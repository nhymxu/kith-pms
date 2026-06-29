package reminders

import (
	"time"

	"github.com/nhymxu/kith-pms/internal/people"
)

// computeNextDue returns the next due date for a recurring reminder.
// lastContact is the most recent contact date for the linked person (zero = not available).
// Returns zero time if the rule cannot produce a next date.
func computeNextDue(r *Reminder, lastContact time.Time) time.Time {
	if r.RecurrenceRule == nil {
		return time.Time{}
	}

	rule := r.RecurrenceRule
	base := r.DueDate

	var next time.Time

	switch rule.Type {
	case RecurrenceDaily:
		next = base.AddDate(0, 0, 1)

	case RecurrenceWeekly:
		next = base.AddDate(0, 0, 7)

	case RecurrenceMonthly:
		next = base.AddDate(0, 1, 0)

	case RecurrenceYearly:
		next = base.AddDate(1, 0, 0)

	case RecurrenceCustom:
		interval := rule.Interval
		if interval < 1 {
			interval = 1
		}

		switch rule.Unit {
		case "weeks":
			next = base.AddDate(0, 0, interval*7)
		case "months":
			next = base.AddDate(0, interval, 0)
		default: // "days"
			next = base.AddDate(0, 0, interval)
		}

	case RecurrenceDayOfWeek:
		if rule.DayOfWeek == nil {
			return time.Time{}
		}

		target := time.Weekday(*rule.DayOfWeek)
		next = base.AddDate(0, 0, 1)

		for next.Weekday() != target {
			next = next.AddDate(0, 0, 1)
		}

	case RecurrenceRelativeContact:
		interval := rule.Interval
		if interval < 1 {
			interval = 1
		}

		ref := lastContact
		if ref.IsZero() {
			if r.CompletedAt != nil {
				ref = *r.CompletedAt
			} else {
				ref = base
			}
		}

		next = ref.AddDate(0, 0, interval)

	default:
		return time.Time{}
	}

	if r.RecurrenceEndDate != nil && !r.RecurrenceEndDate.IsZero() && next.After(*r.RecurrenceEndDate) {
		return time.Time{}
	}

	return next
}

// ComputeBirthdayDueDate returns the reminder due date for the next upcoming birthday occurrence
// at UTC midnight. dob may be yearless (--MM-DD). daysBefore subtracts lead-time from the birthday.
// Feb 29 in a non-leap year is clamped to Feb 28.
func ComputeBirthdayDueDate(dob people.DateOnly, daysBefore int, today time.Time) time.Time {
	today = today.UTC().Truncate(24 * time.Hour)
	t := dob.Time()
	month := t.Month()
	day := t.Day()

	due := birthdayDueInYear(today.Year(), month, day, daysBefore)
	if !due.Before(today) {
		return due
	}

	return birthdayDueInYear(today.Year()+1, month, day, daysBefore)
}

// NextBirthdayDueAfter returns the due date for the first birthday occurrence whose due date is
// strictly after afterDue. Used to advance on completion so the next row always lands in the
// following cycle regardless of when the reminder is completed.
func NextBirthdayDueAfter(dob people.DateOnly, daysBefore int, afterDue time.Time) time.Time {
	afterDue = afterDue.UTC().Truncate(24 * time.Hour)
	t := dob.Time()
	month := t.Month()
	day := t.Day()

	// Try current year of afterDue and next year, return first strictly after afterDue.
	for _, year := range []int{afterDue.Year(), afterDue.Year() + 1} {
		due := birthdayDueInYear(year, month, day, daysBefore)
		if due.After(afterDue) {
			return due
		}
	}

	// Fallback: two years ahead (handles large daysBefore edge cases).
	return birthdayDueInYear(afterDue.Year()+2, month, day, daysBefore)
}

// birthdayDueInYear computes the due date for birthday month/day in a given year with daysBefore offset.
// Feb 29 in a non-leap year is clamped to Feb 28.
func birthdayDueInYear(year int, month time.Month, day int, daysBefore int) time.Time {
	// Clamp Feb 29 to Feb 28 in non-leap years.
	if month == time.February && day == 29 {
		if !isLeapYear(year) {
			day = 28
		}
	}

	birthday := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	return birthday.AddDate(0, 0, -daysBefore)
}

func isLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}
