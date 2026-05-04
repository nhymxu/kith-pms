CREATE TABLE work_history (
  id INTEGER PRIMARY KEY,
  person_id INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
  company TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL DEFAULT '',
  start_date TEXT NOT NULL DEFAULT '',
  end_date TEXT NOT NULL DEFAULT '',
  location TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  position INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX idx_work_history_person ON work_history(person_id);
