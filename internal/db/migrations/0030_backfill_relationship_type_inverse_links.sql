-- Backfill inverse_type_id for relationship types whose reverse_name names another existing
-- type that CreateType previously failed to link (silently skipped when the reverse-name type
-- already existed, e.g. "Child" created standalone before "Parent" set reverse_name="Child").
-- Only links pairs where the named partner isn't already paired with a different type.
UPDATE relationship_type
SET inverse_type_id = (
    SELECT other.id FROM relationship_type other
    WHERE other.name = relationship_type.reverse_name
      AND other.id != relationship_type.id
)
WHERE inverse_type_id IS NULL
  AND reverse_name != ''
  AND reverse_name != name
  AND (
    SELECT other.inverse_type_id FROM relationship_type other
    WHERE other.name = relationship_type.reverse_name
      AND other.id != relationship_type.id
  ) IS NULL;

-- Back-link the partner row so both sides point at each other.
UPDATE relationship_type
SET inverse_type_id = (
    SELECT other.id FROM relationship_type other WHERE other.inverse_type_id = relationship_type.id
)
WHERE inverse_type_id IS NULL
  AND EXISTS (SELECT 1 FROM relationship_type other WHERE other.inverse_type_id = relationship_type.id);
