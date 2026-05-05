package api

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/audit"
)

// BearerAuth returns an Echo middleware that validates a static Bearer token.
// If token is empty, every request gets a 501 "api not configured" response.
func BearerAuth(token string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if token == "" {
				return apiErr(c, http.StatusNotImplemented, "api not configured")
			}

			header := c.Request().Header.Get("Authorization")
			t, ok := extractBearer(header)
			if !ok {
				return apiErr(c, http.StatusUnauthorized, "unauthorized")
			}

			tb := []byte(t)
			kb := []byte(token)
			// Compare lengths first to avoid subtle.ConstantTimeCompare on unequal-length inputs.
			if len(tb) != len(kb) {
				return apiErr(c, http.StatusUnauthorized, "unauthorized")
			}
			if subtle.ConstantTimeCompare(tb, kb) != 1 {
				return apiErr(c, http.StatusUnauthorized, "unauthorized")
			}

			return next(c)
		}
	}
}

// injectAPIActor injects a fixed actor ID (0) into the request context to
// distinguish API-originated mutations from web UI mutations in the audit log.
func injectAPIActor() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := audit.WithActor(c.Request().Context(), 0)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// extractBearer parses "Bearer <token>" from the Authorization header value.
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
