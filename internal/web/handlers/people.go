package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
	"github.com/nhymxu/kith-pms/internal/work_history"
)

// PeopleHandlers groups all /people/* HTTP handlers.
type PeopleHandlers struct {
	Svc             *people.Service
	LabelsSvc       *labels.Service
	JournalSvc      *journal.Service
	DatesSvc        *dates.Service
	WorkHistorySvc  *work_history.Service
	AvatarBasePath  string
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
		CSRFToken:   auth.CSRFToken(c),
		Dates:       []dates.ImportantDate{},
		WorkHistory: []work_history.WorkEntry{},
	})
	return component.Render(c.Request().Context(), c.Response())
}

// PostCreate handles POST /people
func (h *PeopleHandlers) PostCreate(c *echo.Context) error {
	p, contacts, locations, importantDates, workEntries, formErr := parsePersonForm(c)
	if formErr != "" {
		component := templates.PeopleForm(templates.PeopleFormParams{
			Person:      p,
			Contacts:    contacts,
			Locations:   locations,
			Dates:       importantDates,
			WorkHistory: workEntries,
			CSRFToken:   auth.CSRFToken(c),
			Error:       formErr,
		})
		return component.Render(c.Request().Context(), c.Response())
	}

	id, err := h.Svc.Create(c.Request().Context(), p, contacts, locations)
	if err != nil {
		return err
	}

	// Save dates in separate transaction.
	if h.DatesSvc != nil && len(importantDates) > 0 {
		if err := h.DatesSvc.ReplaceForPerson(c.Request().Context(), id, importantDates); err != nil {
			slog.Error("failed to save important dates", "person_id", id, "err", err)
		}
	}

	// Save work history in separate transaction.
	if h.WorkHistorySvc != nil {
		if err := h.WorkHistorySvc.ReplaceForPerson(c.Request().Context(), id, workEntries); err != nil {
			slog.Error("failed to save work history", "person_id", id, "err", err)
		}
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

	// Fetch work history
	var workHistory []work_history.WorkEntry
	if h.WorkHistorySvc != nil {
		workHistory, _ = h.WorkHistorySvc.ListByPerson(c.Request().Context(), id)
	}

	component := templates.PeopleDetail(templates.PeopleDetailParams{
		Person:           *p,
		Labels:           attached,
		AllLabels:        allLabels,
		CSRFToken:        auth.CSRFToken(c),
		RecentActivities: recentActivities,
		Dates:            importantDates,
		WorkHistory:      workHistory,
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

	var workEntries []work_history.WorkEntry
	if h.WorkHistorySvc != nil {
		workEntries, _ = h.WorkHistorySvc.ListByPerson(c.Request().Context(), id)
	}

	component := templates.PeopleForm(templates.PeopleFormParams{
		Person:      *p,
		Contacts:    p.Contacts,
		Locations:   p.Locations,
		Dates:       importantDates,
		WorkHistory: workEntries,
		CSRFToken:   auth.CSRFToken(c),
		IsEdit:      true,
	})
	return component.Render(c.Request().Context(), c.Response())
}

// PostUpdate handles POST /people/:id
func (h *PeopleHandlers) PostUpdate(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	p, contacts, locations, importantDates, workEntries, formErr := parsePersonForm(c)
	p.ID = id

	if formErr != "" {
		component := templates.PeopleForm(templates.PeopleFormParams{
			Person:      p,
			Contacts:    contacts,
			Locations:   locations,
			Dates:       importantDates,
			WorkHistory: workEntries,
			CSRFToken:   auth.CSRFToken(c),
			IsEdit:      true,
			Error:       formErr,
		})
		return component.Render(c.Request().Context(), c.Response())
	}

	if err := h.Svc.Update(c.Request().Context(), p, contacts, locations); err != nil {
		return err
	}

	// Save dates in separate transaction.
	if h.DatesSvc != nil {
		if err := h.DatesSvc.ReplaceForPerson(c.Request().Context(), id, importantDates); err != nil {
			slog.Error("failed to save important dates", "person_id", id, "err", err)
		}
	}

	// Save work history in separate transaction.
	if h.WorkHistorySvc != nil {
		if err := h.WorkHistorySvc.ReplaceForPerson(c.Request().Context(), id, workEntries); err != nil {
			slog.Error("failed to save work history", "person_id", id, "err", err)
		}
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

// PostWorkRow handles POST /people/work-row (htmx fragment)
func (h *PeopleHandlers) PostWorkRow(c *echo.Context) error {
	index, _ := strconv.Atoi(c.QueryParam("index"))
	component := templates.WorkRow(index)
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
func parsePersonForm(c *echo.Context) (people.Person, []people.ContactInfo, []people.Location, []dates.ImportantDate, []work_history.WorkEntry, string) {
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return people.Person{}, nil, nil, nil, nil, "Name is required."
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

		// Validate date format.
		canonical, _, err := dates.ParseFlexible(dateValue)
		if err != nil {
			return p, contacts, locations, nil, nil, "Invalid date format: " + dateValue + " (use YYYY-MM-DD or --MM-DD)"
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

	// Parse indexed work history rows.
	workRows := forms.ParseIndexed(c.Request().Form, "work")
	var workEntries []work_history.WorkEntry
	for i, row := range workRows {
		company := strings.TrimSpace(forms.GetField(row, "company"))
		if company == "" {
			continue // skip empty rows
		}

		startDate := strings.TrimSpace(forms.GetField(row, "start_date"))
		if startDate == "" {
			return p, contacts, locations, importantDates, nil, "Work history entry is missing a start date."
		}
		canonicalStart, err := work_history.ParseWorkDate(startDate)
		if err != nil {
			return p, contacts, locations, importantDates, nil, "Invalid work history start date: " + startDate + " (use YYYY, YYYY-MM, or YYYY-MM-DD)"
		}

		endDate := strings.TrimSpace(forms.GetField(row, "end_date"))
		canonicalEnd := ""
		if endDate != "" {
			canonicalEnd, err = work_history.ParseWorkDate(endDate)
			if err != nil {
				return p, contacts, locations, importantDates, nil, "Invalid work history end date: " + endDate + " (use YYYY, YYYY-MM, or YYYY-MM-DD)"
			}
		}

		workEntries = append(workEntries, work_history.WorkEntry{
			Company:     company,
			Title:       strings.TrimSpace(forms.GetField(row, "title")),
			StartDate:   canonicalStart,
			EndDate:     canonicalEnd,
			Location:    strings.TrimSpace(forms.GetField(row, "location")),
			Description: strings.TrimSpace(forms.GetField(row, "description")),
			Position:    i,
		})
	}

	return p, contacts, locations, importantDates, workEntries, ""
}

// PostUploadAvatar handles POST /people/:id/avatar
func (h *PeopleHandlers) PostUploadAvatar(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	// Limit request body size (5MB + 1MB overhead for multipart)
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 6*1024*1024)

	file, err := c.FormFile("avatar")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			return c.String(http.StatusBadRequest, "File too large (max 5MB)")
		}
		return c.String(http.StatusBadRequest, "No file uploaded")
	}

	src, err := file.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to open file")
	}
	defer src.Close()

	if err := h.Svc.UploadAvatar(c.Request().Context(), personID, src, file); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/people/"+strconv.FormatInt(personID, 10))
}

// GetAvatar handles GET /people/:id/avatar
func (h *PeopleHandlers) GetAvatar(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	p, err := h.Svc.Get(c.Request().Context(), personID)
	if err != nil {
		return err
	}
	if p == nil || p.AvatarPath == "" {
		return echo.ErrNotFound
	}

	fullPath := filepath.Join(h.AvatarBasePath, p.AvatarPath)
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(h.AvatarBasePath)) {
		return echo.ErrNotFound
	}

	f, err := os.Open(cleanPath)
	if err != nil {
		return echo.ErrNotFound
	}
	defer f.Close()

	c.Response().Header().Set("Content-Type", p.AvatarMimeType)
	c.Response().Header().Set("Cache-Control", "public, max-age=86400")
	_, err = io.Copy(c.Response(), f)
	return err
}

// PostDeleteAvatar handles POST /people/:id/avatar/delete
func (h *PeopleHandlers) PostDeleteAvatar(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	if err := h.Svc.DeleteAvatar(c.Request().Context(), personID); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, "/people/"+strconv.FormatInt(personID, 10))
}
