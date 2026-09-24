# Decision result screenshot capture

Use the public-safe deterministic fixture in `src/lib/decision_history.ts`.
It contains no personal Ledger records.

1. Start the UI with `npm run dev -- --host 127.0.0.1`.
2. Open `http://127.0.0.1:5173/assets/decision-history?fixture=synthetic`.
3. Inspect at 1440px and 390px. Wait for
   `[data-testid="decision-history-ready"]`; expect four decision sections.
4. Confirm the summary distinguishes nominal profit from advantage against
   savings; at least one profitable example underperforms the benchmark.
5. On mobile both results must be visible without horizontal page overflow.
6. Expand operations and source disclosures. Confirm contribution dates,
   monetary amounts and the hypothetical payout carry explanation are readable.

Live personal screenshots belong only in the private Ledger repository.
Do not copy live API responses, amounts or screenshots into this public checkout.
