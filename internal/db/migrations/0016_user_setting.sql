CREATE TABLE IF NOT EXISTS user_setting (
    key        TEXT PRIMARY KEY NOT NULL,
    value      TEXT NOT NULL DEFAULT '',
    updated_at TEXT NOT NULL
);
