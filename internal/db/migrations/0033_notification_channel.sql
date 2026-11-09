CREATE TABLE notification_channel (
  id         INTEGER PRIMARY KEY,
  kind       TEXT NOT NULL,
  name       TEXT NOT NULL DEFAULT '',
  enabled    INTEGER NOT NULL DEFAULT 1,
  config_enc BLOB NOT NULL DEFAULT x'',
  enabled_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
