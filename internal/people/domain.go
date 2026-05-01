package people

import "time"

// Person represents a contact in the personal relationship manager.
type Person struct {
	ID               int64
	Prefix           string
	Name             string
	Nickname         string
	DateOfBirth      *time.Time // nullable
	RelationshipType string
	OtherNotes       string
	AvatarPath       string
	AvatarMimeType   string
	AvatarSize       int64
	AvatarUploadedAt *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Contacts         []ContactInfo // populated by service.Get
	Locations        []Location    // populated by service.Get
}

// ContactInfo represents a contact method (phone, email, social, etc.) for a Person.
type ContactInfo struct {
	ID       int64
	PersonID int64
	Type     string // phone | email | social | website | other
	Value    string
	Label    string // optional ("work", "home", etc.)
	Position int    // ordering
}

// Location represents a physical address associated with a Person.
type Location struct {
	ID         int64
	PersonID   int64
	Type       string // home | work | other
	Address    string
	City       string
	Country    string
	PostalCode string
	Position   int
}
