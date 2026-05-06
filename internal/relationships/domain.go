package relationships

import "time"

// RelationshipType describes a named kind of link between two people.
// When InverseTypeID is set, the system auto-creates the inverse junction row.
type RelationshipType struct {
	ID            int64
	Name          string
	ReverseName   string // "" => one-way / symmetric, no auto-inverse
	InverseTypeID *int64 // nil => no partner type configured
	CreatedAt     time.Time
	UsageCount    int // populated by ListWithCounts; 0 elsewhere
}

// PersonRelationship is a raw junction row between two people.
type PersonRelationship struct {
	ID                 int64
	FromPersonID       int64
	ToPersonID         int64
	RelationshipTypeID int64
	Notes              string
	CreatedAt          time.Time
}

// RelationshipView is a denormalised read model used for rendering on the people detail page.
type RelationshipView struct {
	ID                int64
	OtherPersonID     int64
	OtherPersonName   string
	OtherPersonAvatar string // path or ""
	TypeName          string
	Notes             string
}
