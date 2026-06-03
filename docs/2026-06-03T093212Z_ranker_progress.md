# 2026-06-03T09:32:12Z Ranker Progress

## Work Completed

- Added costed active-sleeve Python comparison: `research/ml/compare_active_sleeves.py`.
- Aligned Python composite-sleeve ranking and weighting with Go mechanics:
  - `vol_adjusted_momentum`
  - `risk_adjusted_edge`
  - candidate-symbol allowlists
  - Python half-L1 rebalance band equivalent to Go full-L1 target turnover
- Tightened Python daily-ranker pre-screen status with turnover and low-excess guards.
- Added official Go `benchmark_ranker_proxy` family with PBO variants.

## Best Python Checkpoints

- 2022-2026 ranker: `lambdarank_stocks_sleeve15_top3_reb42_seed17`
  - Return: 98.36%
  - VOO return: 68.74%
  - Excess: 29.63%
  - Sharpe: 0.921 versus VOO 0.730
  - Max drawdown: 25.56% versus VOO 24.52%
  - Turnover: 4.495
  - Status: `candidate` in Python pre-screen
- 2021-2026 smaller sleeve: `lambdarank_stocks_sleeve10_top3_reb42_seed17`
  - Return: 134.02%
  - VOO return: 121.66%
  - Status: `candidate` in Python pre-screen
- 2023-2026 smaller sleeve:
  - Return: 122.70%
  - VOO return: 108.28%
  - Status: `research_only_weak_validation`

## Official Go Result

- Report: `reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_longpanel/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.md`
- Candidate: `benchmark_ranker_proxy_checkpoint`
- Return: 433.00%
- VOO buy-hold return: 351.93%
- Sharpe: 0.915
- DSR: 1.000
- PBO: 0.600
- Promotion: false
- First rejection reason: `PBO 0.600 above 0.200`

## Verdict

The ranker direction is a promising checkpoint, not proven alpha. Python pre-screen returns improved after turnover control, but the official Go gate still rejects the deterministic proxy for PBO fragility.

## Verification

- `cd backend && go test ./...`
- `cd backend && go build ./...`
- `research/ml/.venv/bin/python -m py_compile research/ml/*.py`
- `git diff --check`

## Next

- Build a ranker-specific walk-forward validation path with model-per-fold artifacts.
- Add Python-vs-Go parity for the daily-ranker feature set.
- Continue reducing PBO fragility before claiming alpha.
