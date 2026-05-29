package auth

import (
	"context"
	"testing"
	"time"

	"github.com/uptrace/bun"

	internaldb "github.com/nhymxu/kith-pms/internal/db"
)

// newTestDB opens an in-memory SQLite DB and runs all migrations.
func newTestDB(t *testing.T) *bun.DB {
	t.Helper()

	db, err := internaldb.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := internaldb.Up(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	t.Cleanup(func() { db.Close() })

	return db
}

func TestIssueAndLookup(t *testing.T) {
	db := newTestDB(t)
	repo := NewSessionRepo(db)
	secret := []byte("supersecretkey-at-least-32-bytes!!")
	ctx := context.Background()

	// Seed user row (sessions FK references user.id=1).
	_, err := db.ExecContext(ctx, `INSERT INTO user (id, password_hash) VALUES (1, 'hash')`)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	token, err := Issue(ctx, 1, "127.0.0.1", "test-agent", repo, secret, 30*24*time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	sess, err := Lookup(ctx, token, repo, secret)
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}

	if sess == nil {
		t.Fatal("expected session, got nil")
	}

	if sess.UserID != 1 {
		t.Errorf("expected userID=1, got %d", sess.UserID)
	}
}

func TestLookup_ExpiredSession(t *testing.T) {
	db := newTestDB(t)
	repo := NewSessionRepo(db)
	secret := []byte("supersecretkey-at-least-32-bytes!!")
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `INSERT INTO user (id, password_hash) VALUES (1, 'hash')`)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	// Issue with -1s lifetime so it expires immediately.
	token, err := Issue(ctx, 1, "127.0.0.1", "ua", repo, secret, -time.Second)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	sess, err := Lookup(ctx, token, repo, secret)
	if err != nil {
		t.Fatalf("Lookup error: %v", err)
	}

	if sess != nil {
		t.Error("expected nil session for expired token")
	}
}

func TestRevoke(t *testing.T) {
	db := newTestDB(t)
	repo := NewSessionRepo(db)
	secret := []byte("supersecretkey-at-least-32-bytes!!")
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `INSERT INTO user (id, password_hash) VALUES (1, 'hash')`)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	token, _ := Issue(ctx, 1, "127.0.0.1", "ua", repo, secret, time.Hour)

	if err := Revoke(ctx, token, repo, secret); err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	sess, err := Lookup(ctx, token, repo, secret)
	if err != nil {
		t.Fatalf("Lookup after revoke: %v", err)
	}

	if sess != nil {
		t.Error("expected nil session after revoke")
	}
}

func TestRevokeAll(t *testing.T) {
	db := newTestDB(t)
	repo := NewSessionRepo(db)
	secret := []byte("supersecretkey-at-least-32-bytes!!")
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `INSERT INTO user (id, password_hash) VALUES (1, 'hash')`)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	t1, _ := Issue(ctx, 1, "1.1.1.1", "ua1", repo, secret, time.Hour)
	t2, _ := Issue(ctx, 1, "2.2.2.2", "ua2", repo, secret, time.Hour)

	if err := RevokeAll(ctx, 1, repo); err != nil {
		t.Fatalf("RevokeAll: %v", err)
	}

	for _, tok := range []string{t1, t2} {
		sess, err := Lookup(ctx, tok, repo, secret)
		if err != nil {
			t.Fatalf("Lookup: %v", err)
		}

		if sess != nil {
			t.Errorf("expected nil after RevokeAll, got session for token")
		}
	}
}

func TestLookup_InvalidToken(t *testing.T) {
	db := newTestDB(t)
	repo := NewSessionRepo(db)
	secret := []byte("supersecretkey-at-least-32-bytes!!")
	ctx := context.Background()

	sess, err := Lookup(ctx, "not-valid-base64!!!", repo, secret)
	if err != nil {
		t.Fatalf("expected no error for malformed token, got: %v", err)
	}

	if sess != nil {
		t.Error("expected nil for malformed token")
	}
}
