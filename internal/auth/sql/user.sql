-- name: GetUser :one
SELECT id, password_hash, created_at, updated_at
FROM user
LIMIT 1;

-- name: UpsertUser :exec
INSERT INTO user (id, password_hash, updated_at)
VALUES (1, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now'))
ON CONFLICT(id) DO UPDATE SET
    password_hash = excluded.password_hash,
    updated_at    = excluded.updated_at;

-- name: UpdatePasswordHash :exec
UPDATE user SET password_hash = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE id = ?;

-- name: ClearPassword :exec
UPDATE user SET password_hash = '', updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE id = ?;
