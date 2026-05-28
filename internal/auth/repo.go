package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type UserRepo interface {
	GetUser(ctx context.Context) (*User, error)
	UpsertUser(ctx context.Context, hash string) error
}

type SessionRepo interface {
	CreateSession(ctx context.Context, s Session) error
	GetSession(ctx context.Context, id string) (*Session, error)
	TouchSession(ctx context.Context, id string, expiresAt time.Time) error
	DeleteSession(ctx context.Context, id string) error
	DeleteAllSessions(ctx context.Context, userID int64) error
	DeleteExpiredSessions(ctx context.Context) error
	CountActiveSessions(ctx context.Context) (int64, error)
}

// sqlUserRepo implements UserRepo using raw *bun.DB queries.
type sqlUserRepo struct{ db *bun.DB }

func NewUserRepo(db *bun.DB) UserRepo { return &sqlUserRepo{db: db} }

func (r *sqlUserRepo) GetUser(ctx context.Context) (*User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, password_hash, created_at, updated_at FROM user LIMIT 1`,
	)

	return scanUser(row)
}

func (r *sqlUserRepo) UpsertUser(ctx context.Context, hash string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user (id, password_hash, updated_at)
		 VALUES (1, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		     password_hash = excluded.password_hash,
		     updated_at    = excluded.updated_at`,
		hash, now,
	)
	if err != nil {
		return fmt.Errorf("auth: upsert user: %w", err)
	}

	return nil
}

// sqlSessionRepo implements SessionRepo using raw *bun.DB queries.
type sqlSessionRepo struct{ db *bun.DB }

func NewSessionRepo(db *bun.DB) SessionRepo { return &sqlSessionRepo{db: db} }

func (r *sqlSessionRepo) CreateSession(ctx context.Context, s Session) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO session (id, user_id, expires_at, last_seen_at, ip, user_agent)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		s.ID,
		s.UserID,
		s.ExpiresAt.UTC().Format(time.RFC3339Nano),
		s.LastSeenAt.UTC().Format(time.RFC3339Nano),
		s.IP,
		s.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("auth: create session: %w", err)
	}

	return nil
}

func (r *sqlSessionRepo) GetSession(ctx context.Context, id string) (*Session, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, expires_at, last_seen_at, ip, user_agent
		 FROM session WHERE id = ?`,
		id,
	)

	return scanSession(row)
}

func (r *sqlSessionRepo) TouchSession(ctx context.Context, id string, expiresAt time.Time) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	_, err := r.db.ExecContext(ctx,
		`UPDATE session SET expires_at = ?, last_seen_at = ? WHERE id = ?`,
		expiresAt.UTC().Format(time.RFC3339Nano), now, id,
	)
	if err != nil {
		return fmt.Errorf("auth: touch session: %w", err)
	}

	return nil
}

func (r *sqlSessionRepo) DeleteSession(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM session WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("auth: delete session: %w", err)
	}

	return nil
}

func (r *sqlSessionRepo) DeleteAllSessions(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM session WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("auth: delete all sessions: %w", err)
	}

	return nil
}

func (r *sqlSessionRepo) DeleteExpiredSessions(ctx context.Context) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	_, err := r.db.ExecContext(ctx, `DELETE FROM session WHERE expires_at < ?`, now)
	if err != nil {
		return fmt.Errorf("auth: delete expired sessions: %w", err)
	}

	return nil
}

func (r *sqlSessionRepo) CountActiveSessions(ctx context.Context) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	var n int64

	err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM session WHERE expires_at > ?`, now,
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("auth: count active sessions: %w", err)
	}

	return n, nil
}

// ---- scan helpers -----------------------------------------------------------

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (*User, error) {
	var (
		u                    User
		createdAt, updatedAt string
	)

	err := row.Scan(&u.ID, &u.PasswordHash, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("auth: scan user: %w", err)
	}

	u.CreatedAt, _ = parseTime(createdAt)
	u.UpdatedAt, _ = parseTime(updatedAt)

	return &u, nil
}

func scanSession(row rowScanner) (*Session, error) {
	var (
		s                     Session
		expiresAt, lastSeenAt string
	)

	err := row.Scan(&s.ID, &s.UserID, &expiresAt, &lastSeenAt, &s.IP, &s.UserAgent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("auth: scan session: %w", err)
	}

	s.ExpiresAt, _ = parseTime(expiresAt)
	s.LastSeenAt, _ = parseTime(lastSeenAt)

	return &s, nil
}

// parseTime tries RFC3339Nano then RFC3339 to handle both formats.
func parseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}

	return time.Parse(time.RFC3339, s)
}
