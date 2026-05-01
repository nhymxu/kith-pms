package people

import (
	"context"
	"database/sql"
	"fmt"
)

const defaultPageSize = 50

// ListParams holds query parameters for listing people.
type ListParams struct {
	Query    string
	Page     int
	PageSize int
	LabelIDs []int64 // AND-semantics: person must have ALL listed labels
}

// Service provides business logic for managing people.
type Service struct {
	DB       *sql.DB
	People   PersonRepo
	Contacts ContactRepo
	Locations LocationRepo
}

// NewService constructs a Service wired to db.
func NewService(db *sql.DB) *Service {
	return &Service{
		DB:        db,
		People:    NewPersonRepo(db),
		Contacts:  NewContactRepo(db),
		Locations: NewLocationRepo(db),
	}
}

// Create inserts a new person with their contacts and locations in a single transaction.
// Returns the new person ID.
func (s *Service) Create(ctx context.Context, p Person, contacts []ContactInfo, locations []Location) (int64, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("people: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	id, err := s.People.Create(ctx, tx, p)
	if err != nil {
		return 0, err
	}

	if err := s.Contacts.ReplaceAll(ctx, tx, id, contacts); err != nil {
		return 0, err
	}
	if err := s.Locations.ReplaceAll(ctx, tx, id, locations); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("people: commit create: %w", err)
	}
	return id, nil
}

// Update replaces a person's fields and all child rows in a single transaction.
func (s *Service) Update(ctx context.Context, p Person, contacts []ContactInfo, locations []Location) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("people: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.People.Update(ctx, tx, p); err != nil {
		return err
	}
	if err := s.Contacts.ReplaceAll(ctx, tx, p.ID, contacts); err != nil {
		return err
	}
	if err := s.Locations.ReplaceAll(ctx, tx, p.ID, locations); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("people: commit update: %w", err)
	}
	return nil
}

// Get returns a person by ID with their contacts and locations populated.
// Returns nil, nil when not found.
func (s *Service) Get(ctx context.Context, id int64) (*Person, error) {
	p, err := s.People.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}

	contacts, err := s.Contacts.ListByPerson(ctx, id)
	if err != nil {
		return nil, err
	}
	locations, err := s.Locations.ListByPerson(ctx, id)
	if err != nil {
		return nil, err
	}

	p.Contacts = contacts
	p.Locations = locations
	return p, nil
}

// List returns a paginated, optionally filtered list of people (no child rows).
func (s *Service) List(ctx context.Context, params ListParams) ([]Person, error) {
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > 500 {
		pageSize = 500
	}
	page := params.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	return s.People.List(ctx, params.Query, params.LabelIDs, pageSize, offset)
}

// Delete removes a person and cascades to their contacts and locations.
func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.People.Delete(ctx, id)
}
