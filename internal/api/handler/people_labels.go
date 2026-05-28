package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/labels"
)

// PeopleLabelsAPI handles /v1/people/:id/labels endpoints.
type PeopleLabelsAPI struct {
	Svc *labels.Service
}

// Attach handles POST /v1/people/:id/labels.
// Body: {"label_id": 123}.
func (h *PeopleLabelsAPI) Attach(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	var req struct {
		LabelID int64 `json:"label_id"`
	}
	if err := c.Bind(&req); err != nil || req.LabelID <= 0 {
		return apiErr(c, http.StatusBadRequest, "label_id is required")
	}

	if err := h.Svc.Attach(c.Request().Context(), personID, req.LabelID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]any{"attached": true})
}

// Detach handles DELETE /v1/people/:id/labels/:labelID.
func (h *PeopleLabelsAPI) Detach(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	labelID, err := strconv.ParseInt(c.Param("labelID"), 10, 64)
	if err != nil || labelID <= 0 {
		return apiErr(c, http.StatusBadRequest, "invalid labelID")
	}

	if err := h.Svc.Detach(c.Request().Context(), personID, labelID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return noContent(c)
}
