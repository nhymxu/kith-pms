package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrationStatus describes a single migration entry.
type MigrationStatus struct {
	Version   int
	Name      string
	AppliedAt string // empty string means pending
}

// Up applies all unapplied migrations in ascending version order.
// Each migration runs in its own transaction; a failure rolls back that
// migration only and returns the error immediately.
func Up(db *sql.DB) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	files, err := loadMigrationFiles()
	if err != nil {
		return err
	}

	applied, err := appliedVersions(db)
	if err != nil {
		return err
	}

	for _, mf := range files {
		if applied[mf.version] {
			continue
		}

		if err := applyMigration(db, mf); err != nil {
			return fmt.Errorf("migrations: applying %04d_%s: %w", mf.version, mf.name, err)
		}
	}

	return nil
}

// Status returns the list of all known migrations with their applied/pending state.
func Status(db *sql.DB) ([]MigrationStatus, error) {
	if err := ensureMigrationsTable(db); err != nil {
		return nil, err
	}

	files, err := loadMigrationFiles()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`SELECT version, name, applied_at FROM _schema_migrations ORDER BY version`,
	)
	if err != nil {
		return nil, fmt.Errorf("migrations: query status: %w", err)
	}
	defer func() { _ = rows.Close() }()

	appliedMap := map[int]string{}

	for rows.Next() {
		var (
			ver             int
			name, appliedAt string
		)
		if err := rows.Scan(&ver, &name, &appliedAt); err != nil {
			return nil, fmt.Errorf("migrations: scan status row: %w", err)
		}

		appliedMap[ver] = appliedAt
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("migrations: iterate status rows: %w", err)
	}

	out := make([]MigrationStatus, 0, len(files))
	for _, mf := range files {
		out = append(out, MigrationStatus{
			Version:   mf.version,
			Name:      mf.name,
			AppliedAt: appliedMap[mf.version],
		})
	}

	return out, nil
}

// ---- internal helpers -------------------------------------------------------

type migrationFile struct {
	version  int
	name     string
	filename string
}

func loadMigrationFiles() ([]migrationFile, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("migrations: read dir: %w", err)
	}

	var files []migrationFile

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}

		mf, err := parseMigrationFilename(e.Name())
		if err != nil {
			return nil, err
		}

		files = append(files, mf)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].version < files[j].version
	})

	return files, nil
}

// parseMigrationFilename parses a filename like "0001_init.sql" into version=1, name="init".
func parseMigrationFilename(filename string) (migrationFile, error) {
	base := strings.TrimSuffix(filename, ".sql")

	idx := strings.Index(base, "_")
	if idx < 0 {
		return migrationFile{}, fmt.Errorf("migrations: invalid filename (no underscore): %q", filename)
	}

	versionStr := base[:idx]
	name := base[idx+1:]

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return migrationFile{}, fmt.Errorf("migrations: non-numeric version in %q: %w", filename, err)
	}

	return migrationFile{version: version, name: name, filename: filename}, nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS _schema_migrations (
			version    INTEGER PRIMARY KEY,
			name       TEXT    NOT NULL,
			applied_at TEXT    NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("migrations: ensure table: %w", err)
	}

	return nil
}

func appliedVersions(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query(`SELECT version FROM _schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("migrations: query applied: %w", err)
	}
	defer func() { _ = rows.Close() }()

	m := map[int]bool{}

	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("migrations: scan version: %w", err)
		}

		m[v] = true
	}

	return m, rows.Err()
}

func applyMigration(db *sql.DB, mf migrationFile) error {
	content, err := migrationsFS.ReadFile("migrations/" + mf.filename)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if _, err := tx.Exec(string(content)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("exec sql: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO _schema_migrations (version, name, applied_at) VALUES (?, ?, ?)`,
		mf.version, mf.name, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration: %w", err)
	}

	return tx.Commit()
}
