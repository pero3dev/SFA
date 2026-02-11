# SFA MVP Workspace

SFA (Sales Force Automation) web app workspace.

- Frontend: SvelteKit + TypeScript + Tailwind + TanStack Query + Zod
- Backend: Go + chi + pgx + sqlc
- DB: PostgreSQL 16

## Structure

- `api/openapi.yaml`: REST API contract (typed schemas)
- `db/001_init.sql`: source schema
- `backend/`: Go API scaffold + sqlc config/queries
- `frontend/`: SvelteKit scaffold

## Quick Start (Docker)

```bash
docker compose up --build
```

- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8080`
- Health: `http://localhost:8080/livez`, `http://localhost:8080/readyz`

If you already started Postgres before adding new migrations, reinitialize once:

```bash
docker compose down -v
docker compose up --build
```

## Backend Local

```bash
cd backend
go mod tidy
go test ./...
```

Generate sqlc code:

```bash
cd backend
sqlc generate
```

## Frontend Local

```bash
cd frontend
npm install
npm run check
npm run dev
```

## Added Feature Pack

Implemented features:

1. Next action management (`/opportunities/next-actions`, `/opportunities/{id}/next-action`)
2. Deal health score (`/analytics/deal-health`)
3. Forecast (`/analytics/forecast`)
4. Loss reason analytics (`/analytics/loss-reasons`)
5. Duplicate detection (`/analytics/duplicates`)
6. Email/calendar integration records (`/integrations/connections`, `/integrations/events`)
7. Approval workflow (`/approvals`, `/approvals/{id}/decision`)
8. CSV import/export (`/export/*.csv`, `/import/*.csv`)

Demo tenant data is seeded with:

- `X-Tenant-ID`: `00000000-0000-0000-0000-000000000001`
