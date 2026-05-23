package monica

import (
	"strings"
	"testing"
	"time"
)

func TestMapContactBasicName(t *testing.T) {
	c := Contact{FirstName: "Jane", LastName: "Doe"}

	rec := MapContact(c)
	if rec.Person.Name != "Jane Doe" {
		t.Errorf("expected 'Jane Doe', got %q", rec.Person.Name)
	}
}

func TestMapContactNicknameFallback(t *testing.T) {
	c := Contact{Nickname: "Ace"}

	rec := MapContact(c)
	if rec.Person.Name != "Ace" {
		t.Errorf("expected 'Ace', got %q", rec.Person.Name)
	}
}

func TestMapContactBirthdateWithYear(t *testing.T) {
	c := Contact{
		FirstName: "Bob",
		Information: Information{
			Birthdate:     "1990-06-15",
			IsYearUnknown: false,
		},
	}

	rec := MapContact(c)
	if rec.Person.DateOfBirth == nil {
		t.Fatal("expected DateOfBirth to be set")
	}

	expected := time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)
	if !rec.Person.DateOfBirth.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, rec.Person.DateOfBirth)
	}
	// Should also produce an ImportantDate birthday entry
	found := false

	for _, d := range rec.Dates {
		if d.Kind == "birthday" && d.DateValue == "1990-06-15" {
			found = true
		}
	}

	if !found {
		t.Error("expected ImportantDate birthday with full date")
	}
}

func TestMapContactBirthdateYearUnknown(t *testing.T) {
	c := Contact{
		FirstName: "Bob",
		Information: Information{
			Birthdate:     "0000-05-15",
			IsYearUnknown: true,
		},
	}

	rec := MapContact(c)
	if rec.Person.DateOfBirth != nil {
		t.Error("expected DateOfBirth to be nil when year unknown")
	}

	found := false

	for _, d := range rec.Dates {
		if d.Kind == "birthday" && d.DateValue == "--05-15" {
			found = true
		}
	}

	if !found {
		t.Errorf("expected yearless ImportantDate '--05-15', got %+v", rec.Dates)
	}
}

func TestMapContactWorkNotes(t *testing.T) {
	c := Contact{FirstName: "Alice", Job: "Engineer", Company: "Acme"}

	rec := MapContact(c)
	if !strings.Contains(rec.Person.OtherNotes, "Engineer") {
		t.Errorf("expected OtherNotes to contain job, got %q", rec.Person.OtherNotes)
	}

	if !strings.Contains(rec.Person.OtherNotes, "Acme") {
		t.Errorf("expected OtherNotes to contain company, got %q", rec.Person.OtherNotes)
	}
}

func TestMapContactWorkNotesEmpty(t *testing.T) {
	c := Contact{FirstName: "Alice"}

	rec := MapContact(c)
	if rec.Person.OtherNotes != "" {
		t.Errorf("expected empty OtherNotes, got %q", rec.Person.OtherNotes)
	}
}

func TestMapContactInfoEmail(t *testing.T) {
	c := Contact{
		FirstName:   "Test",
		ContactInfo: []ContactField{{Data: "test@example.com", Type: ContactFieldType{Name: "Email"}}},
	}

	rec := MapContact(c)
	if len(rec.Contacts) != 1 || rec.Contacts[0].Type != "email" {
		t.Errorf("expected email type, got %+v", rec.Contacts)
	}
}

func TestMapContactInfoPhone(t *testing.T) {
	c := Contact{
		FirstName:   "Test",
		ContactInfo: []ContactField{{Data: "+1234567890", Type: ContactFieldType{Name: "Phone"}}},
	}

	rec := MapContact(c)
	if len(rec.Contacts) != 1 || rec.Contacts[0].Type != "phone" {
		t.Errorf("expected phone type, got %+v", rec.Contacts)
	}
}

func TestMapContactInfoSocial(t *testing.T) {
	c := Contact{
		FirstName: "Test",
		ContactInfo: []ContactField{
			{Data: "@testuser", Type: ContactFieldType{Name: "Twitter"}},
			{Data: "testuser", Type: ContactFieldType{Name: "LinkedIn"}},
		},
	}

	rec := MapContact(c)
	for _, ci := range rec.Contacts {
		if ci.Type != "social" {
			t.Errorf("expected social type for %s, got %q", ci.Label, ci.Type)
		}
	}
}

func TestMapContactInfoEmpty(t *testing.T) {
	c := Contact{
		FirstName:   "Test",
		ContactInfo: []ContactField{{Data: "", Type: ContactFieldType{Name: "Email"}}},
	}

	rec := MapContact(c)
	if len(rec.Contacts) != 0 {
		t.Error("expected empty contacts when data is empty")
	}
}

