package calendar

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/important_dates"
	"github.com/nhymxu/kith-pms/internal/reminders"
)

// Part is one selectable slice of feed content.
type Part string

const (
	PartBirthdays Part = "birthdays" // person.date_of_birth
	PartDates     Part = "dates"     // important_date rows (anniversaries etc.)
	PartReminders Part = "reminders"
)

// AllParts is the default content of the combined feed.
var AllParts = []Part{PartBirthdays, PartDates, PartReminders}

// ParseParts reads a comma-separated list, dropping unknown names and duplicates.
func ParseParts(csv string) []Part {
	var out []Part

	for _, raw := range strings.Split(csv, ",") {
		p := Part(strings.TrimSpace(raw))
		if slices.Contains(AllParts, p) && !slices.Contains(out, p) {
			out = append(out, p)
		}
	}

	return out
}

// yearlessAnchorYear is a leap year so a --02-29 date survives parsing.
const yearlessAnchorYear = 2000

// Builder renders feeds from the live database.
type Builder struct {
	DB        *bun.DB
	Reminders *reminders.Repo
	Now       func() time.Time
}

func (b *Builder) now() time.Time {
	if b.Now != nil {
		return b.Now()
	}

	return time.Now()
}

// Build returns the iCalendar document named calName holding the given parts.
func (b *Builder) Build(ctx context.Context, calName string, parts ...Part) (string, error) {
	var events []event

	if slices.Contains(parts, PartBirthdays) || slices.Contains(parts, PartDates) {
		dates, err := b.dateEvents(ctx, slices.Contains(parts, PartBirthdays), slices.Contains(parts, PartDates))
		if err != nil {
			return "", err
		}

		events = append(events, dates...)
	}

	if slices.Contains(parts, PartReminders) {
		rems, err := b.reminderEvents(ctx)
		if err != nil {
			return "", err
		}

		events = append(events, rems...)
	}

	stamp := b.now()

	var w writer

	w.begin(calName)

	for _, e := range events {
		w.event(e, stamp)
	}

	w.end()

	return w.sb.String(), nil
}

type dateRow struct {
	PersonID  int64  `bun:"person_id"`
	DateID    int64  `bun:"date_id"`
	Name      string `bun:"name"`
	Nickname  string `bun:"nickname"`
	Kind      string `bun:"kind"`
	Label     string `bun:"label"`
	DateValue string `bun:"date_value"`
	Recurring bool   `bun:"recurring"`
	Notes     string `bun:"notes"`
}

// dateEvents covers person.date_of_birth (always yearly) and important_date rows.
// Archived and soft-deleted people are excluded, matching the people list.
func (b *Builder) dateEvents(ctx context.Context, birthdays, dates bool) ([]event, error) {
	const q = `
		SELECT p.id AS person_id, 0 AS date_id, p.name AS name, COALESCE(p.nickname, '') AS nickname,
		       'birthday' AS kind, '' AS label, p.date_of_birth AS date_value, 1 AS recurring, '' AS notes
		  FROM person p
		 WHERE p.deleted_at IS NULL AND p.archived_at IS NULL
		   AND p.date_of_birth IS NOT NULL AND p.date_of_birth != ''
		UNION ALL
		SELECT p.id, d.id, p.name, COALESCE(p.nickname, ''),
		       d.kind, COALESCE(d.label, ''), d.date_value, d.recurring, COALESCE(d.notes, '')
		  FROM important_date d
		  JOIN person p ON p.id = d.person_id
		 WHERE p.deleted_at IS NULL AND p.archived_at IS NULL
		 ORDER BY 1, 2`

	var rows []dateRow
	if err := b.DB.NewRaw(q).Scan(ctx, &rows); err != nil {
		return nil, fmt.Errorf("calendar: query dates: %w", err)
	}

	events := make([]event, 0, len(rows))

	for _, r := range rows {
		if (r.Kind == "birthday" && !birthdays) || (r.Kind != "birthday" && !dates) {
			continue
		}

		e, ok := dateRowToEvent(r)
		if ok {
			events = append(events, e)
		}
	}

	return events, nil
}

