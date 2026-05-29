package labels_test

import (
	"context"
	"testing"

	"github.com/uptrace/bun"

	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/internal/labels"
	"github.com/nhymxu/kith-pms/internal/people"
)

// openTestDB opens an in-memory SQLite database with all migrations applied.
func openTestDB(t *testing.T) *bun.DB {
	t.Helper()

	db, err := internaldb.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := internaldb.Up(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

func newSvc(t *testing.T) (*labels.Service, *bun.DB) {
	t.Helper()
	db := openTestDB(t)

	return labels.NewService(db), db
}

// ---- tests ------------------------------------------------------------------

func TestCreate_ListWithCounts(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()

	id, err := svc.Create(ctx, "VIP", "#ff0000")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}

	list, err := svc.ListWithCounts(ctx)
	if err != nil {
		t.Fatalf("ListWithCounts: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 label, got %d", len(list))
	}

	if list[0].Name != "VIP" {
		t.Errorf("Name: got %q, want %q", list[0].Name, "VIP")
	}

	if list[0].Color != "#ff0000" {
		t.Errorf("Color: got %q, want %q", list[0].Color, "#ff0000")
	}
	// No people attached yet — count must be 0.
	if list[0].Count != 0 {
		t.Errorf("Count: got %d, want 0", list[0].Count)
	}
}

func TestCreate_ValidationErrors(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		label   string
		color   string
		wantErr error
	}{
		{"empty name", "", "#aabbcc", labels.ErrNameEmpty},
		{"name too long", string(make([]byte, 65)), "#aabbcc", labels.ErrNameTooLong},
		{"bad color no hash", "x", "aabbcc", labels.ErrInvalidColor},
		{"bad color short", "x", "#abc", labels.ErrInvalidColor},
		{"bad color invalid chars", "x", "#xxyyzz", labels.ErrInvalidColor},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(ctx, tc.label, tc.color)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			// errors.Is traversal
			found := false

			e := err
			for e != nil {
				if e == tc.wantErr {
					found = true
					break
				}

				type unwrapper interface{ Unwrap() error }
				if u, ok := e.(unwrapper); ok {
					e = u.Unwrap()
				} else {
					break
				}
			}

			if !found {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestCreate_UniqueConflict(t *testing.T) {
	svc, _ := newSvc(t)
	ctx := context.Background()

	if _, err := svc.Create(ctx, "dup", "#123456"); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := svc.Create(ctx, "dup", "#654321")
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}

	if err != labels.ErrNameConflict {
		t.Errorf("got %v, want ErrNameConflict", err)
	}
}

func TestAttach_Idempotent(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	// Create a person via the people service.
	peopleSvc := people.NewService(db)

	personID, err := peopleSvc.Create(ctx, people.Person{Name: "Alice"}, nil, nil)
	if err != nil {
		t.Fatalf("create person: %v", err)
	}

	labelID, err := svc.Create(ctx, "Friend", "#00ff00")
	if err != nil {
		t.Fatalf("create label: %v", err)
	}

	// Attach twice — should not error and count stays 1.
	if err := svc.Attach(ctx, personID, labelID); err != nil {
		t.Fatalf("first attach: %v", err)
	}

	if err := svc.Attach(ctx, personID, labelID); err != nil {
		t.Fatalf("second attach (idempotent): %v", err)
	}

	// Verify via ListWithCounts: count == 1.
	list, err := svc.ListWithCounts(ctx)
	if err != nil {
		t.Fatalf("ListWithCounts: %v", err)
	}

	if len(list) != 1 || list[0].Count != 1 {
		t.Errorf("expected count=1, got count=%d", list[0].Count)
	}

	// Verify via ListByPersonID.
	attached, err := svc.ListByPersonID(ctx, personID)
	if err != nil {
		t.Fatalf("ListByPersonID: %v", err)
	}

	if len(attached) != 1 {
		t.Fatalf("expected 1 attached label, got %d", len(attached))
	}

	if attached[0].ID != labelID {
		t.Errorf("attached label ID: got %d, want %d", attached[0].ID, labelID)
	}
}

func TestDetach(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	peopleSvc := people.NewService(db)

	personID, err := peopleSvc.Create(ctx, people.Person{Name: "Bob"}, nil, nil)
	if err != nil {
		t.Fatalf("create person: %v", err)
	}

	labelID, err := svc.Create(ctx, "Colleague", "#0000ff")
	if err != nil {
		t.Fatalf("create label: %v", err)
	}

	if err := svc.Attach(ctx, personID, labelID); err != nil {
		t.Fatalf("attach: %v", err)
	}

	// Confirm attached.
	attached, _ := svc.ListByPersonID(ctx, personID)
	if len(attached) != 1 {
		t.Fatalf("expected 1 attached label before detach, got %d", len(attached))
	}

	if err := svc.Detach(ctx, personID, labelID); err != nil {
		t.Fatalf("detach: %v", err)
	}

	// Confirm detached — association gone, but person and label remain.
	attached, _ = svc.ListByPersonID(ctx, personID)
	if len(attached) != 0 {
		t.Errorf("expected 0 attached labels after detach, got %d", len(attached))
	}

	// Label itself still exists.
	l, err := svc.Get(ctx, labelID)
	if err != nil {
		t.Fatalf("Get label after detach: %v", err)
	}

	if l == nil {
		t.Error("label was deleted unexpectedly")
	}

	// Person still exists.
	p, err := peopleSvc.Get(ctx, personID)
	if err != nil {
		t.Fatalf("Get person after detach: %v", err)
	}

	if p == nil {
		t.Error("person was deleted unexpectedly")
	}
}

func TestFilter_AndSemantics(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	peopleSvc := people.NewService(db)

	personA, _ := peopleSvc.Create(ctx, people.Person{Name: "PersonA"}, nil, nil)
	personB, _ := peopleSvc.Create(ctx, people.Person{Name: "PersonB"}, nil, nil)

	labelX, _ := svc.Create(ctx, "LabelX", "#111111")
	labelY, _ := svc.Create(ctx, "LabelY", "#222222")

	// Person A gets both labels; person B gets only X.
	_ = svc.Attach(ctx, personA, labelX)
	_ = svc.Attach(ctx, personA, labelY)
	_ = svc.Attach(ctx, personB, labelX)

	// Filter by both labels — only person A should be returned.
	results, err := peopleSvc.List(ctx, people.ListParams{
		PageSize: 50,
		LabelIDs: []int64{labelX, labelY},
	})
	if err != nil {
		t.Fatalf("List with AND filter: %v", err)
	}

	if len(results.Items) != 1 {
		t.Fatalf("AND filter: got %d results, want 1", len(results.Items))
	}

	if results.Items[0].ID != personA {
		t.Errorf("AND filter: got person %d, want %d (personA)", results.Items[0].ID, personA)
	}

	// Filter by only X — both A and B should be returned.
	results, err = peopleSvc.List(ctx, people.ListParams{
		PageSize: 50,
		LabelIDs: []int64{labelX},
	})
	if err != nil {
		t.Fatalf("List with single label filter: %v", err)
	}

	if len(results.Items) != 2 {
		t.Errorf("single label filter: got %d results, want 2", len(results.Items))
	}
}

func TestDelete_Cascade(t *testing.T) {
	svc, db := newSvc(t)
	ctx := context.Background()

	peopleSvc := people.NewService(db)
	personID, _ := peopleSvc.Create(ctx, people.Person{Name: "Carol"}, nil, nil)
	labelID, _ := svc.Create(ctx, "ToDelete", "#333333")
	_ = svc.Attach(ctx, personID, labelID)

	// Confirm association exists.
	var count int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM person_label WHERE person_id = ? AND label_id = ?`, personID, labelID,
	).Scan(&count); err != nil {
		t.Fatalf("count person_label: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 person_label row before delete, got %d", count)
	}

	// Delete the label.
	if err := svc.Delete(ctx, labelID); err != nil {
		t.Fatalf("Delete label: %v", err)
	}

	// person_label rows should cascade-delete.
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM person_label WHERE label_id = ?`, labelID,
	).Scan(&count); err != nil {
		t.Fatalf("count person_label after label delete: %v", err)
	}

	if count != 0 {
		t.Errorf("person_label not cascaded: got %d rows", count)
	}

	// Person must still exist.
	p, err := peopleSvc.Get(ctx, personID)
	if err != nil {
		t.Fatalf("Get person after label delete: %v", err)
	}

	if p == nil {
		t.Error("person was deleted by label cascade — must not happen")
	}
}
