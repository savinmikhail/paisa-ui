import assert from "node:assert/strict";
import { describe, test } from "node:test";
import { MONEY_FLOW_SYNTHETIC_FIXTURE } from "../src/lib/fixtures/money-flow";
import { buildMoneyFlowView, type MoneyFlowFilters } from "../src/lib/money_flow";
import { loadMoneyFlow, moneyFlowRoute } from "../src/lib/money_flow_loader";

describe("Money Flow contract", () => {
  test("keeps unresolved transfers outside the default Sankey", () => {
    const filters: MoneyFlowFilters = {
      from: "2024-01-01",
      to: "2024-03-31",
      currency: "ALL",
      categories: ["Personal Expense", "Business", "Tax", "Reserve", "Capital Placement"],
      showUnresolvedTransfers: false
    };
    const hidden = buildMoneyFlowView(MONEY_FLOW_SYNTHETIC_FIXTURE, filters);
    const shown = buildMoneyFlowView(MONEY_FLOW_SYNTHETIC_FIXTURE, { ...filters, showUnresolvedTransfers: true });

    assert.equal(hidden.sankey.transferTotal, 0);
    assert.equal(shown.sankey.transferTotal, 11000);
    assert.equal(shown.sankey.nodes.at(-1)?.label, "Unresolved transfers");
  });

  test("reconciles waterfall steps to the synthetic closing value", () => {
    const view = buildMoneyFlowView(MONEY_FLOW_SYNTHETIC_FIXTURE, {
      from: "2023-12-31",
      to: "2024-03-31",
      currency: "ALL",
      categories: ["Personal Expense", "Business", "Tax", "Reserve", "Capital Placement"],
      showUnresolvedTransfers: false
    });
    const delta = view.waterfall.steps.reduce((sum, step) => sum + step.amount, 0);
    assert.equal(view.waterfall.opening + delta, view.waterfall.closing);
  });

  test("uses an explicit synthetic route and rejects invalid payloads", async () => {
    const routes: string[] = [];
    const report = await loadMoneyFlow("synthetic", async (route) => {
      routes.push(route);
      return MONEY_FLOW_SYNTHETIC_FIXTURE;
    });
    assert.equal(moneyFlowRoute("synthetic"), "/api/money_flow?source=synthetic");
    assert.deepEqual(routes, ["/api/money_flow?source=synthetic"]);
    assert.equal(report.source, "synthetic-fixture");
    await assert.rejects(() => loadMoneyFlow("private", async () => ({ source: "private-adapter" })), /invalid money-flow/);
  });
});
