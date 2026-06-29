package reminders

import (
	"testing"
	"time"

	"github.com/nhymxu/kith-pms/internal/people"
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

func TestComputeBirthdayDueDate(t *testing.T) {
	tests := []struct {
		name       string
		dob        string // YYYY-MM-DD or --MM-DD
		daysBefore int
		today      time.Time
		want       time.Time
	}{
		{
			name:       "future this year",
			dob:        "1990-12-25",
			daysBefore: 0,
			today:      date(2026, 6, 29),
			want:       date(2026, 12, 25),
		},
		{
			name:       "already passed this year, roll to next",
			dob:        "1990-01-10",
			daysBefore: 0,
			today:      date(2026, 6, 29),
			want:       date(2027, 1, 10),
		},
		{
			name:       "days_before within year",
			dob:        "1990-12-25",
			daysBefore: 7,
			today:      date(2026, 6, 29),
			want:       date(2026, 12, 18),
		},
		{
			name:       "due already passed this year, roll to next with offset",
			dob:        "1990-07-02",
			daysBefore: 7,
			today:      date(2026, 6, 29),
			want:       date(2027, 6, 25),
		},
		{
			name:       "Feb 29 in non-leap year clamps to Feb 28",
			dob:        "--02-29",
			daysBefore: 0,
			today:      date(2026, 6, 29),
			want:       date(2027, 2, 28),
		},
		{
			name:       "yearless date after today",
			dob:        "--03-15",
			daysBefore: 0,
			today:      date(2026, 6, 29),
			want:       date(2027, 3, 15),
		},
		{
			name:       "on the day itself (not past)",
			dob:        "1990-06-29",
			daysBefore: 0,
			today:      date(2026, 6, 29),
			want:       date(2026, 6, 29),
		},
		{
			name:       "days_before creates past due, rolls to next year",
			dob:        "1990-12-25",
			daysBefore: 200,
			today:      date(2026, 6, 29),
			want:       date(2027, 6, 8),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dob, err := parseTestDOB(tt.dob)
			if err != nil {
				t.Fatalf("parse dob %q: %v", tt.dob, err)
			}

			got := ComputeBirthdayDueDate(*dob, tt.daysBefore, tt.today)
			if !got.Equal(tt.want) {
				t.Errorf("ComputeBirthdayDueDate(%q, %d, %v) = %v, want %v",
					tt.dob, tt.daysBefore, tt.today, got, tt.want)
			}
		})
	}
}

func TestNextBirthdayDueAfter(t *testing.T) {
	tests := []struct {
		name       string
		dob        string // YYYY-MM-DD or --MM-DD
		daysBefore int
		afterDue   time.Time
		want       time.Time
	}{
		{
			name:       "advance to next birthday",
			dob:        "1990-12-25",
			daysBefore: 7,
			afterDue:   date(2026, 12, 18),
			want:       date(2027, 12, 18),
		},
		{
			name:       "on exact day, still advances to next",
			dob:        "1990-12-25",
			daysBefore: 0,
			afterDue:   date(2026, 12, 25),
			want:       date(2027, 12, 25),
		},
		{
			name:       "deep future rolls ahead",
			dob:        "1990-06-15",
			daysBefore: 5,
			afterDue:   date(2030, 6, 10),
			want:       date(2031, 6, 10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dob, err := parseTestDOB(tt.dob)
			if err != nil {
				t.Fatalf("parse dob %q: %v", tt.dob, err)
			}

			got := NextBirthdayDueAfter(*dob, tt.daysBefore, tt.afterDue)
			if !got.After(tt.afterDue) {
				t.Errorf("NextBirthdayDueAfter result %v must be strictly after %v", got, tt.afterDue)
			}

			if !got.Equal(tt.want) {
				t.Errorf("NextBirthdayDueAfter(%q, %d, %v) = %v, want %v",
					tt.dob, tt.daysBefore, tt.afterDue, got, tt.want)
			}
		})
	}
}

func parseTestDOB(s string) (*people.DateOnly, error) {
	dob, err := people.ParseDateOnly(s)
	if err != nil {
		return nil, err
	}

	return &dob, nil
}
