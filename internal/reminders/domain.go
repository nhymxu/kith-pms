package reminders

import "time"

type Reminder struct {
	ID              int64      `json:"id"`
	Title           string     `json:"title"`
	Notes           string     `json:"notes"`
	DueDate         time.Time  `json:"due_date"`
	PersonID        *int64     `json:"person_id"`
	ImportantDateID *int64     `json:"important_date_id"`
	Completed       bool       `json:"completed"`
	CompletedAt     *time.Time `json:"completed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
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
