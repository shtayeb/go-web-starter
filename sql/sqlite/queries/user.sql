-- name: CreateUser :one
INSERT INTO users (name,email,email_verified,image)
VALUES (?, ?, ?, ?) RETURNING *;

-- name: UpdateUserNameAndImage :one
UPDATE users SET name = ?, image = ? WHERE id = ? RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: GetUserById :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByToken :one
SELECT
    sqlc.embed(users)
FROM users
    INNER JOIN tokens ON users.id = tokens.user_id
WHERE tokens.hash = ?
    AND tokens.scope = ?
    AND tokens.expiry > ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;
