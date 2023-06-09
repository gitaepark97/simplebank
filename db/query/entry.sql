-- name: CreateEntry :one
INSERT INTO entries(
  account_id,
  amount
) VALUES (
  $1, $2
) RETURNING entries.*;

-- name: GetEntry :one
SELECT
  entries.*
FROM entries
WHERE entries.id = $1;

-- name: ListEntries :many
SELECT
  entries.*
FROM entries
WHERE entries.account_id = $1
ORDER BY entries.id
LIMIT $2 OFFSET $3;