package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/uptrace/bun"
	"github.com/urfave/cli/v3"

	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/internal/files"
	"github.com/nhymxu/kith-pms/internal/gifts"
	"github.com/nhymxu/kith-pms/internal/important_dates"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/monica"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/relationships"
	"github.com/nhymxu/kith-pms/internal/reminders"
	"github.com/nhymxu/kith-pms/internal/work_history"
	"github.com/nhymxu/kith-pms/pkg/config"
)

func monicaImportCommand() *cli.Command {
	return &cli.Command{
		Name:  "monica-import",
		Usage: "Import contacts from a Monica v4 JSON export file",
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
			&cli.StringFlag{
				Name:  "inactive-reminders",
				Usage: "How to handle inactive Monica reminders: ask, skip, or completed",
				Value: "ask",
			},
			&cli.StringFlag{
				Name:  "account-journal",
				Usage: "How to handle account-level Monica journal entries: ask, skip, or unlinked",
				Value: "ask",
			},
			&cli.StringFlag{
				Name:  "self-profile",
				Usage: "How to handle self-profile selection: ask or skip",
				Value: "ask",
			},
			&cli.StringFlag{
				Name:  "name-order",
				Usage: "How to combine name parts: ask, western (first middle last), or eastern (last middle first)",
				Value: "ask",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fromPath := cmd.String("from")
			dryRun := cmd.Bool("dry-run")

			const maxExportBytes = 200 * 1024 * 1024
			if fi, err := os.Stat(fromPath); err == nil && fi.Size() > maxExportBytes {
				fmt.Printf(
					"Warning: export file is %.0f MB (>200 MB).\n"+
						"The importer reads the whole file into memory — ensure sufficient RAM.\n",
					float64(fi.Size())/(1024*1024),
				)
			}

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

			options, err := resolveMonicaImportOptions(
				export,
				cmd.String("inactive-reminders"),
				cmd.String("account-journal"),
				cmd.String("self-profile"),
				cmd.String("name-order"),
			)
			if err != nil {
				return err
			}

			if dryRun {
				return printDryRunSummary(export, options)
			}

			database, err := internaldb.Open(config.C.DBPath, config.C.DBMaxOpenConns)
			if err != nil {
				return fmt.Errorf("monica-import: open db: %w", err)
			}
			defer func() { _ = database.Close() }()

			if err := internaldb.Up(database); err != nil {
				return fmt.Errorf("monica-import: migrations: %w", err)
			}

			peopleSvc := people.NewService(database)
			labelsSvc := people.NewLabelService(database)
			journalSvc := journal.NewService(database)
			journalLabelSvc := journal.NewLabelService(database)
			remindersSvc := reminders.NewService(database)
			datesSvc := important_dates.NewService(database)
			giftsSvc := gifts.NewService(database)
			workSvc := work_history.NewService(database)
			relSvc := relationships.NewService(database)
			filesSvc := files.NewLocalFileService(config.C.AvatarStoragePath)

			return runImport(
				ctx,
				export,
				options,
				filesSvc,
				peopleSvc,
				labelsSvc,
				journalSvc,
				journalLabelSvc,
				remindersSvc,
				datesSvc,
				giftsSvc,
				workSvc,
				relSvc,
			)
		},
	}
}

