package handler

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

// Login handles POST /v1/auth/login.
//
// @Summary      Login
// @Description  Authenticate with password. Sets kith_session HttpOnly cookie on success.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      object{password=string}  true  "Login credentials"
// @Success      200   {object}  envelope{data=object{logged_in=bool}}
// @Failure      400   {object}  envelope
// @Failure      401   {object}  envelope
// @Router       /auth/login [post]
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
//
// @Summary      Logout
// @Description  Revoke current session and clear the session cookie.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  envelope{data=object{logged_out=bool}}
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /auth/logout [post]
func (h *AuthAPI) Logout(c *echo.Context) error {
	cookie, err := c.Request().Cookie(apiCookieName)
	if err == nil && cookie.Value != "" {
		_ = h.Svc.Logout(c.Request().Context(), cookie.Value)
	}

	h.clearSessionCookie(c)

	return ok(c, map[string]any{"logged_out": true})
}

// LogoutAll handles POST /v1/auth/logout-all.
//
// @Summary      Logout all sessions
// @Description  Revoke all active sessions and clear the session cookie.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  envelope{data=object{logged_out=bool}}
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /auth/logout-all [post]
func (h *AuthAPI) LogoutAll(c *echo.Context) error {
	_ = h.Svc.LogoutAll(c.Request().Context())
	h.clearSessionCookie(c)

	return ok(c, map[string]any{"logged_out": true})
}

// Me handles GET /v1/auth/me.
//
// @Summary      Get current user
// @Description  Returns the authenticated user's ID and creation time.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  envelope{data=object{id=int,created_at=string}}
// @Failure      401  {object}  envelope
// @Security     CookieAuth
// @Router       /auth/me [get]
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
//
// @Summary      Change password
// @Description  Change the login password. Revokes all sessions and re-issues a new one.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      object{current_password=string,new_password=string,confirm_password=string}  true  "Password change request"
// @Success      200   {object}  envelope{data=object{password_changed=bool}}
// @Failure      400   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /auth/password [post]
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
		lifetime = config.C.SessionLifetime
	}

	cookie := new(http.Cookie)
	cookie.Name = apiCookieName
	cookie.Value = token
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	cookie.Expires = time.Now().Add(lifetime)

	behindTLS := h.BehindTLS || config.C.BehindTLS
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
