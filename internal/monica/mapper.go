package monica

import (
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nhymxu/kith-pms/internal/gifts"
	"github.com/nhymxu/kith-pms/internal/important_dates"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/reminders"
	"github.com/nhymxu/kith-pms/internal/work_history"
)

const (
	LabelConversation = "CONVERSATION"
	LabelLifeEvent    = "LIFE_EVENT"
	LabelDocument     = "DOCUMENT"
	maxConvBodyRunes  = 8000
	maxMsgsPerConv    = 2000
)

// ActivityEntry pairs a Monica activity UUID with its mapped journal.Activity.
// Activities shared between multiple contacts carry the same UUID and are deduplicated
// during import so only one journal entry is created per Monica activity.
type ActivityEntry struct {
	UUID     string
	Activity journal.Activity
}

// ImportRecord holds all kith-pms domain objects for a single Monica contact.
type ImportRecord struct {
	Person    people.Person
	Contacts  []people.ContactInfo
	Locations []people.Location
	TagNames  []string // label names to create-or-find and attach
	// Activities holds per-contact entries (notes, calls) that are always unique to this contact.
	Activities []journal.Activity
	// ActivityEntries holds UUID-keyed Monica activities shared across contacts; deduplicated on import.
	ActivityEntries []ActivityEntry
	// ConversationActivities are mapped from Monica conversations; attach CONVERSATION label on create.
	ConversationActivities []journal.Activity
	// LifeEventActivities are mapped from Monica life events; attach LIFE_EVENT label on create.
	LifeEventActivities []journal.Activity
	Reminders           []reminders.Reminder
	Dates               []important_dates.ImportantDate
	WorkHistory         []work_history.WorkEntry
	Gifts               []gifts.Gift
	// Relationships are resolved after all persons are inserted (UUID→ID mapping needed).
	Relationships []MRelationship
	// AvatarDataURL is non-empty when the contact has a photo avatar in the Monica export.
	// Format: "data:<mime>;base64,<encoded>"
	AvatarDataURL string
	// Documents holds embedded documents to be stored and linked as DOCUMENT journal entries.
	Documents []MDocument
}

// MapContact converts a Monica Contact into an ImportRecord.
func MapContact(c Contact) ImportRecord {
	return MapContactWithOptions(c, ImportOptions{})
}

func MapContactWithOptions(c Contact, options ImportOptions) ImportRecord {
	convActivities := mapConversations(c)
	leActivities, leDates := mapLifeEvents(c)

	allDates := mapDates(c.Information)
	allDates = append(allDates, leDates...)

	return ImportRecord{
		Person:                 mapPerson(c, options.NameOrder),
		Contacts:               mapContactInfo(c.ContactInfo),
		Locations:              mapLocations(c.Addresses),
		TagNames:               mapTags(c.Tags),
		Activities:             mapNoteCallActivities(c),
		ActivityEntries:        mapActivityEntries(c),
		ConversationActivities: convActivities,
		LifeEventActivities:    leActivities,
		Reminders:              mapReminders(c.Reminders, c.Tasks, options),
		Dates:                  allDates,
		WorkHistory:            mapWorkHistory(c),
		Gifts:                  mapGifts(c.Gifts),
		Relationships:          c.Relationships,
		AvatarDataURL:          c.AvatarDataURL,
		Documents:              c.Documents,
	}
}

// BuildFullName joins first/middle/last according to nameOrder ("eastern" = last+middle+first,
// default/western = first+middle+last). Empty parts are omitted.
func BuildFullName(first, middle, last, nameOrder string) string {
	var order []string
	if nameOrder == NameOrderEastern {
		order = []string{last, middle, first}
	} else {
		order = []string{first, middle, last}
	}

	var parts []string

	for _, p := range order {
		if s := strings.TrimSpace(p); s != "" {
			parts = append(parts, s)
		}
	}

	return strings.Join(parts, " ")
}

func mapPerson(c Contact, nameOrder string) people.Person {
	name := BuildFullName(c.FirstName, c.MiddleName, c.LastName, nameOrder)
	if name == "" {
		name = c.Nickname
	}

	p := people.Person{
		Name:     name,
		Nickname: c.Nickname,
	}

	// Build OtherNotes from description + work info.
	var notesParts []string
	if c.Description != "" {
		notesParts = append(notesParts, c.Description)
	}

	var workParts []string
	if c.Job != "" {
		workParts = append(workParts, c.Job)
	}

	if c.Company != "" {
		workParts = append(workParts, "at "+c.Company)
	}

	if len(workParts) > 0 {
		notesParts = append(notesParts, "Work: "+strings.Join(workParts, " "))
	}

	p.OtherNotes = strings.Join(notesParts, "\n")

	// Parse birthdate into Person.DateOfBirth (both year-having and yearless).
	if c.Information.Birthdate != "" {
		var raw string

		if c.Information.IsYearUnknown {
			// Monica stores yearless as "0000-MM-DD"; convert to "--MM-DD".
			if len(c.Information.Birthdate) == 10 {
				raw = "--" + c.Information.Birthdate[5:]
			}
		} else {
			raw = c.Information.Birthdate
		}

		if raw != "" {
			if d, err := people.ParseDateOnly(raw); err == nil {
				p.DateOfBirth = &d
			}
		}
	}

	return p
}

