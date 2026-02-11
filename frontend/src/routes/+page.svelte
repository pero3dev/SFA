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
  import {
    approvalStatusLabel,
    duplicateTypeLabel,
    getText,
    integrationStatusLabel,
    interpolate,
    kpiMetricLabel,
    locale,
    lossReasonLabel,
    stageLabel,
    type I18nKey,
    type Locale
  } from "$lib/i18n";

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

  let lang: Locale = "en";
  let notice = "";

  $: lang = $locale;

  const t = (key: I18nKey): string => getText(lang, key);
  const localeTag = (): string => (lang === "ja" ? "ja-JP" : "en-US");
  const formatNumber = (value: number): string => value.toLocaleString(localeTag());

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
      notice = t("notice_next_action_updated");
    } catch (error) {
      notice = `${t("notice_next_action_failed")}: ${String(error)}`;
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
        scopes: integrationForm.scopes
          .split(",")
          .map((s) => s.trim())
          .filter((s) => s.length > 0)
      });
      await refresh([["integrations", "connections"]]);
      notice = t("notice_integration_saved");
    } catch (error) {
      notice = `${t("notice_integration_failed")}: ${String(error)}`;
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
      notice = t("notice_event_created");
    } catch (error) {
      notice = `${t("notice_event_failed")}: ${String(error)}`;
    }
  }

  async function onCreateApproval(): Promise<void> {
    try {
      await createApproval(approvalForm);
      await refresh([["approvals"]]);
      notice = t("notice_approval_created");
    } catch (error) {
      notice = `${t("notice_approval_create_failed")}: ${String(error)}`;
    }
  }

  async function onDecideApproval(id: string, status: "approved" | "rejected"): Promise<void> {
    try {
      await decideApproval(id, { status, decisionNote: `UI decision: ${status}` });
      await refresh([["approvals"]]);
      notice = interpolate(t("notice_approval_updated"), {
        status: approvalStatusLabel(lang, status)
      });
    } catch (error) {
      notice = `${t("notice_approval_update_failed")}: ${String(error)}`;
    }
  }

  async function onUploadAccounts(): Promise<void> {
    if (!accountsCsv) return;
    try {
      await uploadCSV("/import/accounts.csv", accountsCsv);
      notice = t("notice_import_accounts_done");
      await refresh([["analytics", "duplicates"]]);
    } catch (error) {
      notice = `${t("notice_import_accounts_failed")}: ${String(error)}`;
    }
  }

  async function onUploadOpportunities(): Promise<void> {
    if (!opportunitiesCsv) return;
    try {
      await uploadCSV("/import/opportunities.csv", opportunitiesCsv);
      notice = t("notice_import_opportunities_done");
      await refresh([["opportunities", "next-actions"], ["analytics", "forecast"], ["analytics", "deal-health"]]);
    } catch (error) {
      notice = `${t("notice_import_opportunities_failed")}: ${String(error)}`;
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
        <h2 class="section-title">{t("kpi_pulse")}</h2>
        <span class="status-pill">{t("realtime")}</span>
      </div>
      {#if $kpiQuery.isSuccess}
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {#each $kpiQuery.data.data as item}
            <div class="glass-soft rounded-2xl p-4">
              <p class="meta-label">{kpiMetricLabel(lang, item.metricKey)}</p>
              <p class="kpi-value mt-2 text-cyan-100">{formatNumber(item.metricValue)}</p>
            </div>
          {/each}
        </div>
      {:else}
        <p class="text-sm text-slate-300">{t("loading_kpi")}</p>
      {/if}
    </article>

    <article class="glass-panel rounded-3xl p-5 xl:col-span-4">
      <h2 class="section-title mb-4">{t("pipeline_stage")}</h2>
      {#if $pipelineQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $pipelineQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5">
              <p class="text-sm font-semibold text-slate-100">{stageLabel(lang, row.stage)}</p>
              <p class="mt-1 text-xs text-slate-300">
                {t("deals")} {row.count} - {t("amount")} {formatNumber(row.totalAmount)}
              </p>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="text-sm text-slate-300">{t("loading_pipeline")}</p>
      {/if}
    </article>
  </section>

  <section class="stagger grid grid-cols-1 gap-5 xl:grid-cols-12">
    <article class="glass-panel rounded-3xl p-5 xl:col-span-6">
      <div class="mb-4 flex items-center justify-between">
        <h2 class="section-title">{t("next_action_control")}</h2>
        <span class="status-pill">{t("execution")}</span>
      </div>
      <form class="mb-4 grid gap-2 md:grid-cols-3" on:submit|preventDefault={onUpdateNextAction}>
        <input class="field" bind:value={nextActionForm.id} placeholder={t("opportunity_id")} />
        <input class="field" bind:value={nextActionForm.nextActionAt} type="datetime-local" />
        <input class="field" bind:value={nextActionForm.nextActionNote} placeholder={t("next_action_note")} />
        <button class="btn-primary md:col-span-3" type="submit">{t("update_next_action")}</button>
      </form>
      {#if $nextActionsQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $nextActionsQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <p class="text-sm font-semibold text-slate-100">{row.name}</p>
                <span class="status-pill">{stageLabel(lang, row.stage)}</span>
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
      <h2 class="section-title mb-4">{t("deal_health_radar")}</h2>
      {#if $healthQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $healthQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5">
              <div class="flex items-center justify-between gap-2">
                <p class="text-sm font-semibold text-slate-100">{row.name}</p>
                <p class={`text-sm font-bold ${statusTone(row.healthScore)}`}>{row.healthScore}</p>
              </div>
              <p class="mt-1 text-xs text-slate-300">{stageLabel(lang, row.stage)} - {t("last_activity")} {row.lastActivityAt}</p>
              <p class="mt-1 text-xs text-slate-400">
                {t("probability")} {row.probability}% - {t("amount")} {formatNumber(row.amount)}
              </p>
            </li>
          {/each}
        </ul>
      {:else}
        <p class="text-sm text-slate-300">{t("loading_health")}</p>
      {/if}
    </article>
  </section>

  <section class="stagger grid grid-cols-1 gap-5 xl:grid-cols-12">
    <article class="glass-panel rounded-3xl p-5 xl:col-span-4">
      <h2 class="section-title mb-4">{t("forecast_matrix")}</h2>
      {#if $forecastQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $forecastQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5 text-sm">
              <p class="font-semibold text-slate-100">{row.month}</p>
              <p class="mt-1 text-xs text-slate-300">{t("deals")} {row.dealCount}</p>
              <p class="mt-1 text-cyan-100">{t("pipeline")} {formatNumber(row.pipelineAmount)}</p>
              <p class="text-violet-200">{t("weighted")} {formatNumber(row.weightedAmount)}</p>
            </li>
          {/each}
        </ul>
      {/if}
    </article>

    <article class="glass-panel rounded-3xl p-5 xl:col-span-4">
      <h2 class="section-title mb-4">{t("loss_reason_analytics")}</h2>
      {#if $lossQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $lossQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5 text-sm">
              <p class="font-semibold text-slate-100">{lossReasonLabel(lang, row.reason)}</p>
              <p class="mt-1 text-xs text-slate-300">{t("count")} {row.lostCount}</p>
              <p class="mt-1 text-rose-200">{t("amount")} {formatNumber(row.lostAmount)}</p>
            </li>
          {/each}
        </ul>
      {/if}
    </article>

    <article class="glass-panel rounded-3xl p-5 xl:col-span-4">
      <h2 class="section-title mb-4">{t("duplicate_intelligence")}</h2>
      {#if $duplicatesQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $duplicatesQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5 text-sm">
              <p class="font-semibold text-slate-100">{duplicateTypeLabel(lang, row.type)}</p>
              <p class="mt-1 break-all text-xs text-slate-300">{row.matchValue}</p>
            </li>
          {/each}
        </ul>
      {/if}
    </article>
  </section>

  <section class="stagger grid grid-cols-1 gap-5 xl:grid-cols-12">
    <article class="glass-panel rounded-3xl p-5 xl:col-span-6">
      <h2 class="section-title mb-4">{t("integration_hub")}</h2>
      <form class="mb-4 grid gap-2 md:grid-cols-2" on:submit|preventDefault={onSaveIntegration}>
        <input class="field" bind:value={integrationForm.userId} placeholder={t("user_id")} />
        <input class="field" bind:value={integrationForm.externalAccountId} placeholder={t("external_account")} />
        <select class="field" bind:value={integrationForm.provider}>
          <option value="google">google</option>
          <option value="microsoft">microsoft</option>
        </select>
        <select class="field" bind:value={integrationForm.integrationType}>
          <option value="calendar">calendar</option>
          <option value="email">email</option>
        </select>
        <input class="field md:col-span-2" bind:value={integrationForm.scopes} placeholder={t("scopes")} />
        <button class="btn-primary md:col-span-2" type="submit">{t("save_integration")}</button>
      </form>

      {#if $integrationsQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $integrationsQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5 text-sm">
              <div class="flex items-center justify-between gap-2">
                <p class="font-semibold text-slate-100">{row.provider} / {row.integrationType}</p>
                <span class="status-pill">{integrationStatusLabel(lang, row.status)}</span>
              </div>
              <p class="mt-1 text-xs text-slate-300">{row.externalAccountId}</p>
            </li>
          {/each}
        </ul>
      {/if}
    </article>

    <article class="glass-panel rounded-3xl p-5 xl:col-span-6">
      <h2 class="section-title mb-4">{t("integration_event_stream")}</h2>
      <form class="mb-4 grid gap-2 md:grid-cols-2" on:submit|preventDefault={onCreateIntegrationEvent}>
        <select class="field" bind:value={eventForm.provider}>
          <option value="google">google</option>
          <option value="microsoft">microsoft</option>
        </select>
        <select class="field" bind:value={eventForm.integrationType}>
          <option value="calendar">calendar</option>
          <option value="email">email</option>
        </select>
        <input class="field md:col-span-2" bind:value={eventForm.eventType} placeholder={t("event_type")} />
        <input class="field md:col-span-2" bind:value={eventForm.occurredAt} placeholder={t("rfc3339_timestamp")} />
        <button class="btn-primary md:col-span-2" type="submit">{t("create_event")}</button>
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
      <h2 class="section-title mb-4">{t("approval_workflow")}</h2>
      <form class="mb-4 grid gap-2 md:grid-cols-2" on:submit|preventDefault={onCreateApproval}>
        <input class="field" bind:value={approvalForm.entityType} placeholder={t("entity_type")} />
        <input class="field" bind:value={approvalForm.entityId} placeholder={t("entity_id")} />
        <input class="field" bind:value={approvalForm.requestedBy} placeholder={t("requested_by")} />
        <input class="field" bind:value={approvalForm.approverUserId} placeholder={t("approver_user_id")} />
        <input class="field md:col-span-2" bind:value={approvalForm.reason} placeholder={t("reason")} />
        <button class="btn-primary md:col-span-2" type="submit">{t("create_approval_request")}</button>
      </form>

      {#if $approvalsQuery.isSuccess}
        <ul class="space-y-2.5">
          {#each $approvalsQuery.data.data as row}
            <li class="glass-soft rounded-2xl p-3.5 text-sm">
              <div class="flex items-center justify-between gap-2">
                <p class="font-semibold text-slate-100">{row.entityType}</p>
                <span class="status-pill">{approvalStatusLabel(lang, row.status)}</span>
              </div>
              <p class="mt-1 text-xs text-slate-300">{row.reason}</p>
              {#if row.status === "pending"}
                <div class="mt-2 flex gap-2">
                  <button class="btn-primary !px-3 !py-1.5 !text-xs" type="button" on:click={() => onDecideApproval(row.id, "approved")}>
                    {t("approve")}
                  </button>
                  <button class="btn-ghost" type="button" on:click={() => onDecideApproval(row.id, "rejected")}>{t("reject")}</button>
                </div>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
    </article>

    <article class="glass-panel rounded-3xl p-5 xl:col-span-6">
      <h2 class="section-title mb-4">{t("csv_operations")}</h2>
      <div class="mb-4 flex flex-wrap gap-2">
        <button class="btn-ghost" type="button" on:click={() => downloadCSV("/export/accounts.csv", "accounts.csv")}>
          {t("export_accounts_csv")}
        </button>
        <button class="btn-ghost" type="button" on:click={() => downloadCSV("/export/opportunities.csv", "opportunities.csv")}>
          {t("export_opportunities_csv")}
        </button>
      </div>
      <div class="grid gap-3 md:grid-cols-2">
        <div class="glass-soft rounded-2xl p-3.5">
          <p class="mb-2 text-sm font-semibold text-slate-100">{t("import_accounts_csv")}</p>
          <input
            class="field text-xs"
            type="file"
            accept=".csv,text/csv"
            on:change={(e) => (accountsCsv = (e.currentTarget as HTMLInputElement).files?.[0] ?? null)}
          />
          <button
            class="btn-primary mt-2 w-full !py-2 text-xs disabled:opacity-50"
            type="button"
            on:click={onUploadAccounts}
            disabled={!accountsCsv}
          >
            {t("upload_accounts")}
          </button>
        </div>
        <div class="glass-soft rounded-2xl p-3.5">
          <p class="mb-2 text-sm font-semibold text-slate-100">{t("import_opportunities_csv")}</p>
          <input
            class="field text-xs"
            type="file"
            accept=".csv,text/csv"
            on:change={(e) => (opportunitiesCsv = (e.currentTarget as HTMLInputElement).files?.[0] ?? null)}
          />
          <button
            class="btn-primary mt-2 w-full !py-2 text-xs disabled:opacity-50"
            type="button"
            on:click={onUploadOpportunities}
            disabled={!opportunitiesCsv}
          >
            {t("upload_opportunities")}
          </button>
        </div>
      </div>
    </article>
  </section>
</main>
