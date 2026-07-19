package note

import (
	"context"
	"database/sql"
	"testing"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/audit"
	"github.com/nhymxu/kith-pms/internal/testutil"
)

func insertPerson(t *testing.T, db *bun.DB, name string) int64 {
	t.Helper()

	res, err := db.ExecContext(context.Background(), "INSERT INTO person (name) VALUES (?)", name)
	if err != nil {
		t.Fatalf("insert person: %v", err)
	}

	id, _ := res.LastInsertId()

	return id
}

func TestNoteCRUD(t *testing.T) {
	db := testutil.NewDB(t)

	ctx := context.Background()
	svc := NewService(db)
	personID := insertPerson(t, db, "Alice")

	n := &Note{PersonID: personID, Title: "Ideas", Content: "Buy a gift for Alice"}

	id, err := svc.Create(ctx, n)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if id <= 0 {
		t.Fatalf("Create returned id=%d, want >0", id)
	}

	got, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.Title != "Ideas" || got.Content != "Buy a gift for Alice" {
		t.Errorf("got %+v, want title=Ideas content=%q", got, "Buy a gift for Alice")
	}

	got.Title = "Updated ideas"
	got.Content = "Updated content"

	if err := svc.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, err := svc.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}

	if updated.Title != "Updated ideas" {
		t.Errorf("updated Title = %q, want Updated ideas", updated.Title)
	}

	if !updated.UpdatedAt.After(updated.CreatedAt) && !updated.UpdatedAt.Equal(updated.CreatedAt) {
		t.Errorf("UpdatedAt = %v, want >= CreatedAt %v", updated.UpdatedAt, updated.CreatedAt)
	}

	if err := svc.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := svc.GetByID(ctx, id); err != sql.ErrNoRows {
		t.Errorf("GetByID after delete: got %v, want sql.ErrNoRows", err)
	}
}

func TestListByPersonOrderingAndPagination(t *testing.T) {
	db := testutil.NewDB(t)

	ctx := context.Background()
	svc := NewService(db)
	personID := insertPerson(t, db, "Bob")
	other := insertPerson(t, db, "Other")

	titles := []string{"first", "second", "third"}
	for _, title := range titles {
		if _, err := svc.Create(ctx, &Note{PersonID: personID, Title: title, Content: "c"}); err != nil {
			t.Fatalf("Create %s: %v", title, err)
		}
	}

	if _, err := svc.Create(ctx, &Note{PersonID: other, Title: "not mine", Content: "c"}); err != nil {
		t.Fatalf("Create other: %v", err)
	}

	all, err := svc.ListByPerson(ctx, personID, 1, 50)
	if err != nil {
		t.Fatalf("ListByPerson: %v", err)
	}

	if all.Total != 3 || len(all.Items) != 3 {
		t.Fatalf("ListByPerson: got total=%d items=%d, want 3/3", all.Total, len(all.Items))
	}

	if all.Items[0].Title != "third" || all.Items[2].Title != "first" {
		t.Errorf("ListByPerson order = %q,%q,%q, want third,second,first",
			all.Items[0].Title, all.Items[1].Title, all.Items[2].Title)
	}

	page1, err := svc.ListByPerson(ctx, personID, 1, 2)
	if err != nil {
		t.Fatalf("ListByPerson page1: %v", err)
	}

	if len(page1.Items) != 2 {
		t.Fatalf("page1 items = %d, want 2", len(page1.Items))
	}

	page2, err := svc.ListByPerson(ctx, personID, 2, 2)
	if err != nil {
		t.Fatalf("ListByPerson page2: %v", err)
	}

	if len(page2.Items) != 1 || page2.Items[0].Title != "first" {
		t.Fatalf("page2 = %+v, want 1 item titled first", page2.Items)
	}
}

