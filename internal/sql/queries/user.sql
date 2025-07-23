-- name: CreateUser :one
INSERT INTO user (name,email,created_at,updated_at)
VALUES (
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: UpdateUserName :exec
UPDATE user SET name = ? WHERE id = ?;

-- name: GetUserByEmail :one
SELECT * FROM user WHERE email = ?;

-- name: GetUserById :one
SELECT * FROM user WHERE id = ?;