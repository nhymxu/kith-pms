package audit

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

type Repo struct{}

func NewRepo() *Repo { return &Repo{} }

// Insert writes a single audit entry using the provided executor (db or tx).
func (r *Repo) Insert(ctx context.Context, db bun.IDB, e Entry) error {
	_, err := db.NewInsert().Model(&e).Exec(ctx)
	if err != nil {
		return fmt.Errorf("audit: insert: %w", err)
	}

	return nil
}

// List returns paginated audit entries, optionally filtered by entity type and ID.
func (r *Repo) List(ctx context.Context, db *bun.DB, p ListParams) ([]Entry, error) {
	pageSize := p.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	page := p.Page
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pageSize

	var entries []Entry

	q := db.NewSelect().Model(&entries).OrderExpr("created_at DESC, id DESC").Limit(pageSize).Offset(offset)

	if p.EntityType != "" {
		q = q.Where("entity_type = ?", string(p.EntityType))
		if p.EntityID > 0 {
			q = q.Where("entity_id = ?", p.EntityID)
		}
	}

	if p.FromDate != "" {
		q = q.Where("date(created_at) >= ?", p.FromDate)
	}

	if p.ToDate != "" {
		q = q.Where("date(created_at) <= ?", p.ToDate)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, fmt.Errorf("audit: list: %w", err)
	}

	return entries, nil
}

// Purge deletes audit entries older than `days` days. Returns count deleted.
func (r *Repo) Purge(ctx context.Context, db *bun.DB, days int) (int64, error) {
	res, err := db.NewDelete().
		Model((*Entry)(nil)).
		Where("created_at < datetime('now', ?)", fmt.Sprintf("-%d days", days)).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("audit: purge: %w", err)
	}

	return res.RowsAffected()
}
