CREATE TABLE relationship_type (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  reverse_name TEXT NOT NULL DEFAULT '',
  inverse_type_id INTEGER REFERENCES relationship_type(id) ON DELETE SET NULL,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE TABLE person_relationship (
  id INTEGER PRIMARY KEY,
  from_person_id INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
  to_person_id   INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
  relationship_type_id INTEGER NOT NULL REFERENCES relationship_type(id) ON DELETE RESTRICT,
  notes TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  UNIQUE (from_person_id, to_person_id, relationship_type_id),
  CHECK (from_person_id <> to_person_id)
);
CREATE INDEX idx_person_relationship_from ON person_relationship(from_person_id);
CREATE INDEX idx_person_relationship_to   ON person_relationship(to_person_id);
CREATE INDEX idx_person_relationship_type ON person_relationship(relationship_type_id);
