package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/settings"
)

type SettingsAPI struct {
	Svc *settings.Service
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

	return ok(c, s)
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
			errors.Is(err, settings.ErrInvalidTimezone):
			return apiErr(c, http.StatusUnprocessableEntity, err.Error())
		default:
			return apiErr(c, http.StatusInternalServerError, "internal server error")
		}
	}

	return ok(c, updated)
}
