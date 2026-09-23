import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { test } from "node:test";
import { incomeGroup, rollingMedian, selectedIncomeTotals } from "../src/lib/income_reporting";

const fixture = JSON.parse(
  fs.readFileSync(path.join(process.cwd(), "tests/fixtures/income-screen.synthetic.json"), "utf8")
);

test("income fixture exposes actuals and a four-month rolling median", () => {
  const incomes = fixture.income_timeline.map((income: any) => ({
    date: income.date,
    postings: income.postings
  }));
  const totals = selectedIncomeTotals(incomes, ["Salary"]);

  assert.deepEqual(totals, [1200, 800, 2400, 1000, 1400, 1100]);
  assert.deepEqual(rollingMedian(totals), [null, null, null, 1100, 1200, 1250]);
  assert.equal(incomeGroup({ account: "Income:Coupon:Bond" }), "Coupon");
});

test("legend filtering changes the median input, while passive fixture keeps FX out", () => {
  const incomes = fixture.income_timeline.map((income: any) => ({
    date: income.date,
    postings: income.postings
  }));
  const allTotals = selectedIncomeTotals(incomes, [
    "Salary",
    "Interest",
    "Coupon",
    "Dividend",
    "CapitalGains"
  ]);
  const salaryTotals = selectedIncomeTotals(incomes, ["Salary"]);

  assert.notDeepEqual(rollingMedian(allTotals), rollingMedian(salaryTotals));
  const passive = fixture.income_analysis.passive_income;
  assert.equal(
    passive.timeline.reduce((sum: number, point: any) => sum + point.total, 0),
    250
  );
  assert.equal(passive.timeline[5].total, 35);
  assert.equal(
    passive.timeline.some((point: any) => point.total === 999),
    false
  );
});
