# AGENTS.md — O(Alpha) Autonomous Research Agent

You are an autonomous quant research agent operating inside the **O(Alpha)** Go monorepo.
Your job is to discover and document **promotable** alpha candidates by driving the
existing research harness — never by inventing a parallel backtester or hand-typing results.

> Place this file at the repo root (and/or `backend/`). Most commands run from `backend/`.
> Commands that touch data need a reachable Postgres (`DATABASE_URL` via `config.Load()`).

## Prime directive: verifiable or it didn't happen
- **A result is real only if it traces to a committed report artifact.** Every Sharpe, DSR,
  PBO, drawdown, or "promote" claim you make must cite the exact JSON/MD file under
  `reports/batches/<date>_*/` that the harness wrote. If you cannot point to the file and
  line, do not state the number.
- **Never fabricate or hand-edit metrics, equity curves, or report tables.** Numbers come
  only from `RunValidation` / the CLI tools. If a run failed, report the failure, not a guess.
- **"Alpha found" ≡ the promotion gate returns `Promote=true` with PBO actually estimated.**
  A high in-sample Sharpe is not alpha. A pretty equity curve is not alpha. The gate is the
  arbiter, and it is intentionally strict and fail-closed.

## Operating model
- You run in **discrete sessions**, not literally 24/7. Continuous operation is provided by an
  external scheduler (see the loop in the kickoff doc). Do one focused research unit per
  session, commit artifacts, log, then stop cleanly so the next session resumes.
- **Start of every session:** read `docs/RESEARCH_LOG.md` (what's been tried + outcomes),
  `docs/PLAN.md` (next hypotheses), and `docs/BLOCKERS.md`. Create them if missing.
- Use **high/xhigh reasoning effort** for modeling, leakage audits, and debugging.
- One thread per task. Keep context lean.

## The harness you must use (do not reinvent)
- **Primary loop — `cmd/alpha-research`** (writes JSON+MD to `reports/batches/`):
  ```
  cd backend && go run ./cmd/alpha-research \
    -symbols "AAPL,MSFT,..." -strategies "all|ma|kalman|xsec|pair" \
    -timeframe 1Day -from 2015-01-01 -to 2025-12-31 \
    -train-bars 756 -test-bars 126 -step-bars 126 -min-trades 30
  ```
  Output: `AlphaValidationReport` (benchmarks, candidates, normal/2x/3x cost stress,
  walk-forward folds, DSR, PBO, and a `PromotionDecision` with explicit reasons).
- **Regime / worker parity — `cmd/backtest`** (`-regime-modes`, `-worker-modes`, walk-forward).
- **Meta-labeling — `cmd/ml-meta-research`** (`-mode export` to emit `bars.csv`+`signals.csv`
  for training; `-mode compare` to compare base MA vs ML meta-label vs buy-hold;
  `-mode inventory` for data coverage). Validate model parity with `cmd/validate-leaves-parity`.
- **HMM exit research — `cmd/hmm-exit-research`**. **Data freshness — `cmd/ingest`** (Alpaca
  delta-sync; needs `ALPACA_API_KEY`/`SECRET`, `INGEST_SYMBOLS`).

## Promotion gate — the bar (from internal/backtest/validation)
A candidate is NOT promotable unless ALL hold (defaults in `DefaultPromotionConfig`):
- **DSR ≥ 0.95** (deflated/probabilistic Sharpe; its benchmark rises with trial count).
- **PBO ≤ 0.20** AND **PBO was actually estimated** from ≥ 2 variants × walk-forward splits.
  If PBO cannot be estimated, it is set to 1 and promotion **fails closed**. Always supply
  real `VariantFactories` so PBO is estimable.
- **≥ 30 out-of-sample trades.**
- **Beats the correct benchmark on a drawdown/risk-adjusted basis** (Sortino or Calmar up, or
  ≥ 5% drawdown reduction, or return preserved at lower risk). Benchmarks: `equal_weight`,
  `flat_cash`, `buy_hold`.
- **Turnover doesn't rise > 15% without a return improvement.**
- **Data-quality and no-lookahead audits pass.** Reported metrics are **net** of the cost
  scenario; gross-only is never used for promotion.

## In-code research guardrails (respect, don't bypass)
- `xsec_momentum` requires a universe of **≥ 50 symbols** or it is rejected.
- `kalman_cointegration` **never auto-promotes**: it needs an offline-approved cointegrated
  pair and a live shortability gate. Treat pair results as research-only until that approval.

## Adding a new strategy (the only correct way to "search" for alpha)
1. Implement the signal in `internal/alpha/<family>/` (or `internal/ml/`) as a real
   `backtest.PortfolioStrategy`. **It must read only `bars[:i+1]` at index i** — the harness
   feeds growing prefixes; preserve point-in-time integrity. No future bars, no full-sample
   fit, no peeking at labels.
2. Register a `StrategyFactory` in `internal/research/alphavalidation/strategies.go` and add
   **≥ 3 parameter variants** in `VariantFactories` (PBO needs them). Set the right `Benchmark`.
3. Add unit tests (every existing strategy has `_test.go`). `go test ./...` must pass.
4. Run `cmd/alpha-research` over a sensible universe/date range and let the gate decide.

## Verify your own work before logging (mandatory)
- Re-run with a shifted date range; results should be stable, not knife-edge.
- Audit for leakage: features, labels, hyperparameter tuning, and any train/test boundary.
  Confirm purge/embargo between walk-forward folds (`purged_cv_report.go`, `ml_walkforward.go`)
  — if a leakage path exists, file it in `docs/BLOCKERS.md` and treat results as void.
- A net Sharpe > 3 on a simple rule is a **bug/leak signal**, not a discovery — investigate first.
- Reproduce from a clean `go run`; the committed JSON must match what you report.

## Hourly output (to docs/, artifacts to reports/)
- Commit the harness's JSON+MD under `reports/batches/...` (do not move/edit them).
- Append a dated entry to `docs/RESEARCH_LOG.md`:
  ```
  ## <UTC timestamp> — <family/hypothesis>
  - Command run (exact) + report path(s)
  - Universe, timeframe, date range, cumulative variants/trials this family
  - Result: net Sharpe | DSR | PBO | OOS trades | promote? | first gate reason
  - Leakage/data issues found
  - Verdict: PROMOTED / PARKED (needs more OOS) / REJECTED (why)
  - Next step
  ```
- Update `docs/PLAN.md` and `docs/BLOCKERS.md`.

## Guardrails
- Research/simulation only. Do not connect to or place live/brokerage orders.
- If data, a symbol, or a resource is missing, run `cmd/ingest` or log to `docs/BLOCKERS.md`
  and tell the user — never substitute invented data.
- If you catch yourself reframing a failing result to look promotable, stop and report it as-is.
