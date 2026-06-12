package monica

import (
	"encoding/json"
	"strings"
	"testing"
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

	if rec.Person.DateOfBirth.String() != "1990-06-15" {
		t.Errorf("expected 1990-06-15, got %v", rec.Person.DateOfBirth)
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

	if activities[0].Title != "Reflection" || activities[0].Content != "Account note" ||
		activities[0].OccurredAtDate != "2024-01-02" {
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

func TestParseV4AvatarResolution(t *testing.T) {
	// Array-of-groups format: photo source with matching UUID should populate AvatarDataURL.
	jsonStr := `{
		"account": {
			"data": [
				{"type": "contact", "count": 2, "values": [
					{
						"uuid": "contact-1",
						"properties": {
							"first_name": "Alice",
							"avatar": {"avatar_source": "photo", "avatar_photo": "photo-uuid-1"}
						},
						"data": []
					},
					{
						"uuid": "contact-2",
						"properties": {
							"first_name": "Bob",
							"avatar": {"avatar_source": "gravatar", "avatar_photo": ""}
						},
						"data": []
					}
				]},
				{"type": "relationship", "count": 0, "values": []},
				{"type": "photo", "count": 1, "values": [
					{"uuid": "photo-uuid-1", "properties": {"dataUrl": "data:image/jpeg;base64,abc123"}}
				]}
			],
			"properties": {"journal_entries": []}
		}
	}`

	exp, err := Parse(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(exp.Contacts) != 2 {
		t.Fatalf("expected 2 contacts, got %d", len(exp.Contacts))
	}

	if exp.Contacts[0].AvatarDataURL != "data:image/jpeg;base64,abc123" {
		t.Errorf("expected avatar data URL for Alice, got %q", exp.Contacts[0].AvatarDataURL)
	}

	if exp.Contacts[1].AvatarDataURL != "" {
		t.Errorf("expected no avatar data URL for Bob (gravatar), got %q", exp.Contacts[1].AvatarDataURL)
	}

	// AvatarDataURL should carry through to ImportRecord.
	rec := MapContact(exp.Contacts[0])
	if rec.AvatarDataURL != "data:image/jpeg;base64,abc123" {
		t.Errorf("expected AvatarDataURL in ImportRecord, got %q", rec.AvatarDataURL)
	}
}

// ---- conversation mapping -----------------------------------------------

func TestMapConversation_TranscriptFormat(t *testing.T) {
	c := Contact{
		FirstName: "Alice",
		Conversations: []MConversation{
			{
				HappenedAt: "2024-03-10T14:00:00Z",
				Messages: []MMessage{
					{Content: "Hi there", WrittenAt: "2024-03-10T14:00:00Z", WrittenByMe: true},
					{Content: "Hello back", WrittenAt: "2024-03-10T14:05:00Z", WrittenByMe: false},
				},
			},
		},
	}

	acts := mapConversations(c)
	if len(acts) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(acts))
	}

	act := acts[0]
	if act.Title != "CONVERSATION 2024-03-10" {
		t.Errorf("unexpected title: %q", act.Title)
	}

	if act.OccurredAtDate != "2024-03-10" {
		t.Errorf("unexpected date: %q", act.OccurredAtDate)
	}

	if !strings.Contains(act.Content, "MESSAGE [me, 14:00] Hi there") {
		t.Errorf("expected me author line, got:\n%s", act.Content)
	}

	if !strings.Contains(act.Content, "MESSAGE [them, 14:05] Hello back") {
		t.Errorf("expected them author line, got:\n%s", act.Content)
	}
}

func TestMapConversation_EmptyMessagesSkipped(t *testing.T) {
	c := Contact{
		FirstName: "Bob",
		Conversations: []MConversation{
			{HappenedAt: "2024-03-10", Messages: nil},
		},
	}

	acts := mapConversations(c)
	if len(acts) != 0 {
		t.Errorf("expected 0 activities for empty-message conversation, got %d", len(acts))
	}
}

func TestMapConversation_NoWrittenAtTimestamp(t *testing.T) {
	c := Contact{
		FirstName: "Carol",
		Conversations: []MConversation{
			{
				HappenedAt: "2024-05-01",
				Messages:   []MMessage{{Content: "Hey", WrittenAt: "", WrittenByMe: false}},
			},
		},
	}

	acts := mapConversations(c)
	if len(acts) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(acts))
	}

	if !strings.Contains(acts[0].Content, "MESSAGE [them] Hey") {
		t.Errorf("expected no timestamp when WrittenAt empty, got:\n%s", acts[0].Content)
	}
}

