package web

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/nhymxu/kith-pms/internal/metrics"
	"github.com/nhymxu/kith-pms/pkg/config"
)

func New() *echo.Echo {
	return newEchoApp()
}

func newEchoApp() *echo.Echo {
	e := echo.New()

	skipPaths := []string{
		"/favicon.ico",
		"/swagger",
		"/metrics",
		"/health",
		"/ping",
		"/special-endpoint-can-replace-later",
	}

	skipper := func(c *echo.Context) bool {
		return slices.Contains(skipPaths, c.Request().URL.Path)
	}

	e.Use(
		middleware.RemoveTrailingSlashWithConfig(middleware.RemoveTrailingSlashConfig{
			RedirectCode: http.StatusMovedPermanently,
		}),
		metrics.Middleware(),
		middleware.Recover(),
		middleware.RequestID(),
		//middleware.Secure(),
		//middleware.CORS(),
		middleware.GzipWithConfig(middleware.GzipConfig{
			Level:   config.APIGzipLevel,
			Skipper: skipper,
		}), middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			Skipper:         skipper,
			LogURI:          true,
			LogStatus:       true,
			LogLatency:      true,
			LogRemoteIP:     true,
			LogMethod:       true,
			LogResponseSize: true,
			LogUserAgent:    true,
			LogRequestID:    true,
			LogHost:         true,
			HandleError:     true,
			LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error { // nolint:revive
				attrs := []slog.Attr{
					slog.String("remote_ip", v.RemoteIP),
					slog.Duration("latency", v.Latency),
					slog.String("host", v.Host),
					slog.String("request", fmt.Sprintf("%s %s", v.Method, v.URI)),
					slog.Int("status", v.Status),
					slog.Int64("size", v.ResponseSize),
					slog.String("user_agent", v.UserAgent),
					slog.String("request_id", v.RequestID),
				}

				n := v.Status
				switch {
				case n >= 500:
					slog.LogAttrs(c.Request().Context(), slog.LevelError, "Server error",
						append(attrs, slog.Any("error", v.Error))...)
				case n >= 400:
					slog.LogAttrs(c.Request().Context(), slog.LevelInfo, "Client error", attrs...)
				case n >= 300:
					slog.LogAttrs(c.Request().Context(), slog.LevelInfo, "Redirection", attrs...)
				default:
					slog.LogAttrs(c.Request().Context(), slog.LevelInfo, "Success", attrs...)
				}

				return nil
			},
		}),
		sentryecho.New(sentryecho.Options{}),
	)

	return e
}
