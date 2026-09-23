# Decision history screenshot capture

This capture uses only the deterministic public-safe fixture in
`src/lib/decision_history.ts`. It does not read the private Ledger repository.

1. Start the UI from this worktree with `npm run dev -- --host 127.0.0.1`.
2. Open `http://127.0.0.1:5173/assets/decision-history?fixture=synthetic`.
3. Capture the full page at 1440×1400 and 390×1200. The route should expose
   `[data-testid="decision-history-ready"]` and exactly four
   `[data-testid^="decision-case-"]` elements.
4. Confirm the capture shows the four case names, all four attribution fields
   (FX, Fees, Taxes, RUB counterfactual), source refs, and evidence caveats.
5. Do not capture the route without `fixture=synthetic` unless a supervisor has
   explicitly configured the private adapter. The non-fixture route is expected
   to show the adapter-unavailable state in this public checkout.

The screenshot-test input is
`tests/screenshots/decision-history.synthetic.json`.
