package audit

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
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

// Metadata holds optional structured context for an audit event.
type Metadata struct {
	DetailAction string   `json:"detail_action,omitempty"`
	Changes      []Change `json:"changes,omitempty"`
}

// Change captures a single field mutation within an audit event.
type Change struct {
	Field string `json:"field"`
	Old   any    `json:"old"`
	New   any    `json:"new"`
}

// Value implements driver.Valuer — serializes to a JSON string for SQLite TEXT storage.
func (m Metadata) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("audit metadata value: %w", err)
	}

	return string(b), nil
}

// Scan implements sql.Scanner — deserializes from a SQLite TEXT column.
func (m *Metadata) Scan(src any) error {
	if src == nil {
		return nil
	}

	var raw []byte

	switch v := src.(type) {
	case string:
		raw = []byte(v)
	case []byte:
		raw = v
	default:
		return fmt.Errorf("audit metadata scan: unexpected type %T", src)
	}

	return json.Unmarshal(raw, m)
}

type Entry struct {
	bun.BaseModel `bun:"table:audit_log,alias:al"`

	ID         int64      `bun:",pk,autoincrement"  json:"id"`
	EntityType EntityType `bun:"entity_type"        json:"entity_type"`
	EntityID   int64      `bun:"entity_id"          json:"entity_id"`
	EntityName string     `bun:"entity_name"        json:"entity_name"`
	Action     Action     `bun:"action"             json:"action"`
	ActorID    *int64     `bun:"actor_id"           json:"actor_id"`
	Metadata   *Metadata  `bun:"metadata,nullzero"  json:"metadata,omitempty"`
	CreatedAt  time.Time  `bun:"created_at"         json:"created_at"`
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
