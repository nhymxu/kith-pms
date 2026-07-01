package relationships

import (
	"context"
	"errors"
	"strings"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/audit"
	"github.com/nhymxu/kith-pms/internal/people"
)

// Sentinel errors returned by service methods.
var (
	ErrTypeInUse                = errors.New("relationships: type is in use and cannot be deleted")
	ErrDuplicateRelationship    = errors.New("relationships: relationship already exists")
	ErrSelfRelationship         = errors.New("relationships: cannot relate a person to themselves")
	ErrTypeNotFound             = errors.New("relationships: relationship type not found")
	ErrNameEmpty                = errors.New("relationships: name must not be empty")
	ErrNameTooLong              = errors.New("relationships: name must be 80 characters or fewer")
	ErrAsymmetricTypeNotAllowed = errors.New("relationships: asymmetric types cannot be used for mesh connect")
	ErrMeshTooLarge             = errors.New("relationships: label has too many members for mesh connect (max 500)")
)

// Service provides business logic for relationship types and person-relationship junctions.
type Service struct {
	Types         RelationshipTypeRepo
	Relationships PersonRelationshipRepo
	PeopleLabels  people.LabelRepo
	Audit         *audit.Service
	db            *bun.DB // needed for transactions
	graphEdges    GraphEdgesRepo
}

// NewService constructs a Service wired to db.
func NewService(db *bun.DB) *Service {
	return &Service{
		Types:         NewSQLRelationshipTypeRepo(db),
		Relationships: NewSQLPersonRelationshipRepo(db),
		PeopleLabels:  people.NewLabelRepo(db),
		db:            db,
		graphEdges:    newSQLGraphEdgesRepo(db),
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

	notes = truncateNotes(notes)

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
	} else {
		// Symmetric type (e.g. Friend/Friend, or any type with no distinct inverse):
		// no separate inverse type exists, but we still need the reciprocal row so
		// both people see it. reverse_name is a display label only — it must not
		// gate whether the reciprocal row is created.
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
		s.Audit.Log(ctx, audit.EntityPersonRelationship, fwdID, rt.Name, audit.ActionCreate,
			audit.Metadata{
				DetailAction: "attach",
				Changes: []audit.Change{
					{Field: "from_person_id", Old: nil, New: fromID},
					{Field: "to_person_id", Old: nil, New: toID},
					{Field: "relationship_type", Old: nil, New: rt.Name},
				},
			})
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
	} else if rt != nil {
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
		var typeName string
		if rt != nil {
			typeName = rt.Name
		}

		s.Audit.Log(ctx, audit.EntityPersonRelationship, id, typeName, audit.ActionDelete,
			audit.Metadata{
				DetailAction: "detach",
				Changes: []audit.Change{
					{Field: "from_person_id", Old: row.FromPersonID, New: nil},
					{Field: "to_person_id", Old: row.ToPersonID, New: nil},
					{Field: "relationship_type", Old: typeName, New: nil},
				},
			})
	}

	return nil
}

// ListByPerson returns the outgoing relationship views for a given person.
func (s *Service) ListByPerson(ctx context.Context, personID int64) ([]RelationshipView, error) {
	return s.Relationships.ListByPersonID(ctx, personID)
}

// BulkRelationshipPair is a single from→to relationship to create in a bulk operation.
type BulkRelationshipPair struct {
	ToPersonID int64
	TypeID     int64
	Notes      string
}

