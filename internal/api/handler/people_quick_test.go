package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nhymxu/kith-pms/internal/api/handler"
)

// ---- QuickJournal -----------------------------------------------------------

func TestPeopleQuickJournal_HappyPath(t *testing.T) {
	db := openTestDB(t)
	personID := insertTestPerson(t, db, "Alice")
	h := &handler.PeopleQuickAPI{
		PeopleSvc:  newPeopleService(db),
		JournalSvc: newJournalService(db),
	}

	body := `{"title":"Caught up","occurred_at_date":"2026-01-15","content":"Great chat"}`
	req := jsonRequest(http.MethodPost, "/v1/people/1/journal/quick", body)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.QuickJournal)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"id"`) {
		t.Fatalf("expected id in response, got: %s", rec.Body.String())
	}
}

func TestPeopleQuickJournal_MissingTitle_Returns422(t *testing.T) {
	db := openTestDB(t)
	personID := insertTestPerson(t, db, "Bob")
	h := &handler.PeopleQuickAPI{
		PeopleSvc:  newPeopleService(db),
		JournalSvc: newJournalService(db),
	}

	req := jsonRequest(http.MethodPost, "/v1/people/1/journal/quick", `{"title":""}`)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.QuickJournal)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestPeopleQuickJournal_PersonNotFound_Returns404(t *testing.T) {
	db := openTestDB(t)
	h := &handler.PeopleQuickAPI{
		PeopleSvc:  newPeopleService(db),
		JournalSvc: newJournalService(db),
	}

	body := `{"title":"Note"}`
	req := jsonRequest(http.MethodPost, "/v1/people/999/journal/quick", body)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": "999"}, h.QuickJournal)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestPeopleQuickJournal_InvalidDate_Returns422(t *testing.T) {
	db := openTestDB(t)
	personID := insertTestPerson(t, db, "Carol")
	h := &handler.PeopleQuickAPI{
		PeopleSvc:  newPeopleService(db),
		JournalSvc: newJournalService(db),
	}

	body := `{"title":"Test","occurred_at_date":"not-a-date"}`
	req := jsonRequest(http.MethodPost, "/v1/people/1/journal/quick", body)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.QuickJournal)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

// ---- QuickGift --------------------------------------------------------------

func TestPeopleQuickGift_HappyPath(t *testing.T) {
	db := openTestDB(t)
	personID := insertTestPerson(t, db, "Dave")
	h := &handler.PeopleQuickAPI{
		PeopleSvc: newPeopleService(db),
		GiftsSvc:  newGiftsService(db),
	}

	body := `{"title":"Chocolate","direction":"given","currency":"USD"}`
	req := jsonRequest(http.MethodPost, "/v1/people/1/gifts/quick", body)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.QuickGift)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", rec.Code, rec.Body.String())
	}
}

func TestPeopleQuickGift_MissingTitle_Returns422(t *testing.T) {
	db := openTestDB(t)
	personID := insertTestPerson(t, db, "Eve")
	h := &handler.PeopleQuickAPI{
		PeopleSvc: newPeopleService(db),
		GiftsSvc:  newGiftsService(db),
	}

	req := jsonRequest(http.MethodPost, "/v1/people/1/gifts/quick", `{"title":""}`)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.QuickGift)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestPeopleQuickGift_PersonNotFound_Returns404(t *testing.T) {
	db := openTestDB(t)
	h := &handler.PeopleQuickAPI{
		PeopleSvc: newPeopleService(db),
		GiftsSvc:  newGiftsService(db),
	}

	req := jsonRequest(http.MethodPost, "/v1/people/999/gifts/quick", `{"title":"Gift"}`)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": "999"}, h.QuickGift)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// ---- UpdateLastContact ------------------------------------------------------

func TestPeopleUpdateLastContact_HappyPath(t *testing.T) {
	db := openTestDB(t)
	personID := insertTestPerson(t, db, "Frank")
	h := &handler.PeopleQuickAPI{
		PeopleSvc: newPeopleService(db),
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/people/1/last-contact", nil)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.UpdateLastContact)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}
}

func TestPeopleUpdateLastContact_PersonNotFound_Returns404(t *testing.T) {
	db := openTestDB(t)
	h := &handler.PeopleQuickAPI{
		PeopleSvc: newPeopleService(db),
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/people/999/last-contact", nil)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": "999"}, h.UpdateLastContact)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestPeopleUpdateLastContact_InvalidID_Returns400(t *testing.T) {
	db := openTestDB(t)
	h := &handler.PeopleQuickAPI{
		PeopleSvc: newPeopleService(db),
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/people/bad/last-contact", nil)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": "bad"}, h.UpdateLastContact)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ---- SetFavorite / UnsetFavorite --------------------------------------------

func TestPeopleSetFavorite_HappyPath(t *testing.T) {
	db := openTestDB(t)
	personID := insertTestPerson(t, db, "Grace")
	h := &handler.PeopleQuickAPI{
		PeopleSvc: newPeopleService(db),
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/people/1/favorite", nil)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.SetFavorite)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"favorite":true`) {
		t.Fatalf("expected favorite:true in response, got: %s", rec.Body.String())
	}
}

func TestPeopleSetFavorite_PersonNotFound_Returns404(t *testing.T) {
	db := openTestDB(t)
	h := &handler.PeopleQuickAPI{
		PeopleSvc: newPeopleService(db),
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/people/999/favorite", nil)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": "999"}, h.SetFavorite)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestPeopleUnsetFavorite_HappyPath(t *testing.T) {
	db := openTestDB(t)
	personID := insertTestPerson(t, db, "Heidi")
	h := &handler.PeopleQuickAPI{
		PeopleSvc: newPeopleService(db),
	}

	setReq := httptest.NewRequest(http.MethodPost, "/v1/people/1/favorite", nil)
	execHandler(newTestEcho(), setReq, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.SetFavorite)

	req := httptest.NewRequest(http.MethodDelete, "/v1/people/1/favorite", nil)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.UnsetFavorite)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"favorite":false`) {
		t.Fatalf("expected favorite:false in response, got: %s", rec.Body.String())
	}
}

func TestPeopleUnsetFavorite_PersonNotFound_Returns404(t *testing.T) {
	db := openTestDB(t)
	h := &handler.PeopleQuickAPI{
		PeopleSvc: newPeopleService(db),
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/people/999/favorite", nil)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": "999"}, h.UnsetFavorite)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
