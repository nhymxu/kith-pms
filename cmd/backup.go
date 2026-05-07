package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
	_ "modernc.org/sqlite"

	"github.com/nhymxu/kith-pms/pkg/config"
)

func backupCommand() *cli.Command {
	return &cli.Command{
		Name:  "backup",
		Usage: "Back up the SQLite database using VACUUM INTO",
		Description: `Creates a clean, compacted copy of the live database at the target path.
Safe to run while the API server is running (uses SQLite WAL / VACUUM INTO).
The backup file contains all data including password hashes — store it securely.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "to",
				Usage:    "Destination path for the backup file (required)",
				Required: true,
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			src := config.ENV.DBPath
			dst := cmd.String("to")

			srcInfo, err := os.Stat(src)
			if err != nil {
				return fmt.Errorf("backup: stat source %q: %w", src, err)
			}

			db, err := sql.Open("sqlite", src)
			if err != nil {
				return fmt.Errorf("backup: open db %q: %w", src, err)
			}
			defer db.Close()

			if _, err := db.ExecContext(context.Background(), "VACUUM INTO ?", dst); err != nil {
				return fmt.Errorf("backup: VACUUM INTO %q: %w", dst, err)
			}

			dstInfo, err := os.Stat(dst)
			if err != nil {
				return fmt.Errorf("backup: stat destination %q: %w", dst, err)
			}

			fmt.Printf("Backed up %s → %s  (%s → %s)\n",
				src, dst,
				humanBytes(srcInfo.Size()),
				humanBytes(dstInfo.Size()),
			)

			return nil
		},
	}
}

func humanBytes(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/float64(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
