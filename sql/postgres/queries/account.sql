-- name: CreateAccount :one
INSERT INTO accounts (account_id, user_id, password, provider_id)
VALUES ( $1, $2, $3, $4)
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

-- name: UpdateAccountOAuthTokens :exec
UPDATE accounts 
SET access_token = $1, 
    refresh_token = $2, 
    access_token_expires_at = $3
WHERE user_id = $4 AND provider_id = $5;

-- name: DeleteAccountsByUserId :exec
DELETE FROM accounts WHERE user_id = $1;
