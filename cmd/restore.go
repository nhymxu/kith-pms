package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/nhymxu/kith-pms/pkg/config"
)

func restoreCommand() *cli.Command {
	return &cli.Command{
		Name:  "restore",
		Usage: "Restore the database from a backup file",
		Description: `Replaces the live database file with a backup copy.
IMPORTANT: Stop the API server before running restore to avoid data corruption.
Use --force to confirm you understand the current database will be overwritten.
The restore file is set to 0600 permissions after copy.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "from",
				Usage:    "Path to the backup file to restore from (required)",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "force",
				Usage: "Required: confirms you want to overwrite the current database",
				Value: false,
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			src := cmd.String("from")
			dst := config.ENV.DBPath
			force := cmd.Bool("force")

			if !force {
				return fmt.Errorf(
					"restore: refusing to overwrite %q without --force flag.\n"+
						"Stop the API server first, then re-run with --force.",
					dst,
				)
			}

			// Heuristic safety check: if the DB file was modified in the last 30s,
			// the API server may still be running.
			if info, err := os.Stat(dst); err == nil {
				if time.Since(info.ModTime()) < 30*time.Second {
					return fmt.Errorf(
						"restore: %q was modified within the last 30s — the API server may be running.\n"+
							"Stop the server and retry.",
						dst,
					)
				}
			}

			// Verify source is readable and stat its size.
			srcInfo, err := os.Stat(src)
			if err != nil {
				return fmt.Errorf("restore: stat source %q: %w", src, err)
			}

			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("restore: copy %q → %q: %w", src, dst, err)
			}

			// Lock down permissions — DB contains password hash + sessions.
			if err := os.Chmod(dst, 0o600); err != nil {
				return fmt.Errorf("restore: chmod %q: %w", dst, err)
			}

			fmt.Printf("Restored %s → %s  (%s)\n", src, dst, humanBytes(srcInfo.Size()))
			return nil
		},
	}
}

// copyFile copies src to dst, creating or truncating dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
