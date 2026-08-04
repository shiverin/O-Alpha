# O(Alpha) — Frontend Integration Plan (Paper Trading)

_Scope: wire the validated alphas into the agent UI end-to-end — risk-profile selection → agent
initialization → live paper run → all dashboards (balance, allocation, trade log, buy/sell stream,
risk alerts). Paper only. Live broker execution stays disabled per `docs/BLOCKERS.md`._

---

## 0. The headline finding (read this first)

You have **two separate, non-connected agent systems**, and the frontend is wired to the wrong one:

| | Single-symbol agent (`internal/agent/worker.go`, `manager.go`) | Portfolio agent (`internal/agent/portfolio/`) |
|---|---|---|
| Strategies | MA_CROSSOVER, KALMAN, HMM_ENSEMBLE, ML_META_LABEL — **one symbol** | The validated catalog: `lgbm_ranker_h63_*`, `ranker_proxy_h63_*`, momentum sleeves — **multi-asset** |
| Wired to API? | Yes (`/agent/start`, `/agent/stop`) | **No routes** |
| Instantiated in `main.go`? | Yes | **No** — only referenced in tests |
| Persists to DB (fills/positions/snapshots)? | Yes (`PortfolioRepository`) | **No** — in-memory `PortfolioPaperAccount`, no DB-backed `ExecutionRouter` |
| Drives the dashboard today? | Yes (hardcoded `AAPL` + `HMM_ENSEMBLE`) | No |

So the "working strategies and alphas" you want to surface (the h63 ranker catalog that
`docs/PLAN.md` says to default to — `lgbm_ranker_h63_low`, `ranker_proxy_h63_low`) currently **cannot
reach the UI at all**. The dashboard's "Launch Agent" button starts a single-symbol HMM toy on AAPL,
which is unrelated to your research output.

Everything below is built around closing that gap as the central task.

---

## 1. Verified gap list (what isn't implemented / is wrong)

Each item was confirmed against the code, with the file noted.

1. **Portfolio agent never instantiated.** `NewPortfolioAgentManager` appears only in tests
   (`strategy_catalog_test.go`). `cmd/api/main.go` wires only `agent.NewAgentManager`. → The catalog
   alphas are unreachable from the running service.
2. **No DB-backed `ExecutionRouter`.** Only the interface exists (`portfolio/worker.go`); the worker's
   `PortfolioPaperAccount` is pure in-memory. Portfolio fills/positions/equity are never written to
   Postgres, so no dashboard can show them.
3. **Risk alerts feed is always empty.** `system_alerts` has only a `SELECT` (`GetSystemAlerts`) and a
   table definition — **no `INSERT` anywhere in non-test code**. The worker's stop-loss / take-profit
   interventions in `enforceRiskRules` only `log.Printf`; they never persist an alert. The Activity
   page's alerts panel will be permanently blank for real users.
4. **`estimated_annual_yield` is hardcoded to `0`** in `SavePortfolioSnapshot` (the literal `0` arg)
   but is displayed on the Portfolio page. Either compute it or hide the field.
5. **Dashboard launch is hardcoded.** `dashboard/page.tsx` always sends `symbol: "AAPL"`,
   `strategy_type: "HMM_ENSEMBLE"`. The user's chosen strategy/universe never flows into the launch.
6. **`StrategyControls` sliders are cosmetic.** Risk tolerance / volatility cap / leverage multiplier
   are local React state. Only `riskTolerance` is bucketed into a `risk_profile` string on launch;
   volatility cap and leverage multiplier go nowhere (not saved, not sent).
7. **No agent-status endpoint.** `isAgentActive` is local state, lost on refresh. There is no
   `GET /agent/list` or `/agent/status`, so the UI can't reconcile what's actually running. The
   `agent_runs` table tracks status but is never read back by the API.
8. **Worker ignores most risk settings.** `enforceRiskRules` reads only `stop_loss_pct` and
   `take_profit_pct`. `leverage`, `max_positions`, `rebalance_freq` are saved by Settings but unused
   at runtime.
9. **`PortfolioAllocation` donut is a static mock.** Hardcoded SVG `strokeDashoffset` values; it never
   calls `/user/portfolio/positions`, so allocation is fake even when logged in.
10. **No restart resilience / heartbeat.** Workers live only in memory. `MarkAgentRunRunning` sets
    `last_heartbeat_at` once and nothing updates it again. After an API restart, every worker dies but
    `agent_runs` rows stay `running` — zombie state, and the user's agent silently stops.
11. **Model artifacts are gitignored** (`docs/BLOCKERS.md`). `ML_META_LABEL` and `lgbm_ranker_h63_*`
    load files under `reports/batches/.../fold_artifacts`. In a fresh/deployed environment these
    paths won't exist and the agent will fail to start. They must be mounted or moved to managed
    storage before these strategies are selectable.
