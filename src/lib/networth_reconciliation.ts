import type dayjs from "dayjs";
import type { Networth } from "./utils";

export const NETWORTH_DECOMPOSITION_KEYS = [
  "netSavings",
  "passiveIncome",
  "market",
  "fx",
  "debt",
  "corrections"
] as const;

export type NetworthDecompositionKey = (typeof NETWORTH_DECOMPOSITION_KEYS)[number];

export type NetworthDecomposition = Record<NetworthDecompositionKey, number> & {
  totalDelta: number;
};

export interface NetworthReconciliation {
  from: dayjs.Dayjs;
  to: dayjs.Dayjs;
  opening: Networth;
  closing: Networth;
  decomposition: NetworthDecomposition;
}

export interface NetworthIntervalSummary {
  opening: number;
  closing: number;
  delta: number;
  percentDelta: number | null;
}

export interface NetworthFocus {
  point: Networth;
  series: "networth" | "investment";
}

export interface NetworthInteractionState {
  focus: NetworthFocus | null;
}

export function networthValue(point: Networth): number {
  return point.investmentAmount + point.gainAmount - point.withdrawalAmount;
}

export function netInvestmentValue(point: Networth): number {
  return point.investmentAmount - point.withdrawalAmount;
}

export function summarizeNetworthInterval(points: Networth[]): NetworthIntervalSummary | null {
  if (points.length === 0) {
    return null;
  }

  const opening = networthValue(points[0]);
  const closing = networthValue(points[points.length - 1]);
  const delta = closing - opening;

  return {
    opening,
    closing,
    delta,
    percentDelta: opening === 0 ? null : (delta / Math.abs(opening)) * 100
  };
}

export function sumDecomposition(
  decomposition: Pick<NetworthDecomposition, NetworthDecompositionKey>
): number {
  return NETWORTH_DECOMPOSITION_KEYS.reduce((total, key) => total + decomposition[key], 0);
}

export function reduceNetworthInteraction(
  _state: NetworthInteractionState,
  event: { type: "focus"; focus: NetworthFocus } | { type: "blur" }
): NetworthInteractionState {
  return event.type === "focus" ? { focus: event.focus } : { focus: null };
}

export function dateInputValue(date: dayjs.Dayjs): string {
  return date.format("YYYY-MM-DD");
}
