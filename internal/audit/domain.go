package audit

import (
	"time"

	"github.com/uptrace/bun"
)

type EntityType string
type Action string

const (
	EntityPerson             EntityType = "person"
	EntityJournal            EntityType = "journal"
	EntityLabel              EntityType = "label"
	EntityReminder           EntityType = "reminder"
	EntityDate               EntityType = "date"
	EntityWorkHistory        EntityType = "work_history"
	EntityGift               EntityType = "gift"
	EntityRelationshipType   EntityType = "relationship_type"
	EntityPersonRelationship EntityType = "person_relationship"

	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

type Entry struct {
	bun.BaseModel `bun:"table:audit_log,alias:al"`

	ID         int64      `bun:",pk,autoincrement" json:"id"`
	EntityType EntityType `bun:"entity_type"       json:"entity_type"`
	EntityID   int64      `bun:"entity_id"         json:"entity_id"`
	EntityName string     `bun:"entity_name"       json:"entity_name"`
	Action     Action     `bun:"action"            json:"action"`
	ActorID    *int64     `bun:"actor_id"          json:"actor_id"`
	CreatedAt  time.Time  `bun:"created_at"        json:"created_at"`
}

// ListParams controls filtering and pagination for audit queries.
type ListParams struct {
	EntityType EntityType // optional; empty = all types
	EntityID   int64      // optional; only used when EntityType is set
	Page       int
	PageSize   int
	FromDate   string // optional; YYYY-MM-DD inclusive lower bound
	ToDate     string // optional; YYYY-MM-DD inclusive upper bound
}
