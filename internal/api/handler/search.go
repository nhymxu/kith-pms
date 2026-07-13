package handler

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/gifts"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/people"
)

// searchLimitPerType caps each result group; fixed (no query param) per spec — YAGNI.
const searchLimitPerType = 5

// maxSearchQueryLen bounds q so arbitrarily large inputs never reach the fan-out queries.
const maxSearchQueryLen = 128

// SearchAPI fans a query out to the existing people/journal/gifts List methods and
// assembles a grouped result envelope for the ⌘K command palette. It reuses each
// service's existing query mechanism (people LIKE, journal FTS5, gifts LIKE) rather
// than introducing a new search index.
type SearchAPI struct {
	PeopleSvc  *people.Service
	JournalSvc *journal.Service
	GiftsSvc   *gifts.Service
}

// SearchItem is a single grouped search result row.
// LOCKED shape — consumed verbatim by web/src/schemas/search.ts (Phase 11).
type SearchItem struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	URL      string `json:"url"`
}

// SearchResult groups hits by entity type.
type SearchResult struct {
	People  []SearchItem `json:"people"`
	Journal []SearchItem `json:"journal"`
	Gifts   []SearchItem `json:"gifts"`
}

// Search godoc
//
// @Summary      Global search
// @Description  Fans a query out to people/journal/gifts; each group capped at 5 hits. Blank q returns empty groups.
// @Tags         search
// @Produce      json
// @Param        q  query  string  false  "Search term"
// @Success      200  {object}  envelope
// @Failure      500  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /search [get]
func (h *SearchAPI) Search(c *echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	if utf8.RuneCountInString(q) > maxSearchQueryLen {
		q = string([]rune(q)[:maxSearchQueryLen])
	}

	result := SearchResult{
		People:  []SearchItem{},
		Journal: []SearchItem{},
		Gifts:   []SearchItem{},
	}

	if q == "" {
		return ok(c, result)
	}

	ctx := c.Request().Context()

	if h.PeopleSvc != nil {
		list, err := h.PeopleSvc.List(ctx, people.ListParams{Query: q, Page: 1, PageSize: searchLimitPerType})
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "internal server error")
		}

		for _, p := range list.Items {
			result.People = append(result.People, SearchItem{
				ID:       p.ID,
				Title:    p.Name,
				Subtitle: personSubtitle(p),
				URL:      fmt.Sprintf("/people/%d", p.ID),
			})
		}
	}

	if h.JournalSvc != nil {
		list, err := h.JournalSvc.List(ctx, journal.ListParams{Query: q, Page: 1, PageSize: searchLimitPerType})
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "internal server error")
		}

		for _, a := range list.Items {
			result.Journal = append(result.Journal, SearchItem{
				ID:       a.ID,
				Title:    a.Title,
				Subtitle: a.OccurredAtDate,
				URL:      fmt.Sprintf("/journal/%d", a.ID),
			})
		}
	}

	if h.GiftsSvc != nil {
		list, err := h.GiftsSvc.List(ctx, gifts.ListParams{Query: q, Page: 1, PageSize: searchLimitPerType})
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "internal server error")
		}

		for _, g := range list.Items {
			result.Gifts = append(result.Gifts, SearchItem{
				ID:       g.ID,
				Title:    g.Title,
				Subtitle: g.PersonName,
				URL:      fmt.Sprintf("/gifts/%d", g.ID),
			})
		}
	}

	return ok(c, result)
}

// personSubtitle returns the nickname when set, otherwise the last-contact date
// (YYYY-MM-DD), otherwise an empty string — per the Phase 10 DTO contract.
func personSubtitle(p people.Person) string {
	if p.Nickname != "" {
		return p.Nickname
	}

	if p.LastContactAt != nil {
		return p.LastContactAt.Format("2006-01-02")
	}

	return ""
}
