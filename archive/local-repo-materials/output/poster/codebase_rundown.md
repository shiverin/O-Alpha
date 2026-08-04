# O(Alpha) Current Codebase Rundown

- Backend: Go/Gin API with auth, CORS, migrations, bars repository, backtests, portfolio catalog routes, single-symbol legacy agents, and portfolio catalog agents.
- Research harness: alpha-research, backtest, ml-meta-research, HMM exit research, ranker parity tools, paper ranker signal, Alpaca/Yahoo ingest.
- Strategy catalog: nine paper-only catalog strategies across low/medium/high risk, including LGBM h63 rankers, deterministic h63 proxy, low-vol sleeve, ranked sleeves, TSMOM, and composite momentum.
- Promotion evidence: official artifacts live under reports/batches; low-risk LGBM h63 and ranker proxy pass both primary 2015 and shifted 2016 catalog-bucket windows.
- Paper execution: PortfolioOrchestrator starts one active portfolio run per user, warms daily bars, evaluates the catalog strategy, applies runtime settings, and routes targets through DB or Alpaca-paper routers.
- DB state: orders, fills, positions, cash ledger, portfolio snapshots, system alerts, agent runs, settings, bars, strategy trials, model artifact metadata, universes, pair candidates.
- Execution safety: long-only DB router sells reductions before buys, uses deterministic client_order_id keys, emits rebalance/risk-exit alerts, updates snapshots, and keeps live broker execution distinct.
- Runtime resilience: active portfolio runs can resume after restart, runtime HMM state is saved into agent_runs parameters, and heartbeats update each evaluation loop.
- Frontend: Next.js/React dashboard with onboarding, risk profile selection, catalog strategy selection, streamed backtest acceptance, launch/stop, status polling, summary/history/positions/trades/alerts, and live portfolio stream updates.
- Do not market: live real-money trading, nonzero annual yield, or medium/high PBO-failing variants as promoted alpha.
