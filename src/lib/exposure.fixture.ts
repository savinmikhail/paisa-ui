import type { ExposureReport } from "./exposure";

/**
 * Stable public fixture for the Exposure screen and screenshot capture.
 * Values are fictional and must never be replaced with private Ledger data.
 */
export const syntheticExposureReport: ExposureReport = {
  contract_version: "exposure.v1",
  as_of: "2026-09-01",
  base_currency: "RUB",
  source: "synthetic-fixture",
  assets: [
    {
      id: "asset-aurora-reserve",
      label: "Aurora Reserve",
      kind: "Cash reserve",
      market_value: 210000,
      currency: "RUB",
      failure_domains: {
        institution: {
          value: "Northstar Cooperative",
          source: "synthetic-fixture/exposure-01",
          verified_at: "2026-08-28",
          confidence: "high"
        },
        custodian: {
          value: "Northstar Cooperative",
          source: "synthetic-fixture/exposure-01",
          verified_at: "2026-08-28",
          confidence: "high"
        },
        jurisdiction: {
          value: "RU",
          source: "synthetic-fixture/exposure-01",
          verified_at: "2026-08-28",
          confidence: "high"
        },
        correspondent_bank: {
          value: "Harbor Correspondent",
          source: "synthetic-fixture/exposure-01",
          verified_at: "2026-08-28",
          confidence: "medium"
        },
        withdrawal_route: {
          value: "Instant bank transfer",
          source: "synthetic-fixture/exposure-01",
          verified_at: "2026-08-28",
          confidence: "high"
        },
        key_control: {
          value: "Two-person approval",
          source: "synthetic-fixture/exposure-01",
          verified_at: "2026-08-28",
          confidence: "medium"
        }
      }
    },
    {
      id: "asset-lumen-notes",
      label: "Lumen Treasury Notes",
      kind: "Bond",
      market_value: 180000,
      currency: "RUB",
      failure_domains: {
        institution: {
          value: "Northstar Cooperative",
          source: "synthetic-fixture/exposure-02",
          verified_at: "2026-08-27",
          confidence: "high"
        },
        broker: {
          value: "Lumen Securities",
          source: "synthetic-fixture/exposure-02",
          verified_at: "2026-08-27",
          confidence: "high"
        },
        custodian: {
          value: "Orbit Custody",
          source: "synthetic-fixture/exposure-02",
          verified_at: "2026-08-27",
          confidence: "high"
        },
        depositary: {
          value: "Orbit Depositary",
          source: "synthetic-fixture/exposure-02",
          verified_at: "2026-08-27",
          confidence: "medium"
        },
        jurisdiction: {
          value: "RU",
          source: "synthetic-fixture/exposure-02",
          verified_at: "2026-08-27",
          confidence: "high"
        },
        correspondent_bank: {
          value: "Harbor Correspondent",
          source: "synthetic-fixture/exposure-02",
          verified_at: "2026-08-27",
          confidence: "medium"
        },
        withdrawal_route: {
          value: "T+1 bank transfer",
          source: "synthetic-fixture/exposure-02",
          verified_at: "2026-08-27",
          confidence: "medium"
        },
        key_control: {
          value: "Settlement confirmation",
          source: "synthetic-fixture/exposure-02",
          verified_at: "2026-08-27",
          confidence: "medium"
        }
      }
    },
    {
      id: "asset-helio-etf",
      label: "Helio Global ETF",
      kind: "Fund",
      market_value: 145000,
      currency: "USD",
      failure_domains: {
        institution: {
          value: "Lumen Securities",
          source: "synthetic-fixture/exposure-03",
          verified_at: "2026-08-26",
          confidence: "high"
        },
        broker: {
          value: "Lumen Securities",
          source: "synthetic-fixture/exposure-03",
          verified_at: "2026-08-26",
          confidence: "high"
        },
        custodian: {
          value: "Orbit Custody",
          source: "synthetic-fixture/exposure-03",
          verified_at: "2026-08-26",
          confidence: "high"
        },
        depositary: {
          value: "Orbit Depositary",
          source: "synthetic-fixture/exposure-03",
          verified_at: "2026-08-26",
          confidence: "high"
        },
        jurisdiction: {
          value: "IE",
          source: "synthetic-fixture/exposure-03",
          verified_at: "2026-08-26",
          confidence: "high"
        },
        withdrawal_route: {
          value: "Broker withdrawal",
          source: "synthetic-fixture/exposure-03",
          verified_at: "2026-08-26",
          confidence: "high"
        },
        key_control: {
          value: "Segregated custody",
          source: "synthetic-fixture/exposure-03",
          verified_at: "2026-08-26",
          confidence: "high"
        }
      }
    },
    {
      id: "asset-nova-reserve",
      label: "Nova Digital Reserve",
      kind: "Digital asset",
      market_value: 95000,
      currency: "USD",
      failure_domains: {
        broker: {
          value: "Northstar Prime",
          source: "synthetic-fixture/exposure-04",
          verified_at: "2026-08-25",
          confidence: "low"
        },
        custodian: {
          value: "Aurora Vault",
          source: "synthetic-fixture/exposure-04",
          verified_at: "2026-08-25",
          confidence: "medium"
        },
        withdrawal_route: {
          value: "Whitelisted wallet",
          source: "synthetic-fixture/exposure-04",
          verified_at: "2026-08-25",
          confidence: "medium"
        },
        key_control: {
          value: "Hardware signer quorum",
          source: "synthetic-fixture/exposure-04",
          verified_at: "2026-08-25",
          confidence: "low"
        }
      }
    },
    {
      id: "asset-cedar-note",
      label: "Cedar Opportunity Note",
      kind: "Private note",
      market_value: 70000,
      currency: "RUB",
      failure_domains: {
        institution: {
          value: "Northstar Cooperative",
          source: "synthetic-fixture/exposure-05",
          verified_at: "2026-08-20",
          confidence: "medium"
        },
        custodian: {
          value: "Orbit Custody",
          source: "synthetic-fixture/exposure-05",
          verified_at: "2026-08-20",
          confidence: "medium"
        },
        jurisdiction: {
          value: "RU",
          source: "synthetic-fixture/exposure-05",
          verified_at: "2026-08-20",
          confidence: "medium"
        },
        correspondent_bank: {
          value: "Harbor Correspondent",
          source: "synthetic-fixture/exposure-05",
          verified_at: "2026-08-20",
          confidence: "low"
        }
      }
    }
  ]
};

export const EXPOSURE_SCREENSHOT_INPUT = {
  route: "/assets/exposure",
  viewport: { width: 1440, height: 1100 },
  selected_bucket: "institution:Northstar Cooperative",
  expected_labels: ["Exposure", "Shared exposure map", "Northstar Cooperative", "Unknown"]
} as const;