func TestMapConversation_ControlCharStripping(t *testing.T) {
	c := Contact{
		FirstName: "Dan",
		Conversations: []MConversation{
			{
				HappenedAt: "2024-01-01",
				Messages:   []MMessage{{Content: "hello\x00world\x01\x02", WrittenByMe: true}},
			},
		},
	}

	acts := mapConversations(c)
	if len(acts) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(acts))
	}

	if strings.ContainsAny(acts[0].Content, "\x00\x01\x02") {
		t.Errorf("expected control chars stripped, got: %q", acts[0].Content)
	}

	if !strings.Contains(acts[0].Content, "helloworld") {
		t.Errorf("expected content preserved after stripping, got: %q", acts[0].Content)
	}
}

func TestMapConversation_MessagesPerConvCap(t *testing.T) {
	msgs := make([]MMessage, maxMsgsPerConv+10)
	for i := range msgs {
		msgs[i] = MMessage{Content: "msg", WrittenByMe: true}
	}

	c := Contact{
		FirstName:     "Eve",
		Conversations: []MConversation{{HappenedAt: "2024-01-01", Messages: msgs}},
	}

	acts := mapConversations(c)
	if len(acts) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(acts))
	}

	lineCount := strings.Count(acts[0].Content, "MESSAGE")
	if lineCount > maxMsgsPerConv {
		t.Errorf("expected at most %d MESSAGE lines, got %d", maxMsgsPerConv, lineCount)
	}
}

func TestMapConversation_BodyLengthCap(t *testing.T) {
	bigContent := strings.Repeat("x", maxConvBodyRunes+500)
	c := Contact{
		FirstName: "Frank",
		Conversations: []MConversation{
			{HappenedAt: "2024-01-01", Messages: []MMessage{{Content: bigContent, WrittenByMe: true}}},
		},
	}

	acts := mapConversations(c)
	if len(acts) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(acts))
	}

	if len([]rune(acts[0].Content)) > maxConvBodyRunes {
		t.Errorf("expected body capped at %d runes, got %d", maxConvBodyRunes, len([]rune(acts[0].Content)))
	}
}

// ---- life event mapping -------------------------------------------------

func TestMapLifeEvents_JournalAndImportantDate(t *testing.T) {
	c := Contact{
		FirstName: "Grace",
		LifeEvents: []MLifeEvent{
			{Name: "Got married", Note: "Beautiful ceremony", HappenedAt: "2020-06-15T00:00:00Z"},
		},
	}

	acts, importantDates := mapLifeEvents(c)
	if len(acts) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(acts))
	}

	if acts[0].Title != "LIFE_EVENT Got married" {
		t.Errorf("unexpected title: %q", acts[0].Title)
	}

	if acts[0].Content != "Beautiful ceremony" {
		t.Errorf("unexpected content: %q", acts[0].Content)
	}

	if acts[0].OccurredAtDate != "2020-06-15" {
		t.Errorf("unexpected date: %q", acts[0].OccurredAtDate)
	}

	if len(importantDates) != 1 {
		t.Fatalf("expected 1 ImportantDate, got %d", len(importantDates))
	}

	d := importantDates[0]
	if d.Kind != "other" || d.Label != "Got married" || d.DateValue != "2020-06-15" || d.Recurring {
		t.Errorf("unexpected ImportantDate: %+v", d)
	}
}

func TestMapLifeEvents_EmptyNameSkipped(t *testing.T) {
	c := Contact{
		FirstName:  "Hank",
		LifeEvents: []MLifeEvent{{Name: "   ", Note: "ignored", HappenedAt: "2020-01-01"}},
	}

	acts, dates := mapLifeEvents(c)
	if len(acts) != 0 || len(dates) != 0 {
		t.Errorf("expected empty-name life event skipped, got %d activities %d dates", len(acts), len(dates))
	}
}

func TestMapLifeEvents_MissingHappenedAtFallback(t *testing.T) {
	c := Contact{
		FirstName:  "Iris",
		LifeEvents: []MLifeEvent{{Name: "New job", HappenedAt: ""}},
	}

	acts, _ := mapLifeEvents(c)
	if len(acts) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(acts))
	}
	// dateFromISO("") returns today — just verify it's non-empty and date-shaped
	if len(acts[0].OccurredAtDate) != 10 {
		t.Errorf("expected 10-char date fallback, got %q", acts[0].OccurredAtDate)
	}
}

