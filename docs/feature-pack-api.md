# Feature Pack API

All endpoints require:

- Header: `X-Tenant-ID: <tenant_uuid>`
- Base: `/api/v1`

## 1) Next Action Management

- `GET /opportunities/next-actions`
  - Query: `dueBefore` (optional, RFC3339), `page`, `limit`
- `PATCH /opportunities/{id}/next-action`
  - Body:
    - `nextActionAt` (RFC3339)
    - `nextActionNote`

## 2) Deal Health Score

- `GET /analytics/deal-health`
  - Query: `page`, `limit`

## 3) Forecast

- `GET /analytics/forecast`

## 4) Loss Reason Analytics

- `GET /analytics/loss-reasons`

## 5) Duplicate Detection

- `GET /analytics/duplicates`

## 6) Email/Calendar Integration

- `GET /integrations/connections`
- `POST /integrations/connections`
  - Body:
    - `userId`
    - `provider` (`google` / `microsoft`)
    - `integrationType` (`email` / `calendar`)
    - `externalAccountId`
    - `status` (`active` / `revoked` / `error`, optional)
    - `scopes` (`string[]`, optional)
- `GET /integrations/events`
  - Query: `page`, `limit`
- `POST /integrations/events`
  - Body:
    - `provider`
    - `integrationType`
    - `eventType`
    - `occurredAt` (RFC3339)
    - `externalEventId`, `payload`, `linkedAccountId`, `linkedContactId`, `linkedOpportunityId` (optional)

## 7) Approval Workflow

- `GET /approvals`
  - Query: `status` (`pending` / `approved` / `rejected`, optional), `page`, `limit`
- `POST /approvals`
  - Body:
    - `entityType`
    - `entityId`
    - `requestedBy`
    - `approverUserId`
    - `reason`
- `POST /approvals/{id}/decision`
  - Body:
    - `status` (`approved` / `rejected`)
    - `decisionNote` (optional)

## 8) CSV Import / Export

- `GET /export/accounts.csv`
- `GET /export/opportunities.csv`
- `POST /import/accounts.csv`
  - multipart/form-data (`file`)
- `POST /import/opportunities.csv`
  - multipart/form-data (`file`)