func mapContactInfo(fields []ContactField) []people.ContactInfo {
	out := make([]people.ContactInfo, 0, len(fields))
	for i, f := range fields {
		if f.Data == "" {
			continue
		}

		ci := people.ContactInfo{
			Value:    f.Data,
			Position: i,
		}
		switch strings.ToLower(f.Type.Name) {
		case "email":
			ci.Type = "email"
		case "phone", "telephone":
			ci.Type = "phone"
		case "twitter", "facebook", "linkedin", "github", "instagram":
			ci.Type = "social"
			ci.Label = f.Type.Name
		default:
			ci.Type = "other"
			ci.Label = f.Type.Name
		}

		out = append(out, ci)
	}

	return out
}

func mapLocations(addrs []Address) []people.Location {
	out := make([]people.Location, 0, len(addrs))
	for i, a := range addrs {
		loc := people.Location{
			Type:       strings.ToLower(a.Name),
			Address:    a.Street,
			City:       a.City,
			Country:    a.Country,
			PostalCode: a.PostalCode,
			Position:   i,
		}
		switch loc.Type {
		case "home", "work":
			// valid
		default:
			loc.Type = "other"
		}

		out = append(out, loc)
	}

	return out
}

func mapTags(tags []Tag) []string {
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		if n := strings.TrimSpace(t.Name); n != "" {
			names = append(names, n)
		}
	}

	return names
}

func MapAccountJournalEntries(entries []MAccountJournal) []journal.Activity {
	out := make([]journal.Activity, 0, len(entries))
	for _, entry := range entries {
		if entry.Content == "" {
			continue
		}

		title := strings.TrimSpace(entry.Title)
		if title == "" {
			title = truncate(entry.Content, 60)
		}

		out = append(out, journal.Activity{
			Title:          title,
			Content:        entry.Content,
			OccurredAtDate: dateFromISO(entry.OccurredAtDate),
		})
	}

	return out
}

// mapNoteCallActivities converts per-contact notes and calls to journal entries.
// These are always unique to a single contact and do not need deduplication.
func mapNoteCallActivities(c Contact) []journal.Activity {
	var out []journal.Activity

	for _, n := range c.Notes {
		if n.Body == "" {
			continue
		}

		out = append(out, journal.Activity{
			Title:          truncate(n.Body, 60),
			Content:        n.Body,
			OccurredAtDate: dateFromISO(n.CreatedAt),
		})
	}

	// Calls → journal entries (v4 only)
	for _, call := range c.Calls {
		if call.Content == "" && call.CalledAt == "" {
			continue
		}

		title := "Call"
		if call.CalledAt != "" {
			title = "Call on " + dateFromISO(call.CalledAt)
		}

		out = append(out, journal.Activity{
			Title:          title,
			Content:        call.Content,
			OccurredAtDate: dateFromISO(call.CalledAt),
		})
	}

	return out
}

// mapActivityEntries converts UUID-keyed Monica activities to ActivityEntry values.
// These can be shared across contacts; the importer deduplicates by UUID so only
// one journal entry is created per activity regardless of how many participants it has.
func mapActivityEntries(c Contact) []ActivityEntry {
	var out []ActivityEntry

	for _, a := range c.Activities {
		if a.UUID == "" {
			continue
		}

		title := strings.TrimSpace(a.Summary)
		if title == "" {
			title = truncate(a.Description, 60)
		}

		if title == "" {
			continue
		}

		out = append(out, ActivityEntry{
			UUID: a.UUID,
			Activity: journal.Activity{
				Title:          title,
				Content:        a.Description,
				OccurredAtDate: a.HappenedAt,
			},
		})
	}

	return out
}

// mapReminders maps Monica reminders and incomplete tasks to kith reminders.
func mapReminders(mrs []MReminder, tasks []MTask, options ImportOptions) []reminders.Reminder {
	out := make([]reminders.Reminder, 0, len(mrs)+len(tasks))

	for _, r := range mrs {
		if r.Title == "" || r.InitialDate == "" || (r.Inactive && !options.ImportInactiveReminders) {
			continue
		}

		t, err := time.Parse("2006-01-02", r.InitialDate)
		if err != nil {
			continue
		}

		reminder := reminders.Reminder{
			Title:   r.Title,
			Notes:   r.Description,
			DueDate: t,
		}
		if r.Inactive {
			reminder.Completed = true
			completedAt := t
			reminder.CompletedAt = &completedAt
		}

		out = append(out, reminder)
	}

	// Incomplete tasks become reminders due today.
	for _, task := range tasks {
		if task.Title == "" || task.Completed {
			continue
		}

		out = append(out, reminders.Reminder{
			Title:   task.Title,
			Notes:   task.Description,
			DueDate: time.Now().UTC().Truncate(24 * time.Hour),
		})
	}

	return out
}

