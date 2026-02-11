-- name: GetLatestKpiSnapshot :many
SELECT
  ks.snapshot_at,
  ks.metric_key,
  ks.metric_value::double precision AS metric_value,
  ks.dimension_key,
  ks.dimensions
FROM kpi_snapshots ks
WHERE ks.tenant_id = sqlc.arg(tenant_id)
  AND snapshot_at = (
    SELECT max(snapshot_at)
    FROM kpi_snapshots ks2
    WHERE ks2.tenant_id = sqlc.arg(tenant_id)
  )
ORDER BY metric_key, dimension_key;

-- name: GetPipelineSummary :many
SELECT
  stage,
  count(*)::bigint AS deal_count,
  coalesce(sum(amount), 0)::double precision AS total_amount
FROM opportunities
WHERE tenant_id = sqlc.arg(tenant_id)
GROUP BY stage
ORDER BY stage;
