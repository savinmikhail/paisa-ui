# Reporting semantics

The reporting seam is a read-only view over Ledger postings. It does not add
persisted classifications, rewrite postings, or change importer and account
taxonomy behaviour.

## Periods and buckets

All seam periods use `[start, end)`: the start instant is included and the end
instant is excluded. Calendar-month buckets use the same boundary and are
clipped at the requested period edges. Empty months remain present so a chart
does not change its x-axis when a month has no matching postings.

The existing income API keeps its historical inclusive `end_date` field for
compatibility. The adapter uses the exclusive boundary internally and converts
it back only at that API boundary.

Forecast postings are excluded by default, matching `query.Query.All`. A
caller must opt in with `IncludeForecast`.

## Account-path rules

| Report | Existing predicate | Notes |
| --- | --- | --- |
| Income | `Income:*` | Includes capital gains and refunds, as the current income report does. |
| Expenses | `Expenses:*` except `Expenses:Tax:*` | Tax is reported separately. |
| Tax | `Expenses:Tax:*` | Uses the built-in Paisa tax prefix. |
| Passive income | `Income:Interest:*` or `Income:Dividend:*` | Unrealized FX and market revaluation are not included. |
| Capital gains | `service.IsCapitalGains` | Uses the existing `Income:CapitalGains:*` service rule. |
| Refunds | `service.IsRefund` | Keeps the original `Income:Refund:*` posting visible. |
| Internal transfers | `Equity:Transfers:*` | Never becomes income or expense. |
| Asset movements | `Assets:*` | This is intentionally broad; it does not decide Asset Movement versus Capital Placement. |

Each result retains the complete source posting plus explicit account-path,
transaction, file, and line-range provenance. Consumers can therefore link a
metric back to the journal without copying Ledger data into a new model.

## Deliberate boundary

The seam does not yet define passive-income subcategories, arbitrary Net Worth
range calculations, FX/market/debt decomposition, liquidity policy, or the
distinction between Asset Movement and Capital Placement. Those are derived
views for later tickets and must not be smuggled into this compatibility layer.
