package note

import (
	"time"

	"github.com/uptrace/bun"
)

type Note struct {
	bun.BaseModel `bun:"table:note,alias:note"`

	ID        int64     `bun:",pk,autoincrement" json:"id"`
	PersonID  int64     `bun:"person_id"         json:"person_id"`
	Title     string    `bun:"title"             json:"title"`
	Content   string    `bun:"content"           json:"content"`
	CreatedAt time.Time `bun:"created_at"        json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"        json:"updated_at"`
}

type List struct {
	Items    []Note `json:"items"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// WithPerson pairs a note with its owner's name, for search results and other
// contexts that display a note outside the scope of a single person's list.
type WithPerson struct {
	Note
	PersonName string `json:"person_name"`
}

// WithPersonCard is a note plus the owner fields needed to render an avatar
// and profile link in the cross-person notes feed.
type WithPersonCard struct {
	Note
	PersonName      string `json:"person_name"`
	PersonNickname  string `json:"person_nickname"`
	PersonHasAvatar bool   `json:"person_has_avatar"`
}

type AllList struct {
	Items    []WithPersonCard `json:"items"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// ListAllParams filters the cross-person feed. Dates are YYYY-MM-DD, inclusive,
// compared against the UTC date of created_at.
type ListAllParams struct {
	Page      int
	PageSize  int
	PersonIDs []int64
	FromDate  string
	ToDate    string
}
