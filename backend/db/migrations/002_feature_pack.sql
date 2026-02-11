BEGIN;

CREATE TYPE integration_provider_enum AS ENUM ('google', 'microsoft');
CREATE TYPE integration_type_enum AS ENUM ('email', 'calendar');
CREATE TYPE integration_status_enum AS ENUM ('active', 'revoked', 'error');
CREATE TYPE approval_status_enum AS ENUM ('pending', 'approved', 'rejected');

ALTER TABLE opportunities
  ADD COLUMN next_action_at TIMESTAMPTZ,
  ADD COLUMN next_action_note TEXT;

CREATE TABLE integration_connections (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider integration_provider_enum NOT NULL,
  integration_type integration_type_enum NOT NULL,
  external_account_id TEXT NOT NULL,
  status integration_status_enum NOT NULL DEFAULT 'active',
  access_token TEXT,
  refresh_token TEXT,
  expires_at TIMESTAMPTZ,
  scopes TEXT[] NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, user_id, provider, integration_type, external_account_id)
);

CREATE TABLE integration_events (
  id BIGSERIAL PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  provider integration_provider_enum NOT NULL,
  integration_type integration_type_enum NOT NULL,
  external_event_id TEXT,
  event_type TEXT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  linked_account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
  linked_contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
  linked_opportunity_id UUID REFERENCES opportunities(id) ON DELETE SET NULL,
  occurred_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (tenant_id, provider, integration_type, external_event_id)
);

CREATE TABLE approval_requests (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  entity_type TEXT NOT NULL,
  entity_id UUID NOT NULL,
  requested_by UUID NOT NULL REFERENCES users(id),
  approver_user_id UUID NOT NULL REFERENCES users(id),
  status approval_status_enum NOT NULL DEFAULT 'pending',
  reason TEXT NOT NULL,
  decision_note TEXT,
  decided_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_opportunities_tenant_next_action ON opportunities (tenant_id, next_action_at);
CREATE INDEX idx_integration_connections_tenant_provider ON integration_connections (tenant_id, provider, integration_type);
CREATE INDEX idx_integration_events_tenant_occurred ON integration_events (tenant_id, occurred_at DESC);
CREATE INDEX idx_approval_requests_tenant_status ON approval_requests (tenant_id, status, created_at DESC);

ALTER TABLE integration_connections ENABLE ROW LEVEL SECURITY;
ALTER TABLE integration_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE approval_requests ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_integration_connections ON integration_connections
  USING (tenant_id = current_setting('app.tenant_id', true)::UUID)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::UUID);
CREATE POLICY tenant_isolation_integration_events ON integration_events
  USING (tenant_id = current_setting('app.tenant_id', true)::UUID)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::UUID);
CREATE POLICY tenant_isolation_approval_requests ON approval_requests
  USING (tenant_id = current_setting('app.tenant_id', true)::UUID)
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true)::UUID);

-- Demo data for tenant 000...001 so dashboard/features are visible immediately.
INSERT INTO tenants (id, name)
VALUES ('00000000-0000-0000-0000-000000000001', 'Demo Tenant')
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (id, email, password_hash, display_name, is_active)
VALUES
  ('00000000-0000-0000-0000-000000000010', 'admin@example.com', 'dummy-hash', 'Demo Admin', true),
  ('00000000-0000-0000-0000-000000000011', 'sales@example.com', 'dummy-hash', 'Demo Sales', true),
  ('00000000-0000-0000-0000-000000000012', 'manager@example.com', 'dummy-hash', 'Demo Manager', true)
ON CONFLICT (email) DO NOTHING;

INSERT INTO memberships (tenant_id, user_id, role)
VALUES
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000010', 'admin'),
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000011', 'sales'),
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000012', 'manager')
ON CONFLICT (tenant_id, user_id) DO NOTHING;

