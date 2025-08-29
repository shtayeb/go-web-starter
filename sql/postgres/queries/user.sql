-- name: CreateUser :one
INSERT INTO users (name,email,email_verified,image)
VALUES ($1, $2, $3,$4) RETURNING *;

-- name: UpdateUserNameAndImage :one
UPDATE users SET name = $1, image = $2 WHERE id = $3 RETURNING *;

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

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
