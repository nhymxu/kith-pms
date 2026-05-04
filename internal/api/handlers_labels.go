package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/labels"
)

// LabelsAPI handles /v1/labels CRUD endpoints.
type LabelsAPI struct {
	Svc *labels.Service
}

// labelRequest is the JSON body for create and update.
type labelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"` // "#RRGGBB"
}

// List handles GET /v1/labels
func (h *LabelsAPI) List(c *echo.Context) error {
	list, err := h.Svc.ListWithCounts(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}
	return ok(c, list)
}

// Get handles GET /v1/labels/:id
func (h *LabelsAPI) Get(c *echo.Context) error {
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

// Create handles POST /v1/labels
func (h *LabelsAPI) Create(c *echo.Context) error {
	var req labelRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Name) == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "name is required")
	}

	id, err := h.Svc.Create(c.Request().Context(), req.Name, req.Color)
	if err != nil {
		return labelsServiceErr(c, err)
	}

	return created(c, map[string]any{"id": id})
}

// Update handles PUT /v1/labels/:id
func (h *LabelsAPI) Update(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	var req labelRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.Svc.Update(c.Request().Context(), id, req.Name, req.Color); err != nil {
		return labelsServiceErr(c, err)
	}

	return ok(c, map[string]any{"id": id})
}

// Delete handles DELETE /v1/labels/:id
func (h *LabelsAPI) Delete(c *echo.Context) error {
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

// labelsServiceErr maps known labels validation errors to HTTP responses.
func labelsServiceErr(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, labels.ErrNameEmpty),
		errors.Is(err, labels.ErrNameTooLong),
		errors.Is(err, labels.ErrInvalidColor):
		return apiErr(c, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, labels.ErrNameConflict):
		return apiErr(c, http.StatusUnprocessableEntity, "name already exists")
	default:
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}
}
