// Package work_history provides domain types and business logic for person work history entries.
package work_history

import (
	"fmt"
	"regexp"
	"time"
)

// WorkEntry represents a single work history record for a person.
type WorkEntry struct {
	ID          int64
	PersonID    int64
	Company     string
	Title       string
	StartDate   string // "YYYY", "YYYY-MM", or "YYYY-MM-DD" — required
	EndDate     string // same formats OR "" (= Present)
	Location    string
	Description string
	Position    int
	CreatedAt   time.Time
}

var workDateRe = regexp.MustCompile(`^\d{4}(-\d{2}(-\d{2})?)?$`)

// ParseWorkDate validates that s matches YYYY, YYYY-MM, or YYYY-MM-DD.
// Returns the normalized string or an error.
func ParseWorkDate(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("date is required")
	}

	if !workDateRe.MatchString(s) {
		return "", fmt.Errorf("date must be YYYY, YYYY-MM, or YYYY-MM-DD format: %q", s)
	}
	// Validate month for YYYY-MM and YYYY-MM-DD.
	if len(s) >= 7 {
		var year, month int

		_, err := fmt.Sscanf(s[:7], "%d-%d", &year, &month)
		if err != nil || month < 1 || month > 12 {
			return "", fmt.Errorf("invalid month in date: %q", s)
		}
	}
	// Validate day range for YYYY-MM-DD (e.g. rejects 2020-02-30).
	if len(s) == 10 {
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return "", fmt.Errorf("invalid date: %q", s)
		}
	}

	return s, nil
}

// DisplayStart returns a human-readable form of StartDate.
// "2020" → "2020", "2020-06" → "Jun 2020", "2020-06-15" → "Jun 15, 2020".
func (w WorkEntry) DisplayStart() string {
	return displayWorkDate(w.StartDate)
}

// DisplayEnd returns a human-readable form of EndDate, or "Present" when empty.
func (w WorkEntry) DisplayEnd() string {
	if w.EndDate == "" {
		return "Present"
	}

	return displayWorkDate(w.EndDate)
}

// displayWorkDate formats a work date string for human display.
func displayWorkDate(s string) string {
	switch len(s) {
	case 4: // YYYY
		return s
	case 7: // YYYY-MM
		t, err := time.Parse("2006-01", s)
		if err != nil {
			return s
		}

		return t.Format("Jan 2006")
	case 10: // YYYY-MM-DD
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return s
		}

		return t.Format("Jan 2, 2006")
	default:
		return s
	}
}
