package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/reminders"
)

type RemindersAPI struct {
	Svc *reminders.Service
}

// reminderRequest is the JSON body for create and update.
type reminderRequest struct {
	Title           string `json:"title"`
	Notes           string `json:"notes"`
	DueDate         string `json:"due_date"` // "YYYY-MM-DD"
	PersonID        *int64 `json:"person_id"`
	ImportantDateID *int64 `json:"important_date_id"`
}

// Query params: status (upcoming|overdue|default=all), days (default 30 for upcoming)
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

func (h *RemindersAPI) Create(c *echo.Context) error {
	var req reminderRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Title) == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "title is required")
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return apiErr(c, http.StatusUnprocessableEntity, "due_date must be YYYY-MM-DD")
	}

	rem := &reminders.Reminder{
		Title:           req.Title,
		Notes:           req.Notes,
		DueDate:         dueDate,
		PersonID:        req.PersonID,
		ImportantDateID: req.ImportantDateID,
	}

	id, err := h.Svc.Create(c.Request().Context(), rem)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return created(c, map[string]any{"id": id})
}

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

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return apiErr(c, http.StatusUnprocessableEntity, "due_date must be YYYY-MM-DD")
	}

	rem := &reminders.Reminder{
		ID:              id,
		Title:           req.Title,
		Notes:           req.Notes,
		DueDate:         dueDate,
		PersonID:        req.PersonID,
		ImportantDateID: req.ImportantDateID,
	}

	if err := h.Svc.Update(c.Request().Context(), rem); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]any{"id": id})
}

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
