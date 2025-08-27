-- name: CreateToken :one
INSERT INTO tokens (hash,user_id,expiry,scope) VALUES (?,?,?,?) RETURNING *;

-- name: GetTokensForUser :one
SELECT * FROM tokens WHERE user_id = ?;

-- name: DeleteAllForUser :exec
DELETE FROM tokens WHERE scope = ? AND user_id = ?;

-- name: DeleteToken :exec
DELETE FROM tokens WHERE hash = ?;

-- name: DeleteTokensByUserId :exec
DELETE FROM tokens WHERE user_id = ?;
