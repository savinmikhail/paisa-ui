# Resilience screenshot fixture

Use the explicit synthetic route for deterministic visual review:

```text
http://127.0.0.1:7500/resilience?fixture=resilience
```

This route loads `src/lib/fixtures/resilience.json` through the TypeScript
adapter. It does not call the private Ledger adapter and contains no personal
financial facts. The page should show three synchronized small multiples,
independent y-axis labels, a shared date-range toolbar, and the liquid scope
with `Reserve Asset USDT` included while `Investments`, `IIS`, and
`Receivables` are excluded.

For a local capture after the frontend build, start the Paisa service in this
worktree and capture the URL above with the browser screenshot tool. Hover any
chart point and verify that the same date and values appear across all three
panels. The API's explicit synthetic equivalent is:

```text
GET /api/resilience?source=synthetic
```

The unqualified `GET /api/resilience` remains fail-closed in this slice because
no private runtime adapter is implemented.
