package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/urfave/cli/v3"

	"github.com/nhymxu/kith-pms/app/api"
	"github.com/nhymxu/kith-pms/internal/auth"
	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/labels"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/web"
	"github.com/nhymxu/kith-pms/internal/web/handlers"
	"github.com/nhymxu/kith-pms/pkg/config"
)

func apiCommand() *cli.Command {
	return &cli.Command{
		Name:  "api",
		Usage: "API server",
		Description: `Serve all service on same pod.
Can scale later.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Value: "",
				Usage: "API host listening",
			},
			&cli.Int64Flag{
				Name:  "port",
				Value: 8000,
				Usage: "API port listening",
			},
			&cli.Int64Flag{
				Name:  "shutdown_time",
				Value: config.APIDefaultShutdownTime,
				Usage: "Gracefully shutdown time",
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			host := cmd.String("host")
			port := cmd.Int64("port")
			shutdownTime := cmd.Int64("shutdown_time")

			// Validate SESSION_SECRET before opening anything.
			secret := []byte(config.ENV.SessionSecret)
			if len(secret) < 32 {
				slog.Error("SESSION_SECRET must be at least 32 bytes — refusing to start")
				os.Exit(1)
			}

			// Open SQLite database.
			dbPath := config.ENV.DBPath
			if err := os.MkdirAll(dirOf(dbPath), 0o700); err != nil {
				return fmt.Errorf("api: create db dir: %w", err)
			}
			db, err := internaldb.Open(dbPath)
			if err != nil {
				return fmt.Errorf("api: open db: %w", err)
			}
			defer db.Close()

			// Run migrations automatically when configured (default: true).
			if config.ENV.DBAutoMigrate {
				if err := internaldb.Up(db); err != nil {
					return fmt.Errorf("api: auto-migrate: %w", err)
				}
			}

			// Wire auth service.
			lifetime := config.ENV.SessionLifetime
			if lifetime <= 0 {
				lifetime = 30 * 24 * time.Hour
			}
			authSvc := &auth.Service{
				Users:    auth.NewUserRepo(db),
				Sessions: auth.NewSessionRepo(db),
				Secret:   secret,
				Lifetime: lifetime,
			}

			e := api.New()

			// Set custom error handler for styled HTML error pages.
			e.HTTPErrorHandler = handlers.CustomHTTPErrorHandler

			// Wire people service.
			peopleSvc := people.NewService(db)

			// Wire labels service.
			labelsSvc := labels.NewService(db)

			// Wire journal service.
			journalSvc := journal.NewService(db)

			// Mount HTML UI routes on the same Echo instance.
			web.Mount(e, web.Deps{
				DB:             db,
				AuthService:    authSvc,
				PeopleService:  peopleSvc,
				LabelsService:  labelsSvc,
				JournalService: journalSvc,
			})

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			// Background session GC — runs every hour, cancels on shutdown.
			go runSessionGC(ctx, authSvc.Sessions)

			sc := echo.StartConfig{
				Address:         fmt.Sprintf("%s:%d", host, port),
				HideBanner:      true,
				GracefulTimeout: time.Duration(shutdownTime) * time.Second,
				OnShutdownError: func(err error) {
					slog.Error("graceful shutdown error", "error", err)
				},
			}

			if err := sc.Start(ctx, e); err != nil {
				slog.Error("shutting down the server", "error", err)
			}

			return nil
		},
	}
}

// runSessionGC deletes expired sessions every hour until ctx is cancelled.
func runSessionGC(ctx context.Context, repo auth.SessionRepo) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := repo.DeleteExpiredSessions(ctx); err != nil {
				slog.Warn("session GC error", "error", err)
			} else {
				slog.Debug("session GC: expired sessions deleted")
			}
		}
	}
}

// dirOf returns the directory component of a file path.
// e.g. "data/kith.db" → "data"
func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}
