# Exposure

The Exposure screen is a read-only consumer of the `exposure.v1` reporting
contract. It groups manually curated failure-domain metadata by value and
retains the affected asset IDs for drilldown.

The contract covers institution, broker, custodian, depositary, jurisdiction,
correspondent bank, withdrawal route, and key control. Each field carries a
value or explicit unknown, plus `source`, `verified_at`, and `confidence`.
Unknown fields remain unknown in aggregation; the UI does not infer metadata
from account names or Ledger postings.

The public UI loads the contract through `src/lib/exposure_loader.ts` and the
read-only `GET /api/exposure` API. The default source is the private adapter;
the explicit `?source=synthetic` API route is used for deterministic tests and
screenshots. `src/lib/exposure.fixture.ts` remains the pure-function test
fixture and contains only fictional values. The private adapter reads a local
JSON file configured by the private Ledger repo; it must not copy `data/`, raw
transactions, or private source statements into this repository.

## Supervisor screenshot capture

From the UI worktree, install dependencies once and start the Svelte dev
server. The route itself is fixture-backed, but the existing app shell still
expects the local Paisa `/api/config` endpoint:

```bash
npm ci
PAISA_UI_DIR=/private/tmp/paisa-ui-ticket-06 \
  docker compose -f /private/tmp/paisa-ledger-ticket-06/compose.yaml up -d --build paisa
VITE_PAISA_EXPOSURE_SOURCE=synthetic npm run dev -- --host 127.0.0.1
```

Open `http://127.0.0.1:5173/assets/exposure?fixture=synthetic` at 1440×1100 and
capture the full page. The query fixture loads the pure public fixture without
requiring a backend. The stable screenshot input is exported as
`EXPOSURE_SCREENSHOT_INPUT`; select the default `Northstar Cooperative` row so
the affected-asset drilldown and the `Unknown` rows are visible.
