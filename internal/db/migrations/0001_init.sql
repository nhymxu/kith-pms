-- Bootstrap migration: creates the schema migrations tracking table.
-- This table records which migrations have been applied.
CREATE TABLE IF NOT EXISTS _schema_migrations (
    version    INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL,
    applied_at TEXT    NOT NULL
);