func runImport(
	ctx context.Context,
	export *monica.Export,
	options monica.ImportOptions,
	filesSvc *files.LocalFileService,
	peopleSvc *people.Service,
	labelsSvc *people.LabelService,
	journalSvc *journal.Service,
	journalLabelSvc *journal.LabelService,
	remindersSvc *reminders.Service,
	datesSvc *important_dates.Service,
	giftsSvc *gifts.Service,
	workSvc *work_history.Service,
	relSvc *relationships.Service,
) error {
	// Seed journal labels once; fail fast — labels are a precondition for conversation/life-event/document tagging.
	convLabelID, err := journalLabelSvc.FindOrCreate(ctx, monica.LabelConversation, "#6366f1")
	if err != nil {
		return fmt.Errorf("monica-import: seed CONVERSATION journal label: %w", err)
	}

	lifeLabelID, err := journalLabelSvc.FindOrCreate(ctx, monica.LabelLifeEvent, "#a855f7")
	if err != nil {
		return fmt.Errorf("monica-import: seed LIFE_EVENT journal label: %w", err)
	}

	docLabelID, err := journalLabelSvc.FindOrCreate(ctx, monica.LabelDocument, "#f59e0b")
	if err != nil {
		return fmt.Errorf("monica-import: seed DOCUMENT journal label: %w", err)
	}

	existingLabels, err := labelsSvc.List(ctx)
	if err != nil {
		return fmt.Errorf("monica-import: load labels: %w", err)
	}

	labelMap := make(map[string]int64, len(existingLabels))
	for _, l := range existingLabels {
		labelMap[strings.ToLower(l.Name)] = l.ID
	}

	// If a self-profile was selected, move that contact to the front so it is created first.
	if options.SelfContactUUID != "" {
		for i, c := range export.Contacts {
			if c.ID == options.SelfContactUUID {
				export.Contacts[0], export.Contacts[i] = export.Contacts[i], export.Contacts[0]
				break
			}
		}
	}

	// First pass: insert all persons and build UUID→personID map for relationship resolution.
	uuidToPersonID := make(map[string]int64, len(export.Contacts))

	// pendingActivities accumulates UUID-keyed Monica activities across contacts so that
	// activities with multiple participants are created as a single journal entry.
	type pendingActivity struct {
		Activity  journal.Activity
		PersonIDs []int64
	}

	pendingActivities := make(map[string]*pendingActivity)

	var imported, errCount, avatarImported, avatarSkipped int

	for _, c := range export.Contacts {
		rec := monica.MapContactWithOptions(c, options)
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

		if c.ID != "" {
			uuidToPersonID[c.ID] = personID
		}

		if options.SelfContactUUID != "" && c.ID == options.SelfContactUUID {
			if err := peopleSvc.SetSelf(ctx, personID); err != nil {
				slog.Warn("monica-import: failed to mark self-profile", "person_id", personID, "err", err)
			} else {
				slog.Info("monica-import: self-profile set", "name", rec.Person.Name, "person_id", personID)
			}
		}

		// Attach labels.
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

		// Per-contact journal entries (notes, calls) — always unique to this contact.
		for _, act := range rec.Activities {
			if _, err := journalSvc.Create(ctx, act, []int64{personID}, nil); err != nil {
				slog.Warn("monica-import: failed to create activity", "person_id", personID, "err", err)
			}
		}

		// UUID-keyed Monica activities: accumulate participants for deduplication.
		for _, entry := range rec.ActivityEntries {
			if p, ok := pendingActivities[entry.UUID]; ok {
				p.PersonIDs = append(p.PersonIDs, personID)
			} else {
				pendingActivities[entry.UUID] = &pendingActivity{
					Activity:  entry.Activity,
					PersonIDs: []int64{personID},
				}
			}
		}

		// Conversations → journal entries tagged CONVERSATION.
		for _, act := range rec.ConversationActivities {
			if _, err := journalSvc.Create(ctx, act, []int64{personID}, []int64{convLabelID}); err != nil {
				slog.Warn("monica-import: failed to create conversation activity", "person_id", personID, "err", err)
			}
		}

		// Life events → journal entries tagged LIFE_EVENT.
		for _, act := range rec.LifeEventActivities {
			if _, err := journalSvc.Create(ctx, act, []int64{personID}, []int64{lifeLabelID}); err != nil {
				slog.Warn("monica-import: failed to create life event activity", "person_id", personID, "err", err)
			}
		}

		// Reminders.
		for i := range rec.Reminders {
			rec.Reminders[i].PersonID = &personID
			if _, err := remindersSvc.Create(ctx, &rec.Reminders[i]); err != nil {
				slog.Warn("monica-import: failed to create reminder", "person_id", personID, "err", err)
			}
		}

		// Important dates.
		if len(rec.Dates) > 0 {
			if err := datesSvc.ReplaceForPerson(ctx, personID, rec.Dates); err != nil {
				slog.Warn("monica-import: failed to save dates", "person_id", personID, "err", err)
			}
		}

		// Work history.
		if len(rec.WorkHistory) > 0 {
			for i := range rec.WorkHistory {
				rec.WorkHistory[i].PersonID = personID
			}

			if err := workSvc.ReplaceForPerson(ctx, personID, rec.WorkHistory); err != nil {
				slog.Warn("monica-import: failed to save work history", "person_id", personID, "err", err)
			}
		}

		// Gifts.
		for i := range rec.Gifts {
			rec.Gifts[i].PersonID = personID
			if _, err := giftsSvc.Create(ctx, &rec.Gifts[i]); err != nil {
				slog.Warn("monica-import: failed to create gift", "person_id", personID, "err", err)
			}
		}

		// Avatar: decode base64 dataUrl and save to disk when present.
		if rec.AvatarDataURL != "" {
			if saveAvatar(ctx, peopleSvc.DB, filesSvc, peopleSvc.People, personID, rec.AvatarDataURL) {
				avatarImported++
			} else {
				avatarSkipped++
			}
		}

		// Documents: decode and store each embedded document, then create a DOCUMENT journal entry.
		for _, doc := range rec.Documents {
			saveMonicaDocument(ctx, journalSvc, filesSvc, personID, docLabelID, doc)
		}

		imported++

		slog.Info("monica-import: imported contact", "name", rec.Person.Name, "person_id", personID)
	}

	// Create deduplicated Monica activities with all participants combined.
	for _, p := range pendingActivities {
		if _, err := journalSvc.Create(ctx, p.Activity, p.PersonIDs, nil); err != nil {
			slog.Warn("monica-import: failed to create activity", "err", err)
		}
	}

	// Second pass: resolve and insert relationships now that all persons exist.
	importRelationships(ctx, export, uuidToPersonID, relSvc)

	if options.ImportAccountJournalEntries {
		accountJournalImported, accountJournalErrors := importAccountJournalEntries(ctx, export, journalSvc)
		fmt.Printf(
			"Account journal entries: %d imported, %d skipped/errors\n",
			accountJournalImported,
			accountJournalErrors,
		)
	}

	if avatarImported+avatarSkipped > 0 {
		fmt.Printf("Avatars: %d imported, %d skipped\n", avatarImported, avatarSkipped)
	}

	fmt.Printf("\nImport complete: %d imported, %d skipped/errors\n", imported, errCount)

	return nil
}

