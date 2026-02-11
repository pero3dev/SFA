<script lang="ts">
  import { createQuery, useQueryClient } from "@tanstack/svelte-query";
  import {
    createApproval,
    createIntegrationEvent,
    decideApproval,
    downloadCSV,
    fetchApprovals,
    fetchDealHealth,
    fetchDuplicates,
    fetchForecast,
    fetchIntegrationEvents,
    fetchIntegrations,
    fetchKpi,
    fetchLossReasons,
    fetchNextActions,
    fetchPipeline,
    updateNextAction,
    uploadCSV,
    upsertIntegrationConnection
  } from "$lib/api/client";
  import type {
    ApprovalsResponse,
    DealHealthResponse,
    DuplicatesResponse,
    ForecastResponse,
    IntegrationConnectionsResponse,
    IntegrationEventsResponse,
    KpiResponse,
    LossReasonsResponse,
    NextActionsResponse,
    PipelineResponse
  } from "$lib/api/schemas";

  const queryClient = useQueryClient();

  const kpiQuery = createQuery<KpiResponse>({ queryKey: ["dashboard", "kpi"], queryFn: fetchKpi, refetchInterval: 30000 });
  const pipelineQuery = createQuery<PipelineResponse>({
    queryKey: ["dashboard", "pipeline"],
    queryFn: fetchPipeline,
    refetchInterval: 30000
  });
  const nextActionsQuery = createQuery<NextActionsResponse>({
    queryKey: ["opportunities", "next-actions"],
    queryFn: fetchNextActions,
    refetchInterval: 30000
  });
  const healthQuery = createQuery<DealHealthResponse>({
    queryKey: ["analytics", "deal-health"],
    queryFn: fetchDealHealth,
    refetchInterval: 45000
  });
  const forecastQuery = createQuery<ForecastResponse>({
    queryKey: ["analytics", "forecast"],
    queryFn: fetchForecast,
    refetchInterval: 45000
  });
  const lossQuery = createQuery<LossReasonsResponse>({
    queryKey: ["analytics", "loss-reasons"],
    queryFn: fetchLossReasons,
    refetchInterval: 45000
  });
  const duplicatesQuery = createQuery<DuplicatesResponse>({
    queryKey: ["analytics", "duplicates"],
    queryFn: fetchDuplicates,
    refetchInterval: 45000
  });
  const integrationsQuery = createQuery<IntegrationConnectionsResponse>({
    queryKey: ["integrations", "connections"],
    queryFn: fetchIntegrations,
    refetchInterval: 45000
  });
  const integrationEventsQuery = createQuery<IntegrationEventsResponse>({
    queryKey: ["integrations", "events"],
    queryFn: fetchIntegrationEvents,
    refetchInterval: 45000
  });
  const approvalsQuery = createQuery<ApprovalsResponse>({
    queryKey: ["approvals"],
    queryFn: fetchApprovals,
    refetchInterval: 30000
  });

  let notice = "";
  let nextActionForm = {
    id: "00000000-0000-0000-0000-000000000200",
    nextActionAt: new Date(Date.now() + 86400000).toISOString().slice(0, 16),
    nextActionNote: "Follow-up call"
  };
  let integrationForm = {
    userId: "00000000-0000-0000-0000-000000000011",
    provider: "google",
    integrationType: "calendar",
    externalAccountId: "sales@example.com",
    status: "active",
    scopes: "calendar.readonly"
  };
  let eventForm = {
    provider: "google",
    integrationType: "calendar",
    eventType: "calendar.meeting.created",
    occurredAt: new Date().toISOString()
  };
  let approvalForm = {
    entityType: "quote",
    entityId: "00000000-0000-0000-0000-000000000200",
    requestedBy: "00000000-0000-0000-0000-000000000011",
    approverUserId: "00000000-0000-0000-0000-000000000012",
    reason: "Discount over 20%"
  };
  let accountsCsv: File | null = null;
  let opportunitiesCsv: File | null = null;

  async function refresh(keys: readonly unknown[][]): Promise<void> {
    await Promise.all(keys.map((key) => queryClient.invalidateQueries({ queryKey: key })));
  }

  function statusTone(score: number): string {
    if (score < 70) return "text-rose-300";
    if (score < 90) return "text-amber-200";
    return "text-cyan-200";
  }

  async function onUpdateNextAction(): Promise<void> {
    try {
      await updateNextAction({
        id: nextActionForm.id,
        nextActionAt: new Date(nextActionForm.nextActionAt).toISOString(),
        nextActionNote: nextActionForm.nextActionNote
      });
      await refresh([["opportunities", "next-actions"], ["analytics", "deal-health"]]);
      notice = "Next action updated.";
    } catch (error) {
      notice = `Failed to update next action: ${String(error)}`;
    }
  }

  async function onSaveIntegration(): Promise<void> {
    try {
      await upsertIntegrationConnection({
        userId: integrationForm.userId,
        provider: integrationForm.provider,
        integrationType: integrationForm.integrationType,
        externalAccountId: integrationForm.externalAccountId,
        status: integrationForm.status,
        scopes: integrationForm.scopes.split(",").map((s) => s.trim()).filter((s) => s.length > 0)
      });
      await refresh([["integrations", "connections"]]);
      notice = "Integration settings saved.";
    } catch (error) {
      notice = `Failed to save integration settings: ${String(error)}`;
    }
  }

  async function onCreateIntegrationEvent(): Promise<void> {
    try {
      await createIntegrationEvent({
        provider: eventForm.provider,
        integrationType: eventForm.integrationType,
        eventType: eventForm.eventType,
        occurredAt: eventForm.occurredAt,
        payload: { source: "manual", createdBy: "dashboard-ui" }
      });
      await refresh([["integrations", "events"]]);
      notice = "Integration event created.";
    } catch (error) {
      notice = `Failed to create integration event: ${String(error)}`;
    }
  }

  async function onCreateApproval(): Promise<void> {
    try {
      await createApproval(approvalForm);
      await refresh([["approvals"]]);
      notice = "Approval request created.";
    } catch (error) {
      notice = `Failed to create approval request: ${String(error)}`;
    }
  }

  async function onDecideApproval(id: string, status: "approved" | "rejected"): Promise<void> {
    try {
      await decideApproval(id, { status, decisionNote: `UI decision: ${status}` });
      await refresh([["approvals"]]);
      notice = `Approval updated to ${status}.`;
    } catch (error) {
      notice = `Failed to update approval: ${String(error)}`;
    }
  }

  async function onUploadAccounts(): Promise<void> {
    if (!accountsCsv) return;
    try {
      await uploadCSV("/import/accounts.csv", accountsCsv);
      notice = "accounts.csv imported.";
      await refresh([["analytics", "duplicates"]]);
    } catch (error) {
      notice = `Failed to import accounts.csv: ${String(error)}`;
    }
  }

  async function onUploadOpportunities(): Promise<void> {
    if (!opportunitiesCsv) return;
    try {
      await uploadCSV("/import/opportunities.csv", opportunitiesCsv);
      notice = "opportunities.csv imported.";
      await refresh([["opportunities", "next-actions"], ["analytics", "forecast"], ["analytics", "deal-health"]]);
    } catch (error) {
      notice = `Failed to import opportunities.csv: ${String(error)}`;
    }
  }
