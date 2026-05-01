package web

import (
	"database/sql"
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/dates"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/labels"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/web/handlers"
)

//go:embed static
var staticFS embed.FS

// Deps holds application-level dependencies passed into the web layer.
type Deps struct {
	DB             *sql.DB
	AuthService    *auth.Service
	PeopleService  *people.Service
	LabelsService  *labels.Service
	JournalService *journal.Service
	DatesService   *dates.Service
}

// Mount registers all UI routes and the /static/* file server onto e.
// Call this after api.New() returns the Echo instance.
func Mount(e *echo.Echo, deps Deps) {
	// Serve static assets from the embedded FS with a 1-hour cache header.
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("web: failed to sub static FS: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	e.GET("/static/*", func(c *echo.Context) error {
		c.Response().Header().Set("Cache-Control", "public, max-age=3600")
		http.StripPrefix("/static", fileServer).ServeHTTP(c.Response(), c.Request())
		return nil
	})

	// CSRF middleware for all state-changing requests.
	csrfMiddleware := middleware.CSRF()

	// SessionLoader runs on every route — attaches *User to context if cookie valid.
	sessionLoader := auth.SessionLoader(deps.AuthService)

	// Public routes — no RequireAuth; CSRF still applied for POST forms.
	public := e.Group("", sessionLoader, csrfMiddleware)
	{
		authH := &handlers.AuthHandlers{Svc: deps.AuthService}

		public.GET("/login", authH.GetLogin)
		public.POST("/login", authH.PostLogin,
			auth.RateLimitLogin(5, 15*time.Minute),
		)
		public.GET("/health", func(c *echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})
	}

	// Protected routes — RequireAuth redirects to /login when unauthenticated.
	protected := e.Group("", sessionLoader, csrfMiddleware, auth.RequireAuth())
	{
		authH := &handlers.AuthHandlers{Svc: deps.AuthService}
		protected.POST("/logout", authH.PostLogout)
		protected.POST("/logout-all", authH.PostLogoutAll)

		// Page routes
		homeH := &handlers.HomeHandler{
			DB:         deps.DB,
			PeopleSvc:  deps.PeopleService,
			LabelsSvc:  deps.LabelsService,
			JournalSvc: deps.JournalService,
			DatesSvc:   deps.DatesService,
		}
		protected.GET("/", homeH.Get)

		// People routes
		peopleH := &handlers.PeopleHandlers{
			Svc:        deps.PeopleService,
			LabelsSvc:  deps.LabelsService,
			JournalSvc: deps.JournalService,
			DatesSvc:   deps.DatesService,
		}
		protected.GET("/people", peopleH.GetList)
		protected.GET("/people/new", peopleH.GetNew)
		protected.POST("/people", peopleH.PostCreate)
		protected.POST("/people/contact-row", peopleH.PostContactRow)
		protected.POST("/people/location-row", peopleH.PostLocationRow)
		protected.POST("/people/:id/date-row", peopleH.PostDateRow)
		protected.GET("/people/:id", peopleH.GetDetail)
		protected.GET("/people/:id/edit", peopleH.GetEdit)
		protected.POST("/people/:id", peopleH.PostUpdate)
		protected.GET("/people/:id/delete-confirm", peopleH.GetDeleteConfirm)
		protected.POST("/people/:id/delete", peopleH.PostDelete)
		// Label attach/detach routes (htmx fragments)
		protected.POST("/people/:id/labels/attach", peopleH.PostAttachLabel)
		protected.POST("/people/:id/labels/:labelID/delete", peopleH.PostDetachLabel)

		// Labels CRUD routes
		labelsH := &handlers.LabelsHandlers{Svc: deps.LabelsService}
		protected.GET("/labels", labelsH.GetList)
		protected.POST("/labels", labelsH.PostCreate)
		protected.GET("/labels/:id/edit", labelsH.GetEdit)
		protected.POST("/labels/:id", labelsH.PostUpdate)
		protected.POST("/labels/:id/delete", labelsH.PostDelete)

		// Journal routes — /journal/* must come before /journal/:id to avoid
		// routing collisions with the static sub-paths.
		journalH := &handlers.JournalHandlers{Svc: deps.JournalService, PeopleSvc: deps.PeopleService}
		protected.GET("/journal", journalH.GetList)
		protected.GET("/journal/new", journalH.GetNew)
		protected.GET("/journal/people-search", journalH.GetPeopleSearch)
		protected.POST("/journal", journalH.PostCreate)
		protected.GET("/journal/:id", journalH.GetDetail)
		protected.GET("/journal/:id/edit", journalH.GetEdit)
		protected.GET("/journal/:id/delete-confirm", journalH.GetDeleteConfirm)
		protected.POST("/journal/:id", journalH.PostUpdate)
		protected.POST("/journal/:id/delete", journalH.PostDelete)

		// Dates routes
		datesH := &handlers.DatesHandlers{Svc: deps.DatesService}
		protected.GET("/dates", datesH.GetUpcoming)
	}
}
