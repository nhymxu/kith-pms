package relationships

import (
	"time"

	"github.com/uptrace/bun"
)

// RelationshipType describes a named kind of link between two people.
// When InverseTypeID is set, the system auto-creates the inverse junction row.
type RelationshipType struct {
	bun.BaseModel `bun:"table:relationship_type,alias:rt"`

	ID            int64     `bun:",pk,autoincrement" json:"id"`
	Name          string    `bun:"name"              json:"name"`
	ReverseName   string    `bun:"reverse_name"      json:"reverse_name"`
	InverseTypeID *int64    `bun:"inverse_type_id"   json:"inverse_type_id"`
	CreatedAt     time.Time `bun:"created_at"        json:"created_at"`
	UsageCount    int       `bun:"-"                 json:"usage_count"` // computed via COUNT, not in base select
}

type PersonRelationship struct {
	bun.BaseModel `bun:"table:person_relationship,alias:pr"`

	ID                 int64     `bun:",pk,autoincrement"    json:"id"`
	FromPersonID       int64     `bun:"from_person_id"       json:"from_person_id"`
	ToPersonID         int64     `bun:"to_person_id"         json:"to_person_id"`
	RelationshipTypeID int64     `bun:"relationship_type_id" json:"relationship_type_id"`
	Notes              string    `bun:"notes"                json:"notes"`
	CreatedAt          time.Time `bun:"created_at"           json:"created_at"`
}

type RelationshipView struct {
	ID                int64  `json:"id"`
	OtherPersonID     int64  `json:"other_person_id"`
	OtherPersonName   string `json:"other_person_name"`
	OtherPersonAvatar string `json:"other_person_avatar"`
	TypeName          string `json:"type_name"`
	Notes             string `json:"notes"`
}
