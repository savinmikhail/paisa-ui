# Expense baseline screenshot capture

The screenshot input is deterministic and contains only synthetic values:
`tests/screenshot-fixtures/expense-baseline.capture.json`.

1. Start a synthetic-only config mock on an unused local port, for example
   `7550`, returning the read-only config shape used by the UI. It must not
   proxy to the personal Paisa service on port `7500`.
2. Start the UI from this worktree with
   `PAISA_API_TARGET=http://127.0.0.1:7550 npm run dev -- --host 127.0.0.1 --port 5174`.
3. Open `http://127.0.0.1:5174/expense/monthly?fixture=expense-baseline`.
4. Capture a 1440×1200 screenshot at device scale factor 1 after the text in
   `wait_for_text` is visible.
4. Verify the capture includes the six-month bars, `100` median metrics,
   Business/Tax/Capital Placement exclusions, the Travel one-off, the Food
   `+400% vs median` signal, and the Synthetic rent monthly commitment.

The `fixture=expense-baseline` query parameter bypasses `/api/expense` only
for this route. It does not read or copy the private Ledger and is safe for
browser automation or a supervisor capture.
