ALTER TABLE person ADD COLUMN deleted_at TIMESTAMP NULL;
CREATE INDEX idx_person_deleted_at ON person (deleted_at) WHERE deleted_at IS NOT NULL;
