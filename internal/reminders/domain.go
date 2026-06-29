package reminders

import (
	"time"

	"github.com/uptrace/bun"
)

type RecurrenceType string

const (
	RecurrenceDaily           RecurrenceType = "daily"
	RecurrenceWeekly          RecurrenceType = "weekly"
	RecurrenceMonthly         RecurrenceType = "monthly"
	RecurrenceYearly          RecurrenceType = "yearly"
	RecurrenceCustom          RecurrenceType = "custom"
	RecurrenceRelativeContact RecurrenceType = "relative_contact"
	RecurrenceDayOfWeek       RecurrenceType = "day_of_week"
	RecurrenceBirthday        RecurrenceType = "birthday"
)

type RecurrenceRule struct {
	Type          RecurrenceType `json:"type"`
	Interval      int            `json:"interval,omitempty"`
	Unit          string         `json:"unit,omitempty"`
	DayOfWeek     *int           `json:"day_of_week,omitempty"`
	DaysBeforeDob *int           `json:"days_before_dob,omitempty"`
}

func (r *Reminder) IsBirthday() bool {
	return r.RecurrenceRule != nil && r.RecurrenceRule.Type == RecurrenceBirthday
}

type Reminder struct {
	bun.BaseModel `bun:"table:reminder,alias:r"`

	ID                int64           `bun:",pk,autoincrement"     json:"id"`
	Title             string          `bun:"title"                 json:"title"`
	Notes             string          `bun:"notes"                 json:"notes"`
	DueDate           time.Time       `bun:"due_date"              json:"due_date"`
	PersonID          *int64          `bun:"person_id"             json:"person_id"`
	ImportantDateID   *int64          `bun:"important_date_id"     json:"important_date_id"`
	Completed         bool            `bun:"completed"             json:"completed"`
	CompletedAt       *time.Time      `bun:"completed_at"          json:"completed_at"`
	RecurrenceRule    *RecurrenceRule `bun:"-"               json:"recurrence_rule,omitempty"` // JSON-marshaled separately
	RecurrenceEndDate *time.Time      `bun:"recurrence_end_date"   json:"recurrence_end_date,omitempty"`
	CreatedAt         time.Time       `bun:"created_at"            json:"created_at"`
	UpdatedAt         time.Time       `bun:"updated_at"            json:"updated_at"`
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
