package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/people"
)

// MeAPI handles /v1/me endpoints (self-profile mapping).
type MeAPI struct {
	PeopleSvc *people.Service
}

// GetMe handles GET /v1/me.
//
// @Summary      Get self profile
// @Description  Returns the person designated as "Me", or 404 if not configured.
// @Tags         me
// @Produce      json
// @Success      200  {object}  envelope
// @Failure      404  {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /me [get]
func (h *MeAPI) GetMe(c *echo.Context) error {
	self, err := h.PeopleSvc.GetSelf(c.Request().Context())
	if err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	if self == nil {
		return apiErr(c, http.StatusNotFound, "self person not configured")
	}

	return ok(c, self)
}

// PostSetup handles POST /v1/me/setup.
//
// @Summary      Setup self profile
// @Description  Designates an existing person as the "Me" person.
// @Tags         me
// @Accept       json
// @Produce      json
// @Param        body  body      object{person_id=int}  true  "Person to designate as self"
// @Success      200   {object}  envelope{data=object{person_id=int}}
// @Failure      400   {object}  envelope
// @Security     CookieAuth
// @Security     CSRFHeader
// @Router       /me/setup [post]
func (h *MeAPI) PostSetup(c *echo.Context) error {
	var req struct {
		PersonID int64 `json:"person_id"`
	}

	// Accept both JSON body and form value for flexibility.
	if err := c.Bind(&req); err != nil || req.PersonID <= 0 {
		// Fallback: try raw form value.
		if id, parseErr := strconv.ParseInt(c.FormValue("person_id"), 10, 64); parseErr == nil && id > 0 {
			req.PersonID = id
		} else {
			return apiErr(c, http.StatusBadRequest, "person_id is required")
		}
	}

	if err := h.PeopleSvc.SetSelf(c.Request().Context(), req.PersonID); err != nil {
		return apiErr(c, http.StatusInternalServerError, "internal server error")
	}

	return ok(c, map[string]any{"person_id": req.PersonID})
}
