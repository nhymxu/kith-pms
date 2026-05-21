package people

import (
	"time"

	"github.com/nhymxu/kith-pms/internal/labels"
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
	ID               int64          `json:"id"`
	Prefix           string         `json:"prefix"`
	Name             string         `json:"name"`
	Nickname         string         `json:"nickname"`
	Gender           string         `json:"gender"`
	DateOfBirth      *time.Time     `json:"date_of_birth"`
	RelationshipType string         `json:"relationship_type"`
	OtherNotes       string         `json:"other_notes"`
	AvatarPath       string         `json:"avatar_path"`
	AvatarMimeType   string         `json:"avatar_mime_type"`
	AvatarSize       int64          `json:"avatar_size"`
	AvatarUploadedAt *time.Time     `json:"avatar_uploaded_at"`
	LastContactAt    *time.Time     `json:"last_contact_at"`
	IsSelf           bool           `json:"is_self"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	Contacts         []ContactInfo  `json:"contacts"`
	Locations        []Location     `json:"locations"`
	Labels           []labels.Label `json:"labels"`
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
	ID       int64  `json:"id"`
	PersonID int64  `json:"person_id"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Label    string `json:"label"`
	Position int    `json:"position"`
}

// Location represents a physical address associated with a Person.
type Location struct {
	ID         int64  `json:"id"`
	PersonID   int64  `json:"person_id"`
	Type       string `json:"type"`
	Address    string `json:"address"`
	City       string `json:"city"`
	Country    string `json:"country"`
	PostalCode string `json:"postal_code"`
	Position   int    `json:"position"`
}
