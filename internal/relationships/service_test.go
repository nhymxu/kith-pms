package relationships_test

import (
	"context"
	"testing"

	"github.com/uptrace/bun"

	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/internal/relationships"
)

func openTestDB(t *testing.T) *bun.DB {
	t.Helper()

	db, err := internaldb.Open(":memory:", 1)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := internaldb.Up(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

func newSvc(t *testing.T) (*relationships.Service, *bun.DB) {
	t.Helper()
	db := openTestDB(t)

	return relationships.NewService(db), db
}

// insertPerson inserts a minimal person row directly and returns the id.
func insertPerson(t *testing.T, db *bun.DB, name string) int64 {
	t.Helper()

	res, err := db.Exec(`INSERT INTO person (name) VALUES (?)`, name)
	if err != nil {
		t.Fatalf("insert person %q: %v", name, err)
	}

	id, _ := res.LastInsertId()

	return id
}

// ---- tests ------------------------------------------------------------------

func TestCreateType_NoReverse_OneRow(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	rt, err := svc.CreateType(ctx, "Friend", "")
	if err != nil {
		t.Fatalf("CreateType: %v", err)
	}

	if rt.InverseTypeID != nil {
		t.Errorf("expected nil InverseTypeID, got %v", *rt.InverseTypeID)
	}

	var count int
	db.QueryRowContext(ctx, `SELECT COUNT(*) FROM relationship_type`).Scan(&count)

	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestCreateType_WithReverse_PairedRows(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	rt, err := svc.CreateType(ctx, "Manager", "Reports to")
	if err != nil {
		t.Fatalf("CreateType: %v", err)
	}

	if rt.InverseTypeID == nil {
		t.Fatal("expected InverseTypeID to be set")
	}

	var count int
	db.QueryRowContext(ctx, `SELECT COUNT(*) FROM relationship_type`).Scan(&count)

	if count != 2 {
		t.Errorf("expected 2 rows, got %d", count)
	}

	// Both rows must point at each other.
	inv, err := svc.GetType(ctx, *rt.InverseTypeID)
	if err != nil || inv == nil {
		t.Fatalf("get inverse type: %v", err)
	}

	if inv.InverseTypeID == nil || *inv.InverseTypeID != rt.ID {
		t.Errorf("inverse type does not point back: got %v", inv.InverseTypeID)
	}

	if inv.Name != "Reports to" {
		t.Errorf("inverse name: got %q, want %q", inv.Name, "Reports to")
	}
}

func TestCreateType_SelfReverse_SingleRow(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	rt, err := svc.CreateType(ctx, "Friend", "Friend")
	if err != nil {
		t.Fatalf("CreateType: %v", err)
	}

	if rt.InverseTypeID != nil {
		t.Errorf("expected no inverse when reverse == name, got %v", *rt.InverseTypeID)
	}

	var count int
	db.QueryRowContext(ctx, `SELECT COUNT(*) FROM relationship_type`).Scan(&count)

	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestAttachRelationship_Paired(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")

	rt, _ := svc.CreateType(ctx, "Manager", "Reports to")

	fwdID, err := svc.AttachRelationship(ctx, alice, bob, rt.ID, "since 2024")
	if err != nil {
		t.Fatalf("AttachRelationship: %v", err)
	}

	if fwdID == 0 {
		t.Fatal("expected positive fwdID")
	}

	var count int
	db.QueryRowContext(ctx, `SELECT COUNT(*) FROM person_relationship`).Scan(&count)

	if count != 2 {
		t.Errorf("expected 2 junction rows (paired), got %d", count)
	}
}

func TestAttachRelationship_SymmetricPaired(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")

	rt, _ := svc.CreateType(ctx, "Friend", "Friend")

	_, err := svc.AttachRelationship(ctx, alice, bob, rt.ID, "school")
	if err != nil {
		t.Fatalf("AttachRelationship: %v", err)
	}

	var count int
	db.QueryRowContext(ctx, `SELECT COUNT(*) FROM person_relationship`).Scan(&count)

	if count != 2 {
		t.Errorf("expected 2 junction rows for symmetric type, got %d", count)
	}

	// Both people must see the relationship.
	aliceViews, _ := svc.ListByPerson(ctx, alice)
	bobViews, _ := svc.ListByPerson(ctx, bob)

	if len(aliceViews) != 1 {
		t.Errorf("alice: expected 1 view, got %d", len(aliceViews))
	}

	if len(bobViews) != 1 {
		t.Errorf("bob: expected 1 view, got %d", len(bobViews))
	}
}

func TestDetachRelationship_SymmetricRemovesBoth(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")
	rt, _ := svc.CreateType(ctx, "Friend", "Friend")

	fwdID, _ := svc.AttachRelationship(ctx, alice, bob, rt.ID, "")

	if err := svc.DetachRelationship(ctx, fwdID); err != nil {
		t.Fatalf("DetachRelationship: %v", err)
	}

	var count int
	db.QueryRowContext(ctx, `SELECT COUNT(*) FROM person_relationship`).Scan(&count)

	if count != 0 {
		t.Errorf("expected 0 rows after symmetric detach, got %d", count)
	}
}

func TestAttachRelationship_Unpaired(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")

	rt, _ := svc.CreateType(ctx, "Friend", "")

	_, err := svc.AttachRelationship(ctx, alice, bob, rt.ID, "")
	if err != nil {
		t.Fatalf("AttachRelationship: %v", err)
	}

	var count int
	db.QueryRowContext(ctx, `SELECT COUNT(*) FROM person_relationship`).Scan(&count)

	if count != 1 {
		t.Errorf("expected 1 junction row (unpaired), got %d", count)
	}
}

func TestAttachRelationship_RejectsSelfLoop(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	rt, _ := svc.CreateType(ctx, "Friend", "")

	_, err := svc.AttachRelationship(ctx, alice, alice, rt.ID, "")
	if err == nil {
		t.Fatal("expected ErrSelfRelationship, got nil")
	}

	if !isErr(err, relationships.ErrSelfRelationship) {
		t.Errorf("got %v, want ErrSelfRelationship", err)
	}
}

func TestAttachRelationship_RejectsDuplicate(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")
	rt, _ := svc.CreateType(ctx, "Friend", "")

	if _, err := svc.AttachRelationship(ctx, alice, bob, rt.ID, ""); err != nil {
		t.Fatalf("first attach: %v", err)
	}

	_, err := svc.AttachRelationship(ctx, alice, bob, rt.ID, "")
	if err == nil {
		t.Fatal("expected ErrDuplicateRelationship, got nil")
	}

	if !isErr(err, relationships.ErrDuplicateRelationship) {
		t.Errorf("got %v, want ErrDuplicateRelationship", err)
	}
}

func TestDetachRelationship_RemovesBoth(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")
	rt, _ := svc.CreateType(ctx, "Manager", "Reports to")

	fwdID, _ := svc.AttachRelationship(ctx, alice, bob, rt.ID, "")

	if err := svc.DetachRelationship(ctx, fwdID); err != nil {
		t.Fatalf("DetachRelationship: %v", err)
	}

	var count int
	db.QueryRowContext(ctx, `SELECT COUNT(*) FROM person_relationship`).Scan(&count)

	if count != 0 {
		t.Errorf("expected 0 junction rows after detach, got %d", count)
	}
}

func TestDeleteType_RestrictsWhenInUse(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")
	rt, _ := svc.CreateType(ctx, "Friend", "")

	svc.AttachRelationship(ctx, alice, bob, rt.ID, "")

	err := svc.DeleteType(ctx, rt.ID)
	if err == nil {
		t.Fatal("expected ErrTypeInUse, got nil")
	}

	if !isErr(err, relationships.ErrTypeInUse) {
		t.Errorf("got %v, want ErrTypeInUse", err)
	}
}

func TestListByPerson_ShapeAndOrder(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")
	carol := insertPerson(t, db, "Carol")

	manager, _ := svc.CreateType(ctx, "Manager", "")
	friend, _ := svc.CreateType(ctx, "Friend", "")

	svc.AttachRelationship(ctx, alice, bob, manager.ID, "team lead")
	svc.AttachRelationship(ctx, alice, carol, friend.ID, "college")

	views, err := svc.ListByPerson(ctx, alice)
	if err != nil {
		t.Fatalf("ListByPerson: %v", err)
	}

	if len(views) != 2 {
		t.Fatalf("expected 2 views, got %d", len(views))
	}
	// ORDER BY t.name, p.name → "Friend/Carol" then "Manager/Bob"
	if views[0].TypeName != "Friend" {
		t.Errorf("first view TypeName: got %q, want %q", views[0].TypeName, "Friend")
	}

	if views[0].OtherPersonName != "Carol" {
		t.Errorf("first view OtherPersonName: got %q, want %q", views[0].OtherPersonName, "Carol")
	}

	if views[1].TypeName != "Manager" {
		t.Errorf("second view TypeName: got %q, want %q", views[1].TypeName, "Manager")
	}

	if views[1].Notes != "team lead" {
		t.Errorf("second view Notes: got %q, want %q", views[1].Notes, "team lead")
	}
}

// isErr traverses wrapped errors for equality.
func isErr(err, target error) bool {
	for err != nil {
		if err == target {
			return true
		}

		type unwrapper interface{ Unwrap() error }

		u, ok := err.(unwrapper)
		if !ok {
			break
		}

		err = u.Unwrap()
	}

	return false
}
