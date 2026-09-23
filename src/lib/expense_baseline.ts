import type {
  ExpenseCategorySignal,
  ExpenseCommitment,
  ExpenseRetrospective
} from "$lib/utils";

export function baselineChartMax(data: ExpenseRetrospective): number {
  return Math.max(
    1,
    ...data.months.flatMap((month) => [month.personal_expense, month.regular_personal_burn])
  );
}

export function baselineBarHeight(value: number, max: number): number {
  return Math.max(4, Math.round((value / max) * 100));
}

export function categorySignalLabel(signal: ExpenseCategorySignal): string {
  if (signal.baseline_median === 0) {
    return "new in the latest month";
  }
  return `+${Math.round(signal.growth_percent)}% vs median`;
}

export function commitmentCadence(commitment: ExpenseCommitment): string {
  if (commitment.interval >= 28 && commitment.interval <= 33) {
    return "monthly";
  }
  if (commitment.interval >= 7 && commitment.interval <= 8) {
    return "weekly";
  }
  return `every ${commitment.interval} days`;
}
