import assert from "node:assert/strict";
import { describe, test } from "node:test";
import { affectedAssets, aggregateExposure, bucketKey, fieldValueLabel } from "./exposure";
import { syntheticExposureReport } from "./exposure.fixture";

const summary = aggregateExposure(syntheticExposureReport);

describe("exposure reporting contract", () => {
  test("aggregates a domain value across affected assets", () => {
    const bucket = summary.buckets.find(
      (item) => item.key === bucketKey("institution", "Northstar Cooperative")
    );

    assert.ok(bucket);
    assert.equal(bucket.asset_count, 3);
    assert.equal(bucket.market_value, 460000);
    assert.equal(bucket.share, 460000 / 700000);
  });

  test("retains cross-domain overlap with stable asset membership", () => {
    const overlap = summary.overlaps.find(
      (item) =>
        item.left_field === "institution" &&
        item.left_value === "Northstar Cooperative" &&
        item.right_field === "custodian" &&
        item.right_value === "Orbit Custody"
    );

    assert.deepEqual(overlap?.asset_ids, ["asset-lumen-notes", "asset-cedar-note"]);
    assert.equal(overlap?.market_value, 250000);
  });

  test("represents missing metadata as an explicit unknown bucket", () => {
    const unknownBroker = summary.buckets.find((item) => item.key === bucketKey("broker", null));

    assert.ok(unknownBroker);
    assert.equal(unknownBroker.label, "Unknown");
    assert.equal(unknownBroker.asset_count, 2);
    assert.equal(unknownBroker.market_value, 280000);
    assert.equal(summary.coverage, 29 / 40);
  });

  test("drills down from an aggregate to the original synthetic assets", () => {
    const assets = affectedAssets(syntheticExposureReport, "jurisdiction", "RU");

    assert.deepEqual(
      assets.map((asset) => asset.id),
      ["asset-aurora-reserve", "asset-lumen-notes", "asset-cedar-note"]
    );
    assert.equal(fieldValueLabel("jurisdiction", null), "Jurisdiction · Unknown");
  });
});
