# VOO 5-Year Candidate Strategy Backtest

- Requested window: `2021-06-03` to `2026-06-03`
- Actual daily signal window: `2021-06-03T04:00:00Z` to `2026-06-01T04:00:00Z`
- VOO bars: `1254`
- Initial capital: `$100,000`
- Benchmark note: alpha-validation rows use the costed VOO buy-and-hold benchmark; benchmark-rotation rows use close-to-close VOO because that research tool has no fee/slippage model.

## Verdict

- Best no-leak research checkpoint: `ml_model_status_guard_21_126_no_leak` at 102.62% vs matched VOO 83.51%.
- Status: `provisional_checkpoint`, not production-promoted, because the excess return comes from one accepted alpha trade.
- Costed full-period VOO buy-and-hold: 80.24%, Sharpe 0.791, max DD 25.32%.
- Close-to-close VOO benchmark used by rotation diagnostics: 81.08%.
- Deterministic VOO strategies, HMM overlay, worker parity, and unguarded base rotation all fail to beat VOO.
- `ml_final_model_21_84_full_5y_diagnostic` is the best-looking raw number, but it is rejected as evidence because it uses a model trained through 2024 on a full 2021-2026 window.

## Ranked Results

| Strategy | Family | Window | Decision | Return | Matched VOO | Excess | Sharpe | Max DD | Trades | Notes |
|---|---|---|---|---:|---:|---:|---:|---:|---:|---|
| `ml_final_model_21_84_full_5y_diagnostic` | ml_meta_rotation | full_5y_diagnostic | `diagnostic_only` | 123.20% | 81.08% | 42.12% | 0.573 | 36.04% | 15 | Diagnostic only: beats VOO in full-window test, but this is in-sample/leaky and contradicted by 2025 OOS rejection. |
| `ml_model_status_guard_21_126_no_leak` | ml_meta_rotation | full_5y_walkforward_guarded | `provisional_checkpoint` | 102.62% | 83.51% | 19.11% | n/a | n/a | 1 | Best current research checkpoint. No alpha before candidate model; rejected model folds stay in VOO. Provisional because only one alpha trade. |
| `VOO buy_hold` | benchmark | full_5y | `benchmark` | 80.24% | 80.24% | 0.00% | 0.791 | 25.32% | 3 | Costed full-period buy-and-hold from alpha-validation engine. |
| `ml_final_model_42_126_full_5y_diagnostic` | ml_meta_rotation | full_5y_diagnostic | `diagnostic_only` | 81.08% | 81.08% | 0.00% | 0.757 | 25.32% | 0 | Diagnostic only: model artifact rejected, status guard produces benchmark-only behavior. |
| `ml_final_model_10_63_full_5y_diagnostic` | ml_meta_rotation | full_5y_diagnostic | `diagnostic_only` | 78.45% | 81.08% | -2.63% | 0.502 | 33.33% | 20 | Diagnostic only: trained through 2024; previously rejected versus checkpoint. |
| `ml_final_model_21_126_full_5y_diagnostic` | ml_meta_rotation | full_5y_diagnostic | `diagnostic_only` | 36.04% | 81.08% | -45.04% | 0.280 | 37.61% | 10 | Diagnostic only: final model trained through 2024, so full 5-year run contains in-sample/leaky history. |
| `ensemble_none` | hmm_regime | walk_forward_oos | `reject` | 5.48% | 64.20% | -58.72% | 1.079 | 2.05% | 62 | Below matched OOS VOO buy-and-hold; overlay promotion: total return deteriorated by more than 10 percent |
| `ensemble_overlay` | hmm_regime | walk_forward_oos | `reject` | 3.88% | 64.20% | -60.32% | 1.092 | 1.23% | 62 | Below matched OOS VOO buy-and-hold; overlay promotion: total return deteriorated by more than 10 percent |
| `xsec_momentum_top15` | xsec_momentum | full_5y_costed_multi | `reject` | 17.20% | 80.24% | -63.04% | 1.070 | 3.99% | 1072 | xsec universe size 4 below research minimum 50 |
| `worker_none` | worker_parity | worker_after_warmup | `reject` | 5.43% | 69.71% | -64.27% | 0.812 | 1.99% | 53 | Below matched worker VOO buy-and-hold. |
| `worker_overlay` | worker_parity | worker_after_warmup | `reject` | 4.67% | 69.71% | -65.04% | 0.971 | 1.33% | 51 | Below matched worker VOO buy-and-hold. |
| `kalman_z2` | kalman | full_5y_costed | `reject` | 6.45% | 80.24% | -73.79% | 0.776 | 2.49% | 604 | PBO 0.333 above 0.200; turnover increases without return improvement |
| `ma_crossover_20_50` | ma_crossover | full_5y_costed | `reject` | 6.12% | 80.24% | -74.12% | 0.726 | 2.76% | 648 | PBO 0.333 above 0.200; turnover increases without return improvement |
| `kalman_cointegration_z2` | kalman_cointegration | full_5y_costed_multi | `reject` | 0.00% | 80.24% | -80.24% | 0.000 | 0.00% | 0 | DSR 0.000 below 0.950; OOS trades 0 below 5; no drawdown-adjusted improvement over benchmark; pair sleeve requires offline approved cointegration candidate and live shortability gate before promotion |
| `benchmark_rotation_base_21_126` | ml_meta_rotation | full_5y_diagnostic | `reject` | -27.12% | 81.08% | -108.20% | -0.208 | 54.34% | 13 | No LightGBM filter; benchmark-funded base signals. |

## Report Artifacts

- `reports/batches/2026-06-03_voo_5y_candidate_backtest/regime_worker/voo_1day_regime_worker_5y.json`
- `reports/batches/2026-06-03_voo_5y_candidate_backtest/alpha_validation_voo_single/voo_1day_alpha_validation.md`
- `reports/batches/2026-06-03_voo_5y_candidate_backtest/alpha_validation_voo_spy_qqq_iwm/voo_spy_qqq_iwm_1day_alpha_validation.md`
- `reports/batches/2026-06-03_voo_5y_candidate_backtest/ml_meta/model_status_guard_no_leak_5y_voo.json`
- `reports/batches/2026-06-03_voo_5y_candidate_backtest/ml_meta/rotation_full_5y_21_126_final_model_diagnostic/benchmark_rotation_ml.md`
- `reports/batches/2026-06-03_voo_5y_candidate_backtest/ml_meta/rotation_full_5y_10_63_final_model_diagnostic/benchmark_rotation_ml.md`
- `reports/batches/2026-06-03_voo_5y_candidate_backtest/ml_meta/rotation_full_5y_21_84_final_model_diagnostic/benchmark_rotation_ml.md`
- `reports/batches/2026-06-03_voo_5y_candidate_backtest/ml_meta/rotation_full_5y_42_126_final_model_diagnostic/benchmark_rotation_ml.md`
