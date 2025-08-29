-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE token = $1
  AND revoked_at IS NULL
  AND expires_at > NOW();
