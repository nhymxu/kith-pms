package settings_test

import (
	"context"
	"errors"
	"testing"

	"github.com/uptrace/bun"

	internaldb "github.com/nhymxu/kith-pms/internal/db"
	"github.com/nhymxu/kith-pms/internal/settings"
)

func openTestDB(t *testing.T) *bun.DB {
	t.Helper()

	db, err := internaldb.Open(":memory:", 1)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	if err := internaldb.Up(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

func validBase() settings.UserSettings {
	return settings.UserSettings{
		DateFormat:                "YYYY-MM-DD",
		TimeFormat:                "24h",
		Timezone:                  "UTC",
		AuditLogRetentionDays:     0,
		NetworkColorBy:            "labels",
		NetworkShowAvatar:         false,
		NetworkShowOnlyMine:       false,
		NetworkShowUnconnected:    true,
		NetworkOnlyMineDepth:      "direct",
		AllowFavoriteToggleOnList: true,
		FavoriteFirstDefault:      false,
		DefaultPeopleSort:         "name",
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

func TestSettings_Defaults_FavoritesFields(t *testing.T) {
	svc := settings.NewService(openTestDB(t))

	// On a fresh test DB (no rows), Get() should return the built-in defaults.
	got, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.AllowFavoriteToggleOnList != true {
		t.Errorf("AllowFavoriteToggleOnList: want true, got %v", got.AllowFavoriteToggleOnList)
	}

	if got.FavoriteFirstDefault != false {
		t.Errorf("FavoriteFirstDefault: want false, got %v", got.FavoriteFirstDefault)
	}

	if got.DefaultPeopleSort != "name" {
		t.Errorf("DefaultPeopleSort: want \"name\", got %q", got.DefaultPeopleSort)
	}
}

func TestSettings_Update_FavoritesFields_Roundtrip(t *testing.T) {
	svc := settings.NewService(openTestDB(t))
	ctx := context.Background()

	// Build a valid UserSettings with custom favorite-related fields.
	in := validBase()
	in.AllowFavoriteToggleOnList = false
	in.FavoriteFirstDefault = true
	in.DefaultPeopleSort = "-last_contact"

	// Update and verify the response.
	out, err := svc.Update(ctx, in)
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if out.AllowFavoriteToggleOnList != false {
		t.Errorf("out.AllowFavoriteToggleOnList: want false, got %v", out.AllowFavoriteToggleOnList)
	}

	if out.FavoriteFirstDefault != true {
		t.Errorf("out.FavoriteFirstDefault: want true, got %v", out.FavoriteFirstDefault)
	}

	if out.DefaultPeopleSort != "-last_contact" {
		t.Errorf("out.DefaultPeopleSort: want \"-last_contact\", got %q", out.DefaultPeopleSort)
	}

	// Get again and verify persistence.
	got, err := svc.Get(ctx)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.AllowFavoriteToggleOnList != false {
		t.Errorf("persisted AllowFavoriteToggleOnList: want false, got %v", got.AllowFavoriteToggleOnList)
	}

	if got.FavoriteFirstDefault != true {
		t.Errorf("persisted FavoriteFirstDefault: want true, got %v", got.FavoriteFirstDefault)
	}

	if got.DefaultPeopleSort != "-last_contact" {
		t.Errorf("persisted DefaultPeopleSort: want \"-last_contact\", got %q", got.DefaultPeopleSort)
	}
}

func TestSettings_Update_InvalidDefaultPeopleSort_ReturnsError(t *testing.T) {
	svc := settings.NewService(openTestDB(t))

	// Build a valid UserSettings but with an invalid DefaultPeopleSort.
	in := validBase()
	in.DefaultPeopleSort = "bogus"

	_, err := svc.Update(context.Background(), in)
	if !errors.Is(err, settings.ErrInvalidDefaultPeopleSort) {
		t.Errorf("want ErrInvalidDefaultPeopleSort, got %v", err)
	}
}
