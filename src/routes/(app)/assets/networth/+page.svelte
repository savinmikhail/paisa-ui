<script lang="ts">
  import dayjs from "dayjs";
  import {
    ajax,
    formatCurrency,
    formatFloat,
    isMobile,
    type Legend,
    type Networth
  } from "$lib/utils";
  import COLORS from "$lib/colors";
  import { renderNetworth } from "$lib/networth";
  import {
    dateInputValue,
    reduceNetworthInteraction,
    summarizeNetworthInterval,
    type NetworthFocus,
    type NetworthIntervalSummary,
    type NetworthReconciliation
  } from "$lib/networth_reconciliation";
  import {
    NETWORTH_TICKET_04_FIXTURE,
    ticket04FixtureReconciliation,
    type NetworthFixturePoint
  } from "$lib/fixtures/networth-ticket-04";
  import _ from "lodash";
  import { onDestroy, onMount } from "svelte";
  import { dateRange, setAllowedDateRange } from "../../../../store";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import ZeroState from "$lib/components/ZeroState.svelte";
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import LegendCard from "$lib/components/LegendCard.svelte";

  let networth = 0;
  let investment = 0;
  let gain = 0;
  let xirr = 0;
  let svg: Element;
  let destroy: () => void;
  let points: Networth[] = [];
  let legends: Legend[] = [];
  let visiblePoints: Networth[] = [];
  let intervalSummary: NetworthIntervalSummary = null;
  let focusedPoint: NetworthFocus = null;
  let draftFrom = "";
  let draftTo = "";
  let customRange = false;
  let fixtureMode = false;
  let reconciliation: NetworthReconciliation = null;
  let reconciliationRequest = 0;

  $: rangeFrom = customRange && draftFrom ? dayjs(draftFrom) : $dateRange.from;
  $: rangeTo = customRange && draftTo ? dayjs(draftTo) : $dateRange.to;

  $: visiblePoints = _.filter(
    points,
    (p) => !p.date.isBefore(rangeFrom, "day") && !p.date.isAfter(rangeTo, "day")
  );

  $: intervalSummary = summarizeNetworthInterval(visiblePoints);

  $: if (points.length > 0 && rangeFrom && rangeTo) {
    void loadReconciliation(rangeFrom, rangeTo);
  }

  function replaceNetworthChart(pointsToRender: Networth[], chartElement: Element) {
    if (destroy) {
      destroy();
    }

    return renderNetworth(pointsToRender, chartElement, {
      onHover: (point, series) => {
        focusedPoint = reduceNetworthInteraction(
          { focus: focusedPoint },
          { type: "focus", focus: { point, series } }
        ).focus;
      },
      onLeave: () => {
        focusedPoint = reduceNetworthInteraction({ focus: focusedPoint }, { type: "blur" }).focus;
      },
      onRangeSelect: selectNetworthRange
    });
  }

  $: if (!_.isEmpty(visiblePoints) && svg) {
    const rendered = replaceNetworthChart(visiblePoints, svg);
    destroy = rendered.destroy;
    legends = rendered.legends;
  }

  onDestroy(() => {
    if (destroy) {
      destroy();
    }
  });

  onMount(async () => {
    fixtureMode = new URLSearchParams(window.location.search).get("fixture") === "networth-ticket-04";
    const result = fixtureMode
      ? { networthTimeline: NETWORTH_TICKET_04_FIXTURE, xirr: 0.08 }
      : await ajax("/api/networth");
    points = result.networthTimeline;
    setAllowedDateRange(_.map(points, (p) => p.date));

    if (points.length > 0) {
      draftFrom = dateInputValue(points[0].date);
      draftTo = dateInputValue(points[points.length - 1].date);
    }

    const current = _.last(points);
    if (current) {
      networth = current.investmentAmount + current.gainAmount - current.withdrawalAmount;
      investment = current.investmentAmount - current.withdrawalAmount;
      gain = current.gainAmount;
    }
    xirr = result.xirr;
  });

  async function loadReconciliation(from: dayjs.Dayjs, to: dayjs.Dayjs) {
    const request = ++reconciliationRequest;
    if (fixtureMode) {
      reconciliation = ticket04FixtureReconciliation(from, to);
      return;
    }

    const next = await ajax(
      `/api/networth/reconciliation?from=${dateInputValue(from)}&to=${dateInputValue(to)}`,
      { background: true }
    );
    if (request === reconciliationRequest) {
      reconciliation = next;
    }
  }

  function applyCustomRange() {
    if (!draftFrom || !draftTo || dayjs(draftFrom).isAfter(dayjs(draftTo), "day")) {
      return;
    }
    customRange = true;
    focusedPoint = null;
  }

  function clearCustomRange() {
    customRange = false;
    focusedPoint = null;
    if (points.length > 0) {
      draftFrom = dateInputValue(points[0].date);
      draftTo = dateInputValue(points[points.length - 1].date);
    }
  }

  function selectNetworthRange(from: dayjs.Dayjs, to: dayjs.Dayjs) {
    draftFrom = dateInputValue(from);
    draftTo = dateInputValue(to);
    customRange = true;
    focusedPoint = null;
  }

  const decompositionRows = [
    ["Net savings", "netSavings"],
    ["Interest / coupons / dividends", "passiveIncome"],
    ["Market", "market"],
    ["FX", "fx"],
    ["Debt", "debt"],
    ["Corrections", "corrections"]
  ] as const;
