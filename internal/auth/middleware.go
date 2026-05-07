package auth

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"golang.org/x/time/rate"
)

const (
	cookieName = "kith_session"
	contextKey = "user"
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
