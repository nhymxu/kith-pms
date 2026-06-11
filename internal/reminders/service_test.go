package reminders

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	_ "github.com/uptrace/bun/driver/sqliteshim"
)

func setupTestDB(t *testing.T) *bun.DB {
	sqldb, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())

	_, err = sqldb.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	_, err = sqldb.Exec(`
		CREATE TABLE person (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create person table: %v", err)
	}

	_, err = sqldb.Exec(`
		CREATE TABLE important_date (
			id INTEGER PRIMARY KEY,
			person_id INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
			kind TEXT NOT NULL DEFAULT 'other',
			date_value TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create important_date table: %v", err)
	}

	_, err = sqldb.Exec(`
		CREATE TABLE reminder (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			notes TEXT NOT NULL DEFAULT '',
			due_date TEXT NOT NULL,
			person_id INTEGER REFERENCES person(id) ON DELETE SET NULL,
			important_date_id INTEGER REFERENCES important_date(id) ON DELETE SET NULL,
			completed INTEGER NOT NULL DEFAULT 0,
			completed_at TEXT,
			recurrence_rule TEXT,
			recurrence_end_date TEXT,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		);
		CREATE INDEX idx_reminder_due_date ON reminder(due_date);
		CREATE INDEX idx_reminder_person ON reminder(person_id);
		CREATE INDEX idx_reminder_completed ON reminder(completed);
	`)
	if err != nil {
		t.Fatalf("create reminder table: %v", err)
	}

	return db
}

func TestService_CreateAndGetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	dueDate := time.Now().AddDate(0, 0, 7)
	rem := &Reminder{
		Title:   "Follow up with Alice",
		Notes:   "Discuss project timeline",
		DueDate: dueDate,
	}

	id, err := svc.Create(ctx, rem)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if id == 0 {
		t.Fatal("expected non-zero ID")
	}

	fetched, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if fetched.Title != rem.Title {
		t.Errorf("Title = %q, want %q", fetched.Title, rem.Title)
	}

	if fetched.Notes != rem.Notes {
		t.Errorf("Notes = %q, want %q", fetched.Notes, rem.Notes)
	}
	// Compare timestamps truncated to second precision (SQLite loses nanosecond precision)
	if !fetched.DueDate.Truncate(time.Second).Equal(dueDate.Truncate(time.Second)) {
		t.Errorf("DueDate = %v, want %v", fetched.DueDate, dueDate)
	}

	if fetched.Completed {
		t.Error("expected Completed = false")
	}
}

func TestService_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	rem := &Reminder{
		Title:   "Original title",
		Notes:   "Original notes",
		DueDate: time.Now().AddDate(0, 0, 1),
	}

	id, err := svc.Create(ctx, rem)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	fetched, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	fetched.Title = "Updated title"
	fetched.Notes = "Updated notes"
	fetched.DueDate = time.Now().AddDate(0, 0, 14)

	if err := svc.Update(ctx, fetched); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}

	if updated.Title != "Updated title" {
		t.Errorf("Title = %q, want %q", updated.Title, "Updated title")
	}

	if updated.Notes != "Updated notes" {
		t.Errorf("Notes = %q, want %q", updated.Notes, "Updated notes")
	}
}

func TestService_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	rem := &Reminder{
		Title:   "To be deleted",
		DueDate: time.Now().AddDate(0, 0, 1),
	}

	id, err := svc.Create(ctx, rem)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = svc.GetByID(ctx, id)
	if err == nil {
		t.Error("expected error when fetching deleted reminder")
	}

	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestService_ListUpcoming(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	now := time.Now()

	reminders := []*Reminder{
		{Title: "Due tomorrow", DueDate: now.AddDate(0, 0, 1)},
		{Title: "Due in 3 days", DueDate: now.AddDate(0, 0, 3)},
		{Title: "Due in 10 days", DueDate: now.AddDate(0, 0, 10)},
		{Title: "Due yesterday", DueDate: now.AddDate(0, 0, -1)},
	}

	for _, r := range reminders {
		_, err := svc.Create(ctx, r)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	upcoming, err := svc.GetUpcoming(ctx, 7)
	if err != nil {
		t.Fatalf("GetUpcoming: %v", err)
	}

	if len(upcoming) != 2 {
		t.Errorf("expected 2 upcoming reminders, got %d", len(upcoming))
	}

	if len(upcoming) > 0 && upcoming[0].Title != "Due tomorrow" {
		t.Errorf("first upcoming = %q, want %q", upcoming[0].Title, "Due tomorrow")
	}
}

func TestService_ListOverdue(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	now := time.Now()

	reminders := []*Reminder{
		{Title: "Overdue 1 day", DueDate: now.AddDate(0, 0, -1)},
		{Title: "Overdue 3 days", DueDate: now.AddDate(0, 0, -3)},
		{Title: "Due tomorrow", DueDate: now.AddDate(0, 0, 1)},
	}

	for _, r := range reminders {
		_, err := svc.Create(ctx, r)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	overdue, err := svc.GetOverdue(ctx)
	if err != nil {
		t.Fatalf("GetOverdue: %v", err)
	}

	if len(overdue) != 2 {
		t.Errorf("expected 2 overdue reminders, got %d", len(overdue))
	}
}

func TestService_MarkComplete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	rem := &Reminder{
		Title:   "Task to complete",
		DueDate: time.Now().AddDate(0, 0, 1),
	}

	id, err := svc.Create(ctx, rem)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.MarkComplete(ctx, id); err != nil {
		t.Fatalf("MarkComplete: %v", err)
	}

	fetched, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if !fetched.Completed {
		t.Error("expected Completed = true")
	}

	if fetched.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
}

func TestService_ListWithPersonFilter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	res, err := db.ExecContext(ctx, "INSERT INTO person (name) VALUES (?)", "Alice")
	if err != nil {
		t.Fatalf("insert person: %v", err)
	}

	personID, _ := res.LastInsertId()

	reminders := []*Reminder{
		{Title: "Alice reminder", DueDate: time.Now().AddDate(0, 0, 1), PersonID: &personID},
		{Title: "General reminder", DueDate: time.Now().AddDate(0, 0, 2)},
	}

	for _, r := range reminders {
		_, err := svc.Create(ctx, r)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	filtered, err := svc.List(ctx, ListParams{PersonID: &personID, PageSize: 100, Page: 1})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("expected 1 reminder for person, got %d", len(filtered))
	}

	if len(filtered) > 0 && filtered[0].Title != "Alice reminder" {
		t.Errorf("Title = %q, want %q", filtered[0].Title, "Alice reminder")
	}
}

