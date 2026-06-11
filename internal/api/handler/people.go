package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/people"
)

type PeopleAPI struct {
	Svc *people.Service
}

type personRequest struct {
	Name             string            `json:"name"`
	Nickname         string            `json:"nickname"`
	Gender           string            `json:"gender"` // "" | "male" | "female" | "rather_not_say"
	RelationshipType string            `json:"relationship_type"`
	DateOfBirth      string            `json:"date_of_birth"` // "YYYY-MM-DD" or ""
	OtherNotes       string            `json:"other_notes"`
	Contacts         []contactRequest  `json:"contacts"`
	Locations        []locationRequest `json:"locations"`
}

type contactRequest struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	Label    string `json:"label"`
	Position int    `json:"position"`
}

type locationRequest struct {
	Type       string `json:"type"`
	Address    string `json:"address"`
	City       string `json:"city"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
	Position   int    `json:"position"`
}

// List godoc
//
// @Summary      List people
// @Tags         people
// @Produce      json
// @Param        q          query     string  false  "Search query"
// @Param        page       query     int     false  "Page number"   default(1)
// @Param        page_size  query     int     false  "Page size"     default(50)
// @Param        labels      query     string  false  "Comma-separated label IDs"
// @Param        has_journal query     bool    false  "Only return people linked to at least one journal entry"
// @Success      200  {object}  envelope
// @Failure      500  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people [get]
func (h *PeopleAPI) List(c *echo.Context) error {
	q := c.QueryParam("q")

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 {
		pageSize = 50
	}

	if pageSize > 500 {
		pageSize = 500
	}

	var labelIDs []int64

	if raw := c.QueryParam("labels"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			id, err := strconv.ParseInt(part, 10, 64)
			if err == nil {
				labelIDs = append(labelIDs, id)
			}
		}
	}

	hasJournal := c.QueryParam("has_journal") == "true"

	list, err := h.Svc.List(c.Request().Context(), people.ListParams{
		Query:      q,
		Page:       page,
		PageSize:   pageSize,
		LabelIDs:   labelIDs,
		HasJournal: hasJournal,
	})
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, list)
}

// Get godoc
//
// @Summary      Get person
// @Tags         people
// @Produce      json
// @Param        id   path      int  true  "Person ID"
// @Success      200  {object}  envelope
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id} [get]
func (h *PeopleAPI) Get(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	p, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if p == nil {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	return ok(c, p)
}

// Create godoc
//
// @Summary      Create person
// @Tags         people
// @Accept       json
// @Produce      json
// @Param        body  body      personRequest  true  "Person data"
// @Success      201   {object}  envelope{data=object{id=int}}
// @Failure      400   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people [post]
func (h *PeopleAPI) Create(c *echo.Context) error {
	var req personRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Name) == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "name is required")
	}

	if req.DateOfBirth != "" {
		if _, err := time.Parse("2006-01-02", req.DateOfBirth); err != nil {
			return apiErr(c, http.StatusUnprocessableEntity, "date_of_birth must be YYYY-MM-DD")
		}
	}

	p, contacts, locations := mapPersonRequest(0, req)

	id, err := h.Svc.Create(c.Request().Context(), p, contacts, locations)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return created(c, map[string]any{"id": id})
}

// Update godoc
//
// @Summary      Update person
// @Tags         people
// @Accept       json
// @Produce      json
// @Param        id    path      int            true  "Person ID"
// @Param        body  body      personRequest  true  "Person data"
// @Success      200   {object}  envelope{data=object{id=int}}
// @Failure      400   {object}  envelope
// @Failure      404   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id} [put]
func (h *PeopleAPI) Update(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	existing, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if existing == nil {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	var req personRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if strings.TrimSpace(req.Name) == "" {
		return apiErr(c, http.StatusUnprocessableEntity, "name is required")
	}

	if req.DateOfBirth != "" {
		if _, err := time.Parse("2006-01-02", req.DateOfBirth); err != nil {
			return apiErr(c, http.StatusUnprocessableEntity, "date_of_birth must be YYYY-MM-DD")
		}
	}

	p, contacts, locations := mapPersonRequest(id, req)

	if err := h.Svc.Update(c.Request().Context(), p, contacts, locations); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]any{"id": id})
}

// Delete godoc
//
// @Summary      Delete person
// @Tags         people
// @Produce      json
// @Param        id   path  int  true  "Person ID"
// @Success      204
// @Failure      400  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id} [delete]
func (h *PeopleAPI) Delete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	p, err := h.Svc.Get(c.Request().Context(), id)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if p == nil {
		return apiErr(c, http.StatusNotFound, "not found")
	}

	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return noContent(c)
}

func parseID(c *echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func mapPersonRequest(id int64, req personRequest) (people.Person, []people.ContactInfo, []people.Location) {
	p := people.Person{
		ID:               id,
		Name:             req.Name,
		Nickname:         req.Nickname,
		Gender:           req.Gender,
		RelationshipType: req.RelationshipType,
		OtherNotes:       req.OtherNotes,
	}

	if req.DateOfBirth != "" {
		t, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err == nil {
			p.DateOfBirth = &t
		}
	}

	contacts := make([]people.ContactInfo, 0, len(req.Contacts))
	for _, c := range req.Contacts {
		contacts = append(contacts, people.ContactInfo{
			Type:     c.Type,
			Value:    c.Value,
			Label:    c.Label,
			Position: c.Position,
		})
	}

	locations := make([]people.Location, 0, len(req.Locations))
	for _, l := range req.Locations {
		locations = append(locations, people.Location{
			Type:       l.Type,
			Address:    l.Address,
			City:       l.City,
			Country:    l.Country,
			PostalCode: l.PostalCode,
			Position:   l.Position,
		})
	}

	return p, contacts, locations
}
