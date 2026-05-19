package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ActivityRepo interface {
	Create(ctx context.Context, tx *sql.Tx, a Activity) (int64, error)
	Update(ctx context.Context, tx *sql.Tx, a Activity) error
	Get(ctx context.Context, id int64) (*Activity, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params ListParams) ([]Activity, error)
	Count(ctx context.Context, params ListParams) (int, error)
}

type ActivityPersonRepo interface {
	ReplaceAll(ctx context.Context, tx *sql.Tx, activityID int64, personIDs []int64) error
	ListByActivity(ctx context.Context, activityID int64) ([]ActivityPerson, error)
}

// ---- sqlActivityRepo --------------------------------------------------------

type sqlActivityRepo struct{ db *sql.DB }

func NewActivityRepo(db *sql.DB) ActivityRepo { return &sqlActivityRepo{db: db} }

func (r *sqlActivityRepo) Create(ctx context.Context, tx *sql.Tx, a Activity) (int64, error) {
	const q = `
		INSERT INTO activity (title, occurred_at_date, occurred_at_time, content)
		VALUES (?, ?, ?, ?)`

	var oat any
	if a.OccurredAtTime != "" {
		oat = a.OccurredAtTime
	}

	res, err := tx.ExecContext(ctx, q, a.Title, a.OccurredAtDate, oat, a.Content)
	if err != nil {
		return 0, fmt.Errorf("journal: create activity: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("journal: last insert id: %w", err)
	}

	return id, nil
}

func (r *sqlActivityRepo) Update(ctx context.Context, tx *sql.Tx, a Activity) error {
	const q = `
		UPDATE activity
		SET title = ?, occurred_at_date = ?, occurred_at_time = ?,
		    content = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE id = ?`

	var oat any
	if a.OccurredAtTime != "" {
		oat = a.OccurredAtTime
	}

	_, err := tx.ExecContext(ctx, q, a.Title, a.OccurredAtDate, oat, a.Content, a.ID)
	if err != nil {
		return fmt.Errorf("journal: update activity: %w", err)
	}

	return nil
}

func (r *sqlActivityRepo) Get(ctx context.Context, id int64) (*Activity, error) {
	const q = `
		SELECT id, title, occurred_at_date, occurred_at_time, content, created_at, updated_at
		FROM activity WHERE id = ?`

	row := r.db.QueryRowContext(ctx, q, id)

	a, err := scanActivity(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("journal: get activity: %w", err)
	}

	return a, nil
}

func (r *sqlActivityRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM activity WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("journal: delete activity: %w", err)
	}

	return nil
}

func (r *sqlActivityRepo) Count(ctx context.Context, params ListParams) (int, error) {
	var (
		joins []string
		where []string
		args  []any
	)

	useFTS := strings.TrimSpace(params.Query) != ""
	if useFTS {
		joins = append(joins, "JOIN activity_fts ON activity_fts.rowid = activity.id")
		where = append(where, "activity_fts MATCH ?")
		args = append(args, sanitizeFTSQuery(params.Query))
	}

	if len(params.PersonIDs) > 0 {
		placeholders := strings.Repeat("?,", len(params.PersonIDs))
		placeholders = placeholders[:len(placeholders)-1]

		where = append(
			where,
			"activity.id IN (SELECT activity_id FROM activity_person WHERE person_id IN ("+placeholders+"))",
		)
		for _, pid := range params.PersonIDs {
			args = append(args, pid)
		}
	}

	if len(params.LabelIDs) > 0 {
		placeholders := strings.Repeat("?,", len(params.LabelIDs))
		placeholders = placeholders[:len(placeholders)-1]

		where = append(
			where,
			"activity.id IN (SELECT ap.activity_id FROM activity_person ap JOIN person_label pl ON pl.person_id = ap.person_id WHERE pl.label_id IN ("+placeholders+"))", //nolint:lll
		)
		for _, lid := range params.LabelIDs {
			args = append(args, lid)
		}
	}

	if params.FromDate != "" && params.ToDate != "" {
		where = append(where, "activity.occurred_at_date BETWEEN ? AND ?")
		args = append(args, params.FromDate, params.ToDate)
	} else if params.FromDate != "" {
		where = append(where, "activity.occurred_at_date >= ?")
		args = append(args, params.FromDate)
	} else if params.ToDate != "" {
		where = append(where, "activity.occurred_at_date <= ?")
		args = append(args, params.ToDate)
	}

	query := "SELECT COUNT(*) FROM activity"
	if len(joins) > 0 {
		query += " " + strings.Join(joins, " ")
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("journal: count activities: %w", err)
	}

	return total, nil
}

