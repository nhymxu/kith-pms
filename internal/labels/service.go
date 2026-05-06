package labels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/nhymxu/kith-pms/internal/audit"
)

// Sentinel errors for validation failures — handlers can type-assert to re-render forms.
var (
	ErrNameEmpty    = errors.New("labels: name must not be empty")
	ErrNameTooLong  = errors.New("labels: name must be 64 characters or fewer")
	ErrInvalidColor = errors.New("labels: color must be a 6-digit hex string e.g. #a1b2c3")
	ErrNameConflict = errors.New("labels: name already exists")
)

var reColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Service provides business logic for managing labels and person↔label associations.
type Service struct {
	Labels       LabelRepo
	PersonLabels PersonLabelRepo
	Audit        *audit.Service // optional; nil = no audit logging
}

// NewService constructs a Service wired to db.
func NewService(db *sql.DB) *Service {
	return &Service{
		Labels:       NewLabelRepo(db),
		PersonLabels: NewPersonLabelRepo(db),
	}
}

// Create validates inputs and inserts a new label. Returns the new label ID.
func (s *Service) Create(ctx context.Context, name, color string) (int64, error) {
	if err := validate(name, color); err != nil {
		return 0, err
	}
	id, err := s.Labels.Create(ctx, name, color)
	if err != nil {
		if isUniqueErr(err) {
			return 0, ErrNameConflict
		}
		return 0, err
	}
	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityLabel, id, name, audit.ActionCreate)
	}
	return id, nil
}

// Update validates and updates an existing label's name and color.
func (s *Service) Update(ctx context.Context, id int64, name, color string) error {
	if err := validate(name, color); err != nil {
		return err
	}
	if err := s.Labels.Update(ctx, id, name, color); err != nil {
		if isUniqueErr(err) {
			return ErrNameConflict
		}
		return err
	}
	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityLabel, id, name, audit.ActionUpdate)
	}
	return nil
}

// Delete removes a label; person_label rows cascade automatically.
func (s *Service) Delete(ctx context.Context, id int64) error {
	var name string
	if s.Audit != nil {
		if l, err := s.Labels.Get(ctx, id); err == nil && l != nil {
			name = l.Name
		}
	}
	if err := s.Labels.Delete(ctx, id); err != nil {
		return err
	}
	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityLabel, id, name, audit.ActionDelete)
	}
	return nil
}

// List returns all labels ordered by name (no usage counts).
func (s *Service) List(ctx context.Context) ([]Label, error) {
	return s.Labels.List(ctx)
}

// ListWithCounts returns all labels with their person-attachment counts.
func (s *Service) ListWithCounts(ctx context.Context) ([]Label, error) {
	return s.Labels.ListWithCounts(ctx)
}

// Get returns a single label by ID, or nil if not found.
func (s *Service) Get(ctx context.Context, id int64) (*Label, error) {
	return s.Labels.Get(ctx, id)
}

// Attach associates a label with a person (idempotent).
func (s *Service) Attach(ctx context.Context, personID, labelID int64) error {
	return s.PersonLabels.Attach(ctx, personID, labelID)
}

// Detach removes the association between a label and a person.
func (s *Service) Detach(ctx context.Context, personID, labelID int64) error {
	return s.PersonLabels.Detach(ctx, personID, labelID)
}

// ListByPersonID returns all labels attached to the given person.
func (s *Service) ListByPersonID(ctx context.Context, personID int64) ([]Label, error) {
	return s.PersonLabels.ListByPersonID(ctx, personID)
}

// ListByPersonIDs returns a map of person IDs to their labels for batch loading.
func (s *Service) ListByPersonIDs(ctx context.Context, personIDs []int64) (map[int64][]Label, error) {
	return s.Labels.ListByPersonIDs(ctx, personIDs)
}

// ---- validation -------------------------------------------------------------

func validate(name, color string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrNameEmpty
	}
	if len(name) > 64 {
		return ErrNameTooLong
	}
	if !reColor.MatchString(color) {
		return ErrInvalidColor
	}
	return nil
}

// isUniqueErr returns true when err is a SQLite UNIQUE constraint violation.
func isUniqueErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed") ||
		strings.Contains(err.Error(), "unique constraint")
}

// ValidationError wraps a validation error with a user-visible message.
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("labels: %s: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error { return e.Err }
