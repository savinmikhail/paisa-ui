import assert from "node:assert/strict";
import { describe, test } from "node:test";
import {
  baselineBarHeight,
  baselineChartMax,
  categorySignalLabel,
  commitmentCadence
} from "./expense_baseline";
import { loadExpenseBaselineFixture } from "./fixtures/expense-baseline";

describe("expense retrospective fixture", () => {
  test("keeps exclusions separate from the personal baseline", () => {
    const { retrospective } = loadExpenseBaselineFixture();

    assert.deepEqual(
      retrospective.excluded.map((item) => item.label),
      ["Business", "Tax", "Capital Placement"]
    );
    assert.equal(retrospective.personal_expense_median, 100);
    assert.equal(retrospective.regular_personal_burn_median, 100);
    assert.equal(retrospective.one_offs[0].category, "Travel");
    assert.equal(retrospective.one_offs[0].posting.transaction_note, "; burn-class: one-off");
  });

  test("surfaces retrospective signals and recurring commitments", () => {
    const { retrospective } = loadExpenseBaselineFixture();
    const signal = retrospective.category_signals[0];
    const commitment = retrospective.recurring_commitments[0];

    assert.equal(signal.category, "Food");
    assert.equal(categorySignalLabel(signal), "+400% vs median");
    assert.equal(commitmentCadence(commitment), "monthly");
    assert.equal(baselineChartMax(retrospective), 1180);
    assert.equal(baselineBarHeight(590, 1180), 50);
  });
});
