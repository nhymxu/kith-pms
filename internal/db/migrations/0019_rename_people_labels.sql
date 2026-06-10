ALTER TABLE label RENAME TO people_label;
ALTER TABLE person_label RENAME TO people_label_assignment;
DROP INDEX IF EXISTS idx_person_label_label;
CREATE INDEX idx_people_label_assignment_label ON people_label_assignment(label_id);
