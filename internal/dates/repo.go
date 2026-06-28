package dates

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

type ImportantDateRepo interface {
	ListByPerson(ctx context.Context, personID int64) ([]ImportantDate, error)
	ReplaceAll(ctx context.Context, tx bun.Tx, personID int64, dates []ImportantDate) error
	OnThisDay(ctx context.Context, monthDay, todayISO string) ([]OnThisDayItem, error)
	ListAll(ctx context.Context) ([]OnThisDayItem, error)
}

type sqlRepo struct {
	db *bun.DB
}

func NewRepo(db *bun.DB) ImportantDateRepo {
	return &sqlRepo{db: db}
}

func (r *sqlRepo) ListByPerson(ctx context.Context, personID int64) ([]ImportantDate, error) {
	var dates []ImportantDate

	err := r.db.NewSelect().
		Model(&dates).
		Where("person_id = ?", personID).
		OrderExpr("position, id").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("query dates: %w", err)
	}

	return dates, nil
}

func (r *sqlRepo) ReplaceAll(ctx context.Context, tx bun.Tx, personID int64, dates []ImportantDate) error {
	_, err := tx.NewDelete().Model((*ImportantDate)(nil)).Where("person_id = ?", personID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete existing dates: %w", err)
	}

	if len(dates) == 0 {
		return nil
	}

	now := time.Now().UTC()

	for i := range dates {
		dates[i].PersonID = personID
		dates[i].CreatedAt = now
	}

	_, err = tx.NewInsert().Model(&dates).Exec(ctx)
	if err != nil {
		return fmt.Errorf("insert dates: %w", err)
	}

	return nil
}

// OnThisDay uses raw SQL because it JOINs important_date with person and scans
// into a non-model DTO (OnThisDayItem) with nullable nickname and computed YearsSince.
func (r *sqlRepo) OnThisDay(ctx context.Context, monthDay, todayISO string) ([]OnThisDayItem, error) {
	// The UNION includes synthetic birthday rows from person.date_of_birth for people
	// who have no explicit birthday entry in important_date, to avoid duplicates.
	query := `
		SELECT d.id, d.person_id, d.kind, d.label, d.date_value, d.recurring,
		       d.notes, d.position, d.created_at,
		       p.id, p.name, p.nickname
		  FROM important_date d
		  JOIN person p ON p.id = d.person_id
		 WHERE d.month_day = ?
		   AND (d.recurring = 1 OR d.date_value = ?)
		UNION ALL
		SELECT 0, p.id, 'birthday', 'Birthday', p.date_of_birth, 1,
		       '', 0, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'),
		       p.id, p.name, p.nickname
		  FROM person p
		 WHERE p.date_of_birth IS NOT NULL AND p.date_of_birth != ''
		   AND substr(p.date_of_birth, 6) = ?
		   AND NOT EXISTS (
		     SELECT 1 FROM important_date d2
		      WHERE d2.person_id = p.id AND d2.kind = 'birthday'
		   )
		 ORDER BY name COLLATE NOCASE
	`

	rows, err := r.db.QueryContext(ctx, query, monthDay, todayISO, monthDay)
	if err != nil {
		return nil, fmt.Errorf("query on this day: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []OnThisDayItem

	for rows.Next() {
		var (
			item      OnThisDayItem
			recurring int
			createdAt string
			nickname  sql.NullString
		)

		err := rows.Scan(
			&item.Date.ID,
			&item.Date.PersonID,
			&item.Date.Kind,
			&item.Date.Label,
			&item.Date.DateValue,
			&recurring,
			&item.Date.Notes,
			&item.Date.Position,
			&createdAt,
			&item.Person.ID,
			&item.Person.Name,
			&nickname,
		)
		if err != nil {
			return nil, fmt.Errorf("scan on this day: %w", err)
		}

		item.Date.Recurring = recurring == 1
		if createdAt != "" {
			item.Date.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		}

		if nickname.Valid {
			item.Person.Nickname = nickname.String
		}

		if !item.Date.IsYearless() && item.Date.Recurring {
			dateVal, err := time.Parse("2006-01-02", item.Date.DateValue)
			if err == nil {
				today, err := time.Parse("2006-01-02", todayISO)
				if err == nil {
					item.YearsSince = today.Year() - dateVal.Year()
				}
			}
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate on this day: %w", err)
	}

	return items, nil
}

// ListAll uses raw SQL because it JOINs important_date with person and scans
// into a non-model DTO (OnThisDayItem) with nullable nickname.
func (r *sqlRepo) ListAll(ctx context.Context) ([]OnThisDayItem, error) {
	// The UNION includes synthetic birthday rows from person.date_of_birth for people
	// who have no explicit birthday entry in important_date, to avoid duplicates.
	query := `
		SELECT d.id, d.person_id, d.kind, d.label, d.date_value, d.recurring,
		       d.notes, d.position, d.created_at,
		       p.id, p.name, p.nickname
		  FROM important_date d
		  JOIN person p ON p.id = d.person_id
		UNION ALL
		SELECT 0, p.id, 'birthday', 'Birthday', p.date_of_birth, 1,
		       '', 0, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'),
		       p.id, p.name, p.nickname
		  FROM person p
		 WHERE p.date_of_birth IS NOT NULL AND p.date_of_birth != ''
		   AND NOT EXISTS (
		     SELECT 1 FROM important_date d2
		      WHERE d2.person_id = p.id AND d2.kind = 'birthday'
		   )
		 ORDER BY name COLLATE NOCASE
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query list all: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []OnThisDayItem

	for rows.Next() {
		var (
			item      OnThisDayItem
			recurring int
			createdAt string
			nickname  sql.NullString
		)

		err := rows.Scan(
			&item.Date.ID,
			&item.Date.PersonID,
			&item.Date.Kind,
			&item.Date.Label,
			&item.Date.DateValue,
			&recurring,
			&item.Date.Notes,
			&item.Date.Position,
			&createdAt,
			&item.Person.ID,
			&item.Person.Name,
			&nickname,
		)
		if err != nil {
			return nil, fmt.Errorf("scan list all: %w", err)
		}

		item.Date.Recurring = recurring == 1
		if createdAt != "" {
			item.Date.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		}

		if nickname.Valid {
			item.Person.Nickname = nickname.String
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate list all: %w", err)
	}

	return items, nil
}
