package people

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// PersonRepo defines persistence operations for Person records.
type PersonRepo interface {
	List(ctx context.Context, q string, labelIDs []int64, limit, offset int) ([]Person, error)
	Get(ctx context.Context, id int64) (*Person, error)
	GetSelf(ctx context.Context) (*Person, error)
	Create(ctx context.Context, tx *sql.Tx, p Person) (int64, error)
	Update(ctx context.Context, tx *sql.Tx, p Person) error
	Delete(ctx context.Context, id int64) error
	SetSelf(ctx context.Context, tx *sql.Tx, personID int64) error
	ClearSelf(ctx context.Context, tx *sql.Tx) error
	UpdateAvatar(ctx context.Context, tx *sql.Tx, personID int64, path, mimeType string, size int64, uploadedAt time.Time) error
	ClearAvatar(ctx context.Context, tx *sql.Tx, personID int64) error
}

// ContactRepo defines persistence operations for ContactInfo records.
type ContactRepo interface {
	// ReplaceAll deletes all contacts for personID and inserts the new slice.
	ReplaceAll(ctx context.Context, tx *sql.Tx, personID int64, contacts []ContactInfo) error
	ListByPerson(ctx context.Context, personID int64) ([]ContactInfo, error)
}

// LocationRepo defines persistence operations for Location records.
type LocationRepo interface {
	// ReplaceAll deletes all locations for personID and inserts the new slice.
	ReplaceAll(ctx context.Context, tx *sql.Tx, personID int64, locations []Location) error
	ListByPerson(ctx context.Context, personID int64) ([]Location, error)
}

// ---- sqlPersonRepo ----------------------------------------------------------

type sqlPersonRepo struct{ db *sql.DB }

// NewPersonRepo returns a PersonRepo backed by db.
func NewPersonRepo(db *sql.DB) PersonRepo { return &sqlPersonRepo{db: db} }

func (r *sqlPersonRepo) List(ctx context.Context, q string, labelIDs []int64, limit, offset int) ([]Person, error) {
	// Build WHERE clause and args dynamically.
	var where []string
	var args []any

	if q != "" {
		where = append(where, "name_lower LIKE ?")
		args = append(args, "%"+q+"%")
	}

	// AND-semantics: person must have ALL listed labels.
	// Use INTERSECT subqueries — one per label ID.
	if len(labelIDs) > 0 {
		sub := buildLabelIntersect(labelIDs)
		where = append(where, "id IN ("+sub+")")
		for _, id := range labelIDs {
			args = append(args, id)
		}
	}

	query := `SELECT id, prefix, name, nickname, date_of_birth, relationship_type,
	                 other_notes, avatar_path, avatar_mime_type, avatar_size, avatar_uploaded_at,
	                 created_at, updated_at
	          FROM person`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY name_lower LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("people: list query: %w", err)
	}
	defer rows.Close()

	var people []Person
	for rows.Next() {
		p, err := scanPerson(rows)
		if err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, rows.Err()
}

// buildLabelIntersect builds an INTERSECT subquery for AND-semantics label filtering.
// Each label ID contributes one SELECT so the intersection returns only person IDs
// that have ALL the requested labels.
func buildLabelIntersect(labelIDs []int64) string {
	parts := make([]string, len(labelIDs))
	for i := range labelIDs {
		parts[i] = "SELECT person_id FROM person_label WHERE label_id = ?"
	}
	return strings.Join(parts, " INTERSECT ")
}

