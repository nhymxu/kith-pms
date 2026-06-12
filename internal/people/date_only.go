package people

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// DateOnly wraps time.Time and serializes as YYYY-MM-DD (no time component).
type DateOnly struct {
	t time.Time
}

func NewDateOnly(t time.Time) DateOnly {
	return DateOnly{t: t}
}

func (d DateOnly) Time() time.Time { return d.t }

func (d DateOnly) String() string {
	return d.t.Format("2006-01-02")
}

// Value implements driver.Valuer — stores as YYYY-MM-DD TEXT in SQLite.
func (d DateOnly) Value() (driver.Value, error) {
	return d.t.Format("2006-01-02"), nil
}

// Scan implements sql.Scanner — reads YYYY-MM-DD or full datetime from SQLite TEXT.
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

	if len(s) > 10 {
		s = s[:10]
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("DateOnly: parse %q: %w", s, err)
	}

	d.t = t

	return nil
}

// MarshalJSON serializes as a "YYYY-MM-DD" JSON string.
func (d DateOnly) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.t.Format("2006-01-02"))
}

// UnmarshalJSON parses a "YYYY-MM-DD" JSON string.
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("DateOnly: parse %q: %w", s, err)
	}

	d.t = t

	return nil
}
