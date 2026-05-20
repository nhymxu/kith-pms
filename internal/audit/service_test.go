package audit_test

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/nhymxu/kith-pms/internal/audit"
	internaldb "github.com/nhymxu/kith-pms/internal/db"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if err := internaldb.Up(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

func TestAuditLog_Create(t *testing.T) {
	db := openTestDB(t)
	svc := audit.NewService(db)
	ctx := context.Background()

	svc.Log(ctx, audit.EntityPerson, 1, "Jane Doe", audit.ActionCreate)

	entries, err := svc.List(ctx, audit.ListParams{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}

	e := entries[0]
	if e.EntityType != audit.EntityPerson {
		t.Errorf("entity_type: want %q, got %q", audit.EntityPerson, e.EntityType)
	}

	if e.EntityID != 1 {
		t.Errorf("entity_id: want 1, got %d", e.EntityID)
	}

	if e.EntityName != "Jane Doe" {
		t.Errorf("entity_name: want %q, got %q", "Jane Doe", e.EntityName)
	}

	if e.Action != audit.ActionCreate {
		t.Errorf("action: want %q, got %q", audit.ActionCreate, e.Action)
	}

	if e.ActorID != nil {
		t.Errorf("actor_id: want nil, got %v", *e.ActorID)
	}
}

func TestAuditLog_WithActor(t *testing.T) {
	db := openTestDB(t)
	svc := audit.NewService(db)
	ctx := audit.WithActor(context.Background(), 7)

	svc.Log(ctx, audit.EntityJournal, 5, "Lunch meeting", audit.ActionUpdate)

	entries, err := svc.List(ctx, audit.ListParams{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}

	if entries[0].ActorID == nil {
		t.Fatal("actor_id: want non-nil")
	}

	if *entries[0].ActorID != 7 {
		t.Errorf("actor_id: want 7, got %d", *entries[0].ActorID)
	}
}

func TestAuditLog_NilActor(t *testing.T) {
	db := openTestDB(t)
	svc := audit.NewService(db)
	ctx := context.Background()

	svc.Log(ctx, audit.EntityLabel, 2, "Friend", audit.ActionDelete)

	entries, err := svc.List(ctx, audit.ListParams{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("want 1 entry, got %d", len(entries))
	}

	if entries[0].ActorID != nil {
		t.Errorf("actor_id: want nil, got %v", *entries[0].ActorID)
	}
}

func TestAuditLog_FilterByEntityType(t *testing.T) {
	db := openTestDB(t)
	svc := audit.NewService(db)
	ctx := context.Background()

	svc.Log(ctx, audit.EntityPerson, 1, "Alice", audit.ActionCreate)
	svc.Log(ctx, audit.EntityPerson, 2, "Bob", audit.ActionCreate)
	svc.Log(ctx, audit.EntityJournal, 1, "Notes", audit.ActionCreate)

	all, err := svc.List(ctx, audit.ListParams{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("all: want 3, got %d", len(all))
	}

	people, err := svc.List(ctx, audit.ListParams{EntityType: audit.EntityPerson, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list people: %v", err)
	}

	if len(people) != 2 {
		t.Errorf("people filter: want 2, got %d", len(people))
	}

	journal, err := svc.List(ctx, audit.ListParams{EntityType: audit.EntityJournal, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list journal: %v", err)
	}

	if len(journal) != 1 {
		t.Errorf("journal filter: want 1, got %d", len(journal))
	}
}

func TestAuditLog_FilterByEntityID(t *testing.T) {
	db := openTestDB(t)
	svc := audit.NewService(db)
	ctx := context.Background()

	svc.Log(ctx, audit.EntityPerson, 1, "Alice", audit.ActionCreate)
	svc.Log(ctx, audit.EntityPerson, 1, "Alice", audit.ActionUpdate)
	svc.Log(ctx, audit.EntityPerson, 2, "Bob", audit.ActionCreate)

	results, err := svc.List(ctx, audit.ListParams{
		EntityType: audit.EntityPerson,
		EntityID:   1,
		Page:       1,
		PageSize:   10,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("want 2 entries for entity_id=1, got %d", len(results))
	}

	for _, e := range results {
		if e.EntityID != 1 {
			t.Errorf("entity_id: want 1, got %d", e.EntityID)
		}
	}
}

func TestAuditLog_BestEffort_ClosedDB(t *testing.T) {
	db := openTestDB(t)
	svc := audit.NewService(db)
	_ = db.Close()
	// Must not panic — just logs a warning.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panicked: %v", r)
		}
	}()

	svc.Log(context.Background(), audit.EntityPerson, 1, "Jane", audit.ActionCreate)
}

func TestAuditLog_OrderedNewestFirst(t *testing.T) {
	db := openTestDB(t)
	svc := audit.NewService(db)
	ctx := context.Background()

	svc.Log(ctx, audit.EntityPerson, 1, "Jane Doe", audit.ActionCreate)
	svc.Log(ctx, audit.EntityPerson, 1, "Jane Smith", audit.ActionUpdate)
	svc.Log(ctx, audit.EntityPerson, 1, "Jane Smith", audit.ActionDelete)

	entries, err := svc.List(ctx, audit.ListParams{
		EntityType: audit.EntityPerson,
		EntityID:   1,
		Page:       1,
		PageSize:   10,
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("want 3, got %d", len(entries))
	}
	// Newest first
	if entries[0].Action != audit.ActionDelete {
		t.Errorf("entries[0]: want delete, got %q", entries[0].Action)
	}

	if entries[2].Action != audit.ActionCreate {
		t.Errorf("entries[2]: want create, got %q", entries[2].Action)
	}
}

// ---- Purge tests ------------------------------------------------------------

func insertAuditAt(t *testing.T, db *sql.DB, createdAt string) {
	t.Helper()

	_, err := db.ExecContext(context.Background(),
		`INSERT INTO audit_log (entity_type, entity_id, entity_name, action, created_at)
		 VALUES ('person', 1, 'Test', 'create', ?)`, createdAt)
	if err != nil {
		t.Fatalf("insertAuditAt: %v", err)
	}
}

func TestService_Purge_Disabled(t *testing.T) {
	svc := audit.NewService(openTestDB(t))

	n, err := svc.Purge(context.Background(), 0)
	if err != nil {
		t.Fatalf("purge: %v", err)
	}

	if n != 0 {
		t.Errorf("want 0 deleted, got %d", n)
	}
}

func TestService_Purge_DeletesOldEntries(t *testing.T) {
	db := openTestDB(t)
	svc := audit.NewService(db)

	// one entry 91 days ago, one entry 1 day ago
	insertAuditAt(t, db, "2026-02-18T00:00:00Z") // 91 days before 2026-05-20
	insertAuditAt(t, db, "2026-05-19T00:00:00Z") // 1 day ago

	n, err := svc.Purge(context.Background(), 90)
	if err != nil {
		t.Fatalf("purge: %v", err)
	}

	if n != 1 {
		t.Errorf("want 1 deleted, got %d", n)
	}

	entries, err := svc.List(context.Background(), audit.ListParams{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("want 1 remaining entry, got %d", len(entries))
	}
}

func TestService_Purge_NothingToDelete(t *testing.T) {
	db := openTestDB(t)
	svc := audit.NewService(db)

	insertAuditAt(t, db, "2026-05-19T00:00:00Z") // 1 day ago

	n, err := svc.Purge(context.Background(), 90)
	if err != nil {
		t.Fatalf("purge: %v", err)
	}

	if n != 0 {
		t.Errorf("want 0 deleted, got %d", n)
	}
}

func TestActorContext_RoundTrip(t *testing.T) {
	ctx := audit.WithActor(context.Background(), 42)

	got := audit.ActorFromCtx(ctx)
	if got == nil {
		t.Fatal("ActorFromCtx: want non-nil")
	}

	if *got != 42 {
		t.Errorf("ActorFromCtx: want 42, got %d", *got)
	}
}

func TestActorFromCtx_Missing(t *testing.T) {
	got := audit.ActorFromCtx(context.Background())
	if got != nil {
		t.Errorf("ActorFromCtx: want nil, got %v", *got)
	}
}

func TestActorFromCtx_ZeroIsValid(t *testing.T) {
	ctx := audit.WithActor(context.Background(), 0)

	got := audit.ActorFromCtx(ctx)
	if got == nil {
		t.Fatal("ActorFromCtx: want non-nil for actor=0")
	}

	if *got != 0 {
		t.Errorf("ActorFromCtx: want 0, got %d", *got)
	}
}
