package labels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type LabelRepo interface {
	Create(ctx context.Context, name, color string) (int64, error)
	Update(ctx context.Context, id int64, name, color string) error
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*Label, error)
	List(ctx context.Context) ([]Label, error)
	ListWithCounts(ctx context.Context) ([]Label, error)
	ListByPersonIDs(ctx context.Context, personIDs []int64) (map[int64][]Label, error)
}

type PersonLabelRepo interface {
	Attach(ctx context.Context, personID, labelID int64) error
	Detach(ctx context.Context, personID, labelID int64) error
	ListByPersonID(ctx context.Context, personID int64) ([]Label, error)
}

// PersonLabel is the join table model for person_label.
type PersonLabel struct {
	bun.BaseModel `bun:"table:person_label"`

	PersonID int64 `bun:"person_id"`
	LabelID  int64 `bun:"label_id"`
}

// ---- sqlLabelRepo -----------------------------------------------------------

type sqlLabelRepo struct{ db *bun.DB }

func NewLabelRepo(db *bun.DB) LabelRepo { return &sqlLabelRepo{db: db} }

func (r *sqlLabelRepo) Create(ctx context.Context, name, color string) (int64, error) {
	l := &Label{
		Name:      name,
		Color:     color,
		CreatedAt: time.Now().UTC(),
	}

	_, err := r.db.NewInsert().Model(l).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("labels: create: %w", err)
	}

	return l.ID, nil
}

func (r *sqlLabelRepo) Update(ctx context.Context, id int64, name, color string) error {
	l := &Label{
		ID:    id,
		Name:  name,
		Color: color,
	}

	_, err := r.db.NewUpdate().
		Model(l).
		Column("name", "color").
		WherePK().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("labels: update: %w", err)
	}

	return nil
}

func (r *sqlLabelRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*Label)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("labels: delete: %w", err)
	}

	return nil
}

func (r *sqlLabelRepo) Get(ctx context.Context, id int64) (*Label, error) {
	var l Label

	err := r.db.NewSelect().Model(&l).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("labels: get: %w", err)
	}

	return &l, nil
}

func (r *sqlLabelRepo) List(ctx context.Context) ([]Label, error) {
	var labels []Label

	if err := r.db.NewSelect().Model(&labels).OrderExpr("name").Scan(ctx); err != nil {
		return nil, fmt.Errorf("labels: list: %w", err)
	}

	return labels, nil
}

func (r *sqlLabelRepo) ListWithCounts(ctx context.Context) ([]Label, error) {
	type labelWithCount struct {
		ID        int64     `bun:"id"`
		Name      string    `bun:"name"`
		Color     string    `bun:"color"`
		CreatedAt time.Time `bun:"created_at"`
		Count     int       `bun:"count"`
	}

	var rows []labelWithCount

	err := r.db.NewSelect().
		TableExpr("label l").
		ColumnExpr("l.id, l.name, l.color, l.created_at, COUNT(pl.person_id) AS count").
		Join("LEFT JOIN person_label pl ON pl.label_id = l.id").
		GroupExpr("l.id").
		OrderExpr("l.name").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("labels: list with counts: %w", err)
	}

	labels := make([]Label, len(rows))
	for i, row := range rows {
		labels[i] = Label{
			ID:        row.ID,
			Name:      row.Name,
			Color:     row.Color,
			CreatedAt: row.CreatedAt,
			Count:     row.Count,
		}
	}

	return labels, nil
}

func (r *sqlLabelRepo) ListByPersonIDs(ctx context.Context, personIDs []int64) (map[int64][]Label, error) {
	if len(personIDs) == 0 {
		return make(map[int64][]Label), nil
	}

	type row struct {
		PersonID  int64     `bun:"person_id"`
		ID        int64     `bun:"id"`
		Name      string    `bun:"name"`
		Color     string    `bun:"color"`
		CreatedAt time.Time `bun:"created_at"`
	}

	var rows []row

	err := r.db.NewSelect().
		TableExpr("person_label pl").
		ColumnExpr("pl.person_id, l.id, l.name, l.color, l.created_at").
		Join("JOIN label l ON l.id = pl.label_id").
		Where("pl.person_id IN (?)", bun.List(personIDs)).
		OrderExpr("pl.person_id, l.name").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("labels: list by person ids: %w", err)
	}

	result := make(map[int64][]Label)
	for _, r := range rows {
		result[r.PersonID] = append(result[r.PersonID], Label{
			ID:        r.ID,
			Name:      r.Name,
			Color:     r.Color,
			CreatedAt: r.CreatedAt,
		})
	}

	return result, nil
}

// ---- sqlPersonLabelRepo -----------------------------------------------------

type sqlPersonLabelRepo struct{ db *bun.DB }

func NewPersonLabelRepo(db *bun.DB) PersonLabelRepo { return &sqlPersonLabelRepo{db: db} }

func (r *sqlPersonLabelRepo) Attach(ctx context.Context, personID, labelID int64) error {
	pl := &PersonLabel{PersonID: personID, LabelID: labelID}

	_, err := r.db.NewInsert().
		Model(pl).
		On("CONFLICT DO NOTHING").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("labels: attach: %w", err)
	}

	return nil
}

func (r *sqlPersonLabelRepo) Detach(ctx context.Context, personID, labelID int64) error {
	_, err := r.db.NewDelete().
		Model((*PersonLabel)(nil)).
		Where("person_id = ? AND label_id = ?", personID, labelID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("labels: detach: %w", err)
	}

	return nil
}

func (r *sqlPersonLabelRepo) ListByPersonID(ctx context.Context, personID int64) ([]Label, error) {
	var labels []Label

	err := r.db.NewSelect().
		TableExpr("label l").
		ColumnExpr("l.id, l.name, l.color, l.created_at").
		Join("JOIN person_label pl ON pl.label_id = l.id").
		Where("pl.person_id = ?", personID).
		OrderExpr("l.name").
		Scan(ctx, &labels)
	if err != nil {
		return nil, fmt.Errorf("labels: list by person: %w", err)
	}

	return labels, nil
}
