package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/journal"
)

type JournalAPI struct {
	Svc *journal.Service
}

// journalRequest is the JSON body for create and update.
type journalRequest struct {
	Title          string  `json:"title"`
	Content        string  `json:"content"`
	OccurredAtDate string  `json:"occurred_at_date"` // "YYYY-MM-DD"
	OccurredAtTime string  `json:"occurred_at_time"` // "HH:MM" or ""
	PersonIDs      []int64 `json:"person_ids"`
}

func (h *JournalAPI) List(c *echo.Context) error {
	q := c.QueryParam("q")

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 {
		pageSize = 50
	}

	params := journal.ListParams{
		Query:    q,
		Page:     page,
		PageSize: pageSize,
		FromDate: c.QueryParam("from_date"),
		ToDate:   c.QueryParam("to_date"),
	}

	if raw := c.QueryParam("person_ids"); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
			if err == nil && id > 0 {
				params.PersonIDs = append(params.PersonIDs, id)
			}
		}
	}

	list, err := h.Svc.List(c.Request().Context(), params)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, list)
}

func (h *JournalAPI) Get(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	a, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if a == nil {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	return ok(c, a)
}

func (h *JournalAPI) Create(c *echo.Context) error {
	var req journalRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Title) == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "title is required")
	}

	a := journal.Activity{
		Title:          req.Title,
		Content:        req.Content,
		OccurredAtDate: req.OccurredAtDate,
		OccurredAtTime: req.OccurredAtTime,
	}

	id, err := h.Svc.Create(c.Request().Context(), a, req.PersonIDs)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return created(c, map[string]any{"id": id})
}

func (h *JournalAPI) Update(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	var req journalRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Title) == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "title is required")
	}

	a := journal.Activity{
		ID:             id,
		Title:          req.Title,
		Content:        req.Content,
		OccurredAtDate: req.OccurredAtDate,
		OccurredAtTime: req.OccurredAtTime,
	}

	if err := h.Svc.Update(c.Request().Context(), a, req.PersonIDs); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]any{"id": id})
}

func (h *JournalAPI) Delete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	// Check existence before delete.
	a, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if a == nil {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return noContent(c)
}
