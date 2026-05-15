package relationships

import "time"

// RelationshipType describes a named kind of link between two people.
// When InverseTypeID is set, the system auto-creates the inverse junction row.
type RelationshipType struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	ReverseName   string    `json:"reverse_name"`
	InverseTypeID *int64    `json:"inverse_type_id"`
	CreatedAt     time.Time `json:"created_at"`
	UsageCount    int       `json:"usage_count"`
}

type PersonRelationship struct {
	ID                 int64     `json:"id"`
	FromPersonID       int64     `json:"from_person_id"`
	ToPersonID         int64     `json:"to_person_id"`
	RelationshipTypeID int64     `json:"relationship_type_id"`
	Notes              string    `json:"notes"`
	CreatedAt          time.Time `json:"created_at"`
}

type RelationshipView struct {
	ID                int64  `json:"id"`
	OtherPersonID     int64  `json:"other_person_id"`
	OtherPersonName   string `json:"other_person_name"`
	OtherPersonAvatar string `json:"other_person_avatar"`
	TypeName          string `json:"type_name"`
	Notes             string `json:"notes"`
}
