package audit

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repo struct{}

func NewRepo() *Repo { return &Repo{} }

// Insert writes a single audit entry using the provided executor (db or tx).
func (r *Repo) Insert(ctx context.Context, db sqlExecer, e Entry) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO audit_log (entity_type, entity_id, entity_name, action, actor_id)
		 VALUES (?, ?, ?, ?, ?)`,
		string(e.EntityType), e.EntityID, e.EntityName, string(e.Action), e.ActorID,
	)
	if err != nil {
		return fmt.Errorf("audit: insert: %w", err)
	}

	return nil
}

// List returns paginated audit entries, optionally filtered by entity type and ID.
func (r *Repo) List(ctx context.Context, db *sql.DB, p ListParams) ([]Entry, error) {
	pageSize := p.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	page := p.Page
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pageSize

	query := `SELECT id, entity_type, entity_id, entity_name, action, actor_id, created_at
	          FROM audit_log`
	args := []any{}

	if p.EntityType != "" {
		if p.EntityID > 0 {
			query += ` WHERE entity_type = ? AND entity_id = ?`

			args = append(args, string(p.EntityType), p.EntityID)
		} else {
			query += ` WHERE entity_type = ?`

			args = append(args, string(p.EntityType))
		}
	}

	query += ` ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?`

	args = append(args, pageSize, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("audit: list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entries []Entry

	for rows.Next() {
		var (
			e            Entry
			createdAtStr string
		)
		if err := rows.Scan(&e.ID, &e.EntityType, &e.EntityID, &e.EntityName,
			&e.Action, &e.ActorID, &createdAtStr); err != nil {
			return nil, fmt.Errorf("audit: scan row: %w", err)
		}

		e.CreatedAt, _ = time.Parse("2006-01-02T15:04:05.999Z", createdAtStr)
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// sqlExecer is satisfied by *sql.DB and *sql.Tx.
type sqlExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
