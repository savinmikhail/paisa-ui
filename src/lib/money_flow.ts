import * as d3 from "d3";

export const MONEY_FLOW_CATEGORIES = [
  "Personal Expense",
  "Business",
  "Tax",
  "Reserve",
  "Capital Placement"
] as const;

export type MoneyFlowCategory = (typeof MONEY_FLOW_CATEGORIES)[number];

export interface MoneyFlowEvent {
  date: string;
  category: "Income" | MoneyFlowCategory | "Internal Transfers";
  amount: number;
  currency: string;
  label?: string;
  unresolved?: boolean;
}

export interface MoneyFlowWaterfallEvent {
  date: string;
  key: "netSavings" | "passiveIncome" | "market" | "fx" | "debt" | "corrections";
  amount: number;
}

export interface MoneyFlowSnapshot {
  date: string;
  value: number;
}

export interface MoneyFlowSource {
  contractVersion: "money-flow.v1";
  baseCurrency: string;
  source: "private-adapter" | "synthetic-fixture";
  flowEvents: MoneyFlowEvent[];
  waterfallEvents: MoneyFlowWaterfallEvent[];
  netWorthSnapshots: MoneyFlowSnapshot[];
}

export interface MoneyFlowFilters {
  from: string;
  to: string;
  currency: "ALL" | string;
  categories: readonly MoneyFlowCategory[];
  showUnresolvedTransfers: boolean;
}

export interface MoneyFlowSankeyNode {
  key: string;
  label: string;
  kind: "source" | "destination";
  value: number;
}

export interface MoneyFlowSankeyLink {
  source: string;
  target: string;
  value: number;
}

export interface MoneyFlowSankeyView {
  nodes: MoneyFlowSankeyNode[];
  links: MoneyFlowSankeyLink[];
  incomeTotal: number;
  transferTotal: number;
}

export interface MoneyFlowWaterfallStep {
  key: MoneyFlowWaterfallEvent["key"];
  label: string;
  amount: number;
}

export interface MoneyFlowWaterfallView {
  opening: number;
  closing: number;
  steps: MoneyFlowWaterfallStep[];
}

export interface MoneyFlowView {
  sankey: MoneyFlowSankeyView;
  waterfall: MoneyFlowWaterfallView;
}

const waterfallLabels: Record<MoneyFlowWaterfallEvent["key"], string> = {
  netSavings: "Savings",
  passiveIncome: "Realized passive income",
  market: "Market revaluation",
  fx: "FX revaluation",
  debt: "Debt changes",
  corrections: "Corrections"
};

function formatAmount(value: number): string {
  return new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 }).format(value);
}

export function defaultMoneyFlowFilters(source: MoneyFlowSource): MoneyFlowFilters {
  const dates = source.flowEvents.map((event) => event.date).concat(source.netWorthSnapshots.map((point) => point.date));
  return {
    from: dates.slice().sort()[0] || "",
    to: dates.slice().sort().at(-1) || "",
    currency: "ALL",
    categories: [...MONEY_FLOW_CATEGORIES],
    showUnresolvedTransfers: false
  };
}

function inRange(date: string, filters: MoneyFlowFilters): boolean {
  return (!filters.from || date >= filters.from) && (!filters.to || date <= filters.to);
}

export function buildMoneyFlowView(source: MoneyFlowSource, filters: MoneyFlowFilters): MoneyFlowView {
  const events = source.flowEvents.filter(
    (event) =>
      inRange(event.date, filters) &&
      (filters.currency === "ALL" || event.currency === filters.currency) &&
      (event.category === "Income" || event.category === "Internal Transfers" || filters.categories.includes(event.category))
  );
  const incomeTotal = events
    .filter((event) => event.category === "Income")
    .reduce((total, event) => total + event.amount, 0);
  const destinations: Array<{ category: string; value: number }> = MONEY_FLOW_CATEGORIES.map((category) => ({
    category,
    value: events
      .filter((event) => event.category === category)
      .reduce((total, event) => total + event.amount, 0)
  })).filter((item) => item.value > 0);
  const unresolved = filters.showUnresolvedTransfers
    ? events
        .filter((event) => event.category === "Internal Transfers" && event.unresolved)
        .reduce((total, event) => total + event.amount, 0)
    : 0;
  if (unresolved > 0) destinations.push({ category: "Unresolved transfers", value: unresolved });

  const nodes: MoneyFlowSankeyNode[] = [
    { key: "income", label: "Income", kind: "source", value: incomeTotal },
    ...destinations.map((item) => ({
      key: item.category,
      label: item.category,
      kind: "destination" as const,
      value: item.value
    }))
  ];
  return {
    sankey: {
      nodes,
      links: destinations.map((item) => ({ source: "income", target: item.category, value: item.value })),
      incomeTotal,
      transferTotal: unresolved
    },
    waterfall: buildWaterfallView(source, filters)
  };
}

export function buildWaterfallView(source: MoneyFlowSource, filters: MoneyFlowFilters): MoneyFlowWaterfallView {
  const snapshots = source.netWorthSnapshots.filter((point) => inRange(point.date, filters));
  const opening = snapshots[0]?.value ?? source.netWorthSnapshots[0]?.value ?? 0;
  const closing = snapshots.at(-1)?.value ?? source.netWorthSnapshots.at(-1)?.value ?? opening;
  const selectedEvents = source.waterfallEvents.filter((event) => inRange(event.date, filters));
  const steps = selectedEvents.map((event) => ({ key: event.key, label: waterfallLabels[event.key], amount: event.amount }));
  return { opening, closing, steps };
}