// BulkAttach creates relationships from fromID to each pair in a single transaction.
// Skips duplicates (ON CONFLICT DO NOTHING). Handles inverse rows per type config.
func (s *Service) BulkAttach(
	ctx context.Context,
	fromID int64,
	pairs []BulkRelationshipPair,
) (created, skipped int, err error) {
	if len(pairs) == 0 {
		return 0, 0, nil
	}

	// Prefetch all unique type IDs before opening the transaction.
	// With MaxOpenConns=1, fetching inside a transaction would deadlock because
	// the transaction holds the only connection.
	typeCache := make(map[int64]*RelationshipType)
	for _, p := range pairs {
		if _, ok := typeCache[p.TypeID]; ok {
			continue
		}

		rt, err := s.Types.Get(ctx, p.TypeID)
		if err != nil {
			return 0, 0, err
		}

		if rt == nil {
			return 0, 0, ErrTypeNotFound
		}

		typeCache[p.TypeID] = rt
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	txRepo := &sqlPersonRelationshipRepo{db: tx}

	for _, p := range pairs {
		if p.ToPersonID == fromID {
			return 0, 0, ErrSelfRelationship
		}

		rt := typeCache[p.TypeID]

		p.Notes = truncateNotes(p.Notes)

		_, err = txRepo.Attach(ctx, fromID, p.ToPersonID, p.TypeID, p.Notes)
		if err != nil {
			if isUniqueErr(err) {
				skipped++
				continue
			}

			return 0, 0, err
		}

		created++

		const insertRel = `INSERT OR IGNORE INTO person_relationship` +
			` (from_person_id, to_person_id, relationship_type_id, notes) VALUES (?,?,?,?)`

		if rt.InverseTypeID != nil {
			_, _ = tx.ExecContext(ctx, insertRel, p.ToPersonID, fromID, *rt.InverseTypeID, p.Notes)
		} else {
			_, _ = tx.ExecContext(ctx, insertRel, p.ToPersonID, fromID, p.TypeID, p.Notes)
		}
	}

	return created, skipped, tx.Commit()
}

// BulkAttachMesh creates symmetric relationships between all people sharing labelID.
// Only symmetric types (InverseTypeID == nil) are allowed.
// Inserts all directed pairs (A→B and B→A) with INSERT OR IGNORE.
// Returns created (net-new rows), skipped, totalMembers.
func (s *Service) BulkAttachMesh(
	ctx context.Context,
	labelID, typeID int64,
) (created, skipped, totalMembers int, err error) {
	rt, err := s.Types.Get(ctx, typeID)
	if err != nil || rt == nil {
		return 0, 0, 0, ErrTypeNotFound
	}

	if rt.InverseTypeID != nil {
		return 0, 0, 0, ErrAsymmetricTypeNotAllowed
	}

	ids, err := s.PeopleLabels.ListPersonIDsByLabelID(ctx, labelID)
	if err != nil {
		return 0, 0, 0, err
	}

	totalMembers = len(ids)
	if totalMembers < 2 {
		return 0, 0, totalMembers, nil
	}

	if totalMembers > 500 {
		return 0, 0, totalMembers, ErrMeshTooLarge
	}

	type pair struct{ from, to int64 }

	pairs := make([]pair, 0, totalMembers*(totalMembers-1))

	for i := range ids {
		for j := range ids {
			if i != j {
				pairs = append(pairs, pair{ids[i], ids[j]})
			}
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, totalMembers, err
	}
	defer func() { _ = tx.Rollback() }()

	const chunkSize = 200

	var totalAffected int64

	for start := 0; start < len(pairs); start += chunkSize {
		end := min(start+chunkSize, len(pairs))

		chunk := pairs[start:end]
		args := make([]any, 0, len(chunk)*3)
		placeholders := make([]string, len(chunk))

		for k, p := range chunk {
			placeholders[k] = "(?,?,?)"

			args = append(args, p.from, p.to, typeID)
		}

		q := "INSERT OR IGNORE INTO person_relationship (from_person_id, to_person_id, relationship_type_id) VALUES " +
			strings.Join(placeholders, ",")

		res, err := tx.ExecContext(ctx, q, args...)
		if err != nil {
			return 0, 0, totalMembers, err
		}

		n, _ := res.RowsAffected()
		totalAffected += n
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, totalMembers, err
	}

	created = int(totalAffected)
	skipped = len(pairs) - created

	return created, skipped, totalMembers, nil
}

// ---- helpers ----------------------------------------------------------------

// truncateNotes caps notes at 1 000 Unicode code points to avoid splitting multibyte runes.
func truncateNotes(s string) string {
	const maxCodepoints = 1000
	if len(s) <= maxCodepoints {
		return s
	}

	runes := []rune(s)
	if len(runes) > maxCodepoints {
		runes = runes[:maxCodepoints]
	}

	return string(runes)
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
