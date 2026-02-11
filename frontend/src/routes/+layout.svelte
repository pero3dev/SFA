<script lang="ts">
  import "../app.css";
  import { onMount } from "svelte";
  import { QueryClient, QueryClientProvider } from "@tanstack/svelte-query";
  import { getText, initLocale, locale, setLocale, type I18nKey, type Locale } from "$lib/i18n";

  const queryClient = new QueryClient();
  let lang: Locale = "en";
  let now = "";

  $: lang = $locale;
  $: now = new Intl.DateTimeFormat(lang === "ja" ? "ja-JP" : "en-US", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date());

  const t = (key: I18nKey): string => getText(lang, key);

  onMount(() => {
    initLocale();
  });
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
            <p class="meta-label">{t("layout_kicker")}</p>
            <h1 class="text-[1.45rem] font-bold tracking-tight text-cyan-50 md:text-[1.85rem]">
              {t("layout_title")}
            </h1>
            <p class="max-w-2xl text-sm leading-relaxed text-slate-300">
              {t("layout_description")}
            </p>
          </div>
          <div class="glass-soft rounded-2xl px-4 py-3 text-right">
            <p class="meta-label">{t("live_snapshot")}</p>
            <p class="text-sm font-semibold text-cyan-100">{now}</p>
            <p class="mt-1 text-xs text-slate-400">{t("tenant_label")}: 00000000...0001</p>
            <p class="mt-2 text-[11px] text-slate-400">{t("language")}</p>
            <div class="mt-3 flex justify-end gap-2">
              <button
                class={`btn-ghost !px-2.5 !py-1.5 !text-xs ${lang === "ja" ? "!border-cyan-200/70 !text-cyan-100" : ""}`}
                type="button"
                aria-label={t("language")}
                aria-pressed={lang === "ja"}
                on:click={() => setLocale("ja")}
              >
                {t("lang_ja")}
              </button>
              <button
                class={`btn-ghost !px-2.5 !py-1.5 !text-xs ${lang === "en" ? "!border-cyan-200/70 !text-cyan-100" : ""}`}
                type="button"
                aria-label={t("language")}
                aria-pressed={lang === "en"}
                on:click={() => setLocale("en")}
              >
                {t("lang_en")}
              </button>
            </div>
          </div>
        </div>
      </header>
      <slot />
    </div>
  </div>
</QueryClientProvider>
