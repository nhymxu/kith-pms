package journal

import (
	"context"
	"database/sql"
	"fmt"
)

const defaultPageSize = 30

// ListParams holds filter and pagination parameters for listing activities.
type ListParams struct {
	Query     string
	PersonIDs []int64
	LabelIDs  []int64
	FromDate  string // "YYYY-MM-DD"
	ToDate    string // "YYYY-MM-DD"
	Page      int
	PageSize  int // default 30
}

// Service provides business logic for managing journal activities.
type Service struct {
	DB         *sql.DB
	Activities ActivityRepo
	Links      ActivityPersonRepo
}

// NewService constructs a Service wired to db.
func NewService(db *sql.DB) *Service {
	return &Service{
		DB:         db,
		Activities: NewActivityRepo(db),
		Links:      NewActivityPersonRepo(db),
	}
}

// Create inserts a new activity and links people in a single transaction.
func (s *Service) Create(ctx context.Context, a Activity, personIDs []int64) (int64, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("journal: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	id, err := s.Activities.Create(ctx, tx, a)
	if err != nil {
		return 0, err
	}

	if err := s.Links.ReplaceAll(ctx, tx, id, personIDs); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("journal: commit create: %w", err)
	}
	return id, nil
}

// Update replaces an activity's fields and all person links in a single transaction.
func (s *Service) Update(ctx context.Context, a Activity, personIDs []int64) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("journal: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.Activities.Update(ctx, tx, a); err != nil {
		return err
	}
	if err := s.Links.ReplaceAll(ctx, tx, a.ID, personIDs); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("journal: commit update: %w", err)
	}
	return nil
}

// Get returns an activity by ID with its linked people populated.
// Returns nil, nil when not found.
func (s *Service) Get(ctx context.Context, id int64) (*Activity, error) {
	a, err := s.Activities.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, nil
	}

	people, err := s.Links.ListByActivity(ctx, id)
	if err != nil {
		return nil, err
	}
	a.People = people
	return a, nil
}

// Delete removes an activity; FTS mirror is updated by the activity_ad trigger.
func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.Activities.Delete(ctx, id)
}

// List returns a paginated, optionally filtered list of activities (no people populated).
func (s *Service) List(ctx context.Context, params ListParams) ([]Activity, error) {
	if params.PageSize <= 0 {
		params.PageSize = defaultPageSize
	}
	if params.PageSize > 500 {
		params.PageSize = 500
	}
	if params.Page < 1 {
		params.Page = 1
	}
	return s.Activities.List(ctx, params)
}