</script>

<main class="space-y-6 pb-8">
  {#if notice}
    <aside class="glass-soft fade-in rounded-2xl border border-cyan-200/30 px-4 py-3 text-sm text-cyan-100">
      {notice}
    </aside>
  {/if}

  <section class="stagger grid grid-cols-1 gap-5 xl:grid-cols-12">
    <article class="glass-panel rounded-3xl p-5 xl:col-span-8">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="section-title">KPI Pulse</h2>
        <span class="status-pill">Realtime</span>
      </div>
      {#if $kpiQuery.isSuccess}
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {#each $kpiQuery.data.data as item}
            <div class="glass-soft rounded-2xl p-4">
              <p class="meta-label">{item.metricKey}</p>
              <p class="kpi-value mt-2 text-cyan-100">{item.metricValue.toLocaleString()}</p>
            </div>
          {/each}
        </div>
      {:else}
        <p class="text-sm text-slate-300">Loading KPI stream...</p>
      {/if}
    </article>

    <article class="glass-panel rounded-3xl p-5 xl:col-span-4">
      <h2 class="section-title mb-4">Pipeline Stage</h2>
      {#if $pipelineQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $pipelineQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5">
              <p class="text-sm font-semibold text-slate-100">{row.stage}</p>
              <p class="mt-1 text-xs text-slate-300">Deals {row.count} - Amount {row.totalAmount.toLocaleString()}</p>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="text-sm text-slate-300">Loading pipeline...</p>
      {/if}
    </article>
  </section>

  <section class="stagger grid grid-cols-1 gap-5 xl:grid-cols-12">
    <article class="glass-panel rounded-3xl p-5 xl:col-span-6">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="section-title">Next Action Control</h2>
        <span class="status-pill">Execution</span>
      </div>
      <form class="mb-4 grid gap-2 md:grid-cols-3" on:submit|preventDefault={onUpdateNextAction}>
        <input class="field" bind:value={nextActionForm.id} placeholder="Opportunity ID" />
        <input class="field" bind:value={nextActionForm.nextActionAt} type="datetime-local" />
        <input class="field" bind:value={nextActionForm.nextActionNote} placeholder="Next action note" />
        <button class="btn-primary md:col-span-3" type="submit">Update Next Action</button>
      </form>
      {#if $nextActionsQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $nextActionsQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <p class="text-sm font-semibold text-slate-100">{row.name}</p>
                <span class="status-pill">{row.stage}</span>
              </div>
              <p class="mt-1 text-xs text-slate-300">{row.accountName}</p>
              <p class="mt-1 text-xs text-cyan-100">{row.nextActionAt}</p>
              <p class="mt-1 text-sm text-slate-200">{row.nextActionNote}</p>
            </li>
          {/each}
        </ul>
      {/if}
    </article>

    <article class="glass-panel rounded-3xl p-5 xl:col-span-6">
      <h2 class="section-title mb-4">Deal Health Radar</h2>
      {#if $healthQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $healthQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5">
              <div class="flex items-center justify-between gap-2">
                <p class="text-sm font-semibold text-slate-100">{row.name}</p>
                <p class={`text-sm font-bold ${statusTone(row.healthScore)}`}>{row.healthScore}</p>
              </div>
              <p class="mt-1 text-xs text-slate-300">{row.stage} - Last activity {row.lastActivityAt}</p>
              <p class="mt-1 text-xs text-slate-400">Probability {row.probability}% - Amount {row.amount.toLocaleString()}</p>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="text-sm text-slate-300">Loading health radar...</p>
      {/if}
    </article>
  </section>

  <section class="stagger grid grid-cols-1 gap-5 xl:grid-cols-12">
    <article class="glass-panel rounded-3xl p-5 xl:col-span-4">
      <h2 class="section-title mb-4">Forecast Matrix</h2>
      {#if $forecastQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $forecastQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5 text-sm">
              <p class="font-semibold text-slate-100">{row.month}</p>
              <p class="mt-1 text-xs text-slate-300">Deals {row.dealCount}</p>
              <p class="mt-1 text-cyan-100">Pipeline {row.pipelineAmount.toLocaleString()}</p>
              <p class="text-violet-200">Weighted {row.weightedAmount.toLocaleString()}</p>
            </li>
          {/each}
        </ul>
      {/if}
    </article>

    <article class="glass-panel rounded-3xl p-5 xl:col-span-4">
      <h2 class="section-title mb-4">Loss Reason Analytics</h2>
      {#if $lossQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $lossQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5 text-sm">
              <p class="font-semibold text-slate-100">{row.reason}</p>
              <p class="mt-1 text-xs text-slate-300">Count {row.lostCount}</p>
              <p class="mt-1 text-rose-200">Amount {row.lostAmount.toLocaleString()}</p>
            </li>
          {/each}
        </ul>
      {/if}
    </article>

    <article class="glass-panel rounded-3xl p-5 xl:col-span-4">
      <h2 class="section-title mb-4">Duplicate Intelligence</h2>
      {#if $duplicatesQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $duplicatesQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5 text-sm">
              <p class="font-semibold text-slate-100">{row.type}</p>
              <p class="mt-1 break-all text-xs text-slate-300">{row.matchValue}</p>
            </li>
          {/each}
        </ul>
      {/if}
    </article>
  </section>

  <section class="stagger grid grid-cols-1 gap-5 xl:grid-cols-12">
    <article class="glass-panel rounded-3xl p-5 xl:col-span-6">
      <h2 class="section-title mb-4">Integration Hub</h2>
      <form class="mb-4 grid gap-2 md:grid-cols-2" on:submit|preventDefault={onSaveIntegration}>
        <input class="field" bind:value={integrationForm.userId} placeholder="User ID" />
        <input class="field" bind:value={integrationForm.externalAccountId} placeholder="External account" />
        <select class="field" bind:value={integrationForm.provider}>
          <option value="google">google</option>
          <option value="microsoft">microsoft</option>
        </select>
        <select class="field" bind:value={integrationForm.integrationType}>
          <option value="calendar">calendar</option>
          <option value="email">email</option>
        </select>
        <input class="field md:col-span-2" bind:value={integrationForm.scopes} placeholder="scope1,scope2" />
        <button class="btn-primary md:col-span-2" type="submit">Save Integration</button>
      </form>

      {#if $integrationsQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $integrationsQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5 text-sm">
              <div class="flex items-center justify-between gap-2">
                <p class="font-semibold text-slate-100">{row.provider} / {row.integrationType}</p>
                <span class="status-pill">{row.status}</span>
              </div>
              <p class="mt-1 text-xs text-slate-300">{row.externalAccountId}</p>
            </li>
          {/each}
        </ul>
      {/if}
    </article>

    <article class="glass-panel rounded-3xl p-5 xl:col-span-6">
      <h2 class="section-title mb-4">Integration Event Stream</h2>
      <form class="mb-4 grid gap-2 md:grid-cols-2" on:submit|preventDefault={onCreateIntegrationEvent}>
        <select class="field" bind:value={eventForm.provider}>
          <option value="google">google</option>
          <option value="microsoft">microsoft</option>
        </select>
        <select class="field" bind:value={eventForm.integrationType}>
          <option value="calendar">calendar</option>
          <option value="email">email</option>
        </select>
        <input class="field md:col-span-2" bind:value={eventForm.eventType} placeholder="Event type" />
        <input class="field md:col-span-2" bind:value={eventForm.occurredAt} placeholder="RFC3339 timestamp" />
        <button class="btn-primary md:col-span-2" type="submit">Create Event</button>
      </form>

      {#if $integrationEventsQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $integrationEventsQuery.data.data.slice(0, 6) as row}
            <li class="glass-soft rounded-2xl p-3.5 text-sm">
              <p class="font-semibold text-slate-100">{row.eventType}</p>
              <p class="mt-1 text-xs text-slate-300">{row.provider}/{row.integrationType} - {row.occurredAt}</p>
            </li>
          {/each}
        </ul>
      {/if}
    </article>
  </section>

  <section class="stagger grid grid-cols-1 gap-5 xl:grid-cols-12">
    <article class="glass-panel rounded-3xl p-5 xl:col-span-6">
      <h2 class="section-title mb-4">Approval Workflow</h2>
      <form class="mb-4 grid gap-2 md:grid-cols-2" on:submit|preventDefault={onCreateApproval}>
        <input class="field" bind:value={approvalForm.entityType} placeholder="Entity type" />
        <input class="field" bind:value={approvalForm.entityId} placeholder="Entity ID" />
        <input class="field" bind:value={approvalForm.requestedBy} placeholder="Requested by" />
        <input class="field" bind:value={approvalForm.approverUserId} placeholder="Approver user ID" />
        <input class="field md:col-span-2" bind:value={approvalForm.reason} placeholder="Reason" />
        <button class="btn-primary md:col-span-2" type="submit">Create Approval Request</button>
      </form>

      {#if $approvalsQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $approvalsQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5 text-sm">
              <div class="flex items-center justify-between gap-2">
                <p class="font-semibold text-slate-100">{row.entityType}</p>
                <span class="status-pill">{row.status}</span>
              </div>
              <p class="mt-1 text-xs text-slate-300">{row.reason}</p>
              {#if row.status === "pending"}
                <div class="mt-2 flex gap-2">
                  <button class="btn-primary !px-3 !py-1.5 !text-xs" type="button" on:click={() => onDecideApproval(row.id, "approved")}>
                    Approve
                  </button>
                  <button class="btn-ghost" type="button" on:click={() => onDecideApproval(row.id, "rejected")}>
                    Reject
                  </button>
                </div>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
    </article>

    <article class="glass-panel rounded-3xl p-5 xl:col-span-6">
      <h2 class="section-title mb-4">CSV Operations</h2>
      <div class="mb-4 flex flex-wrap gap-2">
        <button class="btn-ghost" type="button" on:click={() => downloadCSV("/export/accounts.csv", "accounts.csv")}>
          Export Accounts CSV
        </button>
        <button class="btn-ghost" type="button" on:click={() => downloadCSV("/export/opportunities.csv", "opportunities.csv")}>
          Export Opportunities CSV
        </button>
      </div>
      <div class="grid gap-3 md:grid-cols-2">
        <div class="glass-soft rounded-2xl p-3.5">
          <p class="mb-2 text-sm font-semibold text-slate-100">Import Accounts CSV</p>
          <input class="field text-xs" type="file" accept=".csv,text/csv" on:change={(e) => (accountsCsv = (e.currentTarget as HTMLInputElement).files?.[0] ?? null)} />
          <button class="btn-primary mt-2 w-full !py-2 text-xs disabled:opacity-50" type="button" on:click={onUploadAccounts} disabled={!accountsCsv}>
            Upload Accounts
          </button>
        </div>
        <div class="glass-soft rounded-2xl p-3.5">
          <p class="mb-2 text-sm font-semibold text-slate-100">Import Opportunities CSV</p>
          <input class="field text-xs" type="file" accept=".csv,text/csv" on:change={(e) => (opportunitiesCsv = (e.currentTarget as HTMLInputElement).files?.[0] ?? null)} />
          <button class="btn-primary mt-2 w-full !py-2 text-xs disabled:opacity-50" type="button" on:click={onUploadOpportunities} disabled={!opportunitiesCsv}>
            Upload Opportunities
          </button>
        </div>
      </div>
    </article>
  </section>
</main>
