package people

import (
	"time"

	"github.com/uptrace/bun"
)

// Label represents a tag that can be attached to a person.
type Label struct {
	bun.BaseModel `bun:"table:people_label,alias:pl"`

	ID        int64     `bun:",pk,autoincrement" json:"id"`
	Name      string    `bun:"name"              json:"name"`
	Color     string    `bun:"color"             json:"color"`
	CreatedAt time.Time `bun:"created_at"        json:"created_at"`
	Count     int       `bun:"-"                 json:"count"` // populated via JOIN, not in base select
}
