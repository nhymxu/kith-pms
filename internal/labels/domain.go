package labels

import "time"

// Label represents a tag that can be attached to a person.
type Label struct {
	ID        int64
	Name      string
	Color     string
	CreatedAt time.Time
	Count     int // populated by ListWithCounts
}
