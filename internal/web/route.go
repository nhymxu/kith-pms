package web

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/nhymxu/kith-pms/internal/api"
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
	"github.com/nhymxu/kith-pms/internal/web/spa"
	"github.com/nhymxu/kith-pms/internal/work_history"
)

// Deps holds application-level dependencies passed into the web layer.
type Deps struct {
	DB                   *sql.DB
	AuthService          *auth.Service
	PeopleService        *people.Service
	LabelsService        *labels.Service
	JournalService       *journal.Service
	DatesService         *dates.Service
	RemindersService     *reminders.Service
	WorkHistoryService   *work_history.Service
	AuditService         *audit.Service
	GiftsService         *gifts.Service
	RelationshipsService *relationships.Service
	SettingsService      *settings.Service
	FileSvc              files.FileService
	AvatarBasePath       string
	GiftStoragePath      string
	APIToken             string
	SessionLifetime      time.Duration
	BehindTLS            bool
}

// Mount registers all routes onto e.
// Order matters: /v1/* and /health are mounted first; spa.Handler() is last (catch-all).
func Mount(e *echo.Echo, deps Deps) {
	// Health endpoint — no auth, no session overhead.
	e.GET("/health", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// SessionLoader attaches *User to context when a valid session cookie is present.
	sessionLoader := auth.SessionLoader(deps.AuthService)

	// Shared rate limiter: 5 attempts per 15 minutes per IP.
	loginLimiter := auth.RateLimitLogin(5, 15*time.Minute)

	// Mount JSON REST API routes under /v1/.
	apiDeps := api.Deps{
		PeopleService:        deps.PeopleService,
		LabelsService:        deps.LabelsService,
		JournalService:       deps.JournalService,
		RemindersService:     deps.RemindersService,
		WorkHistoryService:   deps.WorkHistoryService,
		DatesService:         deps.DatesService,
		AuditService:         deps.AuditService,
		GiftsService:         deps.GiftsService,
		RelationshipsService: deps.RelationshipsService,
		SettingsService:      deps.SettingsService,
		AuthService:          deps.AuthService,
		FileSvc:              deps.FileSvc,
		APIToken:             deps.APIToken,
		AvatarBasePath:       deps.AvatarBasePath,
		GiftStoragePath:      deps.GiftStoragePath,
		SessionLifetime:      deps.SessionLifetime,
		BehindTLS:            deps.BehindTLS,
		LoginLimiter:         loginLimiter,
	}

	// Login must bypass SessionOrBearer — registered before Mount applies auth middleware.
	api.MountAuthLogin(e, apiDeps)
	api.Mount(e, apiDeps)

	// Inject audit actor from session context for any remaining middleware-wrapped routes.
	e.Use(sessionLoader, injectAuditActor(deps))

	// SPA catch-all: serves index.html for all non-API GET paths.
	// Must be mounted LAST.
	spa.Handler(e)
}

// injectAuditActor copies the authenticated user ID from the Echo context into
// the request context so service-layer audit calls have actor attribution.
func injectAuditActor(_ Deps) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if u := auth.UserFromContext(c); u != nil {
				ctx := audit.WithActor(c.Request().Context(), u.ID)
				c.SetRequest(c.Request().WithContext(ctx))
			}

			return next(c)
		}
	}
}