// List builds a dynamic SQL query based on non-zero filter fields in params.
func (r *sqlActivityRepo) List(ctx context.Context, params ListParams) ([]Activity, error) {
	var (
		joins []string
		where []string
		args  []any
	)

	// FTS full-text search — join activity_fts and use MATCH.
	useFTS := strings.TrimSpace(params.Query) != ""
	if useFTS {
		joins = append(joins, "JOIN activity_fts ON activity_fts.rowid = activity.id")
		where = append(where, "activity_fts MATCH ?")
		args = append(args, sanitizeFTSQuery(params.Query))
	}

	// Filter by person IDs (OR semantics — any linked person).
	if len(params.PersonIDs) > 0 {
		placeholders := strings.Repeat("?,", len(params.PersonIDs))
		placeholders = placeholders[:len(placeholders)-1]

		where = append(
			where,
			"activity.id IN (SELECT activity_id FROM activity_person WHERE person_id IN ("+placeholders+"))",
		)
		for _, pid := range params.PersonIDs {
			args = append(args, pid)
		}
	}

	// Filter by label IDs — activities involving any person with those labels.
	if len(params.LabelIDs) > 0 {
		placeholders := strings.Repeat("?,", len(params.LabelIDs))
		placeholders = placeholders[:len(placeholders)-1]

		where = append(
			where,
			"activity.id IN (SELECT ap.activity_id FROM activity_person ap"+
				" JOIN person_label pl ON pl.person_id = ap.person_id"+
				" WHERE pl.label_id IN ("+placeholders+"))",
		)
		for _, lid := range params.LabelIDs {
			args = append(args, lid)
		}
	}

	// Date range filter.
	if params.FromDate != "" && params.ToDate != "" {
		where = append(where, "activity.occurred_at_date BETWEEN ? AND ?")
		args = append(args, params.FromDate, params.ToDate)
	} else if params.FromDate != "" {
		where = append(where, "activity.occurred_at_date >= ?")
		args = append(args, params.FromDate)
	} else if params.ToDate != "" {
		where = append(where, "activity.occurred_at_date <= ?")
		args = append(args, params.ToDate)
	}

	query := `SELECT activity.id, activity.title, activity.occurred_at_date, activity.occurred_at_time,
	                 activity.content, activity.created_at, activity.updated_at
	          FROM activity`

	if len(joins) > 0 {
		query += " " + strings.Join(joins, " ")
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	if useFTS {
		query += " ORDER BY bm25(activity_fts), activity.occurred_at_date DESC, activity.id DESC"
	} else {
		query += " ORDER BY activity.occurred_at_date DESC, activity.id DESC"
	}

	// Pagination.
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 30
	}

	page := params.Page
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pageSize
	query += " LIMIT ? OFFSET ?"

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("journal: list query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var list []Activity

	for rows.Next() {
		a, err := scanActivityRow(rows)
		if err != nil {
			return nil, fmt.Errorf("journal: scan activity row: %w", err)
		}

		list = append(list, *a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: list rows err: %w", err)
	}

	return list, nil
}

// sanitizeFTSQuery escapes user input for use in FTS5 MATCH queries.
// Replaces double quotes with single quotes, then wraps in double quotes
// for phrase matching — prevents malformed MATCH syntax errors.
func sanitizeFTSQuery(q string) string {
	q = strings.TrimSpace(q)
	// Escape existing double quotes by doubling them (FTS5 phrase literal rules).
	q = strings.ReplaceAll(q, `"`, `""`)

	return `"` + q + `"`
}

// ---- row scanners -----------------------------------------------------------

type rowScanner interface {
	Scan(dest ...any) error
}

func scanActivity(s rowScanner) (*Activity, error) {
	var (
		a                    Activity
		oat                  sql.NullString
		createdAt, updatedAt string
	)

	err := s.Scan(
		&a.ID, &a.Title, &a.OccurredAtDate, &oat,
		&a.Content, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if oat.Valid {
		a.OccurredAtTime = oat.String
	}

	a.CreatedAt = parseTime(createdAt)
	a.UpdatedAt = parseTime(updatedAt)

	return &a, nil
}

func scanActivityRow(rows *sql.Rows) (*Activity, error) {
	var (
		a                    Activity
		oat                  sql.NullString
		createdAt, updatedAt string
	)

	err := rows.Scan(
		&a.ID, &a.Title, &a.OccurredAtDate, &oat,
		&a.Content, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if oat.Valid {
		a.OccurredAtTime = oat.String
	}

	a.CreatedAt = parseTime(createdAt)
	a.UpdatedAt = parseTime(updatedAt)

	return &a, nil
}

func parseTime(s string) time.Time {
	// Try multiple layouts stored in SQLite TEXT columns.
	for _, layout := range []string{
		"2006-01-02T15:04:05.999Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}

	return time.Time{}
}

// ---- sqlActivityPersonRepo --------------------------------------------------

type sqlActivityPersonRepo struct{ db *sql.DB }

func NewActivityPersonRepo(db *sql.DB) ActivityPersonRepo {
	return &sqlActivityPersonRepo{db: db}
}

// ReplaceAll deletes all person links for the activity and inserts new ones.
func (r *sqlActivityPersonRepo) ReplaceAll(ctx context.Context, tx *sql.Tx, activityID int64, personIDs []int64) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM activity_person WHERE activity_id = ?`, activityID); err != nil {
		return fmt.Errorf("journal: delete activity_person: %w", err)
	}

	for _, pid := range personIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO activity_person (activity_id, person_id) VALUES (?, ?)`,
			activityID, pid,
		); err != nil {
			return fmt.Errorf("journal: insert activity_person: %w", err)
		}
	}

	return nil
}

// ListByActivity returns all people linked to the given activity.
func (r *sqlActivityPersonRepo) ListByActivity(ctx context.Context, activityID int64) ([]ActivityPerson, error) {
	const q = `
		SELECT ap.person_id, p.name, COALESCE(p.nickname, ''), COALESCE(p.avatar_path, '')
		FROM activity_person ap
		JOIN person p ON p.id = ap.person_id
		WHERE ap.activity_id = ?
		ORDER BY p.name`

	rows, err := r.db.QueryContext(ctx, q, activityID)
	if err != nil {
		return nil, fmt.Errorf("journal: list activity people: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var list []ActivityPerson

	for rows.Next() {
		var ap ActivityPerson
		if err := rows.Scan(&ap.PersonID, &ap.Name, &ap.Nickname, &ap.AvatarPath); err != nil {
			return nil, fmt.Errorf("journal: scan activity person: %w", err)
		}

		list = append(list, ap)
	}

	return list, rows.Err()
}
