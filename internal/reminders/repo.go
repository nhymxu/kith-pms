package reminders

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

func (r *Repo) Create(ctx context.Context, tx *sql.Tx, rem *Reminder) (int64, error) {
	query := `
		INSERT INTO reminder (title, notes, due_date, person_id, important_date_id, completed, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	var completedInt int
	if rem.Completed {
		completedInt = 1
	}

	var completedAtStr *string

	if rem.CompletedAt != nil {
		s := rem.CompletedAt.Format(time.RFC3339)
		completedAtStr = &s
	}

	result, err := tx.ExecContext(ctx, query,
		rem.Title, rem.Notes, rem.DueDate.Format(time.RFC3339),
		rem.PersonID, rem.ImportantDateID, completedInt, completedAtStr)
	if err != nil {
		return 0, fmt.Errorf("insert reminder: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert id: %w", err)
	}

	return id, nil
}

func (r *Repo) GetByID(ctx context.Context, id int64) (*Reminder, error) {
	query := `
		SELECT id, title, notes, due_date, person_id, important_date_id,
		       completed, completed_at, created_at, updated_at
		FROM reminder
		WHERE id = ?
	`

	var (
		rem                                    Reminder
		dueDateStr, createdAtStr, updatedAtStr string
		completedInt                           int
		completedAtStr                         *string
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rem.ID, &rem.Title, &rem.Notes, &dueDateStr, &rem.PersonID, &rem.ImportantDateID,
		&completedInt, &completedAtStr, &createdAtStr, &updatedAtStr,
	)
	if err != nil {
		return nil, err
	}

	rem.Completed = completedInt == 1
	rem.DueDate, _ = time.Parse(time.RFC3339, dueDateStr)
	rem.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

	rem.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	if completedAtStr != nil {
		t, _ := time.Parse(time.RFC3339, *completedAtStr)
		rem.CompletedAt = &t
	}

	return &rem, nil
}

func (r *Repo) Update(ctx context.Context, tx *sql.Tx, rem *Reminder) error {
	query := `
		UPDATE reminder
		SET title = ?, notes = ?, due_date = ?, person_id = ?, important_date_id = ?,
		    completed = ?, completed_at = ?, updated_at = ?
		WHERE id = ?
	`

	var completedInt int
	if rem.Completed {
		completedInt = 1
	}

	var completedAtStr *string

	if rem.CompletedAt != nil {
		s := rem.CompletedAt.Format(time.RFC3339)
		completedAtStr = &s
	}

	_, err := tx.ExecContext(ctx, query,
		rem.Title, rem.Notes, rem.DueDate.Format(time.RFC3339),
		rem.PersonID, rem.ImportantDateID, completedInt, completedAtStr,
		time.Now().Format(time.RFC3339), rem.ID)
	if err != nil {
		return fmt.Errorf("update reminder: %w", err)
	}

	return nil
}

func (r *Repo) Delete(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `DELETE FROM reminder WHERE id = ?`

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete reminder: %w", err)
	}

	return nil
}

func (r *Repo) List(ctx context.Context, params ListParams) ([]ReminderWithPerson, error) {
	query := `
		SELECT r.id, r.title, r.notes, r.due_date, r.person_id, r.important_date_id,
		       r.completed, r.completed_at, r.created_at, r.updated_at,
		       COALESCE(p.name, '') as person_name
		FROM reminder r
		LEFT JOIN person p ON r.person_id = p.id
		WHERE 1=1
	`
	args := []interface{}{}

	switch params.Status {
	case "pending":
		query += " AND r.completed = 0"
	case "completed":
		query += " AND r.completed = 1"
	case "overdue":
		query += " AND r.completed = 0 AND r.due_date < ?"

		args = append(args, time.Now().Format(time.RFC3339))
	}

	if params.PersonID != nil {
		query += " AND r.person_id = ?"

		args = append(args, *params.PersonID)
	}

	query += " ORDER BY r.due_date ASC"

	if params.PageSize > 0 {
		query += " LIMIT ? OFFSET ?"
		offset := (params.Page - 1) * params.PageSize
		args = append(args, params.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query reminders: %w", err)
	}
	defer rows.Close()

	var results []ReminderWithPerson

	for rows.Next() {
		var (
			rwp                                    ReminderWithPerson
			dueDateStr, createdAtStr, updatedAtStr string
			completedInt                           int
			completedAtStr                         *string
		)

		err := rows.Scan(
			&rwp.ID, &rwp.Title, &rwp.Notes, &dueDateStr, &rwp.PersonID, &rwp.ImportantDateID,
			&completedInt, &completedAtStr, &createdAtStr, &updatedAtStr, &rwp.PersonName,
		)
		if err != nil {
			return nil, fmt.Errorf("scan reminder: %w", err)
		}

		rwp.Completed = completedInt == 1
		rwp.DueDate, _ = time.Parse(time.RFC3339, dueDateStr)
		rwp.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

		rwp.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		if completedAtStr != nil {
			t, _ := time.Parse(time.RFC3339, *completedAtStr)
			rwp.CompletedAt = &t
		}

		results = append(results, rwp)
	}

	return results, rows.Err()
}

func (r *Repo) ListUpcoming(ctx context.Context, days int) ([]ReminderWithPerson, error) {
	now := time.Now()
	future := now.AddDate(0, 0, days)

	query := `
		SELECT r.id, r.title, r.notes, r.due_date, r.person_id, r.important_date_id,
		       r.completed, r.completed_at, r.created_at, r.updated_at,
		       COALESCE(p.name, '') as person_name
		FROM reminder r
		LEFT JOIN person p ON r.person_id = p.id
		WHERE r.completed = 0 AND r.due_date >= ? AND r.due_date <= ?
		ORDER BY r.due_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, now.Format(time.RFC3339), future.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("query upcoming reminders: %w", err)
	}
	defer rows.Close()

	var results []ReminderWithPerson

	for rows.Next() {
		var (
			rwp                                    ReminderWithPerson
			dueDateStr, createdAtStr, updatedAtStr string
			completedInt                           int
			completedAtStr                         *string
		)

		err := rows.Scan(
			&rwp.ID, &rwp.Title, &rwp.Notes, &dueDateStr, &rwp.PersonID, &rwp.ImportantDateID,
			&completedInt, &completedAtStr, &createdAtStr, &updatedAtStr, &rwp.PersonName,
		)
		if err != nil {
			return nil, fmt.Errorf("scan reminder: %w", err)
		}

		rwp.Completed = completedInt == 1
		rwp.DueDate, _ = time.Parse(time.RFC3339, dueDateStr)
		rwp.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

		rwp.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		if completedAtStr != nil {
			t, _ := time.Parse(time.RFC3339, *completedAtStr)
			rwp.CompletedAt = &t
		}

		results = append(results, rwp)
	}

	return results, rows.Err()
}

