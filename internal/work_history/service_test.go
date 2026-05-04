package work_history

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	// Enable foreign keys (required for CASCADE).
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	// Minimal person table.
	_, err = db.Exec(`
		CREATE TABLE person (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create person table: %v", err)
	}

	// work_history table matching migration 0010.
	_, err = db.Exec(`
		CREATE TABLE work_history (
			id INTEGER PRIMARY KEY,
			person_id INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
			company TEXT NOT NULL DEFAULT '',
			title TEXT NOT NULL DEFAULT '',
			start_date TEXT NOT NULL DEFAULT '',
			end_date TEXT NOT NULL DEFAULT '',
			location TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			position INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		);
		CREATE INDEX idx_work_history_person ON work_history(person_id);
	`)
	if err != nil {
		t.Fatalf("create work_history table: %v", err)
	}

	return db
}

func TestParseWorkDate(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"2020", "2020", false},
		{"2020-06", "2020-06", false},
		{"2020-06-15", "2020-06-15", false},
		{"06/2020", "", true},
		{"", "", true},
		{"2020-13", "", true}, // invalid month
		{"2020-00", "", true}, // month zero
		{"abcd", "", true},
		{"2020-6", "", true}, // month must be two digits
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseWorkDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWorkDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseWorkDate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestService_ReplaceForPerson(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	// Insert test person.
	res, err := db.ExecContext(ctx, "INSERT INTO person (name) VALUES (?)", "Alice")
	if err != nil {
		t.Fatalf("insert person: %v", err)
	}
	personID, _ := res.LastInsertId()

	// Insert 3 entries.
	entries := []WorkEntry{
		{Company: "Acme", Title: "Engineer", StartDate: "2018-01", EndDate: "2020-06", Position: 0},
		{Company: "Globex", Title: "Senior Engineer", StartDate: "2020-07", EndDate: "2022-12", Position: 1},
		{Company: "Initech", Title: "Lead", StartDate: "2023-01", EndDate: "", Position: 2},
	}
	if err := svc.ReplaceForPerson(ctx, personID, entries); err != nil {
		t.Fatalf("ReplaceForPerson (3 entries): %v", err)
	}

	got, err := svc.ListByPerson(ctx, personID)
	if err != nil {
		t.Fatalf("ListByPerson: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d entries, want 3", len(got))
	}

	// Replace with 2 entries — only 2 should remain.
	entries2 := []WorkEntry{
		{Company: "Acme", Title: "Engineer", StartDate: "2018-01", EndDate: "2020-06", Position: 0},
		{Company: "Globex", Title: "Senior Engineer", StartDate: "2020-07", EndDate: "", Position: 1},
	}
	if err := svc.ReplaceForPerson(ctx, personID, entries2); err != nil {
		t.Fatalf("ReplaceForPerson (2 entries): %v", err)
	}

	got, err = svc.ListByPerson(ctx, personID)
	if err != nil {
		t.Fatalf("ListByPerson (after replace): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d entries after replace, want 2", len(got))
	}
	if got[0].Company != "Acme" {
		t.Errorf("entry[0].Company = %q, want Acme", got[0].Company)
	}
	if got[1].EndDate != "" {
		t.Errorf("entry[1].EndDate = %q, want empty (Present)", got[1].EndDate)
	}
}

func TestService_CascadeDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	svc := NewService(db)

	// Insert person and work entries.
	res, err := db.ExecContext(ctx, "INSERT INTO person (name) VALUES (?)", "Bob")
	if err != nil {
		t.Fatalf("insert person: %v", err)
	}
	personID, _ := res.LastInsertId()

	entries := []WorkEntry{
		{Company: "Acme", Title: "Dev", StartDate: "2019", EndDate: "2021", Position: 0},
		{Company: "Globex", Title: "Lead", StartDate: "2021", EndDate: "", Position: 1},
	}
	if err := svc.ReplaceForPerson(ctx, personID, entries); err != nil {
		t.Fatalf("ReplaceForPerson: %v", err)
	}

	// Delete the person — should cascade.
	_, err = db.ExecContext(ctx, "DELETE FROM person WHERE id = ?", personID)
	if err != nil {
		t.Fatalf("delete person: %v", err)
	}

	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM work_history WHERE person_id = ?", personID).Scan(&count)
	if err != nil {
		t.Fatalf("count work history: %v", err)
	}
	if count != 0 {
		t.Errorf("got %d work history rows after person delete, want 0", count)
	}
}

func TestWorkEntry_DisplayEnd(t *testing.T) {
	e := WorkEntry{EndDate: ""}
	if got := e.DisplayEnd(); got != "Present" {
		t.Errorf("DisplayEnd() = %q, want Present", got)
	}

	e2 := WorkEntry{EndDate: "2022-06"}
	if got := e2.DisplayEnd(); got != "Jun 2022" {
		t.Errorf("DisplayEnd() = %q, want Jun 2022", got)
	}
}
