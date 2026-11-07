package notify

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/uptrace/bun"
)

var ErrChannelNotFound = errors.New("notify: channel not found")

// Record is a notification_channel row. Config is plaintext and only set when
// the stored blob decrypted; otherwise Unreadable is true.
type Record struct {
	bun.BaseModel `bun:"table:notification_channel"`

	ID         int64           `bun:",pk,autoincrement"`
	Kind       string          `bun:"kind"`
	Name       string          `bun:"name"`
	Enabled    bool            `bun:"enabled"`
	ConfigEnc  []byte          `bun:"config_enc"`
	EnabledAt  time.Time       `bun:"enabled_at"`
	CreatedAt  time.Time       `bun:"created_at"`
	UpdatedAt  time.Time       `bun:"updated_at"`
	Config     json.RawMessage `bun:"-"`
	Unreadable bool            `bun:"-"`
}

type ChannelRepo struct {
	db  *bun.DB
	box *SecretBox
	now func() time.Time
}

func NewChannelRepo(db *bun.DB, box *SecretBox) *ChannelRepo {
	return &ChannelRepo{db: db, box: box, now: time.Now}
}

func aad(kind string, id int64) string { return kind + "|" + strconv.FormatInt(id, 10) }

func (r *ChannelRepo) open(rec *Record) {
	pt, err := r.box.Open(aad(rec.Kind, rec.ID), rec.ConfigEnc)
	if err != nil {
		rec.Unreadable = true
		return
	}

	rec.Config = pt
}

// Create inserts the row first because the AAD needs the id.
func (r *ChannelRepo) Create(
	ctx context.Context,
	kind, name string,
	enabled bool,
	cfg json.RawMessage,
) (*Record, error) {
	now := r.now().UTC()
	rec := &Record{
		Kind:      kind,
		Name:      name,
		Enabled:   enabled,
		ConfigEnc: []byte{},
		EnabledAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(rec).Exec(ctx); err != nil {
			return err
		}

		enc, err := r.box.Seal(aad(kind, rec.ID), cfg)
		if err != nil {
			return err
		}

		rec.ConfigEnc = enc
		_, err = tx.NewUpdate().Model(rec).Column("config_enc").WherePK().Exec(ctx)

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("notify: create channel: %w", err)
	}

	rec.Config = cfg

	return rec, nil
}

func (r *ChannelRepo) Get(ctx context.Context, id int64) (*Record, error) {
	rec := new(Record)
	if err := r.db.NewSelect().Model(rec).Where("id = ?", id).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChannelNotFound
		}

		return nil, fmt.Errorf("notify: get channel: %w", err)
	}

	r.open(rec)

	return rec, nil
}

func (r *ChannelRepo) List(ctx context.Context) ([]*Record, error) {
	var recs []*Record
	if err := r.db.NewSelect().Model(&recs).Order("id").Scan(ctx); err != nil {
		return nil, fmt.Errorf("notify: list channels: %w", err)
	}

	for _, rec := range recs {
		r.open(rec)
	}

	return recs, nil
}

func (r *ChannelRepo) UpdateConfig(ctx context.Context, id int64, cfg json.RawMessage) error {
	rec, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	enc, err := r.box.Seal(aad(rec.Kind, id), cfg)
	if err != nil {
		return err
	}

	return r.exec(ctx, id, "config_enc = ?", enc)
}

func (r *ChannelRepo) Rename(ctx context.Context, id int64, name string) error {
	return r.exec(ctx, id, "name = ?", name)
}

// SetEnabled stamps enabled_at only on a disabled->enabled transition so nothing stale is enqueued later.
func (r *ChannelRepo) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	if !enabled {
		return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			if _, err := tx.NewRaw(`UPDATE reminder_delivery SET status = 'cancelled'
WHERE channel_id = ? AND status = 'pending'`, id).Exec(ctx); err != nil {
				return fmt.Errorf("notify: cancel pending deliveries: %w", err)
			}

			return r.execOn(ctx, tx, id, "enabled = 0")
		})
	}

	return r.exec(ctx, id, "enabled_at = CASE WHEN ? AND NOT enabled THEN ? ELSE enabled_at END, enabled = ?",
		enabled, r.now().UTC(), enabled)
}

func (r *ChannelRepo) Delete(ctx context.Context, id int64) error {
	if _, err := r.db.NewRaw(`DELETE FROM reminder_delivery WHERE channel_id = ?`, id).Exec(ctx); err != nil {
		return fmt.Errorf("notify: delete deliveries: %w", err)
	}

	res, err := r.db.NewDelete().Model((*Record)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("notify: delete channel: %w", err)
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return ErrChannelNotFound
	}

	return nil
}

func (r *ChannelRepo) exec(ctx context.Context, id int64, set string, args ...any) error {
	return r.execOn(ctx, r.db, id, set, args...)
}

func (r *ChannelRepo) execOn(ctx context.Context, db bun.IDB, id int64, set string, args ...any) error {
	res, err := db.NewUpdate().Model((*Record)(nil)).
		Set(set, args...).Set("updated_at = ?", r.now().UTC()).
		Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("notify: update channel: %w", err)
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return ErrChannelNotFound
	}

	return nil
}
