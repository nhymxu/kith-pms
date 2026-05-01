package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/dates"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/labels"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/web/forms"
	"github.com/nhymxu/kith-pms/internal/web/templates"
)

// PeopleHandlers groups all /people/* HTTP handlers.
type PeopleHandlers struct {
	Svc        *people.Service
	LabelsSvc  *labels.Service
	JournalSvc *journal.Service
	DatesSvc   *dates.Service
}

// GetList handles GET /people
func (h *PeopleHandlers) GetList(c *echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	// Parse ?labels=1,3 — comma-separated label IDs for AND-filter.
	labelIDs := parseLabelIDs(c.QueryParam("labels"))

	list, err := h.Svc.List(c.Request().Context(), people.ListParams{
		Query:    q,
		Page:     page,
		PageSize: 50,
		LabelIDs: labelIDs,
	})
	if err != nil {
		return err
	}

	hasMore := len(list) == 50

	// Load all labels for the filter pill row (best-effort; nil on error is fine).
	var allLabels []labels.Label
	if h.LabelsSvc != nil {
		allLabels, _ = h.LabelsSvc.List(c.Request().Context())
	}

	component := templates.PeopleList(templates.PeopleListParams{
		People:       list,
		Query:        q,
		Page:         page,
		HasMore:      hasMore,
		AllLabels:    allLabels,
		ActiveLabels: labelIDs,
	})
	return component.Render(c.Request().Context(), c.Response())
}

// GetNew handles GET /people/new
func (h *PeopleHandlers) GetNew(c *echo.Context) error {
	component := templates.PeopleForm(templates.PeopleFormParams{
		CSRFToken: auth.CSRFToken(c),
		Dates:     []dates.ImportantDate{},
	})
	return component.Render(c.Request().Context(), c.Response())
}

// PostCreate handles POST /people
func (h *PeopleHandlers) PostCreate(c *echo.Context) error {
	p, contacts, locations, importantDates, formErr := parsePersonForm(c)
	if formErr != "" {
		component := templates.PeopleForm(templates.PeopleFormParams{
			Person:    p,
			Contacts:  contacts,
			Locations: locations,
			Dates:     importantDates,
			CSRFToken: auth.CSRFToken(c),
			Error:     formErr,
		})
		return component.Render(c.Request().Context(), c.Response())
	}

	id, err := h.Svc.Create(c.Request().Context(), p, contacts, locations)
	if err != nil {
		return err
	}

	// Save dates in separate transaction
	if h.DatesSvc != nil && len(importantDates) > 0 {
		_ = h.DatesSvc.ReplaceForPerson(c.Request().Context(), id, importantDates)
	}

	return c.Redirect(http.StatusSeeOther, "/people/"+strconv.FormatInt(id, 10))
}

// GetDetail handles GET /people/:id
func (h *PeopleHandlers) GetDetail(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	p, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	if p == nil {
		return echo.ErrNotFound
	}

	var attached, allLabels []labels.Label
	if h.LabelsSvc != nil {
		attached, _ = h.LabelsSvc.ListByPersonID(c.Request().Context(), id)
		allLabels, _ = h.LabelsSvc.List(c.Request().Context())
	}

	// Fetch latest 10 activities involving this person (best-effort).
	var recentActivities []journal.Activity
	if h.JournalSvc != nil {
		recentActivities, _ = h.JournalSvc.List(c.Request().Context(), journal.ListParams{
			PersonIDs: []int64{id},
			PageSize:  10,
		})
	}

	// Fetch important dates
	var importantDates []dates.ImportantDate
	if h.DatesSvc != nil {
		importantDates, _ = h.DatesSvc.ListByPerson(c.Request().Context(), id)
	}

	component := templates.PeopleDetail(templates.PeopleDetailParams{
		Person:           *p,
		Labels:           attached,
		AllLabels:        allLabels,
		CSRFToken:        auth.CSRFToken(c),
		RecentActivities: recentActivities,
		Dates:            importantDates,
	})
	return component.Render(c.Request().Context(), c.Response())
}

