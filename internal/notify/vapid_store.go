package notify

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/uptrace/bun"
)

const (
	// KindWebPush is the channel kind whose subscriptions are bound to the VAPID key.
	KindWebPush = "webpush"
	vapidAAD    = "webpush_vapid|1"
)

// VAPIDStore holds the single server-wide VAPID key pair. The private key is
// sealed with the same SecretBox as channel configs.
type VAPIDStore struct {
	db  *bun.DB
	box *SecretBox
	mu  sync.Mutex
}

func NewVAPIDStore(db *bun.DB, box *SecretBox) *VAPIDStore { return &VAPIDStore{db: db, box: box} }

// Keys returns the key pair, generating it on first use.
func (s *VAPIDStore) Keys(ctx context.Context) (priv, pub string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var row struct {
		PublicKey  string `bun:"public_key"`
		PrivateEnc []byte `bun:"private_enc"`
	}

	err = s.db.NewRaw(`SELECT public_key, private_enc FROM webpush_vapid WHERE id = 1`).Scan(ctx, &row)
	if err == nil {
		pt, openErr := s.box.Open(vapidAAD, row.PrivateEnc)
		if openErr != nil {
			return "", "", ErrSecretUnreadable
		}

		return string(pt), row.PublicKey, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", fmt.Errorf("notify: read vapid: %w", err)
	}

	return s.replace(ctx, s.db)
}

// Rotate swaps in a new key pair and deletes every webpush channel: existing
// subscriptions are bound to the old public key and can never deliver again.
func (s *VAPIDStore) Rotate(ctx context.Context) (pub string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err = s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewRaw(`DELETE FROM reminder_delivery WHERE channel_id IN
(SELECT id FROM notification_channel WHERE kind = ?)`, KindWebPush).Exec(ctx); err != nil {
			return err
		}

		if _, err := tx.NewRaw(`DELETE FROM notification_channel WHERE kind = ?`, KindWebPush).Exec(ctx); err != nil {
			return err
		}

		_, pub, err = s.replace(ctx, tx)

		return err
	})
	if err != nil {
		return "", fmt.Errorf("notify: rotate vapid: %w", err)
	}

	return pub, nil
}

func (s *VAPIDStore) replace(ctx context.Context, db bun.IDB) (priv, pub string, err error) {
	priv, pub, err = webpush.GenerateVAPIDKeys()
	if err != nil {
		return "", "", fmt.Errorf("notify: generate vapid: %w", err)
	}

	enc, err := s.box.Seal(vapidAAD, []byte(priv))
	if err != nil {
		return "", "", err
	}

	_, err = db.NewRaw(`INSERT INTO webpush_vapid (id, public_key, private_enc, created_at) VALUES (1, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET public_key = excluded.public_key, private_enc = excluded.private_enc,
  created_at = excluded.created_at`, pub, enc, time.Now().UTC()).Exec(ctx)
	if err != nil {
		return "", "", fmt.Errorf("notify: store vapid: %w", err)
	}

	return priv, pub, nil
}