func TestMapLifeEvents_ControlCharInNote(t *testing.T) {
	c := Contact{
		FirstName:  "Jake",
		LifeEvents: []MLifeEvent{{Name: "Event", Note: "note\x00with\x01nul"}},
	}

	acts, _ := mapLifeEvents(c)
	if len(acts) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(acts))
	}

	if strings.ContainsAny(acts[0].Content, "\x00\x01") {
		t.Errorf("expected control chars stripped from note, got: %q", acts[0].Content)
	}
}

func TestMapContactWithOptions_LifeEventDatesIncluded(t *testing.T) {
	c := Contact{
		FirstName:   "Kate",
		Information: Information{Birthdate: "1990-05-20"},
		LifeEvents: []MLifeEvent{
			{Name: "Moved abroad", HappenedAt: "2015-09-01"},
		},
	}
	rec := MapContactWithOptions(c, ImportOptions{})
	// Expect both birthday (from Information) and life-event date
	var hasLifeEventDate bool

	for _, d := range rec.Dates {
		if d.Kind == "other" && d.Label == "Moved abroad" {
			hasLifeEventDate = true
		}
	}

	if !hasLifeEventDate {
		t.Errorf("expected life-event ImportantDate in rec.Dates, got: %+v", rec.Dates)
	}
}

