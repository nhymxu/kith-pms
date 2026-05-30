package handler

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/reminders"
)

// peopleLastContacter adapts people.Service to reminders.JournalLastContacter.
type peopleLastContacter struct {
	people *people.Service
}

func (p *peopleLastContacter) LastContactDate(ctx context.Context, personID int64) (time.Time, error) {
	person, err := p.people.Get(ctx, personID)
	if err != nil {
		return time.Time{}, err
	}

	if person.LastContactAt == nil {
		return time.Time{}, nil
	}

	return *person.LastContactAt, nil
}

// NewPeopleLastContacter creates a reminders.JournalLastContacter backed by people.Service.
func NewPeopleLastContacter(svc *people.Service) reminders.JournalLastContacter {
	return &peopleLastContacter{people: svc}
}

type RemindersAPI struct {
	Svc *reminders.Service
}

// reminderRequest is the JSON body for create and update.
type reminderRequest struct {
	Title             string                    `json:"title"`
	Notes             string                    `json:"notes"`
	DueDate           string                    `json:"due_date"` // "YYYY-MM-DD"
	PersonID          *int64                    `json:"person_id"`
	ImportantDateID   *int64                    `json:"important_date_id"`
	RecurrenceRule    *reminders.RecurrenceRule `json:"recurrence_rule"`
	RecurrenceEndDate *string                   `json:"recurrence_end_date"`
}

// List handles reminder listing. Query params: status (upcoming|overdue|default=all), days (default 30 for upcoming).
//
// @Summary      List reminders
// @Tags         reminders
// @Produce      json
// @Param        status  query  string  false  "Filter: upcoming, overdue, or all"
// @Param        days    query  int     false  "Days ahead for upcoming"  default(30)
// @Success      200  {object}  envelope
// @Failure      500  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /reminders [get]
func (h *RemindersAPI) List(c *echo.Context) error {
	status := c.QueryParam("status")

	switch status {
	case "upcoming":
		days, _ := strconv.Atoi(c.QueryParam("days"))
		if days < 1 {
			days = 30
		}

		list, err := h.Svc.GetUpcoming(c.Request().Context(), days)
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "internal server error")
		}

		return ok(c, list)

	case "overdue":
		list, err := h.Svc.GetOverdue(c.Request().Context())
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "internal server error")
		}

		return ok(c, list)

	default:
		list, err := h.Svc.List(c.Request().Context(), reminders.ListParams{})
		if err != nil {
			return apiErr(c, http.StatusInternalServerError, "internal server error")
		}

		return ok(c, list)
	}
}

// Get godoc
//
// @Summary      Get reminder
// @Tags         reminders
// @Produce      json
// @Param        id   path      int  true  "Reminder ID"
// @Success      200  {object}  envelope
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /reminders/{id} [get]
func (h *RemindersAPI) Get(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	r, err := h.Svc.GetByID(c.Request().Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, r)
}

// Create godoc
//
// @Summary      Create reminder
// @Tags         reminders
// @Accept       json
// @Produce      json
// @Param        body  body      reminderRequest  true  "Reminder data"
// @Success      201   {object}  envelope{data=object{id=int}}
// @Failure      400   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /reminders [post]
func (h *RemindersAPI) Create(c *echo.Context) error {
	var req reminderRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Title) == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "title is required")
	}

	dueDate, err := parseDateOrDatetime(req.DueDate)
	if err != nil {
		return apiErr(c, http.StatusUnprocessableEntity, "due_date must be YYYY-MM-DD or YYYY-MM-DDTHH:MM")
	}

	rem := &reminders.Reminder{
		Title:           req.Title,
		Notes:           req.Notes,
		DueDate:         dueDate,
		PersonID:        req.PersonID,
		ImportantDateID: req.ImportantDateID,
		RecurrenceRule:  req.RecurrenceRule,
	}

	if req.RecurrenceEndDate != nil {
		t, err := parseDateOrDatetime(*req.RecurrenceEndDate)
		if err != nil {
			return apiErr(
				c,
				http.StatusUnprocessableEntity,
				"recurrence_end_date must be YYYY-MM-DD or YYYY-MM-DDTHH:MM",
			)
		}

		rem.RecurrenceEndDate = &t
	}

	id, err := h.Svc.Create(c.Request().Context(), rem)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return created(c, map[string]any{"id": id})
}

// Update godoc
//
// @Summary      Update reminder
// @Tags         reminders
// @Accept       json
// @Produce      json
// @Param        id    path      int              true  "Reminder ID"
// @Param        body  body      reminderRequest  true  "Reminder data"
// @Success      200   {object}  envelope{data=object{id=int}}
// @Failure      400   {object}  envelope
// @Failure      404   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /reminders/{id} [put]
func (h *RemindersAPI) Update(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	if _, err := h.Svc.GetByID(c.Request().Context(), id); errors.Is(err, sql.ErrNoRows) {
		return apiErr(c, http.StatusNotFound, "not found")
	} else if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	var req reminderRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Title) == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "title is required")
	}

	dueDate, err := parseDateOrDatetime(req.DueDate)
	if err != nil {
		return apiErr(c, http.StatusUnprocessableEntity, "due_date must be YYYY-MM-DD or YYYY-MM-DDTHH:MM")
	}

	rem := &reminders.Reminder{
		ID:              id,
		Title:           req.Title,
		Notes:           req.Notes,
		DueDate:         dueDate,
		PersonID:        req.PersonID,
		ImportantDateID: req.ImportantDateID,
		RecurrenceRule:  req.RecurrenceRule,
	}

	if req.RecurrenceEndDate != nil {
		t, err := parseDateOrDatetime(*req.RecurrenceEndDate)
		if err != nil {
			return apiErr(
				c,
				http.StatusUnprocessableEntity,
				"recurrence_end_date must be YYYY-MM-DD or YYYY-MM-DDTHH:MM",
			)
		}

		rem.RecurrenceEndDate = &t
	}

	if err := h.Svc.Update(c.Request().Context(), rem); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]any{"id": id})
}

// Delete godoc
//
// @Summary      Delete reminder
// @Tags         reminders
// @Produce      json
// @Param        id   path  int  true  "Reminder ID"
// @Success      204
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /reminders/{id} [delete]
func (h *RemindersAPI) Delete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	if _, err := h.Svc.GetByID(c.Request().Context(), id); errors.Is(err, sql.ErrNoRows) {
		return apiErr(c, http.StatusNotFound, "not found")
	} else if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return noContent(c)
}

// Complete godoc
//
// @Summary      Complete reminder
// @Tags         reminders
// @Produce      json
// @Param        id   path      int  true  "Reminder ID"
// @Success      200  {object}  envelope{data=object{id=int}}
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /reminders/{id}/complete [patch]
func (h *RemindersAPI) Complete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	if _, err := h.Svc.GetByID(c.Request().Context(), id); errors.Is(err, sql.ErrNoRows) {
		return apiErr(c, http.StatusNotFound, "not found")
	} else if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if err := h.Svc.MarkComplete(c.Request().Context(), id); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]any{"id": id})
}
