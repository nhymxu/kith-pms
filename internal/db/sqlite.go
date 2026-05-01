package db

import (
	"database/sql"
	"fmt"

	// modernc.org/sqlite registers itself as driver "sqlite" — pure Go, no CGO.
	_ "modernc.org/sqlite"
)

// Open opens (or creates) the SQLite database at path, applies recommended
// PRAGMAs, and returns a ready-to-use *sql.DB.
//
// PRAGMAs applied:
//   - journal_mode=WAL   — concurrent readers without blocking the writer
//   - foreign_keys=ON    — enforce referential integrity
//   - synchronous=NORMAL — safe with WAL; good balance of durability vs speed
//
// MaxOpenConns is set to 1 to serialise all writes (SQLite single-writer model).
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("db: open %q: %w", path, err)
	}

	// Single writer — avoid SQLITE_BUSY on concurrent writes.
	db.SetMaxOpenConns(1)

	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA synchronous=NORMAL;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("db: exec %q: %w", p, err)
		}
	}

	return db, nil
}
