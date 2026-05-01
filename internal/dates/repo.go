package dates

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ImportantDateRepo interface {
	ListByPerson(ctx context.Context, personID int64) ([]ImportantDate, error)
	ReplaceAll(ctx context.Context, tx *sql.Tx, personID int64, dates []ImportantDate) error
	OnThisDay(ctx context.Context, monthDay, todayISO string) ([]OnThisDayItem, error)
	ListAll(ctx context.Context) ([]OnThisDayItem, error)
}

type sqlRepo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) ImportantDateRepo {
	return &sqlRepo{db: db}
}

func (r *sqlRepo) ListByPerson(ctx context.Context, personID int64) ([]ImportantDate, error) {
	query := `
		SELECT id, person_id, kind, label, date_value, recurring, notes, position, created_at
		FROM important_date
		WHERE person_id = ?
		ORDER BY position, id
	`
	rows, err := r.db.QueryContext(ctx, query, personID)
	if err != nil {
		return nil, fmt.Errorf("query dates: %w", err)
	}
	defer rows.Close()

	var dates []ImportantDate
	for rows.Next() {
		var d ImportantDate
		var recurring int
		var createdAt string
		err := rows.Scan(
			&d.ID,
			&d.PersonID,
			&d.Kind,
			&d.Label,
			&d.DateValue,
			&recurring,
			&d.Notes,
			&d.Position,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan date: %w", err)
		}
		d.Recurring = recurring == 1
		d.CreatedAt, _ = parseTimestamp(createdAt)
		dates = append(dates, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dates: %w", err)
	}
	return dates, nil
}

func (r *sqlRepo) ReplaceAll(ctx context.Context, tx *sql.Tx, personID int64, dates []ImportantDate) error {
	// Delete existing
	_, err := tx.ExecContext(ctx, "DELETE FROM important_date WHERE person_id = ?", personID)
	if err != nil {
		return fmt.Errorf("delete existing dates: %w", err)
	}

	// Insert new
	if len(dates) == 0 {
		return nil
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO important_date (person_id, kind, label, date_value, recurring, notes, position)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, d := range dates {
		recurringInt := 0
		if d.Recurring {
			recurringInt = 1
		}
		_, err := stmt.ExecContext(ctx, personID, d.Kind, d.Label, d.DateValue, recurringInt, d.Notes, d.Position)
		if err != nil {
			return fmt.Errorf("insert date: %w", err)
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

func (r *sqlRepo) OnThisDay(ctx context.Context, monthDay, todayISO string) ([]OnThisDayItem, error) {
	query := `
		SELECT d.id, d.person_id, d.kind, d.label, d.date_value, d.recurring,
		       d.notes, d.position, d.created_at,
		       p.id, p.name, p.nickname
		  FROM important_date d
		  JOIN person p ON p.id = d.person_id
		 WHERE d.month_day = ?
		   AND (d.recurring = 1 OR d.date_value = ?)
		 ORDER BY p.name COLLATE NOCASE
	`
	rows, err := r.db.QueryContext(ctx, query, monthDay, todayISO)
	if err != nil {
		return nil, fmt.Errorf("query on this day: %w", err)
	}
	defer rows.Close()

	var items []OnThisDayItem
	for rows.Next() {
		var item OnThisDayItem
		var recurring int
		var createdAt string
		var nickname sql.NullString
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
		item.Date.CreatedAt, _ = parseTimestamp(createdAt)
		if nickname.Valid {
			item.Person.Nickname = nickname.String
		}

		// Calculate YearsSince if year-having
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

func (r *sqlRepo) ListAll(ctx context.Context) ([]OnThisDayItem, error) {
	query := `
		SELECT d.id, d.person_id, d.kind, d.label, d.date_value, d.recurring,
		       d.notes, d.position, d.created_at,
		       p.id, p.name, p.nickname
		  FROM important_date d
		  JOIN person p ON p.id = d.person_id
		 ORDER BY d.month_day, p.name COLLATE NOCASE
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query list all: %w", err)
	}
	defer rows.Close()

	var items []OnThisDayItem
	for rows.Next() {
		var item OnThisDayItem
		var recurring int
		var createdAt string
		var nickname sql.NullString
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
		item.Date.CreatedAt, _ = parseTimestamp(createdAt)
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
