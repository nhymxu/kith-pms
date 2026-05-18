package journal

import "time"

type Activity struct {
	ID             int64            `json:"id"`
	Title          string           `json:"title"`
	OccurredAtDate string           `json:"occurred_at_date"`
	OccurredAtTime string           `json:"occurred_at_time"`
	Content        string           `json:"content"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	People         []ActivityPerson `json:"people"`
}

type ActivityPerson struct {
	PersonID   int64  `json:"person_id"`
	Name       string `json:"name"`
	Nickname   string `json:"nickname"`
	AvatarPath string `json:"avatar_path"`
}

type ActivityList struct {
	Items    []Activity `json:"items"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}
