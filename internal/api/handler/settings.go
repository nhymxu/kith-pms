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
}

func withServerConfig(s settings.UserSettings) settingsResponse {
	return settingsResponse{
		UserSettings:     s,
		MaxUploadSizeMB:  int(config.C.EffectiveMaxUploadBytes() / (1024 * 1024)),
		ImageMaxEdgePX:   config.C.EffectiveImageMaxEdgePX(),
		ImageJPEGQuality: config.C.EffectiveImageJPEGQuality(),
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
		return apiErr(c, http.StatusInternalServerError, "internal server error")
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
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	updated, err := h.Svc.Update(c.Request().Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, settings.ErrInvalidDateFormat),
			errors.Is(err, settings.ErrInvalidTimeFormat),
			errors.Is(err, settings.ErrInvalidTimezone),
			errors.Is(err, settings.ErrInvalidDefaultPeopleSort),
			errors.Is(err, settings.ErrInvalidDefaultPageSize),
			errors.Is(err, settings.ErrInvalidDashboardFavoritesCount),
			errors.Is(err, settings.ErrInvalidDashboardLastContactCount),
			errors.Is(err, settings.ErrInvalidTheme),
			errors.Is(err, settings.ErrInvalidNavLayout),
			errors.Is(err, settings.ErrInvalidSearchScope):
			return apiErr(c, http.StatusUnprocessableEntity, err.Error())
		default:
			return apiErr(c, http.StatusInternalServerError, "internal server error")
		}
	}

	return ok(c, withServerConfig(updated))
}
