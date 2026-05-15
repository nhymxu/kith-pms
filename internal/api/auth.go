package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/pkg/config"
)

const apiCookieName = "kith_session"

// AuthAPI handles /v1/auth/* endpoints.
type AuthAPI struct {
	Svc             *auth.Service
	SessionLifetime time.Duration
	BehindTLS       bool
}

func newAuthAPI(deps Deps) *AuthAPI {
	return &AuthAPI{
		Svc:             deps.AuthService,
		SessionLifetime: deps.SessionLifetime,
		BehindTLS:       deps.BehindTLS,
	}
}

// Login handles POST /v1/auth/login.
// Body: {"password": "..."}.
// On success sets kith_session cookie and returns {data: {id: <userID>}}.
func (h *AuthAPI) Login(c *echo.Context) error {
	var req struct {
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil || req.Password == "" {
		return apiErr(c, http.StatusBadRequest, "password is required")
	}

	ip := c.RealIP()
	ua := c.Request().Header.Get("User-Agent")

	token, err := h.Svc.Login(c.Request().Context(), req.Password, ip, ua)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return apiErr(c, http.StatusUnauthorized, "invalid credentials")
		}

		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	h.setSessionCookie(c, token)

	return ok(c, map[string]any{"logged_in": true})
}

// Logout handles POST /v1/auth/logout.
// Revokes the current session and clears the cookie.
func (h *AuthAPI) Logout(c *echo.Context) error {
	cookie, err := c.Request().Cookie(apiCookieName)
	if err == nil && cookie.Value != "" {
		_ = h.Svc.Logout(c.Request().Context(), cookie.Value)
	}

	h.clearSessionCookie(c)

	return ok(c, map[string]any{"logged_out": true})
}

// LogoutAll handles POST /v1/auth/logout-all.
// Revokes all sessions for the user and clears the cookie.
func (h *AuthAPI) LogoutAll(c *echo.Context) error {
	_ = h.Svc.LogoutAll(c.Request().Context())
	h.clearSessionCookie(c)

	return ok(c, map[string]any{"logged_out": true})
}

// Me handles GET /v1/auth/me.
// Returns the authenticated user's ID, or 401 if not authenticated.
func (h *AuthAPI) Me(c *echo.Context) error {
	user := auth.UserFromContext(c)
	if user == nil {
		return apiErr(c, http.StatusUnauthorized, "unauthorized")
	}

	return ok(c, map[string]any{
		"id":         user.ID,
		"created_at": user.CreatedAt,
	})
}

// ChangePassword handles POST /v1/auth/password.
// Body: {"current_password": "...", "new_password": "...", "confirm_password": "..."}.
func (h *AuthAPI) ChangePassword(c *echo.Context) error {
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if req.CurrentPassword == "" || req.NewPassword == "" || req.ConfirmPassword == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "all fields are required")
	}

	if req.NewPassword != req.ConfirmPassword {
		return apiErr(c, http.StatusUnprocessableEntity, "new passwords do not match")
	}

	if len(req.NewPassword) < 8 {
		return apiErr(c, http.StatusUnprocessableEntity, "password must be at least 8 characters")
	}

	if req.CurrentPassword == req.NewPassword {
		return apiErr(c, http.StatusUnprocessableEntity, "new password must differ from current")
	}

	if err := h.Svc.ChangePassword(c.Request().Context(), req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return apiErr(c, http.StatusUnprocessableEntity, "current password is incorrect")
		}

		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	// Revoke all sessions and re-issue so the caller remains logged in.
	_ = h.Svc.LogoutAll(c.Request().Context())

	ip := c.RealIP()
	ua := c.Request().Header.Get("User-Agent")

	token, err := h.Svc.Login(c.Request().Context(), req.NewPassword, ip, ua)
	if err != nil {
		// Password changed but re-issue failed — user needs to log in manually.
		h.clearSessionCookie(c)
		return ok(c, map[string]any{"password_changed": true})
	}

	h.setSessionCookie(c, token)

	return ok(c, map[string]any{"password_changed": true})
}

// ---- cookie helpers ---------------------------------------------------------

func (h *AuthAPI) setSessionCookie(c *echo.Context, token string) {
	lifetime := h.SessionLifetime
	if lifetime <= 0 {
		lifetime = config.ENV.SessionLifetime
	}

	cookie := new(http.Cookie)
	cookie.Name = apiCookieName
	cookie.Value = token
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Expires = time.Now().Add(lifetime)

	behindTLS := h.BehindTLS || config.ENV.BehindTLS
	if behindTLS {
		cookie.Secure = true
	}

	http.SetCookie(c.Response(), cookie)
}

func (h *AuthAPI) clearSessionCookie(c *echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = apiCookieName
	cookie.Value = ""
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.MaxAge = -1
	http.SetCookie(c.Response(), cookie)
}