// GetEdit handles GET /people/:id/edit
func (h *PeopleHandlers) GetEdit(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	p, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	if p == nil {
		return echo.ErrNotFound
	}

	var importantDates []dates.ImportantDate
	if h.DatesSvc != nil {
		importantDates, _ = h.DatesSvc.ListByPerson(c.Request().Context(), id)
	}

	component := templates.PeopleForm(templates.PeopleFormParams{
		Person:    *p,
		Contacts:  p.Contacts,
		Locations: p.Locations,
		Dates:     importantDates,
		CSRFToken: auth.CSRFToken(c),
		IsEdit:    true,
	})
	return component.Render(c.Request().Context(), c.Response())
}

// PostUpdate handles POST /people/:id
func (h *PeopleHandlers) PostUpdate(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	p, contacts, locations, importantDates, formErr := parsePersonForm(c)
	p.ID = id

	if formErr != "" {
		component := templates.PeopleForm(templates.PeopleFormParams{
			Person:    p,
			Contacts:  contacts,
			Locations: locations,
			Dates:     importantDates,
			CSRFToken: auth.CSRFToken(c),
			IsEdit:    true,
			Error:     formErr,
		})
		return component.Render(c.Request().Context(), c.Response())
	}

	if err := h.Svc.Update(c.Request().Context(), p, contacts, locations); err != nil {
		return err
	}

	// Save dates in separate transaction
	if h.DatesSvc != nil {
		_ = h.DatesSvc.ReplaceForPerson(c.Request().Context(), id, importantDates)
	}

	return c.Redirect(http.StatusSeeOther, "/people/"+strconv.FormatInt(id, 10))
}

// PostDelete handles POST /people/:id/delete
func (h *PeopleHandlers) PostDelete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/people")
}

// GetDeleteConfirm handles GET /people/:id/delete-confirm (htmx fragment)
func (h *PeopleHandlers) GetDeleteConfirm(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	p, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	if p == nil {
		return echo.ErrNotFound
	}

	component := templates.PeopleDeleteConfirm(*p, auth.CSRFToken(c))
	return component.Render(c.Request().Context(), c.Response())
}

// PostContactRow handles POST /people/contact-row (htmx fragment)
func (h *PeopleHandlers) PostContactRow(c *echo.Context) error {
	count, _ := strconv.Atoi(c.FormValue("count"))
	component := templates.ContactRow(count)
	return component.Render(c.Request().Context(), c.Response())
}

// PostLocationRow handles POST /people/location-row (htmx fragment)
func (h *PeopleHandlers) PostLocationRow(c *echo.Context) error {
	count, _ := strconv.Atoi(c.FormValue("count"))
	component := templates.LocationRow(count)
	return component.Render(c.Request().Context(), c.Response())
}

// PostDateRow handles POST /people/:id/date-row (htmx fragment)
func (h *PeopleHandlers) PostDateRow(c *echo.Context) error {
	index, _ := strconv.Atoi(c.QueryParam("index"))
	component := templates.DateRow(index)
	return component.Render(c.Request().Context(), c.Response())
}

// PostAttachLabel handles POST /people/:id/labels/attach — attach a label (htmx).
func (h *PeopleHandlers) PostAttachLabel(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}
	labelID, err := strconv.ParseInt(c.FormValue("label_id"), 10, 64)
	if err != nil || labelID <= 0 {
		// No label selected — re-render strip unchanged.
		return h.renderLabelStrip(c, personID)
	}
	if h.LabelsSvc != nil {
		_ = h.LabelsSvc.Attach(c.Request().Context(), personID, labelID)
	}
	return h.renderLabelStrip(c, personID)
}

// PostDetachLabel handles POST /people/:id/labels/:labelID/delete — detach a label (htmx).
func (h *PeopleHandlers) PostDetachLabel(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}
	labelID, err := strconv.ParseInt(c.Param("labelID"), 10, 64)
	if err != nil {
		return echo.ErrNotFound
	}
	if h.LabelsSvc != nil {
		_ = h.LabelsSvc.Detach(c.Request().Context(), personID, labelID)
	}
	return h.renderLabelStrip(c, personID)
}

