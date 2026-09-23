export type DecisionHistoryFieldStatus = "known" | "partial" | "unknown";

export interface DecisionHistoryField {
  status: DecisionHistoryFieldStatus;
  value: string;
  detail: string;
}

export interface DecisionHistoryCase {
  id: string;
  title: string;
  period: string;
  decision: string;
  context: string;
  amount: {
    label: string;
    currency: string;
  };
  outcome: string;
  currentStatus: string;
  sourceRefs: string[];
  caveats: string[];
  attribution: {
    fx: DecisionHistoryField;
    fees: DecisionHistoryField;
    taxes: DecisionHistoryField;
    rubCounterfactual: DecisionHistoryField;
  };
}

export interface PrivateDecisionHistoryPayload {
  schemaVersion: "decision-history/v1";
  cases: DecisionHistoryCase[];
}

export type DecisionHistoryLoadResult =
  | { kind: "ready"; payload: PrivateDecisionHistoryPayload }
  | { kind: "unavailable"; reason: string };

export const syntheticDecisionHistory: PrivateDecisionHistoryPayload = {
  schemaVersion: "decision-history/v1",
  cases: [
    {
      id: "aldar-broker-account",
      title: "Aldar broker account",
      period: "2022-04 → 2022-12",
      decision: "Open a broker account to add AED-denominated exposure.",
      context:
        "Diversification outside domestic banking rails, with a separate check on custody, access, and liquidity.",
      amount: { label: "420k", currency: "RUB equivalent" },
      outcome:
        "The account remained part of the allocation map; the original diversification thesis was only partly observed.",
      currentStatus: "Monitoring",
      sourceRefs: ["DH-SYN-001 decision note", "DH-SYN-001 account summary"],
      caveats: [
        "No verified total-return series is attached to the public contract.",
        "Statement-level fees and taxes are still outside this fixture."
      ],
      attribution: {
        fx: {
          status: "partial",
          value: "AED / RUB movement separated only directionally",
          detail: "Exact conversion path is reserved for the private adapter."
        },
        fees: {
          status: "unknown",
          value: "Not supplied",
          detail: "Broker and transfer fees need source evidence."
        },
        taxes: {
          status: "unknown",
          value: "Not supplied",
          detail: "Tax treatment is not inferred from the decision record."
        },
        rubCounterfactual: {
          status: "unknown",
          value: "Not calculated",
          detail: "Requires a selected RUB alternative and a fixed comparison horizon."
        }
      }
    },
    {
      id: "iis-contribution",
      title: "IIS",
      period: "2021-01 → 2024-12",
      decision: "Use a locked investment account for a multi-year allocation.",
      context:
        "Trade near-term access for a tax-advantaged wrapper and a longer horizon for bonds and equities.",
      amount: { label: "650k", currency: "RUB contributed" },
      outcome:
        "The contribution window is recorded; withdrawal eligibility and the tax effect remain separate questions.",
      currentStatus: "Locked / review due",
      sourceRefs: ["DH-SYN-002 decision note", "DH-SYN-002 contribution ledger"],
      caveats: [
        "The public fixture does not assert eligibility for any tax deduction.",
        "Current account availability is supplied as a status, not a conclusion."
      ],
      attribution: {
        fx: {
          status: "known",
          value: "RUB-denominated decision",
          detail: "No FX attribution is needed for the synthetic base amount."
        },
        fees: {
          status: "partial",
          value: "Account fees excluded",
          detail: "Instrument-level costs are not part of this contract example."
        },
        taxes: {
          status: "partial",
          value: "Wrapper treatment noted, result unknown",
          detail: "A private adapter must attach verified tax evidence."
        },
        rubCounterfactual: {
          status: "unknown",
          value: "Not calculated",
          detail: "A comparable liquid RUB allocation has not been selected."
        }
      }
    },
    {
      id: "large-ofz-purchase",
      title: "Large OFZ purchase",
      period: "2023-10-06",
      decision: "Make a concentrated OFZ purchase during a high-rate window.",
      context:
        "Prefer visible coupon cash flow and maturity visibility over adding another equity position.",
      amount: { label: "1.8m", currency: "RUB purchase" },
      outcome:
        "Cash-flow and mark-to-market paths diverged; the realized result needs instrument-level reconciliation.",
      currentStatus: "Needs reconciliation",
      sourceRefs: ["DH-SYN-003 decision note", "DH-SYN-003 bond lot summary"],
      caveats: [
        "Coupon, accrued interest, and sale proceeds are intentionally not expanded here.",
        "The screen records the decision boundary; it does not grade the trade."
      ],
      attribution: {
        fx: {
          status: "known",
          value: "RUB-denominated instrument",
          detail: "FX is not part of the synthetic instrument path."
        },
        fees: {
          status: "partial",
          value: "Brokerage costs pending",
          detail: "The contract expects a source-linked fee amount."
        },
        taxes: {
          status: "partial",
          value: "Tax lot treatment pending",
          detail: "Coupon and disposal tax are kept distinct from price movement."
        },
        rubCounterfactual: {
          status: "partial",
          value: "Cash alternative is the comparison candidate",
          detail: "Rate, duration, and horizon must be fixed before calculating it."
        }
      }
    },
    {
      id: "summer-rub-to-usd-shift",
      title: "Summer RUB-to-USD shift",
      period: "2024-06 → 2024-08",
      decision: "Shift a tranche from RUB to USD over the summer.",
      context:
        "Create a hard-currency reserve while retaining enough RUB liquidity for near-term obligations.",
      amount: { label: "1.25m", currency: "RUB equivalent" },
      outcome:
        "Currency exposure changed; the decision outcome cannot be isolated from FX, fees, taxes, and timing.",
      currentStatus: "Monitoring",
      sourceRefs: ["DH-SYN-004 decision note", "DH-SYN-004 conversion summary"],
      caveats: [
        "The fixture does not contain exchange receipts, counterparties, or account identifiers.",
        "A currency move is not treated as profit without a selected counterfactual."
      ],
      attribution: {
        fx: {
          status: "partial",
          value: "Direction recorded; rate path omitted",
          detail: "The private adapter supplies dated conversion evidence."
        },
        fees: {
          status: "unknown",
          value: "Not supplied",
          detail: "Spread and transfer fees need source evidence."
        },
        taxes: {
          status: "unknown",
          value: "Not supplied",
          detail: "No tax result is inferred from a currency conversion."
        },
        rubCounterfactual: {
          status: "partial",
          value: "RUB reserve is the selected comparison",
          detail: "The horizon and liquidity value still need to be fixed."
        }
      }
    }
  ]
};

export const privateDecisionHistoryAdapter = {
  async load(fetcher: typeof fetch): Promise<DecisionHistoryLoadResult> {
    const response = await fetcher("/api/decision_history");
    if (!response.ok) {
      return {
        kind: "unavailable",
        reason:
          response.status === 404
            ? "Private adapter is not connected."
            : `Adapter returned ${response.status}.`
      };
    }

    const payload = (await response.json()) as PrivateDecisionHistoryPayload;
    if (payload.schemaVersion !== "decision-history/v1" || !Array.isArray(payload.cases)) {
      return { kind: "unavailable", reason: "Adapter returned an unsupported contract." };
    }

    return { kind: "ready", payload };
  }
};
