package relationships

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
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

type sqlRelationshipTypeRepo struct{ db *sql.DB }

func NewSQLRelationshipTypeRepo(db *sql.DB) RelationshipTypeRepo {
	return &sqlRelationshipTypeRepo{db: db}
}

func (r *sqlRelationshipTypeRepo) Create(ctx context.Context, name, reverseName string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO relationship_type (name, reverse_name) VALUES (?, ?)`,
		name, reverseName,
	)
	if err != nil {
		return 0, fmt.Errorf("relationships: create type: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("relationships: create type last id: %w", err)
	}
	return id, nil
}

func (r *sqlRelationshipTypeRepo) Update(ctx context.Context, id int64, name, reverseName string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE relationship_type SET name = ?, reverse_name = ? WHERE id = ?`,
		name, reverseName, id,
	)
	if err != nil {
		return fmt.Errorf("relationships: update type: %w", err)
	}
	return nil
}

func (r *sqlRelationshipTypeRepo) SetInverseTypeID(ctx context.Context, id int64, inverseID *int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE relationship_type SET inverse_type_id = ? WHERE id = ?`,
		inverseID, id,
	)
	if err != nil {
		return fmt.Errorf("relationships: set inverse type: %w", err)
	}
	return nil
}

func (r *sqlRelationshipTypeRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM relationship_type WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("relationships: delete type: %w", err)
	}
	return nil
}

func (r *sqlRelationshipTypeRepo) Get(ctx context.Context, id int64) (*RelationshipType, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, reverse_name, inverse_type_id, created_at FROM relationship_type WHERE id = ?`, id,
	)
	t, err := scanRelationshipType(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *sqlRelationshipTypeRepo) List(ctx context.Context) ([]RelationshipType, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, reverse_name, inverse_type_id, created_at FROM relationship_type ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("relationships: list types: %w", err)
	}
	defer rows.Close()
	return collectRelationshipTypes(rows)
}

func (r *sqlRelationshipTypeRepo) ListWithCounts(ctx context.Context) ([]RelationshipType, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT rt.id, rt.name, rt.reverse_name, rt.inverse_type_id, rt.created_at,
		        COUNT(pr.id) AS usage_count
		 FROM relationship_type rt
		 LEFT JOIN person_relationship pr ON pr.relationship_type_id = rt.id
		 GROUP BY rt.id
		 ORDER BY rt.name`,
	)
	if err != nil {
		return nil, fmt.Errorf("relationships: list types with counts: %w", err)
	}
	defer rows.Close()

	var types []RelationshipType
	for rows.Next() {
		var rt RelationshipType
		var createdAt string
		var inverseID sql.NullInt64
		if err := rows.Scan(&rt.ID, &rt.Name, &rt.ReverseName, &inverseID, &createdAt, &rt.UsageCount); err != nil {
			return nil, fmt.Errorf("relationships: scan type with count: %w", err)
		}
		rt.CreatedAt, _ = parseTime(createdAt)
		if inverseID.Valid {
			v := inverseID.Int64
			rt.InverseTypeID = &v
		}
		types = append(types, rt)
	}
	return types, rows.Err()
}

// ---- PersonRelationshipRepo -------------------------------------------------

type PersonRelationshipRepo interface {
	Attach(ctx context.Context, fromID, toID, typeID int64, notes string) (int64, error)
	Detach(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*PersonRelationship, error)
	FindPair(ctx context.Context, fromID, toID, typeID int64) (*PersonRelationship, error)
	ListByPersonID(ctx context.Context, personID int64) ([]RelationshipView, error)
}

// querier is satisfied by both *sql.DB and *sql.Tx, enabling transactional repo use.
type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type sqlPersonRelationshipRepo struct{ db querier }

func NewSQLPersonRelationshipRepo(db *sql.DB) PersonRelationshipRepo {
	return &sqlPersonRelationshipRepo{db: db}
}

func (r *sqlPersonRelationshipRepo) Attach(ctx context.Context, fromID, toID, typeID int64, notes string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO person_relationship (from_person_id, to_person_id, relationship_type_id, notes)
		 VALUES (?, ?, ?, ?)`,
		fromID, toID, typeID, notes,
	)
	if err != nil {
		return 0, fmt.Errorf("relationships: attach: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("relationships: attach last id: %w", err)
	}
	return id, nil
}

func (r *sqlPersonRelationshipRepo) Detach(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM person_relationship WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("relationships: detach: %w", err)
	}
	return nil
}

func (r *sqlPersonRelationshipRepo) Get(ctx context.Context, id int64) (*PersonRelationship, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, from_person_id, to_person_id, relationship_type_id, notes, created_at
		 FROM person_relationship WHERE id = ?`, id,
	)
	pr, err := scanPersonRelationship(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &pr, nil
}

func (r *sqlPersonRelationshipRepo) FindPair(ctx context.Context, fromID, toID, typeID int64) (*PersonRelationship, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, from_person_id, to_person_id, relationship_type_id, notes, created_at
		 FROM person_relationship
		 WHERE from_person_id = ? AND to_person_id = ? AND relationship_type_id = ?`,
		fromID, toID, typeID,
	)
	pr, err := scanPersonRelationship(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &pr, nil
}

func (r *sqlPersonRelationshipRepo) ListByPersonID(ctx context.Context, personID int64) ([]RelationshipView, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT pr.id, pr.to_person_id, p.name, COALESCE(p.avatar_path, ''), t.name, pr.notes
		 FROM person_relationship pr
		 JOIN person p            ON p.id = pr.to_person_id
		 JOIN relationship_type t ON t.id = pr.relationship_type_id
		 WHERE pr.from_person_id = ?
		 ORDER BY t.name, p.name`,
		personID,
	)
	if err != nil {
		return nil, fmt.Errorf("relationships: list by person: %w", err)
	}
	defer rows.Close()

	var views []RelationshipView
	for rows.Next() {
		var v RelationshipView
		if err := rows.Scan(&v.ID, &v.OtherPersonID, &v.OtherPersonName, &v.OtherPersonAvatar, &v.TypeName, &v.Notes); err != nil {
			return nil, fmt.Errorf("relationships: scan view: %w", err)
		}
		views = append(views, v)
	}
	return views, rows.Err()
}

// ---- scan helpers -----------------------------------------------------------

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRelationshipType(row rowScanner) (RelationshipType, error) {
	var rt RelationshipType
	var createdAt string
	var inverseID sql.NullInt64
	if err := row.Scan(&rt.ID, &rt.Name, &rt.ReverseName, &inverseID, &createdAt); err != nil {
		return RelationshipType{}, fmt.Errorf("relationships: scan type: %w", err)
	}
	rt.CreatedAt, _ = parseTime(createdAt)
	if inverseID.Valid {
		v := inverseID.Int64
		rt.InverseTypeID = &v
	}
	return rt, nil
}

func collectRelationshipTypes(rows *sql.Rows) ([]RelationshipType, error) {
	var types []RelationshipType
	for rows.Next() {
		rt, err := scanRelationshipType(rows)
		if err != nil {
			return nil, err
		}
		types = append(types, rt)
	}
	return types, rows.Err()
}

func scanPersonRelationship(row rowScanner) (PersonRelationship, error) {
	var pr PersonRelationship
	var createdAt string
	if err := row.Scan(&pr.ID, &pr.FromPersonID, &pr.ToPersonID, &pr.RelationshipTypeID, &pr.Notes, &createdAt); err != nil {
		return PersonRelationship{}, fmt.Errorf("relationships: scan junction: %w", err)
	}
	pr.CreatedAt, _ = parseTime(createdAt)
	return pr, nil
}

func parseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
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
