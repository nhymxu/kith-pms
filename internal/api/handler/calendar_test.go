package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/api/handler"
	"github.com/nhymxu/kith-pms/internal/calendar"
	"github.com/nhymxu/kith-pms/internal/reminders"
	"github.com/nhymxu/kith-pms/internal/settings"
)

func TestCalendarFeedRequiresToken(t *testing.T) {
	db := openTestDB(t)
	h := &handler.CalendarAPI{
		Tokens:  &calendar.TokenStore{Repo: settings.NewRepo(db)},
		Builder: &calendar.Builder{DB: db, Reminders: reminders.NewRepo(db)},
	}

	e := echo.New()
	e.GET("/calendar/:token/feed.ics", h.Feed("kith", calendar.AllParts, true))

	tok, err := h.Tokens.Get(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	get := func(token string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/calendar/"+token+"/feed.ics", nil))

		return rec
	}

	if rec := get("wrong"); rec.Code != http.StatusNotFound {
		t.Fatalf("bad token status = %d, want 404", rec.Code)
	}

	rec := get(tok)
	if rec.Code != http.StatusOK {
		t.Fatalf("good token status = %d, want 200", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "text/calendar; charset=utf-8" {
		t.Fatalf("content-type = %q", ct)
	}

	if !strings.HasPrefix(rec.Body.String(), "BEGIN:VCALENDAR\r\n") {
		t.Fatalf("body not an iCalendar: %q", rec.Body.String())
	}

	if _, err := h.Tokens.Regenerate(t.Context()); err != nil {
		t.Fatal(err)
	}

	if rec := get(tok); rec.Code != http.StatusNotFound {
		t.Fatalf("old token after regenerate = %d, want 404", rec.Code)
	}
}
