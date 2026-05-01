package auth

import "time"

// User is the single application user. There is exactly one row in the user table.
type User struct {
	ID           int64
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Session represents a logged-in session stored in the database.
// The ID field holds HMAC-SHA256(token, secret) — not the raw token — so a DB
// leak does not expose live session tokens.
type Session struct {
	ID         string
	UserID     int64
	ExpiresAt  time.Time
	LastSeenAt time.Time
	IP         string
	UserAgent  string
}