// importRelationships resolves Monica UUID-based relationships to kith integer IDs.
func importRelationships(
	ctx context.Context,
	export *monica.Export,
	uuidToPersonID map[string]int64,
	relSvc *relationships.Service,
) {
	if len(uuidToPersonID) == 0 {
		return
	}

	// Build a type name cache to avoid repeated DB lookups.
	typeCache := make(map[string]int64)

	for _, c := range export.Contacts {
		fromID, ok := uuidToPersonID[c.ID]
		if !ok {
			continue
		}

		for _, rel := range c.Relationships {
			toID, ok := uuidToPersonID[rel.ToContactUUID]
			if !ok {
				continue
			}

			typeName := strings.TrimSpace(rel.TypeName)
			if typeName == "" {
				typeName = "other"
			}

			typeID, ok := typeCache[strings.ToLower(typeName)]
			if !ok {
				rt, err := relSvc.CreateType(ctx, typeName, "")
				if err != nil {
					// Type may already exist; try listing to find it.
					types, lerr := relSvc.ListTypes(ctx)
					if lerr != nil {
						slog.Warn("monica-import: failed to resolve relationship type", "type", typeName, "err", err)
						continue
					}

					for _, t := range types {
						if strings.EqualFold(t.Name, typeName) {
							typeID = t.ID
							break
						}
					}

					if typeID == 0 {
						slog.Warn("monica-import: skipping relationship, type not found", "type", typeName)
						continue
					}
				} else {
					typeID = rt.ID
				}

				typeCache[strings.ToLower(typeName)] = typeID
			}

			if _, err := relSvc.AttachRelationship(ctx, fromID, toID, typeID, ""); err != nil {
				if err != relationships.ErrDuplicateRelationship {
					slog.Warn("monica-import: failed to attach relationship",
						"from", fromID, "to", toID, "type", typeName, "err", err)
				}
			}
		}
	}
}

func resolveMonicaImportOptions(
	export *monica.Export,
	inactiveMode, accountJournalMode, selfProfileMode, nameOrderMode string,
) (monica.ImportOptions, error) {
	options := monica.ImportOptions{}
	reader := bufio.NewReader(os.Stdin)

	nameOrder, err := resolveNameOrder(reader, export, nameOrderMode)
	if err != nil {
		return options, err
	}

	options.NameOrder = nameOrder

	inactiveCount := countInactiveReminders(export)
	if inactiveCount > 0 {
		answer, err := resolveChoice(
			reader,
			inactiveMode,
			"completed",
			"skip",
			fmt.Sprintf(
				"Import %d inactive Monica reminders as completed reminders? No = skip permanently.",
				inactiveCount,
			),
		)
		if err != nil {
			return options, err
		}

		options.ImportInactiveReminders = answer
	}

	if len(export.AccountJournalEntries) > 0 {
		answer, err := resolveChoice(
			reader,
			accountJournalMode,
			"unlinked",
			"skip",
			fmt.Sprintf(
				"Import %d account-level Monica journal entries as unlinked journal entries? No = skip.",
				len(export.AccountJournalEntries),
			),
		)
		if err != nil {
			return options, err
		}

		options.ImportAccountJournalEntries = answer
	}

	if len(export.Contacts) > 0 && strings.ToLower(strings.TrimSpace(selfProfileMode)) != "skip" {
		uuid, err := resolveSelfProfile(reader, export, nameOrder)
		if err != nil {
			return options, err
		}

		options.SelfContactUUID = uuid
	}

	return options, nil
}

