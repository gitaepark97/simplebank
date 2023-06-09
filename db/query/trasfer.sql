-- name: CreateTransfer :one
INSERT INTO transfers(
  from_account_id,
  to_account_id,
  amount
) VALUES (
  $1, $2, $3
) RETURNING transfers.*;

-- name: GetTransfer :one
SELECT
  transfers.*
FROM transfers
WHERE transfers.id = $1;

-- name: ListTransfers :many
SELECT
  transfers.*
FROM transfers
WHERE transfers.from_account_id = $1
  OR transfers.to_account_id = $2
ORDER BY transfers.id
LIMIT $3 OFFSET $4;