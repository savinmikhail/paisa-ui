import assert from "node:assert/strict";
import { describe, test } from "node:test";
import { incomeTimelineDomain } from "./income_scale";

describe("income timeline", () => {
  test("scales to the visible income groups", () => {
    const points = [
      { Salary: 100000, Interest: 1000 },
      { Salary: 80000, Interest: 1500 }
    ];

    assert.deepEqual(incomeTimelineDomain(points, ["Salary", "Interest"]), [0, 101000]);
    assert.deepEqual(incomeTimelineDomain(points, ["Interest"]), [0, 1500]);
  });

  test("keeps a usable domain when every group is hidden", () => {
    assert.deepEqual(incomeTimelineDomain([{ Salary: 100000 }], []), [0, 1]);
  });
});
