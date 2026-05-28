package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/work_history"
)

// WorkHistoryAPI handles person-scoped work history endpoints.
type WorkHistoryAPI struct {
	Svc *work_history.Service
}

// workHistoryReplaceRequest is the JSON body for full replace.
type workHistoryReplaceRequest struct {
	Entries []workEntryRequest `json:"entries"`
}

type workEntryRequest struct {
	Company     string `json:"company"`
	Title       string `json:"title"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Location    string `json:"location"`
	Description string `json:"description"`
	Position    int    `json:"position"`
}

// ListByPerson handles GET /v1/people/:id/work-history
func (h *WorkHistoryAPI) ListByPerson(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	entries, err := h.Svc.ListByPerson(c.Request().Context(), personID)
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, entries)
}

// ReplaceForPerson handles PUT /v1/people/:id/work-history
func (h *WorkHistoryAPI) ReplaceForPerson(c *echo.Context) error {
	personID, err := parseID(c)
	if err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid id")
	}

	var req workHistoryReplaceRequest
	if err := c.Bind(&req); err != nil {
		return apiErr(c, http.StatusBadRequest, "invalid request body")
	}

	entries := make([]work_history.WorkEntry, 0, len(req.Entries))
	for _, e := range req.Entries {
		if e.Company == "" {
			return apiErr(c, http.StatusUnprocessableEntity, "company is required for each entry")
		}

		startDate, err := work_history.ParseWorkDate(e.StartDate)
		if err != nil {
			return apiErr(c, http.StatusUnprocessableEntity, "invalid start_date: "+err.Error())
		}

		// EndDate is optional (empty = Present); validate only if non-empty.
		endDate := ""
		if e.EndDate != "" {
			endDate, err = work_history.ParseWorkDate(e.EndDate)
			if err != nil {
				return apiErr(c, http.StatusUnprocessableEntity, "invalid end_date: "+err.Error())
			}
		}

		entries = append(entries, work_history.WorkEntry{
			PersonID:    personID,
			Company:     e.Company,
			Title:       e.Title,
			StartDate:   startDate,
			EndDate:     endDate,
			Location:    e.Location,
			Description: e.Description,
			Position:    e.Position,
		})
	}

	if err := h.Svc.ReplaceForPerson(c.Request().Context(), personID, entries); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, nil)
}
