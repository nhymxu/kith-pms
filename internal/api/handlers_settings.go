package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/settings"
)

type SettingsAPI struct {
	Svc *settings.Service
}

func (h *SettingsAPI) Get(c *echo.Context) error {
	s, err := h.Svc.Get(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, s)
}

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