func TestNotePersonCascadeDelete(t *testing.T) {
	db := testutil.NewDB(t)

	ctx := context.Background()
	svc := NewService(db)
	personID := insertPerson(t, db, "Carol")

	if _, err := svc.Create(ctx, &Note{PersonID: personID, Title: "t", Content: "c"}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := db.ExecContext(ctx, "DELETE FROM person WHERE id = ?", personID); err != nil {
		t.Fatalf("delete person: %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM note WHERE person_id = ?", personID).
		Scan(&count); err != nil {
		t.Fatalf("count notes: %v", err)
	}

	if count != 0 {
		t.Errorf("got %d notes after person delete, want 0", count)
	}
}

func TestNoteAudit(t *testing.T) {
	db := testutil.NewDB(t)

	ctx := context.Background()
	svc := NewService(db)
	svc.Audit = audit.NewService(db)
	personID := insertPerson(t, db, "Dave")

	id, err := svc.Create(ctx, &Note{PersonID: personID, Title: "Titled note", Content: "content"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	entries, err := svc.Audit.List(ctx, audit.ListParams{EntityType: audit.EntityNote, PageSize: 10, Page: 1})
	if err != nil {
		t.Fatalf("Audit.List: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("audit entries = %d, want 1", len(entries))
	}

	entry := entries[0]
	if entry.EntityName != "Dave" {
		t.Errorf("EntityName = %q, want Dave", entry.EntityName)
	}

	if entry.Action != audit.ActionCreate {
		t.Errorf("Action = %q, want create", entry.Action)
	}

	if entry.Metadata == nil || entry.Metadata.Label != "Titled note" {
		t.Errorf("Metadata.Label = %v, want Titled note", entry.Metadata)
	}

	untitled, err := svc.Create(ctx, &Note{
		PersonID: personID,
		Content:  "this content has no title so the audit log should fall back to a snippet of it instead",
	})
	if err != nil {
		t.Fatalf("Create untitled: %v", err)
	}

	entry2, err := svc.Audit.List(
		ctx,
		audit.ListParams{EntityType: audit.EntityNote, EntityID: untitled, PageSize: 10, Page: 1},
	)
	if err != nil {
		t.Fatalf("Audit.List untitled: %v", err)
	}

	if len(entry2) != 1 || entry2[0].Metadata == nil {
		t.Fatalf("audit entry for untitled note missing metadata: %+v", entry2)
	}

	if entry2[0].Metadata.Label == "" {
		t.Error("Metadata.Label snippet fallback is empty, want non-empty content snippet")
	}

	if err := svc.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	delEntries, err := svc.Audit.List(
		ctx,
		audit.ListParams{EntityType: audit.EntityNote, EntityID: id, Page: 1, PageSize: 10},
	)
	if err != nil {
		t.Fatalf("Audit.List after delete: %v", err)
	}

	found := false

	for _, e := range delEntries {
		if e.Action == audit.ActionDelete {
			found = true
		}
	}

	if !found {
		t.Errorf("no delete audit entry found for note %d", id)
	}
}

func TestNoteFTS(t *testing.T) {
	db := testutil.NewDB(t)

	ctx := context.Background()
	svc := NewService(db)
	personID := insertPerson(t, db, "Eve")

	id, err := svc.Create(
		ctx,
		&Note{PersonID: personID, Title: "Searchable title", Content: "uniquefirstoken uniquesecondtoken"},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var ftsCount int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM note_fts WHERE note_fts MATCH 'uniquefirstoken'").
		Scan(&ftsCount); err != nil {
		t.Fatalf("query note_fts after insert: %v", err)
	}

	if ftsCount != 1 {
		t.Errorf("note_fts rows after insert = %d, want 1", ftsCount)
	}

	if err := svc.Delete(ctx, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	var ftsCountAfterDelete int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM note_fts WHERE note_fts MATCH 'uniquefirstoken'").
		Scan(&ftsCountAfterDelete); err != nil {
		t.Fatalf("query note_fts after delete: %v", err)
	}

	if ftsCountAfterDelete != 0 {
		t.Errorf("note_fts rows after delete = %d, want 0", ftsCountAfterDelete)
	}
}
