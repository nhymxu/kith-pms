package journal

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "github.com/uptrace/bun/driver/sqliteshim"
)

// TestCreateNoDeadlock verifies that Create() doesn't deadlock when updateLastContactForParticipants
// is called after the transaction commits (not inside it).
func TestCreateNoDeadlock(t *testing.T) {
	// Setup in-memory database
	sqldb, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())
	defer db.Close()

	// Create minimal schema
	schema := `
		CREATE TABLE activity (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			occurred_at_date TEXT NOT NULL,
			occurred_at_time TEXT,
			content TEXT,
			created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			updated_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		);

		CREATE TABLE activity_person (
			activity_id INTEGER NOT NULL,
			person_id INTEGER NOT NULL,
			PRIMARY KEY (activity_id, person_id)
		);

		CREATE TABLE person (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			is_self INTEGER DEFAULT 0,
			last_contact_at TEXT,
			created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			updated_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		);

		INSERT INTO person (id, name, is_self) VALUES (1, 'Self Person', 1);
		INSERT INTO person (id, name, is_self) VALUES (2, 'Other Person', 0);
	`

	if _, err := sqldb.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	// Create service with mock people service
	svc := NewService(db)
	svc.PeopleSvc = &mockPeopleService{db: db}

	// Test: Create activity with multiple people (including self)
	// This should complete without deadlock
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oat := "10:00"
	activity := Activity{
		Title:          "Test Activity",
		OccurredAtDate: "2026-05-06",
		OccurredAtTime: &oat,
		Content:        "Test content",
	}

	personIDs := []int64{1, 2} // Self and other person

	// This should complete quickly without hanging
	done := make(chan struct{})

	var (
		id        int64
		createErr error
	)

	go func() {
		id, createErr = svc.Create(ctx, activity, personIDs)

		close(done)
	}()

	select {
	case <-done:
		if createErr != nil {
			t.Fatalf("Create failed: %v", createErr)
		}

		if id == 0 {
			t.Fatal("Expected non-zero activity ID")
		}

		t.Logf("✅ Create completed successfully without deadlock (id=%d)", id)

	case <-ctx.Done():
		t.Fatal("❌ Create timed out - deadlock detected!")
	}

	// Verify the activity was created
	var count int

	err = sqldb.QueryRow("SELECT COUNT(*) FROM activity WHERE id = ?", id).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query activity: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 activity, got %d", count)
	}

	// Verify person links were created
	err = sqldb.QueryRow("SELECT COUNT(*) FROM activity_person WHERE activity_id = ?", id).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query activity_person: %v", err)
	}

	if count != 2 {
		t.Fatalf("expected 2 person links, got %d", count)
	}

	t.Log("✅ All verifications passed")
}

// mockPeopleService implements PeopleServiceInterface for testing
type mockPeopleService struct {
	db *bun.DB
}

func (m *mockPeopleService) GetSelf(ctx context.Context) (*PersonAdapter, error) {
	var (
		id          int64
		lastContact sql.NullString
	)

	err := m.db.QueryRowContext(ctx, "SELECT id, last_contact_at FROM person WHERE is_self = 1 LIMIT 1").
		Scan(&id, &lastContact)
	if err != nil {
		return nil, err
	}

	var lastContactTime *time.Time

	if lastContact.Valid {
		t, _ := time.Parse(time.RFC3339, lastContact.String)
		lastContactTime = &t
	}

	return &PersonAdapter{
		PersonID:      id,
		LastContactAt: lastContactTime,
	}, nil
}

func (m *mockPeopleService) Get(ctx context.Context, id int64) (*PersonAdapter, error) {
	var lastContact sql.NullString

	err := m.db.QueryRowContext(ctx, "SELECT last_contact_at FROM person WHERE id = ?", id).
		Scan(&lastContact)
	if err != nil {
		return nil, err
	}

	var lastContactTime *time.Time

	if lastContact.Valid {
		t, _ := time.Parse(time.RFC3339, lastContact.String)
		lastContactTime = &t
	}

	return &PersonAdapter{
		PersonID:      id,
		LastContactAt: lastContactTime,
	}, nil
}

func (m *mockPeopleService) UpdateLastContact(ctx context.Context, personID int64, contactTime time.Time) error {
	// This simulates the nested transaction that was causing the deadlock
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE person SET last_contact_at = ? WHERE id = ?",
		contactTime.Format(time.RFC3339), personID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