func (r *Repo) ListOverdue(ctx context.Context) ([]ReminderWithPerson, error) {
	query := `
		SELECT r.id, r.title, r.notes, r.due_date, r.person_id, r.important_date_id,
		       r.completed, r.completed_at, r.created_at, r.updated_at,
		       COALESCE(p.name, '') as person_name
		FROM reminder r
		LEFT JOIN person p ON r.person_id = p.id
		WHERE r.completed = 0 AND r.due_date < ?
		ORDER BY r.due_date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, time.Now().Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("query overdue reminders: %w", err)
	}
	defer rows.Close()

	var results []ReminderWithPerson

	for rows.Next() {
		var (
			rwp                                    ReminderWithPerson
			dueDateStr, createdAtStr, updatedAtStr string
			completedInt                           int
			completedAtStr                         *string
		)

		err := rows.Scan(
			&rwp.ID, &rwp.Title, &rwp.Notes, &dueDateStr, &rwp.PersonID, &rwp.ImportantDateID,
			&completedInt, &completedAtStr, &createdAtStr, &updatedAtStr, &rwp.PersonName,
		)
		if err != nil {
			return nil, fmt.Errorf("scan reminder: %w", err)
		}

		rwp.Completed = completedInt == 1
		rwp.DueDate, _ = time.Parse(time.RFC3339, dueDateStr)
		rwp.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

		rwp.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		if completedAtStr != nil {
			t, _ := time.Parse(time.RFC3339, *completedAtStr)
			rwp.CompletedAt = &t
		}

		results = append(results, rwp)
	}

	return results, rows.Err()
}

func (r *Repo) MarkComplete(ctx context.Context, tx *sql.Tx, id int64, completedAt time.Time) error {
	query := `
		UPDATE reminder
		SET completed = 1, completed_at = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := tx.ExecContext(ctx, query,
		completedAt.Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
		id)
	if err != nil {
		return fmt.Errorf("mark reminder complete: %w", err)
	}

	return nil
}

func (r *Repo) CountByStatus(ctx context.Context, status string) (int, error) {
	query := `SELECT COUNT(*) FROM reminder WHERE 1=1`
	args := []interface{}{}

	switch status {
	case "pending":
		query += " AND completed = 0"
	case "completed":
		query += " AND completed = 1"
	case "overdue":
		query += " AND completed = 0 AND due_date < ?"

		args = append(args, time.Now().Format(time.RFC3339))
	}

	var count int

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count reminders: %w", err)
	}

	return count, nil
}
