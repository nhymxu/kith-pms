package people

import (
	"time"

	"github.com/uptrace/bun"
)

// Gender enum keys stored in DB; full labels are defined here for reference.
// "male" = Male, "female" = Female, "rather_not_say" = Rather not say, "" = Unselected
type Gender = string

const (
	GenderMale         Gender = "male"
	GenderFemale       Gender = "female"
	GenderRatherNotSay Gender = "rather_not_say"
)

// Person represents a contact in the personal relationship manager.
type Person struct {
	bun.BaseModel `bun:"table:person,alias:p"`

	ID               int64      `bun:",pk,autoincrement"    json:"id"`
	Prefix           string     `bun:"prefix"               json:"prefix"`
	Name             string     `bun:"name"                 json:"name"`
	Nickname         string     `bun:"nickname"             json:"nickname"`
	Gender           string     `bun:"gender"               json:"gender"`
	DateOfBirth      *time.Time `bun:"date_of_birth"        json:"date_of_birth"`
	RelationshipType string     `bun:"relationship_type"    json:"relationship_type"`
	OtherNotes       string     `bun:"other_notes"          json:"other_notes"`
	AvatarPath       string     `bun:"avatar_path"          json:"avatar_path"`
	AvatarMimeType   string     `bun:"avatar_mime_type"     json:"avatar_mime_type"`
	AvatarSize       int64      `bun:"avatar_size"          json:"avatar_size"`
	AvatarUploadedAt *time.Time `bun:"avatar_uploaded_at"   json:"avatar_uploaded_at"`
	LastContactAt    *time.Time `bun:"last_contact_at"      json:"last_contact_at"`
	IsSelf           bool       `bun:"is_self"              json:"is_self"`
	CreatedAt        time.Time  `bun:"created_at"           json:"created_at"`
	UpdatedAt        time.Time  `bun:"updated_at"           json:"updated_at"`

	// Computed/relation fields — populated separately, not stored in person table
	Contacts  []ContactInfo  `bun:"-" json:"contacts"`
	Locations []Location     `bun:"-" json:"locations"`
	Labels    []Label `bun:"-" json:"labels"`
}

// GetLastContactAt returns the last contact timestamp (for interface compatibility).
func (p *Person) GetLastContactAt() *time.Time { return p.LastContactAt }

type PersonList struct {
	Items    []Person `json:"items"`
	Total    int      `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}

// ContactInfo represents a contact method (phone, email, social, etc.) for a Person.
type ContactInfo struct {
	bun.BaseModel `bun:"table:contact_info,alias:ci"`

	ID       int64  `bun:",pk,autoincrement" json:"id"`
	PersonID int64  `bun:"person_id"         json:"person_id"`
	Type     string `bun:"type"              json:"type"`
	Value    string `bun:"value"             json:"value"`
	Label    string `bun:"label"             json:"label"`
	Position int    `bun:"position"          json:"position"`
}

// Location represents a physical address associated with a Person.
type Location struct {
	bun.BaseModel `bun:"table:location,alias:loc"`

	ID         int64  `bun:",pk,autoincrement" json:"id"`
	PersonID   int64  `bun:"person_id"         json:"person_id"`
	Type       string `bun:"type"              json:"type"`
	Address    string `bun:"address"           json:"address"`
	City       string `bun:"city"              json:"city"`
	Country    string `bun:"country"           json:"country"`
	PostalCode string `bun:"postal_code"       json:"postal_code"`
	Position   int    `bun:"position"          json:"position"`
}
