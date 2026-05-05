ALTER TABLE person ADD COLUMN last_contact_at TEXT;
CREATE INDEX idx_person_last_contact ON person(last_contact_at) WHERE last_contact_at IS NOT NULL;
