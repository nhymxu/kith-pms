package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/urfave/cli/v3"

	"github.com/nhymxu/kith-pms/internal/api"
	"github.com/nhymxu/kith-pms/internal/audit"
	"github.com/nhymxu/kith-pms/internal/auth"
	"github.com/nhymxu/kith-pms/internal/dates"
	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/internal/files"
	"github.com/nhymxu/kith-pms/internal/gifts"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/metrics"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/relationships"
	"github.com/nhymxu/kith-pms/internal/reminders"
	"github.com/nhymxu/kith-pms/internal/settings"
	"github.com/nhymxu/kith-pms/internal/work_history"
	"github.com/nhymxu/kith-pms/pkg/config"
	"github.com/nhymxu/kith-pms/pkg/pathutil"
)

func webServerCommand() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Web server",
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

			secret := []byte(config.C.SessionSecret)
			if len(secret) < 32 {
				slog.Error("SESSION_SECRET must be at least 32 bytes — refusing to start")
				os.Exit(1)
			}

			dbPath := config.C.DBPath
			if err := os.MkdirAll(pathutil.DirOf(dbPath), 0o700); err != nil {
				return fmt.Errorf("api: create db dir: %w", err)
			}

			db, err := internaldb.Open(dbPath)
			if err != nil {
				return fmt.Errorf("api: open db: %w", err)
			}
			defer func() { _ = db.Close() }()

			if config.C.Debug {
				db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
			}

			if config.C.DBAutoMigrate {
				if err := internaldb.Up(db); err != nil {
					return fmt.Errorf("api: auto-migrate: %w", err)
				}
			}

			lifetime := config.C.SessionLifetime
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

			metrics.RegisterAppCollectors(db, authSvc.Sessions)
			metrics.RegisterBuildInfo()

			peopleSvc := people.NewService(db)

			avatarPath := config.C.AvatarStoragePath
			if avatarPath == "" {
				avatarPath = "data/avatars"
			}

			if err := os.MkdirAll(avatarPath, 0o700); err != nil {
				return fmt.Errorf("web-server: create avatar dir: %w", err)
			}

			fileSvc := files.NewLocalFileService(avatarPath)
			peopleSvc.FileService = fileSvc

			labelsSvc := people.NewLabelService(db)
			peopleSvc.LabelsSvc = labelsSvc

			journalSvc := journal.NewService(db)
			journalSvc.PeopleSvc = &api.JournalPeopleAdapter{Svc: peopleSvc}
			journalLabelsSvc := journal.NewLabelService(db)

			datesSvc := dates.NewService(db)

			remindersSvc := reminders.NewService(db)

			workHistorySvc := work_history.NewService(db)

			auditSvc := audit.NewService(db)
			peopleSvc.Audit = auditSvc
			labelsSvc.Audit = auditSvc
			journalSvc.Audit = auditSvc
			remindersSvc.Audit = auditSvc
			workHistorySvc.Audit = auditSvc
			datesSvc.Audit = auditSvc

			// Wire gifts service.
			giftsSvc := gifts.NewService(db)
			giftsSvc.Audit = auditSvc

			giftStoragePath := config.C.GiftStoragePath
			if giftStoragePath == "" {
				giftStoragePath = "data/gifts"
			}

			if err := os.MkdirAll(giftStoragePath, 0o700); err != nil {
				return fmt.Errorf("web-server: create gift dir: %w", err)
			}

			giftFileSvc := files.NewLocalFileService(giftStoragePath)
			giftsSvc.FileSvc = giftFileSvc

			// Wire relationships service.
			relsSvc := relationships.NewService(db)
			relsSvc.Audit = auditSvc

			settingsSvc := settings.NewService(db)

			apiToken := os.Getenv("API_TOKEN")

			api.Mount(e, api.Deps{
				DB:                   db,
				AuthService:          authSvc,
				PeopleService:        peopleSvc,
				LabelsService:        labelsSvc,
				JournalService:       journalSvc,
				JournalLabelsService: journalLabelsSvc,
				DatesService:         datesSvc,
				RemindersService:     remindersSvc,
				WorkHistoryService:   workHistorySvc,
				AuditService:         auditSvc,
				GiftsService:         giftsSvc,
				RelationshipsService: relsSvc,
				SettingsService:      settingsSvc,
				FileSvc:              fileSvc,
				AvatarBasePath:       avatarPath,
				GiftStoragePath:      giftStoragePath,
				APIToken:             apiToken,
				SessionLifetime:      lifetime,
				BehindTLS:            config.C.BehindTLS,
			})

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			go api.RunSessionGC(ctx, authSvc.Sessions)

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
