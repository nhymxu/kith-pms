package people

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type PersonRepo interface {
	List(
		ctx context.Context,
		q string,
		labelIDs []int64,
		hasJournal bool,
		limit, offset int,
		sort string,
	) ([]Person, error)
	Count(ctx context.Context, q string, labelIDs []int64, hasJournal bool) (int, error)
	Get(ctx context.Context, id int64) (*Person, error)
	GetSelf(ctx context.Context) (*Person, error)
	Create(ctx context.Context, db bun.IDB, p Person) (int64, error)
	Update(ctx context.Context, db bun.IDB, p Person) error
	Delete(ctx context.Context, id int64) error
	SetSelf(ctx context.Context, db bun.IDB, personID int64) error
	ClearSelf(ctx context.Context, db bun.IDB) error
	UpdateAvatar(ctx context.Context, db bun.IDB, personID int64, path, mimeType string, size int64) error
	ClearAvatar(ctx context.Context, db bun.IDB, personID int64) error
	UpdateLastContact(ctx context.Context, db bun.IDB, personID int64, contactTime time.Time) error
}

type ContactRepo interface {
	// ReplaceAll deletes all contacts for personID and inserts the new slice.
	ReplaceAll(ctx context.Context, db bun.IDB, personID int64, contacts []ContactInfo) error
	ListByPerson(ctx context.Context, personID int64) ([]ContactInfo, error)
}

type LocationRepo interface {
	// ReplaceAll deletes all locations for personID and inserts the new slice.
	ReplaceAll(ctx context.Context, db bun.IDB, personID int64, locations []Location) error
	ListByPerson(ctx context.Context, personID int64) ([]Location, error)
}

// ---- sqlPersonRepo ----------------------------------------------------------

type sqlPersonRepo struct{ db *bun.DB }

func NewPersonRepo(db *bun.DB) PersonRepo { return &sqlPersonRepo{db: db} }

func (r *sqlPersonRepo) List(
	ctx context.Context,
	q string,
	labelIDs []int64,
	hasJournal bool,
	limit, offset int,
	sort string,
) ([]Person, error) {
	var people []Person

	sq := r.db.NewSelect().Model(&people)

	if q != "" {
		ql := "%" + strings.ToLower(q) + "%"
		sq = sq.Where("name_lower LIKE ? OR nickname_lower LIKE ?", ql, ql)
	}

	// AND-semantics: person must have ALL listed labels.
	// Use INTERSECT subqueries — one per label ID (bun has no INTERSECT support).
	if len(labelIDs) > 0 {
		sub := buildLabelIntersect(labelIDs)

		args := make([]any, len(labelIDs))
		for i, id := range labelIDs {
			args[i] = id
		}

		sq = sq.Where("id IN ("+sub+")", args...)
	}

	if hasJournal {
		sq = sq.Where(`EXISTS (SELECT 1 FROM activity_person WHERE person_id = "p"."id")`)
	}

	sq = sq.OrderExpr(buildOrderBy(sort)).Limit(limit).Offset(offset)

	if err := sq.Scan(ctx); err != nil {
		return nil, fmt.Errorf("people: list query: %w", err)
	}

	return people, nil
}

func (r *sqlPersonRepo) Count(ctx context.Context, q string, labelIDs []int64, hasJournal bool) (int, error) {
	sq := r.db.NewSelect().Model((*Person)(nil))

	if q != "" {
		ql := "%" + strings.ToLower(q) + "%"
		sq = sq.Where("name_lower LIKE ? OR nickname_lower LIKE ?", ql, ql)
	}

	if len(labelIDs) > 0 {
		sub := buildLabelIntersect(labelIDs)

		args := make([]any, len(labelIDs))
		for i, id := range labelIDs {
			args[i] = id
		}

		sq = sq.Where("id IN ("+sub+")", args...)
	}

	if hasJournal {
		sq = sq.Where(`EXISTS (SELECT 1 FROM activity_person WHERE person_id = "p"."id")`)
	}

	total, err := sq.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("people: count query: %w", err)
	}

	return total, nil
}

// buildOrderBy returns the ORDER BY expression based on sort parameter.
func buildOrderBy(sort string) string {
	switch sort {
	case "name":
		return "name_lower ASC"
	case "-name":
		return "name_lower DESC"
	case "last_contact":
		return "last_contact_at ASC NULLS LAST"
	case "-last_contact":
		return "last_contact_at DESC NULLS LAST"
	default:
		return "name_lower ASC"
	}
}

// buildLabelIntersect builds an INTERSECT subquery for AND-semantics label filtering.
// Each label ID contributes one SELECT so the intersection returns only person IDs
// that have ALL the requested labels.
func buildLabelIntersect(labelIDs []int64) string {
	parts := make([]string, len(labelIDs))
	for i := range labelIDs {
		parts[i] = "SELECT person_id FROM people_label_assignment WHERE label_id = ?"
	}

	return strings.Join(parts, " INTERSECT ")
}

func (r *sqlPersonRepo) Get(ctx context.Context, id int64) (*Person, error) {
	var p Person

	err := r.db.NewSelect().Model(&p).Where("\"p\".\"id\" = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *sqlPersonRepo) GetSelf(ctx context.Context) (*Person, error) {
	var p Person

	err := r.db.NewSelect().Model(&p).Where("\"p\".\"is_self\" = ?", true).Limit(1).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *sqlPersonRepo) Create(ctx context.Context, db bun.IDB, p Person) (int64, error) {
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt

	_, err := db.NewInsert().Model(&p).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("people: create person: %w", err)
	}

	return p.ID, nil
}

