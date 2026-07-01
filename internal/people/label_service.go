package people

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/audit"
)

var (
	ErrNameEmpty    = errors.New("people labels: name must not be empty")
	ErrNameTooLong  = errors.New("people labels: name must be 64 characters or fewer")
	ErrInvalidColor = errors.New("people labels: color must be a 6-digit hex string e.g. #a1b2c3")
	ErrNameConflict = errors.New("people labels: name already exists")
)

var reLabelColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// LabelService provides business logic for managing people labels and person↔label associations.
type LabelService struct {
	Labels       LabelRepo
	PersonLabels PersonLabelRepo
	Audit        *audit.Service // optional; nil = no audit logging
}

// NewLabelService constructs a LabelService wired to db.
func NewLabelService(db *bun.DB) *LabelService {
	return &LabelService{
		Labels:       NewLabelRepo(db),
		PersonLabels: NewPersonLabelRepo(db),
	}
}

func (s *LabelService) Create(ctx context.Context, name, color string) (int64, error) {
	if err := validateLabel(name, color); err != nil {
		return 0, err
	}

	id, err := s.Labels.Create(ctx, name, color)
	if err != nil {
		if isLabelUniqueErr(err) {
			return 0, ErrNameConflict
		}

		return 0, err
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityLabel, id, name, audit.ActionCreate)
	}

	return id, nil
}

func (s *LabelService) Update(ctx context.Context, id int64, name, color string) error {
	if err := validateLabel(name, color); err != nil {
		return err
	}

	if err := s.Labels.Update(ctx, id, name, color); err != nil {
		if isLabelUniqueErr(err) {
			return ErrNameConflict
		}

		return err
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityLabel, id, name, audit.ActionUpdate)
	}

	return nil
}

// Delete removes a label; people_label_assignment rows cascade automatically.
func (s *LabelService) Delete(ctx context.Context, id int64) error {
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

func (s *LabelService) List(ctx context.Context) ([]Label, error) {
	return s.Labels.List(ctx)
}

func (s *LabelService) ListWithCounts(ctx context.Context) ([]Label, error) {
	return s.Labels.ListWithCounts(ctx)
}

func (s *LabelService) Get(ctx context.Context, id int64) (*Label, error) {
	return s.Labels.Get(ctx, id)
}

func (s *LabelService) Attach(ctx context.Context, personID, labelID int64) error {
	return s.PersonLabels.Attach(ctx, personID, labelID)
}

func (s *LabelService) Detach(ctx context.Context, personID, labelID int64) error {
	return s.PersonLabels.Detach(ctx, personID, labelID)
}

func (s *LabelService) ListByPersonID(ctx context.Context, personID int64) ([]Label, error) {
	return s.PersonLabels.ListByPersonID(ctx, personID)
}

func (s *LabelService) ListByPersonIDs(ctx context.Context, personIDs []int64) (map[int64][]Label, error) {
	return s.Labels.ListByPersonIDs(ctx, personIDs)
}

// BulkAttach attaches labelID to each personID. Idempotent — skips existing rows.
// Returns count of net-new assignments.
func (s *LabelService) BulkAttach(ctx context.Context, labelID int64, personIDs []int64) (int, error) {
	return s.PersonLabels.BulkAttach(ctx, labelID, personIDs)
}

func (s *LabelService) ListPersonIDsByLabelID(ctx context.Context, labelID int64) ([]int64, error) {
	return s.Labels.ListPersonIDsByLabelID(ctx, labelID)
}

// ---- validation -------------------------------------------------------------

func validateLabel(name, color string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrNameEmpty
	}

	if len(name) > 64 {
		return ErrNameTooLong
	}

	if !reLabelColor.MatchString(color) {
		return ErrInvalidColor
	}

	return nil
}

func isLabelUniqueErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "UNIQUE constraint failed") ||
		strings.Contains(err.Error(), "unique constraint")
}

// LabelValidationError wraps a validation error with a user-visible message.
type LabelValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *LabelValidationError) Error() string {
	return fmt.Sprintf("people labels: %s: %s", e.Field, e.Message)
}

func (e *LabelValidationError) Unwrap() error { return e.Err }
