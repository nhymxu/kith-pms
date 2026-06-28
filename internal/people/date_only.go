package people

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DateOnly wraps a date and serializes as YYYY-MM-DD or --MM-DD (yearless).
type DateOnly struct {
	t        time.Time
	yearless bool
}

func NewDateOnly(t time.Time) DateOnly {
	return DateOnly{t: t}
}

// ParseDateOnly parses a "YYYY-MM-DD" or "--MM-DD" string into a DateOnly.
func ParseDateOnly(s string) (DateOnly, error) {
	if strings.HasPrefix(s, "--") {
		// Validate by substituting a leap year so Feb 29 is accepted.
		if _, err := time.Parse("2006-01-02", "2024"+s[1:]); err != nil {
			return DateOnly{}, fmt.Errorf("DateOnly: parse yearless %q: %w", s, err)
		}

		var month, day int
		if _, err := fmt.Sscanf(s, "--%02d-%02d", &month, &day); err != nil {
			return DateOnly{}, fmt.Errorf("DateOnly: parse yearless %q: %w", s, err)
		}

		return DateOnly{
			t:        time.Date(0, time.Month(month), day, 0, 0, 0, 0, time.UTC),
			yearless: true,
		}, nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return DateOnly{}, fmt.Errorf("DateOnly: parse %q: %w", s, err)
	}

	return DateOnly{t: t}, nil
}

func (d DateOnly) Time() time.Time { return d.t }

// IsYearless reports whether the date lacks a year component (--MM-DD format).
func (d DateOnly) IsYearless() bool { return d.yearless }

func (d DateOnly) String() string {
	if d.yearless {
		return fmt.Sprintf("--%02d-%02d", d.t.Month(), d.t.Day())
	}

	return d.t.Format("2006-01-02")
}

// Value implements driver.Valuer — stores as YYYY-MM-DD or --MM-DD TEXT in SQLite.
func (d DateOnly) Value() (driver.Value, error) {
	return d.String(), nil
}

// Scan implements sql.Scanner — reads YYYY-MM-DD, --MM-DD, or full datetime from SQLite TEXT.
func (d *DateOnly) Scan(src any) error {
	if src == nil {
		return nil
	}

	var s string

	switch v := src.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("DateOnly: cannot scan type %T", src)
	}

	parsed, err := ParseDateOnly(truncateToDate(s))
	if err != nil {
		return err
	}

	*d = parsed

	return nil
}

// truncateToDate trims a full datetime string down to its date portion.
func truncateToDate(s string) string {
	if strings.HasPrefix(s, "--") {
		return s
	}

	if len(s) > 10 {
		return s[:10]
	}

	return s
}

// MarshalJSON serializes as a "YYYY-MM-DD" or "--MM-DD" JSON string.
func (d DateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON parses a "YYYY-MM-DD" or "--MM-DD" JSON string.
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	parsed, err := ParseDateOnly(s)
	if err != nil {
		return err
	}

	*d = parsed

	return nil
}
