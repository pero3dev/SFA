import { get, writable } from "svelte/store";

export type Locale = "ja" | "en";

const STORAGE_KEY = "sfa.locale";

export const locale = writable<Locale>("en");

export const messages = {
  en: {
    layout_kicker: "Immersive SFA Workspace",
    layout_title: "Futuristic Revenue Command Center",
    layout_description:
      "A futuristic SaaS dashboard rebuilt with Glassmorphism and Aurora UI. Deals, forecast, approvals, integrations, and CSV operations are managed in one immersive view.",
    live_snapshot: "Live Snapshot",
    tenant_label: "Tenant",
    language: "Language",
    lang_ja: "JP",
    lang_en: "EN",

    notice_next_action_updated: "Next action updated",
    notice_next_action_failed: "Failed to update next action",
    notice_integration_saved: "Integration settings saved",
    notice_integration_failed: "Failed to save integration settings",
    notice_event_created: "Integration event created",
    notice_event_failed: "Failed to create integration event",
    notice_approval_created: "Approval request created",
    notice_approval_create_failed: "Failed to create approval request",
    notice_approval_updated: "Approval updated to {status}",
    notice_approval_update_failed: "Failed to update approval",
    notice_import_accounts_done: "accounts.csv imported",
    notice_import_accounts_failed: "Failed to import accounts.csv",
    notice_import_opportunities_done: "opportunities.csv imported",
    notice_import_opportunities_failed: "Failed to import opportunities.csv",

    kpi_pulse: "KPI Pulse",
    realtime: "Realtime",
    loading_kpi: "Loading KPI stream...",
    pipeline_stage: "Pipeline Stage",
    loading_pipeline: "Loading pipeline...",
    deals: "Deals",
    amount: "Amount",

    next_action_control: "Next Action Control",
    execution: "Execution",
    opportunity_id: "Opportunity ID",
    next_action_note: "Next action note",
    update_next_action: "Update Next Action",
    deal_health_radar: "Deal Health Radar",
    loading_health: "Loading health radar...",
    last_activity: "Last activity",
    probability: "Probability",

    forecast_matrix: "Forecast Matrix",
    loss_reason_analytics: "Loss Reason Analytics",
    duplicate_intelligence: "Duplicate Intelligence",
    count: "Count",
    pipeline: "Pipeline",
    weighted: "Weighted",

    integration_hub: "Integration Hub",
    user_id: "User ID",
    external_account: "External account",
    scopes: "scope1,scope2",
    save_integration: "Save Integration",

    integration_event_stream: "Integration Event Stream",
    event_type: "Event type",
    rfc3339_timestamp: "RFC3339 timestamp",
    create_event: "Create Event",

    approval_workflow: "Approval Workflow",
    entity_type: "Entity type",
    entity_id: "Entity ID",
    requested_by: "Requested by",
    approver_user_id: "Approver user ID",
    reason: "Reason",
    create_approval_request: "Create Approval Request",
    approve: "Approve",
    reject: "Reject",

    csv_operations: "CSV Operations",
    export_accounts_csv: "Export Accounts CSV",
    export_opportunities_csv: "Export Opportunities CSV",
    import_accounts_csv: "Import Accounts CSV",
    import_opportunities_csv: "Import Opportunities CSV",
    upload_accounts: "Upload Accounts",
    upload_opportunities: "Upload Opportunities"
  },
  ja: {
    layout_kicker: "没入型 SFA ワークスペース",
    layout_title: "未来型レベニュー・コマンドセンター",
    layout_description:
      "Glassmorphism と Aurora UI で再構築した未来的な SaaS ダッシュボード。案件、予測、承認、連携、CSV操作を1画面で管理できます。",
    live_snapshot: "ライブスナップショット",
    tenant_label: "テナント",
    language: "言語",
    lang_ja: "JP",
    lang_en: "EN",

    notice_next_action_updated: "次アクションを更新しました",
    notice_next_action_failed: "次アクションの更新に失敗しました",
    notice_integration_saved: "連携設定を保存しました",
    notice_integration_failed: "連携設定の保存に失敗しました",
    notice_event_created: "連携イベントを作成しました",
    notice_event_failed: "連携イベントの作成に失敗しました",
    notice_approval_created: "承認申請を作成しました",
    notice_approval_create_failed: "承認申請の作成に失敗しました",
    notice_approval_updated: "承認ステータスを{status}に更新しました",
    notice_approval_update_failed: "承認の更新に失敗しました",
    notice_import_accounts_done: "accounts.csv を取り込みました",
    notice_import_accounts_failed: "accounts.csv の取り込みに失敗しました",
    notice_import_opportunities_done: "opportunities.csv を取り込みました",
    notice_import_opportunities_failed: "opportunities.csv の取り込みに失敗しました",

    kpi_pulse: "KPI パルス",
    realtime: "リアルタイム",
    loading_kpi: "KPIを読み込み中...",
    pipeline_stage: "パイプラインステージ",
    loading_pipeline: "パイプラインを読み込み中...",
    deals: "案件数",
    amount: "金額",

    next_action_control: "次アクション管理",
    execution: "実行",
    opportunity_id: "案件ID",
    next_action_note: "次アクションメモ",
    update_next_action: "次アクションを更新",
    deal_health_radar: "案件ヘルスレーダー",
    loading_health: "ヘルス情報を読み込み中...",
    last_activity: "最終活動",
    probability: "確度",

    forecast_matrix: "予測マトリクス",
    loss_reason_analytics: "失注理由分析",
    duplicate_intelligence: "重複インテリジェンス",
    count: "件数",
    pipeline: "パイプライン",
    weighted: "加重",

    integration_hub: "連携ハブ",
    user_id: "ユーザーID",
    external_account: "外部アカウント",
    scopes: "scope1,scope2",
    save_integration: "連携設定を保存",

    integration_event_stream: "連携イベントストリーム",
    event_type: "イベント種別",
    rfc3339_timestamp: "RFC3339タイムスタンプ",
    create_event: "イベントを作成",

    approval_workflow: "承認ワークフロー",
    entity_type: "対象タイプ",
    entity_id: "対象ID",
    requested_by: "申請者ID",
    approver_user_id: "承認者ID",
    reason: "理由",
    create_approval_request: "承認申請を作成",
    approve: "承認",
    reject: "却下",

    csv_operations: "CSV操作",
    export_accounts_csv: "アカウントCSVをエクスポート",
    export_opportunities_csv: "案件CSVをエクスポート",
    import_accounts_csv: "アカウントCSVをインポート",
    import_opportunities_csv: "案件CSVをインポート",
    upload_accounts: "アカウントをアップロード",
    upload_opportunities: "案件をアップロード"
  }
} as const;

