package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/journal"
)

type JournalLabelsAPI struct {
	Svc *journal.LabelService
}

type journalLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (h *JournalLabelsAPI) List(c *echo.Context) error {
	list, err := h.Svc.ListWithCounts(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, list)
}

func (h *JournalLabelsAPI) Get(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	label, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if label == nil {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	return ok(c, label)
}

func (h *JournalLabelsAPI) Create(c *echo.Context) error {
	var req journalLabelRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Name) == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "name is required")
	}

	id, err := h.Svc.Create(c.Request().Context(), req.Name, req.Color)
	if err != nil {
		return journalLabelServiceErr(c, err)
	}

	return created(c, map[string]any{"id": id})
}

func (h *JournalLabelsAPI) Update(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	var req journalLabelRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.Svc.Update(c.Request().Context(), id, req.Name, req.Color); err != nil {
		return journalLabelServiceErr(c, err)
	}

	return ok(c, map[string]any{"id": id})
}

func (h *JournalLabelsAPI) Delete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	label, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if label == nil {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return noContent(c)
}

func journalLabelServiceErr(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, journal.ErrLabelNameEmpty),
		errors.Is(err, journal.ErrLabelNameTooLong),
		errors.Is(err, journal.ErrLabelInvalidColor):
		return apiErr(c, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, journal.ErrLabelNameConflict):
		return apiErr(c, http.StatusUnprocessableEntity, "name already exists")
	default:
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}
}
