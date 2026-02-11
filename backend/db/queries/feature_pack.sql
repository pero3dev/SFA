-- name: ListNextActions :many
SELECT
  o.id,
  o.name,
  o.stage,
  o.next_action_at,
  o.next_action_note,
  a.name AS account_name
FROM opportunities o
JOIN accounts a ON a.id = o.account_id
WHERE o.tenant_id = sqlc.arg(tenant_id)
  AND o.next_action_at IS NOT NULL
  AND (sqlc.narg(due_before)::timestamptz IS NULL OR o.next_action_at <= sqlc.narg(due_before)::timestamptz)
ORDER BY o.next_action_at ASC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: UpdateOpportunityNextAction :one
UPDATE opportunities
SET
  next_action_at = sqlc.arg(next_action_at),
  next_action_note = sqlc.narg(next_action_note),
  updated_at = now()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(opportunity_id)
RETURNING *;

-- name: ListDealHealth :many
SELECT
  o.id,
  o.name,
  o.stage,
  o.probability,
  o.amount::double precision AS amount,
  coalesce(last_activity.last_activity_at, o.created_at)::timestamptz AS last_activity_at,
  GREATEST(
    0,
    LEAST(
      100,
      100
      - LEAST(60, EXTRACT(DAY FROM (now() - coalesce(last_activity.last_activity_at, o.created_at)))::int * 2)
      - LEAST(30, EXTRACT(DAY FROM (now() - o.updated_at))::int)
      + CASE
          WHEN o.probability >= 70 THEN 10
          WHEN o.probability <= 20 THEN -10
          ELSE 0
        END
    )
  )::int AS health_score
FROM opportunities o
LEFT JOIN LATERAL (
  SELECT max(a.activity_at) AS last_activity_at
  FROM activities a
  WHERE a.tenant_id = o.tenant_id
    AND a.opportunity_id = o.id
) AS last_activity ON true
WHERE o.tenant_id = sqlc.arg(tenant_id)
ORDER BY health_score ASC, o.updated_at ASC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: GetForecastSummary :many
SELECT
  o.owner_user_id,
  to_char(date_trunc('month', coalesce(o.expected_close_date::timestamptz, now())), 'YYYY-MM') AS month_bucket,
  count(*)::bigint AS deal_count,
  coalesce(sum(o.amount), 0)::double precision AS pipeline_amount,
  coalesce(sum(o.amount * (o.probability::numeric / 100.0)), 0)::double precision AS weighted_amount
FROM opportunities o
WHERE o.tenant_id = sqlc.arg(tenant_id)
  AND o.stage NOT IN ('closed_won', 'closed_lost')
GROUP BY o.owner_user_id, to_char(date_trunc('month', coalesce(o.expected_close_date::timestamptz, now())), 'YYYY-MM')
ORDER BY month_bucket ASC, o.owner_user_id ASC;

-- name: GetLossReasonAnalysis :many
SELECT
  l.reason,
  count(*)::bigint AS lost_count,
  coalesce(sum(o.amount), 0)::double precision AS lost_amount
FROM opportunity_losses l
JOIN opportunities o ON o.id = l.opportunity_id
WHERE l.tenant_id = sqlc.arg(tenant_id)
GROUP BY l.reason
ORDER BY lost_count DESC, lost_amount DESC;

-- name: ListDuplicateCandidates :many
SELECT
  'account_name'::text AS duplicate_type,
  a1.id AS primary_id,
  a2.id AS duplicate_id,
  a1.name AS match_value
FROM accounts a1
JOIN accounts a2
  ON a1.tenant_id = a2.tenant_id
 AND a1.id < a2.id
 AND lower(a1.name) = lower(a2.name)
WHERE a1.tenant_id = sqlc.arg(tenant_id)
UNION ALL
SELECT
  'account_website'::text AS duplicate_type,
  a1.id AS primary_id,
  a2.id AS duplicate_id,
  a1.website AS match_value
FROM accounts a1
JOIN accounts a2
  ON a1.tenant_id = a2.tenant_id
 AND a1.id < a2.id
 AND a1.website IS NOT NULL
 AND a2.website IS NOT NULL
 AND lower(a1.website) = lower(a2.website)
WHERE a1.tenant_id = sqlc.arg(tenant_id)
UNION ALL
SELECT
  'contact_email'::text AS duplicate_type,
  c1.id AS primary_id,
  c2.id AS duplicate_id,
  c1.email AS match_value
