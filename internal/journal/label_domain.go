package journal

import (
	"time"

	"github.com/uptrace/bun"
)

// Label represents a tag that can be attached to a journal entry.
type Label struct {
	bun.BaseModel `bun:"table:journal_label,alias:jl"`

	ID        int64     `bun:",pk,autoincrement" json:"id"`
	Name      string    `bun:"name"              json:"name"`
	Color     string    `bun:"color"             json:"color"`
	CreatedAt time.Time `bun:"created_at"        json:"created_at"`
	Count     int       `bun:"-"                 json:"count"`
}
