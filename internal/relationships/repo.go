package relationships

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/uptrace/bun"
)

// ---- RelationshipTypeRepo ---------------------------------------------------

type RelationshipTypeRepo interface {
	Create(ctx context.Context, name, reverseName string) (int64, error)
	Update(ctx context.Context, id int64, name, reverseName string) error
	SetInverseTypeID(ctx context.Context, id int64, inverseID *int64) error
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*RelationshipType, error)
	List(ctx context.Context) ([]RelationshipType, error)
	ListWithCounts(ctx context.Context) ([]RelationshipType, error)
}

type sqlRelationshipTypeRepo struct{ db *bun.DB }

func NewSQLRelationshipTypeRepo(db *bun.DB) RelationshipTypeRepo {
	return &sqlRelationshipTypeRepo{db: db}
}

func (r *sqlRelationshipTypeRepo) Create(ctx context.Context, name, reverseName string) (int64, error) {
	rt := &RelationshipType{Name: name, ReverseName: reverseName}

	_, err := r.db.NewInsert().Model(rt).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("relationships: create type: %w", err)
	}

	return rt.ID, nil
}

func (r *sqlRelationshipTypeRepo) Update(ctx context.Context, id int64, name, reverseName string) error {
	_, err := r.db.NewUpdate().Model((*RelationshipType)(nil)).
		Set("name = ?", name).
		Set("reverse_name = ?", reverseName).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("relationships: update type: %w", err)
	}

	return nil
}

func (r *sqlRelationshipTypeRepo) SetInverseTypeID(ctx context.Context, id int64, inverseID *int64) error {
	_, err := r.db.NewUpdate().Model((*RelationshipType)(nil)).
		Set("inverse_type_id = ?", inverseID).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("relationships: set inverse type: %w", err)
	}

	return nil
}

func (r *sqlRelationshipTypeRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*RelationshipType)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("relationships: delete type: %w", err)
	}

	return nil
}

func (r *sqlRelationshipTypeRepo) Get(ctx context.Context, id int64) (*RelationshipType, error) {
	var rt RelationshipType

	err := r.db.NewSelect().Model(&rt).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("relationships: get type: %w", err)
	}

	return &rt, nil
}

func (r *sqlRelationshipTypeRepo) List(ctx context.Context) ([]RelationshipType, error) {
	var types []RelationshipType

	err := r.db.NewSelect().Model(&types).OrderExpr("name").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("relationships: list types: %w", err)
	}

	return types, nil
}

func (r *sqlRelationshipTypeRepo) ListWithCounts(ctx context.Context) ([]RelationshipType, error) {
	var rows []struct {
		RelationshipType
		Count int `bun:"count"`
	}

	err := r.db.NewSelect().
		TableExpr("relationship_type rt").
		ColumnExpr("rt.*, COUNT(pr.id) AS count").
		Join("LEFT JOIN person_relationship pr ON pr.relationship_type_id = rt.id").
		GroupExpr("rt.id").
		OrderExpr("rt.name").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("relationships: list types with counts: %w", err)
	}

	types := make([]RelationshipType, 0, len(rows))
	for _, row := range rows {
		rt := row.RelationshipType
		rt.UsageCount = row.Count
		types = append(types, rt)
	}

	return types, nil
}

// ---- PersonRelationshipRepo -------------------------------------------------

type PersonRelationshipRepo interface {
	Attach(ctx context.Context, fromID, toID, typeID int64, notes string) (int64, error)
	Detach(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*PersonRelationship, error)
	FindPair(ctx context.Context, fromID, toID, typeID int64) (*PersonRelationship, error)
	ListByPersonID(ctx context.Context, personID int64) ([]RelationshipView, error)
}

// sqlPersonRelationshipRepo uses bun.IDB so it works with both *bun.DB and bun.Tx.
// The service constructs it with a bun.Tx for transactional operations.
type sqlPersonRelationshipRepo struct{ db bun.IDB }

func NewSQLPersonRelationshipRepo(db *bun.DB) PersonRelationshipRepo {
	return &sqlPersonRelationshipRepo{db: db}
}

func (r *sqlPersonRelationshipRepo) Attach(
	ctx context.Context,
	fromID, toID, typeID int64,
	notes string,
) (int64, error) {
	pr := &PersonRelationship{
		FromPersonID:       fromID,
		ToPersonID:         toID,
		RelationshipTypeID: typeID,
		Notes:              notes,
	}

	_, err := r.db.NewInsert().Model(pr).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("relationships: attach: %w", err)
	}

	return pr.ID, nil
}

func (r *sqlPersonRelationshipRepo) Detach(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*PersonRelationship)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("relationships: detach: %w", err)
	}

	return nil
}

func (r *sqlPersonRelationshipRepo) Get(ctx context.Context, id int64) (*PersonRelationship, error) {
	var pr PersonRelationship

	err := r.db.NewSelect().Model(&pr).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("relationships: get: %w", err)
	}

	return &pr, nil
}

func (r *sqlPersonRelationshipRepo) FindPair(
	ctx context.Context,
	fromID, toID, typeID int64,
) (*PersonRelationship, error) {
	var pr PersonRelationship

	err := r.db.NewSelect().Model(&pr).
		Where("from_person_id = ? AND to_person_id = ? AND relationship_type_id = ?", fromID, toID, typeID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("relationships: find pair: %w", err)
	}

	return &pr, nil
}

func (r *sqlPersonRelationshipRepo) ListByPersonID(ctx context.Context, personID int64) ([]RelationshipView, error) {
	var rows []struct {
		ID                int64  `bun:"id"`
		OtherPersonID     int64  `bun:"other_person_id"`
		OtherPersonName   string `bun:"other_person_name"`
		OtherPersonAvatar string `bun:"other_person_avatar"`
		TypeName          string `bun:"type_name"`
		Notes             string `bun:"notes"`
	}

	err := r.db.NewSelect().
		TableExpr("person_relationship pr").
		ColumnExpr("pr.id, pr.to_person_id AS other_person_id, p.name AS other_person_name, COALESCE(p.avatar_path, '') AS other_person_avatar, t.name AS type_name, pr.notes").
		Join("JOIN person p ON p.id = pr.to_person_id").
		Join("JOIN relationship_type t ON t.id = pr.relationship_type_id").
		Where("pr.from_person_id = ?", personID).
		OrderExpr("t.name, p.name").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("relationships: list by person: %w", err)
	}

	views := make([]RelationshipView, 0, len(rows))
	for _, row := range rows {
		views = append(views, RelationshipView{
			ID:                row.ID,
			OtherPersonID:     row.OtherPersonID,
			OtherPersonName:   row.OtherPersonName,
			OtherPersonAvatar: row.OtherPersonAvatar,
			TypeName:          row.TypeName,
			Notes:             row.Notes,
		})
	}

	return views, nil
}

// isFKConstraintErr returns true for SQLite FOREIGN KEY constraint violations.
func isFKConstraintErr(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	return strings.Contains(msg, "FOREIGN KEY constraint failed") ||
		strings.Contains(msg, "foreign key constraint")
}

// isUniqueErr returns true for SQLite UNIQUE constraint violations.
func isUniqueErr(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "unique constraint")
}
