package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/lmittmann/tint"
	slogmulti "github.com/samber/slog-multi"
	slogsentry "github.com/samber/slog-sentry/v2"
	"github.com/urfave/cli/v3"

	"github.com/nhymxu/kith-pms/pkg/config"
)

func newApp() *cli.Command {
	return &cli.Command{
		Name:  "kith-pms",
		Usage: "Kith - Personal Management System",
		Description: `kith (kith and kin) is a dead simple Personal Management System
with relationship, life log, memory, ...`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "",
				Usage: "config file (default is $APPLICATION_DIR/.env)",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			dependencyInit(cmd.String("config"))
			return ctx, nil
		},
		Commands: []*cli.Command{
			webServerCommand(),
			migrateCommand(),
			setPasswordCommand(),
			backupCommand(),
			restoreCommand(),
			monicaImportCommand(),
		},
	}
}

func dependencyInit(cfgFile string) {
	err := config.Load(cfgFile)
	if err != nil {
		panic("Can't load config from environment")
	}

	initLog()
	initSentry()
}

func newBaseHandler() slog.Handler {
	if config.C.Debug {
		return tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: "15:04:05.000",
		})
	}

	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
}

func initLog() {
	slog.SetDefault(slog.New(newBaseHandler()))
}

func initSentry() {
	if config.C.Sentry.DSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              config.C.Sentry.DSN,
			AttachStacktrace: true,
		})
		if err != nil {
			slog.Error("Sentry initialization failed", "error", err)
		} else {
			slog.Info("Initialized Sentry integration.")
			integrateSlogWithSentry()
		}
	} else {
		slog.Info("SENTRY_DSN not found, sentry integration disabled.")
	}
}

func integrateSlogWithSentry() {
	sentryHandler := slogsentry.Option{
		Level: slog.LevelError,
	}.NewSentryHandler()

	handler := slogmulti.Fanout(newBaseHandler(), sentryHandler)
	slog.SetDefault(slog.New(handler))
	slog.Info("Integrate slog with Sentry successfully.")
}
