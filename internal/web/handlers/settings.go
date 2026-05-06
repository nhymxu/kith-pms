package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/relationships"
	"github.com/nhymxu/kith-pms/internal/web/templates"
)

// SettingsHandlers groups all /settings/* HTTP handlers.
type SettingsHandlers struct {
	Svc *relationships.Service
}

// GetHub handles GET /settings — renders the settings hub with navigation tiles.
func (h *SettingsHandlers) GetHub(c *echo.Context) error {
	component := templates.SettingsHub()
	return component.Render(c.Request().Context(), c.Response())
}

// GetRelationshipTypes handles GET /settings/relationship-types — list all types with create form.
func (h *SettingsHandlers) GetRelationshipTypes(c *echo.Context) error {
	types, err := h.Svc.ListTypesWithCounts(c.Request().Context())
	if err != nil {
		return err
	}
	return templates.RelationshipTypesList(templates.RelationshipTypesListParams{
		Types:     types,
		CSRFToken: auth.CSRFToken(c),
	}).Render(c.Request().Context(), c.Response())
}

// PostRelationshipType handles POST /settings/relationship-types — create a new type.
func (h *SettingsHandlers) PostRelationshipType(c *echo.Context) error {
	ctx := c.Request().Context()
	name := c.FormValue("name")
	reverseName := c.FormValue("reverse_name")

	rerender := func(formErr string) error {
		types, _ := h.Svc.ListTypesWithCounts(ctx)
		c.Response().WriteHeader(http.StatusUnprocessableEntity)
		return templates.RelationshipTypesList(templates.RelationshipTypesListParams{
			Types:       types,
			CSRFToken:   auth.CSRFToken(c),
			CreateError: formErr,
			CreateName:  name,
			CreateReverse: reverseName,
		}).Render(ctx, c.Response())
	}

	if _, err := h.Svc.CreateType(ctx, name, reverseName); err != nil {
		return rerender(typeErrMsg(err))
	}
	return c.Redirect(http.StatusSeeOther, "/settings/relationship-types")
}

// GetRelationshipTypeEdit handles GET /settings/relationship-types/:id/edit — HTMX row swap to edit form.
func (h *SettingsHandlers) GetRelationshipTypeEdit(c *echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseSettingsTypeID(c)
	if err != nil {
		return echo.ErrNotFound
	}
	rt, err := h.Svc.GetType(ctx, id)
	if err != nil {
		return err
	}
	if rt == nil {
		return echo.ErrNotFound
	}
	return templates.TypeRowEditing(*rt, auth.CSRFToken(c), "").Render(ctx, c.Response())
}

// PostUpdateRelationshipType handles POST /settings/relationship-types/:id — update a type inline.
func (h *SettingsHandlers) PostUpdateRelationshipType(c *echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseSettingsTypeID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	name := c.FormValue("name")
	reverseName := c.FormValue("reverse_name")

	if err := h.Svc.UpdateType(ctx, id, name, reverseName); err != nil {
		rt, _ := h.Svc.GetType(ctx, id)
		if rt == nil {
			rt = &relationships.RelationshipType{ID: id, Name: name, ReverseName: reverseName}
		}
		c.Response().WriteHeader(http.StatusUnprocessableEntity)
		return templates.TypeRowEditing(*rt, auth.CSRFToken(c), typeErrMsg(err)).Render(ctx, c.Response())
	}

	rt, err := h.Svc.GetType(ctx, id)
	if err != nil || rt == nil {
		return c.Redirect(http.StatusSeeOther, "/settings/relationship-types")
	}
	return templates.TypeRow(*rt, auth.CSRFToken(c)).Render(ctx, c.Response())
}

// GetRelationshipTypeRow handles GET /settings/relationship-types/:id/row — returns the display row (used by Cancel in inline edit).
func (h *SettingsHandlers) GetRelationshipTypeRow(c *echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseSettingsTypeID(c)
	if err != nil {
		return echo.ErrNotFound
	}
	rt, err := h.Svc.GetType(ctx, id)
	if err != nil {
		return err
	}
	if rt == nil {
		return echo.ErrNotFound
	}
	return templates.TypeRow(*rt, auth.CSRFToken(c)).Render(ctx, c.Response())
}

// PostDeleteRelationshipType handles POST /settings/relationship-types/:id/delete.
func (h *SettingsHandlers) PostDeleteRelationshipType(c *echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseSettingsTypeID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	if err := h.Svc.DeleteType(ctx, id); err != nil {
		if errors.Is(err, relationships.ErrTypeInUse) {
			types, _ := h.Svc.ListTypesWithCounts(ctx)
			c.Response().WriteHeader(http.StatusConflict)
			return templates.RelationshipTypesList(templates.RelationshipTypesListParams{
				Types:       types,
				CSRFToken:   auth.CSRFToken(c),
				DeleteError: "Cannot delete a relationship type that is in use. Remove all relationships using this type first.",
			}).Render(ctx, c.Response())
		}
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/settings/relationship-types")
}

// ---- helpers ----------------------------------------------------------------

func parseSettingsTypeID(c *echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func typeErrMsg(err error) string {
	switch {
	case errors.Is(err, relationships.ErrNameEmpty):
		return "Name is required."
	case errors.Is(err, relationships.ErrNameTooLong):
		return "Name must be 80 characters or fewer."
	default:
		return "An unexpected error occurred."
	}
}
