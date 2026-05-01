package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/web/templates"
	"github.com/nhymxu/kith-pms/pkg/config"
)

const sessionCookieName = "kith_session"

// AuthHandlers groups all auth-related HTTP handlers.
type AuthHandlers struct {
	Svc *auth.Service
}

// GetLogin handles GET /login.
// Redirects to / when the user is already authenticated.
func (h *AuthHandlers) GetLogin(c *echo.Context) error {
	if auth.UserFromContext(c) != nil {
		return c.Redirect(http.StatusFound, "/")
	}

	next := c.QueryParam("next")
	showError := c.QueryParam("error") == "invalid"
	csrfToken := auth.CSRFToken(c)

	component := templates.Login(csrfToken, next, showError)
	return component.Render(c.Request().Context(), c.Response())
}

// PostLogin handles POST /login.
// Validates password, sets session cookie, redirects to next or /.
func (h *AuthHandlers) PostLogin(c *echo.Context) error {
	password := c.FormValue("password")
	next := c.FormValue("next")

	ip := c.RealIP()
	ua := c.Request().Header.Get("User-Agent")

	token, err := h.Svc.Login(c.Request().Context(), password, ip, ua)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			// Redirect back with error flag — no password in redirect.
			redirectURL := "/login?error=invalid"
			if next != "" {
				redirectURL += "&next=" + url.QueryEscape(next)
			}
			return c.Redirect(http.StatusSeeOther, redirectURL)
		}
		return err
	}

	setSessionCookie(c, token, time.Now().Add(config.ENV.SessionLifetime))
	if next != "" {
		// Basic open-redirect guard: only allow relative paths.
		if u, err := url.Parse(next); err == nil && u.Host == "" && len(next) > 0 && next[0] == '/' {
			return c.Redirect(http.StatusSeeOther, next)
		}
	}
	return c.Redirect(http.StatusSeeOther, "/")
}

// PostLogout handles POST /logout.
// Deletes current session and clears cookie.
func (h *AuthHandlers) PostLogout(c *echo.Context) error {
	cookie, err := c.Request().Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		_ = h.Svc.Logout(c.Request().Context(), cookie.Value)
	}
	clearSessionCookie(c)
	return c.Redirect(http.StatusSeeOther, "/login")
}

// PostLogoutAll handles POST /logout-all.
// Deletes all sessions for the user and clears cookie.
func (h *AuthHandlers) PostLogoutAll(c *echo.Context) error {
	_ = h.Svc.LogoutAll(c.Request().Context())
	clearSessionCookie(c)
	return c.Redirect(http.StatusSeeOther, "/login")
}

// ---- cookie helpers ---------------------------------------------------------

func setSessionCookie(c *echo.Context, token string, expires time.Time) {
	cookie := new(http.Cookie)
	cookie.Name = sessionCookieName
	cookie.Value = token
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Expires = expires
	if config.ENV.BehindTLS {
		cookie.Secure = true
	}
	http.SetCookie(c.Response(), cookie)
}

func clearSessionCookie(c *echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = sessionCookieName
	cookie.Value = ""
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.MaxAge = -1
	http.SetCookie(c.Response(), cookie)
}
