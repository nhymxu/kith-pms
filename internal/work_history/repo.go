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
	query := `
		SELECT id, person_id, company, title, start_date, end_date, location, description, position, created_at
		FROM work_history
		WHERE person_id = ?
		ORDER BY position ASC, id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, personID)
	if err != nil {
		return nil, fmt.Errorf("query work history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entries []WorkEntry

	for rows.Next() {
		var (
			e         WorkEntry
			createdAt string
		)

		err := rows.Scan(
			&e.ID,
			&e.PersonID,
			&e.Company,
			&e.Title,
			&e.StartDate,
			&e.EndDate,
			&e.Location,
			&e.Description,
			&e.Position,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan work entry: %w", err)
		}

		e.CreatedAt, _ = parseTimestamp(createdAt)
		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate work history: %w", err)
	}

	return entries, nil
}

func (r *sqlRepo) ReplaceAll(ctx context.Context, tx bun.Tx, personID int64, entries []WorkEntry) error {
	// Delete existing entries for this person.
	_, err := tx.ExecContext(ctx, "DELETE FROM work_history WHERE person_id = ?", personID)
	if err != nil {
		return fmt.Errorf("delete existing work history: %w", err)
	}

	if len(entries) == 0 {
		return nil
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO work_history (person_id, company, title, start_date, end_date, location, description, position)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert work history: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, e := range entries {
		_, err := stmt.ExecContext(
			ctx,
			personID,
			e.Company,
			e.Title,
			e.StartDate,
			e.EndDate,
			e.Location,
			e.Description,
			e.Position,
		)
		if err != nil {
			return fmt.Errorf("insert work entry: %w", err)
		}
	}

	return nil
}

func parseTimestamp(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05.999Z", s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05Z", s)
	}

	return t, err
}
