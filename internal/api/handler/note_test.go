package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/api/handler"
	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/note"
)

func TestNoteCreate_HappyPath_Returns201(t *testing.T) {
	db := openTestDB(t)
	svc := newNoteService(db)
	personID := insertTestPerson(t, db, "Alice")

	h := &handler.NoteAPI{Svc: svc}
	req := jsonRequest(http.MethodPost, fmt.Sprintf("/v1/people/%d/notes", personID),
		`{"title":"Ideas","content":"Buy a gift"}`)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.Create)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", rec.Code, rec.Body.String())
	}

	var envelope struct {
		Data note.Note `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if envelope.Data.Title != "Ideas" || envelope.Data.Content != "Buy a gift" {
		t.Fatalf("unexpected note: %+v", envelope.Data)
	}

	if envelope.Data.PersonID != personID {
		t.Fatalf("PersonID = %d, want %d", envelope.Data.PersonID, personID)
	}
}

func TestNoteCreate_MissingContent_Returns400(t *testing.T) {
	db := openTestDB(t)
	h := &handler.NoteAPI{Svc: newNoteService(db)}
	personID := insertTestPerson(t, db, "Bob")

	req := jsonRequest(http.MethodPost, fmt.Sprintf("/v1/people/%d/notes", personID), `{"title":"No content"}`)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.Create)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d — body: %s", rec.Code, rec.Body.String())
	}
}

func TestNoteCreate_TitleOptional(t *testing.T) {
	db := openTestDB(t)
	h := &handler.NoteAPI{Svc: newNoteService(db)}
	personID := insertTestPerson(t, db, "Carol")

	req := jsonRequest(http.MethodPost, fmt.Sprintf("/v1/people/%d/notes", personID), `{"content":"Just a jot"}`)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.Create)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", rec.Code, rec.Body.String())
	}
}

func TestNoteListByPerson_ReturnsCreatedNote(t *testing.T) {
	db := openTestDB(t)
	svc := newNoteService(db)
	personID := insertTestPerson(t, db, "Dave")

	if _, err := svc.Create(
		context.Background(),
		&note.Note{PersonID: personID, Title: "T", Content: "C"},
	); err != nil {
		t.Fatalf("seed create: %v", err)
	}

	h := &handler.NoteAPI{Svc: svc}
	req := jsonRequest(http.MethodGet, fmt.Sprintf("/v1/people/%d/notes", personID), "")
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", personID)}, h.ListByPerson)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}

	var envelope struct {
		Data note.List `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if envelope.Data.Total != 1 || len(envelope.Data.Items) != 1 {
		t.Fatalf("expected 1 note, got total=%d items=%d", envelope.Data.Total, len(envelope.Data.Items))
	}
}

