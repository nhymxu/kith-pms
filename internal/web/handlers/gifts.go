package handlers

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/gifts"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/web/templates"
)

type GiftsHandlers struct {
	Svc           *gifts.Service
	PeopleSvc     *people.Service
	ImageBasePath string
}

func (h *GiftsHandlers) GetList(c *echo.Context) error {
	direction := gifts.Direction(c.QueryParam("direction"))

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	params := gifts.ListParams{
		Direction: direction,
		PageSize:  50,
		Page:      page,
	}

	if pidStr := c.QueryParam("person_id"); pidStr != "" {
		if pid, err := strconv.ParseInt(pidStr, 10, 64); err == nil && pid > 0 {
			params.PersonID = &pid
		}
	}

	list, err := h.Svc.List(c.Request().Context(), params)
	if err != nil {
		return err
	}

	component := templates.GiftsList(templates.GiftsListParams{
		Gifts:     list,
		Direction: string(direction),
		Page:      page,
		HasMore:   len(list) == 50,
		CSRFToken: auth.CSRFToken(c),
	})

	return component.Render(c.Request().Context(), c.Response())
}

func (h *GiftsHandlers) GetNew(c *echo.Context) error {
	allPeople, _ := h.PeopleSvc.List(c.Request().Context(), people.ListParams{PageSize: 1000, Page: 1})

	// Pre-select person_id from query param (e.g. linked from people detail page)
	var preselected int64
	if pidStr := c.QueryParam("person_id"); pidStr != "" {
		preselected, _ = strconv.ParseInt(pidStr, 10, 64)
	}

	component := templates.GiftsForm(templates.GiftsFormParams{
		Gift:      gifts.Gift{PersonID: preselected},
		CSRFToken: auth.CSRFToken(c),
		AllPeople: allPeople,
	})

	return component.Render(c.Request().Context(), c.Response())
}

