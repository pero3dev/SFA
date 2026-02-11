import {
  approvalsResponseSchema,
  dealHealthResponseSchema,
  duplicatesResponseSchema,
  forecastResponseSchema,
  integrationConnectionsResponseSchema,
  integrationEventsResponseSchema,
  kpiResponseSchema,
  lossReasonsResponseSchema,
  nextActionsResponseSchema,
  pipelineResponseSchema,
  type ApprovalsResponse,
  type DealHealthResponse,
  type DuplicatesResponse,
  type ForecastResponse,
  type IntegrationConnectionsResponse,
  type IntegrationEventsResponse,
  type KpiResponse,
  type LossReasonsResponse,
  type NextActionsResponse,
  type PipelineResponse
} from "./schemas";

const apiBaseURL = import.meta.env.PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";
const tenantID = import.meta.env.PUBLIC_TENANT_ID ?? "00000000-0000-0000-0000-000000000001";

async function request<T>(path: string, parser: (input: unknown) => T): Promise<T> {
  const response = await fetch(`${apiBaseURL}${path}`, {
    headers: {
      "Content-Type": "application/json",
      "X-Tenant-ID": tenantID
    }
  });

  if (!response.ok) {
    throw new Error(`API request failed: ${response.status}`);
  }

  return parser(await response.json());
}

async function requestWithBody<TResponse, TRequest>(
  path: string,
  method: "POST" | "PATCH",
  body: TRequest,
  parser: (input: unknown) => TResponse
): Promise<TResponse> {
  const response = await fetch(`${apiBaseURL}${path}`, {
    method,
    headers: {
      "Content-Type": "application/json",
      "X-Tenant-ID": tenantID
    },
    body: JSON.stringify(body)
  });

  if (!response.ok) {
    throw new Error(`API request failed: ${response.status}`);
  }

  return parser(await response.json());
}

export async function fetchKpi(): Promise<KpiResponse> {
  return request("/dashboard/kpi", (input) => kpiResponseSchema.parse(input));
}

export async function fetchPipeline(): Promise<PipelineResponse> {
  return request("/dashboard/pipeline", (input) => pipelineResponseSchema.parse(input));
}

export async function fetchNextActions(): Promise<NextActionsResponse> {
  return request("/opportunities/next-actions", (input) => nextActionsResponseSchema.parse(input));
}

export async function fetchDealHealth(): Promise<DealHealthResponse> {
  return request("/analytics/deal-health", (input) => dealHealthResponseSchema.parse(input));
}

export async function fetchForecast(): Promise<ForecastResponse> {
  return request("/analytics/forecast", (input) => forecastResponseSchema.parse(input));
}

export async function fetchLossReasons(): Promise<LossReasonsResponse> {
  return request("/analytics/loss-reasons", (input) => lossReasonsResponseSchema.parse(input));
}

export async function fetchDuplicates(): Promise<DuplicatesResponse> {
  return request("/analytics/duplicates", (input) => duplicatesResponseSchema.parse(input));
}

export async function fetchIntegrations(): Promise<IntegrationConnectionsResponse> {
  return request("/integrations/connections", (input) => integrationConnectionsResponseSchema.parse(input));
}

export async function upsertIntegrationConnection(payload: {
  userId: string;
  provider: string;
  integrationType: string;
  externalAccountId: string;
  status?: string;
  scopes?: string[];
}): Promise<IntegrationConnectionsResponse> {
  await requestWithBody("/integrations/connections", "POST", payload, (input) => input);
  return fetchIntegrations();
}

export async function fetchIntegrationEvents(): Promise<IntegrationEventsResponse> {
  return request("/integrations/events", (input) => integrationEventsResponseSchema.parse(input));
}

export async function createIntegrationEvent(payload: {
  provider: string;
  integrationType: string;
  eventType: string;
  payload?: Record<string, unknown>;
  occurredAt: string;
}): Promise<IntegrationEventsResponse> {
  await requestWithBody("/integrations/events", "POST", payload, (input) => input);
  return fetchIntegrationEvents();
}

export async function fetchApprovals(): Promise<ApprovalsResponse> {
  return request("/approvals", (input) => approvalsResponseSchema.parse(input));
}

export async function createApproval(payload: {
  entityType: string;
  entityId: string;
  requestedBy: string;
  approverUserId: string;
  reason: string;
}): Promise<ApprovalsResponse> {
  await requestWithBody("/approvals", "POST", payload, (input) => input);
  return fetchApprovals();
}

export async function decideApproval(
  id: string,
  payload: { status: "approved" | "rejected"; decisionNote?: string }
): Promise<ApprovalsResponse> {
  await requestWithBody(`/approvals/${id}/decision`, "POST", payload, (input) => input);
  return fetchApprovals();
}

export async function updateNextAction(payload: {
  id: string;
  nextActionAt: string;
  nextActionNote: string;
}): Promise<NextActionsResponse> {
  await requestWithBody(`/opportunities/${payload.id}/next-action`, "PATCH", payload, (input) => input);
  return fetchNextActions();
}

export async function uploadCSV(path: "/import/accounts.csv" | "/import/opportunities.csv", file: File): Promise<void> {
  const formData = new FormData();
  formData.append("file", file);

  const response = await fetch(`${apiBaseURL}${path}`, {
    method: "POST",
    headers: {
      "X-Tenant-ID": tenantID
    },
    body: formData
  });
  if (!response.ok) {
    throw new Error(`CSV upload failed: ${response.status}`);
  }
}

export async function downloadCSV(path: "/export/accounts.csv" | "/export/opportunities.csv", filename: string): Promise<void> {
  const response = await fetch(`${apiBaseURL}${path}`, {
    headers: {
      "X-Tenant-ID": tenantID
    }
  });
  if (!response.ok) {
    throw new Error(`CSV download failed: ${response.status}`);
  }

  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}
