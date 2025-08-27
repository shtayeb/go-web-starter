-- (Removed; use UpdateAccountPassword)
-- name: CreateAccount :one
INSERT INTO accounts (account_id, user_id, password, provider_id)
VALUES ( ?, ?, ?, ?)
ON CONFLICT(user_id, provider_id) DO NOTHING
RETURNING id, account_id, user_id, provider_id;

-- Return multiple accounts for a user and only known columns
-- name: GetAccountByUserId :many
SELECT id,
       account_id,
       user_id,
       provider_id,
       access_token,
       refresh_token,
       access_token_expires_at,
       created_at,
       updated_at
FROM accounts
WHERE user_id = ?;

-- name: GetAccountByUserIdAndProvider :one
SELECT id, account_id, user_id, provider_id
FROM accounts
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
