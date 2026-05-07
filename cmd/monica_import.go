package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/nhymxu/kith-pms/internal/dates"
	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/labels"
	"github.com/nhymxu/kith-pms/internal/monica"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/reminders"
	"github.com/nhymxu/kith-pms/pkg/config"
)

func monicaImportCommand() *cli.Command {
	return &cli.Command{
		Name:  "monica-import",
		Usage: "Import contacts from a Monica JSON export file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "from",
				Usage:    "Path to Monica JSON export file",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Parse and report without writing to the database",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fromPath := cmd.String("from")
			dryRun := cmd.Bool("dry-run")

			f, err := os.Open(fromPath)
			if err != nil {
				return fmt.Errorf("monica-import: open %q: %w", fromPath, err)
			}
			defer func() { _ = f.Close() }()

			export, err := monica.Parse(f)
			if err != nil {
				return fmt.Errorf("monica-import: parse: %w", err)
			}

			fmt.Printf("Parsed %d contacts from %s\n", len(export.Contacts), fromPath)

			if dryRun {
				return printDryRunSummary(export)
			}

			database, err := internaldb.Open(config.ENV.DBPath)
			if err != nil {
				return fmt.Errorf("monica-import: open db: %w", err)
			}
			defer func() { _ = database.Close() }()

			if err := internaldb.Up(database); err != nil {
				return fmt.Errorf("monica-import: migrations: %w", err)
			}

			peopleSvc := people.NewService(database)
			labelsSvc := labels.NewService(database)
			journalSvc := journal.NewService(database)
			remindersSvc := reminders.NewService(database)
			datesSvc := dates.NewService(database)

			return runImport(ctx, export, peopleSvc, labelsSvc, journalSvc, remindersSvc, datesSvc)
		},
	}
}

func runImport(
	ctx context.Context,
	export *monica.Export,
	peopleSvc *people.Service,
	labelsSvc *labels.Service,
	journalSvc *journal.Service,
	remindersSvc *reminders.Service,
	datesSvc *dates.Service,
) error {
	existingLabels, err := labelsSvc.List(ctx)
	if err != nil {
		return fmt.Errorf("monica-import: load labels: %w", err)
	}

	labelMap := make(map[string]int64, len(existingLabels))
	for _, l := range existingLabels {
		labelMap[strings.ToLower(l.Name)] = l.ID
	}

	var imported, errCount int

	for _, c := range export.Contacts {
		rec := monica.MapContact(c)
		if rec.Person.Name == "" {
			slog.Warn("monica-import: skipping contact with empty name", "id", c.ID)

			errCount++

			continue
		}

		personID, err := peopleSvc.Create(ctx, rec.Person, rec.Contacts, rec.Locations)
		if err != nil {
			slog.Warn("monica-import: failed to create person", "name", rec.Person.Name, "err", err)

			errCount++

			continue
		}

		for _, tagName := range rec.TagNames {
			key := strings.ToLower(tagName)

			labelID, ok := labelMap[key]
			if !ok {
				labelID, err = labelsSvc.Create(ctx, tagName, "#6366f1")
				if err != nil {
					slog.Warn("monica-import: failed to create label", "name", tagName, "err", err)
					continue
				}

				labelMap[key] = labelID
			}

			if err := labelsSvc.Attach(ctx, personID, labelID); err != nil {
				slog.Warn("monica-import: failed to attach label", "person_id", personID, "label", tagName, "err", err)
			}
		}

		for _, act := range rec.Activities {
			if _, err := journalSvc.Create(ctx, act, []int64{personID}); err != nil {
				slog.Warn("monica-import: failed to create activity", "person_id", personID, "err", err)
			}
		}

		for i := range rec.Reminders {
			rec.Reminders[i].PersonID = &personID
			if _, err := remindersSvc.Create(ctx, &rec.Reminders[i]); err != nil {
				slog.Warn("monica-import: failed to create reminder", "person_id", personID, "err", err)
			}
		}

		if len(rec.Dates) > 0 {
			if err := datesSvc.ReplaceForPerson(ctx, personID, rec.Dates); err != nil {
				slog.Warn("monica-import: failed to save dates", "person_id", personID, "err", err)
			}
		}

		imported++

		slog.Info("monica-import: imported contact", "name", rec.Person.Name, "person_id", personID)
	}

	fmt.Printf("\nImport complete: %d imported, %d skipped/errors\n", imported, errCount)

	return nil
}

func printDryRunSummary(export *monica.Export) error {
	var totalContacts, totalLocations, totalTags, totalActivities, totalReminders, totalDates int

	for _, c := range export.Contacts {
		rec := monica.MapContact(c)
		totalContacts += len(rec.Contacts)
		totalLocations += len(rec.Locations)
		totalTags += len(rec.TagNames)
		totalActivities += len(rec.Activities)
		totalReminders += len(rec.Reminders)
		totalDates += len(rec.Dates)
	}

	fmt.Printf("\nDry-run summary:\n")
	fmt.Printf("  Contacts (people):   %d\n", len(export.Contacts))
	fmt.Printf("  Contact info:        %d\n", totalContacts)
	fmt.Printf("  Locations:           %d\n", totalLocations)
	fmt.Printf("  Tags (labels):       %d\n", totalTags)
	fmt.Printf("  Journal entries:     %d\n", totalActivities)
	fmt.Printf("  Reminders:           %d\n", totalReminders)
	fmt.Printf("  Important dates:     %d\n", totalDates)

	return nil
}
