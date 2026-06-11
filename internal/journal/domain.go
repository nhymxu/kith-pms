package journal

import (
	"time"

	"github.com/uptrace/bun"
)

type Activity struct {
	bun.BaseModel `bun:"table:activity,alias:activity"`

	ID             int64     `bun:",pk,autoincrement"  json:"id"`
	Title          string    `bun:"title"              json:"title"`
	OccurredAtDate string    `bun:"occurred_at_date"   json:"occurred_at_date"`
	OccurredAtTime *string   `bun:"occurred_at_time"   json:"occurred_at_time"` // nullable; NULL when no time recorded
	Content        string    `bun:"content"            json:"content"`
	CreatedAt      time.Time `bun:"created_at"         json:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at"         json:"updated_at"`

	// Populated separately via activity_person JOIN — not stored in activity table
	People []ActivityPerson `bun:"-" json:"people"`
	// Populated separately via journal_label_assignment JOIN — not stored in activity table
	Labels []Label `bun:"-" json:"labels"`
}

// ActivityPerson is a JOIN DTO — not a bun model; scanned from activity_person + person tables.
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
