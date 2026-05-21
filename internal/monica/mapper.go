package monica

import (
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nhymxu/kith-pms/internal/dates"
	"github.com/nhymxu/kith-pms/internal/gifts"
	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/people"
	"github.com/nhymxu/kith-pms/internal/reminders"
	"github.com/nhymxu/kith-pms/internal/work_history"
)

// ImportRecord holds all kith-pms domain objects for a single Monica contact.
type ImportRecord struct {
	Person      people.Person
	Contacts    []people.ContactInfo
	Locations   []people.Location
	TagNames    []string // label names to create-or-find and attach
	Activities  []journal.Activity
	Reminders   []reminders.Reminder
	Dates       []dates.ImportantDate
	WorkHistory []work_history.WorkEntry
	Gifts       []gifts.Gift
	// Relationships are resolved after all persons are inserted (UUID→ID mapping needed).
	Relationships []MRelationship
}

// MapContact converts a Monica Contact into an ImportRecord.
func MapContact(c Contact) ImportRecord {
	return ImportRecord{
		Person:        mapPerson(c),
		Contacts:      mapContactInfo(c.ContactInfo),
		Locations:     mapLocations(c.Addresses),
		TagNames:      mapTags(c.Tags),
		Activities:    mapActivities(c),
		Reminders:     mapReminders(c.Reminders, c.Tasks),
		Dates:         mapDates(c.Information),
		WorkHistory:   mapWorkHistory(c),
		Gifts:         mapGifts(c.Gifts),
		Relationships: c.Relationships,
	}
}

func mapPerson(c Contact) people.Person {
	parts := []string{c.FirstName, c.MiddleName, c.LastName}

	var nameParts []string

	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			nameParts = append(nameParts, s)
		}
	}

	name := strings.Join(nameParts, " ")
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

	// Parse birthdate for Person.DateOfBirth (year-having only).
	if c.Information.Birthdate != "" && !c.Information.IsYearUnknown {
		if t, err := time.Parse("2006-01-02", c.Information.Birthdate); err == nil {
			p.DateOfBirth = &t
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

func mapActivities(c Contact) []journal.Activity {
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

	for _, a := range c.Activities {
		title := strings.TrimSpace(a.Summary)
		if title == "" {
			title = truncate(a.Description, 60)
		}

		if title == "" {
			continue
		}

		out = append(out, journal.Activity{
			Title:          title,
			Content:        a.Description,
			OccurredAtDate: a.HappenedAt,
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

// mapReminders maps Monica reminders and incomplete tasks to kith reminders.
func mapReminders(mrs []MReminder, tasks []MTask) []reminders.Reminder {
	out := make([]reminders.Reminder, 0, len(mrs)+len(tasks))

	for _, r := range mrs {
		if r.Title == "" || r.InitialDate == "" {
			continue
		}

		t, err := time.Parse("2006-01-02", r.InitialDate)
		if err != nil {
			continue
		}

		out = append(out, reminders.Reminder{
			Title:   r.Title,
			Notes:   r.Description,
			DueDate: t,
		})
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

func mapDates(info Information) []dates.ImportantDate {
	var out []dates.ImportantDate

	if info.Birthdate != "" {
		d := dates.ImportantDate{Kind: string(dates.KindBirthday), Recurring: true}

		if info.IsYearUnknown {
			// "0000-MM-DD" → "--MM-DD"
			if len(info.Birthdate) == 10 {
				d.DateValue = "--" + info.Birthdate[5:]
			}
		} else {
			d.DateValue = info.Birthdate
		}

		if d.DateValue != "" {
			out = append(out, d)
		}
	}

	if info.FirstMetDate != "" {
		out = append(out, dates.ImportantDate{
			Kind:      string(dates.KindMet),
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