func mapDates(info Information) []important_dates.ImportantDate {
	var out []important_dates.ImportantDate

	if info.FirstMetDate != "" {
		out = append(out, important_dates.ImportantDate{
			Kind:      string(important_dates.KindMet),
			DateValue: info.FirstMetDate,
			Recurring: false,
		})
	}

	return out
}

// mapWorkHistory converts Monica job/company into a single work_history entry when present.
func mapWorkHistory(c Contact) []work_history.WorkEntry {
	if c.Job == "" && c.Company == "" {
		return nil
	}

	return []work_history.WorkEntry{
		{
			Company:   c.Company,
			Title:     c.Job,
			StartDate: "2000", // Monica has no start date; use placeholder
		},
	}
}

// mapGifts converts Monica gift records to kith Gift domain objects.
func mapGifts(mgifts []MGift) []gifts.Gift {
	out := make([]gifts.Gift, 0, len(mgifts))
	for _, g := range mgifts {
		if g.Name == "" {
			continue
		}

		kg := gifts.Gift{
			Title:    g.Name,
			Notes:    g.Comment,
			Date:     g.Date,
			Currency: "USD",
		}

		// Map Monica status to kith direction.
		switch strings.ToLower(g.Status) {
		case "given":
			kg.Direction = gifts.DirectionGiven
		case "received":
			kg.Direction = gifts.DirectionReceived
		default:
			kg.Direction = gifts.DirectionPlanned
		}

		// Convert float amount to cents.
		if g.Amount > 0 {
			cents := int64(math.Round(g.Amount * 100))
			kg.AmountCents = &cents
		}

		out = append(out, kg)
	}

	return out
}

// sanitizeContent strips NUL and C0 control characters (except \n and \t) from untrusted input.
func sanitizeContent(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		if r == '\n' || r == '\t' || r >= 0x20 {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// mapConversations converts Monica conversations to journal activities (one per conversation).
func mapConversations(c Contact) []journal.Activity {
	var out []journal.Activity

	for _, conv := range c.Conversations {
		msgs := conv.Messages
		if len(msgs) > maxMsgsPerConv {
			msgs = msgs[:maxMsgsPerConv]
		}

		if len(msgs) == 0 {
			continue
		}

		happenedAt := dateFromISO(conv.HappenedAt)
		title := LabelConversation + " " + happenedAt

		lines := make([]string, 0, len(msgs))
		for _, m := range msgs {
			author := "them"
			if m.WrittenByMe {
				author = "me"
			}

			ts := ""
			if len(m.WrittenAt) >= 16 {
				ts = ", " + m.WrittenAt[11:16]
			}

			lines = append(lines, "MESSAGE ["+author+ts+"] "+sanitizeContent(m.Content))
		}

		body := truncate(strings.Join(lines, "\n"), maxConvBodyRunes)

		out = append(out, journal.Activity{
			Title:          title,
			Content:        body,
			OccurredAtDate: happenedAt,
		})
	}

	return out
}

// mapLifeEvents converts Monica life events to journal activities and ImportantDate records.
func mapLifeEvents(c Contact) ([]journal.Activity, []important_dates.ImportantDate) {
	var (
		activities     []journal.Activity
		importantDates []important_dates.ImportantDate
	)

	for _, le := range c.LifeEvents {
		name := strings.TrimSpace(le.Name)
		if name == "" {
			continue
		}

		happenedAt := dateFromISO(le.HappenedAt)
		note := truncate(sanitizeContent(le.Note), maxConvBodyRunes)

		activities = append(activities, journal.Activity{
			Title:          LabelLifeEvent + " " + name,
			Content:        note,
			OccurredAtDate: happenedAt,
		})

		importantDates = append(importantDates, important_dates.ImportantDate{
			Kind:      string(important_dates.KindOther),
			Label:     name,
			DateValue: happenedAt,
			Recurring: false,
		})
	}

	return activities, importantDates
}

// truncate returns the first n runes of s (UTF-8 safe).
func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= n {
		return s
	}

	return string([]rune(s)[:n])
}

// dateFromISO extracts "YYYY-MM-DD" from an ISO 8601 timestamp.
func dateFromISO(ts string) string {
	if len(ts) >= 10 {
		return ts[:10]
	}

	return time.Now().Format("2006-01-02")
}
