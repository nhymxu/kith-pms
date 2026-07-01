package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/auth"
	internaldb "github.com/nhymxu/kith-pms/internal/db"
)

// setupTestDB creates an in-memory SQLite database with schema applied.
func setupTestDB(t *testing.T) *bun.DB {
	t.Helper()

	db, err := internaldb.Open(":memory:", 1)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := internaldb.Up(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return db
}

func TestChangePassword_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	userRepo := auth.NewUserRepo(db)
	sessionRepo := auth.NewSessionRepo(db)

	// Setup: create user with initial password
	initialPwd := "initial-password-123"

	initialHash, err := auth.HashPassword(initialPwd)
	if err != nil {
		t.Fatalf("hash initial password: %v", err)
	}

	if err := userRepo.UpsertUser(ctx, initialHash); err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Create service
	svc := &auth.Service{
		Users:    userRepo,
		Sessions: sessionRepo,
		Secret:   []byte("test-secret-key-32-bytes-long!"),
		Lifetime: 24 * time.Hour,
	}

	// Test: change password
	newPwd := "new-password-456"

	err = svc.ChangePassword(ctx, initialPwd, newPwd)
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	// Verify: new password works
	user, err := userRepo.GetUser(ctx)
	if err != nil {
		t.Fatalf("get user after change: %v", err)
	}

	ok, err := auth.VerifyPassword(user.PasswordHash, newPwd)
	if err != nil {
		t.Fatalf("verify new password: %v", err)
	}

	if !ok {
		t.Error("new password does not verify")
	}

	// Verify: old password no longer works
	ok, err = auth.VerifyPassword(user.PasswordHash, initialPwd)
	if err != nil {
		t.Fatalf("verify old password: %v", err)
	}

	if ok {
		t.Error("old password still verifies (should not)")
	}
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	userRepo := auth.NewUserRepo(db)
	sessionRepo := auth.NewSessionRepo(db)

	// Setup: create user with initial password
	initialPwd := "correct-password-123"

	initialHash, err := auth.HashPassword(initialPwd)
	if err != nil {
		t.Fatalf("hash initial password: %v", err)
	}

	if err := userRepo.UpsertUser(ctx, initialHash); err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Create service
	svc := &auth.Service{
		Users:    userRepo,
		Sessions: sessionRepo,
		Secret:   []byte("test-secret-key-32-bytes-long!"),
		Lifetime: 24 * time.Hour,
	}

	// Test: attempt to change password with wrong current password
	wrongPwd := "wrong-password-999"
	newPwd := "new-password-456"
	err = svc.ChangePassword(ctx, wrongPwd, newPwd)

	// Verify: returns ErrInvalidCredentials
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}

	// Verify: password was NOT changed
	user, err := userRepo.GetUser(ctx)
	if err != nil {
		t.Fatalf("get user after failed change: %v", err)
	}

	ok, err := auth.VerifyPassword(user.PasswordHash, initialPwd)
	if err != nil {
		t.Fatalf("verify original password: %v", err)
	}

	if !ok {
		t.Error("original password no longer works (should still work)")
	}
}

func TestChangePassword_NoUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	userRepo := auth.NewUserRepo(db)
	sessionRepo := auth.NewSessionRepo(db)

	// Create service (no user in database)
	svc := &auth.Service{
		Users:    userRepo,
		Sessions: sessionRepo,
		Secret:   []byte("test-secret-key-32-bytes-long!"),
		Lifetime: 24 * time.Hour,
	}

	// Test: attempt to change password when no user exists
	err := svc.ChangePassword(ctx, "any-password", "new-password")

	// Verify: returns ErrInvalidCredentials
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}
