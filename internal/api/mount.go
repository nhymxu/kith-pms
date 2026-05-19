package api

import (
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/audit"
	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/dates"
	"github.com/nhymxu/kith-pms/internal/files"
	"github.com/nhymxu/kith-pms/internal/gifts"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/labels"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/relationships"
	"github.com/nhymxu/kith-pms/internal/reminders"
	"github.com/nhymxu/kith-pms/internal/settings"
	"github.com/nhymxu/kith-pms/internal/work_history"
)

// Deps holds the dependencies required by the API layer.
type Deps struct {
	PeopleService        *people.Service
	LabelsService        *labels.Service
	JournalService       *journal.Service
	RemindersService     *reminders.Service
	WorkHistoryService   *work_history.Service
	DatesService         *dates.Service
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
	// LoginLimiter is the rate-limiter middleware shared with the templ login route.
	// When non-nil, the same bucket is used for both /login and /v1/auth/login (5/15min total per IP).
	// When nil, MountAuthLogin constructs its own limiter (not shared).
	LoginLimiter echo.MiddlewareFunc
}

// Mount registers all /v1/* API routes onto e, protected by SessionOrBearer + SpaCSRF.
func Mount(e *echo.Echo, deps Deps) {
	v1 := e.Group("/v1",
		auth.SessionOrBearer(deps.APIToken, deps.AuthService),
		auth.SpaCSRF(),
		injectAPIActor(),
	)

	mountAuth(v1, deps)
	mountMe(v1, deps)
	mountPeople(v1, deps)
	mountLabels(v1, deps)
	mountJournal(v1, deps)
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

func mountAudit(g *echo.Group, deps Deps) {
	h := &AuditAPI{Svc: deps.AuditService}
	g.GET("/audit", h.List)
}

func mountGifts(g *echo.Group, deps Deps) {
	h := &GiftsAPI{Svc: deps.GiftsService, GiftStoragePath: deps.GiftStoragePath}
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
	h := newAuthAPI(deps)

	// Protected auth endpoints (require existing auth or cookie).
	g.POST("/auth/logout", h.Logout)
	g.POST("/auth/logout-all", h.LogoutAll)
	g.GET("/auth/me", h.Me)
	g.POST("/auth/password", h.ChangePassword)
}

// MountAuthLogin registers the unauthenticated login route on the root Echo instance.
// Called separately because the /v1 group requires auth — login must bypass it.
// Uses deps.LoginLimiter when set (shared with templ /login) to enforce a single
// 5/15min bucket per IP across both login endpoints.
func MountAuthLogin(e *echo.Echo, deps Deps) {
	h := newAuthAPI(deps)

	limiter := deps.LoginLimiter
	if limiter == nil {
		limiter = auth.RateLimitLogin(5, 15*time.Minute)
	}

	loginGroup := e.Group("/v1/auth", limiter)
	loginGroup.POST("/login", h.Login)
}

func mountMe(g *echo.Group, deps Deps) {
	h := &MeAPI{PeopleSvc: deps.PeopleService}
	g.GET("/me", h.GetMe)
	g.POST("/me/setup", h.PostSetup)
}

func mountRelationships(g *echo.Group, deps Deps) {
	h := &RelationshipsAPI{Svc: deps.RelationshipsService}
	g.GET("/relationship-types", h.ListTypes)
	g.POST("/relationship-types", h.CreateType)
	g.PUT("/relationship-types/:id", h.UpdateType)
	g.DELETE("/relationship-types/:id", h.DeleteType)
	g.GET("/people/:id/relationships", h.ListByPerson)
	g.POST("/people/:id/relationships", h.AttachRelationship)
	g.DELETE("/people/:id/relationships/:relID", h.DetachRelationship)
}

func mountPeopleLabels(g *echo.Group, deps Deps) {
	h := &PeopleLabelsAPI{Svc: deps.LabelsService}
	g.POST("/people/:id/labels", h.Attach)
	g.DELETE("/people/:id/labels/:labelID", h.Detach)
}

func mountPeopleAvatars(g *echo.Group, deps Deps) {
	h := &AvatarsAPI{
		PeopleSvc:      deps.PeopleService,
		FileSvc:        deps.FileSvc,
		AvatarBasePath: deps.AvatarBasePath,
	}
	g.POST("/people/:id/avatar", h.Upload)
	g.DELETE("/people/:id/avatar", h.Delete)
	g.GET("/people/:id/avatar", h.Get)
}

func mountSettings(g *echo.Group, deps Deps) {
	h := &SettingsAPI{Svc: deps.SettingsService}
	g.GET("/settings", h.Get)
	g.PUT("/settings", h.Update)
}

func mountPeopleQuick(g *echo.Group, deps Deps) {
	h := &PeopleQuickAPI{
		PeopleSvc:  deps.PeopleService,
		JournalSvc: deps.JournalService,
		GiftsSvc:   deps.GiftsService,
	}
	g.POST("/people/:id/journal/quick", h.QuickJournal)
	g.POST("/people/:id/gifts/quick", h.QuickGift)
	g.POST("/people/:id/last-contact", h.UpdateLastContact)
}
