package people_test

import (
	"context"
	"testing"

	"github.com/nhymxu/kith-pms/internal/people"
)

func newLabelSvc(t *testing.T) *people.LabelService {
	t.Helper()
	return people.NewLabelService(openTestDB(t))
}

func TestPeopleLabel_Create_ListWithCounts(t *testing.T) {
	svc := newLabelSvc(t)
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

	if list[0].Count != 0 {
		t.Errorf("Count: got %d, want 0", list[0].Count)
	}
}

func TestPeopleLabel_Create_ValidationErrors(t *testing.T) {
	svc := newLabelSvc(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		label   string
		color   string
		wantErr error
	}{
		{"empty name", "", "#aabbcc", people.ErrNameEmpty},
		{"name too long", string(make([]byte, 65)), "#aabbcc", people.ErrNameTooLong},
		{"bad color no hash", "x", "aabbcc", people.ErrInvalidColor},
		{"bad color short", "x", "#abc", people.ErrInvalidColor},
		{"bad color invalid chars", "x", "#xxyyzz", people.ErrInvalidColor},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(ctx, tc.label, tc.color)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

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

func TestPeopleLabel_Create_UniqueConflict(t *testing.T) {
	svc := newLabelSvc(t)
	ctx := context.Background()

	if _, err := svc.Create(ctx, "dup", "#123456"); err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err := svc.Create(ctx, "dup", "#654321")
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}

	if err != people.ErrNameConflict {
		t.Errorf("got %v, want ErrNameConflict", err)
	}
}

func TestPeopleLabel_Attach_Idempotent(t *testing.T) {
	db := openTestDB(t)
	svc := people.NewLabelService(db)
	peopleSvc := people.NewService(db)
	ctx := context.Background()

	personID, err := peopleSvc.Create(ctx, people.Person{Name: "Alice"}, nil, nil)
	if err != nil {
		t.Fatalf("create person: %v", err)
	}

	labelID, err := svc.Create(ctx, "Friend", "#00ff00")
	if err != nil {
		t.Fatalf("create label: %v", err)
	}

	if err := svc.Attach(ctx, personID, labelID); err != nil {
		t.Fatalf("first attach: %v", err)
	}

	if err := svc.Attach(ctx, personID, labelID); err != nil {
		t.Fatalf("second attach (idempotent): %v", err)
	}

	list, err := svc.ListWithCounts(ctx)
	if err != nil {
		t.Fatalf("ListWithCounts: %v", err)
	}

	if len(list) != 1 || list[0].Count != 1 {
		t.Errorf("expected count=1, got count=%d", list[0].Count)
	}

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

func TestPeopleLabel_Detach(t *testing.T) {
	db := openTestDB(t)
	svc := people.NewLabelService(db)
	peopleSvc := people.NewService(db)
	ctx := context.Background()

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

	attached, _ := svc.ListByPersonID(ctx, personID)
	if len(attached) != 1 {
		t.Fatalf("expected 1 attached label before detach, got %d", len(attached))
	}

	if err := svc.Detach(ctx, personID, labelID); err != nil {
		t.Fatalf("detach: %v", err)
	}

	attached, _ = svc.ListByPersonID(ctx, personID)
	if len(attached) != 0 {
		t.Errorf("expected 0 attached labels after detach, got %d", len(attached))
	}

	l, err := svc.Get(ctx, labelID)
	if err != nil {
		t.Fatalf("Get label after detach: %v", err)
	}

	if l == nil {
		t.Error("label was deleted unexpectedly")
	}

	p, err := peopleSvc.Get(ctx, personID)
	if err != nil {
		t.Fatalf("Get person after detach: %v", err)
	}

	if p == nil {
		t.Error("person was deleted unexpectedly")
	}
}

func TestPeopleLabel_Filter_AndSemantics(t *testing.T) {
	db := openTestDB(t)
	svc := people.NewLabelService(db)
	peopleSvc := people.NewService(db)
	ctx := context.Background()

	personA, _ := peopleSvc.Create(ctx, people.Person{Name: "PersonA"}, nil, nil)
	personB, _ := peopleSvc.Create(ctx, people.Person{Name: "PersonB"}, nil, nil)

	labelX, _ := svc.Create(ctx, "LabelX", "#111111")
	labelY, _ := svc.Create(ctx, "LabelY", "#222222")

	_ = svc.Attach(ctx, personA, labelX)
	_ = svc.Attach(ctx, personA, labelY)
	_ = svc.Attach(ctx, personB, labelX)

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

func TestPeopleLabel_Delete_Cascade(t *testing.T) {
	db := openTestDB(t)
	svc := people.NewLabelService(db)
	peopleSvc := people.NewService(db)
	ctx := context.Background()

	personID, _ := peopleSvc.Create(ctx, people.Person{Name: "Carol"}, nil, nil)
	labelID, _ := svc.Create(ctx, "ToDelete", "#333333")
	_ = svc.Attach(ctx, personID, labelID)

	var count int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM people_label_assignment WHERE person_id = ? AND label_id = ?`, personID, labelID,
	).Scan(&count); err != nil {
		t.Fatalf("count people_label_assignment: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 row before delete, got %d", count)
	}

	if err := svc.Delete(ctx, labelID); err != nil {
		t.Fatalf("Delete label: %v", err)
	}

	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM people_label_assignment WHERE label_id = ?`, labelID,
	).Scan(&count); err != nil {
		t.Fatalf("count after label delete: %v", err)
	}

	if count != 0 {
		t.Errorf("people_label_assignment not cascaded: got %d rows", count)
	}

	p, err := peopleSvc.Get(ctx, personID)
	if err != nil {
		t.Fatalf("Get person after label delete: %v", err)
	}

	if p == nil {
		t.Error("person was deleted by label cascade — must not happen")
	}
}
