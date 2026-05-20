package reminders

import (
	"testing"
	"time"
)

func date(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func ptr[T any](v T) *T { return &v }

func TestComputeNextDue(t *testing.T) {
	base := date(2026, 5, 20) // Wednesday

	tests := []struct {
		name        string
		reminder    Reminder
		lastContact time.Time
		want        time.Time
	}{
		{
			name: "daily",
			reminder: Reminder{
				DueDate:        base,
				RecurrenceRule: &RecurrenceRule{Type: RecurrenceDaily},
			},
			want: date(2026, 5, 21),
		},
		{
			name: "weekly",
			reminder: Reminder{
				DueDate:        base,
				RecurrenceRule: &RecurrenceRule{Type: RecurrenceWeekly},
			},
			want: date(2026, 5, 27),
		},
		{
			name: "monthly",
			reminder: Reminder{
				DueDate:        base,
				RecurrenceRule: &RecurrenceRule{Type: RecurrenceMonthly},
			},
			want: date(2026, 6, 20),
		},
		{
			name: "yearly",
			reminder: Reminder{
				DueDate:        base,
				RecurrenceRule: &RecurrenceRule{Type: RecurrenceYearly},
			},
			want: date(2027, 5, 20),
		},
		{
			name: "custom 2 weeks",
			reminder: Reminder{
				DueDate:        base,
				RecurrenceRule: &RecurrenceRule{Type: RecurrenceCustom, Interval: 2, Unit: "weeks"},
			},
			want: date(2026, 6, 3),
		},
		{
			name: "custom 45 days",
			reminder: Reminder{
				DueDate:        base,
				RecurrenceRule: &RecurrenceRule{Type: RecurrenceCustom, Interval: 45, Unit: "days"},
			},
			want: date(2026, 7, 4),
		},
		{
			name: "day_of_week Monday (base is Wednesday)",
			reminder: Reminder{
				DueDate:        base,
				RecurrenceRule: &RecurrenceRule{Type: RecurrenceDayOfWeek, DayOfWeek: ptr(1)},
			},
			want: date(2026, 5, 25),
		},
		{
			name: "relative_contact with last contact",
			reminder: Reminder{
				DueDate:        base,
				RecurrenceRule: &RecurrenceRule{Type: RecurrenceRelativeContact, Interval: 30},
			},
			lastContact: date(2026, 5, 15),
			want:        date(2026, 6, 14),
		},
		{
			name: "relative_contact no journal falls back to completedAt",
			reminder: Reminder{
				DueDate:        base,
				CompletedAt:    ptr(date(2026, 5, 20)),
				RecurrenceRule: &RecurrenceRule{Type: RecurrenceRelativeContact, Interval: 30},
			},
			want: date(2026, 6, 19),
		},
		{
			name: "end_date passed returns zero",
			reminder: Reminder{
				DueDate:           base,
				RecurrenceRule:    &RecurrenceRule{Type: RecurrenceDaily},
				RecurrenceEndDate: ptr(date(2026, 5, 20)),
			},
			want: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeNextDue(&tt.reminder, tt.lastContact)
			if !got.Equal(tt.want) {
				t.Errorf("computeNextDue = %v, want %v", got, tt.want)
			}
		})
	}
}
