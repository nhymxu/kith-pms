package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// ErrInvalidCredentials is returned by Login when the password does not match.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Service orchestrates authentication operations.
type Service struct {
	Users    UserRepo
	Sessions SessionRepo
	Secret   []byte
	Lifetime time.Duration
}

// Login verifies plainPwd against the stored password hash, issues a new session,
// and returns the opaque session token for the cookie.
// ip and ua are stored for auditing; they are not logged on failure.
func (s *Service) Login(ctx context.Context, plainPwd, ip, ua string) (string, error) {
	user, err := s.Users.GetUser(ctx)
	if err != nil {
		return "", fmt.Errorf("auth: login get user: %w", err)
	}
	if user == nil {
		// No user configured — reject all logins.
		return "", ErrInvalidCredentials
	}

	ok, err := VerifyPassword(user.PasswordHash, plainPwd)
	if err != nil {
		return "", fmt.Errorf("auth: login verify: %w", err)
	}
	if !ok {
		slog.Warn("auth: failed login attempt", "ip", ip)
		return "", ErrInvalidCredentials
	}

	lifetime := s.Lifetime
	if lifetime <= 0 {
		lifetime = sessionLifetime
	}

	token, err := Issue(ctx, user.ID, ip, ua, s.Sessions, s.Secret, lifetime)
	if err != nil {
		return "", fmt.Errorf("auth: login issue session: %w", err)
	}
	return token, nil
}

// Logout revokes the session identified by the cookie token.
func (s *Service) Logout(ctx context.Context, token string) error {
	return Revoke(ctx, token, s.Sessions, s.Secret)
}

// LogoutAll revokes all sessions for the single application user.
func (s *Service) LogoutAll(ctx context.Context) error {
	user, err := s.Users.GetUser(ctx)
	if err != nil {
		return fmt.Errorf("auth: logout all get user: %w", err)
	}
	if user == nil {
		return nil
	}
	return RevokeAll(ctx, user.ID, s.Sessions)
}

// LoadUser resolves a cookie token to the owning *User.
// Returns (nil, nil) when the token is missing, invalid, or expired.
func (s *Service) LoadUser(ctx context.Context, token string) (*User, error) {
	sess, err := Lookup(ctx, token, s.Sessions, s.Secret)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, nil
	}

	user, err := s.Users.GetUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("auth: load user: %w", err)
	}
	return user, nil
}
