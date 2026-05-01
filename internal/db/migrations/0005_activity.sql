CREATE TABLE activity (
  id INTEGER PRIMARY KEY,
  title TEXT NOT NULL,
  occurred_at_date TEXT NOT NULL,
  occurred_at_time TEXT,
  content TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX idx_activity_occurred_at ON activity(occurred_at_date DESC, id DESC);

CREATE TABLE activity_person (
  activity_id INTEGER NOT NULL REFERENCES activity(id) ON DELETE CASCADE,
  person_id   INTEGER NOT NULL REFERENCES person(id)   ON DELETE CASCADE,
  PRIMARY KEY (activity_id, person_id)
);
CREATE INDEX idx_activity_person_person ON activity_person(person_id);
