-- Remove the legacy freeform relationship_type column from person.
-- The structured people-to-people relationship system (person_relationship +
-- relationship_type tables) supersedes this field. Existing values are discarded.
ALTER TABLE person DROP COLUMN relationship_type;
