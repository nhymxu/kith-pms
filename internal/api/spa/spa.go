// Package spa serves the embedded React SPA and provides a catch-all fallback
// for client-side routing so deep-link refreshes return index.html (HTTP 200).
package spa

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

//go:embed all:public
var publicFS embed.FS

// Handler mounts SPA routes onto e:
//   - GET /assets/*     → embedded hashed asset files, 1-year immutable cache
//   - GET /favicon.*    → favicon files
//   - GET * (catch-all) → index.html with no-cache (excludes /v1/*, /health)
//
// Call this LAST, after API and health routes, so it acts as the fallback.
// The `public/` directory is populated by `make web` (copies web/dist → internal/api/spa/public).
func Handler(e *echo.Echo) {
	sub, err := fs.Sub(publicFS, "public")
	if err != nil {
		panic("spa: failed to sub public FS: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(sub))

	// Hashed Vite asset files — 1-year immutable cache.
	e.GET("/assets/*", func(c *echo.Context) error {
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		fileServer.ServeHTTP(c.Response(), c.Request())

		return nil
	})

	// Root-level static files (favicons, manifest, robots.txt …).
	e.GET("/favicon.*", func(c *echo.Context) error {
		fileServer.ServeHTTP(c.Response(), c.Request())

		return nil
	})

	// Catch-all: every non-API GET returns index.html so the React router handles
	// the path on the client. Real 404s for unknown /assets/* paths are handled
	// by the /assets/* route above (http.FileServer returns 404).
	e.GET("/*", spaFallback(sub))
}

func spaFallback(sub fs.FS) echo.HandlerFunc {
	return func(c *echo.Context) error {
		path := c.Request().URL.Path

		// Never intercept API or health — let Echo return 404 naturally.
		if path == "/health" ||
			strings.HasPrefix(path, "/v1") ||
			strings.HasPrefix(path, "/assets/") {
			return echo.ErrNotFound
		}

		f, err := sub.Open("index.html")
		if err != nil {
			return echo.ErrNotFound
		}

		defer func() { _ = f.Close() }()

		rs, ok := f.(io.ReadSeeker)
		if !ok {
			// Fallback: read all and write manually (embed.FS always implements ReadSeeker).
			data, readErr := io.ReadAll(f)
			if readErr != nil {
				return echo.ErrInternalServerError
			}

			setIndexHeaders(c)
			_, _ = c.Response().Write(data)

			return nil
		}

		stat, err := f.Stat()
		if err != nil {
			return echo.ErrInternalServerError
		}

		setIndexHeaders(c)
		http.ServeContent(c.Response(), c.Request(), "index.html", stat.ModTime(), rs)

		return nil
	}
}

func setIndexHeaders(c *echo.Context) {
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set(
		"Content-Security-Policy",
		// unsafe-inline needed for Tailwind's runtime style injection
		"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data: blob:; font-src 'self'; connect-src 'self'",
	)
}
