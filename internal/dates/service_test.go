package dates

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestParseFlexible(t *testing.T) {
	tests := []struct {
		input        string
		wantCanon    string
		wantYearless bool
		wantErr      bool
	}{
		{"2024-03-14", "2024-03-14", false, false},
		{"--03-14", "--03-14", true, false},
		{"2024-02-29", "2024-02-29", false, false}, // leap year
		{"2023-02-29", "", false, true},            // non-leap
		{"--02-29", "--02-29", true, false},        // yearless leap day OK
		{"--13-01", "", false, true},               // invalid month
		{"--01-32", "", false, true},               // invalid day
		{"2024/03/14", "", false, true},            // wrong format
		{"03-14", "", false, true},                 // missing year marker
		{"", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			canon, yearless, err := ParseFlexible(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlexible(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if canon != tt.wantCanon {
				t.Errorf("ParseFlexible(%q) canon = %q, want %q", tt.input, canon, tt.wantCanon)
			}

			if yearless != tt.wantYearless {
				t.Errorf("ParseFlexible(%q) yearless = %v, want %v", tt.input, yearless, tt.wantYearless)
			}
		})
	}
}

func TestImportantDate_IsYearless(t *testing.T) {
	tests := []struct {
		dateValue string
		want      bool
	}{
		{"--03-14", true},
		{"2024-03-14", false},
		{"", false},
	}
	for _, tt := range tests {
		d := ImportantDate{DateValue: tt.dateValue}
		if got := d.IsYearless(); got != tt.want {
			t.Errorf("IsYearless(%q) = %v, want %v", tt.dateValue, got, tt.want)
		}
	}
}

func TestImportantDate_MonthDay(t *testing.T) {
	tests := []struct {
		dateValue string
		want      string
	}{
		{"2024-03-14", "03-14"},
		{"--03-14", "03-14"},
		{"", ""},
		{"123", ""},
	}
	for _, tt := range tests {
		d := ImportantDate{DateValue: tt.dateValue}
		if got := d.MonthDay(); got != tt.want {
			t.Errorf("MonthDay(%q) = %q, want %q", tt.dateValue, got, tt.want)
		}
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	// Enable foreign keys (required for CASCADE)
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	// Create person table (minimal)
	_, err = db.Exec(`
		CREATE TABLE person (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			nickname TEXT NOT NULL DEFAULT ''
		)
	`)
	if err != nil {
		t.Fatalf("create person table: %v", err)
	}

	// Create important_date table
	_, err = db.Exec(`
		CREATE TABLE important_date (
			id INTEGER PRIMARY KEY,
			person_id INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
			kind TEXT NOT NULL DEFAULT 'other',
			label TEXT NOT NULL DEFAULT '',
			date_value TEXT NOT NULL,
			recurring INTEGER NOT NULL DEFAULT 1,
			notes TEXT NOT NULL DEFAULT '',
			position INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			month_day TEXT GENERATED ALWAYS AS (substr(date_value, length(date_value) - 4)) VIRTUAL
		);
		CREATE INDEX idx_important_date_person ON important_date(person_id);
		CREATE INDEX idx_important_date_month_day ON important_date(month_day);
	`)
	if err != nil {
		t.Fatalf("create important_date table: %v", err)
	}

	return db
}

func TestService_ReplaceForPerson(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	// Insert test person
	res, err := db.ExecContext(ctx, "INSERT INTO person (name) VALUES (?)", "Alice")
	if err != nil {
		t.Fatalf("insert person: %v", err)
	}

	personID, _ := res.LastInsertId()

	// Replace with 2 dates
	dates := []ImportantDate{
		{Kind: "birthday", Label: "Birthday", DateValue: "1990-05-01", Recurring: true, Position: 0},
		{Kind: "anniversary", Label: "Met", DateValue: "--03-14", Recurring: true, Position: 1},
	}

	err = svc.ReplaceForPerson(ctx, personID, dates)
	if err != nil {
		t.Fatalf("ReplaceForPerson: %v", err)
	}

	// List back
	got, err := svc.ListByPerson(ctx, personID)
	if err != nil {
		t.Fatalf("ListByPerson: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d dates, want 2", len(got))
	}

	if got[0].DateValue != "1990-05-01" {
		t.Errorf("date[0] = %q, want 1990-05-01", got[0].DateValue)
	}

	if got[1].DateValue != "--03-14" {
		t.Errorf("date[1] = %q, want --03-14", got[1].DateValue)
	}

	// Verify month_day virtual column
	var monthDay string

	err = db.QueryRowContext(ctx, "SELECT month_day FROM important_date WHERE id = ?", got[0].ID).Scan(&monthDay)
	if err != nil {
		t.Fatalf("query month_day: %v", err)
	}

	if monthDay != "05-01" {
		t.Errorf("month_day = %q, want 05-01", monthDay)
	}

	// Replace with 1 date (remove one)
	dates = []ImportantDate{
		{Kind: "birthday", Label: "Birthday", DateValue: "1990-05-01", Recurring: true, Position: 0},
	}

	err = svc.ReplaceForPerson(ctx, personID, dates)
	if err != nil {
		t.Fatalf("ReplaceForPerson (2nd): %v", err)
	}

	got, err = svc.ListByPerson(ctx, personID)
	if err != nil {
		t.Fatalf("ListByPerson (2nd): %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("got %d dates, want 1", len(got))
	}
}

func TestService_CascadeDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	// Insert person + date
	res, err := db.ExecContext(ctx, "INSERT INTO person (name) VALUES (?)", "Bob")
	if err != nil {
		t.Fatalf("insert person: %v", err)
	}

	personID, _ := res.LastInsertId()

	dates := []ImportantDate{
		{Kind: "birthday", Label: "Birthday", DateValue: "1985-12-25", Recurring: true, Position: 0},
	}

	err = svc.ReplaceForPerson(ctx, personID, dates)
	if err != nil {
		t.Fatalf("ReplaceForPerson: %v", err)
	}

	// Delete person
	_, err = db.ExecContext(ctx, "DELETE FROM person WHERE id = ?", personID)
	if err != nil {
		t.Fatalf("delete person: %v", err)
	}

	// Verify dates cascade-deleted
	var count int

	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM important_date WHERE person_id = ?", personID).Scan(&count)
	if err != nil {
		t.Fatalf("count dates: %v", err)
	}

	if count != 0 {
		t.Errorf("got %d dates after person delete, want 0", count)
	}
}

func TestService_OnThisDay(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	// Insert test people
	res1, _ := db.ExecContext(ctx, "INSERT INTO person (name, nickname) VALUES (?, ?)", "Alice", "Ali")
	person1ID, _ := res1.LastInsertId()
	res2, _ := db.ExecContext(ctx, "INSERT INTO person (name) VALUES (?)", "Bob")
	person2ID, _ := res2.LastInsertId()

	// Insert dates
	dates1 := []ImportantDate{
		{Kind: "birthday", Label: "Birthday", DateValue: "1990-05-01", Recurring: true, Position: 0},
		{Kind: "anniversary", Label: "Met", DateValue: "--05-01", Recurring: true, Position: 1},
	}

	err := svc.ReplaceForPerson(ctx, person1ID, dates1)
	if err != nil {
		t.Fatalf("ReplaceForPerson person1: %v", err)
	}

	dates2 := []ImportantDate{
		{Kind: "birthday", Label: "Birthday", DateValue: "2026-05-01", Recurring: false, Position: 0},
		{Kind: "other", Label: "Past event", DateValue: "2024-05-01", Recurring: false, Position: 1},
	}

	err = svc.ReplaceForPerson(ctx, person2ID, dates2)
	if err != nil {
		t.Fatalf("ReplaceForPerson person2: %v", err)
	}

	// Test OnThisDay for 2026-05-01
	today := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	items, err := svc.OnThisDay(ctx, today)
	if err != nil {
		t.Fatalf("OnThisDay: %v", err)
	}

	// Should match: Alice's birthday (recurring), Alice's met (yearless recurring),
	// Bob's birthday (non-recurring exact match)
	// Should NOT match: Bob's past event (non-recurring past date)
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}

	// Check Alice's birthday has YearsSince calculated
	found := false

	for _, item := range items {
		if item.Person.Name == "Alice" && item.Date.Kind == "birthday" {
			found = true

			if item.YearsSince != 36 {
				t.Errorf("Alice birthday YearsSince = %d, want 36", item.YearsSince)
			}
		}

		if item.Person.Name == "Alice" && item.Date.Kind == "anniversary" {
			if item.YearsSince != 0 {
				t.Errorf("Alice anniversary (yearless) YearsSince = %d, want 0", item.YearsSince)
			}
		}
	}

	if !found {
		t.Error("Alice's birthday not found in results")
	}
}

func TestService_Upcoming(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	// Insert test person
	res, _ := db.ExecContext(ctx, "INSERT INTO person (name) VALUES (?)", "Charlie")
	personID, _ := res.LastInsertId()

	// Insert dates
	dates := []ImportantDate{
		{Kind: "birthday", Label: "Birthday", DateValue: "--05-10", Recurring: true, Position: 0},
		{Kind: "anniversary", Label: "Anniversary", DateValue: "--05-20", Recurring: true, Position: 1},
		{Kind: "other", Label: "Future", DateValue: "2026-06-15", Recurring: false, Position: 2},
		{Kind: "other", Label: "Past", DateValue: "2026-04-01", Recurring: false, Position: 3},
	}

	err := svc.ReplaceForPerson(ctx, personID, dates)
	if err != nil {
		t.Fatalf("ReplaceForPerson: %v", err)
	}

	// Test Upcoming for 30 days from 2026-05-01
	today := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	items, err := svc.Upcoming(ctx, today, 30)
	if err != nil {
		t.Fatalf("Upcoming: %v", err)
	}

	// Should match: birthday (05-10), anniversary (05-20)
	// Should NOT match: future (06-15 is 45 days away), past (04-01 is in the past)
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}

	// Check order (should be sorted by next occurrence)
	if items[0].Date.Label != "Birthday" {
		t.Errorf("first item label = %q, want Birthday", items[0].Date.Label)
	}

	if items[1].Date.Label != "Anniversary" {
		t.Errorf("second item label = %q, want Anniversary", items[1].Date.Label)
	}
}

