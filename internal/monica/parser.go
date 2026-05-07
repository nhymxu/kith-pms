package monica

import (
	"encoding/json"
	"io"
)

type Export struct {
	Contacts []Contact `json:"contacts"`
}

type Contact struct {
	ID          string         `json:"id"`
	FirstName   string         `json:"first_name"`
	LastName    string         `json:"last_name"`
	Nickname    string         `json:"nickname"`
	Company     string         `json:"company"`
	Job         string         `json:"job"`
	Information Information    `json:"information"`
	Addresses   []Address      `json:"addresses"`
	ContactInfo []ContactField `json:"contactInformation"`
	Notes       []Note         `json:"notes"`
	Activities  []MActivity    `json:"activities"`
	Reminders   []MReminder    `json:"reminders"`
	Tags        []Tag          `json:"tags"`
}

type Information struct {
	Birthdate     string `json:"birthdate"` // "YYYY-MM-DD" | "0000-MM-DD" | ""
	IsYearUnknown bool   `json:"is_year_unknown"`
	FirstMetDate  string `json:"first_met_date"` // "YYYY-MM-DD" | ""
}

type Address struct {
	Name       string `json:"name"`
	Street     string `json:"street"`
	City       string `json:"city"`
	Province   string `json:"province"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type ContactField struct {
	Data string           `json:"data"`
	Type ContactFieldType `json:"contact_field_type"`
}

type ContactFieldType struct {
	Name string `json:"name"`
}

type Note struct {
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"` // ISO 8601
}

type MActivity struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	HappenedAt  string `json:"happened_at"` // "YYYY-MM-DD"
}

type MReminder struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	InitialDate string `json:"initial_date"` // "YYYY-MM-DD"
}

type Tag struct {
	Name string `json:"name"`
}

// Parse decodes a Monica JSON export from r.
func Parse(r io.Reader) (*Export, error) {
	var exp Export
	if err := json.NewDecoder(r).Decode(&exp); err != nil {
		return nil, err
	}

	return &exp, nil
}
