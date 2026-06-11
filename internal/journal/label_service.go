package journal

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/uptrace/bun"
)

var (
	ErrLabelNameEmpty    = errors.New("journal labels: name must not be empty")
	ErrLabelNameTooLong  = errors.New("journal labels: name must be 64 characters or fewer")
	ErrLabelInvalidColor = errors.New("journal labels: color must be a 6-digit hex string e.g. #a1b2c3")
	ErrLabelNameConflict = errors.New("journal labels: name already exists")
)

var reLabelColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// LabelService provides business logic for managing journal labels.
type LabelService struct {
	Labels LabelRepo
}

// NewLabelService constructs a LabelService wired to db.
func NewLabelService(db *bun.DB) *LabelService {
	return &LabelService{Labels: NewLabelRepo(db)}
}

func (s *LabelService) Create(ctx context.Context, name, color string) (int64, error) {
	if err := validateJournalLabel(name, color); err != nil {
		return 0, err
	}

	id, err := s.Labels.Create(ctx, name, color)
	if err != nil {
		if isJournalLabelUniqueErr(err) {
			return 0, ErrLabelNameConflict
		}

		return 0, err
	}

	return id, nil
}

func (s *LabelService) Update(ctx context.Context, id int64, name, color string) error {
	if err := validateJournalLabel(name, color); err != nil {
		return err
	}

	if err := s.Labels.Update(ctx, id, name, color); err != nil {
		if isJournalLabelUniqueErr(err) {
			return ErrLabelNameConflict
		}

		return err
	}

	return nil
}

func (s *LabelService) Delete(ctx context.Context, id int64) error {
	return s.Labels.Delete(ctx, id)
}

func (s *LabelService) Get(ctx context.Context, id int64) (*Label, error) {
	return s.Labels.Get(ctx, id)
}

func (s *LabelService) List(ctx context.Context) ([]Label, error) {
	return s.Labels.List(ctx)
}

func (s *LabelService) ListWithCounts(ctx context.Context) ([]Label, error) {
	return s.Labels.ListWithCounts(ctx)
}

// FindOrCreate returns the ID of the label with the given name, creating it if it does not exist.
// If the name already exists, the existing label's ID is returned and color is ignored.
func (s *LabelService) FindOrCreate(ctx context.Context, name, color string) (int64, error) {
	id, err := s.Create(ctx, name, color)
	if errors.Is(err, ErrLabelNameConflict) {
		existing, getErr := s.Labels.GetByName(ctx, name)
		if getErr != nil {
			return 0, getErr
		}

		if existing == nil {
			return 0, ErrLabelNameConflict
		}

		return existing.ID, nil
	}

	return id, err
}

// ---- validation -------------------------------------------------------------

func validateJournalLabel(name, color string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrLabelNameEmpty
	}

	if len(name) > 64 {
		return ErrLabelNameTooLong
	}

	if !reLabelColor.MatchString(color) {
		return ErrLabelInvalidColor
	}

	return nil
}

func isJournalLabelUniqueErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "UNIQUE constraint failed") ||
		strings.Contains(err.Error(), "unique constraint")
}
