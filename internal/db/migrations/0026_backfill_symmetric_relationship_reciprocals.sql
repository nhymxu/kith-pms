-- Backfill missing reciprocal rows for relationship types with no distinct inverse type
-- (e.g. "Friend"). AttachRelationship/BulkAttach previously skipped creating the reciprocal
-- row whenever reverse_name was blank, so person B never saw the relationship on their profile.
INSERT OR IGNORE INTO person_relationship (from_person_id, to_person_id, relationship_type_id, notes)
SELECT pr.to_person_id, pr.from_person_id, pr.relationship_type_id, pr.notes
FROM person_relationship pr
JOIN relationship_type rt ON rt.id = pr.relationship_type_id
WHERE rt.inverse_type_id IS NULL;
