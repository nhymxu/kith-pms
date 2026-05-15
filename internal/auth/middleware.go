package auth

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"golang.org/x/time/rate"
)

const (
	cookieName = "kith_session"
	contextKey = "user"
	// cookieAuthedKey marks whether the request was authenticated via cookie (vs Bearer).
	cookieAuthedKey = "cookie_authed"
)

// SessionLoader reads the kith_session cookie, resolves it to a *User via svc,
// and stores the result in the Echo context under the key "user".
// It is a no-op (no redirect) when the cookie is absent or invalid — use
// RequireAuth on protected routes to enforce authentication.
func SessionLoader(svc *Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			cookie, err := c.Request().Cookie(cookieName)
			if err == nil && cookie.Value != "" {
				user, err := svc.LoadUser(c.Request().Context(), cookie.Value)
				if err == nil && user != nil {
					c.Set(contextKey, user)
				}
			}

			return next(c)
		}
	}
}

// RequireAuth redirects to /login?next=<escaped_path> when no authenticated
// user is present in the Echo context. Attach after SessionLoader.
func RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if c.Get(contextKey) == nil {
				next := url.QueryEscape(c.Request().URL.RequestURI())
				return c.Redirect(http.StatusFound, "/login?next="+next)
			}

			return next(c)
		}
	}
}

// UserFromContext retrieves the authenticated *User stored by SessionLoader.
// Returns nil when no user is loaded.
func UserFromContext(c *echo.Context) *User {
	u, _ := c.Get(contextKey).(*User)
	return u
}

// ipRateLimiter holds per-IP token buckets.
type ipRateLimiter struct {
	buckets sync.Map
	max     int
	window  time.Duration
}

func (l *ipRateLimiter) allow(ip string) bool {
	val, _ := l.buckets.LoadOrStore(ip, rate.NewLimiter(
		rate.Every(l.window/time.Duration(l.max)),
		l.max,
	))
	lim := val.(*rate.Limiter)

	return lim.Allow()
}

// RateLimitLogin returns middleware that limits login attempts to max requests
// per window per remote IP. Excess requests receive 429 Too Many Requests.
func RateLimitLogin(limit int, window time.Duration) echo.MiddlewareFunc {
	limiter := &ipRateLimiter{max: limit, window: window}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ip := c.RealIP()
			if !limiter.allow(ip) {
				return c.String(http.StatusTooManyRequests, "too many requests")
			}

			return next(c)
		}
	}
}

// SessionOrBearer is an authentication middleware for /v1 routes.
// It accepts either:
//   - A valid Bearer token matching apiToken (machine clients), or
//   - A valid kith_session cookie resolving to a live *User (SPA clients).
//
// On success it stores the *User under the "user" context key.
// On failure it returns JSON 401 {error:"unauthorized"}.
// Auth failures are logged identically for both paths.
func SessionOrBearer(apiToken string, svc *Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// 1. Try Bearer token first.
			if header := c.Request().Header.Get("Authorization"); header != "" {
				t, ok := extractBearer(header)
				if ok && apiToken != "" {
					tb := []byte(t)
					kb := []byte(apiToken)
					if len(tb) == len(kb) && subtle.ConstantTimeCompare(tb, kb) == 1 {
						// Machine client authenticated via Bearer — user stays nil.
						return next(c)
					}
				}
				// Header present but invalid — fall through to cookie check.
				slog.Warn("auth: invalid bearer token", "ip", c.RealIP())
			}

			// 2. Try session cookie.
			cookie, err := c.Request().Cookie(cookieName)
			if err == nil && cookie.Value != "" {
				user, err := svc.LoadUser(c.Request().Context(), cookie.Value)
				if err == nil && user != nil {
					c.Set(contextKey, user)
					c.Set(cookieAuthedKey, true)
					return next(c)
				}
				if err != nil {
					slog.Warn("auth: session lookup error", "ip", c.RealIP(), "err", err)
				}
			}

			slog.Warn("auth: unauthorized request", "ip", c.RealIP(), "path", c.Request().URL.Path)
			return jsonErr(c, http.StatusUnauthorized, "unauthorized")
		}
	}
}

// SpaCSRF enforces a custom-header CSRF gate for cookie-authenticated state-changing requests.
// When the request was authenticated via cookie AND the method is POST/PUT/PATCH/DELETE,
// the header "X-Requested-With: kith-spa" must be present. Bearer-authenticated requests skip
// this check entirely.
func SpaCSRF() echo.MiddlewareFunc {
	stateMethods := map[string]bool{
		http.MethodPost:   true,
		http.MethodPut:    true,
		http.MethodPatch:  true,
		http.MethodDelete: true,
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if !stateMethods[c.Request().Method] {
				return next(c)
			}

			// Only enforce for cookie-authenticated requests.
			if authed, _ := c.Get(cookieAuthedKey).(bool); !authed {
				return next(c)
			}

			if c.Request().Header.Get("X-Requested-With") != "kith-spa" {
				return jsonErr(c, http.StatusForbidden, "missing X-Requested-With")
			}

			return next(c)
		}
	}
}

// IsCookieAuthed reports whether the current request was authenticated via session cookie.
func IsCookieAuthed(c *echo.Context) bool {
	v, _ := c.Get(cookieAuthedKey).(bool)
	return v
}

// jsonErr writes a JSON error envelope. Used by middleware where api package is not importable.
func jsonErr(c *echo.Context, code int, msg string) error {
	return c.JSON(code, map[string]string{"error": msg})
}

func extractBearer(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	t := strings.TrimPrefix(header, prefix)
	if t == "" {
		return "", false
	}

	return t, true
}
