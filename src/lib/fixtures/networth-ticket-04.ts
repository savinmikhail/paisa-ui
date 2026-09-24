import dayjs from "dayjs";
import type { Networth } from "../utils";
import type {
  NetworthDecomposition,
  NetworthReconciliation
} from "../networth_reconciliation";

export type NetworthFixturePoint = Networth & {
  intervalDelta: Omit<NetworthDecomposition, "totalDelta">;
};

const deltas: Array<NetworthFixturePoint["intervalDelta"]> = [
  { netSavings: 0, passiveIncome: 0, market: 0, fx: 0, debt: 0, corrections: 0 },
  { netSavings: 12000, passiveIncome: 900, market: 4500, fx: 900, debt: -3000, corrections: 300 },
  { netSavings: 8000, passiveIncome: 650, market: -1800, fx: 450, debt: 0, corrections: 150 },
  { netSavings: -2200, passiveIncome: 0, market: 3200, fx: -600, debt: -500, corrections: 100 },
  { netSavings: 9500, passiveIncome: 700, market: 2800, fx: 1200, debt: 0, corrections: -200 },
  { netSavings: 4000, passiveIncome: 500, market: -900, fx: 300, debt: 1500, corrections: 0 },
  { netSavings: 11000, passiveIncome: 800, market: 4100, fx: 700, debt: -1000, corrections: 250 },
  { netSavings: 6500, passiveIncome: 450, market: 1600, fx: -250, debt: 0, corrections: 50 }
];

const dates = [
  "2024-01-01",
  "2024-01-31",
  "2024-02-29",
  "2024-03-31",
  "2024-04-30",
  "2024-05-31",
  "2024-06-30",
  "2024-07-31"
];

let runningNetworth = 240000;

export const NETWORTH_TICKET_04_FIXTURE: NetworthFixturePoint[] = dates.map((date, index) => {
  const intervalDelta = deltas[index];
  const totalDelta = Object.values(intervalDelta).reduce((sum, value) => sum + value, 0);
  if (index > 0) {
    runningNetworth += totalDelta;
  }

  return {
    date: dayjs(date),
    investmentAmount: runningNetworth,
    withdrawalAmount: 0,
    gainAmount: 0,
    balanceAmount: runningNetworth,
    balanceUnits: 0,
    netInvestmentAmount: runningNetworth,
    intervalDelta
  };
});

export function ticket04FixtureReconciliation(
  from: dayjs.Dayjs,
  to: dayjs.Dayjs
): NetworthReconciliation {
  const selected = NETWORTH_TICKET_04_FIXTURE.filter(
    (point) => !point.date.isBefore(from, "day") && !point.date.isAfter(to, "day")
  );
  const points = selected.length > 0 ? selected : NETWORTH_TICKET_04_FIXTURE;
  const opening = points[0];
  const closing = points[points.length - 1];
  const decomposition = points.slice(1).reduce(
    (total, point) => {
      for (const key of Object.keys(point.intervalDelta) as Array<keyof typeof total>) {
        total[key] += point.intervalDelta[key];
      }
      return total;
    },
    { netSavings: 0, passiveIncome: 0, market: 0, fx: 0, debt: 0, corrections: 0 }
  );

  return {
    from: opening.date,
    to: closing.date,
    opening,
    closing,
    operations: [],
    decomposition: {
      ...decomposition,
      totalDelta: closing.investmentAmount - opening.investmentAmount
    }
  };
}
