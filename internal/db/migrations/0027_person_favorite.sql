ALTER TABLE person ADD COLUMN is_favorite INTEGER NOT NULL DEFAULT 0;
CREATE INDEX idx_person_is_favorite ON person (is_favorite) WHERE is_favorite = 1;
