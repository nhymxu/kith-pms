CREATE TABLE gift (
  id              INTEGER PRIMARY KEY,
  person_id       INTEGER NOT NULL REFERENCES person(id) ON DELETE CASCADE,
  title           TEXT    NOT NULL,
  direction       TEXT    NOT NULL DEFAULT 'planned',
  date            TEXT,
  notes           TEXT    NOT NULL DEFAULT '',
  amount_cents    INTEGER,
  currency        TEXT    NOT NULL DEFAULT 'USD',
  debt_type       TEXT    NOT NULL DEFAULT '',
  image_path      TEXT    NOT NULL DEFAULT '',
  image_mime_type TEXT    NOT NULL DEFAULT '',
  created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  updated_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);

CREATE INDEX idx_gift_person    ON gift(person_id);
CREATE INDEX idx_gift_direction ON gift(direction);
CREATE INDEX idx_gift_debt_type ON gift(debt_type) WHERE debt_type != '';
