package settings_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "modernc.org/sqlite"

	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/internal/settings"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if err := internaldb.Up(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

func validBase() settings.UserSettings {
	return settings.UserSettings{
		DateFormat:            "YYYY-MM-DD",
		TimeFormat:            "24h",
		Timezone:              "UTC",
		AuditLogRetentionDays: 0,
	}
}

func TestSettings_Defaults(t *testing.T) {
	svc := settings.NewService(openTestDB(t))

	got, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.AuditLogRetentionDays != 0 {
		t.Errorf("default retention: want 0, got %d", got.AuditLogRetentionDays)
	}
}

func TestSettings_Update_RetentionDays(t *testing.T) {
	svc := settings.NewService(openTestDB(t))
	ctx := context.Background()

	in := validBase()
	in.AuditLogRetentionDays = 30

	out, err := svc.Update(ctx, in)
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if out.AuditLogRetentionDays != 30 {
		t.Errorf("want 30, got %d", out.AuditLogRetentionDays)
	}

	got, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.AuditLogRetentionDays != 30 {
		t.Errorf("persisted: want 30, got %d", got.AuditLogRetentionDays)
	}
}

func TestSettings_Update_NegativeRetentionDays(t *testing.T) {
	svc := settings.NewService(openTestDB(t))

	in := validBase()
	in.AuditLogRetentionDays = -1

	_, err := svc.Update(context.Background(), in)
	if !errors.Is(err, settings.ErrInvalidRetentionDays) {
		t.Errorf("want ErrInvalidRetentionDays, got %v", err)
	}
}

func TestSettings_GetRetentionDays(t *testing.T) {
	svc := settings.NewService(openTestDB(t))
	ctx := context.Background()

	in := validBase()

	in.AuditLogRetentionDays = 90
	if _, err := svc.Update(ctx, in); err != nil {
		t.Fatalf("update: %v", err)
	}

	days, err := svc.GetRetentionDays(ctx)
	if err != nil {
		t.Fatalf("get retention days: %v", err)
	}

	if days != 90 {
		t.Errorf("want 90, got %d", days)
	}
}
