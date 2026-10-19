package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/settings"
	"github.com/nhymxu/kith-pms/pkg/config"
)

type SettingsAPI struct {
	Svc *settings.Service
}

// settingsResponse adds read-only, server-config-derived fields alongside the
// persisted user settings. Not part of settings.UserSettings since it isn't
// stored per-user; extra JSON fields sent back on PUT are simply ignored.
type settingsResponse struct {
	settings.UserSettings
	MaxUploadSizeMB int `json:"max_upload_size_mb"`
	// Image encode caps the SPA applies when re-encoding cropped uploads.
	ImageMaxEdgePX   int `json:"image_max_edge_px"`
	ImageJPEGQuality int `json:"image_jpeg_quality"`
	// PeopleDeleteRetentionDays lets the SPA compute "days remaining" for a
	// soft-deleted person without a separate config endpoint.
	PeopleDeleteRetentionDays int `json:"people_delete_retention_days"`
}

func withServerConfig(s settings.UserSettings) settingsResponse {
	return settingsResponse{
		UserSettings:              s,
		MaxUploadSizeMB:           int(config.C.EffectiveMaxUploadBytes() / (1024 * 1024)),
		ImageMaxEdgePX:            config.C.EffectiveImageMaxEdgePX(),
		ImageJPEGQuality:          config.C.EffectiveImageJPEGQuality(),
		PeopleDeleteRetentionDays: config.C.PeopleDeleteRetentionDays,
	}
}

// Get godoc
//
// @Summary      Get settings
// @Tags         settings
// @Produce      json
// @Success      200  {object}  envelope
// @Failure      500  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /settings [get]
func (h *SettingsAPI) Get(c *echo.Context) error {
	s, err := h.Svc.Get(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	return ok(c, withServerConfig(s))
}

// Update godoc
//
// @Summary      Update settings
// @Tags         settings
// @Accept       json
// @Produce      json
// @Param        body  body      settings.UserSettings  true  "Settings"
// @Success      200   {object}  envelope
// @Failure      400   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /settings [put]
func (h *SettingsAPI) Update(c *echo.Context) error {
	var req settings.UserSettings
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid_request_body", "invalid request body")
	}

	updated, err := h.Svc.Update(c.Request().Context(), req)
	if err != nil {
		if code, found := settingsErrCode(err); found {
			return apiErr(c, http.StatusUnprocessableEntity, code, err.Error())
		}

		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	return ok(c, withServerConfig(updated))
}

// settingsErrCodes maps each settings validation sentinel to its client code.
var settingsErrCodes = []struct {
	err  error
	code string
}{
	{settings.ErrInvalidDateFormat, "settings_date_format_invalid"},
	{settings.ErrInvalidTimeFormat, "settings_time_format_invalid"},
	{settings.ErrInvalidTimezone, "settings_timezone_invalid"},
	{settings.ErrInvalidRetentionDays, "settings_retention_days_invalid"},
	{settings.ErrInvalidNetworkColorBy, "settings_network_color_by_invalid"},
	{settings.ErrInvalidNetworkOnlyMineDepth, "settings_network_depth_invalid"},
	{settings.ErrInvalidDefaultPeopleSort, "settings_people_sort_invalid"},
	{settings.ErrInvalidDefaultPageSize, "settings_page_size_invalid"},
	{settings.ErrInvalidDashboardFavoritesCount, "settings_dashboard_favorites_invalid"},
	{settings.ErrInvalidDashboardLastContactCount, "settings_dashboard_last_contact_invalid"},
	{settings.ErrInvalidTheme, "settings_theme_invalid"},
	{settings.ErrInvalidNavLayout, "settings_nav_layout_invalid"},
	{settings.ErrInvalidNumberFormat, "settings_number_format_invalid"},
	{settings.ErrInvalidRelationshipDisplay, "settings_relationship_display_invalid"},
	{settings.ErrInvalidLanguage, "settings_language_invalid"},
	{settings.ErrInvalidSearchScope, "settings_search_scope_invalid"},
}

func settingsErrCode(err error) (string, bool) {
	for _, e := range settingsErrCodes {
		if errors.Is(err, e.err) {
			return e.code, true
		}
	}

	return "", false
}
