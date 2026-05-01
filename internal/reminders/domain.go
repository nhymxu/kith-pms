package reminders

import "time"

type Reminder struct {
	ID              int64
	Title           string
	Notes           string
	DueDate         time.Time
	PersonID        *int64
	ImportantDateID *int64
	Completed       bool
	CompletedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ReminderWithPerson struct {
	Reminder
	PersonName string
}

type ListParams struct {
	Status   string
	PersonID *int64
	PageSize int
	Page     int
}
