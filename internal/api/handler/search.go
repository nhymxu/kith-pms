package handler

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/gifts"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/note"
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
	NoteSvc    *note.Service
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
	Notes   []SearchItem `json:"notes"`
}

// Search godoc
//
// @Summary      Global search
// @Description  Fans a query out to people/journal/gifts; each group capped at 5 hits. Blank q returns empty groups.
// @Tags         search
// @Produce      json
// @Param        q      query  string  false  "Search term"
// @Param        types  query  string  false  "Comma-separated group filter (people,journal,gifts,notes); absent/empty = all"
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

	want := parseSearchTypes(c.QueryParam("types"))

	result := SearchResult{
		People:  []SearchItem{},
		Journal: []SearchItem{},
		Gifts:   []SearchItem{},
		Notes:   []SearchItem{},
	}

	if q == "" {
		return ok(c, result)
	}

	ctx := c.Request().Context()

	if h.PeopleSvc != nil && want["people"] {
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

	if h.JournalSvc != nil && want["journal"] {
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

	if h.GiftsSvc != nil && want["gifts"] {
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

	if h.NoteSvc != nil && want["notes"] {
		var selfID int64

		if h.PeopleSvc != nil {
			if self, err := h.PeopleSvc.GetSelf(ctx); err == nil && self != nil {
				selfID = self.ID
			}
		}

		hits, err := h.NoteSvc.Search(ctx, q, searchLimitPerType)
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "internal server error")
		}

		for _, n := range hits {
			url := fmt.Sprintf("/people/%d", n.PersonID)
			if selfID != 0 && n.PersonID == selfID {
				url = "/notes"
			}

			result.Notes = append(result.Notes, SearchItem{
				ID:       n.ID,
				Title:    noteSearchTitle(n.Title, n.Content),
				Subtitle: n.PersonName,
				URL:      url,
			})
		}
	}

	return ok(c, result)
}

// searchGroups enumerates the SearchResult group keys, the sole source of truth
// for what a "types" token may reference — keep in sync with SearchResult's JSON tags.
var searchGroups = []string{"people", "journal", "gifts", "notes"}

// parseSearchTypes builds the enabled-group set from the "types" query param.
// Absent/empty input enables every group (back-compat); unknown tokens are
// silently ignored rather than rejected — the endpoint stays a pure function
// of its inputs and never errors on this param.
func parseSearchTypes(raw string) map[string]bool {
	want := make(map[string]bool, len(searchGroups))

	raw = strings.TrimSpace(raw)
	if raw == "" {
		for _, g := range searchGroups {
			want[g] = true
		}

		return want
	}

	for _, tok := range strings.Split(raw, ",") {
		tok = strings.ToLower(strings.TrimSpace(tok))
		for _, g := range searchGroups {
			if tok == g {
				want[g] = true
			}
		}
	}

	return want
}

// noteSearchTitle returns the note title, falling back to a truncated content
// snippet — notes, unlike journal entries, allow an empty title.
func noteSearchTitle(title, content string) string {
	if title != "" {
		return title
	}

	content = strings.TrimSpace(content)
	if utf8.RuneCountInString(content) <= 60 {
		return content
	}

	return string([]rune(content)[:60]) + "…"
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
