ALTER TABLE person ADD COLUMN is_self INTEGER NOT NULL DEFAULT 0;
CREATE UNIQUE INDEX idx_person_is_self ON person (is_self) WHERE is_self = 1;