func (r *sqlPersonRepo) Get(ctx context.Context, id int64) (*Person, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, prefix, name, nickname, date_of_birth, relationship_type,
		        other_notes, avatar_path, avatar_mime_type, avatar_size, avatar_uploaded_at,
		        created_at, updated_at
		 FROM person WHERE id = ?`,
		id,
	)
	p, err := scanPerson(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *sqlPersonRepo) GetSelf(ctx context.Context) (*Person, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, prefix, name, nickname, date_of_birth, relationship_type,
		        other_notes, avatar_path, avatar_mime_type, avatar_size, avatar_uploaded_at,
		        created_at, updated_at
		 FROM person WHERE is_self = 1 LIMIT 1`,
	)
	p, err := scanPerson(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *sqlPersonRepo) Create(ctx context.Context, tx *sql.Tx, p Person) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	dob := formatNullableDate(p.DateOfBirth)

	res, err := tx.ExecContext(ctx,
		`INSERT INTO person (prefix, name, nickname, date_of_birth, relationship_type, other_notes, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Prefix, p.Name, p.Nickname, dob, p.RelationshipType, p.OtherNotes, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("people: create person: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("people: create person last id: %w", err)
	}
	return id, nil
}

func (r *sqlPersonRepo) Update(ctx context.Context, tx *sql.Tx, p Person) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	dob := formatNullableDate(p.DateOfBirth)

	_, err := tx.ExecContext(ctx,
		`UPDATE person
		 SET prefix=?, name=?, nickname=?, date_of_birth=?, relationship_type=?, other_notes=?, updated_at=?
		 WHERE id=?`,
		p.Prefix, p.Name, p.Nickname, dob, p.RelationshipType, p.OtherNotes, now, p.ID,
	)
	if err != nil {
		return fmt.Errorf("people: update person: %w", err)
	}
	return nil
}

func (r *sqlPersonRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM person WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("people: delete person: %w", err)
	}
	return nil
}

func (r *sqlPersonRepo) SetSelf(ctx context.Context, tx *sql.Tx, personID int64) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	res, err := tx.ExecContext(ctx, `UPDATE person SET is_self = 1, updated_at = ? WHERE id = ?`, now, personID)
	if err != nil {
		return fmt.Errorf("people: set self: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("people: set self rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("person not found")
	}
	return nil
}

func (r *sqlPersonRepo) ClearSelf(ctx context.Context, tx *sql.Tx) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := tx.ExecContext(ctx, `UPDATE person SET is_self = 0, updated_at = ? WHERE is_self = 1`, now)
	if err != nil {
		return fmt.Errorf("people: clear self: %w", err)
	}
	return nil
}

func (r *sqlPersonRepo) UpdateAvatar(ctx context.Context, tx *sql.Tx, personID int64, path, mimeType string, size int64, uploadedAt time.Time) error {
	uploadedAtStr := uploadedAt.UTC().Format(time.RFC3339Nano)
	now := time.Now().UTC().Format(time.RFC3339Nano)

	_, err := tx.ExecContext(ctx,
		`UPDATE person
		 SET avatar_path=?, avatar_mime_type=?, avatar_size=?, avatar_uploaded_at=?, updated_at=?
		 WHERE id=?`,
		path, mimeType, size, uploadedAtStr, now, personID,
	)
	if err != nil {
		return fmt.Errorf("people: update avatar: %w", err)
	}
	return nil
}

func (r *sqlPersonRepo) ClearAvatar(ctx context.Context, tx *sql.Tx, personID int64) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	_, err := tx.ExecContext(ctx,
		`UPDATE person
		 SET avatar_path='', avatar_mime_type='', avatar_size=0, avatar_uploaded_at=NULL, updated_at=?
		 WHERE id=?`,
		now, personID,
	)
	if err != nil {
		return fmt.Errorf("people: clear avatar: %w", err)
	}
	return nil
}

// ---- sqlContactRepo ---------------------------------------------------------

type sqlContactRepo struct{ db *sql.DB }

// NewContactRepo returns a ContactRepo backed by db.
func NewContactRepo(db *sql.DB) ContactRepo { return &sqlContactRepo{db: db} }

func (r *sqlContactRepo) ReplaceAll(ctx context.Context, tx *sql.Tx, personID int64, contacts []ContactInfo) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM contact_info WHERE person_id = ?`, personID); err != nil {
		return fmt.Errorf("people: delete contacts: %w", err)
	}
	for i, c := range contacts {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO contact_info (person_id, type, value, label, position) VALUES (?, ?, ?, ?, ?)`,
			personID, c.Type, c.Value, c.Label, i,
		)
		if err != nil {
			return fmt.Errorf("people: insert contact[%d]: %w", i, err)
		}
	}
	return nil
}

func (r *sqlContactRepo) ListByPerson(ctx context.Context, personID int64) ([]ContactInfo, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, person_id, type, value, label, position
		 FROM contact_info WHERE person_id = ? ORDER BY position`,
		personID,
	)
	if err != nil {
		return nil, fmt.Errorf("people: list contacts: %w", err)
	}
	defer rows.Close()

	var contacts []ContactInfo
	for rows.Next() {
		var c ContactInfo
		if err := rows.Scan(&c.ID, &c.PersonID, &c.Type, &c.Value, &c.Label, &c.Position); err != nil {
			return nil, fmt.Errorf("people: scan contact: %w", err)
		}
		contacts = append(contacts, c)
	}
	return contacts, rows.Err()
}

