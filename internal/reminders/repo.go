package reminders

import (
	"context"
	"encoding/json"
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

func marshalRecurrenceRule(rule *RecurrenceRule) (*string, error) {
	if rule == nil {
		return nil, nil
	}

	b, err := json.Marshal(rule)
	if err != nil {
		return nil, fmt.Errorf("marshal recurrence_rule: %w", err)
	}

	s := string(b)

	return &s, nil
}

func unmarshalRecurrenceRule(s *string) *RecurrenceRule {
	if s == nil {
		return nil
	}

	var rule RecurrenceRule
	if err := json.Unmarshal([]byte(*s), &rule); err != nil {
		return nil
	}

	return &rule
}

// reminderRow is used to scan Reminder rows that include the raw recurrence_rule JSON.
type reminderRow struct {
	Reminder
	RuleJSON *string `bun:"recurrence_rule"`
}

func (row *reminderRow) toReminder() *Reminder {
	rem := row.Reminder
	rem.RecurrenceRule = unmarshalRecurrenceRule(row.RuleJSON)

	return &rem
}

func (r *Repo) Create(ctx context.Context, tx bun.Tx, rem *Reminder) (int64, error) {
	rem.CreatedAt = time.Now().UTC()
	rem.UpdatedAt = rem.CreatedAt

	ruleStr, err := marshalRecurrenceRule(rem.RecurrenceRule)
	if err != nil {
		return 0, err
	}

	// RecurrenceRule is tagged bun:"-" so bun won't include it automatically.
	// Use Value() to inject the JSON string for that column.
	_, err = tx.NewInsert().Model(rem).
		Value("recurrence_rule", "?", ruleStr).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("insert reminder: %w", err)
	}

	return rem.ID, nil
}

func (r *Repo) GetByID(ctx context.Context, id int64) (*Reminder, error) {
	var row reminderRow

	err := r.db.NewSelect().
		TableExpr("reminder r").
		ColumnExpr("r.*, r.recurrence_rule AS recurrence_rule").
		Where("r.id = ?", id).
		Scan(ctx, &row)
	if err != nil {
		return nil, err
	}

	return row.toReminder(), nil
}

func (r *Repo) Update(ctx context.Context, tx bun.Tx, rem *Reminder) error {
	rem.UpdatedAt = time.Now().UTC()

	ruleStr, err := marshalRecurrenceRule(rem.RecurrenceRule)
	if err != nil {
		return err
	}

	_, err = tx.NewUpdate().
		TableExpr("reminder").
		Set("title = ?", rem.Title).
		Set("notes = ?", rem.Notes).
		Set("due_date = ?", rem.DueDate).
		Set("person_id = ?", rem.PersonID).
		Set("important_date_id = ?", rem.ImportantDateID).
		Set("completed = ?", rem.Completed).
		Set("completed_at = ?", rem.CompletedAt).
		Set("recurrence_rule = ?", ruleStr).
		Set("recurrence_end_date = ?", rem.RecurrenceEndDate).
		Set("updated_at = ?", rem.UpdatedAt).
		Where("id = ?", rem.ID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update reminder: %w", err)
	}

	return nil
}

func (r *Repo) Delete(ctx context.Context, tx bun.Tx, id int64) error {
	_, err := tx.NewDelete().Model((*Reminder)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete reminder: %w", err)
	}

	return nil
}

func (r *Repo) List(ctx context.Context, params ListParams) ([]ReminderWithPerson, error) {
	q := r.db.NewSelect().
		TableExpr("reminder r").
		ColumnExpr("r.*, r.recurrence_rule AS recurrence_rule, COALESCE(p.name, '') AS person_name").
		Join("LEFT JOIN person p ON r.person_id = p.id")

	switch params.Status {
	case "pending":
		q = q.Where("r.completed = ?", false)
	case "completed":
		q = q.Where("r.completed = ?", true)
	case "overdue":
		q = q.Where("r.completed = ? AND r.due_date < ?", false, time.Now().UTC())
	}

	if params.PersonID != nil {
		q = q.Where("r.person_id = ?", *params.PersonID)
	}

	q = q.OrderExpr("r.due_date ASC")

	if params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		q = q.Limit(params.PageSize).Offset(offset)
	}

	var rows []struct {
		reminderRow
		PersonName string `bun:"person_name"`
	}

	if err := q.Scan(ctx, &rows); err != nil {
		return nil, fmt.Errorf("query reminders: %w", err)
	}

	results := make([]ReminderWithPerson, 0, len(rows))
	for _, row := range rows {
		results = append(results, ReminderWithPerson{
			Reminder:   *row.toReminder(),
			PersonName: row.PersonName,
		})
	}

	return results, nil
}

