-- name: CreateToken :one
INSERT INTO tokens (hash,user_id,expiry,scope) VALUES ($1,$2,$3,$4);

-- name: GetTokensForUser :one
SELECT * FROM tokens WHERE user_id = $1;

-- name: DeleteAllForUser :one
DELETE FROM tokens WHERE scope = $1 AND user_id = $2;