FROM contacts c1
JOIN contacts c2
  ON c1.tenant_id = c2.tenant_id
 AND c1.id < c2.id
 AND c1.email IS NOT NULL
 AND c2.email IS NOT NULL
 AND lower(c1.email) = lower(c2.email)
WHERE c1.tenant_id = sqlc.arg(tenant_id)
ORDER BY duplicate_type, match_value;

-- name: UpsertIntegrationConnection :one
INSERT INTO integration_connections (
  tenant_id,
  user_id,
  provider,
  integration_type,
  external_account_id,
  status,
  access_token,
  refresh_token,
  expires_at,
  scopes
) VALUES (
  sqlc.arg(tenant_id),
  sqlc.arg(user_id),
  sqlc.arg(provider),
  sqlc.arg(integration_type),
  sqlc.arg(external_account_id),
  coalesce(sqlc.narg(status)::integration_status_enum, 'active'),
  sqlc.narg(access_token),
  sqlc.narg(refresh_token),
  sqlc.narg(expires_at),
  coalesce(sqlc.narg(scopes), '{}'::text[])
)
ON CONFLICT (tenant_id, user_id, provider, integration_type, external_account_id)
DO UPDATE SET
  status = excluded.status,
  access_token = excluded.access_token,
  refresh_token = excluded.refresh_token,
  expires_at = excluded.expires_at,
  scopes = excluded.scopes,
  updated_at = now()
RETURNING *;

-- name: ListIntegrationConnections :many
SELECT *
FROM integration_connections
WHERE tenant_id = sqlc.arg(tenant_id)
ORDER BY updated_at DESC;

-- name: CreateIntegrationEvent :one
INSERT INTO integration_events (
  tenant_id,
  provider,
  integration_type,
  external_event_id,
  event_type,
  payload,
  linked_account_id,
  linked_contact_id,
  linked_opportunity_id,
  occurred_at
) VALUES (
  sqlc.arg(tenant_id),
  sqlc.arg(provider),
  sqlc.arg(integration_type),
  sqlc.narg(external_event_id),
  sqlc.arg(event_type),
  coalesce(sqlc.narg(payload), '{}'::jsonb),
  sqlc.narg(linked_account_id),
  sqlc.narg(linked_contact_id),
  sqlc.narg(linked_opportunity_id),
  sqlc.arg(occurred_at)
)
RETURNING *;

-- name: ListIntegrationEvents :many
SELECT *
FROM integration_events
WHERE tenant_id = sqlc.arg(tenant_id)
ORDER BY occurred_at DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: CreateApprovalRequest :one
INSERT INTO approval_requests (
  tenant_id,
  entity_type,
  entity_id,
  requested_by,
  approver_user_id,
  reason
) VALUES (
  sqlc.arg(tenant_id),
  sqlc.arg(entity_type),
  sqlc.arg(entity_id),
  sqlc.arg(requested_by),
  sqlc.arg(approver_user_id),
  sqlc.arg(reason)
)
RETURNING *;

-- name: ListApprovalRequests :many
SELECT *
FROM approval_requests
WHERE tenant_id = sqlc.arg(tenant_id)
  AND (sqlc.narg(status)::approval_status_enum IS NULL OR status = sqlc.narg(status))
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);

-- name: DecideApprovalRequest :one
UPDATE approval_requests
SET
  status = sqlc.arg(status),
  decision_note = sqlc.narg(decision_note),
  decided_at = now(),
  updated_at = now()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(approval_id)
RETURNING *;

-- name: ExportAccountsRows :many
SELECT
  id,
  owner_user_id,
  name,
  industry,
  website,
  phone,
  status,
  memo,
  created_at,
  updated_at
FROM accounts
WHERE tenant_id = sqlc.arg(tenant_id)
ORDER BY created_at DESC;

-- name: ExportOpportunitiesRows :many
SELECT
  id,
  account_id,
  contact_id,
  owner_user_id,
  name,
  stage,
  probability,
  amount::double precision AS amount,
  expected_close_date,
  next_action_at,
  next_action_note,
  created_at,
  updated_at
FROM opportunities
WHERE tenant_id = sqlc.arg(tenant_id)
ORDER BY created_at DESC;
