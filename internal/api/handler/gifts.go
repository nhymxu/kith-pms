package handler

import (
	"database/sql"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/gifts"
)

type GiftsAPI struct {
	Svc             *gifts.Service
	GiftStoragePath string
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

	g, err := h.Svc.GetByIDWithPerson(c.Request().Context(), id)
	if err == sql.ErrNoRows {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, g)
}

const maxGiftImageBytes = 5 * 1024 * 1024

var giftImageAllowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// UploadImage handles POST /v1/gifts/:id/image
func (h *GiftsAPI) UploadImage(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, maxGiftImageBytes+1024*1024)

	file, err := c.FormFile("image")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			return apiErr(c, http.StatusRequestEntityTooLarge, "file too large (max 5MB)")
		}

		return apiErr(c, http.StatusBadRequest, "no file uploaded")
	}

	if file.Size > maxGiftImageBytes {
		return apiErr(c, http.StatusRequestEntityTooLarge, "file too large (max 5MB)")
	}

	src, err := file.Open()
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "failed to open file")
	}
	defer func() { _ = src.Close() }()

	detected, err := sniff512(src)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "failed to read file")
	}

	detected = strings.TrimSpace(strings.SplitN(detected, ";", 2)[0])
	if !giftImageAllowedTypes[detected] {
		return apiErr(c, http.StatusUnprocessableEntity, "unsupported file type: use jpeg, png, gif, or webp")
	}

	if err := h.Svc.UploadImage(c.Request().Context(), id, src, file); err != nil {
		if strings.Contains(err.Error(), "file service not configured") {
			return apiErr(c, http.StatusInternalServerError, "file storage not configured")
		}

		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]any{"uploaded": true})
}

// DeleteImage handles DELETE /v1/gifts/:id/image
func (h *GiftsAPI) DeleteImage(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.Svc.DeleteImage(c.Request().Context(), id); err != nil {
		if err == sql.ErrNoRows {
			return apiErr(c, http.StatusNotFound, "not found")
		}

		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return noContent(c)
}

// GetImage handles GET /v1/gifts/:id/image
func (h *GiftsAPI) GetImage(c *echo.Context) error {
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

	if g.ImagePath == "" {
		return apiErr(c, http.StatusNotFound, "no image")
	}

	fullPath := filepath.Join(h.GiftStoragePath, g.ImagePath)

	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(h.GiftStoragePath)) {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	f, err := os.Open(cleanPath)
	if err != nil {
		return apiErr(c, http.StatusNotFound, "image file not found")
	}
	defer func() { _ = f.Close() }()

	mt := g.ImageMimeType
	if mt == "" {
		mt = "application/octet-stream"
	}

	c.Response().Header().Set("Content-Type", mt)
	c.Response().Header().Set("Cache-Control", "public, max-age=86400")
	_, err = io.Copy(c.Response(), f)

	return err
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
