package handlers

import (
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/audit"
	"github.com/nhymxu/kith-pms/internal/web/templates"
)

// AuditHandlers groups the /audit/* HTTP handlers.
type AuditHandlers struct {
	Svc *audit.Service
}

// GetList handles GET /audit
func (h *AuditHandlers) GetList(c *echo.Context) error {
	entityType := audit.EntityType(c.QueryParam("entity_type"))

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	entries, err := h.Svc.List(c.Request().Context(), audit.ListParams{
		EntityType: entityType,
		Page:       page,
		PageSize:   50,
	})
	if err != nil {
		return err
	}

	component := templates.AuditList(templates.AuditListParams{
		Entries:    entries,
		EntityType: entityType,
		Page:       page,
		HasMore:    len(entries) == 50,
	})

	return component.Render(c.Request().Context(), c.Response())
}