// ---- sqlLocationRepo --------------------------------------------------------

type sqlLocationRepo struct{ db *sql.DB }

// NewLocationRepo returns a LocationRepo backed by db.
func NewLocationRepo(db *sql.DB) LocationRepo { return &sqlLocationRepo{db: db} }

func (r *sqlLocationRepo) ReplaceAll(ctx context.Context, tx *sql.Tx, personID int64, locations []Location) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM location WHERE person_id = ?`, personID); err != nil {
		return fmt.Errorf("people: delete locations: %w", err)
	}
	for i, l := range locations {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO location (person_id, type, address, city, country, postal_code, position) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			personID, l.Type, l.Address, l.City, l.Country, l.PostalCode, i,
		)
		if err != nil {
			return fmt.Errorf("people: insert location[%d]: %w", i, err)
		}
	}
	return nil
}

func (r *sqlLocationRepo) ListByPerson(ctx context.Context, personID int64) ([]Location, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, person_id, type, address, city, country, postal_code, position
		 FROM location WHERE person_id = ? ORDER BY position`,
		personID,
	)
	if err != nil {
		return nil, fmt.Errorf("people: list locations: %w", err)
	}
	defer rows.Close()

	var locations []Location
	for rows.Next() {
		var l Location
		if err := rows.Scan(&l.ID, &l.PersonID, &l.Type, &l.Address, &l.City, &l.Country, &l.PostalCode, &l.Position); err != nil {
			return nil, fmt.Errorf("people: scan location: %w", err)
		}
		locations = append(locations, l)
	}
	return locations, rows.Err()
}

// ---- scan helpers -----------------------------------------------------------

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPerson(row rowScanner) (Person, error) {
	var p Person
	var dobStr sql.NullString
	var avatarUploadedAtStr sql.NullString
	var createdAt, updatedAt string

	err := row.Scan(
		&p.ID, &p.Prefix, &p.Name, &p.Nickname,
		&dobStr, &p.RelationshipType, &p.OtherNotes,
		&p.AvatarPath, &p.AvatarMimeType, &p.AvatarSize, &avatarUploadedAtStr,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return Person{}, fmt.Errorf("people: scan person: %w", err)
	}

	if dobStr.Valid && dobStr.String != "" {
		if t, err := parseDate(dobStr.String); err == nil {
			p.DateOfBirth = &t
		}
	}
	if avatarUploadedAtStr.Valid && avatarUploadedAtStr.String != "" {
		if t, err := parseTime(avatarUploadedAtStr.String); err == nil {
			p.AvatarUploadedAt = &t
		}
	}
	p.CreatedAt, _ = parseTime(createdAt)
	p.UpdatedAt, _ = parseTime(updatedAt)
	return p, nil
}

func parseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func formatNullableDate(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Format("2006-01-02")
}
