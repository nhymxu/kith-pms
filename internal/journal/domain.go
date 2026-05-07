package journal

import "time"

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

type ActivityPerson struct {
	PersonID int64
	Name     string
}
