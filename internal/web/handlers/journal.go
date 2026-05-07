package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/web/templates"
)

// JournalHandlers groups all /journal/* HTTP handlers.
type JournalHandlers struct {
	Svc       *journal.Service
	PeopleSvc *people.Service
}

// GetList handles GET /journal
func (h *JournalHandlers) GetList(c *echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	from := strings.TrimSpace(c.QueryParam("from"))
	to := strings.TrimSpace(c.QueryParam("to"))

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	// Parse ?people=1,2 — comma-separated person IDs.
	personIDs := parsePersonIDs(c.QueryParam("people"))
	if filterPersonIDs := parsePersonIDSlice(c.Request().URL.Query()["person_filter[]"]); len(filterPersonIDs) > 0 {
		personIDs = filterPersonIDs
	}

	list, err := h.Svc.List(c.Request().Context(), journal.ListParams{
		Query:     q,
		PersonIDs: personIDs,
		FromDate:  from,
		ToDate:    to,
		Page:      page,
		PageSize:  30,
	})
	if err != nil {
		return err
	}

	hasMore := len(list) == 30

	var selfPersonID int64

	if h.PeopleSvc != nil {
		if selfPerson, err := h.PeopleSvc.GetSelf(c.Request().Context()); err == nil && selfPerson != nil {
			selfPersonID = selfPerson.ID
		}
	}

	// Populate person names map for filter chips
	personNames := make(map[int64]string)

	if h.PeopleSvc != nil {
		for _, pid := range personIDs {
			if p, err := h.PeopleSvc.Get(c.Request().Context(), pid); err == nil && p != nil {
				personNames[pid] = p.Name
			}
		}
	}

	component := templates.JournalList(templates.JournalListParams{
		Activities:   list,
		Query:        q,
		FromDate:     from,
		ToDate:       to,
		Page:         page,
		HasMore:      hasMore,
		PersonIDs:    personIDs,
		PersonNames:  personNames,
		SelfPersonID: selfPersonID,
	})

	return component.Render(c.Request().Context(), c.Response())
}

// GetNew handles GET /journal/new
func (h *JournalHandlers) GetNew(c *echo.Context) error {
	var preselected []journal.ActivityPerson

	// Optional ?person=ID to pre-select a person from the person detail link.
	if pidStr := c.QueryParam("person"); pidStr != "" {
		if pid, err := strconv.ParseInt(pidStr, 10, 64); err == nil && pid > 0 && h.PeopleSvc != nil {
			if p, err := h.PeopleSvc.Get(c.Request().Context(), pid); err == nil && p != nil {
				preselected = append(preselected, journal.ActivityPerson{
					PersonID: p.ID,
					Name:     p.Name,
				})
			}
		}
	}

	component := templates.JournalForm(templates.JournalFormParams{
		Activity:  journal.Activity{People: preselected},
		CSRFToken: auth.CSRFToken(c),
	})

	return component.Render(c.Request().Context(), c.Response())
}

// PostCreate handles POST /journal
func (h *JournalHandlers) PostCreate(c *echo.Context) error {
	a, personIDs, formErr := parseJournalForm(c)
	if formErr != "" {
		component := templates.JournalForm(templates.JournalFormParams{
			Activity:  a,
			CSRFToken: auth.CSRFToken(c),
			Error:     formErr,
		})

		return component.Render(c.Request().Context(), c.Response())
	}

	id, err := h.Svc.Create(c.Request().Context(), a, personIDs)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/journal/"+strconv.FormatInt(id, 10))
}

// GetDetail handles GET /journal/:id
func (h *JournalHandlers) GetDetail(c *echo.Context) error {
	id, err := parseJournalID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	a, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}

	if a == nil {
		return echo.ErrNotFound
	}

	component := templates.JournalDetail(templates.JournalDetailParams{
		Activity:  *a,
		CSRFToken: auth.CSRFToken(c),
	})

	return component.Render(c.Request().Context(), c.Response())
}

