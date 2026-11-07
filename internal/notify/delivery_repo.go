package notify

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/uptrace/bun"
)

const (
	statusPending   = "pending"
	statusSent      = "sent"
	statusFailed    = "failed"
	statusCancelled = "cancelled"

	retention = 30 * 24 * time.Hour
	staleAge  = 24 * time.Hour
)

// Delivery is one (reminder, due_at, channel) notification and its retry state.
// due_at is the verbatim reminder.due_date text so edits are detected by string equality.
type Delivery struct {
	bun.BaseModel `bun:"table:reminder_delivery"`

	ReminderID     int64      `bun:"reminder_id,pk"`
	DueAt          string     `bun:"due_at,pk"`
	ChannelID      int64      `bun:"channel_id,pk"`
	Status         string     `bun:"status"`
	Attempts       int        `bun:"attempts"`
	NextAttemptAt  time.Time  `bun:"next_attempt_at"`
	FireAt         time.Time  `bun:"fire_at"`
	LastErrorClass string     `bun:"last_error_class"`
	LastStatus     int        `bun:"last_status"`
	SentAt         *time.Time `bun:"sent_at"`
	CreatedAt      time.Time  `bun:"created_at"`
}

// liveReminder is the join every claim, cancel and pre-send check shares: the
// reminder must still exist unchanged, be incomplete, and belong to nobody or
// to a person who is neither soft-deleted nor archived. It never relies on FK
// enforcement (the pragma only reaches one pooled connection).
const liveReminder = `r.id = d.reminder_id AND r.due_date = d.due_at AND r.completed = 0
 AND (r.person_id IS NULL OR (p.deleted_at IS NULL AND p.archived_at IS NULL))`

type candidate struct {
	ID        int64     `bun:"id"`
	DueRaw    string    `bun:"due_raw"`
	DueDate   time.Time `bun:"due_date"`
	CreatedAt time.Time `bun:"created_at"`
}

// candidates are incomplete reminders near now. The SQL window is deliberately
// wide because due_date is TEXT; exact checks happen in Go on parsed times.
func candidates(ctx context.Context, db *bun.DB, now time.Time) ([]candidate, error) {
	var out []candidate

	err := db.NewSelect().
		TableExpr("reminder AS r").
		ColumnExpr("r.id, r.due_date AS due_raw, r.due_date, r.created_at").
		Join("LEFT JOIN person AS p ON p.id = r.person_id").
		Where("r.completed = 0").
		Where("r.due_date >= ? AND r.due_date <= ?", now.Add(-72*time.Hour), now.Add(48*time.Hour)).
		Where("(r.person_id IS NULL OR (p.deleted_at IS NULL AND p.archived_at IS NULL))").
		Scan(ctx, &out)
	if err != nil {
		return nil, fmt.Errorf("notify: candidates: %w", err)
	}

	return out, nil
}

func insertDeliveries(ctx context.Context, db *bun.DB, rows []*Delivery) error {
	if len(rows) == 0 {
		return nil
	}

	for chunk := range slices.Chunk(rows, 100) { // keeps one statement under SQLite's size limits
		if _, err := db.NewInsert().Model(&chunk).Ignore().Exec(ctx); err != nil {
			return fmt.Errorf("notify: enqueue: %w", err)
		}
	}

	return nil
}

// cancelDead cancels pending rows whose reminder changed, or that never got a
// try within the stale window.
func cancelDead(ctx context.Context, db *bun.DB, now time.Time) error {
	_, err := db.NewRaw(`UPDATE reminder_delivery AS d SET status = 'cancelled'
WHERE d.status = 'pending' AND (
  NOT EXISTS (SELECT 1 FROM reminder r LEFT JOIN person p ON p.id = r.person_id WHERE `+liveReminder+`)
  OR NOT EXISTS (SELECT 1 FROM notification_channel c WHERE c.id = d.channel_id)
  OR (d.attempts = 0 AND d.fire_at <= ?))`, now.Add(-staleAge)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("notify: cancel dead: %w", err)
	}

	return nil
}

