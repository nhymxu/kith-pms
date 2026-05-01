-- name: CreateSession :exec
INSERT INTO session (id, user_id, expires_at, last_seen_at, ip, user_agent)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetSession :one
SELECT id, user_id, expires_at, last_seen_at, ip, user_agent
FROM session
WHERE id = ?;

-- name: TouchSession :exec
UPDATE session SET expires_at = ?, last_seen_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE id = ?;

-- name: DeleteSession :exec
DELETE FROM session WHERE id = ?;

-- name: DeleteAllSessionsForUser :exec
DELETE FROM session WHERE user_id = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM session WHERE expires_at < strftime('%Y-%m-%dT%H:%M:%fZ','now');
