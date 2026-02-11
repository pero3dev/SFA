<script lang="ts">
  import "../app.css";
  import { QueryClient, QueryClientProvider } from "@tanstack/svelte-query";

  const queryClient = new QueryClient();
  const now = new Intl.DateTimeFormat("ja-JP", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date());
</script>

<QueryClientProvider client={queryClient}>
  <div class="aurora-shell">
    <div aria-hidden="true" class="aurora-blob a"></div>
    <div aria-hidden="true" class="aurora-blob b"></div>
    <div aria-hidden="true" class="aurora-blob c"></div>

    <div class="mx-auto max-w-[1200px] px-4 pb-12 pt-6 md:px-8">
      <header class="glass-panel fade-in mb-7 rounded-3xl px-5 py-5 md:px-7 md:py-6">
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div class="space-y-2">
            <p class="meta-label">Immersive SFA Workspace</p>
            <h1 class="text-[1.45rem] font-bold tracking-tight text-cyan-50 md:text-[1.85rem]">
              Futuristic Revenue Command Center
            </h1>
            <p class="max-w-2xl text-sm leading-relaxed text-slate-300">
              A futuristic SaaS dashboard rebuilt with Glassmorphism and Aurora UI. Deals, forecast,
              approvals, integrations, and CSV operations are managed in one immersive view.
            </p>
          </div>
          <div class="glass-soft rounded-2xl px-4 py-3 text-right">
            <p class="meta-label">Live Snapshot</p>
            <p class="text-sm font-semibold text-cyan-100">{now}</p>
            <p class="mt-1 text-xs text-slate-400">Tenant: 00000000...0001</p>
          </div>
        </div>
      </header>
      <slot />
    </div>
  </div>
</QueryClientProvider>
