package work_history

import (
	"context"
	"database/sql"
	"fmt"
)

// Service provides business logic for work history operations.
type Service struct {
	db   *sql.DB
	repo WorkHistoryRepo
}

// NewService creates a new Service backed by the given database.
func NewService(db *sql.DB) *Service {
	return &Service{
		db:   db,
		repo: NewRepo(db),
	}
}

// ListByPerson returns all work history entries for the given person, ordered by position.
func (s *Service) ListByPerson(ctx context.Context, personID int64) ([]WorkEntry, error) {
	return s.repo.ListByPerson(ctx, personID)
}

// ReplaceForPerson replaces all work history entries for the given person in a single transaction.
func (s *Service) ReplaceForPerson(ctx context.Context, personID int64, entries []WorkEntry) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := s.repo.ReplaceAll(ctx, tx, personID, entries); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