func (r *Repo) ListUpcoming(ctx context.Context, days int) ([]ReminderWithPerson, error) {
	now := time.Now().UTC()
	future := now.AddDate(0, 0, days)

	var rows []struct {
		reminderRow
		PersonName string `bun:"person_name"`
	}

	err := r.db.NewSelect().
		TableExpr("reminder r").
		ColumnExpr("r.*, r.recurrence_rule AS recurrence_rule, COALESCE(p.name, '') AS person_name").
		Join("LEFT JOIN person p ON r.person_id = p.id").
		Where("r.completed = ? AND r.due_date >= ? AND r.due_date <= ?", false, now, future).
		OrderExpr("r.due_date ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("query upcoming reminders: %w", err)
	}

	results := make([]ReminderWithPerson, 0, len(rows))
	for _, row := range rows {
		results = append(results, ReminderWithPerson{
			Reminder:   *row.toReminder(),
			PersonName: row.PersonName,
		})
	}

	return results, nil
}

func (r *Repo) ListOverdue(ctx context.Context) ([]ReminderWithPerson, error) {
	var rows []struct {
		reminderRow
		PersonName string `bun:"person_name"`
	}

	err := r.db.NewSelect().
		TableExpr("reminder r").
		ColumnExpr("r.*, r.recurrence_rule AS recurrence_rule, COALESCE(p.name, '') AS person_name").
		Join("LEFT JOIN person p ON r.person_id = p.id").
		Where("r.completed = ? AND r.due_date < ?", false, time.Now().UTC()).
		OrderExpr("r.due_date ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("query overdue reminders: %w", err)
	}

	results := make([]ReminderWithPerson, 0, len(rows))
	for _, row := range rows {
		results = append(results, ReminderWithPerson{
			Reminder:   *row.toReminder(),
			PersonName: row.PersonName,
		})
	}

	return results, nil
}

func (r *Repo) MarkComplete(ctx context.Context, tx bun.Tx, id int64, completedAt time.Time) error {
	_, err := tx.NewUpdate().
		TableExpr("reminder").
		Set("completed = ?", true).
		Set("completed_at = ?", completedAt).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("mark reminder complete: %w", err)
	}

	return nil
}

// FindBirthdayRemindersByPersonID returns PENDING birthday reminders for a person.
func (r *Repo) FindBirthdayRemindersByPersonID(ctx context.Context, personID int64) ([]*Reminder, error) {
	var rows []reminderRow

	err := r.db.NewSelect().
		TableExpr("reminder r").
		ColumnExpr("r.*, r.recurrence_rule AS recurrence_rule").
		Where("r.person_id = ?", personID).
		Where("r.completed = ?", false).
		Where("json_extract(r.recurrence_rule, '$.type') = ?", "birthday").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("find birthday reminders: %w", err)
	}

	rems := make([]*Reminder, 0, len(rows))
	for i := range rows {
		rems = append(rems, rows[i].toReminder())
	}

	return rems, nil
}

// DeleteBirthdayRemindersByPersonID deletes ALL birthday reminders (pending + completed) for a person.
func (r *Repo) DeleteBirthdayRemindersByPersonID(ctx context.Context, tx bun.Tx, personID int64) error {
	_, err := tx.NewDelete().
		TableExpr("reminder").
		Where("person_id = ? AND json_extract(recurrence_rule, '$.type') = ?", personID, "birthday").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete birthday reminders: %w", err)
	}

	return nil
}

// HasBirthdayReminderForPerson returns true if the person has at least one non-completed birthday reminder.
func (r *Repo) HasBirthdayReminderForPerson(ctx context.Context, personID int64) (bool, error) {
	var count int

	err := r.db.NewSelect().
		TableExpr("reminder").
		ColumnExpr("COUNT(*)").
		Where("person_id = ? AND completed = ? AND json_extract(recurrence_rule, '$.type') = ?", personID, false, "birthday").
		Scan(ctx, &count)
	if err != nil {
		return false, fmt.Errorf("check birthday reminder: %w", err)
	}

	return count > 0, nil
}

func (r *Repo) CountByStatus(ctx context.Context, status string) (int, error) {
	q := r.db.NewSelect().TableExpr("reminder").ColumnExpr("COUNT(*)")

	switch status {
	case "pending":
		q = q.Where("completed = ?", false)
	case "completed":
		q = q.Where("completed = ?", true)
	case "overdue":
		q = q.Where("completed = ? AND due_date < ?", false, time.Now().UTC())
	}

	var count int
	if err := q.Scan(ctx, &count); err != nil {
		return 0, fmt.Errorf("count reminders: %w", err)
	}

	return count, nil
}
