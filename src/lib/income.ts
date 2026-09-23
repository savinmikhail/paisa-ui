import * as d3 from "d3";
import dayjs from "dayjs";
import _ from "lodash";
import COLORS, { generateColorScheme } from "./colors";
import {
  formatCurrency,
  formatCurrencyCrude,
  type Income,
  type Posting,
  restName,
  secondName,
  skipTicks,
  tooltip,
  type IncomeYearlyCard,
  type Legend,
  type PassiveIncomeAnalysis
} from "./utils";
import { incomeTimelineDomain } from "./income_scale";
import { incomeGroup, rollingMedian, selectedIncomeTotals } from "./income_reporting";

export { incomeGroup } from "./income_reporting";

export { incomeTimelineDomain } from "./income_scale";

export function renderMonthlyInvestmentTimeline(incomes: Income[]): Legend[] {
  return renderIncomeTimeline(incomes, "#d3-income-timeline", "MMM-YYYY");
}

function renderIncomeTimeline(incomes: Income[], id: string, timeFormat: string): Legend[] {
  const MAX_BAR_WIDTH = 40;
  const svg = d3.select(id),
    margin = { top: 20, right: 30, bottom: 80, left: 40 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    height = +svg.attr("height") - margin.top - margin.bottom,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const postings = _.flatMap(incomes, (i) => i.postings);
  const groupKeys = _.chain(postings)
    .map((p) => incomeGroup(p))
    .uniq()
    .sort()
    .value();

  const groupTotal = _.chain(postings)
    .groupBy((p) => incomeGroup(p))
    .map((postings, key) => {
      const total = _.sumBy(postings, (p) => -p.amount);
      return [key, `${key}\n${formatCurrency(total)}`];
    })
    .fromPairs()
    .value();

  const defaultValues = _.zipObject(
    groupKeys,
    _.map(groupKeys, () => 0)
  );

  interface Point {
    date: dayjs.Dayjs;
    month: string;
    [key: string]: number | string | dayjs.Dayjs;
  }
  let points: Point[] = [];

  points = _.map(incomes, (i) => {
    const values = _.chain(i.postings)
      .groupBy((p) => incomeGroup(p))
      .flatMap((postings, key) => [[key, _.sumBy(postings, (p) => -p.amount)]])
      .fromPairs()
      .value();

    return _.merge(
      {
        month: i.date.format(timeFormat),
        date: i.date,
        postings: i.postings
      },
      defaultValues,
      values
    );
  });

  const x = d3.scaleBand().range([0, width]).paddingInner(0.1).paddingOuter(0);
  const y = d3.scaleLinear().range([height, 0]);
  const z = generateColorScheme(groupKeys);

  const xAxis = g.append("g").attr("class", "axis x");
  const yAxis = g.append("g").attr("class", "axis y");
  const bars = g.append("g");
  const medianPath = g
    .append("path")
    .attr("fill", "none")
    .attr("stroke", COLORS.danger)
    .attr("stroke-width", 2)
    .attr("stroke-dasharray", "7,5");
  const medianHitTargets = g.append("g").attr("class", "median-hit-targets");

  const tooltipContent = (visibleGroups: string[]) => (d: any) => {
    const postings = _.filter((d.data as unknown as { postings: Posting[] }).postings, (posting) =>
      visibleGroups.includes(incomeGroup(posting))
    );
    const total = _.sumBy(postings, (p) => -p.amount);

    return tooltip(
      _.sortBy(
        postings.map((p) => [
          restName(p.account),
          [formatCurrency(-p.amount), "has-text-weight-bold has-text-right"]
        ]),
        (r) => r[0]
      ),
      { total: formatCurrency(total) }
    );
  };

  let firstRender = true;
  let showMedian = true;
  const render = (visibleGroups: string[]) => {
    x.domain(points.map((p) => p.month));
    const selectedTotals = selectedIncomeTotals(incomes, visibleGroups);
    const medians = rollingMedian(selectedTotals, 4);
    const actualDomain = incomeTimelineDomain(points, visibleGroups);
    const medianValues = medians.filter((value): value is number => value !== null);
    const medianMin = medianValues.length ? Math.min(...medianValues) : 0;
    const medianMax = medianValues.length ? Math.max(...medianValues) : 0;
    const domain: [number, number] = [
      Math.min(actualDomain[0], medianMin),
      Math.max(actualDomain[1], medianMax)
    ];
    y.domain(domain[0] === domain[1] ? [0, 1] : domain);

    const transition = svg.transition().duration(firstRender ? 0 : 350);
    firstRender = false;

    xAxis
      .attr("transform", "translate(0," + height + ")")
      .transition(transition)
      .call(
        d3
          .axisBottom(x)
          .ticks(5)
          .tickFormat(skipTicks(30, x, (d) => d.toString()))
      )
      .selectAll("text")
      .attr("y", 10)
      .attr("x", -8)
      .attr("dy", ".35em")
      .attr("transform", "rotate(-45)")
      .style("text-anchor", "end");

    yAxis
      .transition(transition)
      .call(d3.axisLeft(y).tickSize(-width).tickFormat(formatCurrencyCrude));

    bars
      .selectAll("g")
      .data(
        d3.stack().offset(d3.stackOffsetDiverging).keys(visibleGroups)(
          points as { [key: string]: number }[]
        ),
        (d: any) => d.key
      )
      .join(
        (enter) => enter.append("g"),
        (update) => update,
        (exit) => exit.remove()
      )
      .attr("fill", (d: any) => z(d.key.split("-")[0]))
      .selectAll("rect")
      .data(
        (d: any) => d,
        (d: any) => d.data.month
      )
      .join(
        (enter) => enter.append("rect"),
        (update) => update,
        (exit) => exit.remove()
      )
      .attr("data-tippy-content", tooltipContent(visibleGroups))
      .attr(
        "x",
        (d: any) => x(d.data.month) + (x.bandwidth() - Math.min(x.bandwidth(), MAX_BAR_WIDTH)) / 2
      )
      .attr("y", (d: any) => y(d[1]))
      .attr("height", (d: any) => y(d[0]) - y(d[1]))
      .attr("width", Math.min(x.bandwidth(), MAX_BAR_WIDTH));

    const line = d3
      .line<{ month: string; median: number | null }>()
      .defined((d) => d.median !== null)
      .x((d) => x(d.month) + x.bandwidth() / 2)
      .y((d) => y(d.median as number));
    medianPath
      .datum(points.map((point, index) => ({ month: point.month, median: medians[index] })))
      .attr("display", showMedian ? null : "none")
      .attr("d", line);

    medianHitTargets
      .selectAll("rect")
      .data(
        points
          .map((point, index) => ({ month: point.month, median: medians[index] }))
          .filter((point): point is { month: string; median: number } => point.median !== null),
        (d) => d.month
      )
      .join("rect")
      .attr("display", showMedian ? null : "none")
      .attr("fill", "transparent")
      .attr("x", (d) => x(d.month) + x.bandwidth() / 2 - 10)
      .attr("y", (d) => y(d.median) - 10)
      .attr("width", 20)
      .attr("height", 20)
      .attr("tabindex", 0)
      .attr("role", "img")
      .attr("aria-label", (d) => `${d.month}: ${formatCurrency(d.median)}`)
      .attr("data-tippy-content", (d) => `${d.month}: ${formatCurrency(d.median)}`);
  };

  const visibleGroups = new Set(groupKeys);
  const legends = _.map(groupKeys, (k) => {
    return {
      label: groupTotal[k],
      color: z(k),
      shape: "square",
      toggle: true,
      selected: true,
      onClick: (legend: Legend) => {
        if (legend.selected) {
          visibleGroups.add(k);
        } else {
          visibleGroups.delete(k);
        }
        render(groupKeys.filter((group) => visibleGroups.has(group)));
      }
    } as Legend;
  });

  legends.push({
    label: "4-month rolling median",
    color: COLORS.danger,
    shape: "line",
    toggle: true,
    selected: true,
    onClick: (legend: Legend) => {
      showMedian = legend.selected;
      render(groupKeys.filter((group) => visibleGroups.has(group)));
    }
  });

  render(groupKeys);

  return legends;
}

function financialYear(card: IncomeYearlyCard) {
  return `${card.start_date.format("YYYY")} - ${card.end_date.format("YY")}`;
}

export function renderYearlyIncomeTimeline(yearlyCards: IncomeYearlyCard[]): Legend[] {
  const id = "#d3-yearly-income-timeline";
  const BAR_HEIGHT = 20;
  const svg = d3.select(id),
    margin = { top: 15, right: 20, bottom: 20, left: 70 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const groups = _.chain(yearlyCards)
    .flatMap((c) => c.postings)
    .map((p) => secondName(p.account))
    .uniq()
    .sort()
    .value();

  const defaultValues = _.zipObject(
    groups,
    _.map(groups, () => 0)
  );

  const start = _.min(_.map(yearlyCards, (c) => c.start_date)),
    end = _.max(_.map(yearlyCards, (c) => c.end_date));

  if (!start || !end) {
    return [];
  }

  const height = BAR_HEIGHT * (end.year() - start.year());
  svg.attr("height", height + margin.top + margin.bottom);

  interface Point {
    year: string;
    [key: string]: number | string | dayjs.Dayjs;
  }
  const points: Point[] = [];

  _.each(yearlyCards, (card) => {
    const postings = card.postings;
    const values = _.chain(postings)
      .groupBy((t) => secondName(t.account))
      .map((postings, key) => [key, _.sum(_.map(postings, (p) => -p.amount))])
      .fromPairs()
      .value();

    points.push(
      _.merge(
        {
          year: financialYear(card),
          postings: postings
        },
        defaultValues,
        values
      )
    );
  });

  const x = d3.scaleLinear().range([0, width]);
  const y = d3.scaleBand().range([height, 0]).paddingInner(0.1).paddingOuter(0);

  y.domain(points.map((p) => p.year));
  x.domain([0, d3.max(points, (p: Point) => _.sum(_.map(groups, (k) => p[k])))]);

  const z = generateColorScheme(groups);

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(d3.axisBottom(x).tickSize(-height).tickFormat(formatCurrencyCrude));

  g.append("g").attr("class", "axis y dark").call(d3.axisLeft(y));

  g.append("g")
    .selectAll("g")
    .data(
      d3.stack().offset(d3.stackOffsetDiverging).keys(groups)(points as { [key: string]: number }[])
    )
    .enter()
    .append("g")
    .attr("fill", function (d) {
      return z(d.key.split("-")[0]);
    })
    .selectAll("rect")
    .data(function (d) {
      return d;
    })
    .enter()
    .append("rect")
    .attr("data-tippy-content", (d) => {
      let grandTotal = 0;
      return tooltip(
        _.sortBy(
          groups.flatMap((k) => {
            const total = d.data[k];
            if (total == 0) {
              return [];
            }
            grandTotal += total;
            return [[k, [formatCurrency(total), "has-text-weight-bold has-text-right"]]];
          }),
          (r) => r[0]
        ),
        { total: formatCurrency(grandTotal) }
      );
    })
    .attr("x", function (d) {
      return x(d[0]);
    })
    .attr("y", function (d) {
      return y((d.data as any).year) + (y.bandwidth() - Math.min(y.bandwidth(), BAR_HEIGHT)) / 2;
    })
    .attr("width", function (d) {
      return x(d[1]) - x(d[0]);
    })
    .attr("height", y.bandwidth());

  return _.map(groups, (k) => ({
    label: k,
    color: z(k),
    shape: "square"
  }));
}

export function renderYearlyTimelineOf(
  label: string,
  key: "net_tax" | "net_income",
  color: string,
  yearlyCards: IncomeYearlyCard[]
): Legend[] {
  const id = `#d3-yearly-${key}-timeline`;
  const BAR_HEIGHT = 20;
  const svg = d3.select(id),
    margin = { top: 15, right: 20, bottom: 20, left: 70 },
    width =
      document.getElementById(id.substring(1)).parentElement.clientWidth -
      margin.left -
      margin.right,
    g = svg.append("g").attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  const colorKeys = [label];
  const colorScale = d3.scaleOrdinal<string>().domain(colorKeys).range([color]);

  const start = _.min(_.map(yearlyCards, (c) => c.start_date)),
    end = _.max(_.map(yearlyCards, (c) => c.end_date));

  if (!start || !end) {
    return [];
  }

  const height = BAR_HEIGHT * (end.year() - start.year());
  svg.attr("height", height + margin.top + margin.bottom);

  interface Point {
    year: string;
    value: number;
  }

  const points: Point[] = _.map(yearlyCards, (card) => {
    return {
      year: financialYear(card),
      value: card[key]
    };
  });

  const x = d3.scaleLinear().range([0, width]);
  const y = d3.scaleBand().range([height, 0]).paddingInner(0.1).paddingOuter(0);

  y.domain(points.map((p) => p.year));
  x.domain([0, d3.max(points, (p: Point) => p.value)]);

  g.append("g")
    .attr("class", "axis y")
    .attr("transform", "translate(0," + height + ")")
    .call(d3.axisBottom(x).tickSize(-height).tickFormat(formatCurrencyCrude));

  g.append("g").attr("class", "axis y dark").call(d3.axisLeft(y));

  g.append("g")
    .selectAll("rect")
    .data(points)
    .join("rect")
    .attr("fill", color)
    .attr("data-tippy-content", (d) => {
      return tooltip([[label, [formatCurrency(d.value), "has-text-weight-bold has-text-right"]]]);
    })
    .attr("x", x(0))
    .attr("y", function (d) {
      return y(d.year) + (y.bandwidth() - Math.min(y.bandwidth(), BAR_HEIGHT)) / 2;
    })
    .attr("width", function (d) {
      return x(d.value) - x(0);
    })
    .attr("height", y.bandwidth());

  return _.map(colorKeys, (k) => ({
    label: k,
    color: colorScale(k),
    shape: "square"
  }));
}

export function renderPassiveIncomeTimeline(report: PassiveIncomeAnalysis): Legend[] {
  const id = "#d3-passive-income-timeline";
  const svg = d3.select(id);
  const element = document.getElementById(id.substring(1));
  if (!element || !element.parentElement || report.timeline.length === 0) {
    return [];
  }

  const margin = { top: 20, right: 30, bottom: 80, left: 55 };
  const width = element.parentElement.clientWidth - margin.left - margin.right;
  const height = +svg.attr("height") - margin.top - margin.bottom;
  const g = svg.append("g").attr("transform", `translate(${margin.left},${margin.top})`);
  const sourceKeys = ["interest", "coupons", "dividends"] as const;
  const sourceLabels: Record<(typeof sourceKeys)[number], string> = {
    interest: "Interest",
    coupons: "Coupons",
    dividends: "Dividends"
  };
  const sourceColors: Record<(typeof sourceKeys)[number], string> = {
    interest: COLORS.primary,
    coupons: COLORS.secondary,
    dividends: COLORS.tertiary
  };
  const points = report.timeline.map((point) => ({
    ...point,
    month: dayjs(point.date).format("MMM-YYYY")
  }));
  const x = d3.scaleBand<string>().range([0, width]).paddingInner(0.12);
  const y = d3.scaleLinear().range([height, 0]);
  const xAxis = g.append("g").attr("class", "axis x");
  const yAxis = g.append("g").attr("class", "axis y");
  const bars = g.append("g");
  const benchmarks = g.append("g");
  const visibleSources = new Set<string>(sourceKeys);
  const visibleBenchmarks = new Set(["regular_personal_burn", "freedom_income_floor"]);
  let firstRender = true;

  const render = () => {
    x.domain(points.map((point) => point.month));
    const actualTotals = points.map((point) =>
      sourceKeys.reduce((total, key) => total + (visibleSources.has(key) ? point[key] : 0), 0)
    );
    const benchmarkValues = [
      visibleBenchmarks.has("regular_personal_burn") ? report.regular_personal_burn : 0,
      visibleBenchmarks.has("freedom_income_floor") ? report.freedom_income_floor : 0
    ];
    y.domain([0, Math.max(1, ...actualTotals, ...benchmarkValues)]);
    const transition = svg.transition().duration(firstRender ? 0 : 350);
    firstRender = false;

    xAxis
      .attr("transform", `translate(0,${height})`)
      .transition(transition)
      .call(
        d3
          .axisBottom(x)
          .ticks(6)
          .tickFormat(skipTicks(30, x, (d) => d.toString()))
      )
      .selectAll("text")
      .attr("y", 10)
      .attr("x", -8)
      .attr("dy", ".35em")
      .attr("transform", "rotate(-45)")
      .style("text-anchor", "end");

    yAxis
      .transition(transition)
      .call(d3.axisLeft(y).tickSize(-width).tickFormat(formatCurrencyCrude));

    bars
      .selectAll("g")
      .data(
        d3.stack().keys([...visibleSources])(points as unknown as Array<Record<string, number>>),
        (d: any) => d.key
      )
      .join(
        (enter) => enter.append("g"),
        (update) => update,
        (exit) => exit.remove()
      )
      .attr("fill", (d: any) => sourceColors[d.key as (typeof sourceKeys)[number]])
      .selectAll("rect")
      .data(
        (d: any) => d,
        (d: any) => d.data.month
      )
      .join("rect")
      .attr("x", (d: any) => x(d.data.month))
      .attr("y", (d: any) => y(d[1]))
      .attr("height", (d: any) => y(d[0]) - y(d[1]))
      .attr("width", x.bandwidth());

    const benchmarkLine = d3
      .line<{ month: string; value: number }>()
      .x((d) => x(d.month) + x.bandwidth() / 2)
      .y((d) => y(d.value));
    benchmarks
      .selectAll("path")
      .data(
        [
          {
            key: "regular_personal_burn",
            value: report.regular_personal_burn,
            color: COLORS.lossText
          },
          {
            key: "freedom_income_floor",
            value: report.freedom_income_floor,
            color: COLORS.gainText
          }
        ].filter((benchmark) => visibleBenchmarks.has(benchmark.key))
      )
      .join("path")
      .attr("fill", "none")
      .attr("stroke", (d) => d.color)
      .attr("stroke-width", 2)
      .attr("stroke-dasharray", "5,5")
      .attr("d", (d) =>
        benchmarkLine(points.map((point) => ({ month: point.month, value: d.value })))
      );
  };

  render();
  return [
    ...sourceKeys.map((key) => ({
      label: sourceLabels[key],
      color: sourceColors[key],
      shape: "square" as const,
      toggle: true,
      selected: true,
      onClick: (legend: Legend) => {
        legend.selected ? visibleSources.add(key) : visibleSources.delete(key);
        render();
      }
    })),
    {
      label: `Regular Personal Burn\n${formatCurrency(report.regular_personal_burn)}`,
      color: COLORS.lossText,
      shape: "line" as const,
      toggle: true,
      selected: true,
      onClick: (legend: Legend) => {
        legend.selected
          ? visibleBenchmarks.add("regular_personal_burn")
          : visibleBenchmarks.delete("regular_personal_burn");
        render();
      }
    },
    {
      label: `Freedom Income Floor\n${formatCurrency(report.freedom_income_floor)}`,
      color: COLORS.gainText,
      shape: "line" as const,
      toggle: true,
      selected: true,
      onClick: (legend: Legend) => {
        legend.selected
          ? visibleBenchmarks.add("freedom_income_floor")
          : visibleBenchmarks.delete("freedom_income_floor");
        render();
      }
    }
  ];
}
