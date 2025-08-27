-- name: GetSessionByToken :one
SELECT * FROM sessions WHERE token = ?;
