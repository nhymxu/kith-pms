package note

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type Repo struct {
	db *bun.DB
}

func NewRepo(db *bun.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, tx bun.Tx, n *Note) (int64, error) {
	n.CreatedAt = time.Now().UTC()
	n.UpdatedAt = n.CreatedAt

	_, err := tx.NewInsert().Model(n).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("insert note: %w", err)
	}

	return n.ID, nil
}

func (r *Repo) GetByID(ctx context.Context, id int64) (*Note, error) {
	var n Note

	if err := r.db.NewSelect().Model(&n).Where("id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}

	return &n, nil
}

// PersonName looks up a person's name by ID, for audit logging where the note
// itself isn't the loggable subject the person it belongs to is.
func (r *Repo) PersonName(ctx context.Context, personID int64) (string, error) {
	var name string

	err := r.db.NewSelect().
		TableExpr("person").
		Column("name").
		Where("id = ?", personID).
		Scan(ctx, &name)
	if err != nil {
		return "", fmt.Errorf("lookup person name: %w", err)
	}

	return name, nil
}

func (r *Repo) Update(ctx context.Context, tx bun.Tx, n *Note) error {
	n.UpdatedAt = time.Now().UTC()

	_, err := tx.NewUpdate().Model(n).WherePK().
		Column("title", "content", "updated_at").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("update note: %w", err)
	}

	return nil
}

func (r *Repo) Delete(ctx context.Context, tx bun.Tx, id int64) error {
	_, err := tx.NewDelete().Model((*Note)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	return nil
}

func (r *Repo) ListByPerson(ctx context.Context, personID int64, page, pageSize int) (*List, error) {
	var total int

	err := r.db.NewSelect().Model((*Note)(nil)).
		Where("person_id = ?", personID).
		ColumnExpr("COUNT(*)").
		Scan(ctx, &total)
	if err != nil {
		return nil, fmt.Errorf("count notes: %w", err)
	}

	var items []Note

	q := r.db.NewSelect().Model(&items).
		Where("person_id = ?", personID).
		OrderExpr("created_at DESC, id DESC")

	if pageSize > 0 {
		q = q.Limit(pageSize).Offset((page - 1) * pageSize)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}

	if items == nil {
		items = []Note{}
	}

	return &List{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Search finds notes matching query via note_fts, ranked by relevance (bm25),
// each paired with its owner's name for display in cross-person search results.
func (r *Repo) Search(ctx context.Context, query string, limit int) ([]WithPerson, error) {
	var rows []struct {
		Note
		PersonName string `bun:"person_name"`
	}

	err := r.db.NewSelect().
		TableExpr("note n").
		ColumnExpr("n.*, p.name AS person_name").
		Join("JOIN note_fts ON note_fts.rowid = n.id").
		Join("JOIN person p ON p.id = n.person_id").
		Where("note_fts MATCH ?", sanitizeFTSQuery(query)).
		OrderExpr("bm25(note_fts)").
		Limit(limit).
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("search notes: %w", err)
	}

	results := make([]WithPerson, 0, len(rows))
	for _, row := range rows {
		results = append(results, WithPerson{Note: row.Note, PersonName: row.PersonName})
	}

	return results, nil
}

// sanitizeFTSQuery escapes user input for use in FTS5 MATCH queries.
// Replaces double quotes with single quotes, then wraps in double quotes
// for phrase matching — prevents malformed MATCH syntax errors.
func sanitizeFTSQuery(q string) string {
	q = strings.TrimSpace(q)
	q = strings.ReplaceAll(q, `"`, `""`)

	return `"` + q + `"`
}
