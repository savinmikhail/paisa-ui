import assert from "node:assert/strict";
import { describe, test } from "node:test";
import { loadResilienceFixture } from "./fixtures/resilience";
import {
  filterResiliencePoints,
  resilienceDomain,
  resilienceY,
  type ResilienceMetric
} from "./resilience";

describe("resilience fixture", () => {
  test("keeps arbitrary date ranges synchronized across all metrics", () => {
    const report = loadResilienceFixture();
    const visible = filterResiliencePoints(report.points, "2025-02-01", "2025-05-01");

    assert.deepEqual(
      visible.map((point) => point.date.format("YYYY-MM-DD")),
      ["2025-02-01", "2025-03-01", "2025-04-01", "2025-05-01"]
    );
    for (const metric of ["rub", "usd", "life_months"] as ResilienceMetric[]) {
      assert.equal(visible.length, visible.map((point) => point[metric]).length);
    }
  });

  test("uses independent domains and the ExpenseBaseline burn provenance", () => {
    const report = loadResilienceFixture();
    assert.equal(report.burn_provenance.source, "ExpenseBaseline.RegularPersonalBurnMedian");
    assert.equal(report.burn_provenance.value, 30000);
    assert.equal(report.liquid_scope.included.at(-1)?.name, "Reserve Asset USDT");
    assert.deepEqual(
      report.liquid_scope.excluded.map((item) => item.name),
      ["Investments", "IIS", "Receivables"]
    );

    const rubDomain = resilienceDomain(report.points, "rub");
    const usdDomain = resilienceDomain(report.points, "usd");
    const lifeDomain = resilienceDomain(report.points, "life_months");
    assert.notDeepEqual(rubDomain, usdDomain);
    assert.notDeepEqual(usdDomain, lifeDomain);
    assert.equal(resilienceY(lifeDomain[1], lifeDomain, 100), 0);
  });
});
