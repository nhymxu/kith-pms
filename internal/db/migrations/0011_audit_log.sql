CREATE TABLE audit_log (
  id          INTEGER PRIMARY KEY,
  entity_type TEXT    NOT NULL,
  entity_id   INTEGER NOT NULL,
  entity_name TEXT    NOT NULL DEFAULT '',
  action      TEXT    NOT NULL,
  actor_id    INTEGER,
  created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX idx_audit_log_entity  ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_actor   ON audit_log(actor_id);
CREATE INDEX idx_audit_log_created ON audit_log(created_at DESC);
