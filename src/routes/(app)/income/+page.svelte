<script lang="ts">
  import COLORS from "$lib/colors";
  import BoxLabel from "$lib/components/BoxLabel.svelte";
  import LegendCard from "$lib/components/LegendCard.svelte";
  import LevelItem from "$lib/components/LevelItem.svelte";
  import {
    renderMonthlyInvestmentTimeline,
    renderPassiveIncomeTimeline,
    renderYearlyIncomeTimeline,
    renderYearlyTimelineOf
  } from "$lib/income";
  import {
    ajax,
    formatCurrency,
    formatPercentage,
    type Legend,
    type PassiveIncomeAnalysis,
    type PassiveIncomeWindow
  } from "$lib/utils";
  import _ from "lodash";
  import { onMount, tick } from "svelte";

  let grossIncome = 0;
  let netTax = 0;

  let monthlyInvestmentTimelineLegends: Legend[] = [];
  let yearlyIncomeTimelineLegends: Legend[] = [];
  let yearlyNetIncomeTimelineLegends: Legend[] = [];
  let yearlyNetTaxTimelineLegends: Legend[] = [];
  let passiveIncome: PassiveIncomeAnalysis;
  let passiveIncomeTimelineLegends: Legend[] = [];

  function coverage(value: number) {
    return value > 0 ? formatPercentage(value) : "Not configured";
  }

  function coverageSummary(window: PassiveIncomeWindow) {
    return `Burn ${coverage(window.coverage_regular_personal_burn)} · Floor ${coverage(
      window.coverage_freedom_income_floor
    )}`;
  }

  onMount(async () => {
    const {
      income_timeline: incomes,
      tax_timeline: taxes,
      yearly_cards: yearlyCards
    } = await ajax("/api/income");
    const [analysis, { cash_flows: cashFlows }] = await Promise.all([
      ajax("/api/income/analysis"),
      ajax("/api/cash_flow")
    ]);
    passiveIncome = analysis.passive_income;
    await tick();
    monthlyInvestmentTimelineLegends = renderMonthlyInvestmentTimeline(incomes, cashFlows);
    yearlyIncomeTimelineLegends = renderYearlyIncomeTimeline(yearlyCards);
    yearlyNetIncomeTimelineLegends = renderYearlyTimelineOf(
      "Net Income",
      "net_income",
      COLORS.gain,
      yearlyCards
    );
    yearlyNetTaxTimelineLegends = renderYearlyTimelineOf(
      "Net Tax",
      "net_tax",
      COLORS.loss,
      yearlyCards
    );
    passiveIncomeTimelineLegends = renderPassiveIncomeTimeline(passiveIncome);

    grossIncome = _.sumBy(incomes, (i) => _.sumBy(i.postings, (p) => -p.amount));
    netTax = _.sumBy(taxes, (t) => _.sumBy(t.postings, (p) => p.amount));
  });
</script>

<section class="section tab-income">
  <div class="container">
    <nav class="level">
      <LevelItem title="Gross Income" value={formatCurrency(grossIncome)} color={COLORS.gainText} />
      <LevelItem title="Net Tax" value={formatCurrency(netTax)} color={COLORS.lossText} />
    </nav>
  </div>
</section>
<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-12">
        <div class="box">
          <LegendCard legends={monthlyInvestmentTimelineLegends} clazz="ml-4" />
          <svg id="d3-income-timeline" width="100%" height="500" />
        </div>
      </div>
    </div>
    <BoxLabel
      text="Monthly Income Timeline · Savings rate: values outside −100%…100% appear at the edge; hover for exact values"
    />
  </div>
</section>
{#if passiveIncome}
  <section class="section">
    <div class="container is-fluid">
      <nav class="level passive-income-summary">
        <LevelItem
          title="Passive income · current month"
          value={formatCurrency(passiveIncome.current_month.average)}
          color={COLORS.gainText}
          subtitle={coverageSummary(passiveIncome.current_month)}
        />
        <LevelItem
          title="Passive income · 3m average"
          value={formatCurrency(passiveIncome.trailing_3_months.average)}
          color={COLORS.gainText}
          subtitle={coverageSummary(passiveIncome.trailing_3_months)}
        />
        <LevelItem
          title="Passive income · 12m average"
          value={formatCurrency(passiveIncome.trailing_12_months.average)}
          color={COLORS.gainText}
          subtitle={`${coverageSummary(passiveIncome.trailing_12_months)} · Total ${formatCurrency(
            passiveIncome.trailing_12_months.total
          )}`}
        />
        <LevelItem
          title="Coverage baselines"
          value={formatCurrency(passiveIncome.regular_personal_burn)}
          subtitle={`Burn · ${formatCurrency(
            passiveIncome.regular_personal_burn
          )} / Freedom floor · ${
            passiveIncome.freedom_income_floor > 0
              ? formatCurrency(passiveIncome.freedom_income_floor)
              : "not configured"
          }`}
        />
      </nav>
    </div>
  </section>
  <section class="section pt-0">
    <div class="container is-fluid">
      <div class="columns">
        <div class="column is-12">
          <div class="box">
            <LegendCard legends={passiveIncomeTimelineLegends} clazz="ml-4" />
            <svg id="d3-passive-income-timeline" width="100%" height="380" />
          </div>
        </div>
      </div>
      <BoxLabel text="Realized Passive Income · coverage excludes FX and market revaluation" />
    </div>
  </section>
{/if}
<section class="section">
  <div class="container is-fluid">
    <div class="columns">
      <div class="column is-one-third">
        <div class="box px-3">
          <LegendCard legends={yearlyIncomeTimelineLegends} clazz="ml-4" />
          <svg id="d3-yearly-income-timeline" width="100%" />
        </div>
      </div>
      <div class="column is-one-third">
        <div class="box px-3">
          <LegendCard legends={yearlyNetIncomeTimelineLegends} clazz="ml-4" />
          <svg id="d3-yearly-net_income-timeline" width="100%" />
        </div>
      </div>
      <div class="column is-one-third">
        <div class="box px-3">
          <LegendCard legends={yearlyNetTaxTimelineLegends} clazz="ml-4" />
          <svg id="d3-yearly-net_tax-timeline" width="100%" />
        </div>
      </div>
    </div>
    <BoxLabel text="Financial Year Income Timeline" />
  </div>
</section>

<style lang="scss">
  .passive-income-summary {
    display: grid !important;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1rem;
    align-items: stretch;

    :global(.level-item) {
      min-width: 0;
      margin: 0;
      padding: 1rem;
      border: 1px solid rgba(128, 128, 128, 0.25);
      border-radius: 4px;
    }

    :global(.title) {
      font-variant-numeric: tabular-nums;
      white-space: nowrap;
    }
  }

  @media screen and (max-width: 1023px) {
    .passive-income-summary {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media screen and (max-width: 600px) {
    .passive-income-summary {
      grid-template-columns: 1fr;
    }
  }
</style>
