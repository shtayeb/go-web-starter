-- name: ChangeAccountPassword :exec
UPDATE accounts SET password = ? WHERE id = ?;

-- name: CreateAccount :one
INSERT INTO accounts (account_id, user_id, password, provider_id)
VALUES ( ?, ?, ?, ?)
RETURNING *;

-- name: GetAccountByUserId :one
SELECT * FROM accounts WHERE user_id = ?;

-- name: GetAccountByUserIdAndProvider :one
SELECT * FROM accounts
WHERE user_id = ? AND provider_id = ?;

-- name: GetAccountById :one
SELECT * FROM accounts WHERE id = ?;

-- name: UpdateAccountPassword :exec
UPDATE accounts SET password = ? WHERE id = ?;

-- name: UpdateAccountOAuthTokens :exec
UPDATE accounts
SET access_token = ?,
    refresh_token = ?,
    access_token_expires_at = ?
WHERE user_id = ? AND provider_id = ?;

-- name: DeleteAccountsByUserId :exec
DELETE FROM accounts WHERE user_id = ?;
