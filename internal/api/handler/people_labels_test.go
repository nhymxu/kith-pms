package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nhymxu/kith-pms/internal/api/handler"
)

func TestPeopleLabelsAttach_HappyPath(t *testing.T) {
	db := openTestDB(t)
	labelsSvc := newLabelsService(db)
	personID := insertTestPerson(t, db, "Alice")

	labelID, err := labelsSvc.Create(context.Background(), "friend", "#aabbcc")
	if err != nil {
		t.Fatalf("create label: %v", err)
	}

	h := &handler.PeopleLabelsAPI{Svc: labelsSvc}
	req := jsonRequest(http.MethodPost, "/v1/people/1/labels",
		fmt.Sprintf(`{"label_id":%d}`, labelID))
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.Attach)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}
}

func TestPeopleLabelsAttach_MissingLabelID_Returns400(t *testing.T) {
	db := openTestDB(t)
	h := &handler.PeopleLabelsAPI{Svc: newLabelsService(db)}

	req := jsonRequest(http.MethodPost, "/v1/people/1/labels", `{"label_id":0}`)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": "1"}, h.Attach)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPeopleLabelsAttach_InvalidPersonID_Returns400(t *testing.T) {
	db := openTestDB(t)
	h := &handler.PeopleLabelsAPI{Svc: newLabelsService(db)}

	req := jsonRequest(http.MethodPost, "/v1/people/bad/labels", `{"label_id":1}`)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": "bad"}, h.Attach)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPeopleLabelsDetach_HappyPath(t *testing.T) {
	db := openTestDB(t)
	labelsSvc := newLabelsService(db)
	personID := insertTestPerson(t, db, "Bob")

	labelID, err := labelsSvc.Create(context.Background(), "colleague", "#112233")
	if err != nil {
		t.Fatalf("create label: %v", err)
	}

	_ = labelsSvc.Attach(context.Background(), personID, labelID)

	h := &handler.PeopleLabelsAPI{Svc: labelsSvc}
	req := httptest.NewRequest(http.MethodDelete,
		fmt.Sprintf("/v1/people/%d/labels/%d", personID, labelID), nil)
	rec := execHandler(newTestEcho(), req,
		map[string]string{"id": fmt.Sprintf("%d", personID), "labelID": fmt.Sprintf("%d", labelID)},
		h.Detach)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestPeopleLabelsDetach_InvalidLabelID_Returns400(t *testing.T) {
	db := openTestDB(t)
	h := &handler.PeopleLabelsAPI{Svc: newLabelsService(db)}

	req := httptest.NewRequest(http.MethodDelete, "/v1/people/1/labels/bad", nil)
	rec := execHandler(newTestEcho(), req,
		map[string]string{"id": "1", "labelID": "bad"}, h.Detach)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
