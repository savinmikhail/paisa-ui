# Income screen screenshot input

Use `tests/fixtures/income-screen.synthetic.json` as the deterministic API
input for a screenshot harness. It contains only synthetic postings and fixed
2024 dates; it is not a runtime Ledger or a copy of private data.

## Supervisor capture

1. Start the public UI worktree at `/private/tmp/paisa-ui-income-baseline` with
   the supervisor's fixture/mock API adapter returning `income_timeline` from
   the fixture for `GET /api/income` and `income_analysis` as the response for
   `GET /api/income/analysis`.
2. Open the Income route at `/income` at a desktop viewport at least 1280px
   wide and wait for both requests to settle.
3. Capture the full page, including the first Monthly Income Timeline and the
   lower Realized Passive Income section.
4. Capture a second state after toggling `Salary` off in the first chart
   legend. The red dashed `4-month rolling median` must remain visible and the
   y-axis must rescale to the visible groups.
5. Capture a third state after toggling `Coupons` off in the passive legend.
   The passive bars must update while the burn/freedom-floor lines remain
   independently toggleable.

Acceptance evidence: the captures show actual income bars, the optional
four-month median, passive interest/coupon/dividend bars, current/3m/12m
summary values, and burn/freedom-floor coverage without any personal data.