func TestParseV4PreservesPromptedData(t *testing.T) {
	jsonStr := `{
		"account": {
			"data": [
				{"type": "contact", "count": 1, "values": [
					{
						"uuid": "contact-1",
						"properties": {"first_name": "Alice"},
						"data": [
							{"type": "reminder", "count": 1, "values": [
								{"properties": {"title": "Paused", "initial_date": "2024-06-01", "inactive": true}}
							]}
						]
					}
				]},
				{"type": "relationship", "count": 0, "values": []}
			],
			"properties": {"journal_entries": [
				{"created_at": "2024-01-02T00:00:00Z", "properties": {"title": "Private", "post": "Account note"}}
			]}
		}
	}`

	exp, err := Parse(strings.NewReader(jsonStr))
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

// ---- group decoder tests ----------------------------------------------------

func TestV4GroupsValues(t *testing.T) {
	groups := v4Groups{
		{Type: "contact", Values: []json.RawMessage{[]byte(`"a"`), []byte(`"b"`)}},
		{Type: "photo", Values: []json.RawMessage{[]byte(`"c"`)}},
	}

	if v := groups.values("contact"); len(v) != 2 {
		t.Errorf("expected 2 values for 'contact', got %d", len(v))
	}

	if v := groups.values("photo"); len(v) != 1 {
		t.Errorf("expected 1 value for 'photo', got %d", len(v))
	}

	if v := groups.values("absent"); v != nil {
		t.Errorf("expected nil for absent type, got %v", v)
	}
}

func TestNormDate(t *testing.T) {
	tests := []struct{ input, want string }{
		{"1997-09-17T00:00:00.000000Z", "1997-09-17"},
		{"2024-06-01T12:34:56Z", "2024-06-01"},
		{"2024-06-01", "2024-06-01"},
		{"0000-05-15", "0000-05-15"},
		{"", ""},
		{"2024", "2024"},
	}
	for _, tt := range tests {
		got := normDate(tt.input)
		if got != tt.want {
			t.Errorf("normDate(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseV4DateNormalization(t *testing.T) {
	// RFC3339 timestamps in birthdate and reminder should be trimmed to YYYY-MM-DD.
	jsonStr := `{
		"account": {
			"data": [
				{"type": "contact", "count": 1, "values": [{
					"uuid": "c1",
					"properties": {
						"first_name": "Test",
						"birthdate": {"date": "1990-03-15T00:00:00.000000Z", "is_year_unknown": false},
						"first_met_date": {"date": "2010-07-20T08:00:00Z"}
					},
					"data": [
						{"type": "reminder", "count": 1, "values": [
							{"properties": {"title": "Birthday call", "initial_date": "2024-06-01T00:00:00Z", "inactive": false}}
						]},
						{"type": "gift", "count": 1, "values": [
							{"properties": {"name": "Book", "date": "2023-12-25T00:00:00.000000Z"}}
						]}
					]
				}]},
				{"type": "relationship", "count": 0, "values": []}
			],
			"properties": {"journal_entries": []}
		}
	}`

	exp, err := Parse(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(exp.Contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(exp.Contacts))
	}

	c := exp.Contacts[0]
	if c.Information.Birthdate != "1990-03-15" {
		t.Errorf("birthdate: got %q, want %q", c.Information.Birthdate, "1990-03-15")
	}

	if c.Information.FirstMetDate != "2010-07-20" {
		t.Errorf("first_met_date: got %q, want %q", c.Information.FirstMetDate, "2010-07-20")
	}

	if len(c.Reminders) != 1 || c.Reminders[0].InitialDate != "2024-06-01" {
		t.Errorf("reminder initial_date: got %+v", c.Reminders)
	}

	if len(c.Gifts) != 1 || c.Gifts[0].Date != "2023-12-25" {
		t.Errorf("gift date: got %+v", c.Gifts)
	}
}

func TestParseV4DocumentGroup(t *testing.T) {
	jsonStr := `{
		"account": {
			"data": [
				{"type": "contact", "count": 1, "values": [{
					"uuid": "c1",
					"properties": {"first_name": "Alice"},
					"data": [
						{"type": "document", "count": 1, "values": [
							{"uuid": "doc-1", "properties": {
								"original_filename": "memory.jpg",
								"filesize": 1234,
								"mime_type": "image/jpeg",
								"dataUrl": "data:image/jpeg;base64,abc"
							}}
						]}
					]
				}]},
				{"type": "document", "count": 1, "values": [
					{"uuid": "acct-doc-1", "properties": {
						"original_filename": "report.pdf",
						"dataUrl": "data:application/pdf;base64,xyz"
					}}
				]}
			],
			"properties": {"journal_entries": []}
		}
	}`

	exp, err := Parse(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(exp.Contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(exp.Contacts))
	}

	docs := exp.Contacts[0].Documents
	if len(docs) != 1 {
		t.Fatalf("expected 1 document on contact, got %d", len(docs))
	}

	d := docs[0]
	if d.OriginalFilename != "memory.jpg" || d.MimeType != "image/jpeg" || d.Filesize != 1234 {
		t.Errorf("unexpected document metadata: %+v", d)
	}

	if d.DataURL != "data:image/jpeg;base64,abc" {
		t.Errorf("unexpected document dataURL: %q", d.DataURL)
	}

	// Account-level document with no contact link should be counted but not attached.
	if exp.AccountDocumentCount != 1 {
		t.Errorf("expected AccountDocumentCount == 1, got %d", exp.AccountDocumentCount)
	}

	// Documents carry through to ImportRecord.
	rec := MapContact(exp.Contacts[0])
	if len(rec.Documents) != 1 {
		t.Errorf("expected 1 document in ImportRecord, got %d", len(rec.Documents))
	}
}

func TestParseV4ActivityResolution(t *testing.T) {
	jsonStr := `{
		"account": {
			"data": [
				{"type": "contact", "count": 1, "values": [{
					"uuid": "c1",
					"properties": {"first_name": "Alice"},
					"data": [
						{"type": "activity", "count": 2, "values": ["act-1", "act-2"]}
					]
				}]},
				{"type": "activity", "count": 3, "values": [
					{"uuid": "act-1", "properties": {
						"summary": "Coffee chat", "description": "Caught up over coffee",
						"happened_at": "2021-05-10T00:00:00.000000Z"}},
					{"uuid": "act-2", "properties": {
						"summary": "", "description": "Long walk in the park",
						"happened_at": "2021-06-15T00:00:00.000000Z"}},
					{"uuid": "act-3", "properties": {
						"summary": "Not linked", "description": "This activity has no contact link",
						"happened_at": "2021-07-01T00:00:00.000000Z"}}
				]}
			],
			"properties": {"journal_entries": []}
		}
	}`

	exp, err := Parse(strings.NewReader(jsonStr))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(exp.Contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(exp.Contacts))
	}

	acts := exp.Contacts[0].Activities
	if len(acts) != 2 {
		t.Fatalf("expected 2 activities resolved for contact, got %d", len(acts))
	}

	if acts[0].Summary != "Coffee chat" {
		t.Errorf("expected 'Coffee chat', got %q", acts[0].Summary)
	}

	if acts[0].HappenedAt != "2021-05-10" {
		t.Errorf("expected '2021-05-10', got %q", acts[0].HappenedAt)
	}

	// act-2 has empty summary → falls back to truncated description
	if acts[1].Summary != "Long walk in the park" {
		t.Errorf("expected description fallback, got %q", acts[1].Summary)
	}

	// act-3 is not linked to the contact → should not appear
	rec := MapContact(exp.Contacts[0])
	if len(rec.ActivityEntries) != 2 {
		t.Errorf("expected 2 activity entries in ImportRecord, got %d", len(rec.ActivityEntries))
	}
}
