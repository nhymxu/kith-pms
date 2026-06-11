package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// LabelRepo handles CRUD for journal_label rows.
type LabelRepo interface {
	Create(ctx context.Context, name, color string) (int64, error)
	Update(ctx context.Context, id int64, name, color string) error
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*Label, error)
	GetByName(ctx context.Context, name string) (*Label, error)
	List(ctx context.Context) ([]Label, error)
	ListWithCounts(ctx context.Context) ([]Label, error)
	// FilterExisting returns only IDs that exist in journal_label.
	FilterExisting(ctx context.Context, ids []int64) ([]int64, error)
}

// LabelAssignmentRepo manages journal_label_assignment join rows.
type LabelAssignmentRepo interface {
	ReplaceAll(ctx context.Context, tx bun.Tx, activityID int64, labelIDs []int64) error
	ListByActivityID(ctx context.Context, activityID int64) ([]Label, error)
	// ListByActivityIDs returns a map[activityID][]Label for batch population.
	ListByActivityIDs(ctx context.Context, activityIDs []int64) (map[int64][]Label, error)
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
		return 0, fmt.Errorf("journal labels: create: %w", err)
	}

	return l.ID, nil
}

func (r *sqlLabelRepo) Update(ctx context.Context, id int64, name, color string) error {
	l := &Label{ID: id, Name: name, Color: color}

	_, err := r.db.NewUpdate().
		Model(l).
		Column("name", "color").
		WherePK().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("journal labels: update: %w", err)
	}

	return nil
}

func (r *sqlLabelRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*Label)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("journal labels: delete: %w", err)
	}

	return nil
}

func (r *sqlLabelRepo) Get(ctx context.Context, id int64) (*Label, error) {
	var l Label

	err := r.db.NewSelect().Model(&l).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("journal labels: get: %w", err)
	}

	return &l, nil
}

func (r *sqlLabelRepo) GetByName(ctx context.Context, name string) (*Label, error) {
	var l Label

	err := r.db.NewSelect().Model(&l).Where("name = ?", name).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("journal labels: get by name: %w", err)
	}

	return &l, nil
}

func (r *sqlLabelRepo) List(ctx context.Context) ([]Label, error) {
	var labels []Label

	if err := r.db.NewSelect().Model(&labels).OrderExpr("name").Scan(ctx); err != nil {
		return nil, fmt.Errorf("journal labels: list: %w", err)
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
		TableExpr("journal_label jl").
		ColumnExpr("jl.id, jl.name, jl.color, jl.created_at, COUNT(jla.activity_id) AS count").
		Join("LEFT JOIN journal_label_assignment jla ON jla.label_id = jl.id").
		GroupExpr("jl.id").
		OrderExpr("jl.name").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("journal labels: list with counts: %w", err)
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

func (r *sqlLabelRepo) FilterExisting(ctx context.Context, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var existing []int64

	err := r.db.NewSelect().
		TableExpr("journal_label").
		ColumnExpr("id").
		Where("id IN (?)", bun.List(ids)).
		Scan(ctx, &existing)
	if err != nil {
		return nil, fmt.Errorf("journal labels: filter existing: %w", err)
	}

	return existing, nil
}

// ---- sqlLabelAssignmentRepo -------------------------------------------------

type journalLabelAssignment struct {
	bun.BaseModel `bun:"table:journal_label_assignment"`
	ActivityID    int64 `bun:"activity_id"`
	LabelID       int64 `bun:"label_id"`
}

type sqlLabelAssignmentRepo struct{ db *bun.DB }

func NewLabelAssignmentRepo(db *bun.DB) LabelAssignmentRepo {
	return &sqlLabelAssignmentRepo{db: db}
}

func (r *sqlLabelAssignmentRepo) ReplaceAll(ctx context.Context, tx bun.Tx, activityID int64, labelIDs []int64) error {
	_, err := tx.NewDelete().Model((*journalLabelAssignment)(nil)).Where("activity_id = ?", activityID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("journal labels: delete assignments: %w", err)
	}

	if len(labelIDs) == 0 {
		return nil
	}

	rows := make([]journalLabelAssignment, len(labelIDs))
	for i, lid := range labelIDs {
		rows[i] = journalLabelAssignment{ActivityID: activityID, LabelID: lid}
	}

	_, err = tx.NewInsert().Model(&rows).Exec(ctx)
	if err != nil {
		return fmt.Errorf("journal labels: insert assignments: %w", err)
	}

	return nil
}

func (r *sqlLabelAssignmentRepo) ListByActivityID(ctx context.Context, activityID int64) ([]Label, error) {
	var labels []Label

	err := r.db.NewSelect().
		TableExpr("journal_label jl").
		ColumnExpr("jl.id, jl.name, jl.color, jl.created_at").
		Join("JOIN journal_label_assignment jla ON jla.label_id = jl.id").
		Where("jla.activity_id = ?", activityID).
		OrderExpr("jl.name").
		Scan(ctx, &labels)
	if err != nil {
		return nil, fmt.Errorf("journal labels: list by activity: %w", err)
	}

	return labels, nil
}

func (r *sqlLabelAssignmentRepo) ListByActivityIDs(
	ctx context.Context,
	activityIDs []int64,
) (map[int64][]Label, error) {
	if len(activityIDs) == 0 {
		return make(map[int64][]Label), nil
	}

	type row struct {
		ActivityID int64     `bun:"activity_id"`
		ID         int64     `bun:"id"`
		Name       string    `bun:"name"`
		Color      string    `bun:"color"`
		CreatedAt  time.Time `bun:"created_at"`
	}

	var rows []row

	err := r.db.NewSelect().
		TableExpr("journal_label_assignment jla").
		ColumnExpr("jla.activity_id, jl.id, jl.name, jl.color, jl.created_at").
		Join("JOIN journal_label jl ON jl.id = jla.label_id").
		Where("jla.activity_id IN (?)", bun.List(activityIDs)).
		OrderExpr("jla.activity_id, jl.name").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("journal labels: list by activity ids: %w", err)
	}

	result := make(map[int64][]Label)
	for _, r := range rows {
		result[r.ActivityID] = append(result[r.ActivityID], Label{
			ID:        r.ID,
			Name:      r.Name,
			Color:     r.Color,
			CreatedAt: r.CreatedAt,
		})
	}

	return result, nil
}
