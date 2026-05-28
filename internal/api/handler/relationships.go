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
func (h *RelationshipsAPI) ListTypes(c *echo.Context) error {
	types, err := h.Svc.ListTypesWithCounts(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, types)
}

// CreateType handles POST /v1/relationship-types.
// Body: {"name": "...", "reverse_name": "..."}.
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
// Body: {"name": "...", "reverse_name": "..."}.
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
// Body: {"to_person_id": 123, "relationship_type_id": 1, "notes": "..."}.
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
