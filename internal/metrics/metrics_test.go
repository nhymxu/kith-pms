package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/metrics"
)

func newTestEcho() *echo.Echo {
	e := echo.New()
	e.Use(metrics.Middleware())

	return e
}

func TestMiddleware_CountsRequests(t *testing.T) {
	e := newTestEcho()
	e.GET("/v1/people", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	e.GET("/v1/people/:id", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Two requests to /v1/people
	for range 2 {
		req := httptest.NewRequest(http.MethodGet, "/v1/people", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
	}

	// One request to /v1/people/:id (raw path, route template must be used)
	req := httptest.NewRequest(http.MethodGet, "/v1/people/42", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Scrape /metrics
	body := scrapeMetrics(t, e)

	if !strings.Contains(body, `http_requests_total`) {
		t.Fatal("expected http_requests_total in metrics output")
	}

	// Route template label — must NOT contain raw ID
	if strings.Contains(body, `route="/v1/people/42"`) {
		t.Error("metrics must use route template /v1/people/:id, not raw path /v1/people/42")
	}

	if !strings.Contains(body, `route="/v1/people/:id"`) {
		t.Error("expected route template /v1/people/:id in metrics output")
	}
}

func TestMiddleware_UnknownRouteLabel(t *testing.T) {
	e := newTestEcho()

	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	body := scrapeMetrics(t, e)

	// Unmatched routes must use "unknown" label, not the raw path
	if strings.Contains(body, `route="/does-not-exist"`) {
		t.Error("unmatched route must use label unknown, not raw path")
	}
}

func TestHandler_Returns200(t *testing.T) {
	e := echo.New()
	e.GET("/metrics", metrics.Handler())

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("expected text/plain content-type, got %s", ct)
	}
}

// scrapeMetrics registers the /metrics route on e and returns the body.
func scrapeMetrics(t *testing.T, e *echo.Echo) string {
	t.Helper()
	e.GET("/metrics", metrics.Handler())

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("metrics endpoint returned %d", rec.Code)
	}

	return rec.Body.String()
}
