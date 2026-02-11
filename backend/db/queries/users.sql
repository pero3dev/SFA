-- name: ListTenantUsers :many
SELECT
  u.id,
  u.email,
  u.display_name,
  u.is_active,
  m.role,
  u.created_at,
  u.updated_at
FROM memberships m
JOIN users u ON u.id = m.user_id
WHERE m.tenant_id = sqlc.arg(tenant_id)
  AND (sqlc.narg(role)::role_enum IS NULL OR m.role = sqlc.narg(role))
ORDER BY u.created_at DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: CountTenantUsers :one
SELECT count(*)::bigint
FROM memberships
WHERE tenant_id = sqlc.arg(tenant_id)
  AND (sqlc.narg(role)::role_enum IS NULL OR role = sqlc.narg(role));

-- name: CreateMembership :one
INSERT INTO memberships (
  tenant_id,
  user_id,
  role
) VALUES (
  sqlc.arg(tenant_id),
  sqlc.arg(user_id),
  sqlc.arg(role)
)
RETURNING *;

-- name: UpdateMembershipRole :one
UPDATE memberships
SET role = sqlc.arg(role), updated_at = now()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND user_id = sqlc.arg(user_id)
RETURNING *;