func (r *sqlPersonRepo) Update(ctx context.Context, db bun.IDB, p Person) error {
	p.UpdatedAt = time.Now().UTC()

	_, err := db.NewUpdate().Model(&p).WherePK().
		Column("prefix", "name", "nickname", "gender", "date_of_birth",
			"relationship_type", "last_contact_at", "other_notes", "updated_at").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("people: update person: %w", err)
	}

	return nil
}

func (r *sqlPersonRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*Person)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("people: delete person: %w", err)
	}

	return nil
}

func (r *sqlPersonRepo) SetSelf(ctx context.Context, db bun.IDB, personID int64) error {
	res, err := db.NewUpdate().Model((*Person)(nil)).
		Set("is_self = ?, updated_at = ?", true, time.Now().UTC()).
		Where("id = ?", personID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("people: set self: %w", err)
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("person not found: %d", personID)
	}

	return nil
}

func (r *sqlPersonRepo) ClearSelf(ctx context.Context, db bun.IDB) error {
	_, err := db.NewUpdate().Model((*Person)(nil)).
		Set("is_self = ?, updated_at = ?", false, time.Now().UTC()).
		Where("is_self = ?", true).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("people: clear self: %w", err)
	}

	return nil
}

func (r *sqlPersonRepo) UpdateAvatar(
	ctx context.Context,
	db bun.IDB,
	personID int64,
	path, mimeType string,
	size int64,
) error {
	now := time.Now().UTC()
	p := &Person{
		ID:               personID,
		AvatarPath:       path,
		AvatarMimeType:   mimeType,
		AvatarSize:       size,
		AvatarUploadedAt: &now,
		UpdatedAt:        now,
	}

	_, err := db.NewUpdate().Model(p).WherePK().
		Column("avatar_path", "avatar_mime_type", "avatar_size", "avatar_uploaded_at", "updated_at").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("people: update avatar: %w", err)
	}

	return nil
}

func (r *sqlPersonRepo) ClearAvatar(ctx context.Context, db bun.IDB, personID int64) error {
	now := time.Now().UTC()
	p := &Person{
		ID:        personID,
		UpdatedAt: now,
	}

	_, err := db.NewUpdate().Model(p).WherePK().
		Column("avatar_path", "avatar_mime_type", "avatar_size", "avatar_uploaded_at", "updated_at").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("people: clear avatar: %w", err)
	}

	return nil
}

func (r *sqlPersonRepo) UpdateLastContact(
	ctx context.Context,
	db bun.IDB,
	personID int64,
	contactTime time.Time,
) error {
	p := &Person{
		ID:            personID,
		LastContactAt: &contactTime,
		UpdatedAt:     time.Now().UTC(),
	}

	_, err := db.NewUpdate().Model(p).WherePK().
		Column("last_contact_at", "updated_at").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("people: update last contact: %w", err)
	}

	return nil
}

// ---- sqlContactRepo ---------------------------------------------------------

type sqlContactRepo struct{ db *bun.DB }

func NewContactRepo(db *bun.DB) ContactRepo { return &sqlContactRepo{db: db} }

func (r *sqlContactRepo) ReplaceAll(ctx context.Context, db bun.IDB, personID int64, contacts []ContactInfo) error {
	_, err := db.NewDelete().Model((*ContactInfo)(nil)).Where("person_id = ?", personID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("people: delete contacts: %w", err)
	}

	if len(contacts) == 0 {
		return nil
	}

	// Ensure PersonID and Position are set before bulk insert.
	for i := range contacts {
		contacts[i].PersonID = personID
		contacts[i].Position = i
	}

	_, err = db.NewInsert().Model(&contacts).Exec(ctx)
	if err != nil {
		return fmt.Errorf("people: insert contacts: %w", err)
	}

	return nil
}

func (r *sqlContactRepo) ListByPerson(ctx context.Context, personID int64) ([]ContactInfo, error) {
	var contacts []ContactInfo

	err := r.db.NewSelect().Model(&contacts).
		Where("person_id = ?", personID).
		OrderExpr("position ASC, id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("people: list contacts: %w", err)
	}

	return contacts, nil
}

// ---- sqlLocationRepo --------------------------------------------------------

type sqlLocationRepo struct{ db *bun.DB }

func NewLocationRepo(db *bun.DB) LocationRepo { return &sqlLocationRepo{db: db} }

func (r *sqlLocationRepo) ReplaceAll(ctx context.Context, db bun.IDB, personID int64, locations []Location) error {
	_, err := db.NewDelete().Model((*Location)(nil)).Where("person_id = ?", personID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("people: delete locations: %w", err)
	}

	if len(locations) == 0 {
		return nil
	}

	// Ensure PersonID and Position are set before bulk insert.
	for i := range locations {
		locations[i].PersonID = personID
		locations[i].Position = i
	}

	_, err = db.NewInsert().Model(&locations).Exec(ctx)
	if err != nil {
		return fmt.Errorf("people: insert locations: %w", err)
	}

	return nil
}

func (r *sqlLocationRepo) ListByPerson(ctx context.Context, personID int64) ([]Location, error) {
	var locations []Location

	err := r.db.NewSelect().Model(&locations).
		Where("person_id = ?", personID).
		OrderExpr("position ASC, id ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("people: list locations: %w", err)
	}

	return locations, nil
}
