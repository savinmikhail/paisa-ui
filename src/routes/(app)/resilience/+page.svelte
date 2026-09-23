<script lang="ts">
  import { onMount } from "svelte";
  import { ajax, formatCurrency, formatFloat } from "$lib/utils";
  import {
    filterResiliencePoints,
    normalizeResilienceReport,
    resilienceDateInputValue,
    resilienceDomain,
    resilienceX,
    resilienceY,
    RESILIENCE_METRICS,
    type ResilienceMetric,
    type ResiliencePoint,
    type ResilienceReport
  } from "$lib/resilience";
  import { loadResilienceFixture } from "$lib/fixtures/resilience";

  const plotWidth = 580;
  const plotHeight = 190;
  const plotLeft = 58;
  const plotRight = 14;
  const plotTop = 12;
  const plotBottom = 30;

  let report: ResilienceReport = null;
  let visiblePoints: ResiliencePoint[] = [];
  let draftFrom = "";
  let draftTo = "";
  let hoveredIndex: number = null;
  let loadError = "";
  let fixtureMode = false;

  $: if (report && !draftFrom && !draftTo) {
    draftFrom = resilienceDateInputValue(report.range_start);
    draftTo = resilienceDateInputValue(report.range_end);
  }

  $: visiblePoints =
    report && draftFrom && draftTo
      ? filterResiliencePoints(report.points, draftFrom, draftTo)
      : report?.points || [];

  $: chartConfigs = RESILIENCE_METRICS.map((metric) => ({
    ...metric,
    domain: resilienceDomain(visiblePoints, metric.key)
  }));

  $: hoveredPoint = hoveredIndex === null ? null : visiblePoints[hoveredIndex] || null;

  onMount(async () => {
    fixtureMode = new URLSearchParams(window.location.search).get("fixture") === "resilience";
    try {
      const result = fixtureMode
        ? loadResilienceFixture()
        : ((await ajax("/api/resilience")) as ResilienceReport & { error?: string });
      if ("error" in result && result.error) {
        loadError = String(result.error);
        return;
      }
      report = normalizeResilienceReport(result);
      draftFrom = resilienceDateInputValue(report.range_start);
      draftTo = resilienceDateInputValue(report.range_end);
    } catch (error) {
      loadError = error instanceof Error ? error.message : "Resilience API could not be loaded";
    }
  });

  function metricValue(point: ResiliencePoint, metric: ResilienceMetric): number {
    return point[metric];
  }

  function linePath(points: ResiliencePoint[], metric: ResilienceMetric, domain: [number, number]) {
    return points
      .map((point, index) => {
        const x = resilienceX(index, points.length, plotWidth);
        const y = resilienceY(metricValue(point, metric), domain, plotHeight);
        return `${index === 0 ? "M" : "L"} ${x} ${y}`;
      })
      .join(" ");
  }

  function formatMetricValue(metric: ResilienceMetric, value: number): string {
    if (metric === "life_months") {
      return `${formatFloat(value, 1)} months`;
    }
    return `${formatCurrency(value, 0)} ${metric.toUpperCase()}`;
  }

  function hoverPoint(event: MouseEvent) {
    if (visiblePoints.length === 0) return;
    const svg = event.currentTarget as SVGSVGElement;
    const bounds = svg.getBoundingClientRect();
    const x = Math.max(0, Math.min(bounds.width, event.clientX - bounds.left));
    hoveredIndex = Math.min(
      visiblePoints.length - 1,
      Math.max(0, Math.round((x / bounds.width) * (visiblePoints.length - 1)))
    );
  }

  function clearHover() {
    hoveredIndex = null;
  }
</script>

<svelte:head>
  <title>Resilience · Paisa</title>
</svelte:head>