12. **No idempotency on fills.** The schema has a unique index on `orders.client_order_id`, but
    `RecordLongFill` never sets it. A retried tick or duplicate signal can double-book a fill. Trading
    OMS best practice is a unique client order id per intended order so retries dedupe instead of
    duplicating.
13. **"Live" dashboards are polling, not live.** All panels use SWR polling. Fine for v1, but the
    "Live Execution Log" label oversells it; consider SSE later.

Not a bug (by design, keep it): `executeLiveTrade` is a stub and `paperTrade` is forced `true`. That
matches `BLOCKERS.md` ("do not connect the ranker to broker execution without explicit approval").
Keep the live path stubbed.

---

## 2. Target end-to-end flow

```
Risk profile (onboarding/settings)
      │  persists agent_settings  + selects a default strategy bucket
      ▼
Strategy selection  ──►  POST /agent/start { strategy_key, symbols[], risk_profile, initial_cash }
      │
      ▼
PortfolioAgentManager.StartCatalogPortfolioAgent(...)   (multi-asset, validated alphas)
      │  ExecutionRouter (DB-backed) writes fills/positions/snapshots/alerts
      ▼
Postgres  ──►  GET summary | history | positions | trades | alerts | agent status
      ▼
Dashboards: Balance · Allocation · Execution Log · Risk Alerts · Agent status badge
```

The single-symbol agent stays available for demos, but the **primary path is the portfolio agent**.

---

## 3. Phased build plan

### Phase 1 — Make the validated alphas runnable & persistent (backend)

This is the unlock. Nothing in the UI matters until the portfolio agent can run and write to the DB.

1. **Add a DB-backed `ExecutionRouter`.** Create `internal/agent/portfolio/db_execution_router.go`
   implementing `ExecutePortfolioTargets`. It should:
   - translate target weights → per-symbol target notional → buy/sell deltas vs current DB positions;
   - reuse `PortfolioRepository.RecordLongFill` for each leg (long-only first; short support later);
   - call `MarkPositionPrice` + `SavePortfolioSnapshot` after applying the batch;
   - set a deterministic `client_order_id` (e.g. `runID:symbol:barTime`) for idempotency — add a
     `RecordLongFillIdempotent` that early-returns if the `client_order_id` already exists.
2. **Persist portfolio fills with the same account model** the single-symbol path uses
   (`accounts`/`positions`/`orders`/`fills`/`cash_ledger`/`portfolio_snapshots`) so all dashboards
   read one source of truth regardless of which agent produced the trade.
3. **Instantiate `PortfolioAgentManager` in `main.go`** and pass it into `api.NewHandler`.
4. **Provision model artifacts** (gap #11): decide on a mounted volume or object-store path, set
   `OALPHA_DAILY_RANKER_ARTIFACT_ROOT` / `OALPHA_DAILY_RANKER_PIT_UNIVERSE`, and **fail closed with a
   clear error** if a selected strategy's artifacts are missing (don't silently fall back).
