package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/api/handler"
	"github.com/nhymxu/kith-pms/internal/api/spa"
	"github.com/nhymxu/kith-pms/internal/audit"
	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/files"
	"github.com/nhymxu/kith-pms/internal/gifts"
	"github.com/nhymxu/kith-pms/internal/important_dates"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/metrics"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/relationships"
	"github.com/nhymxu/kith-pms/internal/reminders"
	"github.com/nhymxu/kith-pms/internal/settings"
	"github.com/nhymxu/kith-pms/internal/work_history"
)

// Deps holds the dependencies required by the API layer.
type Deps struct {
	DB                   *bun.DB
	PeopleService        *people.Service
	LabelsService        *people.LabelService
	JournalService       *journal.Service
	JournalLabelsService *journal.LabelService
	RemindersService     *reminders.Service
	WorkHistoryService   *work_history.Service
	DatesService         *important_dates.Service
	AuditService         *audit.Service
	GiftsService         *gifts.Service
	RelationshipsService *relationships.Service
	SettingsService      *settings.Service
	AuthService          *auth.Service
	FileSvc              files.FileService
	APIToken             string
	AvatarBasePath       string
	GiftStoragePath      string
	SessionLifetime      time.Duration
	BehindTLS            bool
}

// Mount registers all routes onto e.
// Order matters: /health, /metrics, /v1/* are mounted first; spa.Handler() is last (catch-all).
func Mount(e *echo.Echo, deps Deps) {
	e.GET("/health", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.GET("/metrics", metrics.Handler())

	sessionLoader := auth.SessionLoader(deps.AuthService)

	loginLimiter := auth.RateLimitLogin(5, 15*time.Minute)

	MountAuthLogin(e, deps, loginLimiter)

	v1 := e.Group("/v1",
		auth.SessionOrBearer(deps.APIToken, deps.AuthService),
		auth.SpaCSRF(),
		injectAPIActor(),
	)

	// Cross-wire birthday services before mounting routes.
	if deps.PeopleService != nil && deps.RemindersService != nil {
		deps.RemindersService.PersonDOB = handler.NewPersonDOBLookup(deps.PeopleService)
		deps.PeopleService.BirthdaySync = handler.NewBirthdayReminderSyncer(deps.RemindersService)
	}

	mountAuth(v1, deps)
	mountMe(v1, deps)
	mountPeople(v1, deps)
	mountPeopleLabelsCRUD(v1, deps)
	mountJournal(v1, deps)
	mountJournalLabels(v1, deps)
	mountReminders(v1, deps)
	mountWorkHistory(v1, deps)
	mountDates(v1, deps)
	mountAudit(v1, deps)
	mountGifts(v1, deps)
	mountRelationships(v1, deps)
	mountPeopleLabels(v1, deps)
	mountPeopleAvatars(v1, deps)
	mountPeopleQuick(v1, deps)
	mountSettings(v1, deps)
	mountAppInfo(v1)

	e.Use(sessionLoader, injectAuditActor())

	mountSwagger(e)
	spa.Handler(e)
}

func mountPeople(g *echo.Group, deps Deps) {
	h := &handler.PeopleAPI{Svc: deps.PeopleService, ReminderSvc: deps.RemindersService}
	g.GET("/people", h.List)
	g.POST("/people", h.Create)
	g.GET("/people/:id", h.Get)
	g.PUT("/people/:id", h.Update)
	g.DELETE("/people/:id", h.Delete)
}

func mountPeopleLabelsCRUD(g *echo.Group, deps Deps) {
	h := &handler.PeopleLabelsCRUD{Svc: deps.LabelsService}
	g.GET("/people-labels", h.List)
	g.POST("/people-labels", h.Create)
	g.GET("/people-labels/:id", h.Get)
	g.PUT("/people-labels/:id", h.Update)
	g.DELETE("/people-labels/:id", h.Delete)
}

func mountJournal(g *echo.Group, deps Deps) {
	h := &handler.JournalAPI{Svc: deps.JournalService}
	g.GET("/journal", h.List)
	g.POST("/journal", h.Create)
	g.GET("/journal/:id", h.Get)
	g.PUT("/journal/:id", h.Update)
	g.DELETE("/journal/:id", h.Delete)
}

func mountJournalLabels(g *echo.Group, deps Deps) {
	h := &handler.JournalLabelsAPI{Svc: deps.JournalLabelsService}
	g.GET("/journal-labels", h.List)
	g.POST("/journal-labels", h.Create)
	g.GET("/journal-labels/:id", h.Get)
	g.PUT("/journal-labels/:id", h.Update)
	g.DELETE("/journal-labels/:id", h.Delete)
}

func mountReminders(g *echo.Group, deps Deps) {
	svc := deps.RemindersService
	if deps.PeopleService != nil {
		svc.Journal = handler.NewPeopleLastContacter(deps.PeopleService)
	}

	h := &handler.RemindersAPI{Svc: svc}
	g.GET("/reminders", h.List)
	g.POST("/reminders", h.Create)
	g.GET("/reminders/:id", h.Get)
	g.PUT("/reminders/:id", h.Update)
	g.DELETE("/reminders/:id", h.Delete)
	g.PATCH("/reminders/:id/complete", h.Complete)
}

func mountWorkHistory(g *echo.Group, deps Deps) {
	h := &handler.WorkHistoryAPI{Svc: deps.WorkHistoryService}
	g.GET("/people/:id/work-history", h.ListByPerson)
	g.PUT("/people/:id/work-history", h.ReplaceForPerson)
}

func mountDates(g *echo.Group, deps Deps) {
	h := &handler.DatesAPI{Svc: deps.DatesService}
	g.GET("/people/:id/dates", h.ListByPerson)
	g.PUT("/people/:id/dates", h.ReplaceForPerson)
	g.GET("/dates/upcoming", h.Upcoming)
}

func mountAudit(g *echo.Group, deps Deps) {
	h := &handler.AuditAPI{Svc: deps.AuditService, SettingsSvc: deps.SettingsService}
	g.GET("/audit", h.List)
	g.POST("/audit/cleanup", h.Cleanup)
}

func mountGifts(g *echo.Group, deps Deps) {
	h := &handler.GiftsAPI{Svc: deps.GiftsService, GiftStoragePath: deps.GiftStoragePath}
	g.GET("/gifts", h.List)
	g.POST("/gifts", h.Create)
	g.GET("/gifts/:id", h.GetByID)
	g.PUT("/gifts/:id", h.Update)
	g.DELETE("/gifts/:id", h.Delete)
	g.GET("/people/:id/gifts", h.ListByPerson)
	g.POST("/gifts/:id/image", h.UploadImage)
	g.DELETE("/gifts/:id/image", h.DeleteImage)
	g.GET("/gifts/:id/image", h.GetImage)
}

func mountAuth(g *echo.Group, deps Deps) {
	// Login is public — mount on a separate group without the parent auth middleware.
	// We reach here via the parent /v1 group which already has SessionOrBearer, so we
	// register a subgroup. Login itself must be accessible pre-auth, so mount it on
	// the echo instance directly via a sibling group that bypasses SessionOrBearer.
	//
	// NOTE: The parent /v1 group enforces auth. To allow unauthenticated login, we
	// attach the /v1/auth/login endpoint to a separate unprotected group in the echo
	// instance. The other auth endpoints (logout, me, password) remain inside the
	// protected group.
	h := &handler.AuthAPI{
		Svc:             deps.AuthService,
		SessionLifetime: deps.SessionLifetime,
		BehindTLS:       deps.BehindTLS,
	}

	// Protected auth endpoints (require existing auth or cookie).
	g.POST("/auth/logout", h.Logout)
	g.POST("/auth/logout-all", h.LogoutAll)
	g.GET("/auth/me", h.Me)
	g.POST("/auth/password", h.ChangePassword)
}

// MountAuthLogin registers the unauthenticated login route on the root Echo instance.
// Called separately because the /v1 group requires auth — login must bypass it.
// The limiter argument enforces a shared 5/15min bucket per IP across all login endpoints.
func MountAuthLogin(e *echo.Echo, deps Deps, limiter echo.MiddlewareFunc) {
	h := &handler.AuthAPI{
		Svc:             deps.AuthService,
		SessionLifetime: deps.SessionLifetime,
		BehindTLS:       deps.BehindTLS,
	}

	loginGroup := e.Group("/v1/auth", limiter)
	loginGroup.POST("/login", h.Login)
}

func mountMe(g *echo.Group, deps Deps) {
	h := &handler.MeAPI{PeopleSvc: deps.PeopleService}
	g.GET("/me", h.GetMe)
	g.POST("/me/setup", h.PostSetup)
}

func mountRelationships(g *echo.Group, deps Deps) {
	h := &handler.RelationshipsAPI{Svc: deps.RelationshipsService}
	g.GET("/relationship-types", h.ListTypes)
	g.POST("/relationship-types", h.CreateType)
	g.PUT("/relationship-types/:id", h.UpdateType)
	g.DELETE("/relationship-types/:id", h.DeleteType)
	g.GET("/people/:id/relationships", h.ListByPerson)
	g.POST("/people/:id/relationships", h.AttachRelationship)
	g.DELETE("/people/:id/relationships/:relID", h.DetachRelationship)
}

func mountPeopleLabels(g *echo.Group, deps Deps) {
	h := &handler.PeopleLabelsAPI{Svc: deps.LabelsService}
	g.POST("/people/:id/labels", h.Attach)
	g.DELETE("/people/:id/labels/:labelID", h.Detach)
}

func mountPeopleAvatars(g *echo.Group, deps Deps) {
	h := &handler.AvatarsAPI{
		PeopleSvc:      deps.PeopleService,
		FileSvc:        deps.FileSvc,
		AvatarBasePath: deps.AvatarBasePath,
	}
	g.POST("/people/:id/avatar", h.Upload)
	g.DELETE("/people/:id/avatar", h.Delete)
	g.GET("/people/:id/avatar", h.Get)
}

func mountSettings(g *echo.Group, deps Deps) {
	h := &handler.SettingsAPI{Svc: deps.SettingsService}
	g.GET("/settings", h.Get)
	g.PUT("/settings", h.Update)
}

func mountAppInfo(g *echo.Group) {
	g.GET("/app/info", handler.GetAppInfo)
}

func mountPeopleQuick(g *echo.Group, deps Deps) {
	h := &handler.PeopleQuickAPI{
		PeopleSvc:  deps.PeopleService,
		JournalSvc: deps.JournalService,
		GiftsSvc:   deps.GiftsService,
	}
	g.POST("/people/:id/journal/quick", h.QuickJournal)
	g.POST("/people/:id/gifts/quick", h.QuickGift)
	g.POST("/people/:id/last-contact", h.UpdateLastContact)
}
