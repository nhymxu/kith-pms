-- Migrate any birthday rows from important_date into person.date_of_birth (where not already set),
-- then remove birthday kind from important_date entirely.

UPDATE person
SET date_of_birth = (
    SELECT date_value
    FROM important_date
    WHERE important_date.person_id = person.id
      AND important_date.kind = 'birthday'
    LIMIT 1
)
WHERE (date_of_birth IS NULL OR date_of_birth = '')
  AND EXISTS (
    SELECT 1 FROM important_date
    WHERE important_date.person_id = person.id
      AND important_date.kind = 'birthday'
  );

DELETE FROM important_date WHERE kind = 'birthday';