// renderLabelStrip fetches current label state and returns the htmx fragment.
func (h *PeopleHandlers) renderLabelStrip(c *echo.Context, personID int64) error {
	var attached, allLabels []labels.Label
	if h.LabelsSvc != nil {
		attached, _ = h.LabelsSvc.ListByPersonID(c.Request().Context(), personID)
		allLabels, _ = h.LabelsSvc.List(c.Request().Context())
	}
	component := templates.PersonLabelsStrip(attached, allLabels, personID, auth.CSRFToken(c))
	return component.Render(c.Request().Context(), c.Response())
}

// ---- helpers ----------------------------------------------------------------

func parseID(c *echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

// parseLabelIDs parses a comma-separated string of int64 label IDs.
// Invalid tokens are silently skipped.
func parseLabelIDs(s string) []int64 {
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

// parsePersonForm reads form values and returns domain structs plus an error message.
// Returns a non-empty error string when validation fails.
func parsePersonForm(c *echo.Context) (people.Person, []people.ContactInfo, []people.Location, []dates.ImportantDate, string) {
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return people.Person{}, nil, nil, nil, "Name is required."
	}

	p := people.Person{
		Prefix:           strings.TrimSpace(c.FormValue("prefix")),
		Name:             name,
		Nickname:         strings.TrimSpace(c.FormValue("nickname")),
		RelationshipType: strings.TrimSpace(c.FormValue("relationship_type")),
		OtherNotes:       strings.TrimSpace(c.FormValue("other_notes")),
	}

	if dob := strings.TrimSpace(c.FormValue("date_of_birth")); dob != "" {
		if t, err := time.Parse("2006-01-02", dob); err == nil {
			p.DateOfBirth = &t
		}
	}

	// Parse indexed contact rows.
	contactRows := forms.ParseIndexed(c.Request().Form, "contact")
	var contacts []people.ContactInfo
	for _, row := range contactRows {
		val := strings.TrimSpace(forms.GetField(row, "value"))
		if val == "" {
			continue // skip empty rows
		}
		contacts = append(contacts, people.ContactInfo{
			Type:  strings.TrimSpace(forms.GetField(row, "type")),
			Value: val,
			Label: strings.TrimSpace(forms.GetField(row, "label")),
		})
	}

	// Parse indexed location rows.
	locationRows := forms.ParseIndexed(c.Request().Form, "location")
	var locations []people.Location
	for _, row := range locationRows {
		addr := strings.TrimSpace(forms.GetField(row, "address"))
		city := strings.TrimSpace(forms.GetField(row, "city"))
		country := strings.TrimSpace(forms.GetField(row, "country"))
		postal := strings.TrimSpace(forms.GetField(row, "postal_code"))
		if addr == "" && city == "" && country == "" && postal == "" {
			continue // skip entirely empty rows
		}
		locations = append(locations, people.Location{
			Type:       strings.TrimSpace(forms.GetField(row, "type")),
			Address:    addr,
			City:       city,
			Country:    country,
			PostalCode: postal,
		})
	}

	// Parse indexed date rows.
	dateRows := forms.ParseIndexed(c.Request().Form, "date")
	var importantDates []dates.ImportantDate
	for i, row := range dateRows {
		dateValue := strings.TrimSpace(forms.GetField(row, "date_value"))
		if dateValue == "" {
			continue // skip empty rows
		}

		// Validate date format
		canonical, _, err := dates.ParseFlexible(dateValue)
		if err != nil {
			return p, contacts, locations, nil, "Invalid date format: " + dateValue + " (use YYYY-MM-DD or --MM-DD)"
		}

		recurring := forms.GetField(row, "recurring") == "1"
		importantDates = append(importantDates, dates.ImportantDate{
			Kind:      strings.TrimSpace(forms.GetField(row, "kind")),
			Label:     strings.TrimSpace(forms.GetField(row, "label")),
			DateValue: canonical,
			Recurring: recurring,
			Notes:     strings.TrimSpace(forms.GetField(row, "notes")),
			Position:  i,
		})
	}

	return p, contacts, locations, importantDates, ""
}
