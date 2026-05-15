package labels

import "time"

// Label represents a tag that can be attached to a person.
type Label struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
	Count     int       `json:"count"`
}
