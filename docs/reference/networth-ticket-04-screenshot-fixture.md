# Net Worth ticket 04 screenshot fixture

The UI fixture is deterministic and contains no Ledger data. It is activated
with this URL:

```text
http://127.0.0.1:5173/assets/networth?fixture=networth-ticket-04
```

Run it from the UI worktree with `npm run dev -- --host 127.0.0.1 --port 5173`.
The fixture-only app layout bypasses `/api/config`, `/api/networth`, and the
reconciliation endpoint, so the capture does not read the private Ledger.

Capture these three states at the same viewport (1440 x 1200 is recommended):

1. Initial interval/decomposition: wait for `[data-testid="networth-decomposition"]`
   to show the `Reconciled` tag and capture the full page.
2. Selected interval: fill `#networth-from` with `2024-02-29`, fill
   `#networth-to` with `2024-06-30`, click the `Apply` button, wait for the
   `Reconciled` tag, and capture the full page. The summary should read opening
   `263,050`, closing `298,300`, absolute delta `35,250`, percent delta
   `13.40%` (currency formatting follows the fixture's USD config).
3. Hover detail: hover
   `#d3-networth-timeline path[data-series="networth"][data-point-index="4"]`
   and capture the chart plus `[data-testid="networth-hover-detail"]`. The
   detail panel must show `30 Apr 2024` and the corresponding point value.

The chart hit areas expose `data-series` and `data-point-index` specifically so
the hover capture is stable across runs. The deterministic source is
`src/lib/fixtures/networth-ticket-04.ts`; the interaction assertions are in
`src/lib/networth_reconciliation.test.ts`.