// resolveNameOrder asks how to combine first/middle/last into a full name.
// Returns monica.NameOrderWestern or monica.NameOrderEastern.
func resolveNameOrder(reader *bufio.Reader, export *monica.Export, mode string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case monica.NameOrderWestern:
		return monica.NameOrderWestern, nil
	case monica.NameOrderEastern:
		return monica.NameOrderEastern, nil
	case "", "ask":
		// fall through to interactive prompt
	default:
		return "", fmt.Errorf("monica-import: invalid --name-order %q, expected ask, western, or eastern", mode)
	}

	// Build an example from the first contact that has at least two name parts.
	exFirst, exMiddle, exLast := "John", "", "Smith"

	for _, c := range export.Contacts {
		first := strings.TrimSpace(c.FirstName)

		last := strings.TrimSpace(c.LastName)
		if first != "" && last != "" {
			exFirst = first
			exMiddle = strings.TrimSpace(c.MiddleName)
			exLast = last

			break
		}
	}

	westernExample := monica.BuildFullName(exFirst, exMiddle, exLast, monica.NameOrderWestern)
	easternExample := monica.BuildFullName(exFirst, exMiddle, exLast, monica.NameOrderEastern)

	fmt.Println("\nHow should full names be combined from first / middle / last?")
	fmt.Printf("  1. Western order  (first middle last)  e.g. %q\n", westernExample)
	fmt.Printf("  2. Eastern order  (last middle first)   e.g. %q\n", easternExample)

	for {
		fmt.Print("Enter 1 or 2: ")

		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("monica-import: read name-order choice: %w", err)
		}

		switch strings.TrimSpace(line) {
		case "1":
			fmt.Printf("Name order: Western (%s)\n", westernExample)
			return monica.NameOrderWestern, nil
		case "2":
			fmt.Printf("Name order: Eastern (%s)\n", easternExample)
			return monica.NameOrderEastern, nil
		default:
			fmt.Println("Please enter 1 or 2.")
		}
	}
}

// resolveSelfProfile asks the user whether to designate one contact as the self-profile.
// Returns the chosen contact UUID, or "" if the user declines.
func resolveSelfProfile(reader *bufio.Reader, export *monica.Export, nameOrder string) (string, error) {
	want, err := askYesNo(reader, fmt.Sprintf("Select a self-profile from %d contacts?", len(export.Contacts)))
	if err != nil || !want {
		return "", err
	}

	fmt.Println("\nContacts:")

	contactName := func(c monica.Contact) string {
		name := monica.BuildFullName(c.FirstName, c.MiddleName, c.LastName, nameOrder)
		if name == "" {
			name = c.Nickname
		}

		if name == "" {
			name = "(no name)"
		}

		return name
	}

	for i, c := range export.Contacts {
		fmt.Printf("  %3d. %s\n", i+1, contactName(c))
	}

	for {
		fmt.Printf("Enter number (1-%d): ", len(export.Contacts))

		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("monica-import: read self-profile choice: %w", err)
		}

		line = strings.TrimSpace(line)

		var n int
		if _, scanErr := fmt.Sscanf(line, "%d", &n); scanErr != nil || n < 1 || n > len(export.Contacts) {
			fmt.Printf("Please enter a number between 1 and %d.\n", len(export.Contacts))
			continue
		}

		chosen := export.Contacts[n-1]

		fmt.Printf("Self-profile set to: %s\n", contactName(chosen))

		return chosen.ID, nil
	}
}

func resolveChoice(reader *bufio.Reader, mode, yesMode, noMode, question string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case yesMode:
		return true, nil
	case "", "ask":
		return askYesNo(reader, question)
	case noMode:
		return false, nil
	default:
		return false, fmt.Errorf("monica-import: invalid option %q, expected ask, %s, or %s", mode, noMode, yesMode)
	}
}

