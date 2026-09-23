import type { MoneyFlowSource } from "./money_flow";

export type MoneyFlowSourceKind = "private" | "synthetic";
export type MoneyFlowRequest = (route: string) => Promise<unknown>;

export function moneyFlowRoute(source: MoneyFlowSourceKind = "private"): string {
  return source === "synthetic" ? "/api/money_flow?source=synthetic" : "/api/money_flow";
}

function isMoneyFlowSource(value: unknown): value is MoneyFlowSource {
  if (!value || typeof value !== "object") return false;
  const report = value as Partial<MoneyFlowSource>;
  return (
    report.contractVersion === "money-flow.v1" &&
    (report.source === "private-adapter" || report.source === "synthetic-fixture") &&
    typeof report.baseCurrency === "string" &&
    Array.isArray(report.flowEvents) &&
    Array.isArray(report.waterfallEvents) &&
    Array.isArray(report.netWorthSnapshots)
  );
}

export async function loadMoneyFlow(
  source: MoneyFlowSourceKind = "private",
  request: MoneyFlowRequest
): Promise<MoneyFlowSource> {
  const report = await request(moneyFlowRoute(source));
  if (!isMoneyFlowSource(report)) {
    throw new Error("Money Flow API returned an invalid money-flow.v1 report");
  }
  return report;
}
