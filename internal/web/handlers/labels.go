package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/labels"
	"github.com/nhymxu/kith-pms/internal/web/templates"
)

// LabelsHandlers groups all /labels/* HTTP handlers.
type LabelsHandlers struct {
	Svc *labels.Service
}

// GetList handles GET /labels — shows all labels with usage counts.
func (h *LabelsHandlers) GetList(c *echo.Context) error {
	list, err := h.Svc.ListWithCounts(c.Request().Context())
	if err != nil {
		return err
	}
	component := templates.LabelsList(templates.LabelsListParams{
		Labels:    list,
		CSRFToken: auth.CSRFToken(c),
	})
	return component.Render(c.Request().Context(), c.Response())
}

// PostCreate handles POST /labels — create a new label.
func (h *LabelsHandlers) PostCreate(c *echo.Context) error {
	name := c.FormValue("name")
	color := c.FormValue("color")
	if color == "" {
		color = "#9ea096"
	}

	_, err := h.Svc.Create(c.Request().Context(), name, color)
	if err != nil {
		list, _ := h.Svc.ListWithCounts(c.Request().Context())
		formErr := labelErrMsg(err)
		component := templates.LabelsList(templates.LabelsListParams{
			Labels:      list,
			CSRFToken:   auth.CSRFToken(c),
			CreateError: formErr,
			CreateName:  name,
			CreateColor: color,
		})
		c.Response().WriteHeader(http.StatusUnprocessableEntity)
		return component.Render(c.Request().Context(), c.Response())
	}
	return c.Redirect(http.StatusSeeOther, "/labels")
}

// GetEdit handles GET /labels/:id/edit — show the edit form for a label.
func (h *LabelsHandlers) GetEdit(c *echo.Context) error {
	id, err := parseLabelID(c)
	if err != nil {
		return echo.ErrNotFound
	}
	l, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	if l == nil {
		return echo.ErrNotFound
	}
	component := templates.LabelEdit(templates.LabelEditParams{
		Label:     *l,
		CSRFToken: auth.CSRFToken(c),
	})
	return component.Render(c.Request().Context(), c.Response())
}

// PostUpdate handles POST /labels/:id — update label name and color.
func (h *LabelsHandlers) PostUpdate(c *echo.Context) error {
	id, err := parseLabelID(c)
	if err != nil {
		return echo.ErrNotFound
	}
	name := c.FormValue("name")
	color := c.FormValue("color")
	if color == "" {
		color = "#9ea096"
	}

	if err := h.Svc.Update(c.Request().Context(), id, name, color); err != nil {
		l, _ := h.Svc.Get(c.Request().Context(), id)
		if l == nil {
			return echo.ErrNotFound
		}
		l.Name = name
		l.Color = color
		component := templates.LabelEdit(templates.LabelEditParams{
			Label:     *l,
			CSRFToken: auth.CSRFToken(c),
			Error:     labelErrMsg(err),
		})
		c.Response().WriteHeader(http.StatusUnprocessableEntity)
		return component.Render(c.Request().Context(), c.Response())
	}
	return c.Redirect(http.StatusSeeOther, "/labels")
}

// PostDelete handles POST /labels/:id/delete — delete a label.
func (h *LabelsHandlers) PostDelete(c *echo.Context) error {
	id, err := parseLabelID(c)
	if err != nil {
		return echo.ErrNotFound
	}
	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/labels")
}

// ---- helpers ----------------------------------------------------------------

func parseLabelID(c *echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

// labelErrMsg converts a known service error to a user-visible message.
func labelErrMsg(err error) string {
	switch {
	case errors.Is(err, labels.ErrNameEmpty):
		return "Label name is required."
	case errors.Is(err, labels.ErrNameTooLong):
		return "Label name must be 64 characters or fewer."
	case errors.Is(err, labels.ErrInvalidColor):
		return "Color must be a valid 6-digit hex color (e.g. #a1b2c3)."
	case errors.Is(err, labels.ErrNameConflict):
		return "A label with that name already exists."
	default:
		return "An unexpected error occurred."
	}
}
