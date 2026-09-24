<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import {
    privateDecisionHistoryAdapter,
    syntheticDecisionHistory,
    type DecisionHistoryCase,
    type DecisionHistoryField,
    type DecisionHistoryMetricTone,
    type DecisionHistoryTheme,
    type PrivateDecisionHistoryPayload
  } from "$lib/decision_history";

  type LoadState = "loading" | "ready" | "unavailable";
  type IndexItem = {
    id: string;
    title: string;
    period: string;
    kind: "theme" | "case";
  };

  let loadState: LoadState = "loading";
  let decisionCases: DecisionHistoryCase[] = [];
  let decisionThemes: DecisionHistoryTheme[] = [];
  let indexItems: IndexItem[] = [];
  let isFixture = false;
  let activeItemId = "";

  $: indexItems = [
    ...decisionThemes.map((item) => ({
      id: item.id,
      title: item.title,
      period: item.period,
      kind: "theme" as const
    })),
    ...decisionCases.map((item) => ({
      id: item.id,
      title: item.title,
      period: item.period,
      kind: "case" as const
    }))
  ];

  onMount(async () => {
    isFixture = $page.url.searchParams.get("fixture") === "synthetic";

    if (isFixture) {
      applyPayload(syntheticDecisionHistory);
      loadState = "ready";
      return;
    }

    try {
      const result = await privateDecisionHistoryAdapter.load(fetch);
      if (result.kind === "ready") {
        applyPayload(result.payload);
        loadState = "ready";
      } else {
        loadState = "unavailable";
      }
    } catch (_error) {
      loadState = "unavailable";
    }
  });

  function applyPayload(payload: PrivateDecisionHistoryPayload) {
    decisionCases = payload.cases;
    decisionThemes = payload.themes ?? [];
    activeItemId = indexItems[0]?.id ?? payload.themes?.[0]?.id ?? payload.cases[0]?.id ?? "";
  }

  function retryLoading() {
    window.location.reload();
  }

  function selectItem(id: string) {
    activeItemId = id;
  }

  function fieldLabel(status: DecisionHistoryField["status"]) {
    if (status === "known") return "Известно";
    if (status === "partial") return "Частично";
    return "Неизвестно";
  }

  function attributionFields(item: DecisionHistoryCase): [string, DecisionHistoryField][] {
    return [
      ["Валютный эффект", item.attribution.fx],
      ["Комиссии", item.attribution.fees],
      ["Налоги", item.attribution.taxes],
      ["Сценарий сравнения в RUB", item.attribution.rubCounterfactual]
    ];
  }

  function metricToneClass(tone: DecisionHistoryMetricTone | undefined) {
    return tone ? `metric-${tone}` : "metric-neutral";
  }

  function sectionCountLabel(count: number) {
    const remainderTen = count % 10;
    const remainderHundred = count % 100;

    if (remainderTen === 1 && remainderHundred !== 11) return "раздел";
    if (
      remainderTen >= 2 &&
      remainderTen <= 4 &&
      (remainderHundred < 12 || remainderHundred > 14)
    ) {
      return "раздела";
    }
    return "разделов";
  }
</script>

<svelte:head>
  <title>История решений · Paisa</title>
  <meta
    name="description"
    content="История финансовых решений с выводами, денежными потоками, источниками и границами доказательств."
  />
</svelte:head>

