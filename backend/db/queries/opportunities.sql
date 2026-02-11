-- name: ListOpportunities :many
SELECT *
FROM opportunities
WHERE tenant_id = sqlc.arg(tenant_id)
  AND (sqlc.narg(stage)::opportunity_stage_enum IS NULL OR stage = sqlc.narg(stage))
  AND (sqlc.narg(owner_user_id)::uuid IS NULL OR owner_user_id = sqlc.narg(owner_user_id))
ORDER BY updated_at DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: CountOpportunities :one
SELECT count(*)::bigint
FROM opportunities
WHERE tenant_id = sqlc.arg(tenant_id)
  AND (sqlc.narg(stage)::opportunity_stage_enum IS NULL OR stage = sqlc.narg(stage))
  AND (sqlc.narg(owner_user_id)::uuid IS NULL OR owner_user_id = sqlc.narg(owner_user_id));

-- name: CreateOpportunity :one
INSERT INTO opportunities (
  tenant_id,
  account_id,
  contact_id,
  owner_user_id,
  name,
  stage,
  probability,
  amount,
  expected_close_date,
  memo,
  created_by
) VALUES (
  sqlc.arg(tenant_id),
  sqlc.arg(account_id),
  sqlc.narg(contact_id),
  sqlc.arg(owner_user_id),
  sqlc.arg(name),
  coalesce(sqlc.narg(stage)::opportunity_stage_enum, 'new_lead'),
  coalesce(sqlc.narg(probability), 0),
  coalesce(sqlc.narg(amount), 0),
  sqlc.narg(expected_close_date),
  sqlc.narg(memo),
  sqlc.arg(created_by)
)
RETURNING *;

-- name: UpdateOpportunity :one
UPDATE opportunities
SET
  contact_id = coalesce(sqlc.narg(contact_id), contact_id),
  owner_user_id = coalesce(sqlc.narg(owner_user_id), owner_user_id),
  name = coalesce(sqlc.narg(name), name),
  stage = coalesce(sqlc.narg(stage)::opportunity_stage_enum, stage),
  probability = coalesce(sqlc.narg(probability), probability),
  amount = coalesce(sqlc.narg(amount), amount),
  expected_close_date = coalesce(sqlc.narg(expected_close_date), expected_close_date),
  memo = coalesce(sqlc.narg(memo), memo),
  updated_at = now()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(opportunity_id)
RETURNING *;

-- name: ListActivitiesByOpportunity :many
SELECT *
FROM activities
WHERE tenant_id = sqlc.arg(tenant_id)
  AND opportunity_id = sqlc.arg(opportunity_id)
ORDER BY activity_at DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: CreateActivity :one
INSERT INTO activities (
  tenant_id,
  opportunity_id,
  activity_type,
  subject,
  detail,
  activity_at,
  created_by
) VALUES (
  sqlc.arg(tenant_id),
  sqlc.arg(opportunity_id),
  sqlc.arg(activity_type),
  sqlc.arg(subject),
  sqlc.narg(detail),
  sqlc.arg(activity_at),
  sqlc.arg(created_by)
)
RETURNING *;

-- name: ListQuotesByOpportunity :many
SELECT *
FROM quotes
WHERE tenant_id = sqlc.arg(tenant_id)
  AND opportunity_id = sqlc.arg(opportunity_id)
ORDER BY created_at DESC;

-- name: CreateQuote :one
INSERT INTO quotes (
  tenant_id,
  opportunity_id,
  quote_no,
  amount,
  status,
  issued_on,
  valid_until,
  note,
  created_by
) VALUES (
  sqlc.arg(tenant_id),
  sqlc.arg(opportunity_id),
  sqlc.arg(quote_no),
  sqlc.arg(amount),
  coalesce(sqlc.narg(status)::quote_status_enum, 'draft'),
  sqlc.narg(issued_on),
  sqlc.narg(valid_until),
  sqlc.narg(note),
  sqlc.arg(created_by)
)
RETURNING *;

-- name: ListOrdersByOpportunity :many
SELECT *
FROM orders
WHERE tenant_id = sqlc.arg(tenant_id)
  AND opportunity_id = sqlc.arg(opportunity_id)
ORDER BY created_at DESC;

-- name: CreateOrder :one
INSERT INTO orders (
  tenant_id,
  opportunity_id,
  order_no,
  amount,
  status,
  ordered_on,
  note,
  created_by
) VALUES (
  sqlc.arg(tenant_id),
  sqlc.arg(opportunity_id),
  sqlc.arg(order_no),
  sqlc.arg(amount),
  coalesce(sqlc.narg(status)::order_status_enum, 'pending'),
  sqlc.narg(ordered_on),
  sqlc.narg(note),
  sqlc.arg(created_by)
)
RETURNING *;

-- name: CloseOpportunityAsLost :exec
UPDATE opportunities
SET stage = 'closed_lost',
    closed_at = now(),
    updated_at = now()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(opportunity_id);

-- name: CreateOpportunityLoss :one
INSERT INTO opportunity_losses (
  tenant_id,
  opportunity_id,
  reason,
  detail,
  lost_at,
  created_by
)
VALUES (
  sqlc.arg(tenant_id),
  sqlc.arg(opportunity_id),
  sqlc.arg(reason),
  sqlc.narg(detail),
  coalesce(sqlc.narg(lost_at), now()),
  sqlc.arg(created_by)
)
RETURNING *;
