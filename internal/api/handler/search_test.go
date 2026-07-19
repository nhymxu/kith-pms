package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/api/handler"
	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/gifts"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/people"
)

// seedSearchFixtures inserts one person plus a linked journal entry and gift, all
// containing "alice" in a searchable field, and returns the person ID.
func seedSearchFixtures(t *testing.T, h *handler.SearchAPI) int64 {
	t.Helper()

	ctx := context.Background()

	// h.PeopleSvc is required by every fixture; callers always populate it.
	personID, err := h.PeopleSvc.Create(ctx, people.Person{Name: "Alice Wonderland"}, nil, nil)
	if err != nil {
		t.Fatalf("create person: %v", err)
	}

	if h.JournalSvc != nil {
		_, err = h.JournalSvc.Create(ctx, journal.Activity{
			Title:          "Coffee with Alice",
			OccurredAtDate: "2026-01-10",
			Content:        "Great catch-up conversation",
		}, []int64{personID}, nil)
		if err != nil {
			t.Fatalf("create activity: %v", err)
		}
	}

	if h.GiftsSvc != nil {
		_, err = h.GiftsSvc.Create(ctx, &gifts.Gift{
			PersonID:  personID,
			Title:     "Alice's birthday gift",
			Direction: gifts.DirectionGiven,
			Currency:  "USD",
		})
		if err != nil {
			t.Fatalf("create gift: %v", err)
		}
	}

	return personID
}

// ---- grouped shape / happy path ---------------------------------------------

func TestSearch_HappyPath_ReturnsGroupedResults(t *testing.T) {
	db := openTestDB(t)
	h := &handler.SearchAPI{
		PeopleSvc:  newPeopleService(db),
		JournalSvc: newJournalService(db),
		GiftsSvc:   newGiftsService(db),
	}
	seedSearchFixtures(t, h)

	req := jsonRequest(http.MethodGet, "/v1/search?q=alice", "")
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

	if len(envelope.Data.People) != 1 {
		t.Fatalf("expected 1 people hit, got %d", len(envelope.Data.People))
	}

	if envelope.Data.People[0].Title != "Alice Wonderland" {
		t.Fatalf("expected person title 'Alice Wonderland', got %q", envelope.Data.People[0].Title)
	}

	if envelope.Data.People[0].URL == "" || !strings.HasPrefix(envelope.Data.People[0].URL, "/people/") {
		t.Fatalf("expected /people/ URL, got %q", envelope.Data.People[0].URL)
	}

	if len(envelope.Data.Journal) != 1 {
		t.Fatalf("expected 1 journal hit, got %d", len(envelope.Data.Journal))
	}

	if envelope.Data.Journal[0].Subtitle != "2026-01-10" {
		t.Fatalf("expected journal subtitle to be the entry date, got %q", envelope.Data.Journal[0].Subtitle)
	}

	if len(envelope.Data.Gifts) != 1 {
		t.Fatalf("expected 1 gift hit, got %d", len(envelope.Data.Gifts))
	}

	if envelope.Data.Gifts[0].Subtitle != "Alice Wonderland" {
		t.Fatalf("expected gift subtitle to be the person name, got %q", envelope.Data.Gifts[0].Subtitle)
	}
}

func TestSearch_NoMatch_ReturnsEmptyGroups(t *testing.T) {
	db := openTestDB(t)
	h := &handler.SearchAPI{
		PeopleSvc:  newPeopleService(db),
		JournalSvc: newJournalService(db),
		GiftsSvc:   newGiftsService(db),
	}
	seedSearchFixtures(t, h)

	req := jsonRequest(http.MethodGet, "/v1/search?q=zzz-nonexistent", "")
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

	if len(envelope.Data.People) != 0 || len(envelope.Data.Journal) != 0 || len(envelope.Data.Gifts) != 0 {
		t.Fatalf("expected all-empty groups, got: %+v", envelope.Data)
	}
}

// ---- blank q short-circuit ---------------------------------------------------

func TestSearch_BlankQuery_ReturnsEmptyGroupsWithoutError(t *testing.T) {
	db := openTestDB(t)
	h := &handler.SearchAPI{
		PeopleSvc:  newPeopleService(db),
		JournalSvc: newJournalService(db),
		GiftsSvc:   newGiftsService(db),
	}
	seedSearchFixtures(t, h)

	for _, q := range []string{"", "   "} {
		req := jsonRequest(http.MethodGet, "/v1/search?q="+escapeQueryParam(q), "")
		rec := execHandler(newTestEcho(), req, nil, h.Search)

		if rec.Code != http.StatusOK {
			t.Fatalf("q=%q: expected 200, got %d — body: %s", q, rec.Code, rec.Body.String())
		}

		if !strings.Contains(rec.Body.String(), `"people":[]`) ||
			!strings.Contains(rec.Body.String(), `"journal":[]`) ||
			!strings.Contains(rec.Body.String(), `"gifts":[]`) {
			t.Fatalf("q=%q: expected empty groups, got: %s", q, rec.Body.String())
		}
	}
}

