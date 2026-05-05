package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/web/templates"
)

type MeHandlers struct {
	PeopleSvc *people.Service
}

func (h *MeHandlers) GetMe(c *echo.Context) error {
	self, err := h.PeopleSvc.GetSelf(c.Request().Context())
	if err != nil {
		return err
	}
	if self == nil {
		return c.Redirect(http.StatusFound, "/me/setup")
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("/people/%d", self.ID))
}

func (h *MeHandlers) GetSetup(c *echo.Context) error {
	all, err := h.PeopleSvc.List(c.Request().Context(), people.ListParams{PageSize: 200})
	if err != nil {
		return err
	}
	self, err := h.PeopleSvc.GetSelf(c.Request().Context())
	if err != nil {
		return err
	}
	component := templates.MeSetup(templates.MeSetupParams{
		People:     all,
		SelfPerson: self,
		CSRFToken:  auth.CSRFToken(c),
	})
	return component.Render(c.Request().Context(), c.Response())
}

func (h *MeHandlers) PostSetup(c *echo.Context) error {
	personID, err := strconv.ParseInt(c.FormValue("person_id"), 10, 64)
	if err != nil || personID <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid person_id")
	}
	if err := h.PeopleSvc.SetSelf(c.Request().Context(), personID); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/people/%d", personID))
}
