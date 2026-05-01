package dates

import (
	"fmt"
	"regexp"
	"time"

	"github.com/nhymxu/kith-pms/internal/people"
)

type Kind string

const (
	KindBirthday    Kind = "birthday"
	KindAnniversary Kind = "anniversary"
	KindMet         Kind = "met"
	KindOther       Kind = "other"
)

type ImportantDate struct {
	ID        int64
	PersonID  int64
	Kind      string
	Label     string
	DateValue string
	Recurring bool
	Notes     string
	Position  int
	CreatedAt time.Time
}

// OnThisDayItem represents a date match with person info and years since.
type OnThisDayItem struct {
	Person     people.Person
	Date       ImportantDate
	YearsSince int // 0 if yearless or non-recurring
}

var (
	yearHavingRe = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`)
	yearlessRe   = regexp.MustCompile(`^--(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`)
)

// IsYearless reports whether DateValue lacks a year component.
func (d ImportantDate) IsYearless() bool {
	return len(d.DateValue) > 0 && d.DateValue[0] == '-' && d.DateValue[1] == '-'
}

// MonthDay returns the trailing "MM-DD" of DateValue.
func (d ImportantDate) MonthDay() string {
	if len(d.DateValue) < 5 {
		return ""
	}
	return d.DateValue[len(d.DateValue)-5:]
}

// ParseFlexible validates and normalizes a user-supplied date string.
// Returns (canonical, yearless, error).
// Accepts "YYYY-MM-DD" or "--MM-DD".
func ParseFlexible(s string) (string, bool, error) {
	if yearHavingRe.MatchString(s) {
		// Validate actual date
		_, err := time.Parse("2006-01-02", s)
		if err != nil {
			return "", false, fmt.Errorf("invalid date: %w", err)
		}
		return s, false, nil
	}
	if yearlessRe.MatchString(s) {
		// Validate month/day by attempting parse with a leap year
		testDate := "2024" + s[1:] // 2024 is leap year
		_, err := time.Parse("2006-01-02", testDate)
		if err != nil {
			return "", false, fmt.Errorf("invalid month/day: %w", err)
		}
		return s, true, nil
	}
	return "", false, fmt.Errorf("date must be YYYY-MM-DD or --MM-DD format")
}

// nextOccurrence returns the next occurrence of this date on or after today.
// Returns zero time if the date is non-recurring and in the past, or if invalid.
func nextOccurrence(d ImportantDate, today time.Time) time.Time {
	if d.IsYearless() {
		// Yearless recurring: try this year, then next year
		monthDay := d.MonthDay()
		if len(monthDay) != 5 {
			return time.Time{}
		}
		thisYear := time.Date(today.Year(), 1, 1, 0, 0, 0, 0, today.Location())
		candidate, err := time.Parse("2006-01-02", fmt.Sprintf("%d-%s", today.Year(), monthDay))
		if err != nil {
			return time.Time{}
		}
		candidate = time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 0, 0, 0, 0, today.Location())
		if !candidate.Before(today) {
			return candidate
		}
		// Try next year
		nextYear := thisYear.AddDate(1, 0, 0)
		candidate, err = time.Parse("2006-01-02", fmt.Sprintf("%d-%s", nextYear.Year(), monthDay))
		if err != nil {
			return time.Time{}
		}
		return time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 0, 0, 0, 0, today.Location())
	}

	// Year-having date
	exact, err := time.Parse("2006-01-02", d.DateValue)
	if err != nil {
		return time.Time{}
	}
	exact = time.Date(exact.Year(), exact.Month(), exact.Day(), 0, 0, 0, 0, today.Location())

	if d.Recurring {
		// Recurring: roll forward to this year or next
		monthDay := d.MonthDay()
		if len(monthDay) != 5 {
			return time.Time{}
		}
		candidate, err := time.Parse("2006-01-02", fmt.Sprintf("%d-%s", today.Year(), monthDay))
		if err != nil {
			return time.Time{}
		}
		candidate = time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 0, 0, 0, 0, today.Location())
		if !candidate.Before(today) {
			return candidate
		}
		// Try next year
		candidate, err = time.Parse("2006-01-02", fmt.Sprintf("%d-%s", today.Year()+1, monthDay))
		if err != nil {
			return time.Time{}
		}
		return time.Date(candidate.Year(), candidate.Month(), candidate.Day(), 0, 0, 0, 0, today.Location())
	}

	// Non-recurring: only if exact date is today or future
	if exact.Before(today) {
		return time.Time{}
	}
	return exact
}
