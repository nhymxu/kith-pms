package handlers

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/dates"
	"github.com/nhymxu/kith-pms/internal/web/templates"
)

// DatesHandlers groups all /dates/* HTTP handlers.
type DatesHandlers struct {
	Svc *dates.Service
}

// GetUpcoming handles GET /dates?days=N
func (h *DatesHandlers) GetUpcoming(c *echo.Context) error {
	daysStr := c.QueryParam("days")
	days := 30 // default
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	today := time.Now()
	items, err := h.Svc.Upcoming(c.Request().Context(), today, days)
	if err != nil {
		return err
	}

	component := templates.DatesList(templates.DatesListParams{
		Items: items,
		Days:  days,
	})
	return component.Render(c.Request().Context(), c.Response())
}