func askYesNo(reader *bufio.Reader, question string) (bool, error) {
	for {
		fmt.Printf("%s [y/N]: ", question)

		answer, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("monica-import: read confirmation: %w", err)
		}

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "", "n", "no":
			return false, nil
		case "y", "yes":
			return true, nil
		default:
			fmt.Println("Please answer y or n.")
		}
	}
}

func countInactiveReminders(export *monica.Export) int {
	var count int

	for _, c := range export.Contacts {
		for _, reminder := range c.Reminders {
			if reminder.Inactive {
				count++
			}
		}
	}

	return count
}

func importAccountJournalEntries(ctx context.Context, export *monica.Export, journalSvc *journal.Service) (int, int) {
	var imported, errCount int

	for _, act := range monica.MapAccountJournalEntries(export.AccountJournalEntries) {
		if _, err := journalSvc.Create(ctx, act, nil, nil); err != nil {
			slog.Warn("monica-import: failed to create account journal entry", "title", act.Title, "err", err)

			errCount++

			continue
		}

		imported++
	}

	return imported, errCount
}

func printDryRunSummary(export *monica.Export, options monica.ImportOptions) error {
	var totalContacts, totalLocations, totalTags, totalActivities, totalReminders, totalDates, totalGifts, totalWork, totalRels, totalAvatars, totalDocuments int //nolint:lll

	var totalConversations, totalMessages, totalLifeEvents int

	for _, c := range export.Contacts {
		rec := monica.MapContactWithOptions(c, options)
		totalContacts += len(rec.Contacts)
		totalLocations += len(rec.Locations)
		totalTags += len(rec.TagNames)
		totalActivities += len(rec.Activities) + len(rec.ActivityEntries)
		totalReminders += len(rec.Reminders)
		totalDates += len(rec.Dates)
		totalGifts += len(rec.Gifts)
		totalWork += len(rec.WorkHistory)
		totalRels += len(rec.Relationships)
		totalConversations += len(rec.ConversationActivities)
		totalDocuments += len(rec.Documents)

		totalLifeEvents += len(rec.LifeEventActivities)
		for _, conv := range c.Conversations {
			totalMessages += len(conv.Messages)
		}

		if rec.AvatarDataURL != "" {
			totalAvatars++
		}
	}

	fmt.Printf("\nDry-run summary:\n")
	fmt.Printf("  Contacts (people):   %d\n", len(export.Contacts))
	fmt.Printf("  Contact info:        %d\n", totalContacts)
	fmt.Printf("  Locations:           %d\n", totalLocations)
	fmt.Printf("  Tags (labels):       %d\n", totalTags)
	fmt.Printf("  Journal entries:     %d\n", totalActivities)
	fmt.Printf("  Conversations:       %d (%d messages → CONVERSATION label)\n", totalConversations, totalMessages)
	fmt.Printf("  Life events:         %d (→ LIFE_EVENT label + ImportantDate)\n", totalLifeEvents)

	if options.ImportAccountJournalEntries {
		fmt.Printf("  Account journals:    %d\n", len(monica.MapAccountJournalEntries(export.AccountJournalEntries)))
	} else if len(export.AccountJournalEntries) > 0 {
		fmt.Printf("  Account journals:    0 (%d skipped)\n", len(export.AccountJournalEntries))
	}

	fmt.Printf("  Reminders:           %d\n", totalReminders)

	if skippedInactive := countInactiveReminders(export); skippedInactive > 0 && !options.ImportInactiveReminders {
		fmt.Printf("  Inactive reminders:  0 (%d skipped)\n", skippedInactive)
	}

	fmt.Printf("  Important dates:     %d\n", totalDates)
	fmt.Printf("  Gifts:               %d\n", totalGifts)
	fmt.Printf("  Work history:        %d\n", totalWork)
	fmt.Printf("  Relationships:       %d\n", totalRels)
	fmt.Printf("  Avatars:             %d\n", totalAvatars)
	fmt.Printf("  Documents:           %d (→ DOCUMENT journal entries)\n", totalDocuments)

	if export.AccountDocumentCount > 0 {
		fmt.Printf("  Account documents:   0 (%d skipped — no contact link)\n", export.AccountDocumentCount)
	}

	return nil
}

