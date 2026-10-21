import type { Income, Posting } from "./utils";

export function incomeGroup(posting: Pick<Posting, "account">): string {
  return posting.account.split(":")[1] || posting.account;
}

export function monthIncomeTotal(
  postings: Pick<Posting, "account" | "amount">[],
  visibleGroups: Set<string> | string[]
): number {
  const groups = visibleGroups instanceof Set ? visibleGroups : new Set(visibleGroups);
  return postings.reduce((total, posting) => {
    return groups.has(incomeGroup(posting)) ? total - posting.amount : total;
  }, 0);
}

export function selectedIncomeTotals(incomes: Income[], visibleGroups: string[]): number[] {
  return incomes.map((income) => monthIncomeTotal(income.postings, visibleGroups));
}

export function rollingMedian(values: number[], windowSize = 4): Array<number | null> {
  if (windowSize < 1) {
    throw new Error("rolling median window must be positive");
  }

  return values.map((_value, index) => {
    if (index + 1 < windowSize) {
      return null;
    }

    const window = values.slice(index - windowSize + 1, index + 1).sort((a, b) => a - b);
    const middle = Math.floor(window.length / 2);
    if (window.length % 2 === 0) {
      return (window[middle - 1] + window[middle]) / 2;
    }
    return window[middle];
  });
}

// Cash Flow reports tax separately from other recognized expenses.
export function savingsRate(flow: {
  income: number;
  expenses: number;
  tax: number;
}): number | null {
  return flow.income > 0 ? (flow.income - flow.expenses - flow.tax) / flow.income : null;
}
