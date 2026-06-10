CREATE TABLE journal_label (
  id         INTEGER PRIMARY KEY,
  name       TEXT    NOT NULL UNIQUE,
  color      TEXT    NOT NULL DEFAULT '#9ea096',
  created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE journal_label_assignment (
  activity_id INTEGER NOT NULL REFERENCES activity(id)       ON DELETE CASCADE,
  label_id    INTEGER NOT NULL REFERENCES journal_label(id)  ON DELETE CASCADE,
  PRIMARY KEY (activity_id, label_id)
);

CREATE INDEX idx_journal_label_assignment_label ON journal_label_assignment(label_id);
