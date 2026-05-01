CREATE TABLE person (
  id INTEGER PRIMARY KEY,
  prefix TEXT NOT NULL DEFAULT '',
  name TEXT NOT NULL,
  nickname TEXT NOT NULL DEFAULT '',
  date_of_birth TEXT,
  relationship_type TEXT NOT NULL DEFAULT '',
  other_notes TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  name_lower TEXT GENERATED ALWAYS AS (lower(name)) VIRTUAL
);
CREATE INDEX idx_person_name_lower ON person(name_lower);

CREATE TABLE contact_info (
  id INTEGER PRIMARY KEY,
  person_id INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  value TEXT NOT NULL,
  label TEXT NOT NULL DEFAULT '',
  position INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_contact_info_person ON contact_info(person_id);

CREATE TABLE location (
  id INTEGER PRIMARY KEY,
  person_id INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  address TEXT NOT NULL DEFAULT '',
  city TEXT NOT NULL DEFAULT '',
  country TEXT NOT NULL DEFAULT '',
  postal_code TEXT NOT NULL DEFAULT '',
  position INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_location_person ON location(person_id);
