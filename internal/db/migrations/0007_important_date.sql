CREATE TABLE important_date (
  id INTEGER PRIMARY KEY,
  person_id INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
  kind TEXT NOT NULL DEFAULT 'other',
  label TEXT NOT NULL DEFAULT '',
  date_value TEXT NOT NULL,
  recurring INTEGER NOT NULL DEFAULT 1,
  notes TEXT NOT NULL DEFAULT '',
  position INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  month_day TEXT GENERATED ALWAYS AS (substr(date_value, length(date_value) - 4)) VIRTUAL
);

CREATE INDEX idx_important_date_person ON important_date(person_id);
CREATE INDEX idx_important_date_month_day ON important_date(month_day);
