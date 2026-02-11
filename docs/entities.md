# SFA Requirements v0.1 - Entity Definition

## 1. Scope
This document defines the MVP data model for a multi-tenant SFA web application.

- Tenant isolation: row-level by `tenant_id`
- Roles: `admin`, `manager`, `sales` (RBAC only)
- Main audit target operations: `create`, `update`, `delete`, `login`

## 2. Core Entities

### tenants
- Purpose: tenant master (company/workspace)
- Primary key: `id` (UUID)
- Notes: all business data belongs to one tenant

### users
- Purpose: login identity (global)
- Primary key: `id` (UUID)
- Unique: `email`
- Notes: tenant affiliation is expressed by `memberships`

### memberships
- Purpose: user-to-tenant mapping and role assignment
- Primary key: `id` (UUID)
- Foreign keys: `tenant_id -> tenants.id`, `user_id -> users.id`
- Unique: `(tenant_id, user_id)`
- Role values: `admin`, `manager`, `sales`

### accounts
- Purpose: customer company
- Primary key: `id` (UUID)
- Foreign keys: `tenant_id`, `owner_user_id`, `created_by`
- Notes: parent for contacts, locations, opportunities

### account_locations
- Purpose: department/branch/site under account
- Primary key: `id` (UUID)
- Foreign keys: `tenant_id`, `account_id`

### contacts
- Purpose: person in charge at customer
- Primary key: `id` (UUID)
- Foreign keys: `tenant_id`, `account_id`, `location_id (optional)`, `owner_user_id`, `created_by`

### opportunities
- Purpose: sales deal/opportunity
- Primary key: `id` (UUID)
- Foreign keys: `tenant_id`, `account_id`, `contact_id (optional)`, `owner_user_id`, `created_by`
- Main fields: `stage`, `probability`, `amount`, `expected_close_date`

### activities
- Purpose: timeline activities (meeting/call/email/note/task)
- Primary key: `id` (UUID)
- Foreign keys: `tenant_id`, `opportunity_id`, `created_by`

### quotes
- Purpose: quote records tied to opportunities
- Primary key: `id` (UUID)
- Foreign keys: `tenant_id`, `opportunity_id`, `created_by`
- Unique: `(tenant_id, quote_no)`

### orders
- Purpose: order records tied to opportunities
- Primary key: `id` (UUID)
- Foreign keys: `tenant_id`, `opportunity_id`, `created_by`
- Unique: `(tenant_id, order_no)`

### opportunity_losses
- Purpose: lost reason detail (1 record per lost opportunity)
- Primary key: `id` (UUID)
- Foreign keys: `tenant_id`, `opportunity_id`, `created_by`
- Unique: `opportunity_id`

### audit_logs
- Purpose: operation audit trail for critical actions
- Primary key: `id` (BIGSERIAL)
- Foreign keys: `tenant_id`, `actor_user_id`
- Notes: `metadata` JSONB stores structured payload snapshot

### refresh_tokens
- Purpose: refresh token session management
- Primary key: `id` (UUID)
- Foreign keys: `user_id`
- Unique: `token_hash`

### kpi_snapshots
- Purpose: near real-time dashboard aggregation output
- Primary key: `id` (BIGSERIAL)
- Foreign keys: `tenant_id`
- Unique: `(tenant_id, snapshot_at, metric_key, dimension_key)`

### integration_connections
- Purpose: store integration account linkage (Google/Microsoft email/calendar)
- Primary key: `id` (UUID)
- Foreign keys: `tenant_id`, `user_id`
- Unique: `(tenant_id, user_id, provider, integration_type, external_account_id)`

### integration_events
- Purpose: imported email/calendar activity events
- Primary key: `id` (BIGSERIAL)
- Foreign keys: `tenant_id`, optional links to `accounts/contacts/opportunities`

### approval_requests
- Purpose: approval workflow records for high-risk operations (e.g., high discount quote)
- Primary key: `id` (UUID)
- Foreign keys: `tenant_id`, `requested_by`, `approver_user_id`

## 3. Enum Definitions

- `role_enum`: `admin`, `manager`, `sales`
- `account_status_enum`: `prospect`, `active`, `inactive`
- `opportunity_stage_enum`: `new_lead`, `qualified`, `proposal`, `negotiation`, `closed_won`, `closed_lost`
- `activity_type_enum`: `meeting`, `call`, `email`, `note`, `task`
- `quote_status_enum`: `draft`, `sent`, `accepted`, `rejected`, `expired`
- `order_status_enum`: `pending`, `confirmed`, `cancelled`, `invoiced`
- `loss_reason_enum`: `budget`, `competitor`, `timing`, `no_decision`, `other`
- `audit_action_enum`: `create`, `update`, `delete`, `login`
- `integration_provider_enum`: `google`, `microsoft`
- `integration_type_enum`: `email`, `calendar`
- `integration_status_enum`: `active`, `revoked`, `error`
- `approval_status_enum`: `pending`, `approved`, `rejected`

## 4. Relationship Summary

- `tenants 1 - n memberships`
- `users 1 - n memberships`
- `accounts 1 - n account_locations`
- `accounts 1 - n contacts`
- `accounts 1 - n opportunities`
- `opportunities 1 - n activities`
- `opportunities 1 - n quotes`
- `opportunities 1 - n orders`
- `opportunities 1 - 0..1 opportunity_losses`

## 5. RBAC MVP Intent

- `sales`: create/update own opportunities and activities, view customer data
- `manager`: team-level visibility and update rights for opportunities
- `admin`: full tenant-level access, user and role administration, audit log viewing

## 6. Tenant Isolation

- All business tables include `tenant_id`.
- PostgreSQL RLS policies are enabled on tenant-scoped tables.
- Application layer must set `SET app.tenant_id = '<tenant_uuid>'` per request transaction.
