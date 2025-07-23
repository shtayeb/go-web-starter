-- name: GetSessionByToken :one
SELECT * FROM session WHERE token = ?;