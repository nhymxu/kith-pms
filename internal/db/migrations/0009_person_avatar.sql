-- Add avatar fields to person table
ALTER TABLE person ADD COLUMN avatar_path TEXT DEFAULT '';
ALTER TABLE person ADD COLUMN avatar_mime_type TEXT DEFAULT '';
ALTER TABLE person ADD COLUMN avatar_size INTEGER DEFAULT 0;
ALTER TABLE person ADD COLUMN avatar_uploaded_at TEXT;

-- Index for avatar queries
CREATE INDEX idx_person_avatar ON person(avatar_path) WHERE avatar_path != '';
