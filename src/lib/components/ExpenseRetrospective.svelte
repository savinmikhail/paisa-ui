<script lang="ts">
  import { baselineBarHeight, baselineChartMax, categorySignalLabel, commitmentCadence } from "$lib/expense_baseline";
  import { formatCurrency, type ExpenseRetrospective } from "$lib/utils";

  export let data: ExpenseRetrospective;

  $: chartMax = baselineChartMax(data);
  $: latestMonth = data.months[data.months.length - 1];

  function height(value: number) {
    return `${baselineBarHeight(value, chartMax)}%`;
  }
</script>

<section
  class="box expense-retrospective"
  data-testid="expense-retrospective"
  aria-labelledby="expense-retrospective-title"
>
  <div class="retrospective-heading">
    <div>
      <h2 id="expense-retrospective-title" class="title is-5 mb-1">Personal expense retrospective</h2>
      <p class="has-text-grey">
        Six complete months · Personal Expense excludes business, tax, and Capital Placement.
      </p>
    </div>
    <span class="tag is-light">{data.start_date.format("MMM YYYY")} — {data.end_date.subtract(1, "day").format("MMM YYYY")}</span>
  </div>

  <div class="retrospective-overview">
    <div class="baseline-panel">
      <div class="panel-heading">
        <span>Monthly baseline</span>
        <span class="legend-inline"><i class="legend-swatch personal" /> Personal Expense <i class="legend-swatch regular" /> Regular burn</span>
      </div>
      <div class="baseline-chart" aria-label="Six-month personal expense baseline">
        <div class="median-rule" style={`bottom: ${height(data.personal_expense_median)}`}>
          <span>6m median {formatCurrency(data.personal_expense_median)}</span>
        </div>
        {#each data.months as month}
          <div class="month-column">
            <div class="bar-pair">
              <div class="bar personal" style={`height: ${height(month.personal_expense)}`} title={`Personal Expense ${formatCurrency(month.personal_expense)}`} />
              <div class="bar regular" style={`height: ${height(month.regular_personal_burn)}`} title={`Regular burn ${formatCurrency(month.regular_personal_burn)}`} />
            </div>
            <span class="month-label">{month.date.format("MMM")}</span>
          </div>
        {/each}
      </div>
      <div class="chart-footnote">
        Latest month: <strong>{formatCurrency(latestMonth.personal_expense)}</strong> personal ·
        <strong>{formatCurrency(latestMonth.regular_personal_burn)}</strong> regular
      </div>
    </div>

    <div class="metric-stack" aria-label="Expense baseline metrics">
      <div class="metric-row">
        <span>Personal Expense median</span>
        <strong>{formatCurrency(data.personal_expense_median)}</strong>
        <small>includes confirmed one-offs</small>
      </div>
      <div class="metric-row accent">
        <span>Regular Personal Burn</span>
        <strong>{formatCurrency(data.regular_personal_burn_median)}</strong>
        <small>one-offs excluded</small>
      </div>
    </div>
  </div>

  <div class="signal-grid">
    <div class="signal-panel">
      <div class="panel-heading">
        <span>Growing faster than baseline</span>
        <span class="has-text-grey is-size-7">latest month vs six-month median</span>
      </div>
      {#if data.category_signals.length === 0}
        <p class="has-text-grey is-size-7">No categories are above their baseline.</p>
      {:else}
        {#each data.category_signals as signal}
          <div class="signal-row">
            <span class="signal-name">{signal.category}</span>
            <span class="signal-values">{formatCurrency(signal.current)} <span class="has-text-grey">vs {formatCurrency(signal.baseline_median)}</span></span>
            <span class="signal-delta">{categorySignalLabel(signal)}</span>
          </div>
        {/each}
      {/if}
    </div>

    <div class="signal-panel">
      <div class="panel-heading">
        <span>Recurring commitments</span>
        <a class="secondary-link is-size-7" href="/cash_flow/recurring">View all</a>
      </div>
      {#if data.recurring_commitments.length === 0}
        <p class="has-text-grey is-size-7">No tagged recurring personal expenses.</p>
      {:else}
        {#each data.recurring_commitments as commitment}
          <div class="signal-row">
            <span class="signal-name">{commitment.category || commitment.key}</span>
            <span class="signal-values">{formatCurrency(commitment.amount)}</span>
            <span class="has-text-grey is-size-7">{commitmentCadence(commitment)}</span>
          </div>
        {/each}
      {/if}
    </div>
  </div>

  <div class="evidence-grid">
    <div>
      <div class="panel-heading"><span>Excluded from Personal Expense</span><span class="has-text-grey is-size-7">still accounted for elsewhere</span></div>
      <div class="evidence-list">
        {#each data.excluded as item}
          <div class="evidence-row"><span>{item.label}</span><strong>{formatCurrency(item.amount)}</strong><small>{item.count} {item.count === 1 ? "flow" : "flows"}</small></div>
        {/each}
      </div>
    </div>
    <div>
      <div class="panel-heading"><span>One-offs stay visible</span><span class="has-text-grey is-size-7">excluded from regular burn</span></div>
      <div class="evidence-list">
        {#each data.one_offs as oneOff}
          <div class="evidence-row"><span><strong>{oneOff.category}</strong> · {oneOff.posting.payee}</span><strong>{formatCurrency(oneOff.posting.amount)}</strong><small>confirmed one-off</small></div>
        {/each}
      </div>
    </div>
  </div>
</section>

<style>
  .expense-retrospective {
    padding: 1.25rem;
    border: 1px solid rgba(122, 132, 148, 0.22);
    background: rgba(255, 255, 255, 0.72);
  }

  .retrospective-heading,
  .panel-heading,
  .signal-row,
  .evidence-row {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 0.75rem;
  }

  .retrospective-heading {
    margin-bottom: 1rem;
  }

  .retrospective-overview {
    display: grid;
    grid-template-columns: minmax(0, 1.65fr) minmax(220px, 0.85fr);
    gap: 1rem;
  }

  .baseline-panel,
  .metric-stack,
  .signal-panel {
    min-width: 0;
  }

  .baseline-panel {
    position: relative;
    border-top: 1px solid rgba(122, 132, 148, 0.22);
    padding-top: 0.75rem;
  }

  .panel-heading {
    color: #475569;
    font-size: 0.8rem;
    font-weight: 600;
    margin-bottom: 0.65rem;
  }

  .legend-inline {
    color: #64748b;
    font-size: 0.7rem;
    font-weight: 400;
    white-space: nowrap;
  }

  .legend-swatch {
    display: inline-block;
    width: 0.55rem;
    height: 0.55rem;
    margin: 0 0.2rem 0 0.55rem;
    border-radius: 2px;
  }

  .legend-swatch.personal,
  .bar.personal {
    background: #e76f51;
  }

  .legend-swatch.regular,
  .bar.regular {
    background: #264653;
  }

  .baseline-chart {
    position: relative;
    display: flex;
    align-items: flex-end;
    gap: 0.65rem;
    height: 168px;
    padding: 0.75rem 0.5rem 1.35rem;
    border-bottom: 1px solid rgba(122, 132, 148, 0.22);
  }

  .median-rule {
    position: absolute;
    left: 0.5rem;
    right: 0.5rem;
    border-top: 1px dashed rgba(38, 70, 83, 0.42);
    pointer-events: none;
  }

  .median-rule span {
    position: absolute;
    top: -1.1rem;
    right: 0;
    color: #264653;
    font-size: 0.65rem;
  }

  .month-column {
    position: relative;
    z-index: 1;
    display: flex;
    flex: 1;
    align-items: center;
    flex-direction: column;
    height: 100%;
    justify-content: flex-end;
    min-width: 2rem;
  }

  .bar-pair {
    display: flex;
    align-items: flex-end;
    justify-content: center;
    gap: 3px;
    width: 100%;
    height: 100%;
  }

  .bar {
    width: min(1rem, 40%);
    min-height: 4px;
    border-radius: 3px 3px 0 0;
  }

  .month-label {
    position: absolute;
    bottom: -1.1rem;
    color: #64748b;
    font-size: 0.7rem;
  }

  .chart-footnote {
    margin-top: 0.7rem;
    color: #64748b;
    font-size: 0.75rem;
  }

  .metric-stack {
    display: grid;
    align-content: start;
    gap: 0.6rem;
  }

  .metric-row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 0.1rem 0.75rem;
    padding: 0.85rem 0.95rem;
    border: 1px solid rgba(122, 132, 148, 0.22);
    border-radius: 8px;
  }

  .metric-row span,
  .metric-row small {
    color: #64748b;
  }

  .metric-row span {
    font-size: 0.78rem;
  }

  .metric-row strong {
    color: #1e293b;
    font-size: 1.2rem;
    font-variant-numeric: tabular-nums;
  }

  .metric-row small {
    grid-column: 1 / -1;
    font-size: 0.7rem;
  }

  .metric-row.accent {
    background: rgba(38, 70, 83, 0.06);
    border-color: rgba(38, 70, 83, 0.25);
  }

  .signal-grid,
  .evidence-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 1rem;
    margin-top: 1rem;
  }

  .signal-panel,
  .evidence-grid > div {
    border-top: 1px solid rgba(122, 132, 148, 0.22);
    padding-top: 0.75rem;
  }

  .signal-row,
  .evidence-row {
    padding: 0.45rem 0;
    border-bottom: 1px solid rgba(122, 132, 148, 0.14);
    font-size: 0.78rem;
  }

  .signal-name {
    color: #1e293b;
    font-weight: 600;
  }

  .signal-values {
    margin-left: auto;
    color: #1e293b;
    font-variant-numeric: tabular-nums;
  }

  .signal-delta {
    min-width: 7.5rem;
    color: #b45309;
    text-align: right;
    font-size: 0.7rem;
    font-weight: 600;
  }

  .evidence-row small {
    color: #64748b;
    font-size: 0.7rem;
  }

  @media (max-width: 768px) {
    .retrospective-overview,
    .signal-grid,
    .evidence-grid {
      grid-template-columns: 1fr;
    }

    .retrospective-heading {
      align-items: flex-start;
      flex-direction: column;
    }

    .legend-inline {
      display: block;
      margin-top: 0.35rem;
      white-space: normal;
    }
  }
</style>
