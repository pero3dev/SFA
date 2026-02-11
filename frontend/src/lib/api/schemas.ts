import { z } from "zod";

export const kpiItemSchema = z.object({
  metricKey: z.string(),
  metricValue: z.number(),
  dimensions: z.record(z.any()).optional()
});

export const kpiResponseSchema = z.object({
  snapshotAt: z.string().datetime().or(z.string()),
  data: z.array(kpiItemSchema)
});

export type KpiResponse = z.infer<typeof kpiResponseSchema>;

export const pipelineItemSchema = z.object({
  stage: z.enum(["new_lead", "qualified", "proposal", "negotiation", "closed_won", "closed_lost"]),
  count: z.number(),
  totalAmount: z.number()
});

export const pipelineResponseSchema = z.object({
  data: z.array(pipelineItemSchema)
});

export type PipelineResponse = z.infer<typeof pipelineResponseSchema>;

export const nextActionSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  stage: z.string(),
  accountName: z.string(),
  nextActionAt: z.string(),
  nextActionNote: z.string()
});
export const nextActionsResponseSchema = z.object({
  data: z.array(nextActionSchema)
});
export type NextActionsResponse = z.infer<typeof nextActionsResponseSchema>;

export const dealHealthSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  stage: z.string(),
  probability: z.number(),
  amount: z.number(),
  lastActivityAt: z.string(),
  healthScore: z.number()
});
export const dealHealthResponseSchema = z.object({
  data: z.array(dealHealthSchema)
});
export type DealHealthResponse = z.infer<typeof dealHealthResponseSchema>;

export const forecastSchema = z.object({
  ownerUserId: z.string().uuid().or(z.literal("")),
  month: z.string(),
  dealCount: z.number(),
  pipelineAmount: z.number(),
  weightedAmount: z.number()
});
export const forecastResponseSchema = z.object({
  data: z.array(forecastSchema)
});
export type ForecastResponse = z.infer<typeof forecastResponseSchema>;

export const lossReasonSchema = z.object({
  reason: z.string(),
  lostCount: z.number(),
  lostAmount: z.number()
});
export const lossReasonsResponseSchema = z.object({
  data: z.array(lossReasonSchema)
});
export type LossReasonsResponse = z.infer<typeof lossReasonsResponseSchema>;

export const duplicateSchema = z.object({
  type: z.string(),
  primaryId: z.string().uuid(),
  duplicateId: z.string().uuid(),
  matchValue: z.string()
});
export const duplicatesResponseSchema = z.object({
  data: z.array(duplicateSchema)
});
export type DuplicatesResponse = z.infer<typeof duplicatesResponseSchema>;

export const integrationConnectionSchema = z.object({
  id: z.string().uuid(),
  userId: z.string().uuid(),
  provider: z.string(),
  integrationType: z.string(),
  externalAccountId: z.string(),
  status: z.string(),
  scopes: z.array(z.string()),
  expiresAt: z.string().optional(),
  updatedAt: z.string()
});
export const integrationConnectionsResponseSchema = z.object({
  data: z.array(integrationConnectionSchema)
});
export type IntegrationConnectionsResponse = z.infer<typeof integrationConnectionsResponseSchema>;

export const integrationEventSchema = z.object({
  id: z.number(),
  provider: z.string(),
  integrationType: z.string(),
  externalEventId: z.string().optional(),
  eventType: z.string(),
  payload: z.record(z.any()),
  linkedAccountId: z.string().optional(),
  linkedContactId: z.string().optional(),
  linkedOpportunityId: z.string().optional(),
  occurredAt: z.string()
});
export const integrationEventsResponseSchema = z.object({
  data: z.array(integrationEventSchema)
});
export type IntegrationEventsResponse = z.infer<typeof integrationEventsResponseSchema>;

export const approvalSchema = z.object({
  id: z.string().uuid(),
  entityType: z.string(),
  entityId: z.string().uuid(),
  requestedBy: z.string().uuid(),
  approverUserId: z.string().uuid(),
  status: z.string(),
  reason: z.string(),
  decisionNote: z.string().optional(),
  decidedAt: z.string().optional(),
  createdAt: z.string()
});
export const approvalsResponseSchema = z.object({
  data: z.array(approvalSchema)
});
export type ApprovalsResponse = z.infer<typeof approvalsResponseSchema>;
