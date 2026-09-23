<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import {
    privateDecisionHistoryAdapter,
    syntheticDecisionHistory,
    type DecisionHistoryCase,
    type DecisionHistoryField
  } from "$lib/decision_history";

  type LoadState = "loading" | "ready" | "unavailable";

  let loadState: LoadState = "loading";
  let decisionCases: DecisionHistoryCase[] = [];
  let isSynthetic = false;
  let activeCaseId = "";

  onMount(async () => {
    isSynthetic = $page.url.searchParams.get("fixture") === "synthetic";

    if (isSynthetic) {
      decisionCases = syntheticDecisionHistory.cases;
      activeCaseId = decisionCases[0]?.id ?? "";
      loadState = "ready";
      return;
    }

    try {
      const result = await privateDecisionHistoryAdapter.load(fetch);
      if (result.kind === "ready") {
        decisionCases = result.payload.cases;
        activeCaseId = decisionCases[0]?.id ?? "";
        loadState = "ready";
      } else {
        loadState = "unavailable";
      }
    } catch (_error) {
      loadState = "unavailable";
    }
  });

  function selectCase(id: string) {
    activeCaseId = id;
  }

  function fieldLabel(status: DecisionHistoryField["status"]) {
    if (status === "known") return "Known";
    if (status === "partial") return "Partial";
    return "Unknown";
  }

  function attributionFields(item: DecisionHistoryCase): [string, DecisionHistoryField][] {
    return [
      ["FX", item.attribution.fx],
      ["Fees", item.attribution.fees],
      ["Taxes", item.attribution.taxes],
      ["RUB counterfactual", item.attribution.rubCounterfactual]
    ];
  }
</script>

<svelte:head>
  <title>Decision history · Paisa</title>
  <meta
    name="description"
    content="A manually curated, evidence-aware history of financial decisions."
  />
</svelte:head>

