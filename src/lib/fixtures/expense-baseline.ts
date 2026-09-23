import dayjs from "dayjs";
import rawFixture from "./expense-baseline.json";
import type { ExpenseResponse, Posting } from "$lib/utils";

function toPosting(raw: any): Posting {
  return { ...raw, date: dayjs(raw.date) } as Posting;
}

function groupByMonth(postings: Posting[]) {
  return postings.reduce<Record<string, Posting[]>>((groups, posting) => {
    const key = posting.date.format("YYYY-MM");
    groups[key] = [...(groups[key] || []), posting];
    return groups;
  }, {});
}

function groupByYear(postings: Posting[]) {
  return postings.reduce<Record<string, Posting[]>>((groups, posting) => {
    const key = posting.date.format("YYYY");
    groups[key] = [...(groups[key] || []), posting];
    return groups;
  }, {});
}

export function loadExpenseBaselineFixture(): ExpenseResponse {
  const expenses = rawFixture.expenses.map(toPosting);
  const retrospective = {
    ...rawFixture.retrospective,
    start_date: dayjs(rawFixture.retrospective.start_date),
    end_date: dayjs(rawFixture.retrospective.end_date),
    months: rawFixture.retrospective.months.map((month) => ({
      ...month,
      date: dayjs(month.date)
    })),
    one_offs: rawFixture.retrospective.one_offs.map((oneOff) => ({
      ...oneOff,
      posting: toPosting(oneOff.posting)
    })),
    recurring_commitments: rawFixture.retrospective.recurring_commitments.map((commitment) => ({
      ...commitment,
      last_date: dayjs(commitment.last_date)
    }))
  };

  return {
    expenses,
    month_wise: {
      expenses: groupByMonth(expenses),
      incomes: {},
      investments: {},
      taxes: {}
    },
    year_wise: {
      expenses: groupByYear(expenses),
      incomes: {},
      investments: {},
      taxes: {}
    },
    graph: {},
    retrospective
  } as ExpenseResponse;
}
