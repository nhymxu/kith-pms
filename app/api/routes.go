package api

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/nhymxu/kith-pms/pkg/config"
)

func groupV1Routes(e *echo.Group) {
	e.GET("/test", testFuncRequestID)

	publicGroup := e.Group("")
	publicGroup.GET("/test_public", testFunc)

	sec := privateRoutesV1(publicGroup)
	adminRoutesV1(sec)
}

func privateRoutesV1(e *echo.Group) *echo.Group {
	privateGroup := e.Group("")
	// TODO: can change to JWT auth later: https://echo.labstack.com/docs/middleware/jwt
	privateGroup.Use(middleware.KeyAuth(func(_ *echo.Context, key string, _ middleware.ExtractorSource) (bool, error) {
		return key == config.ENV.TokenAuth, nil
	}))
	privateGroup.Use(validateUserMiddleware)
	privateGroup.GET("/test_private", testFuncPrivate)

	return privateGroup
}

func adminRoutesV1(e *echo.Group) *echo.Group {
	adminGroup := e.Group("/admin")
	// TODO: can change to JWT auth later: https://echo.labstack.com/docs/middleware/jwt
	adminGroup.Use(validateAdminMiddleware)
	adminGroup.GET("/test", testFuncPrivate)

	return adminGroup
}

func validateUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		// TODO: do something like set user from jwt token
		var err error
		if err != nil {
			slog.Warn(err.Error())
			return ctx.Redirect(http.StatusFound, "/logout")
		}

		return next(ctx)
	}
}

func validateAdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		// TODO: get user info from context jwt
		type User struct {
			Active bool
			Admin  bool
		}

		var u *User
		if u == nil || !u.Active {
			slog.Warn("User is not found")
			return ctx.Redirect(http.StatusFound, "/logout")
		}

		if !u.Admin {
			slog.Warn("User is not an admin")
			return ctx.Redirect(http.StatusFound, "/dashboard")
		}

		return next(ctx)
	}
}
