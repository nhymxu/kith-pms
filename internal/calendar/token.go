package calendar

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/nhymxu/kith-pms/internal/settings"
)

const tokenSettingKey = "calendar_feed_token"

// TokenStore keeps the single feed token in user_setting. It is stored in
// plaintext because the Settings page must show the full subscription URL again.
type TokenStore struct {
	Repo settings.Repo
	mu   sync.Mutex
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("calendar: generate token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Get returns the current token, creating one on first use.
func (s *TokenStore) Get(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	all, err := s.Repo.GetAll(ctx)
	if err != nil {
		return "", fmt.Errorf("calendar: read token: %w", err)
	}

	if tok := all[tokenSettingKey]; tok != "" {
		return tok, nil
	}

	return s.rotateLocked(ctx)
}

// Regenerate replaces the token, invalidating every existing subscription URL.
func (s *TokenStore) Regenerate(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.rotateLocked(ctx)
}

func (s *TokenStore) rotateLocked(ctx context.Context) (string, error) {
	tok, err := newToken()
	if err != nil {
		return "", err
	}

	if err := s.Repo.Set(ctx, tokenSettingKey, tok, time.Now().UTC()); err != nil {
		return "", fmt.Errorf("calendar: store token: %w", err)
	}

	return tok, nil
}

// Valid reports whether candidate matches the stored token. A missing token never matches.
func (s *TokenStore) Valid(ctx context.Context, candidate string) (bool, error) {
	all, err := s.Repo.GetAll(ctx)
	if err != nil {
		return false, fmt.Errorf("calendar: read token: %w", err)
	}

	stored := all[tokenSettingKey]
	if stored == "" || candidate == "" {
		return false, nil
	}

	return subtle.ConstantTimeCompare([]byte(stored), []byte(candidate)) == 1, nil
}