func dateRowToEvent(r dateRow) (event, bool) {
	canonical, yearless, err := important_dates.ParseFlexible(r.DateValue)
	if err != nil {
		return event{}, false
	}

	var start time.Time

	if yearless {
		start, err = time.Parse("2006-01-02", fmt.Sprintf("%d%s", yearlessAnchorYear, canonical[1:]))
	} else {
		start, err = time.Parse("2006-01-02", canonical)
	}

	if err != nil {
		return event{}, false
	}

	who := r.Name
	if r.Nickname != "" {
		who = fmt.Sprintf("%s (%s)", r.Name, r.Nickname)
	}

	e := event{start: start, allDay: true}

	if r.Recurring {
		e.rrule = "FREQ=YEARLY"
	}

	var desc []string

	if r.Kind == "birthday" {
		e.uid = fmt.Sprintf("kith-birthday-%d@kith", r.PersonID)
		e.summary = who + "'s birthday"

		if !yearless {
			desc = append(desc, fmt.Sprintf("Born %d", start.Year()))
		}
	} else {
		label := r.Label
		if label == "" {
			label = "Date"
			if r.Kind != "" {
				label = strings.ToUpper(r.Kind[:1]) + r.Kind[1:]
			}
		}

		e.uid = fmt.Sprintf("kith-date-%d@kith", r.DateID)
		e.summary = label + ": " + who

		if !yearless {
			desc = append(desc, fmt.Sprintf("Since %d", start.Year()))
		}
	}

	if r.Notes != "" {
		desc = append(desc, r.Notes)
	}

	e.description = strings.Join(desc, "\n")

	return e, true
}

// reminderEvents lists incomplete reminders. Birthday-type reminders are skipped
// because the birthdays feed already carries them from the person record.
func (b *Builder) reminderEvents(ctx context.Context) ([]event, error) {
	items, err := b.Reminders.List(ctx, reminders.ListParams{Status: "pending"})
	if err != nil {
		return nil, fmt.Errorf("calendar: list reminders: %w", err)
	}

	var deleted []int64
	if err := b.DB.NewRaw(`SELECT id FROM person WHERE deleted_at IS NOT NULL`).Scan(ctx, &deleted); err != nil {
		return nil, fmt.Errorf("calendar: list deleted people: %w", err)
	}

	events := make([]event, 0, len(items))

	for _, it := range items {
		if it.IsBirthday() || (it.PersonID != nil && slices.Contains(deleted, *it.PersonID)) {
			continue
		}

		events = append(events, reminderToEvent(it))
	}

	return events, nil
}

func reminderToEvent(it reminders.ReminderWithPerson) event {
	due := it.DueDate.UTC()
	// Always timed: the SPA stores due_date as a UTC datetime, so midnight UTC is not "all day".
	const allDay = false

	e := event{
		uid:     fmt.Sprintf("kith-reminder-%d@kith", it.ID),
		summary: it.Title,
		start:   due,
		allDay:  allDay,
		rrule:   reminderRRule(it.Reminder, allDay),
	}

	var desc []string

	if it.PersonName != "" {
		desc = append(desc, "Person: "+it.PersonName)
	}

	if it.Notes != "" {
		desc = append(desc, it.Notes)
	}

	e.description = strings.Join(desc, "\n")

	return e
}

var byDay = [7]string{"SU", "MO", "TU", "WE", "TH", "FR", "SA"}

// reminderRRule maps the rules that have a plain RFC 5545 equivalent.
// relative_contact and birthday depend on app state, so they export as one event.
// shortcut: monthly on the 29th-31st skips short months per RFC while the app clamps, upgrade if users report it.
func reminderRRule(r reminders.Reminder, allDay bool) string {
	rule := r.RecurrenceRule
	if rule == nil {
		return ""
	}

	var rr string

	switch rule.Type {
	case reminders.RecurrenceDaily:
		rr = "FREQ=DAILY"
	case reminders.RecurrenceWeekly:
		rr = "FREQ=WEEKLY"
	case reminders.RecurrenceMonthly:
		rr = "FREQ=MONTHLY"
	case reminders.RecurrenceYearly:
		rr = "FREQ=YEARLY"
	case reminders.RecurrenceCustom:
		interval := max(rule.Interval, 1)

		freq := "DAILY"

		switch rule.Unit {
		case "weeks":
			freq = "WEEKLY"
		case "months":
			freq = "MONTHLY"
		}

		rr = fmt.Sprintf("FREQ=%s;INTERVAL=%d", freq, interval)
	case reminders.RecurrenceDayOfWeek:
		if rule.DayOfWeek == nil || *rule.DayOfWeek < 0 || *rule.DayOfWeek > 6 {
			return ""
		}

		rr = "FREQ=WEEKLY;BYDAY=" + byDay[*rule.DayOfWeek]
	default:
		return ""
	}

	if r.RecurrenceEndDate != nil {
		end := r.RecurrenceEndDate.UTC()
		if allDay {
			rr += ";UNTIL=" + formatDate(end)
		} else {
			endOfDay := time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, time.UTC)
			rr += ";UNTIL=" + formatDateTime(endOfDay)
		}
	}

	return rr
}
