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

// sqlUserRepo implements UserRepo using bun query builder.
type sqlUserRepo struct{ db *bun.DB }

func NewUserRepo(db *bun.DB) UserRepo { return &sqlUserRepo{db: db} }

func (r *sqlUserRepo) GetUser(ctx context.Context) (*User, error) {
	var u User

	err := r.db.NewSelect().Model(&u).Limit(1).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("auth: get user: %w", err)
	}

	return &u, nil
}

func (r *sqlUserRepo) UpsertUser(ctx context.Context, hash string) error {
	u := &User{
		ID:           1,
		PasswordHash: hash,
		UpdatedAt:    time.Now().UTC(),
	}

	_, err := r.db.NewInsert().
		Model(u).
		On("CONFLICT (id) DO UPDATE SET password_hash = EXCLUDED.password_hash, updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("auth: upsert user: %w", err)
	}

	return nil
}

// sqlSessionRepo implements SessionRepo using bun query builder.
type sqlSessionRepo struct{ db *bun.DB }

func NewSessionRepo(db *bun.DB) SessionRepo { return &sqlSessionRepo{db: db} }

func (r *sqlSessionRepo) CreateSession(ctx context.Context, s Session) error {
	_, err := r.db.NewInsert().Model(&s).Exec(ctx)
	if err != nil {
		return fmt.Errorf("auth: create session: %w", err)
	}

	return nil
}

func (r *sqlSessionRepo) GetSession(ctx context.Context, id string) (*Session, error) {
	var s Session

	err := r.db.NewSelect().Model(&s).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("auth: get session: %w", err)
	}

	return &s, nil
}

func (r *sqlSessionRepo) TouchSession(ctx context.Context, id string, expiresAt time.Time) error {
	s := &Session{
		ID:         id,
		ExpiresAt:  expiresAt.UTC(),
		LastSeenAt: time.Now().UTC(),
	}

	_, err := r.db.NewUpdate().
		Model(s).
		Column("expires_at", "last_seen_at").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("auth: touch session: %w", err)
	}

	return nil
}

func (r *sqlSessionRepo) DeleteSession(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*Session)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("auth: delete session: %w", err)
	}

	return nil
}

func (r *sqlSessionRepo) DeleteAllSessions(ctx context.Context, userID int64) error {
	_, err := r.db.NewDelete().Model((*Session)(nil)).Where("user_id = ?", userID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("auth: delete all sessions: %w", err)
	}

	return nil
}

func (r *sqlSessionRepo) DeleteExpiredSessions(ctx context.Context) error {
	_, err := r.db.NewDelete().Model((*Session)(nil)).Where("expires_at < ?", time.Now().UTC()).Exec(ctx)
	if err != nil {
		return fmt.Errorf("auth: delete expired sessions: %w", err)
	}

	return nil
}

func (r *sqlSessionRepo) CountActiveSessions(ctx context.Context) (int64, error) {
	n, err := r.db.NewSelect().Model((*Session)(nil)).Where("expires_at > ?", time.Now().UTC()).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("auth: count active sessions: %w", err)
	}

	return int64(n), nil
}
