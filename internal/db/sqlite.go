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
// PRAGMAs applied:
//   - journal_mode=WAL   — concurrent readers without blocking the writer
//   - foreign_keys=ON    — enforce referential integrity
//   - synchronous=NORMAL — safe with WAL; good balance of durability vs speed
//
// MaxOpenConns is set to 1 to serialise all writes (SQLite single-writer model).
func Open(path string) (*bun.DB, error) {
	sqldb, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("db: open %q: %w", path, err)
	}

	// Single writer — avoid SQLITE_BUSY on concurrent writes.
	sqldb.SetMaxOpenConns(1)

	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA synchronous=NORMAL;",
	}
	for _, p := range pragmas {
		if _, err := sqldb.Exec(p); err != nil {
			_ = sqldb.Close()
			return nil, fmt.Errorf("db: exec %q: %w", p, err)
		}
	}

	return bun.NewDB(sqldb, sqlitedialect.New()), nil
}
