package api

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
	swaggerFiles "github.com/swaggo/files/v2"
	"github.com/swaggo/swag"

	// Import generated docs so the swag runtime registers the spec.
	_ "github.com/nhymxu/kith-pms/internal/api/swagger"
)

// mountSwagger serves Swagger UI at /swagger/* for Echo v5.
// Disabled in production (ENV=production) — exposes full API surface.
// swaggo/echo-swagger targets Echo v4; this thin wrapper avoids that dep.
func mountSwagger(e *echo.Echo) {
	if os.Getenv("ENV") == "production" {
		return
	}

	e.GET("/swagger/*", func(c *echo.Context) error {
		path := strings.TrimPrefix(c.Request().URL.Path, "/swagger/")
		if path == "" {
			return c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
		}

		// Serve the OpenAPI spec from the swag runtime.
		if path == "doc.json" {
			spec, err := swag.ReadDoc()
			if err != nil {
				return echo.ErrInternalServerError
			}

			return c.JSONBlob(http.StatusOK, []byte(spec))
		}

		f, err := swaggerFiles.FS.Open(path)
		if err != nil {
			return echo.ErrNotFound
		}
		defer func() { _ = f.Close() }()

		stat, err := f.Stat()
		if err != nil {
			return echo.ErrNotFound
		}

		rs, ok := f.(io.ReadSeeker)
		if !ok {
			return echo.ErrInternalServerError
		}

		http.ServeContent(c.Response(), c.Request(), stat.Name(), stat.ModTime(), rs)

		return nil
	})
}
