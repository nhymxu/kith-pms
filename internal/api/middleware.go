package api

import (
	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/audit"
	"github.com/nhymxu/kith-pms/internal/auth"
)

func injectAPIActor() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			var actorID int64
			if u := auth.UserFromContext(c); u != nil {
				actorID = u.ID
			}

			ctx := audit.WithActor(c.Request().Context(), actorID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
