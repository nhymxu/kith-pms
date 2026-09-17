ALTER TABLE person ADD COLUMN archived_at TIMESTAMP NULL;
CREATE INDEX idx_person_archived_at ON person (archived_at) WHERE archived_at IS NOT NULL;
