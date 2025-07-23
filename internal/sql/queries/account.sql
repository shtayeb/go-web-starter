-- name: ChangeAccountPassword :exec
UPDATE account SET password = ? WHERE id = ?;

-- name: CreateAccount :one
INSERT INTO account (account_id,user_id,password,created_at,updated_at)
VALUES ( ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetAccountByUserId :one
SELECT * FROM account WHERE user_id=?;

-- name: GetAccountById :one
SELECT * FROM account WHERE id = ?;