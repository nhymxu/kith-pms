package settings

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// Setting is the DB model for user_setting rows.
type Setting struct {
	bun.BaseModel `bun:"table:user_setting"`

	Key       string    `bun:",pk"`
	Value     string    `bun:"value"`
	UpdatedAt time.Time `bun:"updated_at"`
}

type Repo interface {
	GetAll(ctx context.Context) (map[string]string, error)
	Set(ctx context.Context, key, value string, updatedAt time.Time) error
}

type sqlRepo struct{ db *bun.DB }

func NewRepo(db *bun.DB) Repo { return &sqlRepo{db: db} }

func (r *sqlRepo) GetAll(ctx context.Context) (map[string]string, error) {
	var settings []Setting

	if err := r.db.NewSelect().Model(&settings).Scan(ctx); err != nil {
		return nil, fmt.Errorf("settings: get all: %w", err)
	}

	result := make(map[string]string, len(settings))
	for _, s := range settings {
		result[s.Key] = s.Value
	}

	return result, nil
}

func (r *sqlRepo) Set(ctx context.Context, key, value string, updatedAt time.Time) error {
	s := &Setting{Key: key, Value: value, UpdatedAt: updatedAt}

	_, err := r.db.NewInsert().
		Model(s).
		On("CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("settings: set %s: %w", key, err)
	}

	return nil
}