type claimed struct {
	ReminderID int64  `bun:"reminder_id"`
	DueAt      string `bun:"due_at"`
	ChannelID  int64  `bun:"channel_id"`
	Attempts   int    `bun:"attempts"`
}

// claim leases up to limit due rows of one channel by bumping attempts and
// pushing next_attempt_at to the lease end. One short statement, no tx.
func claim(ctx context.Context, db *bun.DB, channelID int64, now, lease time.Time, limit int) ([]claimed, error) {
	var out []claimed

	err := db.NewRaw(`UPDATE reminder_delivery SET attempts = attempts + 1, next_attempt_at = ?
WHERE rowid IN (
  SELECT d.rowid FROM reminder_delivery d
  JOIN reminder r ON r.id = d.reminder_id
  JOIN notification_channel c ON c.id = d.channel_id AND c.enabled = 1 AND d.fire_at >= c.enabled_at
  LEFT JOIN person p ON p.id = r.person_id
  WHERE d.channel_id = ? AND d.status = 'pending' AND d.next_attempt_at <= ?
    AND `+liveReminder+`
  ORDER BY d.fire_at LIMIT ?)
RETURNING reminder_id, due_at, channel_id, attempts`, lease, channelID, now, limit).Scan(ctx, &out)
	if err != nil {
		return nil, fmt.Errorf("notify: claim: %w", err)
	}

	return out, nil
}

type liveMessage struct {
	Title      string    `bun:"title"`
	Notes      string    `bun:"notes"`
	DueDate    time.Time `bun:"due_date"`
	PersonName string    `bun:"person_name"`
}

// loadLive returns the reminder content if the delivery is still valid, else sql.ErrNoRows.
func loadLive(ctx context.Context, db *bun.DB, c claimed) (liveMessage, error) {
	var m liveMessage

	err := db.NewRaw(`SELECT r.title, r.notes, r.due_date, COALESCE(p.name, '') AS person_name
FROM reminder_delivery d
JOIN reminder r ON r.id = d.reminder_id
LEFT JOIN person p ON p.id = r.person_id
WHERE d.reminder_id = ? AND d.due_at = ? AND d.channel_id = ? AND `+liveReminder, c.ReminderID, c.DueAt, c.ChannelID).
		Scan(ctx, &m)

	return m, err
}

// finish records an outcome. The status='pending' guard keeps a row cancelled
// mid-send (channel disabled, reminder completed) cancelled.
func finish(ctx context.Context, db *bun.DB, c claimed, status string, se *SendError, next, now time.Time) error {
	class, httpStatus := "", 0
	if se != nil {
		class, httpStatus = se.Class, se.Status
	}

	var sentAt *time.Time
	if status == statusSent {
		sentAt = &now
	}

	_, err := db.NewRaw(`UPDATE reminder_delivery SET status = ?, last_error_class = ?, last_status = ?,
  next_attempt_at = ?, sent_at = ?
WHERE reminder_id = ? AND due_at = ? AND channel_id = ? AND status = 'pending'`,
		status, class, httpStatus, next, sentAt, c.ReminderID, c.DueAt, c.ChannelID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("notify: finish: %w", err)
	}

	return nil
}

func cancelOne(ctx context.Context, db *bun.DB, c claimed) error {
	_, err := db.NewRaw(`UPDATE reminder_delivery SET status = 'cancelled'
WHERE reminder_id = ? AND due_at = ? AND channel_id = ? AND status = 'pending'`,
		c.ReminderID, c.DueAt, c.ChannelID).Exec(ctx)

	return err
}

func housekeeping(ctx context.Context, db *bun.DB, now time.Time) error {
	_, err := db.NewRaw(`DELETE FROM reminder_delivery WHERE status != 'pending' AND created_at < ?`,
		now.Add(-retention)).Exec(ctx)

	return err
}
