package auth

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:user,alias:u"`

	ID           int64     `bun:",pk,autoincrement" json:"id"`
	PasswordHash string    `bun:"password_hash"     json:"password_hash"`
	CreatedAt    time.Time `bun:"created_at"        json:"created_at"`
	UpdatedAt    time.Time `bun:"updated_at"        json:"updated_at"`
}

// Session represents a logged-in session stored in the database.
// The ID field holds HMAC-SHA256(token, secret) — not the raw token — so a DB
// leak does not expose live session tokens.
type Session struct {
	bun.BaseModel `bun:"table:session,alias:s"`

	ID         string    `bun:",pk"        json:"id"`
	UserID     int64     `bun:"user_id"    json:"user_id"`
	ExpiresAt  time.Time `bun:"expires_at" json:"expires_at"`
	LastSeenAt time.Time `bun:"last_seen_at" json:"last_seen_at"`
	IP         string    `bun:"ip"         json:"ip"`
	UserAgent  string    `bun:"user_agent" json:"user_agent"`
}
