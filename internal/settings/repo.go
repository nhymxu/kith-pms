package settings

import (
	"context"
	"database/sql"
	"fmt"
)

type Repo interface {
	GetAll(ctx context.Context) (map[string]string, error)
	Set(ctx context.Context, key, value, updatedAt string) error
}

type sqlRepo struct{ db *sql.DB }

func NewRepo(db *sql.DB) Repo { return &sqlRepo{db: db} }

func (r *sqlRepo) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM user_setting`)
	if err != nil {
		return nil, fmt.Errorf("settings: get all: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string]string)

	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("settings: scan: %w", err)
		}

		result[k] = v
	}

	return result, rows.Err()
}

func (r *sqlRepo) Set(ctx context.Context, key, value, updatedAt string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO user_setting (key, value, updated_at) VALUES (?, ?, ?)`,
		key, value, updatedAt,
	)
	if err != nil {
		return fmt.Errorf("settings: set %s: %w", key, err)
	}

	return nil
}
