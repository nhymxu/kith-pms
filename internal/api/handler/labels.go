package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/labels"
)

type LabelsAPI struct {
	Svc *labels.Service
}

// labelRequest is the JSON body for create and update.
type labelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"` // "#RRGGBB"
}

// List godoc
//
// @Summary      List labels
// @Tags         labels
// @Produce      json
// @Success      200  {object}  envelope
// @Failure      500  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /labels [get]
func (h *LabelsAPI) List(c *echo.Context) error {
	list, err := h.Svc.ListWithCounts(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, list)
}

// Get godoc
//
// @Summary      Get label
// @Tags         labels
// @Produce      json
// @Param        id   path      int  true  "Label ID"
// @Success      200  {object}  envelope
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /labels/{id} [get]
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

// Create godoc
//
// @Summary      Create label
// @Tags         labels
// @Accept       json
// @Produce      json
// @Param        body  body      labelRequest  true  "Label data"
// @Success      201   {object}  envelope{data=object{id=int}}
// @Failure      400   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /labels [post]
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

// Update godoc
//
// @Summary      Update label
// @Tags         labels
// @Accept       json
// @Produce      json
// @Param        id    path      int           true  "Label ID"
// @Param        body  body      labelRequest  true  "Label data"
// @Success      200   {object}  envelope{data=object{id=int}}
// @Failure      400   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /labels/{id} [put]
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

// Delete godoc
//
// @Summary      Delete label
// @Tags         labels
// @Produce      json
// @Param        id   path  int  true  "Label ID"
// @Success      204
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /labels/{id} [delete]
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
