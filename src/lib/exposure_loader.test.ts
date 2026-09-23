import assert from "node:assert/strict";
import { describe, test } from "node:test";
import { syntheticExposureReport } from "./exposure.fixture";
import { exposureRoute, loadExposureReport } from "./exposure_loader";

describe("exposure API loader", () => {
  test("selects the read-only private API by default", async () => {
    const routes: string[] = [];
    const report = await loadExposureReport("private", async (route) => {
      routes.push(route);
      return syntheticExposureReport;
    });

    assert.deepEqual(routes, ["/api/exposure"]);
    assert.equal(report.contract_version, "exposure.v1");
  });

  test("selects synthetic API source explicitly for deterministic tests", async () => {
    const routes: string[] = [];
    const report = await loadExposureReport("synthetic", async (route) => {
      routes.push(route);
      return { ...syntheticExposureReport, source: "synthetic-fixture" };
    });

    assert.equal(exposureRoute("synthetic"), "/api/exposure?source=synthetic");
    assert.deepEqual(routes, ["/api/exposure?source=synthetic"]);
    assert.equal(report.assets[0].failure_domains.broker, undefined);
  });

  test("fails closed on an invalid API payload", async () => {
    await assert.rejects(
      () => loadExposureReport("private", async () => ({ source: "private-adapter" })),
      /invalid exposure\.v1 report/
    );
  });
});