func TestNextOccurrence(t *testing.T) {
	today := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		date      ImportantDate
		wantYear  int
		wantMonth time.Month
		wantDay   int
		wantZero  bool
	}{
		{
			name:      "yearless recurring - today",
			date:      ImportantDate{DateValue: "--05-01", Recurring: true},
			wantYear:  2026,
			wantMonth: 5,
			wantDay:   1,
		},
		{
			name:      "yearless recurring - future this year",
			date:      ImportantDate{DateValue: "--05-10", Recurring: true},
			wantYear:  2026,
			wantMonth: 5,
			wantDay:   10,
		},
		{
			name:      "yearless recurring - past this year",
			date:      ImportantDate{DateValue: "--04-01", Recurring: true},
			wantYear:  2027,
			wantMonth: 4,
			wantDay:   1,
		},
		{
			name:      "year-having recurring - today",
			date:      ImportantDate{DateValue: "1990-05-01", Recurring: true},
			wantYear:  2026,
			wantMonth: 5,
			wantDay:   1,
		},
		{
			name:      "year-having recurring - future this year",
			date:      ImportantDate{DateValue: "1990-05-10", Recurring: true},
			wantYear:  2026,
			wantMonth: 5,
			wantDay:   10,
		},
		{
			name:      "year-having recurring - past this year",
			date:      ImportantDate{DateValue: "1990-04-01", Recurring: true},
			wantYear:  2027,
			wantMonth: 4,
			wantDay:   1,
		},
		{
			name:      "non-recurring exact match",
			date:      ImportantDate{DateValue: "2026-05-01", Recurring: false},
			wantYear:  2026,
			wantMonth: 5,
			wantDay:   1,
		},
		{
			name:      "non-recurring future",
			date:      ImportantDate{DateValue: "2026-06-15", Recurring: false},
			wantYear:  2026,
			wantMonth: 6,
			wantDay:   15,
		},
		{
			name:     "non-recurring past",
			date:     ImportantDate{DateValue: "2026-04-01", Recurring: false},
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextOccurrence(tt.date, today)
			if tt.wantZero {
				if !got.IsZero() {
					t.Errorf("nextOccurrence() = %v, want zero time", got)
				}

				return
			}

			if got.Year() != tt.wantYear || got.Month() != tt.wantMonth || got.Day() != tt.wantDay {
				t.Errorf("nextOccurrence() = %v, want %d-%02d-%02d", got, tt.wantYear, tt.wantMonth, tt.wantDay)
			}
		})
	}
}
