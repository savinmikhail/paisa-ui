<script lang="ts">
  import { onMount } from "svelte";
  import { loadMoneyFlow } from "$lib/money_flow_loader";
  import {
    buildMoneyFlowView,
    defaultMoneyFlowFilters,
    MONEY_FLOW_CATEGORIES,
    renderMoneyFlowSankey,
    renderMoneyFlowWaterfall,
    type MoneyFlowFilters,
    type MoneyFlowSource
  } from "$lib/money_flow";
  import { MONEY_FLOW_SYNTHETIC_FIXTURE } from "$lib/fixtures/money-flow";
  import { formatCurrency } from "$lib/utils";

  let source: MoneyFlowSource | null = null;
  let activeTab: "sankey" | "waterfall" = "sankey";
  let filters: MoneyFlowFilters;
  let draftFrom = "";
  let draftTo = "";
  let draftCurrency: MoneyFlowFilters["currency"] = "ALL";
  let loading = true;
  let errorMessage = "";
  let sankeySvg: SVGSVGElement;
  let waterfallSvg: SVGSVGElement;

  $: view = source && filters ? buildMoneyFlowView(source, filters) : null;
  $: if (view && sankeySvg && activeTab === "sankey") renderMoneyFlowSankey(sankeySvg, view.sankey);
  $: if (view && waterfallSvg && activeTab === "waterfall") renderMoneyFlowWaterfall(waterfallSvg, view.waterfall);

  onMount(async () => {
    const fixtureMode = new URLSearchParams(window.location.search).get("fixture") === "money-flow";
    try {
      source = fixtureMode ? MONEY_FLOW_SYNTHETIC_FIXTURE : await loadMoneyFlow("private", (route) => fetch(route).then((response) => response.json()));
      filters = defaultMoneyFlowFilters(source);
      draftFrom = filters.from;
      draftTo = filters.to;
      draftCurrency = filters.currency;
    } catch (error) {
      errorMessage = error instanceof Error ? error.message : "Money Flow API could not be loaded";
    } finally {
      loading = false;
    }
  });

  function applyFilters() {
    if (!filters || !draftFrom || !draftTo || draftFrom > draftTo) return;
    filters = { ...filters, from: draftFrom, to: draftTo, currency: draftCurrency };
  }

  function resetFilters() {
    if (!source) return;
    filters = defaultMoneyFlowFilters(source);
    draftFrom = filters.from;
    draftTo = filters.to;
    draftCurrency = filters.currency;
  }

  function toggleCategory(category: MoneyFlowFilters["categories"][number]) {
    filters = {
      ...filters,
      categories: filters.categories.includes(category)
        ? filters.categories.filter((item) => item !== category)
        : [...filters.categories, category]
    };
  }
</script>

<svelte:head><title>Money Flow · Paisa</title></svelte:head>

