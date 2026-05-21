package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/urfave/cli/v3"
	"golang.org/x/term"

	"github.com/nhymxu/kith-pms/internal/auth"
	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/pkg/config"
	"github.com/nhymxu/kith-pms/pkg/pathutil"
)

func setPasswordCommand() *cli.Command {
	return &cli.Command{
		Name:  "set-password",
		Usage: "Set or reset the application login password",
		Description: `Prompts for a new password (twice for confirmation), hashes it with
argon2id, and stores the hash in the database. Creates the user row if
it does not exist yet.`,
		Action: func(_ context.Context, _ *cli.Command) error {
			pwd1, err := promptPassword("New password: ")
			if err != nil {
				return fmt.Errorf("set-password: prompt: %w", err)
			}

			pwd2, err := promptPassword("Confirm password: ")
			if err != nil {
				return fmt.Errorf("set-password: confirm: %w", err)
			}

			if pwd1 != pwd2 {
				return fmt.Errorf("set-password: passwords do not match")
			}

			if len(pwd1) == 0 {
				return fmt.Errorf("set-password: password must not be empty")
			}

			hash, err := auth.HashPassword(pwd1)
			if err != nil {
				return fmt.Errorf("set-password: hash: %w", err)
			}

			dbPath := config.ENV.DBPath
			if err := os.MkdirAll(pathutil.DirOf(dbPath), 0o700); err != nil {
				return fmt.Errorf("set-password: create db dir: %w", err)
			}

			db, err := internaldb.Open(dbPath)
			if err != nil {
				return fmt.Errorf("set-password: open db: %w", err)
			}
			defer func() { _ = db.Close() }()

			if err := internaldb.Up(db); err != nil {
				return fmt.Errorf("set-password: migrate: %w", err)
			}

			repo := auth.NewUserRepo(db)
			if err := repo.UpsertUser(context.Background(), hash); err != nil {
				return fmt.Errorf("set-password: upsert user: %w", err)
			}

			_, _ = fmt.Fprintln(os.Stdout, "Password updated successfully.")

			return nil
		},
	}
}

func promptPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	raw, err := term.ReadPassword(int(syscall.Stdin))

	fmt.Fprintln(os.Stderr)

	if err != nil {
		return "", err
	}

	return string(raw), nil
}
