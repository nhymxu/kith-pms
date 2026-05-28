package work_history

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type WorkHistoryRepo interface {
	ListByPerson(ctx context.Context, personID int64) ([]WorkEntry, error)
	ReplaceAll(ctx context.Context, tx bun.Tx, personID int64, entries []WorkEntry) error
}

type sqlRepo struct {
	db *bun.DB
}

// NewRepo creates a new SQL-backed WorkHistoryRepo.
func NewRepo(db *bun.DB) WorkHistoryRepo {
	return &sqlRepo{db: db}
}

func (r *sqlRepo) ListByPerson(ctx context.Context, personID int64) ([]WorkEntry, error) {
	var entries []WorkEntry

	err := r.db.NewSelect().
		Model(&entries).
		Where("person_id = ?", personID).
		OrderExpr("position ASC, id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("query work history: %w", err)
	}

	return entries, nil
}

func (r *sqlRepo) ReplaceAll(ctx context.Context, tx bun.Tx, personID int64, entries []WorkEntry) error {
	_, err := tx.NewDelete().Model((*WorkEntry)(nil)).Where("person_id = ?", personID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete existing work history: %w", err)
	}

	if len(entries) == 0 {
		return nil
	}

	now := time.Now().UTC()
	for i := range entries {
		entries[i].PersonID = personID
		entries[i].CreatedAt = now
	}

	_, err = tx.NewInsert().Model(&entries).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert work history: %w", err)
	}

	return nil
}
