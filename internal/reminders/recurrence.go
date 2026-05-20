package reminders

import "time"

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
