package journal_test

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/internal/journal"
)

// openTestDB opens an in-memory SQLite database and runs all migrations.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	// Enable foreign keys for cascade behaviour.
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if err := internaldb.Up(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

func newSvc(t *testing.T) *journal.Service {
	t.Helper()
	return journal.NewService(openTestDB(t))
}

// insertPerson inserts a bare person row and returns its ID.
func insertPerson(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()

	res, err := db.Exec(`INSERT INTO person (name) VALUES (?)`, name)
	if err != nil {
		t.Fatalf("insert person %q: %v", name, err)
	}

	id, _ := res.LastInsertId()

	return id
}

// insertLabel inserts a label row and returns its ID.
func insertLabel(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()

	res, err := db.Exec(`INSERT INTO label (name, color) VALUES (?, '#aabbcc')`, name)
	if err != nil {
		t.Fatalf("insert label %q: %v", name, err)
	}

	id, _ := res.LastInsertId()

	return id
}

// attachLabel links a label to a person.
func attachLabel(t *testing.T, db *sql.DB, personID, labelID int64) {
	t.Helper()

	if _, err := db.Exec(
		`INSERT INTO person_label (person_id, label_id) VALUES (?, ?)`,
		personID,
		labelID,
	); err != nil {
		t.Fatalf("attach label: %v", err)
	}
}

// mustCreate creates an activity and asserts no error.
func mustCreate(t *testing.T, svc *journal.Service, a journal.Activity, personIDs []int64) int64 {
	t.Helper()

	id, err := svc.Create(context.Background(), a, personIDs)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	return id
}

// ---- tests ------------------------------------------------------------------

// TestCreate_FTSRoundtrip verifies that a newly created activity is found by
// FTS search on a word in its content.
func TestCreate_FTSRoundtrip(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	mustCreate(t, svc, journal.Activity{
		Title:          "Beach day",
		OccurredAtDate: "2024-07-01",
		Content:        "We went to the seaside and had a wonderful picnic.",
	}, nil)

	list, err := svc.List(ctx, journal.ListParams{Query: "picnic"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(list.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(list.Items))
	}

	if list.Items[0].Title != "Beach day" {
		t.Errorf("unexpected title %q", list.Items[0].Title)
	}
}

// TestUpdate_FTSUpdated verifies that after updating an activity's content the
// FTS index reflects the change: old terms no longer match, new terms do.
func TestUpdate_FTSUpdated(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	id := mustCreate(t, svc, journal.Activity{
		Title:          "Work meeting",
		OccurredAtDate: "2024-08-15",
		Content:        "Discussed quarterly roadmap.",
	}, nil)

	// Search for original word — should be found.
	list, _ := svc.List(ctx, journal.ListParams{Query: "quarterly"})
	if len(list.Items) != 1 {
		t.Fatalf("pre-update: expected 1, got %d", len(list.Items))
	}

	// Update: replace content entirely.
	err := svc.Update(ctx, journal.Activity{
		ID:             id,
		Title:          "Work meeting",
		OccurredAtDate: "2024-08-15",
		Content:        "Talked about new product launch.",
	}, nil)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	// Old term should no longer match.
	old, _ := svc.List(ctx, journal.ListParams{Query: "quarterly"})
	if len(old.Items) != 0 {
		t.Errorf("post-update: old term should not match, got %d results", len(old.Items))
	}

	// New term should match.
	newList, _ := svc.List(ctx, journal.ListParams{Query: "launch"})
	if len(newList.Items) != 1 {
		t.Errorf("post-update: expected 1 for new term, got %d", len(newList.Items))
	}
}

// TestDelete_FTSGone verifies that deleting an activity removes it from FTS results.
func TestDelete_FTSGone(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	id := mustCreate(t, svc, journal.Activity{
		Title:          "Hiking trip",
		OccurredAtDate: "2024-09-10",
		Content:        "Reached the summit at noon.",
	}, nil)

	list, _ := svc.List(ctx, journal.ListParams{Query: "summit"})
	if len(list.Items) != 1 {
		t.Fatalf("pre-delete: expected 1, got %d", len(list.Items))
	}

	if err := svc.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	after, _ := svc.List(ctx, journal.ListParams{Query: "summit"})
	if len(after.Items) != 0 {
		t.Errorf("post-delete: expected 0, got %d", len(after.Items))
	}
}

// TestFilter_PersonIDs verifies that filtering by personID returns only
// activities linked to that person.
func TestFilter_PersonIDs(t *testing.T) {
	db := openTestDB(t)
	svc := journal.NewService(db)
	ctx := context.Background()

	aliceID := insertPerson(t, db, "Alice")
	bobID := insertPerson(t, db, "Bob")

	aID := mustCreate(t, svc, journal.Activity{
		Title:          "Alice event",
		OccurredAtDate: "2024-01-01",
		Content:        "Something with Alice.",
	}, []int64{aliceID})

	mustCreate(t, svc, journal.Activity{
		Title:          "Bob event",
		OccurredAtDate: "2024-01-02",
		Content:        "Something with Bob.",
	}, []int64{bobID})

	list, err := svc.List(ctx, journal.ListParams{PersonIDs: []int64{aliceID}})
	if err != nil {
		t.Fatalf("List by personID: %v", err)
	}

	if len(list.Items) != 1 {
		t.Fatalf("expected 1 activity for Alice, got %d", len(list.Items))
	}

	if list.Items[0].ID != aID {
		t.Errorf("unexpected activity ID %d", list.Items[0].ID)
	}
}

// TestFilter_Combined verifies that multiple simultaneous filters (people +
// label + date + text) return the correct intersection.
func TestFilter_Combined(t *testing.T) {
	db := openTestDB(t)
	svc := journal.NewService(db)
	ctx := context.Background()

	aliceID := insertPerson(t, db, "Alice")
	bobID := insertPerson(t, db, "Bob")
	labelID := insertLabel(t, db, "vip")
	attachLabel(t, db, aliceID, labelID)

	// Activity matches all filters: Alice (has vip label), date in range, text match.
	targetID := mustCreate(t, svc, journal.Activity{
		Title:          "VIP dinner",
		OccurredAtDate: "2024-03-15",
		Content:        "An exclusive gathering with VIP guests.",
	}, []int64{aliceID})

	// Activity outside date range.
	mustCreate(t, svc, journal.Activity{
		Title:          "Old dinner",
		OccurredAtDate: "2023-01-01",
		Content:        "An exclusive gathering.",
	}, []int64{aliceID})

	// Activity with Bob only (label filter should exclude).
	mustCreate(t, svc, journal.Activity{
		Title:          "Bob dinner",
		OccurredAtDate: "2024-03-16",
		Content:        "An exclusive event.",
	}, []int64{bobID})

	list, err := svc.List(ctx, journal.ListParams{
		Query:    "exclusive",
		LabelIDs: []int64{labelID},
		FromDate: "2024-01-01",
		ToDate:   "2024-12-31",
	})
	if err != nil {
		t.Fatalf("List combined: %v", err)
	}

	if len(list.Items) != 1 {
		t.Fatalf("expected 1 combined result, got %d", len(list.Items))
	}

	if list.Items[0].ID != targetID {
		t.Errorf("wrong activity returned: ID %d", list.Items[0].ID)
	}
}

// TestGet_PopulatesPeople verifies that Get returns the activity with its
// linked people populated.
func TestGet_PopulatesPeople(t *testing.T) {
	db := openTestDB(t)
	svc := journal.NewService(db)
	ctx := context.Background()

	aliceID := insertPerson(t, db, "Alice")
	bobID := insertPerson(t, db, "Bob")

	id := mustCreate(t, svc, journal.Activity{
		Title:          "Team lunch",
		OccurredAtDate: "2024-05-20",
		Content:        "Lunch with the team.",
	}, []int64{aliceID, bobID})

	a, err := svc.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if a == nil {
		t.Fatal("expected activity, got nil")
	}

	if len(a.People) != 2 {
		t.Errorf("expected 2 people, got %d", len(a.People))
	}
}

// TestDelete_CascadesLinks verifies that deleting an activity also removes its
// activity_person rows (via ON DELETE CASCADE).
func TestDelete_CascadesLinks(t *testing.T) {
	db := openTestDB(t)
	svc := journal.NewService(db)
	ctx := context.Background()

	aliceID := insertPerson(t, db, "Alice")
	id := mustCreate(t, svc, journal.Activity{
		Title:          "Event",
		OccurredAtDate: "2024-06-01",
		Content:        "Some event.",
	}, []int64{aliceID})

	if err := svc.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM activity_person WHERE activity_id = ?`, id).Scan(&count); err != nil {
		t.Fatalf("count activity_person: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 activity_person rows after delete, got %d", count)
	}
}