<section class="section decision-history-page">
  <div class="container is-fluid">
    {#if loadState === "loading"}
      <div class="decision-history-loading" data-testid="decision-history-loading">
        <span class="loader" aria-hidden="true"></span>
        <span>Loading decision history…</span>
      </div>
    {:else if loadState === "unavailable"}
      <div class="decision-history-empty" data-testid="decision-history-unavailable">
        <p class="title is-4">Decision history is not connected</p>
        <p>
          The private facts adapter is not available in this public UI checkout. No personal Ledger
          data is bundled here.
        </p>
        <p class="has-text-grey">
          For screenshot review, open this route with <code>?fixture=synthetic</code> to load the deterministic,
          public-safe fixture.
        </p>
      </div>
    {:else}
      <div class="decision-history-heading">
        <div>
          <h1 class="title is-3 mb-1">Decision history</h1>
          <p class="subtitle is-6 mb-0">
            Four manually curated decisions, kept separate from automatic performance reporting.
          </p>
        </div>
      </div>

      <div class="interpretation-note" data-testid="decision-history-ready">
        <span class="icon is-small" aria-hidden="true"><i class="fas fa-compass"></i></span>
        <div>
          <strong>Interpretation boundary</strong>
          <span>
            This view records what was decided, what is evidenced, and what remains open. It never
            auto-labels a decision as success or failure.
          </span>
        </div>
      </div>

      <div class="history-layout">
        <aside class="history-index" aria-label="Decision history index">
          <div class="index-heading">
            <span class="eyebrow">Case index</span>
            <span class="index-count">{decisionCases.length} cases</span>
          </div>
          <ol>
            {#each decisionCases as item, index}
              <li>
                <a
                  href={`#${item.id}`}
                  class:active={activeCaseId === item.id}
                  aria-current={activeCaseId === item.id ? "true" : undefined}
                  on:click={() => selectCase(item.id)}
                >
                  <span class="index-number">{String(index + 1).padStart(2, "0")}</span>
                  <span class="index-copy">
                    <strong>{item.title}</strong>
                    <small>{item.period}</small>
                  </span>
                </a>
              </li>
            {/each}
          </ol>
          <p class="index-footnote">
            Curated records are not a complete transaction timeline. Follow source refs for the
            evidence boundary.
          </p>
        </aside>

        <div class="case-stack">
          {#each decisionCases as item, index}
            <article
              id={item.id}
              class="decision-case"
              class:active-case={activeCaseId === item.id}
              data-testid={`decision-case-${item.id}`}
            >
              <header class="case-header">
                <div class="case-number" aria-hidden="true">
                  {String(index + 1).padStart(2, "0")}
                </div>
                <div class="case-title">
                  <span class="case-period">{item.period}</span>
                  <h2>{item.title}</h2>
                </div>
                <span class="status-tag">{item.currentStatus}</span>
              </header>

              <div class="case-body">
                <div class="narrative-column">
                  <div class="narrative-block">
                    <span class="field-label">Decision</span>
                    <p class="decision-copy">{item.decision}</p>
                  </div>
                  <div class="narrative-block">
                    <span class="field-label">Context</span>
                    <p>{item.context}</p>
                  </div>
                  <div class="narrative-block outcome-block">
                    <span class="field-label">Outcome</span>
                    <p>{item.outcome}</p>
                  </div>
                </div>

                <div class="evidence-column">
                  <div class="amount-panel">
                    <span class="field-label">Amount</span>
                    <strong>{item.amount.label}</strong>
                    <span>{item.amount.currency}</span>
                  </div>

                  <div class="attribution-grid" aria-label="Attribution fields">
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
                  <span class="field-label">Source refs</span>
                  <ul class="source-list">
                    {#each item.sourceRefs as source}
                      <li><span class="source-mark" aria-hidden="true">↗</span>{source}</li>
                    {/each}
                  </ul>
                </div>
                <div class="footer-section caveat-section">
                  <span class="field-label">Evidence caveats</span>
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
    padding-top: 2.5rem;
    padding-bottom: 4rem;
  }

  .decision-history-heading {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 2rem;
    margin-bottom: 1.25rem;
  }

  .decision-history-heading .subtitle {
    color: color-mix(in srgb, currentColor 68%, transparent);
    max-width: 48rem;
  }

  .interpretation-note {
    align-items: flex-start;
    background: color-mix(in srgb, #{$link} 8%, transparent);
    border: 1px solid color-mix(in srgb, #{$link} 24%, transparent);
    border-radius: $radius;
    color: inherit;
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
    color: $grey;
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .index-count {
    color: $grey-light;
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
    color: $grey-dark;
    display: flex;
    gap: 0.75rem;
    margin: 0.25rem 0;
    padding: 0.7rem 0.55rem;
    transition:
      background-color 150ms ease,
      color 150ms ease;
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
    color: $grey-light;
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
    color: $grey;
    font-size: 0.7rem;
  }

  .index-footnote {
    border-top: 1px solid $grey-lighter;
    color: $grey;
    font-size: 0.72rem;
    line-height: 1.55;
    margin: 0;
    padding: 1rem 0.25rem 0;
  }

  .case-stack {
    display: grid;
    gap: 1rem;
  }

  .decision-case {
    background: transparent;
    border: 1px solid color-mix(in srgb, currentColor 18%, transparent);
    border-radius: $radius;
    scroll-margin-top: 1.25rem;
  }

  .decision-case.active-case {
    border-color: rgba($link, 0.5);
    box-shadow: 0 0.5rem 1.5rem color-mix(in srgb, currentColor 10%, transparent);
  }

  .case-header {
    align-items: flex-start;
    border-bottom: 1px solid color-mix(in srgb, currentColor 14%, transparent);
    display: grid;
    gap: 0.85rem;
    grid-template-columns: 2.25rem minmax(0, 1fr) auto;
    padding: 1.15rem 1.25rem 1rem;
  }

  .case-number {
    color: $link;
    font-family: $family-monospace;
    font-size: 0.78rem;
    font-weight: 700;
    padding-top: 0.15rem;
  }

  .case-title h2 {
    color: $black-bis;
    font-size: 1.15rem;
    font-weight: 650;
    line-height: 1.25;
    margin: 0.2rem 0 0;
  }

  .status-tag {
    background: color-mix(in srgb, currentColor 7%, transparent);
    border: 1px solid color-mix(in srgb, currentColor 18%, transparent);
    border-radius: 999px;
    color: $grey-dark;
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

  .narrative-block p {
    color: $grey-dark;
    font-size: 0.88rem;
    line-height: 1.55;
    margin: 0.35rem 0 0;
  }

  .narrative-block .decision-copy {
    color: $black-bis;
    font-size: 1rem;
    font-weight: 600;
    line-height: 1.45;
  }

  .outcome-block {
    border-top: 1px solid color-mix(in srgb, currentColor 14%, transparent);
    padding-top: 0.9rem;
  }

  .amount-panel {
    background: color-mix(in srgb, currentColor 5%, transparent);
    border-radius: $radius;
    display: grid;
    gap: 0.2rem;
    padding: 0.8rem 0.9rem;
  }

  .amount-panel strong {
    color: $black-bis;
    font-family: $family-monospace;
    font-size: 1.35rem;
    font-weight: 650;
    letter-spacing: -0.03em;
  }

  .amount-panel > span:last-child {
    color: $grey;
    font-size: 0.75rem;
  }

  .attribution-grid {
    display: grid;
    gap: 0.6rem;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .evidence-field {
    border-top: 2px solid rgba($link, 0.5);
    min-width: 0;
    padding-top: 0.45rem;
  }

  .evidence-field.field-unknown {
    border-top-color: $grey-lighter;
  }

  .evidence-field-heading {
    align-items: baseline;
    display: flex;
    gap: 0.45rem;
    justify-content: space-between;
  }

  .field-status {
    color: $grey;
    font-size: 0.65rem;
  }

  .evidence-field p {
    color: $black-bis;
    font-size: 0.78rem;
    font-weight: 600;
    line-height: 1.35;
    margin: 0.35rem 0 0.25rem;
  }

  .evidence-field small {
    color: $grey;
    display: block;
    font-size: 0.68rem;
    line-height: 1.45;
  }

  .case-footer {
    background: $white-bis;
    border-top: 1px solid $grey-lightest;
    display: grid;
    gap: 1.25rem;
    grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.2fr);
    padding: 1rem 1.25rem 1.1rem;
  }

  .source-list,
  .caveat-list {
    color: $grey-dark;
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
    color: $grey;
    content: "•";
    left: 0.15rem;
    position: absolute;
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
    color: $grey-dark;
    display: flex;
    gap: 0.75rem;
    justify-content: center;
  }

  .loader {
    border: 2px solid $grey-lighter;
    border-radius: 50%;
    border-top-color: $link;
    display: inline-block;
    height: 1rem;
    width: 1rem;
  }

  .decision-history-empty p:not(.title) {
    color: $grey-dark;
    line-height: 1.55;
    margin: 0.7rem auto;
    max-width: 34rem;
  }

  @media screen and (max-width: 768px) {
    .decision-history-heading {
      display: block;
    }

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
    .case-footer {
      grid-template-columns: 1fr;
    }

    .case-header {
      grid-template-columns: 2rem minmax(0, 1fr);
    }

    .status-tag {
      grid-column: 2;
      justify-self: start;
    }
  }

  @media screen and (max-width: 480px) {
    .decision-history-page {
      padding-left: 0.8rem;
      padding-right: 0.8rem;
    }

    .case-body,
    .case-footer,
    .case-header {
      padding-left: 0.9rem;
      padding-right: 0.9rem;
    }

    .attribution-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
