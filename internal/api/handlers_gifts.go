package api

import (
	"database/sql"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/gifts"
)

type GiftsAPI struct {
	Svc *gifts.Service
}

type giftRequest struct {
	PersonID    int64  `json:"person_id"`
	Title       string `json:"title"`
	Direction   string `json:"direction"`
	Date        string `json:"date"`
	Notes       string `json:"notes"`
	AmountCents *int64 `json:"amount_cents"`
	Currency    string `json:"currency"`
	DebtType    string `json:"debt_type"`
}

func (h *GiftsAPI) List(c *echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	var personID *int64

	if pidStr := c.QueryParam("person_id"); pidStr != "" {
		if pid, err := strconv.ParseInt(pidStr, 10, 64); err == nil {
			personID = &pid
		}
	}

	list, err := h.Svc.List(c.Request().Context(), gifts.ListParams{
		Direction: gifts.Direction(c.QueryParam("direction")),
		DebtType:  gifts.DebtType(c.QueryParam("debt_type")),
		PersonID:  personID,
		PageSize:  50,
		Page:      page,
	})
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, list)
}

func (h *GiftsAPI) Create(c *echo.Context) error {
	var req giftRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Title) == "" {
		return apiErr(c, http.StatusBadRequest, "title is required")
	}

	if req.PersonID <= 0 {
		return apiErr(c, http.StatusBadRequest, "person_id is required")
	}

	g := giftFromRequest(req)

	id, err := h.Svc.Create(c.Request().Context(), g)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	g.ID = id

	return created(c, g)
}

// GetByID handles GET /v1/gifts/:id
func (h *GiftsAPI) GetByID(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	g, err := h.Svc.GetByID(c.Request().Context(), id)
	if err == sql.ErrNoRows {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, g)
}

func (h *GiftsAPI) Update(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	var req giftRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Title) == "" {
		return apiErr(c, http.StatusBadRequest, "title is required")
	}

	g := giftFromRequest(req)

	g.ID = id
	if err := h.Svc.Update(c.Request().Context(), g); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, g)
}

func (h *GiftsAPI) Delete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return noContent(c)
}

// ListByPerson handles GET /v1/people/:id/gifts
func (h *GiftsAPI) ListByPerson(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	list, err := h.Svc.List(c.Request().Context(), gifts.ListParams{
		PersonID: &personID,
		PageSize: 200,
		Page:     1,
	})
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, list)
}

func giftFromRequest(req giftRequest) *gifts.Gift {
	direction := gifts.Direction(req.Direction)
	if direction != gifts.DirectionGiven && direction != gifts.DirectionReceived &&
		direction != gifts.DirectionPlanned {
		direction = gifts.DirectionPlanned
	}

	debtType := gifts.DebtType(req.DebtType)
	if debtType != gifts.DebtIOwe && debtType != gifts.DebtTheyOwe {
		debtType = gifts.DebtNone
	}

	currency := strings.TrimSpace(req.Currency)
	if currency == "" {
		currency = "USD"
	}

	// Normalize amount_cents: if caller sends a float amount field, convert.
	amountCents := req.AmountCents
	if amountCents != nil {
		v := int64(math.Abs(float64(*amountCents)))
		amountCents = &v
	}

	return &gifts.Gift{
		PersonID:    req.PersonID,
		Title:       strings.TrimSpace(req.Title),
		Direction:   direction,
		Date:        req.Date,
		Notes:       req.Notes,
		AmountCents: amountCents,
		Currency:    currency,
		DebtType:    debtType,
	}
}