5. **Write alerts** (gap #3): add `PortfolioRepository.InsertSystemAlert(...)` and call it from:
   - stop-loss / take-profit interventions,
   - regime transitions from the HMM overlay (INFO),
   - agent start/stop/failed (INFO/CRITICAL).
6. **Heartbeat + reconciliation** (gap #10): have each worker update `last_heartbeat_at` every loop;
   on `main.go` startup, mark any `running` row with a stale heartbeat as `failed`
   (`reason="orphaned_on_restart"`). This matches the regulatory expectation that algo systems
   reconcile their own trade logs and support timely resumption after disruption.

**Acceptance:** start `lgbm_ranker_h63_low` for a small symbol set via a temporary curl; confirm
fills, positions, snapshots, and at least one alert appear in Postgres.

### Phase 2 — API surface for the portfolio agent

1. **Extend `AgentControlRequest`** (or add a dedicated `/agent/portfolio/start`) to accept:
   `strategy_key` (from the catalog), `symbols []string`, `risk_profile`, `initial_cash`, `timeframe`.
2. **`POST /agent/start`**: when `strategy_key` is a catalog key, route to
   `StartCatalogPortfolioAgent` with the DB execution router; otherwise keep the single-symbol path.
   Create an `agent_runs` row (reuse `CreateAgentRun`; store `strategy_key` + `symbols` in
   `parameters`).
3. **`GET /agent/list`** (new): return active runs for the user from `agent_runs`
   (status, strategy, symbols, started_at, last_heartbeat) so the UI can reconcile after refresh.
4. **`GET /strategies/catalog`** (new): expose `AvailableStrategySpecs` (key, display name, risk
   profile, deployment status, `paper_only`, description, notes). This powers the selection UI and
   lets you surface the "research/paper only" provenance honestly.
5. **Validate symbols** against what's ingestable; reject empty universes with a clear 400.

**Acceptance:** the five endpoints return correct JSON; starting/stopping a catalog strategy is
visible in `GET /agent/list`.

### Phase 3 — Risk-profile → strategy selection (frontend)

1. **Map risk profile → default catalog strategy** (align with `docs/PLAN.md`):
   conservative/moderate → `lgbm_ranker_h63_low` or `ranker_proxy_h63_low`; expose medium/high as
   explicit "research/paper-only" opt-ins with the catalog's own warning notes shown in the UI.
2. **Build a Strategy Selection step** (reuse/extend `StrategySelector.tsx`, currently only on the
   performance page): fetch `/strategies/catalog`, show cards with risk profile + deployment status +
   notes, let the user pick a strategy and a symbol universe (default to the benchmark + a curated
   set you have bars for).
3. **Wire `StrategyControls` sliders to real settings** (gap #6) or relabel them as preview-only.
   Preferred: persist volatility cap / leverage intent through `agent_settings` and actually consume
   them in sizing, or remove them to avoid implying control that doesn't exist.
4. **Replace the hardcoded launch** (gap #5) in `dashboard/page.tsx` with the selected
   `{ strategy_key, symbols, risk_profile, initial_cash }`.

**Acceptance:** choosing a risk profile preselects the right alpha; launching sends the real payload.

### Phase 4 — Dashboards on real data

1. **Agent status badge** (gap #7): drive `isAgentActive` from `GET /agent/list` via SWR, not local
   state. Show strategy, symbols, and run status; survive refresh.
2. **`PortfolioAllocation`** (gap #9): fetch `/user/portfolio/positions`, compute real allocation
   ring from exposures; remove the static `strokeDashoffset` mock.
3. **Execution log / buy-sell stream**: already wired (`ExecutionLog`, Activity page) — verify it
   renders portfolio-agent fills now that they persist (it will, since they share the fills table).
4. **Risk Alerts** (gap #3): Activity page already reads `/user/portfolio/alerts`; it lights up once
   Phase 1.5 writes rows. Add severity styling by `alert_type`.
5. **`estimated_annual_yield`** (gap #4): either compute a trailing annualized figure in
   `SavePortfolioSnapshot` or hide the metric until you do. Don't display a hardcoded 0.

**Acceptance:** with an agent running, every panel reflects DB state and updates on poll; a forced
stop-loss produces a visible alert and a SELL row.

### Phase 5 — Hardening (before any wider use)

- Idempotent fills enforced end-to-end (gap #12); add a test that a duplicated tick doesn't
  double-book.
- Graceful worker shutdown writes a final snapshot and a stop alert.
- Concurrency: confirm one active portfolio run per user (or per `(user, strategy)`), mirroring the
  single-symbol manager's duplicate guard.
- Optional: SSE endpoint for the execution log to make "live" honest (gap #13).
- Keep `executeLiveTrade` stubbed; gate any live path behind an explicit env flag **and** the
  `docs/PLAN.md` promotion checklist (PBO ≤ 0.20, DSR ≥ 0.95, PIT coverage pass).

---

## 4. Suggested order of work (smallest path to a working demo)

1. DB-backed `ExecutionRouter` + instantiate `PortfolioAgentManager` (Phase 1.1–1.3).
2. `/agent/start` routes to catalog + `/strategies/catalog` + `/agent/list` (Phase 2).
3. Frontend: catalog selection + real launch payload + status badge (Phase 3 + 4.1).
4. Alerts writing + real allocation + yield fix (Phase 1.5, 4.2, 4.5).
5. Idempotency + heartbeat/reconciliation (Phase 1.6, 5).

After step 3 you have a genuine end-to-end demo: pick risk profile → pick validated alpha → launch →
watch real fills and balance move.

---

## 5. Explicit non-goals / guardrails

- **No live brokerage execution.** Your own `BLOCKERS.md` is right: the panel is survivorship-biased
  and PIT price coverage isn't cleared. Keep everything paper.
- **Don't promote medium/high catalog entries as "recommended."** Surface them as research/paper-only
  with their notes, per the catalog metadata.
- **Don't show numbers you don't compute** (annual yield = 0). Honesty in the UI matters more here
  than completeness.

---

## 6. Open questions for you

1. Concurrency model: one portfolio agent per user, or allow several (e.g. one per strategy)? This
   decides the `activeWorkers` key and the duplicate guard.
2. Symbol universe source for live paper: a fixed curated list you have Alpaca/Yahoo bars for, or
   user-selectable? The catalog alphas assume a benchmark (`VOO`) plus a stock universe.
3. Where will model artifacts live in deployment (mounted volume vs object store)? This blocks the
   LGBM ranker strategies from being selectable.