<section class="section resilience-page" data-testid="resilience-screen">
  <div class="container is-fluid">
    <div class="resilience-heading">
      <div>
        <h1 class="title is-3 mb-2">Resilience</h1>
        <p class="subtitle is-6 mb-0">
          How long the deliberately liquid set can carry regular personal burn.
        </p>
      </div>
      <div class="has-text-right is-size-7 has-text-grey">
        <div>As of {report?.as_of || "loading"}</div>
        <div>{report?.contract_version || "resilience.v1"} · read-only</div>
      </div>
    </div>

    <div class="notification is-info is-light invertable resilience-notice" data-testid="source-notice">
      <span class="icon mr-2"><i class="fas fa-flask" aria-hidden="true" /></span>
      {#if fixtureMode}
        <span><strong>Synthetic fixture.</strong> No private Ledger facts are loaded in this capture.</span>
      {:else}
        <span><strong>Read-only API.</strong> This page does not write Ledger data.</span>
      {/if}
    </div>

    <div class="box resilience-toolbar" data-testid="resilience-range-controls">
      <div>
        <div class="has-text-weight-semibold">Shared date range</div>
        <div class="is-size-7 has-text-grey">All three small multiples use the same inclusive range and hover.</div>
      </div>
      <form class="resilience-range-form" on:submit|preventDefault={() => (hoveredIndex = null)}>
        <label class="label is-size-7 mb-1" for="resilience-from">From</label>
        <input id="resilience-from" class="input is-small" type="date" bind:value={draftFrom} />
        <label class="label is-size-7 mb-1" for="resilience-to">To</label>
        <input id="resilience-to" class="input is-small" type="date" bind:value={draftTo} />
        <button class="button is-small is-primary" type="submit">Apply</button>
      </form>
    </div>

    {#if loadError}
      <div class="message invertable is-warning" data-testid="resilience-error">
        <div class="message-header">Resilience data is unavailable</div>
        <div class="message-body">
          {loadError}
          <div class="mt-2 is-size-7">
            The current slice ships a deterministic synthetic adapter only; it does not claim a private runtime adapter.
          </div>
        </div>
      </div>
    {:else if !report}
      <div class="box has-text-centered has-text-grey" data-testid="resilience-loading">Loading read-only resilience…</div>
    {:else}
      <div class="resilience-chart-grid" data-testid="resilience-charts">
        {#each chartConfigs as chart}
          <article class="box resilience-chart-card">
            <div class="resilience-chart-heading">
              <div>
                <h2 class="title is-5 mb-1">{chart.label}</h2>
                <p class="is-size-7 has-text-grey">Independent {chart.unit} y-scale</p>
              </div>
              <strong>{formatMetricValue(chart.key, hoveredPoint?.[chart.key] ?? visiblePoints.at(-1)?.[chart.key] ?? 0)}</strong>
            </div>
            <svg
              class="resilience-chart"
              viewBox={`0 0 ${plotWidth + plotLeft + plotRight} ${plotHeight + plotTop + plotBottom}`}
              role="img"
              aria-label={`${chart.label} over the selected date range`}
              on:mousemove={hoverPoint}
              on:mouseleave={clearHover}
            >
              <g transform={`translate(${plotLeft}, ${plotTop})`}>
                {#each [0, 0.5, 1] as tick}
                  <line x1="0" x2={plotWidth} y1={tick * plotHeight} y2={tick * plotHeight} class="resilience-gridline" />
                  <text x="-8" y={tick * plotHeight + 4} text-anchor="end" class="resilience-axis-label">
                    {formatMetricValue(chart.key, chart.domain[1] * (1 - tick))}
                  </text>
                {/each}
                {#if visiblePoints.length > 0}
                  <path d={linePath(visiblePoints, chart.key, chart.domain)} class="resilience-line" style={`stroke: ${chart.color}`} />
                  {#each visiblePoints as point, index}
                    <circle
                      cx={resilienceX(index, visiblePoints.length, plotWidth)}
                      cy={resilienceY(metricValue(point, chart.key), chart.domain, plotHeight)}
                      r={hoveredIndex === index ? 5 : 3}
                      class="resilience-point"
                      style={`fill: ${chart.color}`}
                    />
                  {/each}
                  {#if hoveredIndex !== null && hoveredPoint}
                    <line
                      x1={resilienceX(hoveredIndex, visiblePoints.length, plotWidth)}
                      x2={resilienceX(hoveredIndex, visiblePoints.length, plotWidth)}
                      y1="0"
                      y2={plotHeight}
                      class="resilience-hover-line"
                    />
                  {/if}
                  <text x="0" y={plotHeight + 22} class="resilience-axis-label">
                    {visiblePoints[0].date.format("MMM YYYY")}
                  </text>
                  <text x={plotWidth} y={plotHeight + 22} text-anchor="end" class="resilience-axis-label">
                    {visiblePoints.at(-1).date.format("MMM YYYY")}
                  </text>
                {/if}
              </g>
            </svg>
          </article>
        {/each}
      </div>

      {#if hoveredPoint}
        <div class="resilience-hover-summary" data-testid="resilience-hover-summary">
          <span class="has-text-weight-semibold">{hoveredPoint.date.format("D MMM YYYY")}</span>
          <span>{formatMetricValue("rub", hoveredPoint.rub)}</span>
          <span>{formatMetricValue("usd", hoveredPoint.usd)}</span>
          <span>{formatMetricValue("life_months", hoveredPoint.life_months)}</span>
        </div>
      {/if}

      <div class="columns is-variable is-5 mt-1">
        <div class="column is-7">
          <div class="box resilience-panel">
            <div class="panel-heading-row">
              <div>
                <h2 class="title is-5 mb-1">Liquid scope</h2>
                <p class="is-size-7 has-text-grey">Included balances are the only balances represented by the series.</p>
              </div>
              <span class="tag is-rounded is-success is-light">{report.liquid_scope.included.length} included</span>
            </div>
            <div class="scope-list">
              {#each report.liquid_scope.included as item}
                <div class="scope-row included">
                  <div><strong>{item.name}</strong><span class="is-size-7 has-text-grey">{item.kind} · {item.currency}</span></div>
                  <span class="tag is-success is-light">included</span>
                </div>
              {/each}
              {#each report.liquid_scope.excluded as item}
                <div class="scope-row excluded">
                  <div><strong>{item.name}</strong><span class="is-size-7 has-text-grey">{item.kind} · {item.currency}</span></div>
                  <span class="tag is-light">excluded</span>
                </div>
              {/each}
            </div>
          </div>
        </div>
        <div class="column is-5">
          <div class="box resilience-panel" data-testid="burn-provenance">
            <div class="has-text-grey is-size-7">Life Months denominator</div>
            <h2 class="title is-5 mb-2">{formatCurrency(report.burn_provenance.value, 0)} {report.burn_provenance.currency} / month</h2>
            <div class="provenance-callout">
              <strong>{report.burn_provenance.metric}</strong>
              <span>{report.burn_provenance.source}</span>
              <span>{report.burn_provenance.period_start.format("D MMM YYYY")} – {report.burn_provenance.period_end.format("D MMM YYYY")}</span>
            </div>
            <p class="is-size-7 has-text-grey mt-3 mb-0">{report.burn_provenance.description}</p>
          </div>
        </div>
      </div>
    {/if}
  </div>
</section>

<style>
  .resilience-page {
    padding-top: 1rem;
  }

  .resilience-heading,
  .resilience-toolbar,
  .resilience-chart-heading,
  .panel-heading-row,
  .scope-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
  }

  .resilience-notice {
    margin-top: 1.25rem;
  }

  .resilience-toolbar {
    align-items: flex-end;
    margin-top: 1rem;
    padding: 1rem 1.25rem;
  }

  .resilience-range-form {
    display: flex;
    align-items: flex-end;
    gap: 0.5rem;
  }

  .resilience-range-form .label {
    margin-left: 0.25rem;
  }

  .resilience-chart-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 1rem;
    margin-top: 1rem;
  }

  .resilience-chart-card,
  .resilience-panel {
    padding: 1rem;
  }

  .resilience-chart-heading strong {
    color: #363636;
    white-space: nowrap;
  }

  .resilience-chart {
    display: block;
    width: 100%;
    overflow: visible;
    margin-top: 0.75rem;
  }

  .resilience-gridline {
    stroke: #e8e8e8;
    stroke-width: 1;
  }

  .resilience-axis-label {
    fill: #7a7a7a;
    font-size: 11px;
  }

  .resilience-line {
    fill: none;
    stroke-width: 3;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .resilience-point {
    stroke: white;
    stroke-width: 2;
  }

  .resilience-hover-line {
    stroke: #7a7a7a;
    stroke-dasharray: 3 3;
    stroke-width: 1;
  }

  .resilience-hover-summary {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    align-items: center;
    padding: 0.75rem 1rem;
    margin: 0.5rem 0 1rem;
    background: white;
    border: 1px solid #dbdbdb;
    border-radius: 4px;
  }

  .scope-list {
    margin-top: 1rem;
  }

  .scope-row {
    align-items: center;
    padding: 0.7rem 0;
    border-top: 1px solid #f0f0f0;
  }

  .scope-row div {
    display: flex;
    flex-direction: column;
  }

  .scope-row.excluded {
    opacity: 0.72;
  }

  .provenance-callout {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    padding: 0.75rem;
    border-left: 3px solid #f14668;
    background: #fff5f7;
  }

  @media screen and (max-width: 1023px) {
    .resilience-chart-grid {
      grid-template-columns: 1fr;
    }
  }

  @media screen and (max-width: 768px) {
    .resilience-heading,
    .resilience-toolbar,
    .resilience-range-form {
      align-items: stretch;
      flex-direction: column;
    }

    .resilience-heading > div:last-child {
      text-align: left !important;
    }

    .resilience-range-form .label {
      margin-left: 0;
      margin-bottom: -0.3rem !important;
    }
  }
</style>