func TestMapContactNotesTruncation(t *testing.T) {
	longBody := strings.Repeat("a", 100)
	c := Contact{
		FirstName: "Test",
		Notes:     []Note{{Body: longBody, CreatedAt: "2024-01-01T00:00:00Z"}},
	}

	rec := MapContact(c)
	if len(rec.Activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(rec.Activities))
	}

	if len([]rune(rec.Activities[0].Title)) != 60 {
		t.Errorf("expected title truncated to 60 runes, got %d", len([]rune(rec.Activities[0].Title)))
	}
}

func TestMapContactReminderInvalidDateSkipped(t *testing.T) {
	c := Contact{
		FirstName: "Test",
		Reminders: []MReminder{
			{Title: "Valid", InitialDate: "2024-06-01"},
			{Title: "Invalid", InitialDate: "not-a-date"},
		},
	}

	rec := MapContact(c)
	if len(rec.Reminders) != 1 {
		t.Errorf("expected 1 reminder (invalid date skipped), got %d", len(rec.Reminders))
	}
}

func TestMapContactReminderEmptySkipped(t *testing.T) {
	c := Contact{
		FirstName: "Test",
		Reminders: []MReminder{
			{Title: "", InitialDate: "2024-06-01"},
			{Title: "OK", InitialDate: ""},
		},
	}

	rec := MapContact(c)
	if len(rec.Reminders) != 0 {
		t.Errorf("expected 0 reminders, got %d", len(rec.Reminders))
	}
}

func TestMapContactInactiveReminderOption(t *testing.T) {
	c := Contact{
		FirstName: "Test",
		Reminders: []MReminder{{Title: "Paused", InitialDate: "2024-06-01", Inactive: true}},
	}

	withoutOption := MapContact(c)
	if len(withoutOption.Reminders) != 0 {
		t.Errorf("expected inactive reminder skipped by default, got %d", len(withoutOption.Reminders))
	}

	withOption := MapContactWithOptions(c, ImportOptions{ImportInactiveReminders: true})
	if len(withOption.Reminders) != 1 {
		t.Fatalf("expected inactive reminder imported, got %d", len(withOption.Reminders))
	}
	if !withOption.Reminders[0].Completed || withOption.Reminders[0].CompletedAt == nil {
		t.Fatalf("expected inactive reminder imported as completed, got %+v", withOption.Reminders[0])
	}
}

func TestMapAccountJournalEntries(t *testing.T) {
	entries := []MAccountJournal{{Title: "Reflection", Content: "Account note", OccurredAtDate: "2024-01-02"}}

	activities := MapAccountJournalEntries(entries)
	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}
	if activities[0].Title != "Reflection" || activities[0].Content != "Account note" || activities[0].OccurredAtDate != "2024-01-02" {
		t.Fatalf("unexpected account journal mapping: %+v", activities[0])
	}
}

func TestMapContactFirstMetDate(t *testing.T) {
	c := Contact{
		FirstName:   "Test",
		Information: Information{FirstMetDate: "2020-03-10"},
	}
	rec := MapContact(c)
	found := false

	for _, d := range rec.Dates {
		if d.Kind == "met" && d.DateValue == "2020-03-10" {
			found = true
		}
	}

	if !found {
		t.Errorf("expected met ImportantDate, got %+v", rec.Dates)
	}
}

func TestParseV4PreservesPromptedData(t *testing.T) {
	json := `{
		"account": {
			"data": {
				"contacts": [{
					"uuid": "contact-1",
					"properties": {"first_name": "Alice"},
					"data": {
						"reminders": {"data": [{"properties": {"title": "Paused", "initial_date": "2024-06-01", "inactive": true}}]},
						"addresses": {"data": []}, "contact_fields": {"data": []}, "notes": {"data": []},
						"calls": {"data": []}, "tasks": {"data": []}, "gifts": {"data": []}, "activities": {"data": []}
					}
				}],
				"relationships": []
			},
			"properties": {"journal_entries": [{"created_at": "2024-01-02T00:00:00Z", "properties": {"title": "Private", "post": "Account note"}}]}
		}
	}`

	exp, err := Parse(strings.NewReader(json))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(exp.Contacts) != 1 || len(exp.Contacts[0].Reminders) != 1 || !exp.Contacts[0].Reminders[0].Inactive {
		t.Fatalf("expected inactive reminder preserved, got %+v", exp.Contacts)
	}
	if len(exp.AccountJournalEntries) != 1 || exp.AccountJournalEntries[0].Title != "Private" {
		t.Fatalf("expected account journal preserved, got %+v", exp.AccountJournalEntries)
	}
}
