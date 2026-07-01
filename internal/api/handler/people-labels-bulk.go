package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/relationships"
)

// PeopleLabelsBulkAPI handles bulk label and mesh-connect operations.
type PeopleLabelsBulkAPI struct {
	Svc       *people.LabelService
	PeopleSvc *people.Service
	RelSvc    *relationships.Service
}

// BulkAssign handles POST /v1/people-labels/bulk-assign.
func (h *PeopleLabelsBulkAPI) BulkAssign(c *echo.Context) error {
	var req struct {
		LabelID   int64   `json:"label_id"`
		PersonIDs []int64 `json:"person_ids"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if req.LabelID == 0 || len(req.PersonIDs) == 0 {
		return apiErr(c, http.StatusBadRequest, "label_id and person_ids required")
	}

	if len(req.PersonIDs) > 500 {
		return apiErr(c, http.StatusBadRequest, "max 500 people per request")
	}

	label, err := h.Svc.Get(c.Request().Context(), req.LabelID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if label == nil {
		return apiErr(c, http.StatusUnprocessableEntity, "label not found")
	}

	// Deduplicate so already_had_label count is accurate.
	seen := make(map[int64]struct{}, len(req.PersonIDs))

	deduped := req.PersonIDs[:0]
	for _, id := range req.PersonIDs {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			deduped = append(deduped, id)
		}
	}

	req.PersonIDs = deduped

	invalid, err := h.PeopleSvc.ValidatePeopleExist(c.Request().Context(), req.PersonIDs)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if len(invalid) > 0 {
		return c.JSON(http.StatusUnprocessableEntity, map[string]any{
			"error":       "invalid person ids",
			"invalid_ids": invalid,
		})
	}

	attached, err := h.Svc.BulkAttach(c.Request().Context(), req.LabelID, req.PersonIDs)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]int{
		"attached":          attached,
		"already_had_label": len(req.PersonIDs) - attached,
	})
}

// ConnectAllPreview handles GET /v1/people-labels/:id/connect-all/preview.
func (h *PeopleLabelsBulkAPI) ConnectAllPreview(c *echo.Context) error {
	labelID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid label id")
	}

	ids, err := h.Svc.ListPersonIDsByLabelID(c.Request().Context(), labelID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	n := len(ids)

	return ok(c, map[string]int{"member_count": n, "pair_count": n * (n - 1)})
}

// ConnectAll handles POST /v1/people-labels/:id/connect-all.
func (h *PeopleLabelsBulkAPI) ConnectAll(c *echo.Context) error {
	labelID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid label id")
	}

	var req struct {
		RelationshipTypeID int64 `json:"relationship_type_id"`
	}
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	if req.RelationshipTypeID == 0 {
		return apiErr(c, http.StatusBadRequest, "relationship_type_id required")
	}

	created, skipped, total, err := h.RelSvc.BulkAttachMesh(c.Request().Context(), labelID, req.RelationshipTypeID)
	if errors.Is(err, relationships.ErrTypeNotFound) {
		return apiErr(c, http.StatusUnprocessableEntity, "relationship type not found")
	}

	if errors.Is(err, relationships.ErrMeshTooLarge) {
		return apiErr(c, http.StatusUnprocessableEntity, "label has too many members for mesh connect (max 500)")
	}

	if errors.Is(err, relationships.ErrAsymmetricTypeNotAllowed) {
		return apiErr(
			c,
			http.StatusUnprocessableEntity,
			"asymmetric relationship types cannot be used for mesh connect",
		)
	}

	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]int{"created": created, "skipped": skipped, "total_members": total})
}
