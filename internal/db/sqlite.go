package db

import (
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	// sqliteshim registers "sqlite" driver, auto-selects modernc when CGO is disabled.
	_ "github.com/uptrace/bun/driver/sqliteshim"
)

// Open opens (or creates) the SQLite database at path, applies recommended
// PRAGMAs, and returns a *bun.DB wrapping the underlying connection.
//
// PRAGMAs always applied:
//   - journal_mode=WAL   — concurrent readers without blocking the writer
//   - foreign_keys=ON    — enforce referential integrity
//   - synchronous=NORMAL — safe with WAL; good balance of durability vs speed
//
// When maxOpenConns > 1, busy_timeout=5000 is also applied so SQLite retries
// for up to 5 s on write contention instead of returning SQLITE_BUSY immediately.
// At maxOpenConns=1 (default) the Go pool serialises everything, so busy_timeout
// is unnecessary. See docs/system-architecture.md for the full tradeoff.
func Open(path string, maxOpenConns int) (*bun.DB, error) {
	sqldb, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("db: open %q: %w", path, err)
	}

	sqldb.SetMaxOpenConns(maxOpenConns)

	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA synchronous=NORMAL;",
	}
	if maxOpenConns > 1 {
		pragmas = append(pragmas, "PRAGMA busy_timeout=5000;")
	}

	for _, p := range pragmas {
		if _, err := sqldb.Exec(p); err != nil {
			_ = sqldb.Close()
			return nil, fmt.Errorf("db: exec %q: %w", p, err)
		}
	}

	return bun.NewDB(sqldb, sqlitedialect.New()), nil
}
