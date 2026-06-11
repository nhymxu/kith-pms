package people

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

// LabelAssignment is the join table model for people_label_assignment.
type LabelAssignment struct {
	bun.BaseModel `bun:"table:people_label_assignment"`

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
		return 0, fmt.Errorf("people labels: create: %w", err)
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
		return fmt.Errorf("people labels: update: %w", err)
	}

	return nil
}

func (r *sqlLabelRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*Label)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("people labels: delete: %w", err)
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

		return nil, fmt.Errorf("people labels: get: %w", err)
	}

	return &l, nil
}

func (r *sqlLabelRepo) List(ctx context.Context) ([]Label, error) {
	var labels []Label

	if err := r.db.NewSelect().Model(&labels).OrderExpr("name").Scan(ctx); err != nil {
		return nil, fmt.Errorf("people labels: list: %w", err)
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
		TableExpr("people_label pl").
		ColumnExpr("pl.id, pl.name, pl.color, pl.created_at, COUNT(pla.person_id) AS count").
		Join("LEFT JOIN people_label_assignment pla ON pla.label_id = pl.id").
		GroupExpr("pl.id").
		OrderExpr("pl.name").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("people labels: list with counts: %w", err)
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
		TableExpr("people_label_assignment pla").
		ColumnExpr("pla.person_id, pl.id, pl.name, pl.color, pl.created_at").
		Join("JOIN people_label pl ON pl.id = pla.label_id").
		Where("pla.person_id IN (?)", bun.List(personIDs)).
		OrderExpr("pla.person_id, pl.name").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("people labels: list by person ids: %w", err)
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
	pla := &LabelAssignment{PersonID: personID, LabelID: labelID}

	_, err := r.db.NewInsert().
		Model(pla).
		On("CONFLICT DO NOTHING").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("people labels: attach: %w", err)
	}

	return nil
}

func (r *sqlPersonLabelRepo) Detach(ctx context.Context, personID, labelID int64) error {
	_, err := r.db.NewDelete().
		Model((*LabelAssignment)(nil)).
		Where("person_id = ? AND label_id = ?", personID, labelID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("people labels: detach: %w", err)
	}

	return nil
}

func (r *sqlPersonLabelRepo) ListByPersonID(ctx context.Context, personID int64) ([]Label, error) {
	var labels []Label

	err := r.db.NewSelect().
		TableExpr("people_label pl").
		ColumnExpr("pl.id, pl.name, pl.color, pl.created_at").
		Join("JOIN people_label_assignment pla ON pla.label_id = pl.id").
		Where("pla.person_id = ?", personID).
		OrderExpr("pl.name").
		Scan(ctx, &labels)
	if err != nil {
		return nil, fmt.Errorf("people labels: list by person: %w", err)
	}

	return labels, nil
}
