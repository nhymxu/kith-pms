package handlers_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	_ "modernc.org/sqlite"

	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/web/handlers"
)

// openTestDB opens an in-memory SQLite database and runs all migrations.
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

// insertPerson inserts a person and returns ID.
func insertPerson(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()
	res, err := db.Exec(`INSERT INTO person (name) VALUES (?)`, name)
	if err != nil {
		t.Fatalf("insert person: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

// insertActivity inserts an activity and returns ID.
func insertActivity(t *testing.T, db *sql.DB, title, date string) int64 {
	t.Helper()
	res, err := db.Exec(`INSERT INTO activity (title, occurred_at_date) VALUES (?, ?)`, title, date)
	if err != nil {
		t.Fatalf("insert activity: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

// linkActivityPerson links an activity to a person.
func linkActivityPerson(t *testing.T, db *sql.DB, activityID, personID int64) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO activity_person (activity_id, person_id) VALUES (?, ?)`, activityID, personID)
	if err != nil {
		t.Fatalf("link activity person: %v", err)
	}
}

// TestGetList_PeopleFilter_Single verifies filtering by a single person.
func TestGetList_PeopleFilter_Single(t *testing.T) {
	db := openTestDB(t)
	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")

	a1 := insertActivity(t, db, "Meeting with Alice", "2024-01-01")
	linkActivityPerson(t, db, a1, alice)

	a2 := insertActivity(t, db, "Meeting with Bob", "2024-01-02")
	linkActivityPerson(t, db, a2, bob)

	h := &handlers.JournalHandlers{
		Svc:       journal.NewService(db),
		PeopleSvc: people.NewService(db),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/journal?people=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetList(c); err != nil {
		t.Fatalf("GetList: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Meeting with Alice") {
		t.Errorf("expected Alice's meeting in response")
	}
	if strings.Contains(body, "Meeting with Bob") {
		t.Errorf("unexpected Bob's meeting in response")
	}
}

// TestGetList_PeopleFilter_Multiple verifies OR semantics with multiple people.
func TestGetList_PeopleFilter_Multiple(t *testing.T) {
	db := openTestDB(t)
	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")
	charlie := insertPerson(t, db, "Charlie")

	a1 := insertActivity(t, db, "Alice solo", "2024-01-01")
	linkActivityPerson(t, db, a1, alice)

	a2 := insertActivity(t, db, "Bob solo", "2024-01-02")
	linkActivityPerson(t, db, a2, bob)

	a3 := insertActivity(t, db, "Charlie solo", "2024-01-03")
	linkActivityPerson(t, db, a3, charlie)

	h := &handlers.JournalHandlers{
		Svc:       journal.NewService(db),
		PeopleSvc: people.NewService(db),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/journal?people=1,2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetList(c); err != nil {
		t.Fatalf("GetList: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Alice solo") {
		t.Errorf("expected Alice's entry")
	}
	if !strings.Contains(body, "Bob solo") {
		t.Errorf("expected Bob's entry")
	}
	if strings.Contains(body, "Charlie solo") {
		t.Errorf("unexpected Charlie's entry")
	}
}

// TestGetList_PeopleFilter_Empty verifies empty filter shows all entries.
func TestGetList_PeopleFilter_Empty(t *testing.T) {
	db := openTestDB(t)
	alice := insertPerson(t, db, "Alice")

	a1 := insertActivity(t, db, "Entry 1", "2024-01-01")
	linkActivityPerson(t, db, a1, alice)

	_ = insertActivity(t, db, "Entry 2", "2024-01-02")

	h := &handlers.JournalHandlers{
		Svc:       journal.NewService(db),
		PeopleSvc: people.NewService(db),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/journal?people=", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetList(c); err != nil {
		t.Fatalf("GetList: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Entry 1") {
		t.Errorf("expected Entry 1")
	}
	if !strings.Contains(body, "Entry 2") {
		t.Errorf("expected Entry 2")
	}
}

// TestGetList_PeopleFilter_Invalid verifies invalid IDs are ignored.
func TestGetList_PeopleFilter_Invalid(t *testing.T) {
	db := openTestDB(t)
	alice := insertPerson(t, db, "Alice")

	a1 := insertActivity(t, db, "Entry 1", "2024-01-01")
	linkActivityPerson(t, db, a1, alice)

	h := &handlers.JournalHandlers{
		Svc:       journal.NewService(db),
		PeopleSvc: people.NewService(db),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/journal?people=abc,999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetList(c); err != nil {
		t.Fatalf("GetList: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

// TestGetList_PeopleFilter_WithQuery verifies people filter combines with text search.
func TestGetList_PeopleFilter_WithQuery(t *testing.T) {
	db := openTestDB(t)
	alice := insertPerson(t, db, "Alice")

	a1 := insertActivity(t, db, "Meeting about project", "2024-01-01")
	linkActivityPerson(t, db, a1, alice)
	db.Exec(`INSERT INTO activity_fts (rowid, title, content) VALUES (?, ?, ?)`, a1, "Meeting about project", "")

	a2 := insertActivity(t, db, "Lunch with Alice", "2024-01-02")
	linkActivityPerson(t, db, a2, alice)
	db.Exec(`INSERT INTO activity_fts (rowid, title, content) VALUES (?, ?, ?)`, a2, "Lunch with Alice", "")

	h := &handlers.JournalHandlers{
		Svc:       journal.NewService(db),
		PeopleSvc: people.NewService(db),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/journal?people=1&q=project", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetList(c); err != nil {
		t.Fatalf("GetList: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Meeting about project") {
		t.Errorf("expected meeting entry")
	}
	if strings.Contains(body, "Lunch with Alice") {
		t.Errorf("unexpected lunch entry (no 'project' in title)")
	}
}

// TestGetList_PeopleFilter_WithDates verifies people filter combines with date range.
func TestGetList_PeopleFilter_WithDates(t *testing.T) {
	db := openTestDB(t)
	alice := insertPerson(t, db, "Alice")

	a1 := insertActivity(t, db, "January meeting", "2024-01-15")
	linkActivityPerson(t, db, a1, alice)

	a2 := insertActivity(t, db, "February meeting", "2024-02-15")
	linkActivityPerson(t, db, a2, alice)

	h := &handlers.JournalHandlers{
		Svc:       journal.NewService(db),
		PeopleSvc: people.NewService(db),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/journal?people=1&from=2024-01-01&to=2024-01-31", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetList(c); err != nil {
		t.Fatalf("GetList: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "January meeting") {
		t.Errorf("expected January meeting")
	}
	if strings.Contains(body, "February meeting") {
		t.Errorf("unexpected February meeting (outside date range)")
	}
}

// TestGetList_PeopleFilter_Pagination verifies people param preserved in pagination links.
func TestGetList_PeopleFilter_Pagination(t *testing.T) {
	db := openTestDB(t)
	alice := insertPerson(t, db, "Alice")

	for i := 1; i <= 35; i++ {
		aid := insertActivity(t, db, "Entry", "2024-01-01")
		linkActivityPerson(t, db, aid, alice)
	}

	h := &handlers.JournalHandlers{
		Svc:       journal.NewService(db),
		PeopleSvc: people.NewService(db),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/journal?people=1&page=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetList(c); err != nil {
		t.Fatalf("GetList: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "people=1") {
		t.Errorf("expected people=1 in pagination links")
	}
	if !strings.Contains(body, "Next") {
		t.Errorf("expected Next link (35 entries > 30 page size)")
	}
}

// TestGetList_PersonFilterArray verifies person_filter[] form submission.
func TestGetList_PersonFilterArray(t *testing.T) {
	db := openTestDB(t)
	alice := insertPerson(t, db, "Alice")
	bob := insertPerson(t, db, "Bob")

	a1 := insertActivity(t, db, "Alice entry", "2024-01-01")
	linkActivityPerson(t, db, a1, alice)

	a2 := insertActivity(t, db, "Bob entry", "2024-01-02")
	linkActivityPerson(t, db, a2, bob)

	h := &handlers.JournalHandlers{
		Svc:       journal.NewService(db),
		PeopleSvc: people.NewService(db),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/journal?person_filter[]=1&person_filter[]=2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetList(c); err != nil {
		t.Fatalf("GetList: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Alice entry") {
		t.Errorf("expected Alice entry")
	}
	if !strings.Contains(body, "Bob entry") {
		t.Errorf("expected Bob entry")
	}
}