func TestService_ListByStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	now := time.Now()

	reminders := []*Reminder{
		{Title: "Pending future", DueDate: now.AddDate(0, 0, 1)},
		{Title: "Overdue", DueDate: now.AddDate(0, 0, -1)},
	}

	for _, r := range reminders {
		id, err := svc.Create(ctx, r)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		if r.Title == "Pending future" {
			if err := svc.MarkComplete(ctx, id); err != nil {
				t.Fatalf("MarkComplete: %v", err)
			}
		}
	}

	pending, err := svc.List(ctx, ListParams{Status: "pending", PageSize: 100, Page: 1})
	if err != nil {
		t.Fatalf("List pending: %v", err)
	}

	if len(pending) != 1 {
		t.Errorf("expected 1 pending, got %d", len(pending))
	}

	completed, err := svc.List(ctx, ListParams{Status: "completed", PageSize: 100, Page: 1})
	if err != nil {
		t.Fatalf("List completed: %v", err)
	}

	if len(completed) != 1 {
		t.Errorf("expected 1 completed, got %d", len(completed))
	}

	overdue, err := svc.List(ctx, ListParams{Status: "overdue", PageSize: 100, Page: 1})
	if err != nil {
		t.Fatalf("List overdue: %v", err)
	}

	if len(overdue) != 1 {
		t.Errorf("expected 1 overdue, got %d", len(overdue))
	}
}

func TestMarkComplete_OneOff_NoSpawn(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	id, err := svc.Create(ctx, &Reminder{
		Title:   "One-off",
		DueDate: time.Now().AddDate(0, 0, 1),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.MarkComplete(ctx, id); err != nil {
		t.Fatalf("MarkComplete: %v", err)
	}

	all, err := svc.List(ctx, ListParams{PageSize: 100, Page: 1})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(all) != 1 {
		t.Errorf("expected 1 row (no spawn), got %d", len(all))
	}
}

func TestMarkComplete_Recurring_Daily_Spawns(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	due := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)

	id, err := svc.Create(ctx, &Reminder{
		Title:          "Daily standup",
		DueDate:        due,
		RecurrenceRule: &RecurrenceRule{Type: RecurrenceDaily},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.MarkComplete(ctx, id); err != nil {
		t.Fatalf("MarkComplete: %v", err)
	}

	all, err := svc.List(ctx, ListParams{PageSize: 100, Page: 1})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(all) != 2 {
		t.Fatalf("expected 2 rows (original + spawn), got %d", len(all))
	}

	var spawned *ReminderWithPerson

	for i := range all {
		if !all[i].Completed {
			spawned = &all[i]
		}
	}

	if spawned == nil {
		t.Fatal("expected a non-completed spawned reminder")
		return
	}

	wantDue := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)
	if !spawned.DueDate.Equal(wantDue) {
		t.Errorf("spawned DueDate = %v, want %v", spawned.DueDate, wantDue)
	}
}

func TestMarkComplete_Recurring_EndDateReached_NoSpawn(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	due := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)

	id, err := svc.Create(ctx, &Reminder{
		Title:             "Expiring daily",
		DueDate:           due,
		RecurrenceRule:    &RecurrenceRule{Type: RecurrenceDaily},
		RecurrenceEndDate: &endDate,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.MarkComplete(ctx, id); err != nil {
		t.Fatalf("MarkComplete: %v", err)
	}

	all, err := svc.List(ctx, ListParams{PageSize: 100, Page: 1})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(all) != 1 {
		t.Errorf("expected 1 row (end date reached, no spawn), got %d", len(all))
	}
}