// saveAvatar decodes a "data:<mime>;base64,..." dataUrl, writes it to disk, and
// updates the person row with file path metadata.
// Returns true on success; logs and returns false so the caller's import loop continues.
func saveAvatar(
	ctx context.Context,
	db *bun.DB,
	filesSvc *files.LocalFileService,
	personRepo people.PersonRepo,
	personID int64,
	dataURL string,
) bool {
	mimeType, data, err := parseDataURL(dataURL)
	if err != nil {
		slog.Warn("monica-import: skip avatar, bad data URL", "person_id", personID, "err", err)
		return false
	}

	path, err := filesSvc.SaveAvatarBytes(personID, data, mimeType)
	if err != nil {
		slog.Warn("monica-import: skip avatar, save failed", "person_id", personID, "err", err)
		return false
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		_ = filesSvc.DeleteAvatar(personID, path)
		slog.Warn("monica-import: skip avatar, begin tx failed", "person_id", personID, "err", err)

		return false
	}
	defer func() { _ = tx.Rollback() }()

	if err := personRepo.UpdateAvatar(
		ctx,
		tx,
		personID,
		path,
		int64(len(data)),
	); err != nil {
		_ = filesSvc.DeleteAvatar(personID, path)
		slog.Warn("monica-import: skip avatar, db update failed", "person_id", personID, "err", err)

		return false
	}

	if err := tx.Commit(); err != nil {
		_ = filesSvc.DeleteAvatar(personID, path)
		slog.Warn("monica-import: skip avatar, commit failed", "person_id", personID, "err", err)

		return false
	}

	return true
}

// parseDataURL splits a "data:<mime>;base64,..." string using the avatar (5 MB) size limit.
func parseDataURL(dataURL string) (mimeType string, data []byte, err error) {
	const maxAvatarBytes = 5 * 1024 * 1024
	return parseDataURLLimit(dataURL, maxAvatarBytes)
}

// parseDataURLLimit splits a "data:<mime>;base64,..." string, rejecting payloads over maxBytes.
func parseDataURLLimit(dataURL string, maxBytes int) (mimeType string, data []byte, err error) {
	const prefix = "data:"
	if !strings.HasPrefix(dataURL, prefix) {
		return "", nil, fmt.Errorf("not a data URL")
	}

	rest := dataURL[len(prefix):]

	mimeType, encoded, ok := strings.Cut(rest, ";base64,")
	if !ok {
		return "", nil, fmt.Errorf("not a base64 data URL")
	}

	// Reject oversized payloads before allocating: base64 encodes 3 bytes as 4 chars.
	if len(encoded) > maxBytes*4/3+4 {
		return "", nil, fmt.Errorf("data URL exceeds size limit")
	}

	// Try standard encoding first; fall back to raw (no-padding) for Monica exports
	// that may omit trailing '=' characters.
	data, err = base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(encoded)
	}

	if err != nil {
		return "", nil, fmt.Errorf("decode base64: %w", err)
	}

	return mimeType, data, nil
}

// saveMonicaDocument decodes an embedded document dataUrl, stores it to disk, and creates a
// DOCUMENT-labelled journal entry linking it to the person.
func saveMonicaDocument(
	ctx context.Context,
	journalSvc *journal.Service,
	filesSvc *files.LocalFileService,
	personID int64,
	docLabelID int64,
	doc monica.MDocument,
) {
	const maxDocBytes = 50 * 1024 * 1024

	_, data, err := parseDataURLLimit(doc.DataURL, maxDocBytes)
	if err != nil {
		slog.Warn("monica-import: skip document, bad data URL",
			"person_id", personID, "filename", doc.OriginalFilename, "err", err)

		return
	}

	path, err := filesSvc.SaveDocument(personID, data, doc.OriginalFilename)
	if err != nil {
		slog.Warn("monica-import: skip document, save failed",
			"person_id", personID, "filename", doc.OriginalFilename, "err", err)

		return
	}

	title := monica.LabelDocument + " " + doc.OriginalFilename
	body := fmt.Sprintf(
		"Imported Monica document. file=%s type=%s size=%d",
		path, doc.MimeType, len(data),
	)

	act := journal.Activity{
		Title:   title,
		Content: body,
	}

	if _, err := journalSvc.Create(ctx, act, []int64{personID}, []int64{docLabelID}); err != nil {
		slog.Warn("monica-import: failed to create document journal entry",
			"person_id", personID, "filename", doc.OriginalFilename, "err", err)
	}
}
