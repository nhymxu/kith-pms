package testutil

import (
	"testing"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/db"
)

// NewDB opens an in-memory SQLite bun.DB, runs all migrations,
// and registers t.Cleanup to close it.
func NewDB(t *testing.T) *bun.DB {
	t.Helper()

	bunDB, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("testutil.NewDB: %v", err)
	}

	if err := db.Up(bunDB); err != nil {
		t.Fatalf("testutil.NewDB migrate: %v", err)
	}

	t.Cleanup(func() { _ = bunDB.Close() })

	return bunDB
}
