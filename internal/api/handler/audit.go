package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/audit"
	"github.com/nhymxu/kith-pms/internal/settings"
)

// AuditAPI handles audit endpoints.
type AuditAPI struct {
	Svc         *audit.Service
	SettingsSvc *settings.Service
}

// List returns paginated audit entries, optionally filtered by entity_type and entity_id.
//
// @Summary      List audit log entries
// @Tags         audit
// @Produce      json
// @Param        entity_type  query  string  false  "Filter by entity type"
// @Param        entity_id    query  int     false  "Filter by entity ID"
// @Param        page         query  int     false  "Page number"  default(1)
// @Param        from_date    query  string  false  "From date YYYY-MM-DD"
// @Param        to_date      query  string  false  "To date YYYY-MM-DD"
// @Success      200  {object}  object{data=[]object,page=int,page_size=int,has_more=bool}
// @Failure      501  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /audit [get]
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
		FromDate:   c.QueryParam("from_date"),
		ToDate:     c.QueryParam("to_date"),
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
		Metadata   *audit.Metadata  `json:"metadata,omitempty"`
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
			Metadata:   e.Metadata,
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

// Cleanup deletes audit entries older than the configured retention period.
//
// @Summary      Clean up old audit entries
// @Tags         audit
// @Produce      json
// @Success      200  {object}  object{deleted=int}
// @Failure      501  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /audit/cleanup [post]
func (h *AuditAPI) Cleanup(c *echo.Context) error {
	if h.Svc == nil || h.SettingsSvc == nil {
		return apiErr(c, http.StatusNotImplemented, "audit not configured")
	}

	days, err := h.SettingsSvc.GetRetentionDays(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if days <= 0 {
		return c.JSON(http.StatusOK, map[string]int64{"deleted": 0})
	}

	n, err := h.Svc.Purge(c.Request().Context(), days)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, map[string]int64{"deleted": n})
}
