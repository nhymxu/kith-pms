package reminders

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nhymxu/kith-pms/internal/audit"
)

type Service struct {
	db    *sql.DB
	repo  *Repo
	Audit *audit.Service // optional; nil = no audit logging
}

func NewService(db *sql.DB) *Service {
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

	if err := s.repo.MarkComplete(ctx, tx, id, time.Now()); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityReminder, id, title, audit.ActionUpdate)
	}

	return nil
}

func (s *Service) CountByStatus(ctx context.Context, status string) (int, error) {
	return s.repo.CountByStatus(ctx, status)
}
