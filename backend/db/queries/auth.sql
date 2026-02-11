-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = sqlc.arg(email)
LIMIT 1;

-- name: InsertRefreshToken :one
INSERT INTO refresh_tokens (
  user_id,
  token_hash,
  expires_at
) VALUES (
  sqlc.arg(user_id),
  sqlc.arg(token_hash),
  sqlc.arg(expires_at)
)
RETURNING *;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE token_hash = sqlc.arg(token_hash)
  AND revoked_at IS NULL;
