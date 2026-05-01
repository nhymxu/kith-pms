CREATE TABLE reminder (
  id INTEGER PRIMARY KEY,
  title TEXT NOT NULL,
  notes TEXT NOT NULL DEFAULT '',
  due_date TEXT NOT NULL,
  person_id INTEGER REFERENCES person(id) ON DELETE SET NULL,
  important_date_id INTEGER REFERENCES important_date(id) ON DELETE SET NULL,
  completed INTEGER NOT NULL DEFAULT 0,
  completed_at TEXT,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX idx_reminder_due_date ON reminder(due_date);
CREATE INDEX idx_reminder_person ON reminder(person_id);
CREATE INDEX idx_reminder_completed ON reminder(completed);