func (h *GiftsHandlers) PostCreate(c *echo.Context) error {
	g, formErr := parseGiftForm(c)
	if formErr != "" {
		allPeople, _ := h.PeopleSvc.List(c.Request().Context(), people.ListParams{PageSize: 1000, Page: 1})

		return templates.GiftsForm(templates.GiftsFormParams{
			Gift:      *g,
			CSRFToken: auth.CSRFToken(c),
			IsEdit:    false,
			Error:     formErr,
			AllPeople: allPeople,
		}).Render(c.Request().Context(), c.Response())
	}

	id, err := h.Svc.Create(c.Request().Context(), g)
	if err != nil {
		return err
	}

	if src, file, ferr := c.Request().FormFile("image"); ferr == nil {
		defer func() { _ = src.Close() }()

		_ = h.Svc.UploadImage(c.Request().Context(), id, src, file)
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/gifts/%d", id))
}

func (h *GiftsHandlers) GetDetail(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	g, err := h.Svc.GetByID(c.Request().Context(), id)
	if err == sql.ErrNoRows {
		return echo.ErrNotFound
	}

	if err != nil {
		return err
	}

	var personName string
	if p, err := h.PeopleSvc.Get(c.Request().Context(), g.PersonID); err == nil && p != nil {
		personName = p.Name
	}

	return templates.GiftsDetail(templates.GiftsDetailParams{
		Gift:       *g,
		PersonName: personName,
	}).Render(c.Request().Context(), c.Response())
}

func (h *GiftsHandlers) GetEdit(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	g, err := h.Svc.GetByID(c.Request().Context(), id)
	if err == sql.ErrNoRows {
		return echo.ErrNotFound
	}

	if err != nil {
		return err
	}

	allPeople, _ := h.PeopleSvc.List(c.Request().Context(), people.ListParams{PageSize: 1000, Page: 1})

	return templates.GiftsForm(templates.GiftsFormParams{
		Gift:      *g,
		CSRFToken: auth.CSRFToken(c),
		IsEdit:    true,
		AllPeople: allPeople,
	}).Render(c.Request().Context(), c.Response())
}

func (h *GiftsHandlers) PostUpdate(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	g, formErr := parseGiftForm(c)
	if formErr != "" {
		allPeople, _ := h.PeopleSvc.List(c.Request().Context(), people.ListParams{PageSize: 1000, Page: 1})
		g.ID = id

		return templates.GiftsForm(templates.GiftsFormParams{
			Gift:      *g,
			CSRFToken: auth.CSRFToken(c),
			IsEdit:    true,
			Error:     formErr,
			AllPeople: allPeople,
		}).Render(c.Request().Context(), c.Response())
	}

	g.ID = id

	if err := h.Svc.Update(c.Request().Context(), g); err != nil {
		return err
	}

	if src, file, ferr := c.Request().FormFile("image"); ferr == nil {
		defer func() { _ = src.Close() }()

		_ = h.Svc.UploadImage(c.Request().Context(), id, src, file)
	} else if c.FormValue("remove_image") == "1" {
		_ = h.Svc.DeleteImage(c.Request().Context(), id)
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/gifts/%d", id))
}

func (h *GiftsHandlers) GetDeleteConfirm(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	g, err := h.Svc.GetByID(c.Request().Context(), id)
	if err == sql.ErrNoRows {
		return echo.ErrNotFound
	}

	if err != nil {
		return err
	}

	return templates.GiftDeleteConfirm(*g, auth.CSRFToken(c)).Render(c.Request().Context(), c.Response())
}

func (h *GiftsHandlers) PostDelete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/gifts")
}

func (h *GiftsHandlers) GetImage(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return echo.ErrNotFound
	}

	g, err := h.Svc.GetByID(c.Request().Context(), id)
	if err == sql.ErrNoRows || g == nil || !g.HasImage() {
		return echo.ErrNotFound
	}

	if err != nil {
		return err
	}

	fullPath := filepath.Join(h.ImageBasePath, g.ImagePath)

	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(h.ImageBasePath)) {
		return echo.ErrForbidden
	}

	mimeType := g.ImageMimeType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	c.Response().Header().Set("Content-Type", mimeType)
	c.Response().Header().Set("Cache-Control", "private, max-age=3600")

	return c.File(fullPath)
}

// parseGiftForm reads and validates the gift HTML form.
func parseGiftForm(c *echo.Context) (*gifts.Gift, string) {
	title := strings.TrimSpace(c.FormValue("title"))
	if title == "" {
		return &gifts.Gift{}, "Title is required"
	}

	personIDStr := c.FormValue("person_id")

	personID, err := strconv.ParseInt(personIDStr, 10, 64)
	if err != nil || personID <= 0 {
		return &gifts.Gift{Title: title}, "Person is required"
	}

	direction := gifts.Direction(c.FormValue("direction"))
	if direction != gifts.DirectionGiven && direction != gifts.DirectionReceived &&
		direction != gifts.DirectionPlanned {
		direction = gifts.DirectionPlanned
	}

	date := strings.TrimSpace(c.FormValue("date"))
	notes := strings.TrimSpace(c.FormValue("notes"))

	currency := strings.TrimSpace(c.FormValue("currency"))
	if currency == "" {
		currency = "USD"
	}

	debtType := gifts.DebtType(c.FormValue("debt_type"))
	if debtType != gifts.DebtIOwe && debtType != gifts.DebtTheyOwe {
		debtType = gifts.DebtNone
	}

	var amountCents *int64

	if amtStr := strings.TrimSpace(c.FormValue("amount")); amtStr != "" {
		f, err := strconv.ParseFloat(amtStr, 64)
		if err == nil && f >= 0 {
			cents := int64(math.Round(f * 100))
			amountCents = &cents
		}
	}

	return &gifts.Gift{
		PersonID:    personID,
		Title:       title,
		Direction:   direction,
		Date:        date,
		Notes:       notes,
		AmountCents: amountCents,
		Currency:    currency,
		DebtType:    debtType,
	}, ""
}
