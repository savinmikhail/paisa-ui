# Cross-surface screenshot fixtures

The reporting screens use explicit, public-safe fixture IDs for deterministic
browser captures. Every ID in the table is handled by the app layout before it
requests `/api/config`, so a capture can run against the UI dev server without
the private Ledger service. Fixture mode sets `USER_CONFIG.readonly = true` and
uses fictional values only; it must never be used as a way to select a private
source.

| Surface | URL | Fixture source | Range interaction |
| --- | --- | --- | --- |
| Expense retrospective | `/expense/monthly?fixture=expense-baseline` | `src/lib/fixtures/expense-baseline.json` | month selector and retrospective six-month window |
| Net Worth decomposition | `/assets/networth?fixture=networth-ticket-04` | `src/lib/fixtures/networth-ticket-04.ts` | arbitrary inclusive fixture dates |
| Decision History | `/assets/decision-history?fixture=synthetic` | `src/lib/decision_history.ts` | curated records; no automatic performance range |
| Exposure | `/assets/exposure?fixture=synthetic` | `src/lib/exposure.fixture.ts` | as-of metadata snapshot |
| Resilience | `/resilience?fixture=resilience` | `src/lib/fixtures/resilience.json` | shared inclusive range across all metrics |
| Money Flow | `/assets/money-flow?fixture=money-flow` | `src/lib/fixtures/money-flow.ts` | shared date, currency, and category filters across tabs |

The fixture query must be present in the URL. The unqualified routes use the
configured private adapter or legacy Ledger API and are not screenshot inputs.
The fixtures are read-only presentation contracts: they do not call write
endpoints, mutate Ledger files, or copy private data into this repository.

For each range-capable report, select an arbitrary range within the fixture's
available dates and verify that all visualizations on that surface use the same
selection. Net Worth and Money Flow additionally expose their reconciliation
contracts; their additive decomposition must equal closing minus opening. The
Income/passive-income and Expense screens keep their reporting semantics from
`docs/reference/reporting-semantics.md`: internal transfers are not income or
expense, and passive income excludes FX and market revaluation.

Income and passive income are intentionally not in this query-fixture
allowlist. Their existing capture uses the synthetic API mock described in
`tests/screenshot/income-screen.capture.md` on a separate local port (the
current workflow uses `5175`), serving the public-safe fixture to
`GET /api/income` and `GET /api/income/analysis`. This keeps the Income surface
explicitly separate from the route-level fixture bypass and does not promise a
private Ledger-free `?fixture=` URL for it.

## Reproducible capture

From this worktree:

```bash
npm run dev -- --host 127.0.0.1 --port 5173
```

Use a 1440px-wide viewport and wait for the surface-specific ready marker:

```text
expense:    [data-testid="expense-retrospective"]
net worth:  [data-testid="networth-decomposition"]
decisions:  [data-testid="decision-history-ready"]
exposure:   [data-testid="exposure-screen"]
resilience: [data-testid="resilience-charts"]
money flow: [data-testid="money-flow-sankey-panel"]
```

Capture the initial state, then the documented range/tab interaction for that
surface. Never capture without the fixture query unless private adapter access
has been explicitly configured for the local session.
