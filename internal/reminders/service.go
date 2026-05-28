package reminders

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/audit"
)

// JournalLastContacter is satisfied by journal.Service — used for relative_contact recurrence.
type JournalLastContacter interface {
	LastContactDate(ctx context.Context, personID int64) (time.Time, error)
}

type Service struct {
	db      *bun.DB
	repo    *Repo
	Audit   *audit.Service       // optional; nil = no audit logging
	Journal JournalLastContacter // optional; nil = fallback to completedAt
}

func NewService(db *bun.DB) *Service {
	return &Service{
		db:   db,
		repo: NewRepo(db),
	}
}

func (s *Service) Create(ctx context.Context, rem *Reminder) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	id, err := s.repo.Create(ctx, tx, rem)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityReminder, id, rem.Title, audit.ActionCreate)
	}

	return id, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Reminder, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, rem *Reminder) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.Update(ctx, tx, rem); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityReminder, rem.ID, rem.Title, audit.ActionUpdate)
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	var title string

	if s.Audit != nil {
		if r, err := s.repo.GetByID(ctx, id); err == nil && r != nil {
			title = r.Title
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.Delete(ctx, tx, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityReminder, id, title, audit.ActionDelete)
	}

	return nil
}

func (s *Service) List(ctx context.Context, params ListParams) ([]ReminderWithPerson, error) {
	return s.repo.List(ctx, params)
}

func (s *Service) GetUpcoming(ctx context.Context, days int) ([]ReminderWithPerson, error) {
	return s.repo.ListUpcoming(ctx, days)
}

func (s *Service) GetOverdue(ctx context.Context) ([]ReminderWithPerson, error) {
	return s.repo.ListOverdue(ctx)
}

func (s *Service) MarkComplete(ctx context.Context, id int64) error {
	rem, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get reminder: %w", err)
	}

	now := time.Now()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.MarkComplete(ctx, tx, id, now); err != nil {
		return err
	}

	if rem.RecurrenceRule != nil {
		var lastContact time.Time

		if s.Journal != nil && rem.PersonID != nil {
			lastContact, _ = s.Journal.LastContactDate(ctx, *rem.PersonID)
		}

		rem.CompletedAt = &now
		nextDue := computeNextDue(rem, lastContact)

		if !nextDue.IsZero() {
			next := &Reminder{
				Title:             rem.Title,
				Notes:             rem.Notes,
				PersonID:          rem.PersonID,
				ImportantDateID:   rem.ImportantDateID,
				RecurrenceRule:    rem.RecurrenceRule,
				RecurrenceEndDate: rem.RecurrenceEndDate,
				DueDate:           nextDue,
			}

			if _, err := s.repo.Create(ctx, tx, next); err != nil {
				return fmt.Errorf("spawn next occurrence: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityReminder, id, rem.Title, audit.ActionUpdate)
	}

	return nil
}

func (s *Service) CountByStatus(ctx context.Context, status string) (int, error) {
	return s.repo.CountByStatus(ctx, status)
}
