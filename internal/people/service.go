package people

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/nhymxu/kith-pms/internal/audit"
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
	DB          *sql.DB
	People      PersonRepo
	Contacts    ContactRepo
	Locations   LocationRepo
	FileService FileService
	Audit       *audit.Service // optional; nil = no audit logging
}

type FileService interface {
	SaveAvatar(personID int64, file multipart.File, header *multipart.FileHeader) (path string, err error)
	DeleteAvatar(personID int64, path string) error
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
	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPerson, id, p.Name, audit.ActionCreate)
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
	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPerson, p.ID, p.Name, audit.ActionUpdate)
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
	var name string
	if s.Audit != nil {
		if p, err := s.People.Get(ctx, id); err == nil && p != nil {
			name = p.Name
		}
	}
	if err := s.People.Delete(ctx, id); err != nil {
		return err
	}
	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPerson, id, name, audit.ActionDelete)
	}
	return nil
}

func (s *Service) GetSelf(ctx context.Context) (*Person, error) {
	return s.People.GetSelf(ctx)
}

func (s *Service) SetSelf(ctx context.Context, personID int64) error {
	person, err := s.People.Get(ctx, personID)
	if err != nil {
		return err
	}
	if person == nil {
		return fmt.Errorf("person not found")
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("people: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.People.ClearSelf(ctx, tx); err != nil {
		return err
	}
	if err := s.People.SetSelf(ctx, tx, personID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("people: commit set self: %w", err)
	}
	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPerson, personID, person.Name, audit.ActionUpdate)
	}
	return nil
}

// UploadAvatar saves a new avatar file and updates the person's avatar metadata.
// If the person already has an avatar, the old file is deleted after the transaction commits.
func (s *Service) UploadAvatar(ctx context.Context, personID int64, file multipart.File, header *multipart.FileHeader) error {
	if s.FileService == nil {
		return fmt.Errorf("file service not configured")
	}

	person, err := s.People.Get(ctx, personID)
	if err != nil {
		return fmt.Errorf("get person: %w", err)
	}
	if person == nil {
		return fmt.Errorf("person not found")
	}

	oldAvatarPath := person.AvatarPath

	path, err := s.FileService.SaveAvatar(personID, file, header)
	if err != nil {
		return fmt.Errorf("save avatar file: %w", err)
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		s.FileService.DeleteAvatar(personID, path)
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	mimeType := header.Header.Get("Content-Type")
	size := header.Size
	uploadedAt := time.Now().UTC()

	if err := s.People.UpdateAvatar(ctx, tx, personID, path, mimeType, size, uploadedAt); err != nil {
		s.FileService.DeleteAvatar(personID, path)
		return err
	}

	if err := tx.Commit(); err != nil {
		s.FileService.DeleteAvatar(personID, path)
		return fmt.Errorf("commit: %w", err)
	}

	if oldAvatarPath != "" {
		_ = s.FileService.DeleteAvatar(personID, oldAvatarPath)
	}
	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPerson, personID, person.Name, audit.ActionUpdate)
	}
	return nil
}

// DeleteAvatar removes the avatar file and clears the person's avatar metadata.
func (s *Service) DeleteAvatar(ctx context.Context, personID int64) error {
	if s.FileService == nil {
		return fmt.Errorf("file service not configured")
	}

	person, err := s.People.Get(ctx, personID)
	if err != nil {
		return fmt.Errorf("get person: %w", err)
	}
	if person == nil {
		return fmt.Errorf("person not found")
	}

	avatarPath := person.AvatarPath

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.People.ClearAvatar(ctx, tx, personID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if avatarPath != "" {
		_ = s.FileService.DeleteAvatar(personID, avatarPath)
	}
	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPerson, personID, person.Name, audit.ActionUpdate)
	}
	return nil
}