// GetEdit handles GET /journal/:id/edit
func (h *JournalHandlers) GetEdit(c *echo.Context) error {
	id, err := parseJournalID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	a, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}

	if a == nil {
		return echo.ErrNotFound
	}

	component := templates.JournalForm(templates.JournalFormParams{
		Activity:  *a,
		CSRFToken: auth.CSRFToken(c),
		IsEdit:    true,
	})

	return component.Render(c.Request().Context(), c.Response())
}

// PostUpdate handles POST /journal/:id
func (h *JournalHandlers) PostUpdate(c *echo.Context) error {
	id, err := parseJournalID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	a, personIDs, formErr := parseJournalForm(c)
	a.ID = id

	if formErr != "" {
		component := templates.JournalForm(templates.JournalFormParams{
			Activity:  a,
			CSRFToken: auth.CSRFToken(c),
			IsEdit:    true,
			Error:     formErr,
		})

		return component.Render(c.Request().Context(), c.Response())
	}

	if err := h.Svc.Update(c.Request().Context(), a, personIDs); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/journal/"+strconv.FormatInt(id, 10))
}

// PostDelete handles POST /journal/:id/delete
func (h *JournalHandlers) PostDelete(c *echo.Context) error {
	id, err := parseJournalID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/journal")
}

// GetDeleteConfirm handles GET /journal/:id/delete-confirm (htmx fragment)
func (h *JournalHandlers) GetDeleteConfirm(c *echo.Context) error {
	id, err := parseJournalID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	a, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}

	if a == nil {
		return echo.ErrNotFound
	}

	component := templates.JournalDeleteConfirm(*a, auth.CSRFToken(c))

	return component.Render(c.Request().Context(), c.Response())
}

// GetPeopleSearch handles GET /journal/people-search?q= (htmx fragment)
func (h *JournalHandlers) GetPeopleSearch(c *echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	if q == "" {
		component := templates.PersonPickerResults(nil)
		return component.Render(c.Request().Context(), c.Response())
	}

	list, err := h.PeopleSvc.List(c.Request().Context(), people.ListParams{
		Query:    q,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		return err
	}

	if c.QueryParam("mode") == "filter" {
		component := templates.PersonFilterPickerResults(list)
		return component.Render(c.Request().Context(), c.Response())
	}

	component := templates.PersonPickerResults(list)

	return component.Render(c.Request().Context(), c.Response())
}

// ---- form parsing -----------------------------------------------------------

// parseJournalForm reads journal form values and returns domain struct, person IDs,
// and a non-empty error string when validation fails.
func parseJournalForm(c *echo.Context) (journal.Activity, []int64, string) {
	title := strings.TrimSpace(c.FormValue("title"))
	if title == "" {
		return journal.Activity{}, nil, "Title is required."
	}

	date := strings.TrimSpace(c.FormValue("occurred_at_date"))
	if date == "" {
		return journal.Activity{}, nil, "Date is required."
	}

	a := journal.Activity{
		Title:          title,
		OccurredAtDate: date,
		OccurredAtTime: strings.TrimSpace(c.FormValue("occurred_at_time")),
		Content:        strings.TrimSpace(c.FormValue("content")),
	}

	// Parse person_id[] multi-value hidden inputs.
	personIDs := parsePersonIDSlice(c.Request().Form["person_id[]"])

	return a, personIDs, ""
}

// ---- helpers ----------------------------------------------------------------

func parseJournalID(c *echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

// parsePersonIDs parses a comma-separated string of int64 person IDs.
func parsePersonIDs(s string) []int64 {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")

	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if id, err := strconv.ParseInt(p, 10, 64); err == nil && id > 0 {
			out = append(out, id)
		}
	}

	return out
}

// parsePersonIDSlice parses a slice of string person IDs (from form multi-value).
func parsePersonIDSlice(vals []string) []int64 {
	out := make([]int64, 0, len(vals))

	seen := make(map[int64]bool)
	for _, v := range vals {
		if id, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil && id > 0 && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}

	return out
}
