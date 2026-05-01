package labels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// LabelRepo defines persistence operations for Label records.
type LabelRepo interface {
	Create(ctx context.Context, name, color string) (int64, error)
	Update(ctx context.Context, id int64, name, color string) error
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*Label, error)
	List(ctx context.Context) ([]Label, error)
	ListWithCounts(ctx context.Context) ([]Label, error)
}

// PersonLabelRepo defines persistence operations for person↔label associations.
type PersonLabelRepo interface {
	Attach(ctx context.Context, personID, labelID int64) error
	Detach(ctx context.Context, personID, labelID int64) error
	ListByPersonID(ctx context.Context, personID int64) ([]Label, error)
}

// ---- sqlLabelRepo -----------------------------------------------------------

type sqlLabelRepo struct{ db *sql.DB }

// NewLabelRepo returns a LabelRepo backed by db.
func NewLabelRepo(db *sql.DB) LabelRepo { return &sqlLabelRepo{db: db} }

func (r *sqlLabelRepo) Create(ctx context.Context, name, color string) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO label (name, color, created_at) VALUES (?, ?, ?)`,
		name, color, now,
	)
	if err != nil {
		return 0, fmt.Errorf("labels: create: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("labels: create last id: %w", err)
	}
	return id, nil
}

func (r *sqlLabelRepo) Update(ctx context.Context, id int64, name, color string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE label SET name = ?, color = ? WHERE id = ?`,
		name, color, id,
	)
	if err != nil {
		return fmt.Errorf("labels: update: %w", err)
	}
	return nil
}

func (r *sqlLabelRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM label WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("labels: delete: %w", err)
	}
	return nil
}

func (r *sqlLabelRepo) Get(ctx context.Context, id int64) (*Label, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, color, created_at FROM label WHERE id = ?`, id,
	)
	l, err := scanLabel(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (r *sqlLabelRepo) List(ctx context.Context) ([]Label, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, color, created_at FROM label ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("labels: list: %w", err)
	}
	defer rows.Close()
	return collectLabels(rows)
}

func (r *sqlLabelRepo) ListWithCounts(ctx context.Context) ([]Label, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT l.id, l.name, l.color, l.created_at, COUNT(pl.person_id) as count
		 FROM label l
		 LEFT JOIN person_label pl ON pl.label_id = l.id
		 GROUP BY l.id
		 ORDER BY l.name`,
	)
	if err != nil {
		return nil, fmt.Errorf("labels: list with counts: %w", err)
	}
	defer rows.Close()

	var labels []Label
	for rows.Next() {
		var l Label
		var createdAt string
		if err := rows.Scan(&l.ID, &l.Name, &l.Color, &createdAt, &l.Count); err != nil {
			return nil, fmt.Errorf("labels: scan label with count: %w", err)
		}
		l.CreatedAt, _ = parseTime(createdAt)
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

// ---- sqlPersonLabelRepo -----------------------------------------------------

type sqlPersonLabelRepo struct{ db *sql.DB }

// NewPersonLabelRepo returns a PersonLabelRepo backed by db.
func NewPersonLabelRepo(db *sql.DB) PersonLabelRepo { return &sqlPersonLabelRepo{db: db} }

func (r *sqlPersonLabelRepo) Attach(ctx context.Context, personID, labelID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO person_label (person_id, label_id) VALUES (?, ?)`,
		personID, labelID,
	)
	if err != nil {
		return fmt.Errorf("labels: attach: %w", err)
	}
	return nil
}

func (r *sqlPersonLabelRepo) Detach(ctx context.Context, personID, labelID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM person_label WHERE person_id = ? AND label_id = ?`,
		personID, labelID,
	)
	if err != nil {
		return fmt.Errorf("labels: detach: %w", err)
	}
	return nil
}

func (r *sqlPersonLabelRepo) ListByPersonID(ctx context.Context, personID int64) ([]Label, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT l.id, l.name, l.color, l.created_at
		 FROM label l
		 JOIN person_label pl ON pl.label_id = l.id
		 WHERE pl.person_id = ?
		 ORDER BY l.name`,
		personID,
	)
	if err != nil {
		return nil, fmt.Errorf("labels: list by person: %w", err)
	}
	defer rows.Close()
	return collectLabels(rows)
}

// ---- scan helpers -----------------------------------------------------------

type rowScanner interface {
	Scan(dest ...any) error
}

func scanLabel(row rowScanner) (Label, error) {
	var l Label
	var createdAt string
	if err := row.Scan(&l.ID, &l.Name, &l.Color, &createdAt); err != nil {
		return Label{}, fmt.Errorf("labels: scan label: %w", err)
	}
	l.CreatedAt, _ = parseTime(createdAt)
	return l, nil
}

func collectLabels(rows *sql.Rows) ([]Label, error) {
	var labels []Label
	for rows.Next() {
		l, err := scanLabel(rows)
		if err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

func parseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}
