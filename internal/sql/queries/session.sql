-- name: GetSessionByToken :one
SELECT * FROM sessions WHERE token = $1;