package gifts

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, tx *sql.Tx, g *Gift) (int64, error) {
	var dateVal *string
	if g.Date != "" {
		dateVal = &g.Date
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO gift (person_id, title, direction, date, notes,
			amount_cents, currency, debt_type, image_path, image_mime_type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		g.PersonID, g.Title, string(g.Direction), dateVal, g.Notes,
		g.AmountCents, g.Currency, string(g.DebtType), g.ImagePath, g.ImageMimeType,
	)
	if err != nil {
		return 0, fmt.Errorf("insert gift: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	return id, nil
}

func (r *Repo) GetByID(ctx context.Context, id int64) (*Gift, error) {
	var (
		g                    Gift
		dateVal              sql.NullString
		createdAt, updatedAt string
	)

	err := r.db.QueryRowContext(ctx, `
		SELECT id, person_id, title, direction, date, notes, amount_cents, currency, debt_type,
		       image_path, image_mime_type, created_at, updated_at
		FROM gift WHERE id = ?`, id).Scan(
		&g.ID, &g.PersonID, &g.Title, &g.Direction, &dateVal, &g.Notes,
		&g.AmountCents, &g.Currency, &g.DebtType,
		&g.ImagePath, &g.ImageMimeType, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if dateVal.Valid {
		g.Date = dateVal.String
	}

	g.CreatedAt, _ = parseTimestamp(createdAt)
	g.UpdatedAt, _ = parseTimestamp(updatedAt)

	return &g, nil
}

func (r *Repo) List(ctx context.Context, params ListParams) ([]GiftWithPerson, error) {
	args := []any{}
	query := `
		SELECT g.id, g.person_id, g.title, g.direction, g.date, g.notes,
		       g.amount_cents, g.currency, g.debt_type, g.image_path, g.image_mime_type,
		       g.created_at, g.updated_at, p.name AS person_name
		FROM gift g
		JOIN person p ON p.id = g.person_id
		WHERE 1=1`

	if params.Direction != "" {
		query += " AND g.direction = ?"

		args = append(args, string(params.Direction))
	}

	if params.PersonID != nil {
		query += " AND g.person_id = ?"

		args = append(args, *params.PersonID)
	}

	if params.DebtType != "" {
		query += " AND g.debt_type = ?"

		args = append(args, string(params.DebtType))
	}

	query += " ORDER BY g.created_at DESC"

	if params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		if offset < 0 {
			offset = 0
		}

		query += " LIMIT ? OFFSET ?"

		args = append(args, params.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query gifts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []GiftWithPerson

	for rows.Next() {
		var (
			gwp                  GiftWithPerson
			dateVal              sql.NullString
			createdAt, updatedAt string
		)

		err := rows.Scan(
			&gwp.ID, &gwp.PersonID, &gwp.Title, &gwp.Direction, &dateVal, &gwp.Notes,
			&gwp.AmountCents, &gwp.Currency, &gwp.DebtType,
			&gwp.ImagePath, &gwp.ImageMimeType, &createdAt, &updatedAt, &gwp.PersonName,
		)
		if err != nil {
			return nil, fmt.Errorf("scan gift: %w", err)
		}

		if dateVal.Valid {
			gwp.Date = dateVal.String
		}

		gwp.CreatedAt, _ = parseTimestamp(createdAt)
		gwp.UpdatedAt, _ = parseTimestamp(updatedAt)
		results = append(results, gwp)
	}

	return results, rows.Err()
}

func (r *Repo) Update(ctx context.Context, tx *sql.Tx, g *Gift) error {
	var dateVal *string
	if g.Date != "" {
		dateVal = &g.Date
	}

	_, err := tx.ExecContext(ctx, `
		UPDATE gift
		SET title=?, direction=?, date=?, notes=?, amount_cents=?, currency=?, debt_type=?,
		    updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE id=?`,
		g.Title, string(g.Direction), dateVal, g.Notes,
		g.AmountCents, g.Currency, string(g.DebtType), g.ID,
	)
	if err != nil {
		return fmt.Errorf("update gift: %w", err)
	}

	return nil
}

func (r *Repo) UpdateImage(ctx context.Context, tx *sql.Tx, id int64, imagePath, imageMimeType string) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE gift SET image_path=?, image_mime_type=?, updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE id=?`, imagePath, imageMimeType, id)
	if err != nil {
		return fmt.Errorf("update gift image: %w", err)
	}

	return nil
}

func (r *Repo) Delete(ctx context.Context, tx *sql.Tx, id int64) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM gift WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete gift: %w", err)
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