export type I18nKey = keyof (typeof messages)["en"];

const stageLabels: Record<Locale, Record<string, string>> = {
  en: {
    new_lead: "New Lead",
    qualified: "Qualified",
    proposal: "Proposal",
    negotiation: "Negotiation",
    closed_won: "Closed Won",
    closed_lost: "Closed Lost"
  },
  ja: {
    new_lead: "新規リード",
    qualified: "有望案件",
    proposal: "提案",
    negotiation: "交渉",
    closed_won: "受注",
    closed_lost: "失注"
  }
};

const approvalStatusLabels: Record<Locale, Record<string, string>> = {
  en: { pending: "Pending", approved: "Approved", rejected: "Rejected" },
  ja: { pending: "保留", approved: "承認済み", rejected: "却下" }
};

const integrationStatusLabels: Record<Locale, Record<string, string>> = {
  en: { active: "Active", revoked: "Revoked", error: "Error" },
  ja: { active: "有効", revoked: "無効", error: "エラー" }
};

const lossReasonLabels: Record<Locale, Record<string, string>> = {
  en: {
    budget: "Budget",
    competitor: "Competitor",
    timing: "Timing",
    no_decision: "No Decision",
    other: "Other"
  },
  ja: {
    budget: "予算不足",
    competitor: "競合選定",
    timing: "時期未定",
    no_decision: "見送り",
    other: "その他"
  }
};

export function getText(lang: Locale, key: I18nKey): string {
  return messages[lang][key] ?? messages.en[key];
}

export function stageLabel(lang: Locale, value: string): string {
  return stageLabels[lang][value] ?? value;
}

export function approvalStatusLabel(lang: Locale, value: string): string {
  return approvalStatusLabels[lang][value] ?? value;
}

export function integrationStatusLabel(lang: Locale, value: string): string {
  return integrationStatusLabels[lang][value] ?? value;
}

export function lossReasonLabel(lang: Locale, value: string): string {
  return lossReasonLabels[lang][value] ?? value;
}

export function detectLocale(): Locale {
  if (typeof window === "undefined") return "en";
  const stored = window.localStorage.getItem(STORAGE_KEY);
  if (stored === "ja" || stored === "en") return stored;
  return window.navigator.language.toLowerCase().startsWith("ja") ? "ja" : "en";
}

export function initLocale(): void {
  if (typeof window === "undefined") return;
  locale.set(detectLocale());
}

export function setLocale(next: Locale): void {
  locale.set(next);
  if (typeof window !== "undefined") {
    window.localStorage.setItem(STORAGE_KEY, next);
  }
}

export function toggleLocale(): void {
  setLocale(get(locale) === "ja" ? "en" : "ja");
}

export function interpolate(template: string, values: Record<string, string>): string {
  let result = template;
  Object.entries(values).forEach(([key, value]) => {
    result = result.replaceAll(`{${key}}`, value);
  });
  return result;
}
