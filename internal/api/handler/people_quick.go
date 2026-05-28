package handler

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/gifts"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/people"
)

// PeopleQuickAPI handles quick-action endpoints under /v1/people/:id/*.
type PeopleQuickAPI struct {
	PeopleSvc  *people.Service
	JournalSvc *journal.Service
	GiftsSvc   *gifts.Service
}

// QuickJournal handles POST /v1/people/:id/journal/quick.
// Body: {"title":"...","occurred_at_date":"YYYY-MM-DD","occurred_at_time":"HH:MM","content":"...","person_ids":[...]}.
func (h *PeopleQuickAPI) QuickJournal(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	if h.JournalSvc == nil {
		return apiErr(c, http.StatusInternalServerError, "journal service not configured")
	}

	// Verify person exists.
	p, err := h.PeopleSvc.Get(c.Request().Context(), personID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if p == nil {
		return apiErr(c, http.StatusNotFound, "person not found")
	}

	var req struct {
		Title          string  `json:"title"`
		OccurredAtDate string  `json:"occurred_at_date"`
		OccurredAtTime string  `json:"occurred_at_time"`
		Content        string  `json:"content"`
		PersonIDs      []int64 `json:"person_ids"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "title is required")
	}

	today := time.Now().Format("2006-01-02")
	if req.OccurredAtDate == "" {
		req.OccurredAtDate = today
	}

	if _, err := time.Parse("2006-01-02", req.OccurredAtDate); err != nil {
		return apiErr(c, http.StatusUnprocessableEntity, "invalid date format (use YYYY-MM-DD)")
	}

	activity := journal.Activity{
		Title:          req.Title,
		OccurredAtDate: req.OccurredAtDate,
		OccurredAtTime: strings.TrimSpace(req.OccurredAtTime),
		Content:        strings.TrimSpace(req.Content),
	}

	// Always include the current person; deduplicate.
	seen := make(map[int64]bool, len(req.PersonIDs)+1)
	seen[personID] = true

	personIDs := []int64{personID}

	for _, id := range req.PersonIDs {
		if !seen[id] {
			seen[id] = true
			personIDs = append(personIDs, id)
		}
	}

	id, err := h.JournalSvc.Create(c.Request().Context(), activity, personIDs)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "failed to save journal entry")
	}

	return created(c, map[string]any{"id": id})
}

// QuickGift handles POST /v1/people/:id/gifts/quick.
// Body: {"title":"...","direction":"planned|given|received","date":"YYYY-MM-DD",
//
//	"notes":"...","amount":"12.50","currency":"USD","debt_type":"i_owe|they_owe|"}.
func (h *PeopleQuickAPI) QuickGift(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	if h.GiftsSvc == nil {
		return apiErr(c, http.StatusInternalServerError, "gifts service not configured")
	}

	// Verify person exists.
	p, err := h.PeopleSvc.Get(c.Request().Context(), personID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if p == nil {
		return apiErr(c, http.StatusNotFound, "person not found")
	}

	var req struct {
		Title     string `json:"title"`
		Direction string `json:"direction"`
		Date      string `json:"date"`
		Notes     string `json:"notes"`
		Amount    string `json:"amount"`
		Currency  string `json:"currency"`
		DebtType  string `json:"debt_type"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "title is required")
	}

	direction := gifts.Direction(strings.TrimSpace(req.Direction))
	if direction == "" {
		direction = gifts.DirectionPlanned
	}

	currency := strings.TrimSpace(req.Currency)
	if currency == "" {
		currency = "USD"
	}

	debtType := gifts.DebtType(req.DebtType)
	if debtType != gifts.DebtIOwe && debtType != gifts.DebtTheyOwe {
		debtType = gifts.DebtNone
	}

	var amountCents *int64

	if amtStr := strings.TrimSpace(req.Amount); amtStr != "" {
		if f, parseErr := strconv.ParseFloat(amtStr, 64); parseErr == nil && f >= 0 {
			cents := int64(math.Round(f * 100))
			amountCents = &cents
		}
	}

	g := &gifts.Gift{
		PersonID:    personID,
		Title:       req.Title,
		Direction:   direction,
		Date:        strings.TrimSpace(req.Date),
		Notes:       strings.TrimSpace(req.Notes),
		AmountCents: amountCents,
		Currency:    currency,
		DebtType:    debtType,
	}

	giftID, err := h.GiftsSvc.Create(c.Request().Context(), g)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "failed to save gift")
	}

	return created(c, map[string]any{"id": giftID})
}

// UpdateLastContact handles POST /v1/people/:id/last-contact.
// Records now as the last contact timestamp for the person.
func (h *PeopleQuickAPI) UpdateLastContact(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.PeopleSvc.UpdateLastContact(c.Request().Context(), personID, time.Now().UTC()); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return apiErr(c, http.StatusNotFound, "person not found")
		}

		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]any{"updated": true})
}