function svgText(selection: d3.Selection<SVGTextElement, unknown, null, undefined>, text: string) {
  selection.attr("class", "money-flow-svg-label").text(text);
}

export function renderMoneyFlowSankey(element: SVGSVGElement, view: MoneyFlowSankeyView) {
  const root = d3.select(element);
  root.selectAll("*").remove();
  const width = Math.max(element.parentElement?.clientWidth || 760, 680);
  const height = 410;
  root.attr("viewBox", `0 0 ${width} ${height}`).attr("width", width).attr("height", height);
  const max = Math.max(view.incomeTotal, ...view.nodes.map((node) => node.value), 1);
  const scale = d3.scaleLinear().domain([0, max]).range([0, 270]);
  const sourceX = 90;
  const targetX = width - 240;
  const sourceY = 70;
  const sourceHeight = Math.max(scale(view.incomeTotal), 8);
  const destinations = view.nodes.filter((node) => node.kind === "destination");
  const gap = 18;
  const totalHeight = destinations.reduce((sum, node) => sum + Math.max(scale(node.value), 8), 0) + gap * Math.max(destinations.length - 1, 0);
  let y = Math.max(55, (height - totalHeight) / 2);
  const positions = new Map<string, { y: number; height: number }>();
  destinations.forEach((node) => {
    const nodeHeight = Math.max(scale(node.value), 8);
    positions.set(node.key, { y, height: nodeHeight });
    y += nodeHeight + gap;
  });
  const source = root.append("g").attr("class", "money-flow-sankey-source");
  source.append("rect").attr("x", sourceX).attr("y", sourceY).attr("width", 22).attr("height", sourceHeight).attr("rx", 4).attr("fill", "var(--bulma-link)");
  svgText(source.append("text").attr("x", sourceX - 10).attr("y", sourceY - 12).attr("text-anchor", "end"), `Income ${formatAmount(view.incomeTotal)}`);
  let sourceOffset = sourceY;
  view.links.forEach((link) => {
    const target = positions.get(link.target);
    if (!target) return;
    const thickness = Math.max(scale(link.value), 8);
    const sy = sourceOffset + thickness / 2;
    const ty = target.y + target.height / 2;
    const path = `M ${sourceX + 22} ${sy} C ${sourceX + 150} ${sy}, ${targetX - 100} ${ty}, ${targetX} ${ty}`;
    root.append("path").attr("d", path).attr("fill", "none").attr("stroke", "var(--bulma-link)").attr("stroke-opacity", 0.25).attr("stroke-width", thickness);
    sourceOffset += thickness;
  });
  destinations.forEach((node) => {
    const position = positions.get(node.key);
    if (!position) return;
    const group = root.append("g");
    group.append("rect").attr("x", targetX).attr("y", position.y).attr("width", 22).attr("height", position.height).attr("rx", 4).attr("fill", node.key === "Unresolved transfers" ? "var(--bulma-warning)" : "var(--bulma-primary)");
    svgText(group.append("text").attr("x", targetX + 36).attr("y", position.y + position.height / 2 + 4), `${node.label} ${formatAmount(node.value)}`);
  });
  svgText(root.append("text").attr("x", sourceX).attr("y", height - 22), "Recognized flows only · transfers stay outside expenses");
}

export function renderMoneyFlowWaterfall(element: SVGSVGElement, view: MoneyFlowWaterfallView) {
  const root = d3.select(element);
  root.selectAll("*").remove();
  const width = Math.max(element.parentElement?.clientWidth || 760, 680);
  const height = 410;
  root.attr("viewBox", `0 0 ${width} ${height}`).attr("width", width).attr("height", height);
  const values = [view.opening, ...view.steps.map((step) => step.amount), view.closing];
  const max = Math.max(...values.map((value) => Math.abs(value)), view.opening, view.closing, 1);
  const y = d3.scaleLinear().domain([-max * 0.1, max * 1.1]).range([height - 55, 50]);
  const labels = ["Opening Net Worth", ...view.steps.map((step) => step.label), "Closing Net Worth"];
  const x = d3.scaleBand().domain(labels).range([55, width - 30]).padding(0.28);
  let running = view.opening;
  const bars = [{ label: labels[0], start: 0, end: running, color: "var(--bulma-link)" }];
  view.steps.forEach((step, index) => {
    const next = running + step.amount;
    bars.push({ label: labels[index + 1], start: Math.min(running, next), end: Math.max(running, next), color: step.amount >= 0 ? "var(--bulma-success)" : "var(--bulma-danger)" });
    running = next;
  });
  bars.push({ label: labels.at(-1) || "Closing Net Worth", start: 0, end: view.closing, color: "var(--bulma-primary)" });
  bars.forEach((bar) => {
    const position = x(bar.label) || 0;
    const top = y(bar.end);
    const bottom = y(bar.start);
    root.append("rect").attr("x", position).attr("y", Math.min(top, bottom)).attr("width", x.bandwidth()).attr("height", Math.max(3, Math.abs(bottom - top))).attr("rx", 3).attr("fill", bar.color).attr("opacity", 0.85);
    svgText(root.append("text").attr("x", position + x.bandwidth() / 2).attr("y", Math.min(top, bottom) - 9).attr("text-anchor", "middle"), formatAmount(bar.end - bar.start));
    svgText(root.append("text").attr("x", position + x.bandwidth() / 2).attr("y", height - 20).attr("text-anchor", "middle"), bar.label);
  });
  root.append("line").attr("x1", 45).attr("x2", width - 25).attr("y1", y(0)).attr("y2", y(0)).attr("stroke", "currentColor").attr("stroke-opacity", 0.16);
}
