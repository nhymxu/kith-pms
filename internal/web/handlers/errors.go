package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/web/templates"
)

// CustomHTTPErrorHandler is set as echo.Echo.HTTPErrorHandler.
// For browser requests (Accept: text/html or non-/v1/* paths) it renders
// a styled templ error page. For API paths (/v1/*) it falls back to JSON.
// Echo v5 signature: func(c *Context, err error).
func CustomHTTPErrorHandler(c *echo.Context, err error) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	// Decide whether to render HTML or JSON.
	if wantsHTML(c) {
		renderHTMLError(c, code)
		return
	}

	// JSON fallback for /v1/* API routes.
	msg := http.StatusText(code)
	if jsonErr := c.JSON(code, map[string]string{"error": msg}); jsonErr != nil {
		slog.Error("error handler: json response", "error", jsonErr)
	}
}

// wantsHTML returns true when the request looks like a browser page request.
// Heuristic: Accept header contains "text/html" OR path does not start with "/v1".
func wantsHTML(c *echo.Context) bool {
	accept := c.Request().Header.Get("Accept")
	// Requests without an Accept header (e.g. direct URL bar navigation) get HTML.
	if accept == "" {
		return true
	}
	if strings.Contains(accept, "text/html") {
		return true
	}
	return !strings.HasPrefix(c.Request().URL.Path, "/v1")
}

// renderHTMLError writes the appropriate templ error page to the response.
func renderHTMLError(c *echo.Context, code int) {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(code)

	var renderErr error
	switch code {
	case http.StatusNotFound:
		renderErr = templates.Error404().Render(c.Request().Context(), c.Response())
	default:
		renderErr = templates.Error500(http.StatusText(code)).Render(c.Request().Context(), c.Response())
	}
	if renderErr != nil {
		slog.Error("error handler: render template", "error", renderErr)
	}
}
