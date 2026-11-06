package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/calendar"
)

type CalendarAPI struct {
	Tokens  *calendar.TokenStore
	Builder *calendar.Builder
}

type feedTokenResponse struct {
	Token string `json:"token"`
}

// GetToken godoc
//
// @Summary      Get calendar feed token
// @Tags         calendar
// @Produce      json
// @Success      200  {object}  envelope
// @Failure      500  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /calendar/feed-token [get]
func (h *CalendarAPI) GetToken(c *echo.Context) error {
	tok, err := h.Tokens.Get(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	return ok(c, feedTokenResponse{Token: tok})
}

// RegenerateToken godoc
//
// @Summary      Regenerate calendar feed token
// @Description  Invalidates every existing subscription URL.
// @Tags         calendar
// @Produce      json
// @Success      200  {object}  envelope
// @Failure      500  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /calendar/feed-token/regenerate [post]
func (h *CalendarAPI) RegenerateToken(c *echo.Context) error {
	tok, err := h.Tokens.Regenerate(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal", "internal server error")
	}

	return ok(c, feedTokenResponse{Token: tok})
}

// Feed serves the public, token-guarded ICS feed. Any token failure is a bare
// 404 so the URL reveals nothing about whether a token exists.
func (h *CalendarAPI) Feed(calName string, parts []calendar.Part, selectable bool) echo.HandlerFunc {
	return func(c *echo.Context) error {
		ctx := c.Request().Context()

		valid, err := h.Tokens.Valid(ctx, c.Param("token"))
		if err != nil {
			return c.String(http.StatusInternalServerError, "internal server error")
		}

		if !valid {
			return c.String(http.StatusNotFound, "not found")
		}

		use := parts

		if inc := c.QueryParam("include"); selectable && inc != "" {
			use = calendar.ParseParts(inc)
			if len(use) == 0 {
				return c.String(http.StatusBadRequest, "include must list birthdays, dates or reminders")
			}
		}

		body, err := h.Builder.Build(ctx, calName, use...)
		if err != nil {
			return c.String(http.StatusInternalServerError, "internal server error")
		}

		c.Response().Header().Set("Cache-Control", "private, max-age=300")

		return c.Blob(http.StatusOK, "text/calendar; charset=utf-8", []byte(body))
	}
}