// ---- types param filtering ----------------------------------------------------

func TestSearch_TypesParam_FiltersToRequestedGroup(t *testing.T) {
	db := openTestDB(t)
	h := &handler.SearchAPI{
		PeopleSvc:  newPeopleService(db),
		JournalSvc: newJournalService(db),
		GiftsSvc:   newGiftsService(db),
	}
	seedSearchFixtures(t, h)

	req := jsonRequest(http.MethodGet, "/v1/search?q=alice&types=people", "")
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

	if len(envelope.Data.People) != 1 {
		t.Fatalf("expected 1 people hit, got %d", len(envelope.Data.People))
	}

	if len(envelope.Data.Journal) != 0 || len(envelope.Data.Gifts) != 0 {
		t.Fatalf("expected journal/gifts empty when types=people, got: %+v", envelope.Data)
	}
}

func TestSearch_TypesParam_MultipleGroups(t *testing.T) {
	db := openTestDB(t)
	h := &handler.SearchAPI{
		PeopleSvc:  newPeopleService(db),
		JournalSvc: newJournalService(db),
		GiftsSvc:   newGiftsService(db),
	}
	seedSearchFixtures(t, h)

	req := jsonRequest(http.MethodGet, "/v1/search?q=alice&types=people,gifts", "")
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

	if len(envelope.Data.People) != 1 || len(envelope.Data.Gifts) != 1 {
		t.Fatalf("expected people+gifts populated, got: %+v", envelope.Data)
	}

	if len(envelope.Data.Journal) != 0 {
		t.Fatalf("expected journal empty when types=people,gifts, got: %+v", envelope.Data)
	}
}

func TestSearch_TypesParam_UnknownTokenIgnoredNotError(t *testing.T) {
	db := openTestDB(t)
	h := &handler.SearchAPI{
		PeopleSvc:  newPeopleService(db),
		JournalSvc: newJournalService(db),
		GiftsSvc:   newGiftsService(db),
	}
	seedSearchFixtures(t, h)

	req := jsonRequest(http.MethodGet, "/v1/search?q=alice&types=bogus", "")
	rec := execHandler(newTestEcho(), req, nil, h.Search)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 (unknown token ignored, not an error), got %d — body: %s", rec.Code, rec.Body.String())
	}

	var envelope struct {
		Data handler.SearchResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(envelope.Data.People) != 0 || len(envelope.Data.Journal) != 0 || len(envelope.Data.Gifts) != 0 {
		t.Fatalf("expected all-empty groups for unknown token, got: %+v", envelope.Data)
	}
}

// ---- injection safety --------------------------------------------------------

func TestSearch_MaliciousQuery_IsSafelyParameterized(t *testing.T) {
	db := openTestDB(t)
	h := &handler.SearchAPI{
		PeopleSvc:  newPeopleService(db),
		JournalSvc: newJournalService(db),
		GiftsSvc:   newGiftsService(db),
	}
	seedSearchFixtures(t, h)

	malicious := []string{
		`"`,
		`'`,
		`%`,
		`_`,
		`'; DROP TABLE person; --`,
		`" OR "1"="1`,
		`\%\_\\`,
	}

	for _, q := range malicious {
		req := jsonRequest(http.MethodGet, "/v1/search?q="+escapeQueryParam(q), "")
		rec := execHandler(newTestEcho(), req, nil, h.Search)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"q=%q: expected 200 (safe, parameterized query), got %d — body: %s",
				q,
				rec.Code,
				rec.Body.String(),
			)
		}
	}

	// Verify the DROP TABLE payload never executed — the person table (and seed row) survive.
	var count int
	if err := db.NewSelect().Table("person").ColumnExpr("COUNT(*)").Scan(context.Background(), &count); err != nil {
		t.Fatalf("person table should still exist and be queryable: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 person row to survive injection attempts, got %d", count)
	}
}

// escapeQueryParam minimally encodes a raw string for safe inclusion in a test request URL.
func escapeQueryParam(s string) string {
	r := strings.NewReplacer(
		"%", "%25",
		"&", "%26",
		"#", "%23",
		" ", "%20",
		"\"", "%22",
		"'", "%27",
		";", "%3B",
		"\\", "%5C",
	)

	return r.Replace(s)
}

// ---- auth (middleware smoke, matching TestSessionOrBearer_AuthFailure_Returns401JSON) ----

func TestSearch_Unauthenticated_Returns401ViaMiddleware(t *testing.T) {
	db := openTestDB(t)
	svc := newTestAuthSvc(t, db, "pw")

	e := newTestEcho()
	req := jsonRequest(http.MethodGet, "/v1/search?q=alice", "")
	// No Bearer, no cookie — SessionOrBearer must reject before the handler runs.
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
