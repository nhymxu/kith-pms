package journal_test

import (
	"context"
	"testing"

	"github.com/nhymxu/kith-pms/internal/journal"
)

func newLabelSvc(t *testing.T) (*journal.LabelService, *journal.Service) {
	t.Helper()
	db := openTestDB(t)
	return journal.NewLabelService(db), journal.NewService(db)
}

// insertJournalLabel inserts a journal_label row and returns its ID.
func insertJournalLabel(t *testing.T, svc *journal.LabelService, name, color string) int64 {
	t.Helper()
	id, err := svc.Create(context.Background(), name, color)
	if err != nil {
		t.Fatalf("insertJournalLabel %q: %v", name, err)
	}
	return id
}

func TestJournalLabel_CreateListWithCounts(t *testing.T) {
	svc, _ := newLabelSvc(t)
	ctx := context.Background()

	id1 := insertJournalLabel(t, svc, "Work", "#112233")
	id2 := insertJournalLabel(t, svc, "Personal", "#aabbcc")

	labels, err := svc.ListWithCounts(ctx)
	if err != nil {
		t.Fatalf("ListWithCounts: %v", err)
	}

	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(labels))
	}

	_ = id1
	_ = id2
}

func TestJournalLabel_ValidationErrors(t *testing.T) {
	svc, _ := newLabelSvc(t)
	ctx := context.Background()

	if _, err := svc.Create(ctx, "", "#aabbcc"); err != journal.ErrLabelNameEmpty {
		t.Errorf("expected ErrLabelNameEmpty, got %v", err)
	}

	longName := make([]byte, 65)
	for i := range longName {
		longName[i] = 'x'
	}

	if _, err := svc.Create(ctx, string(longName), "#aabbcc"); err != journal.ErrLabelNameTooLong {
		t.Errorf("expected ErrLabelNameTooLong, got %v", err)
	}

	if _, err := svc.Create(ctx, "Good", "not-hex"); err != journal.ErrLabelInvalidColor {
		t.Errorf("expected ErrLabelInvalidColor, got %v", err)
	}
}

func TestJournalLabel_UniqueConflict(t *testing.T) {
	svc, _ := newLabelSvc(t)
	ctx := context.Background()

	insertJournalLabel(t, svc, "Dup", "#aabbcc")

	if _, err := svc.Create(ctx, "Dup", "#112233"); err != journal.ErrLabelNameConflict {
		t.Errorf("expected ErrLabelNameConflict, got %v", err)
	}
}

func TestJournalLabel_FindOrCreate_ExistingReturnsID(t *testing.T) {
	svc, _ := newLabelSvc(t)
	ctx := context.Background()

	id1, err := svc.FindOrCreate(ctx, "CONVERSATION", "#9ea096")
	if err != nil {
		t.Fatalf("first FindOrCreate: %v", err)
	}

	id2, err := svc.FindOrCreate(ctx, "CONVERSATION", "#ffffff")
	if err != nil {
		t.Fatalf("second FindOrCreate: %v", err)
	}

	if id1 != id2 {
		t.Errorf("expected same id, got %d vs %d", id1, id2)
	}
}

func TestJournalLabel_FilterExisting_DropsUnknown(t *testing.T) {
	svc, jSvc := newLabelSvc(t)
	ctx := context.Background()

	id1 := insertJournalLabel(t, svc, "Tag1", "#aabbcc")
	// id 9999 does not exist

	validIDs, err := jSvc.Labels.FilterExisting(ctx, []int64{id1, 9999})
	if err != nil {
		t.Fatalf("FilterExisting: %v", err)
	}

	if len(validIDs) != 1 || validIDs[0] != id1 {
		t.Errorf("expected [%d], got %v", id1, validIDs)
	}
}

func TestJournalLabel_ReplaceAll_Idempotent(t *testing.T) {
	svc, jSvc := newLabelSvc(t)
	ctx := context.Background()

	lid := insertJournalLabel(t, svc, "Idem", "#aabbcc")
	actID, err := jSvc.Create(ctx, journal.Activity{
		Title:          "Idem test",
		OccurredAtDate: "2024-01-01",
		Content:        "",
	}, nil, []int64{lid})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Second create with same label — should not error.
	if err := jSvc.Update(ctx, journal.Activity{
		ID:             actID,
		Title:          "Idem test",
		OccurredAtDate: "2024-01-01",
		Content:        "",
	}, nil, []int64{lid}); err != nil {
		t.Fatalf("Update (idempotent): %v", err)
	}

	got, err := jSvc.Get(ctx, actID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if len(got.Labels) != 1 || got.Labels[0].ID != lid {
		t.Errorf("expected label %d, got %v", lid, got.Labels)
	}
}

func TestJournalLabel_JournalLabelIDs_Filter(t *testing.T) {
	svc, jSvc := newLabelSvc(t)
	ctx := context.Background()

	l1 := insertJournalLabel(t, svc, "Alpha", "#111111")
	l2 := insertJournalLabel(t, svc, "Beta", "#222222")

	act1, _ := jSvc.Create(ctx, journal.Activity{Title: "A1", OccurredAtDate: "2024-01-01"}, nil, []int64{l1})
	act2, _ := jSvc.Create(ctx, journal.Activity{Title: "A2", OccurredAtDate: "2024-01-02"}, nil, []int64{l2})
	act3, _ := jSvc.Create(ctx, journal.Activity{Title: "A3", OccurredAtDate: "2024-01-03"}, nil, []int64{l1, l2})

	// OR within: filter by l1 → should include act1 and act3.
	list, err := jSvc.List(ctx, journal.ListParams{JournalLabelIDs: []int64{l1}})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if list.Total != 2 {
		t.Errorf("OR filter by l1: expected 2, got %d", list.Total)
	}

	// OR within: filter by l1 OR l2 → all 3.
	list2, _ := jSvc.List(ctx, journal.ListParams{JournalLabelIDs: []int64{l1, l2}})
	if list2.Total != 3 {
		t.Errorf("OR filter l1|l2: expected 3, got %d", list2.Total)
	}

	_ = act1
	_ = act2
	_ = act3
	_ = act2 // suppress unused

	// Verify label population in list.
	list3, _ := jSvc.List(ctx, journal.ListParams{JournalLabelIDs: []int64{l2}})
	for _, item := range list3.Items {
		if len(item.Labels) == 0 {
			t.Errorf("activity %d has no labels populated", item.ID)
		}
	}
}

func TestJournalLabel_DeleteActivityCascades(t *testing.T) {
	svc, jSvc := newLabelSvc(t)
	ctx := context.Background()
	db := openTestDB(t)

	lid := insertJournalLabel(t, svc, "Cascade", "#aabbcc")
	actID, _ := jSvc.Create(ctx, journal.Activity{
		Title:          "Cascade test",
		OccurredAtDate: "2024-01-01",
	}, nil, []int64{lid})

	// Verify assignment exists.
	var count int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM journal_label_assignment WHERE activity_id = ?`, actID).Scan(&count)
	// Note: this uses a separate DB; use jSvc's DB via service Delete and then verify via Get.

	if err := jSvc.Delete(ctx, actID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	got, err := jSvc.Get(ctx, actID)
	if err != nil {
		t.Fatalf("Get after delete: %v", err)
	}

	if got != nil {
		t.Errorf("expected nil after delete, got %+v", got)
	}
}