</script>

<section class="section tab-networth">
  <div class="container is-fluid">
    <div class="networth-range-toolbar" data-testid="networth-range-controls">
      <div>
        <h1 class="title is-5 mb-1">Net Worth interval</h1>
        <p class="is-size-7 has-text-grey">Choose any two dates to explain the change.</p>
      </div>
      <form class="networth-range-form" on:submit|preventDefault={applyCustomRange}>
        <label class="label is-size-7 mb-1" for="networth-from">From</label>
        <input id="networth-from" class="input is-small" type="date" bind:value={draftFrom} />
        <label class="label is-size-7 mb-1" for="networth-to">To</label>
        <input id="networth-to" class="input is-small" type="date" bind:value={draftTo} />
        <button class="button is-small is-primary" type="submit">Apply</button>
        {#if customRange}
          <button class="button is-small" type="button" on:click={clearCustomRange}>Reset</button>
        {/if}
      </form>
    </div>

    {#if intervalSummary}
      <div class="networth-interval-summary" data-testid="networth-interval-summary">
        <div>
          <span class="is-size-7 has-text-grey">Opening</span>
          <strong>{formatCurrency(intervalSummary.opening)}</strong>
        </div>
        <div>
          <span class="is-size-7 has-text-grey">Closing</span>
          <strong>{formatCurrency(intervalSummary.closing)}</strong>
        </div>
        <div>
          <span class="is-size-7 has-text-grey">Absolute delta</span>
          <strong class:has-text-danger={intervalSummary.delta < 0} class:has-text-success={intervalSummary.delta >= 0}
            >{formatCurrency(intervalSummary.delta)}</strong
          >
        </div>
        <div>
          <span class="is-size-7 has-text-grey">Percent delta</span>
          <strong
            >{intervalSummary.percentDelta === null
              ? "—"
              : `${formatFloat(intervalSummary.percentDelta)}%`}</strong
          >
        </div>
      </div>
    {/if}

    <nav class="level {isMobile() && 'grid-2'}">
      <LevelItem title="Net worth" color={COLORS.primary} value={formatCurrency(networth)} />
      <LevelItem
        title="Net Investment"
        color={COLORS.secondary}
        value={formatCurrency(investment)}
      />
      <LevelItem
        title="Gain / Loss"
        color={gain >= 0 ? COLORS.gainText : COLORS.lossText}
        value={formatCurrency(gain)}
      />
      <LevelItem title="XIRR" value={formatFloat(xirr)} />
    </nav>
  </div>
</section>

<section class="section tab-networth">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box overflow-x-auto">
          <ZeroState item={points}>
            <strong>Oops!</strong> You have no transactions.
          </ZeroState>

          <LegendCard {legends} clazz="ml-4" />
          <svg id="d3-networth-timeline" data-testid="networth-chart" height="500" bind:this={svg} />
        </div>
      </div>
    </div>

    <div class="columns is-variable is-4 mt-1">
      <div class="column is-5">
        <div class="box networth-detail" data-testid="networth-hover-detail">
          <div class="is-flex is-justify-content-space-between is-align-items-baseline">
            <h2 class="title is-6 mb-2">Detail</h2>
            {#if focusedPoint}<span class="tag is-light">Hover focus</span>{/if}
          </div>
          {#if focusedPoint}
            <p class="is-size-7 has-text-grey mb-1">{focusedPoint.point.date.format("DD MMM YYYY")}</p>
            <div class="networth-focus-value">{formatCurrency(
              focusedPoint.point.investmentAmount + focusedPoint.point.gainAmount - focusedPoint.point.withdrawalAmount
            )}</div>
            <p class="is-size-7 has-text-grey">{focusedPoint.series === "networth" ? "Net Worth" : "Net Investment"}</p>
          {:else}
            <p class="is-size-7 has-text-grey">Hover the chart to inspect a synchronized point.</p>
          {/if}
        </div>
      </div>
      <div class="column is-7">
        <div class="box networth-detail" data-testid="networth-decomposition">
          <div class="is-flex is-justify-content-space-between is-align-items-baseline">
            <h2 class="title is-6 mb-2">Interval decomposition</h2>
            {#if reconciliation}<span class="tag is-light">Reconciled</span>{/if}
          </div>
          {#if reconciliation}
            <table class="table is-fullwidth is-narrow is-size-7">
              <tbody>
                {#each decompositionRows as [label, key]}
                  <tr>
                    <th>{label}</th>
                    <td class:has-text-danger={reconciliation.decomposition[key] < 0} class="has-text-right">
                      {formatCurrency(reconciliation.decomposition[key])}
                    </td>
                  </tr>
                {/each}
                <tr class="has-text-weight-bold">
                  <th>Closing − opening</th>
                  <td class="has-text-right">{formatCurrency(reconciliation.decomposition.totalDelta)}</td>
                </tr>
              </tbody>
            </table>
          {:else}
            <p class="is-size-7 has-text-grey">Loading reconciliation…</p>
          {/if}
        </div>
      </div>
    </div>
    <BoxLabel text="Networth Timeline" />
  </div>
</section>

<style lang="scss">
  .networth-range-toolbar {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 1.25rem;
  }

  .networth-range-form {
    display: grid;
    grid-template-columns: repeat(2, minmax(9rem, 1fr)) auto auto;
    align-items: end;
    gap: 0.5rem;
  }

  .networth-range-form .label {
    grid-row: 1;
  }

  .networth-range-form input {
    grid-row: 2;
  }

  .networth-range-form button {
    grid-row: 2;
  }

  .networth-interval-summary {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 0.75rem;
    margin-bottom: 1rem;
  }

  .networth-interval-summary > div {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    padding: 0.75rem 0;
    border-top: 1px solid var(--bulma-border);
  }

  .networth-interval-summary strong,
  .networth-focus-value {
    font-variant-numeric: tabular-nums;
  }

  .networth-detail {
    height: 100%;
  }

  .networth-focus-value {
    font-size: 1.65rem;
    font-weight: 600;
  }

  .networth-detail table {
    margin-bottom: 0;
  }

  :global(.networth-selection) {
    fill: var(--bulma-primary);
    fill-opacity: 0.12;
    stroke: var(--bulma-primary);
    stroke-opacity: 0.45;
    pointer-events: none;
  }

  @media screen and (max-width: 768px) {
    .networth-range-toolbar {
      align-items: stretch;
      flex-direction: column;
    }

    .networth-range-form {
      grid-template-columns: 1fr 1fr;
    }

    .networth-range-form button {
      grid-row: 3;
    }

    .networth-interval-summary {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }
</style>
