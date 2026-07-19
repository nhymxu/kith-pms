package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/note"
)

type NoteAPI struct {
	Svc *note.Service
}

type noteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ListByPerson handles GET /v1/people/:id/notes
//
// @Summary      List notes for person
// @Tags         notes
// @Produce      json
// @Param        id         path   int     true   "Person ID"
// @Param        page       query  int     false  "Page number"  default(1)
// @Param        page_size  query  int     false  "Page size"    default(50)
// @Success      200  {object}  envelope
// @Failure      400  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id}/notes [get]
func (h *NoteAPI) ListByPerson(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 {
		pageSize = 50
	}

	if pageSize > 200 {
		pageSize = 200
	}

	list, err := h.Svc.ListByPerson(c.Request().Context(), personID, page, pageSize)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, list)
}

// Create handles POST /v1/people/:id/notes
//
// @Summary      Create note
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        id    path      int          true  "Person ID"
// @Param        body  body      noteRequest  true  "Note data"
// @Success      201   {object}  envelope
// @Failure      400   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id}/notes [post]
func (h *NoteAPI) Create(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	var req noteRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Content) == "" {
		return apiErr(c, http.StatusBadRequest, "content is required")
	}

	n := &note.Note{
		PersonID: personID,
		Title:    strings.TrimSpace(req.Title),
		Content:  req.Content,
	}

	id, err := h.Svc.Create(c.Request().Context(), n)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	n.ID = id

	return created(c, n)
}

// Get handles GET /v1/notes/:id
//
// @Summary      Get note
// @Tags         notes
// @Produce      json
// @Param        id   path      int  true  "Note ID"
// @Success      200  {object}  envelope
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /notes/{id} [get]
func (h *NoteAPI) Get(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	n, err := h.Svc.GetByID(c.Request().Context(), id)
	if err == sql.ErrNoRows {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, n)
}

// Update handles PUT /v1/notes/:id
//
// @Summary      Update note
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        id    path      int          true  "Note ID"
// @Param        body  body      noteRequest  true  "Note data"
// @Success      200   {object}  envelope
// @Failure      400   {object}  envelope
// @Failure      404   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /notes/{id} [put]
func (h *NoteAPI) Update(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	var req noteRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Content) == "" {
		return apiErr(c, http.StatusBadRequest, "content is required")
	}

	existing, err := h.Svc.GetByID(c.Request().Context(), id)
	if err == sql.ErrNoRows {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	existing.Title = strings.TrimSpace(req.Title)
	existing.Content = req.Content

	if err := h.Svc.Update(c.Request().Context(), existing); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, existing)
}

// Delete handles DELETE /v1/notes/:id
//
// @Summary      Delete note
// @Tags         notes
// @Produce      json
// @Param        id   path  int  true  "Note ID"
// @Success      204
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /notes/{id} [delete]
func (h *NoteAPI) Delete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	if _, err := h.Svc.GetByID(c.Request().Context(), id); err == sql.ErrNoRows {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return noContent(c)
}
