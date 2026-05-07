package relationships

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/nhymxu/kith-pms/internal/audit"
)

// Sentinel errors returned by service methods.
var (
	ErrTypeInUse             = errors.New("relationships: type is in use and cannot be deleted")
	ErrDuplicateRelationship = errors.New("relationships: relationship already exists")
	ErrSelfRelationship      = errors.New("relationships: cannot relate a person to themselves")
	ErrTypeNotFound          = errors.New("relationships: relationship type not found")
	ErrNameEmpty             = errors.New("relationships: name must not be empty")
	ErrNameTooLong           = errors.New("relationships: name must be 80 characters or fewer")
)

// Service provides business logic for relationship types and person-relationship junctions.
type Service struct {
	Types         RelationshipTypeRepo
	Relationships PersonRelationshipRepo
	Audit         *audit.Service
	db            *sql.DB // needed for transactions
}

// NewService constructs a Service wired to db.
func NewService(db *sql.DB) *Service {
	return &Service{
		Types:         NewSQLRelationshipTypeRepo(db),
		Relationships: NewSQLPersonRelationshipRepo(db),
		db:            db,
	}
}

// ---- RelationshipType methods -----------------------------------------------

// CreateType validates and inserts a new relationship type.
// If reverseName is non-empty and differs from name, a paired inverse type is also created and both rows are linked.
func (s *Service) CreateType(ctx context.Context, name, reverseName string) (RelationshipType, error) {
	name = strings.TrimSpace(name)
	reverseName = strings.TrimSpace(reverseName)

	if err := validateTypeName(name); err != nil {
		return RelationshipType{}, err
	}

	id, err := s.Types.Create(ctx, name, reverseName)
	if err != nil {
		return RelationshipType{}, err
	}

	// Auto-create the inverse type when a reverse name is given and differs from the forward name.
	if reverseName != "" && reverseName != name {
		inverseID, err := s.Types.Create(ctx, reverseName, name)
		if err == nil {
			// Link both types to each other.
			_ = s.Types.SetInverseTypeID(ctx, id, &inverseID)
			_ = s.Types.SetInverseTypeID(ctx, inverseID, &id)
		}
		// If the inverse name already exists we simply skip linking — not a fatal error.
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityRelationshipType, id, name, audit.ActionCreate)
	}

	rt, err := s.Types.Get(ctx, id)
	if err != nil || rt == nil {
		return RelationshipType{ID: id, Name: name, ReverseName: reverseName}, err
	}

	return *rt, nil
}

// UpdateType updates the name and reverse_name of an existing type.
// Editing does not auto-propagate to the partner type — document this to users.
func (s *Service) UpdateType(ctx context.Context, id int64, name, reverseName string) error {
	name = strings.TrimSpace(name)
	reverseName = strings.TrimSpace(reverseName)

	if err := validateTypeName(name); err != nil {
		return err
	}

	if err := s.Types.Update(ctx, id, name, reverseName); err != nil {
		return err
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityRelationshipType, id, name, audit.ActionUpdate)
	}

	return nil
}

// DeleteType removes a relationship type.
// If a paired inverse type exists, its inverse_type_id pointer is cleared first.
// Returns ErrTypeInUse when junction rows reference this type.
func (s *Service) DeleteType(ctx context.Context, id int64) error {
	rt, err := s.Types.Get(ctx, id)
	if err != nil {
		return err
	}

	var name string
	if rt != nil {
		name = rt.Name
		// Clear the partner's back-pointer before deleting so no dangling FK issue.
		if rt.InverseTypeID != nil {
			_ = s.Types.SetInverseTypeID(ctx, *rt.InverseTypeID, nil)
		}
	}

	if err := s.Types.Delete(ctx, id); err != nil {
		if isFKConstraintErr(err) {
			return ErrTypeInUse
		}

		return err
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityRelationshipType, id, name, audit.ActionDelete)
	}

	return nil
}

// ListTypes returns all relationship types ordered by name.
func (s *Service) ListTypes(ctx context.Context) ([]RelationshipType, error) {
	return s.Types.List(ctx)
}

