package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/pkg/config"
)

func migrateCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrate",
		Usage: "Database migration management",
		Commands: []*cli.Command{
			{
				Name:  "up",
				Usage: "Apply all unapplied migrations",
				Action: func(_ context.Context, _ *cli.Command) error {
					db, err := internaldb.Open(config.C.DBPath, 1)
					if err != nil {
						return fmt.Errorf("migrate up: open db: %w", err)
					}
					defer func() { _ = db.Close() }()

					before, err := internaldb.Status(db)
					if err != nil {
						return fmt.Errorf("migrate up: status before: %w", err)
					}

					pendingBefore := 0

					for _, s := range before {
						if s.AppliedAt == "" {
							pendingBefore++
						}
					}

					if err := internaldb.Up(db); err != nil {
						return fmt.Errorf("migrate up: %w", err)
					}

					fmt.Printf("Applied %d migration(s).\n", pendingBefore)

					return nil
				},
			},
			{
				Name:  "status",
				Usage: "List applied and pending migrations",
				Action: func(_ context.Context, _ *cli.Command) error {
					db, err := internaldb.Open(config.C.DBPath, 1)
					if err != nil {
						return fmt.Errorf("migrate status: open db: %w", err)
					}
					defer func() { _ = db.Close() }()

					statuses, err := internaldb.Status(db)
					if err != nil {
						return fmt.Errorf("migrate status: %w", err)
					}

					if len(statuses) == 0 {
						fmt.Println("No migrations found.")
						return nil
					}

					fmt.Printf("%-8s %-30s %s\n", "VERSION", "NAME", "APPLIED AT")
					fmt.Printf("%-8s %-30s %s\n", "-------", "----", "----------")

					for _, s := range statuses {
						appliedAt := s.AppliedAt
						if appliedAt == "" {
							appliedAt = "(pending)"
						}

						fmt.Printf("%-8d %-30s %s\n", s.Version, s.Name, appliedAt)
					}

					return nil
				},
			},
		},
	}
}
