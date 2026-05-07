package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/audit"
)

// AuditAPI handles GET /v1/audit.
type AuditAPI struct {
	Svc *audit.Service
}

// List returns paginated audit entries, optionally filtered by entity_type and entity_id.
func (h *AuditAPI) List(c *echo.Context) error {
	if h.Svc == nil {
		return apiErr(c, http.StatusNotImplemented, "audit not configured")
	}

	entityType := audit.EntityType(c.QueryParam("entity_type"))
	entityID, _ := strconv.ParseInt(c.QueryParam("entity_id"), 10, 64)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	entries, err := h.Svc.List(c.Request().Context(), audit.ListParams{
		EntityType: entityType,
		EntityID:   entityID,
		Page:       page,
		PageSize:   50,
	})
	if err != nil {
		return err
	}

	type row struct {
		ID         int64            `json:"id"`
		EntityType audit.EntityType `json:"entity_type"`
		EntityID   int64            `json:"entity_id"`
		EntityName string           `json:"entity_name"`
		Action     audit.Action     `json:"action"`
		ActorID    *int64           `json:"actor_id"`
		CreatedAt  string           `json:"created_at"`
	}

	out := make([]row, 0, len(entries))
	for _, e := range entries {
		out = append(out, row{
			ID:         e.ID,
			EntityType: e.EntityType,
			EntityID:   e.EntityID,
			EntityName: e.EntityName,
			Action:     e.Action,
			ActorID:    e.ActorID,
			CreatedAt:  e.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data":      out,
		"page":      page,
		"page_size": 50,
		"has_more":  len(entries) == 50,
	})
}
