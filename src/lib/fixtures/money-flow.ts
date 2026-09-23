import type { MoneyFlowSource } from "../money_flow";

export const MONEY_FLOW_SYNTHETIC_FIXTURE: MoneyFlowSource = {
  contractVersion: "money-flow.v1",
  baseCurrency: "RUB",
  source: "synthetic-fixture",
  flowEvents: [
    { date: "2024-01-31", category: "Income", amount: 176000, currency: "RUB", label: "Salary" },
    { date: "2024-02-29", category: "Income", amount: 92000, currency: "RUB", label: "Contract" },
    { date: "2024-03-31", category: "Income", amount: 138000, currency: "RUB", label: "Salary" },
    { date: "2024-01-31", category: "Personal Expense", amount: 70000, currency: "RUB" },
    { date: "2024-02-29", category: "Personal Expense", amount: 58000, currency: "RUB" },
    { date: "2024-03-31", category: "Business", amount: 26000, currency: "RUB" },
    { date: "2024-02-29", category: "Tax", amount: 42000, currency: "RUB" },
    { date: "2024-01-31", category: "Reserve", amount: 64000, currency: "RUB" },
    { date: "2024-03-31", category: "Capital Placement", amount: 86000, currency: "RUB" },
    { date: "2024-02-29", category: "Internal Transfers", amount: 24000, currency: "RUB" },
    { date: "2024-03-31", category: "Internal Transfers", amount: 11000, currency: "RUB", unresolved: true }
  ],
  waterfallEvents: [
    { date: "2024-03-31", key: "netSavings", amount: 82000 },
    { date: "2024-03-31", key: "passiveIncome", amount: 14000 },
    { date: "2024-03-31", key: "market", amount: 26000 },
    { date: "2024-03-31", key: "fx", amount: -9000 },
    { date: "2024-03-31", key: "debt", amount: -18000 },
    { date: "2024-03-31", key: "corrections", amount: 5000 }
  ],
  netWorthSnapshots: [
    { date: "2023-12-31", value: 640000 },
    { date: "2024-03-31", value: 740000 }
  ]
};
