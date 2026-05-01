package handlers

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/dates"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/labels"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/web/templates"
)

// HomeHandler handles the dashboard homepage (GET /).
// It pulls summary counts and recent records from the database.
type HomeHandler struct {
	DB         *sql.DB
	PeopleSvc  *people.Service
	LabelsSvc  *labels.Service
	JournalSvc *journal.Service
	DatesSvc   *dates.Service
}

// Get handles GET / and renders the dashboard.
func (h *HomeHandler) Get(c *echo.Context) error {
	ctx := c.Request().Context()

	data := templates.DashboardData{}

	// Fetch counts in parallel using goroutines with a simple error aggregation.
	type countResult struct {
		name  string
		count int
		err   error
	}
	ch := make(chan countResult, 3)

	go func() {
		n, err := countRows(ctx, h.DB, "SELECT COUNT(*) FROM person")
		ch <- countResult{"people", n, err}
	}()
	go func() {
		n, err := countRows(ctx, h.DB, "SELECT COUNT(*) FROM label")
		ch <- countResult{"labels", n, err}
	}()
	go func() {
		n, err := countRows(ctx, h.DB, "SELECT COUNT(*) FROM activity")
		ch <- countResult{"activities", n, err}
	}()

	for i := 0; i < 3; i++ {
		r := <-ch
		if r.err != nil {
			slog.Warn("dashboard: count error", "table", r.name, "error", r.err)
		}
		switch r.name {
		case "people":
			data.PeopleCount = r.count
		case "labels":
			data.LabelsCount = r.count
		case "activities":
			data.ActivitiesCount = r.count
		}
	}

	// Latest 5 activities.
	activities, err := h.JournalSvc.List(ctx, journal.ListParams{PageSize: 5, Page: 1})
	if err != nil {
		slog.Warn("dashboard: list activities", "error", err)
	} else {
		data.RecentActivities = activities
	}

	// Recently added 5 people (service.List orders by created_at DESC via repo).
	recentPeople, err := h.PeopleSvc.List(ctx, people.ListParams{PageSize: 5, Page: 1})
	if err != nil {
		slog.Warn("dashboard: list people", "error", err)
	} else {
		data.RecentPeople = recentPeople
	}

	// On this day dates
	if h.DatesSvc != nil {
		today := time.Now()
		onThisDay, err := h.DatesSvc.OnThisDay(ctx, today)
		if err != nil {
			slog.Warn("dashboard: on this day", "error", err)
		} else {
			data.OnThisDay = onThisDay
		}
	}

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)
	return templates.Home(data).Render(ctx, c.Response())
}

// countRows executes a scalar COUNT(*) query and returns the integer result.
func countRows(ctx context.Context, db *sql.DB, query string) (int, error) {
	var n int
	if err := db.QueryRowContext(ctx, query).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
