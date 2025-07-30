-- name: CreateUser :one
INSERT INTO users (name,email,created_at,updated_at)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateUserName :exec
UPDATE users SET name = $1 WHERE id = $2;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByToken :one
SELECT 
    sqlc.embed(users)
FROM users
    INNER JOIN tokens ON users.id = tokens.user_id
WHERE tokens.hash = $1
    AND tokens.scope = $2
    AND tokens.expiry > $3;