<section class="section money-flow-screen" data-testid="money-flow-screen">
  <div class="container is-fluid">
    <div class="money-flow-heading">
      <div>
        <h1 class="title is-3 mb-2">Money Flow</h1>
        <p class="subtitle is-6 mb-0">Follow recognized income into household uses, then explain the change in Net Worth.</p>
      </div>
      <div class="has-text-right is-size-7 has-text-grey" data-testid="money-flow-contract">
        <div>money-flow.v1 · read-only</div>
        <div>{source?.baseCurrency || "—"} · {source?.source || "loading"}</div>
      </div>
    </div>

    <div class="notification is-info is-light invertable money-flow-notice">
      <strong>Reporting view.</strong> Internal Transfers never become Income or Expenses; unresolved ones are opt-in for investigation.
    </div>

    <div class="box money-flow-toolbar" data-testid="money-flow-filters">
      <div class="toolbar-copy"><strong>Shared period and filters</strong><span class="is-size-7 has-text-grey">Both tabs use the same selected inputs.</span></div>
      <form class="money-flow-form" on:submit|preventDefault={applyFilters}>
        <label class="label is-size-7 mb-1" for="money-flow-from">From</label>
        <label class="label is-size-7 mb-1" for="money-flow-to">To</label>
        <label class="label is-size-7 mb-1" for="money-flow-currency">Currency</label>
        <span></span>
        <input id="money-flow-from" class="input is-small" type="date" bind:value={draftFrom} />
        <input id="money-flow-to" class="input is-small" type="date" bind:value={draftTo} />
        <div class="select is-small"><select id="money-flow-currency" bind:value={draftCurrency}><option value="ALL">All currencies</option><option value={source?.baseCurrency || "RUB"}>{source?.baseCurrency || "RUB"}</option></select></div>
        <div class="money-flow-actions"><button class="button is-small is-primary" type="submit">Apply</button><button class="button is-small" type="button" on:click={resetFilters}>Reset</button></div>
      </form>
      {#if filters}
        <div class="category-filters" aria-label="Money Flow categories">
          {#each MONEY_FLOW_CATEGORIES as category}
            <label class="checkbox is-size-7"><input type="checkbox" checked={filters.categories.includes(category)} on:change={() => toggleCategory(category)} /> {category}</label>
          {/each}
        </div>
        <label class="checkbox is-size-7 unresolved-toggle"><input data-testid="unresolved-transfers-toggle" type="checkbox" bind:checked={filters.showUnresolvedTransfers} /> Show unresolved Internal Transfers</label>
      {/if}
    </div>

    {#if filters}<div class="range-line is-size-7 has-text-grey" data-testid="money-flow-active-range">{filters.from} → {filters.to} · {filters.currency === "ALL" ? "all currencies" : filters.currency}{filters.showUnresolvedTransfers ? " · unresolved transfers visible" : ""}</div>{/if}

    {#if errorMessage}
      <div class="message is-danger" data-testid="money-flow-error"><div class="message-header">Money Flow is unavailable</div><div class="message-body">{errorMessage}</div></div>
    {:else if loading || !source || !view}
      <div class="box has-text-centered has-text-grey" data-testid="money-flow-loading">Loading read-only Money Flow…</div>
    {:else}
      <div class="tabs is-boxed money-flow-tabs" role="tablist" aria-label="Money Flow views">
        <ul><li class:is-active={activeTab === "sankey"}><button role="tab" aria-selected={activeTab === "sankey"} data-testid="money-flow-tab-sankey" on:click={() => (activeTab = "sankey")}>Sankey</button></li><li class:is-active={activeTab === "waterfall"}><button role="tab" aria-selected={activeTab === "waterfall"} data-testid="money-flow-tab-waterfall" on:click={() => (activeTab = "waterfall")}>Waterfall</button></li></ul>
      </div>
      {#if activeTab === "sankey"}
        <div class="columns is-variable is-5" data-testid="money-flow-sankey-panel">
          <div class="column is-8"><div class="box money-flow-panel"><div class="panel-heading-row"><div><h2 class="title is-5 mb-1">Where recognized income went</h2><p class="is-size-7 has-text-grey">Personal Expense, business, tax, reserve, and Capital Placement remain separate.</p></div><span class="tag is-link is-light">{formatCurrency(view.sankey.incomeTotal)} income</span></div><div class="chart-frame"><svg bind:this={sankeySvg} data-testid="money-flow-sankey-chart" aria-label="Income Sankey diagram"></svg></div></div></div>
          <div class="column is-4"><div class="box money-flow-panel"><h2 class="title is-6 mb-3">Destinations</h2><div class="flow-legend">{#each view.sankey.nodes.filter((node) => node.kind === "destination") as node}<div class="flow-legend-row"><span>{node.label}</span><strong>{formatCurrency(node.value)}</strong></div>{/each}</div>{#if filters.showUnresolvedTransfers}<div class="transfer-callout" data-testid="unresolved-transfer-summary"><span>Unresolved Internal Transfers</span><strong>{formatCurrency(view.sankey.transferTotal)}</strong><small>Shown outside Income and Expenses.</small></div>{/if}</div></div>
        </div>
      {:else}
        <div class="columns is-variable is-5" data-testid="money-flow-waterfall-panel">
          <div class="column is-8"><div class="box money-flow-panel"><div class="panel-heading-row"><div><h2 class="title is-5 mb-1">Why Net Worth changed</h2><p class="is-size-7 has-text-grey">Same six-part decomposition semantics used by Net Worth reconciliation.</p></div><span class="tag is-success is-light">{formatCurrency(view.waterfall.closing - view.waterfall.opening)} change</span></div><div class="chart-frame"><svg bind:this={waterfallSvg} data-testid="money-flow-waterfall-chart" aria-label="Net Worth waterfall"></svg></div></div></div>
          <div class="column is-4"><div class="box money-flow-panel waterfall-summary"><div><span>Opening Net Worth</span><strong>{formatCurrency(view.waterfall.opening)}</strong></div>{#each view.waterfall.steps as step}<div><span>{step.label}</span><strong class:has-text-danger={step.amount < 0}>{formatCurrency(step.amount)}</strong></div>{/each}<div class="closing-row"><span>Closing Net Worth</span><strong>{formatCurrency(view.waterfall.closing)}</strong></div></div></div>
        </div>
      {/if}
    {/if}
  </div>
</section>

<style lang="scss">
  .money-flow-heading, .panel-heading-row { display: flex; justify-content: space-between; align-items: flex-start; gap: 1rem; }
  .money-flow-notice { border: 1px solid var(--bulma-info); }
  .money-flow-toolbar { display: grid; gap: 0.75rem; margin-bottom: 0.5rem; }
  .toolbar-copy { display: flex; gap: 0.75rem; align-items: baseline; }
  .money-flow-form { display: grid; grid-template-columns: repeat(3, minmax(9rem, 1fr)) auto; gap: 0.5rem; align-items: end; }
  .money-flow-actions { display: flex; gap: 0.5rem; }
  .category-filters { display: flex; flex-wrap: wrap; gap: 0.8rem; }
  .unresolved-toggle { justify-self: start; }
  .range-line { margin: 0 0 1rem; }
  .money-flow-tabs button { border: 0; background: transparent; padding: 0.7rem 1rem; color: inherit; cursor: pointer; }
  .money-flow-panel { height: 100%; }
  .chart-frame { overflow-x: auto; min-height: 410px; padding-top: 1rem; }
  .flow-legend { display: grid; gap: 0.75rem; }
  .flow-legend-row, .waterfall-summary > div { display: flex; justify-content: space-between; gap: 1rem; border-bottom: 1px solid var(--bulma-border); padding-bottom: 0.6rem; }
  .waterfall-summary span { color: var(--bulma-text-weak); }
  .closing-row { border-top: 2px solid var(--bulma-border); border-bottom: 0 !important; padding-top: 0.8rem; font-weight: 700; }
  .transfer-callout { display: grid; gap: 0.25rem; margin-top: 1.25rem; padding: 0.85rem; border: 1px dashed var(--bulma-warning); border-radius: 0.6rem; }
  .transfer-callout small { color: var(--bulma-text-weak); }
  :global(.money-flow-svg-label) { fill: currentColor; font-size: 12px; font-variant-numeric: tabular-nums; }
  @media screen and (max-width: 900px) { .money-flow-form { grid-template-columns: 1fr 1fr; } .money-flow-form span { display: none; } }
  @media screen and (max-width: 768px) { .money-flow-heading, .panel-heading-row, .toolbar-copy { flex-direction: column; } .money-flow-form { grid-template-columns: 1fr; } .money-flow-actions { grid-column: 1; } }
</style>