func TestNoteGet_UnknownID_Returns404(t *testing.T) {
	db := openTestDB(t)
	h := &handler.NoteAPI{Svc: newNoteService(db)}

	req := jsonRequest(http.MethodGet, "/v1/notes/999", "")
	rec := execHandler(newTestEcho(), req, map[string]string{"id": "999"}, h.Get)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestNoteUpdate_HappyPath(t *testing.T) {
	db := openTestDB(t)
	svc := newNoteService(db)
	personID := insertTestPerson(t, db, "Eve")

	id, err := svc.Create(context.Background(), &note.Note{PersonID: personID, Title: "Old", Content: "Old content"})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	h := &handler.NoteAPI{Svc: svc}
	req := jsonRequest(http.MethodPut, fmt.Sprintf("/v1/notes/%d", id), `{"title":"New","content":"New content"}`)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", id)}, h.Update)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}

	var envelope struct {
		Data note.Note `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if envelope.Data.Title != "New" || envelope.Data.Content != "New content" {
		t.Fatalf("unexpected note after update: %+v", envelope.Data)
	}
}

func TestNoteUpdate_UnknownID_Returns404(t *testing.T) {
	db := openTestDB(t)
	h := &handler.NoteAPI{Svc: newNoteService(db)}

	req := jsonRequest(http.MethodPut, "/v1/notes/999", `{"content":"x"}`)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": "999"}, h.Update)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestNoteDelete_HappyPath(t *testing.T) {
	db := openTestDB(t)
	svc := newNoteService(db)
	personID := insertTestPerson(t, db, "Frank")

	id, err := svc.Create(context.Background(), &note.Note{PersonID: personID, Content: "C"})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	h := &handler.NoteAPI{Svc: svc}
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/notes/%d", id), nil)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": fmt.Sprintf("%d", id)}, h.Delete)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d — body: %s", rec.Code, rec.Body.String())
	}

	getReq := jsonRequest(http.MethodGet, fmt.Sprintf("/v1/notes/%d", id), "")
	getRec := execHandler(newTestEcho(), getReq, map[string]string{"id": fmt.Sprintf("%d", id)}, h.Get)

	if getRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", getRec.Code)
	}
}

func TestNoteDelete_UnknownID_Returns404(t *testing.T) {
	db := openTestDB(t)
	h := &handler.NoteAPI{Svc: newNoteService(db)}

	req := httptest.NewRequest(http.MethodDelete, "/v1/notes/999", nil)
	rec := execHandler(newTestEcho(), req, map[string]string{"id": "999"}, h.Delete)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestNote_Unauthenticated_Returns401ViaMiddleware(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "pw")

	e := newTestEcho()
	req := jsonRequest(http.MethodGet, "/v1/people/1/notes", "")
	rec := execHandler(e, req, nil, func(c *echo.Context) error {
		mw := auth.SessionOrBearer(testAPIToken, svc)
		return mw(func(_ *echo.Context) error { return nil })(c)
	})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"error"`) {
		t.Fatalf("expected JSON error body, got: %s", rec.Body.String())
	}
}

// ---- search integration ------------------------------------------------------

func TestSearch_ReturnsNoteHit(t *testing.T) {
	db := openTestDB(t)
	peopleSvc := newPeopleService(db)
	noteSvc := newNoteService(db)

	h := &handler.SearchAPI{
		PeopleSvc: peopleSvc,
		NoteSvc:   noteSvc,
	}

	personID := insertTestPerson(t, db, "Grace")

	_, err := noteSvc.Create(context.Background(), &note.Note{
		PersonID: personID,
		Title:    "Searchable note title",
		Content:  "unrelated content",
	})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}

	req := jsonRequest(http.MethodGet, "/v1/search?q=searchable", "")
	rec := execHandler(newTestEcho(), req, nil, h.Search)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}

	var envelope struct {
		Data handler.SearchResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(envelope.Data.Notes) != 1 {
		t.Fatalf("expected 1 note hit, got %d: %+v", len(envelope.Data.Notes), envelope.Data.Notes)
	}

	if envelope.Data.Notes[0].Title != "Searchable note title" {
		t.Fatalf("expected note title, got %q", envelope.Data.Notes[0].Title)
	}

	if envelope.Data.Notes[0].Subtitle != "Grace" {
		t.Fatalf("expected subtitle to be owner name, got %q", envelope.Data.Notes[0].Subtitle)
	}

	if !strings.HasPrefix(envelope.Data.Notes[0].URL, "/people/") {
		t.Fatalf("expected /people/ URL for non-self owner, got %q", envelope.Data.Notes[0].URL)
	}
}

func TestSearch_NoteHit_SelfOwner_UsesNotesURL(t *testing.T) {
	db := openTestDB(t)
	peopleSvc := newPeopleService(db)
	noteSvc := newNoteService(db)

	h := &handler.SearchAPI{
		PeopleSvc: peopleSvc,
		NoteSvc:   noteSvc,
	}

	selfID := insertTestPerson(t, db, "Me")
	if err := peopleSvc.SetSelf(context.Background(), selfID); err != nil {
		t.Fatalf("SetSelf: %v", err)
	}

	if _, err := noteSvc.Create(context.Background(), &note.Note{
		PersonID: selfID,
		Title:    "Self searchable note",
		Content:  "content",
	}); err != nil {
		t.Fatalf("create note: %v", err)
	}

	req := jsonRequest(http.MethodGet, "/v1/search?q=searchable", "")
	rec := execHandler(newTestEcho(), req, nil, h.Search)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rec.Code, rec.Body.String())
	}

	var envelope struct {
		Data handler.SearchResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(envelope.Data.Notes) != 1 {
		t.Fatalf("expected 1 note hit, got %d", len(envelope.Data.Notes))
	}

	if envelope.Data.Notes[0].URL != "/notes" {
		t.Fatalf("expected /notes URL for self owner, got %q", envelope.Data.Notes[0].URL)
	}
}
