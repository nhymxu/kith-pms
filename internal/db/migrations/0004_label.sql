CREATE TABLE label (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  color TEXT NOT NULL DEFAULT '#9ea096',
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE person_label (
  person_id INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
  label_id  INTEGER NOT NULL REFERENCES label(id)  ON DELETE CASCADE,
  PRIMARY KEY (person_id, label_id)
);
CREATE INDEX idx_person_label_label ON person_label(label_id);
