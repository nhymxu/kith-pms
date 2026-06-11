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
	LabelIDs       []int64 `json:"label_ids"`
}

// List godoc
//
// @Summary      List journal entries
// @Tags         journal
// @Produce      json
// @Param        q           query  string  false  "Full-text search query"
// @Param        page        query  int     false  "Page number"   default(1)
// @Param        page_size   query  int     false  "Page size"     default(50)
// @Param        from_date   query  string  false  "From date YYYY-MM-DD"
// @Param        to_date     query  string  false  "To date YYYY-MM-DD"
// @Param        person_ids  query  string  false  "Comma-separated person IDs"
// @Success      200  {object}  envelope
// @Failure      500  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /journal [get]
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

	if raw := c.QueryParam("labels"); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
			if err == nil && id > 0 {
				params.JournalLabelIDs = append(params.JournalLabelIDs, id)
			}
		}
	}

	list, err := h.Svc.List(c.Request().Context(), params)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, list)
}

// Get godoc
//
// @Summary      Get journal entry
// @Tags         journal
// @Produce      json
// @Param        id   path      int  true  "Entry ID"
// @Success      200  {object}  envelope
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /journal/{id} [get]
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

// Create godoc
//
// @Summary      Create journal entry
// @Tags         journal
// @Accept       json
// @Produce      json
// @Param        body  body      journalRequest  true  "Journal entry"
// @Success      201   {object}  envelope{data=object{id=int}}
// @Failure      400   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /journal [post]
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
		OccurredAtTime: nullableString(req.OccurredAtTime),
	}

	id, err := h.Svc.Create(c.Request().Context(), a, req.PersonIDs, req.LabelIDs)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return created(c, map[string]any{"id": id})
}

// Update godoc
//
// @Summary      Update journal entry
// @Tags         journal
// @Accept       json
// @Produce      json
// @Param        id    path      int             true  "Entry ID"
// @Param        body  body      journalRequest  true  "Journal entry"
// @Success      200   {object}  envelope{data=object{id=int}}
// @Failure      400   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /journal/{id} [put]
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
		OccurredAtTime: nullableString(req.OccurredAtTime),
	}

	if err := h.Svc.Update(c.Request().Context(), a, req.PersonIDs, req.LabelIDs); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]any{"id": id})
}

// Delete godoc
//
// @Summary      Delete journal entry
// @Tags         journal
// @Produce      json
// @Param        id   path  int  true  "Entry ID"
// @Success      204
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /journal/{id} [delete]
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
