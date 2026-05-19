package metrics

import (
	"context"
	"database/sql"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SessionCounter is satisfied by auth.SessionRepo — avoids import cycle.
type SessionCounter interface {
	CountActiveSessions(ctx context.Context) (int64, error)
}

var (
	Registry = prometheus.NewRegistry()

	httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by method, route template, and status code.",
	}, []string{"method", "route", "status"})

	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency by method and route template.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)

func init() {
	Registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		httpRequests,
		httpDuration,
	)
}

// RegisterAppCollectors registers DB-size and active-session gauges.
func RegisterAppCollectors(db *sql.DB, sessions SessionCounter) {
	Registry.MustRegister(
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "kith_db_size_bytes",
			Help: "SQLite database size in bytes (page_count * page_size).",
		}, func() float64 {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			var size float64

			row := db.QueryRowContext(ctx,
				`SELECT page_count * page_size FROM pragma_page_count(), pragma_page_size()`)
			_ = row.Scan(&size)

			return size
		}),
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "kith_sessions_active",
			Help: "Number of non-expired sessions.",
		}, func() float64 {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			n, _ := sessions.CountActiveSessions(ctx)

			return float64(n)
		}),
	)
}

// RegisterBuildInfo registers a build-info gauge (value always 1).
func RegisterBuildInfo() {
	version := "dev"
	commit := "unknown"

	if info, ok := debug.ReadBuildInfo(); ok {
		version = info.Main.Version
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				if len(s.Value) > 8 {
					commit = s.Value[:8]
				} else {
					commit = s.Value
				}
			}
		}
	}

	Registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "kith_build_info",
		Help:        "Build metadata (value is always 1).",
		ConstLabels: prometheus.Labels{"version": version, "commit": commit},
	}, func() float64 { return 1 }))
}

// Middleware instruments every request with request count and latency.
// Uses c.Path() (Echo route template) to keep label cardinality bounded.
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			err := next(c)

			route := c.Path()
			if route == "" {
				route = "unknown"
			}

			method := c.Request().Method

			statusCode := 0
			if r, ok := c.Response().(*echo.Response); ok {
				statusCode = r.Status
			}

			status := strconv.Itoa(statusCode)
			elapsed := time.Since(start).Seconds()

			httpRequests.WithLabelValues(method, route, status).Inc()
			httpDuration.WithLabelValues(method, route).Observe(elapsed)

			return err
		}
	}
}

// Handler returns an Echo handler that serves the Prometheus metrics page.
func Handler() echo.HandlerFunc {
	h := promhttp.HandlerFor(Registry, promhttp.HandlerOpts{})

	return func(c *echo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