<section class="section decision-history-page">
  <div class="container is-fluid">
    {#if loadState === "loading"}
      <div
        class="decision-history-loading"
        data-testid="decision-history-loading"
        role="status"
        aria-live="polite"
      >
        <span class="loader" aria-hidden="true"></span>
        <span>Загрузка истории решений…</span>
      </div>
    {:else if loadState === "unavailable"}
      <div class="decision-history-empty" data-testid="decision-history-unavailable">
        <p class="title is-4">История решений недоступна</p>
        <p>Не удалось загрузить данные. Проверьте подключение и повторите попытку.</p>
        <button class="button is-link is-light" type="button" on:click={retryLoading}>
          Повторить загрузку
        </button>
      </div>
    {:else}
      <div class="decision-history-heading">
        <div class="page-heading-copy">
          <h1 class="title is-3 mb-1">История решений</h1>
          <p class="subtitle is-6 mb-0">
            Не лента операций, а разбор того, что изменилось после решения и насколько это видно в
            Ledger.
          </p>
        </div>
      </div>

      <div class="interpretation-note" data-testid="decision-history-ready">
        <span class="icon is-small" aria-hidden="true"><i class="fas fa-compass"></i></span>
        <div>
          <strong>Как читать этот экран</strong>
          <span>
            Сначала показан вывод и измеримые суммы, ниже — доказательства и ограничения. Перевод
            денег, доход и рыночную переоценку не смешиваем в одну доходность.
          </span>
        </div>
      </div>

      <div class="history-layout">
        <aside class="history-index" aria-label="Индекс решений">
          <div class="index-heading">
            <span class="eyebrow">Индекс решений</span>
            <span class="index-count">{indexItems.length} {sectionCountLabel(indexItems.length)}</span>
          </div>
          <ol>
            {#each indexItems as item, index}
              <li>
                <a
                  href={`#${item.id}`}
                  class:active={activeItemId === item.id}
                  aria-current={activeItemId === item.id ? "true" : undefined}
                  on:click={() => selectItem(item.id)}
                >
                  <span class="index-number">{String(index + 1).padStart(2, "0")}</span>
                  <span class="index-copy">
                    <strong>{item.title}</strong>
                    <small>{item.kind === "theme" ? "Главный блок" : item.period}</small>
                  </span>
                </a>
              </li>
            {/each}
          </ol>
          <p class="index-footnote">
            Вверху — решения с собранным результатом. Внизу — остальные маршруты и поля, которые
            ещё требуют отдельной сверки.
          </p>
        </aside>

        <div class="history-content">
          {#each decisionThemes as theme}
            <article
              id={theme.id}
              class="decision-theme"
              class:active-theme={activeItemId === theme.id}
              data-testid={`decision-theme-${theme.id}`}
            >
              <header class="theme-header">
                <div class="theme-title">
                  <span class="case-period">{theme.period}</span>
                  <h2>{theme.title}</h2>
                  <p>{theme.decision}</p>
                </div>
                <div class="theme-result">
                  <span class="field-label">Что видно сейчас</span>
                  <strong>{theme.headline}</strong>
                </div>
              </header>

              <p class="theme-summary">{theme.summary}</p>

              <div class="theme-metrics">
                {#each theme.metrics as metric}
                  <div class={`metric-card ${metricToneClass(metric.tone)}`}>
                    <span class="field-label">{metric.label}</span>
                    <strong>{metric.value}</strong>
                    {#if metric.detail}<small>{metric.detail}</small>{/if}
                  </div>
                {/each}
              </div>

              <div class="branch-stack">
                {#each theme.branches as branch}
                  <section class="decision-branch" id={branch.id}>
                    <div class="branch-marker" aria-hidden="true">{branch.number}</div>
                    <div class="branch-content">
                      <header class="branch-header">
                        <div>
                          <span class="case-period">{branch.period}</span>
                          <h3>{branch.title}</h3>
                        </div>
                        <strong>{branch.headline}</strong>
                      </header>

                      <div class="branch-grid">
                        <div class="branch-narrative">
                          <span class="field-label">Решение</span>
                          <p class="decision-copy">{branch.decision}</p>
                          <div class="outcome-block">
                            <span class="field-label">Результат, который уже можно утверждать</span>
                            <p>{branch.outcome}</p>
                          </div>
                        </div>

                        <div class="branch-evidence">
                          <div class="branch-metrics">
                            {#each branch.metrics as metric}
                              <div class={`metric-card compact ${metricToneClass(metric.tone)}`}>
                                <span class="field-label">{metric.label}</span>
                                <strong>{metric.value}</strong>
                                {#if metric.detail}<small>{metric.detail}</small>{/if}
                              </div>
                            {/each}
                          </div>

                          {#if branch.timeline?.length}
                            <div class="timeline-block">
                              <span class="field-label">Динамика по месяцам</span>
                              <div class="timeline-table-wrap">
                                <table class="timeline-table">
                                  <thead>
                                    <tr>
                                      <th scope="col">Период</th>
                                      <th scope="col">Движение</th>
                                      <th scope="col">Что за ним стоит</th>
                                    </tr>
                                  </thead>
                                  <tbody>
                                    {#each branch.timeline as row}
                                      <tr>
                                        <th scope="row">{row.label}</th>
                                        <td>{row.value}</td>
                                        <td>{row.detail}</td>
                                      </tr>
                                    {/each}
                                  </tbody>
                                </table>
                              </div>
                            </div>
                          {/if}
                        </div>
                      </div>

                      <footer class="evidence-footer">
                        <div>
                          <span class="field-label">Доказательства</span>
                          <ul class="source-list">
                            {#each branch.sourceRefs as source}
                              <li><span class="source-mark" aria-hidden="true">↗</span>{source}</li>
                            {/each}
                          </ul>
                        </div>
                        <div>
                          <span class="field-label">Границы вывода</span>
                          <ul class="caveat-list">
                            {#each branch.caveats as caveat}
                              <li>{caveat}</li>
                            {/each}
                          </ul>
                        </div>
                      </footer>
                    </div>
                  </section>
                {/each}
              </div>

              <footer class="theme-footer">
                <div>
                  <span class="field-label">Общие источники</span>
                  <ul class="source-list">
                    {#each theme.sourceRefs as source}
                      <li><span class="source-mark" aria-hidden="true">↗</span>{source}</li>
                    {/each}
                  </ul>
                </div>
                <div>
                  <span class="field-label">Что не смешиваем</span>
                  <ul class="caveat-list">
                    {#each theme.caveats as caveat}
                      <li>{caveat}</li>
                    {/each}
                  </ul>
                </div>
              </footer>
            </article>
          {/each}

          {#if decisionCases.length}
            <div class="secondary-heading">
              <span class="eyebrow">Другие решения</span>
              <p>Те же принципы: сначала итог, затем атрибуция и незакрытые вопросы.</p>
            </div>
          {/if}

          {#each decisionCases as item, index}
            <article
              id={item.id}
              class="decision-case"
              class:active-case={activeItemId === item.id}
              data-testid={`decision-case-${item.id}`}
            >
              <header class="case-header">
                <div class="case-number" aria-hidden="true">{String(index + 1).padStart(2, "0")}</div>
                <div class="case-title">
                  <span class="case-period">{item.period}</span>
                  <h2>{item.title}</h2>
                </div>
                <span class="status-tag">{item.currentStatus}</span>
              </header>

              <div class="case-body">
                <div class="narrative-column">
                  <div class="narrative-block">
                    <span class="field-label">Решение</span>
                    <p class="decision-copy">{item.decision}</p>
                  </div>
                  <div class="narrative-block">
                    <span class="field-label">Контекст</span>
                    <p>{item.context}</p>
                  </div>
                  <div class="narrative-block outcome-block">
                    <span class="field-label">Итог</span>
                    <p>{item.outcome}</p>
                  </div>
                </div>

                <div class="evidence-column">
                  <div class="amount-panel">
                    <span class="field-label">Сумма / результат</span>
                    <strong>{item.amount.label}</strong>
                    <span>{item.amount.currency}</span>
                  </div>

                  <div class="attribution-grid" aria-label="Поля атрибуции">
                    {#each attributionFields(item) as [label, field]}
                      <div class="evidence-field" class:field-unknown={field.status === "unknown"}>
                        <div class="evidence-field-heading">
                          <span class="field-label">{label}</span>
                          <span class="field-status">{fieldLabel(field.status)}</span>
                        </div>
                        <p>{field.value}</p>
                        <small>{field.detail}</small>
                      </div>
                    {/each}
                  </div>
                </div>
              </div>

              <footer class="case-footer">
                <div class="footer-section">
                  <span class="field-label">Ссылки на источники</span>
                  <ul class="source-list">
                    {#each item.sourceRefs as source}
                      <li><span class="source-mark" aria-hidden="true">↗</span>{source}</li>
                    {/each}
                  </ul>
                </div>
                <div class="footer-section caveat-section">
                  <span class="field-label">Ограничения доказательств</span>
                  <ul class="caveat-list">
                    {#each item.caveats as caveat}
                      <li>{caveat}</li>
                    {/each}
                  </ul>
                </div>
              </footer>
            </article>
          {/each}
        </div>
      </div>
    {/if}
  </div>
</section>

<style lang="scss">
  @import "bulma/sass/utilities/_all.sass";

  .decision-history-page {
    padding-bottom: 4rem;
    padding-top: 2.5rem;
  }

  .decision-history-heading {
    margin-bottom: 1.25rem;
  }

  .page-heading-copy {
    display: grid;
    gap: 0.4rem;
    min-width: 0;
  }

  .page-heading-copy > .title,
  .page-heading-copy > .subtitle {
    margin-bottom: 0;
    margin-top: 0;
  }

  .decision-history-heading .subtitle {
    color: color-mix(in srgb, currentColor 68%, transparent);
    max-width: 54rem;
  }

  .decision-theme,
  .decision-case {
    border: 1px solid color-mix(in srgb, currentColor 18%, transparent);
    border-radius: $radius;
  }

  .theme-title h2 {
    font-size: 1.35rem;
    font-weight: 700;
    line-height: 1.2;
    margin: 0.35rem 0 0;
  }

  .theme-title p,
  .theme-summary {
    color: color-mix(in srgb, currentColor 76%, transparent);
    line-height: 1.55;
    margin: 0.55rem 0 0;
    max-width: 70ch;
  }

  .theme-result {
    border-left: 2px solid color-mix(in srgb, #{$link} 70%, transparent);
    display: grid;
    gap: 0.35rem;
    padding-left: 1rem;
  }

  .theme-result strong {
    font-size: 1rem;
    line-height: 1.4;
  }

  .theme-metrics,
  .branch-metrics {
    display: grid;
    gap: 0.7rem;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    margin-top: 1.25rem;
  }

  .theme-metrics {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .metric-card {
    background: color-mix(in srgb, currentColor 5%, transparent);
    border-top: 2px solid color-mix(in srgb, currentColor 28%, transparent);
    display: grid;
    gap: 0.28rem;
    min-width: 0;
    padding: 0.72rem 0.8rem 0.8rem;
  }

  .metric-card strong {
    font-family: $family-monospace;
    font-size: 1.1rem;
    font-weight: 700;
    letter-spacing: -0.03em;
    line-height: 1.2;
  }

  .metric-card small {
    color: color-mix(in srgb, currentColor 67%, transparent);
    font-size: 0.7rem;
    line-height: 1.4;
  }

  .metric-accent {
    border-top-color: color-mix(in srgb, #{$link} 72%, transparent);
  }

  .metric-positive {
    border-top-color: color-mix(in srgb, #{$success} 78%, transparent);
  }

  .metric-negative {
    border-top-color: color-mix(in srgb, #{$danger} 72%, transparent);
  }

  .interpretation-note {
    align-items: flex-start;
    background: color-mix(in srgb, #{$link} 8%, transparent);
    border: 1px solid color-mix(in srgb, #{$link} 24%, transparent);
    border-radius: $radius;
    display: flex;
    gap: 0.75rem;
    line-height: 1.5;
    margin-bottom: 2rem;
    padding: 0.9rem 1rem;
  }

  .interpretation-note .icon {
    color: $link;
    margin-top: 0.15rem;
  }

  .interpretation-note strong {
    display: block;
    font-size: 0.8rem;
    letter-spacing: 0.02em;
    margin-bottom: 0.15rem;
    text-transform: uppercase;
  }

  .history-layout {
    align-items: start;
    display: grid;
    gap: 2rem;
    grid-template-columns: minmax(12rem, 16rem) minmax(0, 1fr);
  }

  .history-index {
    position: sticky;
    top: 1.25rem;
  }

  .index-heading {
    align-items: baseline;
    border-bottom: 1px solid color-mix(in srgb, currentColor 18%, transparent);
    display: flex;
    justify-content: space-between;
    padding: 0 0.25rem 0.7rem;
  }

  .eyebrow,
  .field-label,
  .case-period {
    color: color-mix(in srgb, currentColor 68%, transparent);
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .index-count {
    color: color-mix(in srgb, currentColor 52%, transparent);
    font-size: 0.7rem;
  }

  .history-index ol {
    list-style: none;
    margin: 0;
    padding: 0.75rem 0;
  }

  .history-index a {
    align-items: flex-start;
    border-radius: $radius;
    color: inherit;
    display: flex;
    gap: 0.75rem;
    margin: 0.25rem 0;
    padding: 0.7rem 0.55rem;
    transition: background-color 150ms ease, color 150ms ease;
  }

  .history-index a:hover,
  .history-index a:focus-visible,
  .history-index a.active {
    background: color-mix(in srgb, currentColor 7%, transparent);
    color: $link;
  }

  .history-index a:focus-visible,
  .history-index a.active {
    box-shadow: 0 0 0 2px color-mix(in srgb, #{$link} 36%, transparent);
    outline: none;
  }

  .index-number {
    color: color-mix(in srgb, currentColor 52%, transparent);
    font-family: $family-monospace;
    font-size: 0.72rem;
    padding-top: 0.1rem;
  }

  .index-copy {
    display: grid;
    gap: 0.2rem;
  }

  .index-copy strong {
    font-size: 0.85rem;
    font-weight: 600;
    line-height: 1.3;
  }

  .index-copy small {
    color: color-mix(in srgb, currentColor 68%, transparent);
    font-size: 0.7rem;
  }

  .index-footnote {
    border-top: 1px solid color-mix(in srgb, currentColor 18%, transparent);
    color: color-mix(in srgb, currentColor 68%, transparent);
    font-size: 0.72rem;
    line-height: 1.55;
    margin: 0;
    padding: 1rem 0.25rem 0;
  }

  .history-content {
    display: grid;
    gap: 1rem;
    min-width: 0;
  }

  .decision-theme {
    background: color-mix(in srgb, #{$link} 2%, transparent);
    scroll-margin-top: 1.25rem;
  }

  .decision-theme.active-theme,
  .decision-case.active-case {
    border-color: color-mix(in srgb, #{$link} 58%, transparent);
    box-shadow: 0 0.5rem 1.5rem color-mix(in srgb, currentColor 10%, transparent);
  }

  .theme-header {
    align-items: start;
    border-bottom: 1px solid color-mix(in srgb, currentColor 14%, transparent);
    display: grid;
    gap: 1.5rem;
    grid-template-columns: minmax(0, 1.15fr) minmax(16rem, 0.85fr);
    padding: 1.25rem 1.35rem 1rem;
  }

  .theme-summary {
    margin: 1rem 1.35rem 0;
  }

  .branch-stack {
    display: grid;
    gap: 0;
    margin-top: 1rem;
  }

  .decision-branch {
    border-top: 1px solid color-mix(in srgb, currentColor 14%, transparent);
    display: grid;
    gap: 1rem;
    grid-template-columns: 2.2rem minmax(0, 1fr);
    margin: 0 1.35rem;
    padding: 1.15rem 0;
    scroll-margin-top: 1.25rem;
  }

  .branch-marker,
  .case-number {
    color: $link;
    font-family: $family-monospace;
    font-size: 0.78rem;
    font-weight: 700;
    padding-top: 0.15rem;
  }

  .branch-header {
    align-items: start;
    display: grid;
    gap: 1rem;
    grid-template-columns: minmax(0, 1fr) minmax(14rem, 0.8fr);
  }

  .branch-header h3 {
    font-size: 1.08rem;
    font-weight: 700;
    line-height: 1.25;
    margin: 0.25rem 0 0;
  }

  .branch-header > strong {
    color: color-mix(in srgb, currentColor 88%, transparent);
    font-family: $family-monospace;
    font-size: 0.95rem;
    line-height: 1.35;
    text-align: right;
  }

  .branch-grid {
    display: grid;
    gap: 1.5rem;
    grid-template-columns: minmax(0, 0.8fr) minmax(20rem, 1.2fr);
    margin-top: 1rem;
  }

  .branch-narrative,
  .branch-evidence {
    min-width: 0;
  }

  .decision-copy {
    color: inherit;
    font-size: 0.92rem;
    font-weight: 600;
    line-height: 1.5;
    margin: 0.35rem 0 0;
  }

  .outcome-block {
    border-top: 1px solid color-mix(in srgb, currentColor 14%, transparent);
    margin-top: 1.2rem;
    padding-top: 0.85rem;
  }

  .outcome-block p,
  .narrative-block p {
    color: color-mix(in srgb, currentColor 78%, transparent);
    font-size: 0.84rem;
    line-height: 1.55;
    margin: 0.35rem 0 0;
  }

  .branch-metrics {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    margin-top: 0;
  }

  .metric-card.compact {
    padding: 0.6rem 0.65rem 0.7rem;
  }

  .metric-card.compact strong {
    font-size: 0.92rem;
  }

  .timeline-block {
    margin-top: 1.05rem;
  }

  .timeline-table-wrap {
    overflow-x: auto;
  }

  .timeline-table {
    border-collapse: collapse;
    font-size: 0.72rem;
    margin-top: 0.45rem;
    min-width: 32rem;
    width: 100%;
  }

  .timeline-table th,
  .timeline-table td {
    border-bottom: 1px solid color-mix(in srgb, currentColor 12%, transparent);
    padding: 0.45rem 0.4rem;
    text-align: left;
    vertical-align: top;
  }

  .timeline-table thead th {
    color: color-mix(in srgb, currentColor 56%, transparent);
    font-size: 0.63rem;
    font-weight: 700;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .timeline-table tbody th {
    font-weight: 650;
    white-space: nowrap;
  }

  .timeline-table tbody td:nth-child(2) {
    font-family: $family-monospace;
    font-weight: 650;
    white-space: nowrap;
  }

  .timeline-table tbody td:last-child {
    color: color-mix(in srgb, currentColor 70%, transparent);
  }

  .evidence-footer,
  .theme-footer,
  .case-footer {
    background: color-mix(in srgb, currentColor 4%, transparent);
    border-top: 1px solid color-mix(in srgb, currentColor 14%, transparent);
    display: grid;
    gap: 1.25rem;
    grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.2fr);
    padding: 0.9rem 1.35rem 1rem;
  }

  .evidence-footer {
    margin-left: -3.2rem;
    margin-top: 1rem;
  }

  .theme-footer {
    margin-top: 0.25rem;
  }

  .source-list,
  .caveat-list {
    color: color-mix(in srgb, currentColor 78%, transparent);
    font-size: 0.72rem;
    line-height: 1.45;
    list-style: none;
    margin: 0.4rem 0 0;
    padding: 0;
  }

  .source-list li,
  .caveat-list li {
    margin-bottom: 0.25rem;
  }

  .source-mark {
    color: $link;
    font-family: $family-monospace;
    margin-right: 0.35rem;
  }

  .caveat-list li {
    padding-left: 0.9rem;
    position: relative;
  }

  .caveat-list li::before {
    color: color-mix(in srgb, currentColor 68%, transparent);
    content: "•";
    left: 0.15rem;
    position: absolute;
  }

  .secondary-heading {
    align-items: baseline;
    display: flex;
    gap: 0.8rem;
    margin: 1.25rem 0 0.15rem;
  }

  .secondary-heading p {
    color: color-mix(in srgb, currentColor 65%, transparent);
    font-size: 0.78rem;
    margin: 0;
  }

  .decision-case {
    background: transparent;
    scroll-margin-top: 1.25rem;
  }

  .case-header {
    align-items: flex-start;
    border-bottom: 1px solid color-mix(in srgb, currentColor 14%, transparent);
    display: grid;
    gap: 0.85rem;
    grid-template-columns: 2.25rem minmax(0, 1fr) auto;
    padding: 1.15rem 1.25rem 1rem;
  }

  .case-title h2 {
    font-size: 1.1rem;
    font-weight: 650;
    line-height: 1.25;
    margin: 0.2rem 0 0;
  }

  .status-tag {
    background: color-mix(in srgb, currentColor 7%, transparent);
    border: 1px solid color-mix(in srgb, currentColor 18%, transparent);
    border-radius: 999px;
    color: color-mix(in srgb, currentColor 78%, transparent);
    font-size: 0.7rem;
    line-height: 1.2;
    padding: 0.4rem 0.65rem;
    text-align: right;
  }

  .case-body {
    display: grid;
    gap: 1.5rem;
    grid-template-columns: minmax(0, 1fr) minmax(18rem, 0.85fr);
    padding: 1.2rem 1.25rem 1.35rem;
  }

  .narrative-column,
  .evidence-column {
    display: grid;
    gap: 1rem;
  }

  .narrative-block {
    max-width: 62ch;
  }

  .amount-panel {
    background: color-mix(in srgb, currentColor 5%, transparent);
    border-radius: $radius;
    display: grid;
    gap: 0.2rem;
    padding: 0.8rem 0.9rem;
  }

  .amount-panel strong {
    font-family: $family-monospace;
    font-size: 1.25rem;
    font-weight: 650;
    letter-spacing: -0.03em;
  }

  .amount-panel > span:last-child {
    color: color-mix(in srgb, currentColor 68%, transparent);
    font-size: 0.75rem;
  }

  .attribution-grid {
    display: grid;
    gap: 0.6rem;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .evidence-field {
    border-top: 2px solid color-mix(in srgb, #{$link} 58%, transparent);
    min-width: 0;
    padding-top: 0.45rem;
  }

  .evidence-field.field-unknown {
    border-top-color: color-mix(in srgb, currentColor 18%, transparent);
  }

  .evidence-field-heading {
    align-items: baseline;
    display: flex;
    gap: 0.45rem;
    justify-content: space-between;
  }

  .field-status {
    color: color-mix(in srgb, currentColor 68%, transparent);
    font-size: 0.65rem;
  }

  .evidence-field p {
    font-size: 0.78rem;
    font-weight: 600;
    line-height: 1.35;
    margin: 0.35rem 0 0.25rem;
  }

  .evidence-field small {
    color: color-mix(in srgb, currentColor 68%, transparent);
    display: block;
    font-size: 0.68rem;
    line-height: 1.45;
  }

  .case-footer {
    padding-left: 1.25rem;
    padding-right: 1.25rem;
  }

  .decision-history-loading,
  .decision-history-empty {
    background: transparent;
    border: 1px solid color-mix(in srgb, currentColor 18%, transparent);
    border-radius: $radius;
    margin: 3rem auto;
    max-width: 42rem;
    padding: 2rem;
    text-align: center;
  }

  .decision-history-loading {
    align-items: center;
    color: color-mix(in srgb, currentColor 78%, transparent);
    display: flex;
    gap: 0.75rem;
    justify-content: center;
  }

  .loader {
    border: 2px solid color-mix(in srgb, currentColor 18%, transparent);
    border-radius: 50%;
    border-top-color: $link;
    display: inline-block;
    height: 1rem;
    width: 1rem;
  }

  .decision-history-empty p:not(.title) {
    color: color-mix(in srgb, currentColor 78%, transparent);
    line-height: 1.55;
    margin: 0.7rem auto;
    max-width: 34rem;
  }

  @media screen and (max-width: 1000px) {
    .theme-header,
    .branch-grid {
      grid-template-columns: 1fr;
    }

    .theme-result {
      border-left: 0;
      border-top: 1px solid color-mix(in srgb, #{$link} 45%, transparent);
      padding-left: 0;
      padding-top: 0.75rem;
    }

    .branch-header {
      grid-template-columns: 1fr;
    }

    .branch-header > strong {
      text-align: left;
    }

    .theme-metrics {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media screen and (max-width: 768px) {
    .history-layout {
      display: block;
    }

    .history-index {
      background: color-mix(in srgb, currentColor 5%, transparent);
      margin-bottom: 1.25rem;
      overflow-x: auto;
      position: static;
    }

    .index-heading,
    .index-footnote {
      display: none;
    }

    .history-index ol {
      display: flex;
      gap: 0.45rem;
      padding: 0;
    }

    .history-index li {
      min-width: 10.5rem;
    }

    .history-index a {
      border: 1px solid color-mix(in srgb, currentColor 18%, transparent);
      height: 100%;
    }

    .case-body,
    .case-footer,
    .evidence-footer,
    .theme-footer {
      grid-template-columns: 1fr;
    }

    .evidence-footer {
      margin-left: 0;
    }

    .case-header {
      grid-template-columns: 2rem minmax(0, 1fr);
    }

    .status-tag {
      grid-column: 2;
      justify-self: start;
    }

    .branch-metrics,
    .attribution-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media screen and (max-width: 480px) {
    .decision-history-page {
      padding-left: 0.8rem;
      padding-right: 0.8rem;
    }

    .theme-header,
    .theme-footer,
    .case-body,
    .case-footer,
    .case-header {
      padding-left: 0.9rem;
      padding-right: 0.9rem;
    }

    .decision-branch {
      grid-template-columns: 1.8rem minmax(0, 1fr);
      margin-left: 0.9rem;
      margin-right: 0.9rem;
    }

    .theme-summary {
      margin-left: 0.9rem;
      margin-right: 0.9rem;
    }

    .theme-metrics,
    .branch-metrics,
    .attribution-grid {
      grid-template-columns: 1fr;
    }

    .secondary-heading {
      align-items: flex-start;
      display: grid;
    }
  }
</style>
