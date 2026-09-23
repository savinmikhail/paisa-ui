import assert from "node:assert/strict";
import { describe, test } from "node:test";
import dayjs from "dayjs";
import type { Networth } from "./utils";
import {
  dateInputValue,
  netInvestmentValue,
  networthValue,
  reduceNetworthInteraction,
  sumDecomposition,
  summarizeNetworthInterval
} from "./networth_reconciliation";

function point(date: string, investmentAmount: number, gainAmount: number): Networth {
  return {
    date: dayjs(date),
    investmentAmount,
    withdrawalAmount: 0,
    gainAmount,
    balanceAmount: investmentAmount + gainAmount,
    balanceUnits: 0,
    netInvestmentAmount: investmentAmount
  };
}

describe("net worth interval interaction model", () => {
  test("calculates opening, closing, absolute and percent delta", () => {
    const points = [point("2024-01-01", 1000, 0), point("2024-04-01", 1000, 250)];

    assert.equal(networthValue(points[1]), 1250);
    assert.equal(netInvestmentValue(points[1]), 1000);
    assert.deepEqual(summarizeNetworthInterval(points), {
      opening: 1000,
      closing: 1250,
      delta: 250,
      percentDelta: 25
    });
  });

  test("keeps zero opening percentage explicit instead of emitting infinity", () => {
    assert.equal(summarizeNetworthInterval([point("2024-01-01", 0, 0), point("2024-02-01", 50, 0)])
      ?.percentDelta, null);
  });

  test("keeps the six-part decomposition reconciled", () => {
    const decomposition = {
      netSavings: 120,
      passiveIncome: 10,
      market: 35,
      fx: -5,
      debt: -20,
      corrections: 2,
      totalDelta: 142
    };

    assert.equal(sumDecomposition(decomposition), decomposition.totalDelta);
  });

  test("synchronizes hover focus and clears it on chart leave", () => {
    const focused = reduceNetworthInteraction(
      { focus: null },
      { type: "focus", focus: { point: point("2024-03-01", 1100, 40), series: "networth" } }
    );

    assert.equal(focused.focus?.point.date.format("YYYY-MM-DD"), "2024-03-01");
    assert.equal(reduceNetworthInteraction(focused, { type: "blur" }).focus, null);
  });

  test("serializes dates for deterministic screenshot inputs", () => {
    assert.equal(dateInputValue(dayjs("2024-02-15")), "2024-02-15");
  });
});
