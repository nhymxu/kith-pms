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
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	return ok(c, list)
}

func (h *JournalLabelsAPI) Get(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid_id", "invalid id")
	}

	label, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	if label == nil {
		return apiErr(c, http.StatusNotFound, "not_found", "not found")
	}

	return ok(c, label)
}

func (h *JournalLabelsAPI) Create(c *echo.Context) error {
	var req journalLabelRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid_request_body", "invalid request body")
	}

	if strings.TrimSpace(req.Name) == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "name_required", "name is required")
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
		return apiErr(c, http.StatusBadRequest, "invalid_id", "invalid id")
	}

	var req journalLabelRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid_request_body", "invalid request body")
	}

	if err := h.Svc.Update(c.Request().Context(), id, req.Name, req.Color); err != nil {
		return journalLabelServiceErr(c, err)
	}

	return ok(c, map[string]any{"id": id})
}

func (h *JournalLabelsAPI) Delete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid_id", "invalid id")
	}

	label, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	if label == nil {
		return apiErr(c, http.StatusNotFound, "not_found", "not found")
	}

	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	return noContent(c)
}

func journalLabelServiceErr(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, journal.ErrLabelNameEmpty):
		return apiErr(c, http.StatusUnprocessableEntity, "label_name_empty", err.Error())
	case errors.Is(err, journal.ErrLabelNameTooLong):
		return apiErrParams(
			c,
			http.StatusUnprocessableEntity,
			"label_name_too_long",
			err.Error(),
			map[string]any{"max": journal.MaxLabelNameLen},
		)
	case errors.Is(err, journal.ErrLabelInvalidColor):
		return apiErr(c, http.StatusUnprocessableEntity, "label_color_invalid", err.Error())
	case errors.Is(err, journal.ErrLabelNameConflict):
		return apiErr(c, http.StatusUnprocessableEntity, "name_exists", "name already exists")
	default:
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}
}
