package notify

import "time"

// FireTime returns when a reminder should notify. A due date at exactly
// 00:00:00 UTC is date-only (birthday, imported): it fires at notifyHour on
// that calendar date in loc. Anything else fires at its exact instant.
func FireTime(due time.Time, loc *time.Location, notifyHour int) time.Time {
	u := due.UTC()
	if u.Hour() != 0 || u.Minute() != 0 || u.Second() != 0 || u.Nanosecond() != 0 {
		return due
	}

	return time.Date(u.Year(), u.Month(), u.Day(), notifyHour, 0, 0, 0, loc)
}
