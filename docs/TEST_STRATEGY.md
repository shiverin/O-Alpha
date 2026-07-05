# O(Alpha) Test Strategy

This document is the developer contract for safety and system testing. The goal is to make paper-agent behavior, quant validation, and dashboard controls verifiable without relying on manual inspection.

## Test Tiers

| Tier               | Command                                            | Required For                       | Notes                                                                                                                                        |
| ------------------ | -------------------------------------------------- | ---------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Backend unit       | `cd backend && go test ./...`                      | Every PR                           | Fast, offline tests for strategy math, validation gates, API helpers, repositories without live DB dependencies, and portfolio-agent policy. |
| Frontend unit      | `cd frontend && npm test`                          | Every PR                           | Vitest + React Testing Library for stateful UI behavior and client-side guardrails.                                                          |
| Frontend quality   | `cd frontend && npm run lint && npm run typecheck` | Every PR                           | Static safety checks before build/deploy.                                                                                                    |
| Frontend e2e smoke | `cd frontend && npm run test:e2e`                  | PR smoke                           | Playwright browser test with mocked backend routes for critical dashboard/settings behavior.                                                 |
| DB integration     | `cd backend && go test -tags=integration ./...`    | Local/nightly when DB is available | Requires `OALPHA_TEST_DATABASE_URL`; tests may run migrations and must use disposable test users/data.                                       |
| Full system        | Docker/API/frontend/manual or scheduled harness    | Release confidence                 | Seed deterministic bars, run migrations, exercise onboarding, start/stop agent, persistence, dashboard reads, and settings lockout.          |
| Live/external      | Workflow dispatch/scheduled only                   | Operational checks                 | Alpaca/data-ingest checks require secrets and must never be mandatory for ordinary PRs.                                                      |

## Safety Invariants

- A user may have at most one active portfolio catalog agent.
- Saved agent settings are immutable while any agent run is `starting` or `running`; exact no-op saves may pass.
- Portfolio-agent settings changes must be server-enforced. Frontend disabled controls are only a UX layer.
- Runtime fills must be idempotent by deterministic order/fill keys.
- Research metrics must be artifact-backed under `reports/batches/`; hand-entered Sharpe, DSR, PBO, drawdown, or promotion claims are invalid.
- Promotion is fail-closed unless PBO is actually estimated, DSR/PBO/trade-count gates pass, and benchmark-aware risk improvement is present.
- Strategies and ML features must preserve point-in-time integrity and read only data available at the evaluated bar.

## Fixture Policy

- Unit tests should use deterministic in-memory bars, prices, and strategy outputs.
- Research/system tests should write reports to temporary directories unless the report is intentionally part of a research session artifact.
- DB integration tests must use `OALPHA_TEST_DATABASE_URL`, not production `DATABASE_URL`.
- External market data and Alpaca paper checks are opt-in. Missing secrets should skip or fail with a clear configuration error, not substitute fake live data.

## Current Critical Flows

- Settings lockout: active portfolio agent appears in `/api/v1/agent/list`; settings page disables all controls; backend rejects changed settings with `409`.
- Portfolio run lifecycle: launch writes `agent_runs`, marks running, evaluates, records fills/positions/snapshots/alerts, and stop marks the latest active run stopped.
- Onboarding: user saves risk settings, accepts a matching catalog backtest, completes onboarding, and dashboard launches the matching strategy bucket.
- Research harness: `cmd/alpha-research` produces JSON/MD reports with cost stress, DSR, PBO, promotion reasons, and benchmark comparison.

## CI Expectations

- Pull requests run backend unit tests, frontend formatting/lint/typecheck/unit tests, frontend build, and Playwright smoke tests.
- Integration tests are intentionally tagged because they require a reachable disposable Postgres database.
- Nightly or release workflows should add `go test -tags=integration ./...` and full-stack seed/start/stop assertions when infrastructure is available.

## Local Reproduction

```bash
make test-all
cd frontend && npm run lint && npm run typecheck
cd frontend && npm run test:e2e
OALPHA_TEST_DATABASE_URL=postgres://... cd backend && go test -tags=integration ./...
```

When a test fails, prefer adding a focused regression test at the lowest tier that can reproduce the bug. Use broader system tests only when the bug depends on multiple processes, database state, or browser behavior.
