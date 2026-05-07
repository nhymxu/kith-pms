package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/labels"
	"github.com/nhymxu/kith-pms/internal/relationships"
	"github.com/nhymxu/kith-pms/internal/web/templates"
	"github.com/nhymxu/kith-pms/pkg/config"
)

// SettingsHandlers groups all /settings/* HTTP handlers.
type SettingsHandlers struct {
	RelSvc    *relationships.Service
	LabelsSvc *labels.Service
	AuthSvc   *auth.Service
}

// GetHub handles GET /settings — renders the settings hub with navigation tiles.
func (h *SettingsHandlers) GetHub(c *echo.Context) error {
	component := templates.SettingsHub()
	return component.Render(c.Request().Context(), c.Response())
}

// GetRelationshipTypes handles GET /settings/relationship-types — list all types with create form.
func (h *SettingsHandlers) GetRelationshipTypes(c *echo.Context) error {
	types, err := h.RelSvc.ListTypesWithCounts(c.Request().Context())
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
		types, _ := h.RelSvc.ListTypesWithCounts(ctx)

		c.Response().WriteHeader(http.StatusUnprocessableEntity)

		return templates.RelationshipTypesList(templates.RelationshipTypesListParams{
			Types:         types,
			CSRFToken:     auth.CSRFToken(c),
			CreateError:   formErr,
			CreateName:    name,
			CreateReverse: reverseName,
		}).Render(ctx, c.Response())
	}

	if _, err := h.RelSvc.CreateType(ctx, name, reverseName); err != nil {
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

	rt, err := h.RelSvc.GetType(ctx, id)
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

	if err := h.RelSvc.UpdateType(ctx, id, name, reverseName); err != nil {
		rt, _ := h.RelSvc.GetType(ctx, id)
		if rt == nil {
			rt = &relationships.RelationshipType{ID: id, Name: name, ReverseName: reverseName}
		}

		c.Response().WriteHeader(http.StatusUnprocessableEntity)

		return templates.TypeRowEditing(*rt, auth.CSRFToken(c), typeErrMsg(err)).Render(ctx, c.Response())
	}

	rt, err := h.RelSvc.GetType(ctx, id)
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

	rt, err := h.RelSvc.GetType(ctx, id)
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

	if err := h.RelSvc.DeleteType(ctx, id); err != nil {
		if errors.Is(err, relationships.ErrTypeInUse) {
			types, _ := h.RelSvc.ListTypesWithCounts(ctx)

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

// ---- Labels handlers --------------------------------------------------------

// GetLabels handles GET /settings/labels.
func (h *SettingsHandlers) GetLabels(c *echo.Context) error {
	list, err := h.LabelsSvc.ListWithCounts(c.Request().Context())
	if err != nil {
		return err
	}

	return templates.LabelsList(templates.LabelsListParams{
		Labels:    list,
		CSRFToken: auth.CSRFToken(c),
	}).Render(c.Request().Context(), c.Response())
}

// PostCreateLabel handles POST /settings/labels.
func (h *SettingsHandlers) PostCreateLabel(c *echo.Context) error {
	ctx := c.Request().Context()
	name := c.FormValue("name")

	color := c.FormValue("color")
	if color == "" {
		color = "#9ea096"
	}

	_, err := h.LabelsSvc.Create(ctx, name, color)
	if err != nil {
		list, _ := h.LabelsSvc.ListWithCounts(ctx)

		c.Response().WriteHeader(http.StatusUnprocessableEntity)

		return templates.LabelsList(templates.LabelsListParams{
			Labels:      list,
			CSRFToken:   auth.CSRFToken(c),
			CreateError: labelErrMsg(err),
			CreateName:  name,
			CreateColor: color,
		}).Render(ctx, c.Response())
	}

	return c.Redirect(http.StatusSeeOther, "/settings/labels")
}

// GetLabelEdit handles GET /settings/labels/:id/edit.
func (h *SettingsHandlers) GetLabelEdit(c *echo.Context) error {
	id, err := parseSettingsLabelID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	l, err := h.LabelsSvc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}

	if l == nil {
		return echo.ErrNotFound
	}

	return templates.LabelEdit(templates.LabelEditParams{
		Label:     *l,
		CSRFToken: auth.CSRFToken(c),
	}).Render(c.Request().Context(), c.Response())
}

// PostUpdateLabel handles POST /settings/labels/:id.
func (h *SettingsHandlers) PostUpdateLabel(c *echo.Context) error {
	ctx := c.Request().Context()

	id, err := parseSettingsLabelID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	name := c.FormValue("name")

	color := c.FormValue("color")
	if color == "" {
		color = "#9ea096"
	}

	if err := h.LabelsSvc.Update(ctx, id, name, color); err != nil {
		l, _ := h.LabelsSvc.Get(ctx, id)
		if l == nil {
			return echo.ErrNotFound
		}

		l.Name = name
		l.Color = color

		c.Response().WriteHeader(http.StatusUnprocessableEntity)

		return templates.LabelEdit(templates.LabelEditParams{
			Label:     *l,
			CSRFToken: auth.CSRFToken(c),
			Error:     labelErrMsg(err),
		}).Render(ctx, c.Response())
	}

	return c.Redirect(http.StatusSeeOther, "/settings/labels")
}

// PostDeleteLabel handles POST /settings/labels/:id/delete.
func (h *SettingsHandlers) PostDeleteLabel(c *echo.Context) error {
	id, err := parseSettingsLabelID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	if err := h.LabelsSvc.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/settings/labels")
}

// ---- helpers ----------------------------------------------------------------

func parseSettingsTypeID(c *echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func parseSettingsLabelID(c *echo.Context) (int64, error) {
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

// GetSecurity handles GET /settings/security.
// Renders the password change form.
func (h *SettingsHandlers) GetSecurity(c *echo.Context) error {
	user := auth.UserFromContext(c)
	csrfToken := auth.CSRFToken(c)

	// Check for success message from redirect
	successMsg := ""
	if c.QueryParam("success") == "1" {
		successMsg = "Password changed successfully"
	}

	component := templates.SettingsSecurity(user, csrfToken, successMsg, "")

	return component.Render(c.Request().Context(), c.Response())
}

// PostChangePassword handles POST /settings/security/password.
// Validates and changes the user's password.
func (h *SettingsHandlers) PostChangePassword(c *echo.Context) error {
	user := auth.UserFromContext(c)
	csrfToken := auth.CSRFToken(c)

	currentPwd := c.FormValue("current_password")
	newPwd := c.FormValue("new_password")
	confirmPwd := c.FormValue("confirm_password")

	// Validation
	if currentPwd == "" || newPwd == "" || confirmPwd == "" {
		component := templates.SettingsSecurity(user, csrfToken, "", "All fields are required")
		return component.Render(c.Request().Context(), c.Response())
	}

	if newPwd != confirmPwd {
		component := templates.SettingsSecurity(user, csrfToken, "", "New passwords do not match")
		return component.Render(c.Request().Context(), c.Response())
	}

	if len(newPwd) < 8 {
		component := templates.SettingsSecurity(user, csrfToken, "", "Password must be at least 8 characters")
		return component.Render(c.Request().Context(), c.Response())
	}

	if currentPwd == newPwd {
		component := templates.SettingsSecurity(
			user,
			csrfToken,
			"",
			"New password must be different from current password",
		)

		return component.Render(c.Request().Context(), c.Response())
	}

	// Change password
	err := h.AuthSvc.ChangePassword(c.Request().Context(), currentPwd, newPwd)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			component := templates.SettingsSecurity(user, csrfToken, "", "Current password is incorrect")
			return component.Render(c.Request().Context(), c.Response())
		}

		return err
	}

	// Invalidate all other sessions
	if err := h.AuthSvc.LogoutAll(c.Request().Context()); err != nil {
		slog.Warn("settings: failed to logout all sessions after password change", "error", err)
	}

	// Re-issue new session for current request
	ip := c.RealIP()
	ua := c.Request().Header.Get("User-Agent")

	token, err := h.AuthSvc.Login(c.Request().Context(), newPwd, ip, ua)
	if err != nil {
		return fmt.Errorf("settings: re-issue session after password change: %w", err)
	}

	// Set new session cookie
	cookie := new(http.Cookie)
	cookie.Name = "kith_session"
	cookie.Value = token
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode

	cookie.Expires = time.Now().Add(config.ENV.SessionLifetime)
	if config.ENV.BehindTLS {
		cookie.Secure = true
	}

	http.SetCookie(c.Response(), cookie)

	// Redirect with success message
	return c.Redirect(http.StatusSeeOther, "/settings/security?success=1")
}
