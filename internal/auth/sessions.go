package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	tokenBytes      = 32
	sessionLifetime = 30 * 24 * time.Hour
)

// tokenToID computes HMAC-SHA256(tokenBytes, secret) and returns the hex string
// used as session.id in the database. The raw token is never stored.
func tokenToID(tokenRaw []byte, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(tokenRaw)

	return hex.EncodeToString(mac.Sum(nil))
}

// Issue creates a new session for userID, stores it via repo, and returns the
// opaque base64url-encoded token that is placed in the cookie.
// The token itself is never stored — only its HMAC is persisted.
func Issue(
	ctx context.Context,
	userID int64,
	ip, ua string,
	repo SessionRepo,
	secret []byte,
	lifetime time.Duration,
) (string, error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("auth: generate token: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(raw)
	id := tokenToID(raw, secret)

	now := time.Now().UTC()

	s := Session{
		ID:         id,
		UserID:     userID,
		ExpiresAt:  now.Add(lifetime),
		LastSeenAt: now,
		IP:         ip,
		UserAgent:  ua,
	}
	if err := repo.CreateSession(ctx, s); err != nil {
		return "", fmt.Errorf("auth: issue session: %w", err)
	}

	return token, nil
}

// Lookup resolves a cookie token to the stored Session.
// Returns (nil, nil) when the session does not exist or has expired.
func Lookup(ctx context.Context, token string, repo SessionRepo, secret []byte) (*Session, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		// Malformed token — treat as not found.
		return nil, nil
	}

	id := tokenToID(raw, secret)

	s, err := repo.GetSession(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("auth: lookup session: %w", err)
	}

	if s == nil {
		return nil, nil
	}

	// Reject expired sessions (GC may not have run yet).
	if time.Now().UTC().After(s.ExpiresAt) {
		return nil, nil
	}

	return s, nil
}

// Revoke deletes a single session identified by the cookie token.
func Revoke(ctx context.Context, token string, repo SessionRepo, secret []byte) error {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil // already invalid
	}

	id := tokenToID(raw, secret)

	return repo.DeleteSession(ctx, id)
}

// RevokeAll deletes all sessions for userID (logout-everywhere).
func RevokeAll(ctx context.Context, userID int64, repo SessionRepo) error {
	return repo.DeleteAllSessions(ctx, userID)
}
