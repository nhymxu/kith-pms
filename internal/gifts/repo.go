package gifts

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type Repo struct {
	db *bun.DB
}

func NewRepo(db *bun.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, tx bun.Tx, g *Gift) (int64, error) {
	g.CreatedAt = time.Now().UTC()
	g.UpdatedAt = g.CreatedAt

	_, err := tx.NewInsert().Model(g).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("insert gift: %w", err)
	}

	return g.ID, nil
}

func (r *Repo) GetByID(ctx context.Context, id int64) (*Gift, error) {
	var g Gift

	if err := r.db.NewSelect().Model(&g).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}

	return &g, nil
}

func (r *Repo) GetByIDWithPerson(ctx context.Context, id int64) (*GiftWithPerson, error) {
	var row struct {
		Gift
		PersonName string `bun:"person_name"`
	}

	err := r.db.NewSelect().
		TableExpr("gift g").
		ColumnExpr("g.*, p.name AS person_name").
		Join("JOIN person p ON p.id = g.person_id").
		Where("g.id = ?", id).
		Scan(ctx, &row)
	if err != nil {
		return nil, err
	}

	return &GiftWithPerson{
		Gift:       row.Gift,
		PersonName: row.PersonName,
	}, nil
}

func (r *Repo) List(ctx context.Context, params ListParams) ([]GiftWithPerson, error) {
	var rows []struct {
		Gift
		PersonName string `bun:"person_name"`
	}

	q := r.db.NewSelect().
		TableExpr("gift g").
		ColumnExpr("g.*, p.name AS person_name").
		Join("JOIN person p ON p.id = g.person_id")

	if params.Direction != "" {
		q = q.Where("g.direction = ?", string(params.Direction))
	}

	if params.PersonID != nil {
		q = q.Where("g.person_id = ?", *params.PersonID)
	}

	if params.DebtType != "" {
		q = q.Where("g.debt_type = ?", string(params.DebtType))
	}

	q = q.OrderExpr("g.created_at DESC")

	if params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		if offset < 0 {
			offset = 0
		}

		q = q.Limit(params.PageSize).Offset(offset)
	}

	err := q.Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("query gifts: %w", err)
	}

	results := make([]GiftWithPerson, 0, len(rows))
	for _, row := range rows {
		results = append(results, GiftWithPerson{
			Gift:       row.Gift,
			PersonName: row.PersonName,
		})
	}

	return results, nil
}

func (r *Repo) Count(ctx context.Context, params ListParams) (int, error) {
	q := r.db.NewSelect().TableExpr("gift g").ColumnExpr("COUNT(*)")

	if params.Direction != "" {
		q = q.Where("g.direction = ?", string(params.Direction))
	}

	if params.PersonID != nil {
		q = q.Where("g.person_id = ?", *params.PersonID)
	}

	if params.DebtType != "" {
		q = q.Where("g.debt_type = ?", string(params.DebtType))
	}

	var total int
	if err := q.Scan(ctx, &total); err != nil {
		return 0, fmt.Errorf("count gifts: %w", err)
	}

	return total, nil
}

func (r *Repo) Update(ctx context.Context, tx bun.Tx, g *Gift) error {
	g.UpdatedAt = time.Now().UTC()

	_, err := tx.NewUpdate().Model(g).WherePK().
		Column("title", "direction", "date", "notes", "amount_cents", "currency", "debt_type", "updated_at").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update gift: %w", err)
	}

	return nil
}

func (r *Repo) UpdateImage(ctx context.Context, tx bun.Tx, id int64, imagePath string) error {
	_, err := tx.NewUpdate().Model((*Gift)(nil)).
		Set("image_path = ?", imagePath).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update gift image: %w", err)
	}

	return nil
}

func (r *Repo) Delete(ctx context.Context, tx bun.Tx, id int64) error {
	_, err := tx.NewDelete().Model((*Gift)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete gift: %w", err)
	}

	return nil
}
