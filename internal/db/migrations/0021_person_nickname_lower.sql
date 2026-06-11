ALTER TABLE person ADD COLUMN nickname_lower TEXT GENERATED ALWAYS AS (lower(nickname)) VIRTUAL;
CREATE INDEX idx_person_nickname_lower ON person(nickname_lower);
