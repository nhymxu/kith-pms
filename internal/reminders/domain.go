package reminders

import "time"

type RecurrenceType string

const (
	RecurrenceDaily           RecurrenceType = "daily"
	RecurrenceWeekly          RecurrenceType = "weekly"
	RecurrenceMonthly         RecurrenceType = "monthly"
	RecurrenceYearly          RecurrenceType = "yearly"
	RecurrenceCustom          RecurrenceType = "custom"
	RecurrenceRelativeContact RecurrenceType = "relative_contact"
	RecurrenceDayOfWeek       RecurrenceType = "day_of_week"
)

type RecurrenceRule struct {
	Type      RecurrenceType `json:"type"`
	Interval  int            `json:"interval,omitempty"`
	Unit      string         `json:"unit,omitempty"`
	DayOfWeek *int           `json:"day_of_week,omitempty"`
}

type Reminder struct {
	ID                int64           `json:"id"`
	Title             string          `json:"title"`
	Notes             string          `json:"notes"`
	DueDate           time.Time       `json:"due_date"`
	PersonID          *int64          `json:"person_id"`
	ImportantDateID   *int64          `json:"important_date_id"`
	Completed         bool            `json:"completed"`
	CompletedAt       *time.Time      `json:"completed_at"`
	RecurrenceRule    *RecurrenceRule `json:"recurrence_rule,omitempty"`
	RecurrenceEndDate *time.Time      `json:"recurrence_end_date,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type ReminderWithPerson struct {
	Reminder
	PersonName string `json:"person_name"`
}

type ListParams struct {
	Status   string
	PersonID *int64
	PageSize int
	Page     int
}
