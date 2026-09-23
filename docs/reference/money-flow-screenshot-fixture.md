# Money Flow screenshot fixture

The public-safe screenshot fixture contains fixed fictional values only. It is
activated at:

```text
http://127.0.0.1:5173/assets/money-flow?fixture=money-flow
```

This route does not call the private Ledger API. The backend has a matching
read-only deterministic contract at `GET /api/money_flow?source=synthetic` for
API and integration capture tests. The browser fixture is
`src/lib/fixtures/money-flow.ts`.

Capture at 1440×1100 (device scale factor 1):

1. Wait for `[data-testid="money-flow-sankey-panel"]` and capture the initial
   Sankey state. The diagram must show Income flowing to Personal Expense,
   Business, Tax, Reserve, and Capital Placement; unresolved transfers are
   absent.
2. Click `[data-testid="unresolved-transfers-toggle"]` and capture again. The
   separate `Unresolved transfers` destination and callout must be visible.
3. Click `[data-testid="money-flow-tab-waterfall"]` and capture the Waterfall.
   It must show Opening Net Worth, Savings, Realized passive income, Market
   revaluation, FX revaluation, Debt changes, Corrections, and Closing Net
   Worth, with the six steps summing to the closing minus opening value.

The date range, currency selector, and category filters are shared by both
tabs. Do not capture the route without the fixture query unless a supervisor
has explicitly authorized access to a private Ledger runtime.