// ListTypesWithCounts returns all relationship types with their junction row counts.
func (s *Service) ListTypesWithCounts(ctx context.Context) ([]RelationshipType, error) {
	return s.Types.ListWithCounts(ctx)
}

// GetType returns a single relationship type by ID.
func (s *Service) GetType(ctx context.Context, id int64) (*RelationshipType, error) {
	return s.Types.Get(ctx, id)
}

// ---- PersonRelationship methods ---------------------------------------------

// AttachRelationship creates a person-relationship junction row.
// When the chosen type has an InverseTypeID, the inverse row is also written in the same transaction.
func (s *Service) AttachRelationship(ctx context.Context, fromID, toID, typeID int64, notes string) (int64, error) {
	if fromID == toID {
		return 0, ErrSelfRelationship
	}

	if len(notes) > 1000 {
		notes = notes[:1000]
	}

	rt, err := s.Types.Get(ctx, typeID)
	if err != nil {
		return 0, err
	}

	if rt == nil {
		return 0, ErrTypeNotFound
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	txRepo := &sqlPersonRelationshipRepo{db: tx}

	fwdID, err := txRepo.Attach(ctx, fromID, toID, typeID, notes)
	if err != nil {
		if isUniqueErr(err) {
			return 0, ErrDuplicateRelationship
		}

		return 0, err
	}

	if rt.InverseTypeID != nil {
		// Asymmetric paired type (e.g. Manager / Reports-to): insert inverse with the partner type.
		_, _ = tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO person_relationship (from_person_id, to_person_id, relationship_type_id, notes)
			 VALUES (?, ?, ?, ?)`,
			toID, fromID, *rt.InverseTypeID, notes,
		)
	} else if rt.ReverseName != "" {
		// Symmetric type (e.g. Friend/Friend): reverse name equals forward name so no separate
		// inverse type is created, but we still need the reciprocal row so both people see it.
		_, _ = tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO person_relationship (from_person_id, to_person_id, relationship_type_id, notes)
			 VALUES (?, ?, ?, ?)`,
			toID, fromID, typeID, notes,
		)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPersonRelationship, fwdID, "", audit.ActionCreate)
	}

	return fwdID, nil
}

// DetachRelationship deletes a junction row and, if the type has a paired inverse, deletes the inverse row too.
func (s *Service) DetachRelationship(ctx context.Context, id int64) error {
	row, err := s.Relationships.Get(ctx, id)
	if err != nil {
		return err
	}

	if row == nil {
		return nil // already gone
	}

	rt, _ := s.Types.Get(ctx, row.RelationshipTypeID)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	txRepo := &sqlPersonRelationshipRepo{db: tx}

	if err := txRepo.Detach(ctx, id); err != nil {
		return err
	}

	// Delete paired inverse row if it exists.
	if rt != nil && rt.InverseTypeID != nil {
		inv, _ := txRepo.FindPair(ctx, row.ToPersonID, row.FromPersonID, *rt.InverseTypeID)
		if inv != nil {
			_ = txRepo.Detach(ctx, inv.ID)
		}
	} else if rt != nil && rt.ReverseName != "" {
		// Symmetric type: inverse row uses the same type ID.
		inv, _ := txRepo.FindPair(ctx, row.ToPersonID, row.FromPersonID, row.RelationshipTypeID)
		if inv != nil {
			_ = txRepo.Detach(ctx, inv.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPersonRelationship, id, "", audit.ActionDelete)
	}

	return nil
}

// ListByPerson returns the outgoing relationship views for a given person.
func (s *Service) ListByPerson(ctx context.Context, personID int64) ([]RelationshipView, error) {
	return s.Relationships.ListByPersonID(ctx, personID)
}

// ---- validation -------------------------------------------------------------

func validateTypeName(name string) error {
	if name == "" {
		return ErrNameEmpty
	}

	if len(name) > 80 {
		return ErrNameTooLong
	}

	return nil
}
