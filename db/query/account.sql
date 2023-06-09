-- name: CreateAccount :one
INSERT INTO accounts(
  owner,
  balance,
  currency
) VALUES (
  $1, $2, $3
) RETURNING accounts.*;

-- name: GetAccount :one
SELECT
  accounts.*
FROM accounts
WHERE accounts.id = $1;

-- name: GetAccountForUpdate :one
SELECT
  accounts.*
FROM accounts
WHERE accounts.id = $1
FOR NO KEY UPDATE;

-- name: ListAccounts :many
SELECT
  accounts.*
FROM accounts
ORDER BY accounts.id
LIMIT $1 OFFSET $2;

-- name: UpdateAccount :one
UPDATE accounts
SET
  balance = $2
WHERE accounts.id = $1
RETURNING accounts.*;

-- name: AddAccountBalance :one
UPDATE accounts
SET
  balance = balance + sqlc.arg(amount)
WHERE accounts.id = sqlc.arg(id)
RETURNING accounts.*;

-- name: DeleteAccount :exec
DELETE
FROM accounts
WHERE accounts.id = $1;