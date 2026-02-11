-- name: CreateAuditLog :one
INSERT INTO audit_logs (
  tenant_id,
  actor_user_id,
  action,
  entity_type,
  entity_id,
  metadata,
  ip_address,
  user_agent
) VALUES (
  sqlc.arg(tenant_id),
  sqlc.narg(actor_user_id),
  sqlc.arg(action),
  sqlc.narg(entity_type),
  sqlc.narg(entity_id),
  coalesce(sqlc.narg(metadata), '{}'::jsonb),
  sqlc.narg(ip_address),
  sqlc.narg(user_agent)
)
RETURNING *;

-- name: ListAuditLogs :many
SELECT *
FROM audit_logs
WHERE tenant_id = sqlc.arg(tenant_id)
  AND (sqlc.narg(action)::audit_action_enum IS NULL OR action = sqlc.narg(action))
  AND (sqlc.narg(actor_user_id)::uuid IS NULL OR actor_user_id = sqlc.narg(actor_user_id))
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);
