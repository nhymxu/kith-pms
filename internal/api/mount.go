package api

import (
	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/dates"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/labels"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/reminders"
	"github.com/nhymxu/kith-pms/internal/work_history"
)

// Deps holds the dependencies required by the API layer.
type Deps struct {
	PeopleService      *people.Service
	LabelsService      *labels.Service
	JournalService     *journal.Service
	RemindersService   *reminders.Service
	WorkHistoryService *work_history.Service
	DatesService       *dates.Service
	APIToken           string
}

// Mount registers all /v1/* API routes onto e, protected by BearerAuth.
func Mount(e *echo.Echo, deps Deps) {
	v1 := e.Group("/v1", BearerAuth(deps.APIToken))
	mountPeople(v1, deps)
	mountLabels(v1, deps)
	mountJournal(v1, deps)
	mountReminders(v1, deps)
	mountWorkHistory(v1, deps)
	mountDates(v1, deps)
}

func mountPeople(g *echo.Group, deps Deps) {
	h := &PeopleAPI{Svc: deps.PeopleService}
	g.GET("/people", h.List)
	g.POST("/people", h.Create)
	g.GET("/people/:id", h.Get)
	g.PUT("/people/:id", h.Update)
	g.DELETE("/people/:id", h.Delete)
}

func mountLabels(g *echo.Group, deps Deps) {
	h := &LabelsAPI{Svc: deps.LabelsService}
	g.GET("/labels", h.List)
	g.POST("/labels", h.Create)
	g.GET("/labels/:id", h.Get)
	g.PUT("/labels/:id", h.Update)
	g.DELETE("/labels/:id", h.Delete)
}

func mountJournal(g *echo.Group, deps Deps) {
	h := &JournalAPI{Svc: deps.JournalService}
	g.GET("/journal", h.List)
	g.POST("/journal", h.Create)
	g.GET("/journal/:id", h.Get)
	g.PUT("/journal/:id", h.Update)
	g.DELETE("/journal/:id", h.Delete)
}

func mountReminders(g *echo.Group, deps Deps) {
	h := &RemindersAPI{Svc: deps.RemindersService}
	g.GET("/reminders", h.List)
	g.POST("/reminders", h.Create)
	g.GET("/reminders/:id", h.Get)
	g.PUT("/reminders/:id", h.Update)
	g.DELETE("/reminders/:id", h.Delete)
	g.PATCH("/reminders/:id/complete", h.Complete)
}

func mountWorkHistory(g *echo.Group, deps Deps) {
	h := &WorkHistoryAPI{Svc: deps.WorkHistoryService}
	g.GET("/people/:id/work-history", h.ListByPerson)
	g.PUT("/people/:id/work-history", h.ReplaceForPerson)
}

func mountDates(g *echo.Group, deps Deps) {
	h := &DatesAPI{Svc: deps.DatesService}
	g.GET("/people/:id/dates", h.ListByPerson)
	g.PUT("/people/:id/dates", h.ReplaceForPerson)
	g.GET("/dates/upcoming", h.Upcoming)
}
