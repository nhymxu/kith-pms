package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/relationships"
)

// RelationshipsAPI handles /v1/relationship-types and /v1/people/:id/relationships.
type RelationshipsAPI struct {
	Svc *relationships.Service
}

// ListTypes handles GET /v1/relationship-types.
//
// @Summary      List relationship types
// @Tags         relationships
// @Produce      json
// @Success      200  {object}  envelope
// @Failure      500  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /relationship-types [get]
func (h *RelationshipsAPI) ListTypes(c *echo.Context) error {
	types, err := h.Svc.ListTypesWithCounts(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, types)
}

// CreateType handles POST /v1/relationship-types.
//
// @Summary      Create relationship type
// @Tags         relationships
// @Accept       json
// @Produce      json
// @Param        body  body      object{name=string,reverse_name=string}  true  "Relationship type"
// @Success      201   {object}  envelope
// @Failure      400   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /relationship-types [post]
func (h *RelationshipsAPI) CreateType(c *echo.Context) error {
	var req struct {
		Name        string `json:"name"`
		ReverseName string `json:"reverse_name"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	rt, err := h.Svc.CreateType(c.Request().Context(), req.Name, req.ReverseName)
	if err != nil {
		return relTypeErrResponse(c, err)
	}

	return created(c, rt)
}

// UpdateType handles PUT /v1/relationship-types/:id.
//
// @Summary      Update relationship type
// @Tags         relationships
// @Accept       json
// @Produce      json
// @Param        id    path      int                                       true  "Type ID"
// @Param        body  body      object{name=string,reverse_name=string}  true  "Relationship type"
// @Success      200   {object}  envelope{data=object{id=int}}
// @Failure      400   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /relationship-types/{id} [put]
func (h *RelationshipsAPI) UpdateType(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	var req struct {
		Name        string `json:"name"`
		ReverseName string `json:"reverse_name"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.Svc.UpdateType(c.Request().Context(), id, req.Name, req.ReverseName); err != nil {
		return relTypeErrResponse(c, err)
	}

	return ok(c, map[string]any{"id": id})
}

// DeleteType handles DELETE /v1/relationship-types/:id.
//
// @Summary      Delete relationship type
// @Tags         relationships
// @Produce      json
// @Param        id   path  int  true  "Type ID"
// @Success      204
// @Failure      400  {object}  envelope
// @Failure      409  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /relationship-types/{id} [delete]
func (h *RelationshipsAPI) DeleteType(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.Svc.DeleteType(c.Request().Context(), id); err != nil {
		if errors.Is(err, relationships.ErrTypeInUse) {
			return apiErr(c, http.StatusConflict, "type is in use and cannot be deleted")
		}

		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return noContent(c)
}

// ListByPerson handles GET /v1/people/:id/relationships.
//
// @Summary      List relationships for person
// @Tags         relationships
// @Produce      json
// @Param        id   path      int  true  "Person ID"
// @Success      200  {object}  envelope
// @Failure      400  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id}/relationships [get]
func (h *RelationshipsAPI) ListByPerson(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	views, err := h.Svc.ListByPerson(c.Request().Context(), personID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, views)
}

// AttachRelationship handles POST /v1/people/:id/relationships.
//
// @Summary      Attach relationship to person
// @Tags         relationships
// @Accept       json
// @Produce      json
// @Param        id    path      int                                                              true  "Person ID"
// @Param        body  body      object{to_person_id=int,relationship_type_id=int,notes=string}  true  "Relationship data"
// @Success      201   {object}  envelope{data=object{id=int}}
// @Failure      400   {object}  envelope
// @Failure      409   {object}  envelope
// @Failure      422   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id}/relationships [post]
func (h *RelationshipsAPI) AttachRelationship(c *echo.Context) error {
	fromID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	var req struct {
		ToPersonID         int64  `json:"to_person_id"`
		RelationshipTypeID int64  `json:"relationship_type_id"`
		Notes              string `json:"notes"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if req.ToPersonID <= 0 {
		return apiErr(c, http.StatusUnprocessableEntity, "to_person_id is required")
	}

	if req.RelationshipTypeID <= 0 {
		return apiErr(c, http.StatusUnprocessableEntity, "relationship_type_id is required")
	}

	relID, err := h.Svc.AttachRelationship(
		c.Request().Context(),
		fromID,
		req.ToPersonID,
		req.RelationshipTypeID,
		req.Notes,
	)
	if err != nil {
		switch {
		case errors.Is(err, relationships.ErrDuplicateRelationship):
			return apiErr(c, http.StatusConflict, "relationship already exists")
		case errors.Is(err, relationships.ErrSelfRelationship):
			return apiErr(c, http.StatusUnprocessableEntity, "cannot relate a person to themselves")
		case errors.Is(err, relationships.ErrTypeNotFound):
			return apiErr(c, http.StatusUnprocessableEntity, "relationship type not found")
		default:
			return apiErr(c, http.StatusInternalServerError, "internal server error")
		}
	}

	return created(c, map[string]any{"id": relID})
}

// DetachRelationship handles DELETE /v1/people/:id/relationships/:relID.
//
// @Summary      Detach relationship from person
// @Tags         relationships
// @Produce      json
// @Param        id     path  int  true  "Person ID"
// @Param        relID  path  int  true  "Relationship ID"
// @Success      204
// @Failure      400  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /people/{id}/relationships/{relID} [delete]
func (h *RelationshipsAPI) DetachRelationship(c *echo.Context) error {
	relID, err := strconv.ParseInt(c.Param("relID"), 10, 64)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid relID")
	}

	if err := h.Svc.DetachRelationship(c.Request().Context(), relID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return noContent(c)
}

// Graph handles GET /v1/relationships/graph[?person_id=N].
//
// @Summary      Get relationship network graph
// @Tags         relationships
// @Produce      json
// @Param        person_id  query     int  false  "Ego-network focal person ID"
// @Success      200        {object}  envelope
// @Failure      400        {object}  envelope
// @Security     CookieAuth
// @Router       /relationships/graph [get]
func (h *RelationshipsAPI) Graph(c *echo.Context) error {
	var personID int64

	if raw := c.QueryParam("person_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			return apiErr(c, http.StatusBadRequest, "invalid person_id")
		}

		personID = id
	}

	graph, err := h.Svc.Graph(c.Request().Context(), personID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, graph)
}

// BulkAttach handles POST /v1/people/:id/relationships/bulk.
func (h *RelationshipsAPI) BulkAttach(c *echo.Context) error {
	fromID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid person id")
	}

	var req struct {
		Relationships []struct {
			ToPersonID int64  `json:"to_person_id"`
			TypeID     int64  `json:"relationship_type_id"`
			Notes      string `json:"notes"`
		} `json:"relationships"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if len(req.Relationships) == 0 {
		return apiErr(c, http.StatusBadRequest, "relationships required")
	}

	if len(req.Relationships) > 50 {
		return apiErr(c, http.StatusBadRequest, "max 50 relationships per request")
	}

	pairs := make([]relationships.BulkRelationshipPair, len(req.Relationships))
	for i, r := range req.Relationships {
		pairs[i] = relationships.BulkRelationshipPair{ToPersonID: r.ToPersonID, TypeID: r.TypeID, Notes: r.Notes}
	}

	created, skipped, err := h.Svc.BulkAttach(c.Request().Context(), fromID, pairs)
	if errors.Is(err, relationships.ErrTypeNotFound) {
		return apiErr(c, http.StatusUnprocessableEntity, "relationship type not found")
	}

	if errors.Is(err, relationships.ErrSelfRelationship) {
		return apiErr(c, http.StatusUnprocessableEntity, "cannot relate a person to themselves")
	}

	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]int{"created": created, "skipped": skipped})
}

// ---- helpers ----------------------------------------------------------------

func relTypeErrResponse(c *echo.Context, err error) error {
	switch {
	case errors.Is(err, relationships.ErrNameEmpty):
		return apiErr(c, http.StatusUnprocessableEntity, "name is required")
	case errors.Is(err, relationships.ErrNameTooLong):
		return apiErr(c, http.StatusUnprocessableEntity, "name must be 80 characters or fewer")
	default:
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}
}
