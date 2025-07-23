-- name: ChangeAccountPassword :exec
UPDATE accounts SET password = $1 WHERE id = $2;

-- name: CreateAccount :one
INSERT INTO accounts (account_id,user_id,password,created_at,updated_at)
VALUES ( $1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAccountByUserId :one
SELECT * FROM accounts WHERE user_id = $1;

-- name: GetAccountById :one
SELECT * FROM accounts WHERE id = $1;