INSERT INTO accounts (
  id, tenant_id, owner_user_id, name, industry, website, status, created_by
)
VALUES
  ('00000000-0000-0000-0000-000000000100', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000011', 'Acme Corp', 'Manufacturing', 'https://acme.example.com', 'active', '00000000-0000-0000-0000-000000000010'),
  ('00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000011', 'acme corp', 'Manufacturing', 'https://acme.example.com', 'prospect', '00000000-0000-0000-0000-000000000010'),
  ('00000000-0000-0000-0000-000000000102', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000012', 'Beta Ltd', 'Software', 'https://beta.example.com', 'active', '00000000-0000-0000-0000-000000000010')
ON CONFLICT (id) DO NOTHING;

INSERT INTO contacts (
  id, tenant_id, account_id, owner_user_id, full_name, email, created_by
)
VALUES
  ('00000000-0000-0000-0000-000000000120', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000100', '00000000-0000-0000-0000-000000000011', 'Taro A', 'buyer@acme.example.com', '00000000-0000-0000-0000-000000000010'),
  ('00000000-0000-0000-0000-000000000121', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000011', 'Taro B', 'buyer@acme.example.com', '00000000-0000-0000-0000-000000000010')
ON CONFLICT (id) DO NOTHING;

INSERT INTO opportunities (
  id, tenant_id, account_id, contact_id, owner_user_id, name, stage, probability, amount, expected_close_date, next_action_at, next_action_note, created_by
)
VALUES
  ('00000000-0000-0000-0000-000000000200', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000100', '00000000-0000-0000-0000-000000000120', '00000000-0000-0000-0000-000000000011', 'Acme Renewal FY26', 'proposal', 70, 2200000, current_date + 20, now() + interval '2 day', 'Send final quote', '00000000-0000-0000-0000-000000000010'),
  ('00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000102', null, '00000000-0000-0000-0000-000000000012', 'Beta Expansion', 'negotiation', 55, 4800000, current_date + 35, now() + interval '5 day', 'Executive meeting', '00000000-0000-0000-0000-000000000010'),
  ('00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000100', null, '00000000-0000-0000-0000-000000000011', 'Acme Add-on Module', 'closed_lost', 0, 900000, current_date - 10, null, null, '00000000-0000-0000-0000-000000000010')
ON CONFLICT (id) DO NOTHING;

INSERT INTO activities (
  tenant_id, opportunity_id, activity_type, subject, activity_at, created_by
)
VALUES
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000200', 'meeting', 'Proposal walkthrough', now() - interval '1 day', '00000000-0000-0000-0000-000000000011'),
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000201', 'call', 'Pricing call', now() - interval '3 day', '00000000-0000-0000-0000-000000000012')
ON CONFLICT DO NOTHING;

INSERT INTO opportunity_losses (
  id, tenant_id, opportunity_id, reason, detail, lost_at, created_by
)
VALUES
  ('00000000-0000-0000-0000-000000000230', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000202', 'competitor', 'Price undercut by competitor', now() - interval '10 day', '00000000-0000-0000-0000-000000000011')
ON CONFLICT (opportunity_id) DO NOTHING;

WITH snap AS (
  SELECT now()::timestamptz AS ts
)
INSERT INTO kpi_snapshots (
  tenant_id, snapshot_at, metric_key, metric_value, dimension_key, dimensions
)
SELECT
  '00000000-0000-0000-0000-000000000001',
  snap.ts,
  v.metric_key,
  v.metric_value,
  '',
  '{}'::jsonb
FROM snap
JOIN (
  VALUES
    ('open_pipeline_amount', 7000000.0),
    ('weighted_pipeline_amount', 4180000.0),
    ('open_deal_count', 2.0),
    ('win_rate_90d', 0.43)
) AS v(metric_key, metric_value) ON true
ON CONFLICT (tenant_id, snapshot_at, metric_key, dimension_key) DO NOTHING;

INSERT INTO integration_connections (
  tenant_id, user_id, provider, integration_type, external_account_id, status, scopes
)
VALUES
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000011', 'google', 'calendar', 'sales@example.com', 'active', ARRAY['calendar.readonly']),
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000012', 'microsoft', 'email', 'manager@example.com', 'active', ARRAY['mail.read'])
ON CONFLICT (tenant_id, user_id, provider, integration_type, external_account_id) DO NOTHING;

INSERT INTO approval_requests (
  tenant_id, entity_type, entity_id, requested_by, approver_user_id, status, reason
)
VALUES
  ('00000000-0000-0000-0000-000000000001', 'quote', '00000000-0000-0000-0000-000000000200', '00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000012', 'pending', 'High discount requires approval')
ON CONFLICT DO NOTHING;

COMMIT;
