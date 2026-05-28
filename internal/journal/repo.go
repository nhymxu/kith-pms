package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type ActivityRepo interface {
	Create(ctx context.Context, tx bun.Tx, a Activity) (int64, error)
	Update(ctx context.Context, tx bun.Tx, a Activity) error
	Get(ctx context.Context, id int64) (*Activity, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params ListParams) ([]Activity, error)
	Count(ctx context.Context, params ListParams) (int, error)
}

type ActivityPersonRepo interface {
	ReplaceAll(ctx context.Context, tx bun.Tx, activityID int64, personIDs []int64) error
	ListByActivity(ctx context.Context, activityID int64) ([]ActivityPerson, error)
}

// ---- sqlActivityRepo --------------------------------------------------------

type sqlActivityRepo struct{ db *bun.DB }

func NewActivityRepo(db *bun.DB) ActivityRepo { return &sqlActivityRepo{db: db} }

func (r *sqlActivityRepo) Create(ctx context.Context, tx bun.Tx, a Activity) (int64, error) {
	a.CreatedAt = time.Now().UTC()
	a.UpdatedAt = a.CreatedAt
	_, err := tx.NewInsert().Model(&a).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("journal: create activity: %w", err)
	}

	return a.ID, nil
}

func (r *sqlActivityRepo) Update(ctx context.Context, tx bun.Tx, a Activity) error {
	a.UpdatedAt = time.Now().UTC()
	_, err := tx.NewUpdate().Model(&a).WherePK().
		Column("title", "occurred_at_date", "occurred_at_time", "content", "updated_at").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("journal: update activity: %w", err)
	}

	return nil
}

func (r *sqlActivityRepo) Get(ctx context.Context, id int64) (*Activity, error) {
	var a Activity
	err := r.db.NewSelect().Model(&a).Where("\"activity\".\"id\" = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("journal: get activity: %w", err)
	}

	return &a, nil
}

func (r *sqlActivityRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*Activity)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("journal: delete activity: %w", err)
	}

	return nil
}

// buildActivityQuery constructs a shared base SELECT with all WHERE/JOIN filters applied.
func (r *sqlActivityRepo) buildActivityQuery(params ListParams) *bun.SelectQuery {
	q := r.db.NewSelect().TableExpr("activity")

	if strings.TrimSpace(params.Query) != "" {
		q = q.Join("JOIN activity_fts ON activity_fts.rowid = activity.id").
			Where("activity_fts MATCH ?", sanitizeFTSQuery(params.Query))
	}

	if len(params.PersonIDs) > 0 {
		q = q.Where(
			"activity.id IN (SELECT activity_id FROM activity_person WHERE person_id IN (?))",
			bun.List(params.PersonIDs),
		)
	}

	if len(params.LabelIDs) > 0 {
		q = q.Where(
			"activity.id IN (SELECT ap.activity_id FROM activity_person ap JOIN person_label pl ON pl.person_id = ap.person_id WHERE pl.label_id IN (?))",
			bun.List(params.LabelIDs),
		)
	}

	if params.FromDate != "" && params.ToDate != "" {
		q = q.Where("activity.occurred_at_date BETWEEN ? AND ?", params.FromDate, params.ToDate)
	} else if params.FromDate != "" {
		q = q.Where("activity.occurred_at_date >= ?", params.FromDate)
	} else if params.ToDate != "" {
		q = q.Where("activity.occurred_at_date <= ?", params.ToDate)
	}

	return q
}

func (r *sqlActivityRepo) Count(ctx context.Context, params ListParams) (int, error) {
	count, err := r.buildActivityQuery(params).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("journal: count activities: %w", err)
	}

	return count, nil
}

// List builds a dynamic query based on non-zero filter fields in params.
func (r *sqlActivityRepo) List(ctx context.Context, params ListParams) ([]Activity, error) {
	useFTS := strings.TrimSpace(params.Query) != ""

	q := r.buildActivityQuery(params).
		ColumnExpr("activity.id, activity.title, activity.occurred_at_date, activity.occurred_at_time, activity.content, activity.created_at, activity.updated_at")

	if useFTS {
		q = q.OrderExpr("bm25(activity_fts), activity.occurred_at_date DESC, activity.id DESC")
	} else {
		q = q.OrderExpr("activity.occurred_at_date DESC, activity.id DESC")
	}

	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 30
	}

	page := params.Page
	if page < 1 {
		page = 1
	}

	var list []Activity
	err := q.Limit(pageSize).Offset((page-1)*pageSize).Scan(ctx, &list)
	if err != nil {
		return nil, fmt.Errorf("journal: list activities: %w", err)
	}

	return list, nil
}

// sanitizeFTSQuery escapes user input for use in FTS5 MATCH queries.
// Replaces double quotes with single quotes, then wraps in double quotes
// for phrase matching — prevents malformed MATCH syntax errors.
func sanitizeFTSQuery(q string) string {
	q = strings.TrimSpace(q)
	// Escape existing double quotes by doubling them (FTS5 phrase literal rules).
	q = strings.ReplaceAll(q, `"`, `""`)

	return `"` + q + `"`
}

// ---- sqlActivityPersonRepo --------------------------------------------------

// activityPersonLink is a private bun model for the activity_person join table.
type activityPersonLink struct {
	bun.BaseModel `bun:"table:activity_person"`
	ActivityID    int64 `bun:"activity_id"`
	PersonID      int64 `bun:"person_id"`
}

type sqlActivityPersonRepo struct{ db *bun.DB }

func NewActivityPersonRepo(db *bun.DB) ActivityPersonRepo {
	return &sqlActivityPersonRepo{db: db}
}

// ReplaceAll deletes all person links for the activity and bulk-inserts new ones.
func (r *sqlActivityPersonRepo) ReplaceAll(ctx context.Context, tx bun.Tx, activityID int64, personIDs []int64) error {
	_, err := tx.NewDelete().Model((*activityPersonLink)(nil)).Where("activity_id = ?", activityID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("journal: delete activity_person: %w", err)
	}

	if len(personIDs) == 0 {
		return nil
	}

	links := make([]activityPersonLink, len(personIDs))
	for i, pid := range personIDs {
		links[i] = activityPersonLink{ActivityID: activityID, PersonID: pid}
	}

	_, err = tx.NewInsert().Model(&links).Exec(ctx)
	if err != nil {
		return fmt.Errorf("journal: insert activity_person: %w", err)
	}

	return nil
}

// ListByActivity returns all people linked to the given activity.
func (r *sqlActivityPersonRepo) ListByActivity(ctx context.Context, activityID int64) ([]ActivityPerson, error) {
	var rows []struct {
		PersonID   int64  `bun:"person_id"`
		Name       string `bun:"name"`
		Nickname   string `bun:"nickname"`
		AvatarPath string `bun:"avatar_path"`
	}

	err := r.db.NewSelect().
		TableExpr("activity_person ap").
		Join("JOIN person p ON p.id = ap.person_id").
		ColumnExpr("ap.person_id, p.name, COALESCE(p.nickname, '') AS nickname, COALESCE(p.avatar_path, '') AS avatar_path").
		Where("ap.activity_id = ?", activityID).
		OrderExpr("p.name").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("journal: list activity people: %w", err)
	}

	result := make([]ActivityPerson, len(rows))
	for i, row := range rows {
		result[i] = ActivityPerson{
			PersonID:   row.PersonID,
			Name:       row.Name,
			Nickname:   row.Nickname,
			AvatarPath: row.AvatarPath,
		}
	}

	return result, nil
}
