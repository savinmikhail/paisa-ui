import dayjs, { type Dayjs } from "dayjs";

export interface ResiliencePoint {
  date: Dayjs;
  rub: number;
  usd: number;
  life_months: number;
}

export interface ResilienceScopeItem {
  name: string;
  kind: string;
  currency: string;
  reason: string;
}

export interface ResilienceReport {
  contract_version: string;
  source: "synthetic-fixture" | "private-adapter" | string;
  as_of: string;
  base_currency: string;
  range_start: Dayjs;
  range_end: Dayjs;
  liquid_scope: {
    included: ResilienceScopeItem[];
    excluded: ResilienceScopeItem[];
  };
  burn_provenance: {
    metric: "RegularPersonalBurnMedian" | string;
    source: "ExpenseBaseline.RegularPersonalBurnMedian" | string;
    value: number;
    currency: string;
    period_start: Dayjs;
    period_end: Dayjs;
    description: string;
  };
  points: ResiliencePoint[];
}

export interface ResilienceReportInput
  extends Omit<ResilienceReport, "range_start" | "range_end" | "burn_provenance" | "points"> {
  range_start: string | Dayjs;
  range_end: string | Dayjs;
  burn_provenance: Omit<ResilienceReport["burn_provenance"], "period_start" | "period_end"> & {
    period_start: string | Dayjs;
    period_end: string | Dayjs;
  };
  points: Array<Omit<ResiliencePoint, "date"> & { date: string | Dayjs }>;
}

export type ResilienceMetric = "rub" | "usd" | "life_months";

export const RESILIENCE_METRICS: Array<{
  key: ResilienceMetric;
  label: string;
  unit: string;
  color: string;
}> = [
  { key: "rub", label: "Liquid RUB", unit: "RUB", color: "#3273dc" },
  { key: "usd", label: "Liquid USD", unit: "USD", color: "#48c78e" },
  { key: "life_months", label: "Life Months", unit: "months", color: "#f14668" }
];

export function normalizeResilienceReport(raw: ResilienceReportInput): ResilienceReport {
  return {
    ...raw,
    range_start: dayjs(raw.range_start),
    range_end: dayjs(raw.range_end),
    burn_provenance: {
      ...raw.burn_provenance,
      period_start: dayjs(raw.burn_provenance.period_start),
      period_end: dayjs(raw.burn_provenance.period_end)
    },
    points: raw.points.map((point) => ({ ...point, date: dayjs(point.date) }))
  };
}

export function filterResiliencePoints(
  points: ResiliencePoint[],
  from: string | Dayjs,
  to: string | Dayjs
): ResiliencePoint[] {
  const start = dayjs(from);
  const end = dayjs(to);
  return points.filter((point) => !point.date.isBefore(start, "day") && !point.date.isAfter(end, "day"));
}

export function resilienceDomain(points: ResiliencePoint[], metric: ResilienceMetric): [number, number] {
  const max = Math.max(1, ...points.map((point) => point[metric]));
  return [0, max];
}

export function resilienceX(index: number, count: number, width: number): number {
  return count <= 1 ? width / 2 : (index / (count - 1)) * width;
}

export function resilienceY(value: number, domain: [number, number], height: number): number {
  const span = domain[1] - domain[0];
  return span === 0 ? height : height - ((value - domain[0]) / span) * height;
}

export function resilienceDateInputValue(date: Dayjs): string {
  return date.format("YYYY-MM-DD");
}
