-- name: ChangeAccountPassword :exec
UPDATE accounts SET password = $1 WHERE id = $2;

-- name: CreateAccount :one
INSERT INTO accounts (account_id,user_id,password,provider_id,created_at,updated_at)
VALUES ( $1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAccountByUserId :one
SELECT * FROM accounts WHERE user_id = $1;

-- name: GetAccountByUserIdAndProvider :one
SELECT * FROM accounts
WHERE user_id = $1 AND provider_id = $2;

-- name: GetAccountById :one
SELECT * FROM accounts WHERE id = $1;

-- name: UpdateAccountPassword :exec
UPDATE accounts SET password = $1 WHERE id = $2;