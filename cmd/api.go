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

			e := api.New()

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

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
