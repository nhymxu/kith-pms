package journal

import "time"

// Activity is a journal entry linked to zero or more people.
type Activity struct {
	ID             int64
	Title          string
	OccurredAtDate string // "YYYY-MM-DD"
	OccurredAtTime string // "HH:MM" or ""
	Content        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	People         []ActivityPerson
}

// ActivityPerson is a person linked to an activity.
type ActivityPerson struct {
	PersonID int64
	Name     string
}
