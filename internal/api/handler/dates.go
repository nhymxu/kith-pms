package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/dates"
)

// DatesAPI handles person-scoped important dates and global upcoming endpoints.
type DatesAPI struct {
	Svc *dates.Service
}

// datesReplaceRequest is the JSON body for full replace of a person's dates.
type datesReplaceRequest struct {
	Dates []importantDateRequest `json:"dates"`
}

type importantDateRequest struct {
	Kind      string `json:"kind"`
	Label     string `json:"label"`
	DateValue string `json:"date_value"` // "YYYY-MM-DD" or "--MM-DD"
	Recurring bool   `json:"recurring"`
	Notes     string `json:"notes"`
	Position  int    `json:"position"`
}

// ListByPerson handles GET /v1/people/:id/dates
//
// @Summary      List important dates for person
// @Tags         dates
// @Produce      json
// @Param        id   path      int  true  "Person ID"
// @Success      200  {object}  envelope
// @Failure      400  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id}/dates [get]
func (h *DatesAPI) ListByPerson(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	list, err := h.Svc.ListByPerson(c.Request().Context(), personID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, list)
}

// ReplaceForPerson handles PUT /v1/people/:id/dates
//
// @Summary      Replace dates for person
// @Tags         dates
// @Accept       json
// @Produce      json
// @Param        id    path      int                  true  "Person ID"
// @Param        body  body      datesReplaceRequest  true  "Dates list"
// @Success      200   {object}  envelope
// @Failure      400   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id}/dates [put]
func (h *DatesAPI) ReplaceForPerson(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	var req datesReplaceRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	domainDates := make([]dates.ImportantDate, 0, len(req.Dates))
	for _, d := range req.Dates {
		domainDates = append(domainDates, dates.ImportantDate{
			PersonID:  personID,
			Kind:      d.Kind,
			Label:     d.Label,
			DateValue: d.DateValue,
			Recurring: d.Recurring,
			Notes:     d.Notes,
			Position:  d.Position,
		})
	}

	if err := h.Svc.ReplaceForPerson(c.Request().Context(), personID, domainDates); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, nil)
}

// upcomingItem is the slim JSON representation of an OnThisDayItem.
type upcomingItem struct {
	Person         personRef `json:"person"`
	Kind           string    `json:"kind"`
	DateValue      string    `json:"date_value"`
	YearsSince     int       `json:"years_since"`
	NextOccurrence string    `json:"next_occurrence"`
}

type personRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Upcoming godoc
//
// @Summary      Get upcoming dates
// @Tags         dates
// @Produce      json
// @Param        days  query     int  false  "Days ahead to look"  default(30)
// @Success      200   {object}  envelope
// @Failure      500   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /dates/upcoming [get]
func (h *DatesAPI) Upcoming(c *echo.Context) error {
	days, _ := strconv.Atoi(c.QueryParam("days"))
	if days < 1 {
		days = 30
	}

	if days > 365 {
		days = 365
	}

	today := time.Now()

	items, err := h.Svc.Upcoming(c.Request().Context(), today, days)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	result := make([]upcomingItem, 0, len(items))
	for _, item := range items {
		// Compute next occurrence date string for display.
		nextOcc := nextOccurrenceStr(item.Date, today)
		result = append(result, upcomingItem{
			Person:         personRef{ID: item.Person.ID, Name: item.Person.Name},
			Kind:           item.Date.Kind,
			DateValue:      item.Date.DateValue,
			YearsSince:     item.YearsSince,
			NextOccurrence: nextOcc,
		})
	}

	return ok(c, result)
}

// nextOccurrenceStr computes the next occurrence date as "YYYY-MM-DD".
// Mirrors the logic in dates.nextOccurrence but returns a string instead of time.Time.
func nextOccurrenceStr(d dates.ImportantDate, today time.Time) string {
	monthDay := d.MonthDay()
	if d.IsYearless() {
		if len(monthDay) != 5 {
			return ""
		}

		candidate, err := time.Parse("2006-01-02", today.Format("2006")+"-"+monthDay)
		if err != nil {
			return ""
		}

		if !candidate.Before(today.Truncate(24 * time.Hour)) {
			return candidate.Format("2006-01-02")
		}

		next, err := time.Parse("2006-01-02", strconv.Itoa(today.Year()+1)+"-"+monthDay)
		if err != nil {
			return ""
		}

		return next.Format("2006-01-02")
	}

	exact, err := time.Parse("2006-01-02", d.DateValue)
	if err != nil {
		return ""
	}

	todayTrunc := today.Truncate(24 * time.Hour)

	if d.Recurring {
		if len(monthDay) != 5 {
			return ""
		}

		candidate, err := time.Parse("2006-01-02", today.Format("2006")+"-"+monthDay)
		if err != nil {
			return ""
		}

		if !candidate.Before(todayTrunc) {
			return candidate.Format("2006-01-02")
		}

		next, err := time.Parse("2006-01-02", strconv.Itoa(today.Year()+1)+"-"+monthDay)
		if err != nil {
			return ""
		}

		return next.Format("2006-01-02")
	}

	if exact.Before(todayTrunc) {
		return ""
	}

	return exact.Format("2006-01-02")
}
