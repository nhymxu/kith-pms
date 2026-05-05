package audit

import "time"

type EntityType string
type Action string

const (
	EntityPerson      EntityType = "person"
	EntityJournal     EntityType = "journal"
	EntityLabel       EntityType = "label"
	EntityReminder    EntityType = "reminder"
	EntityDate        EntityType = "date"
	EntityWorkHistory EntityType = "work_history"

	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// Entry is a single audit log record.
type Entry struct {
	ID         int64
	EntityType EntityType
	EntityID   int64
	EntityName string
	Action     Action
	ActorID    *int64
	CreatedAt  time.Time
}

// ListParams controls filtering and pagination for audit queries.
type ListParams struct {
	EntityType EntityType // optional; empty = all types
	EntityID   int64      // optional; only used when EntityType is set
	Page       int
	PageSize   int
}
