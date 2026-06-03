# Alpha Validation Report

- Generated: `2026-06-03T15:41:35Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2015-01-02` to `2026-06-01`
- Bars: `2869`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 351.93% | 14.17% | 0.837 | 0.790 | 0.417 | 34.00% | 0 | 1.000 |
| equal_weight | 1170.53% | 25.03% | 1.087 | 1.039 | 0.673 | 37.17% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| lgbm_ranker_h63_low | buy_hold | true | 459.70% | 16.34% | 0.924 | 0.881 | 0.484 | 33.73% | 1.000 | 0.067 | 49 | pass |
| lowvol_sleeve_low | buy_hold | false | 335.94% | 13.81% | 0.843 | 0.795 | 0.412 | 33.52% | 1.000 | 0.067 | 328 | turnover increases without return improvement |
| ranker_proxy_h63_low | buy_hold | true | 420.45% | 15.60% | 0.907 | 0.854 | 0.441 | 35.39% | 1.000 | 0.067 | 103 | pass |
| lgbm_ranker_h63_medium | buy_hold | false | 631.93% | 19.11% | 1.013 | 0.966 | 0.561 | 34.07% | 1.000 | 0.333 | 104 | PBO 0.333 above 0.200 |
| ranked_sleeve_medium | buy_hold | false | 474.41% | 16.60% | 0.944 | 0.888 | 0.482 | 34.45% | 1.000 | 0.333 | 561 | PBO 0.333 above 0.200 |
| ranker_proxy_h63_medium | buy_hold | false | 488.98% | 16.86% | 0.964 | 0.901 | 0.473 | 35.67% | 1.000 | 0.333 | 142 | PBO 0.333 above 0.200 |
| benchmark_tsmom_high | buy_hold | false | 447.24% | 16.11% | 0.916 | 0.866 | 0.462 | 34.89% | 1.000 | 0.333 | 396 | PBO 0.333 above 0.200 |
| composite_momentum_high | buy_hold | false | 421.90% | 15.63% | 0.916 | 0.854 | 0.462 | 33.83% | 1.000 | 0.333 | 703 | PBO 0.333 above 0.200 |
| lgbm_ranker_h63_high | buy_hold | false | 624.09% | 19.00% | 0.996 | 0.938 | 0.556 | 34.17% | 1.000 | 0.333 | 236 | PBO 0.333 above 0.200 |

## lgbm_ranker_h63_low

- Family: `agent_catalog_low`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `true`

### Metadata Audit

| Key | Value |
|---|---|
| active_weight | `0.05` |
| agent_strategy_deployment_status | `conservative_variant` |
| agent_strategy_family | `benchmark_lgbm_ranker_h63` |
| agent_strategy_key | `lgbm_ranker_h63_low` |
| agent_strategy_name | `LGBM h63 active sleeve low risk` |
| agent_strategy_paper_only | `true` |
| agent_strategy_promoted_checkpoint | `false` |
| agent_strategy_risk_profile | `low` |
| benchmark | `VOO` |
| benchmark_drawdown | `0` |
| benchmark_vol_20 | `0.0585979383` |
| benchmark_weight | `0.95` |
| candidate_count | `1` |
| engine | `daily_lgbm_ranker_sleeve` |
| point_in_time_universe | `false` |
| ranker_model_artifact_root | `../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts` |
| ranker_model_feature_count | `31` |
| ranker_model_feature_spec_version | `daily_ranker_v1` |
| ranker_model_loaded | `true` |
| ranker_model_path | `../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts/stocks_h63_s15_top3_reb63_z10/2018/model.txt` |
| ranker_model_sha256 | `dce2cb0d05a535e50158aeb0513a2e66b4f5afeee8f331b0c98965241e834bfe` |
| ranker_model_variant | `stocks_h63_s15_top3_reb63_z10` |
| ranker_model_year | `2018` |
| rebalance | `true` |
| sleeve_scale | `1` |
| turnover | `0.1` |
| turnover_band | `0.05` |

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 459.70% | 16.34% | 0.924 | 0.881 | 0.484 | 33.73% | 49 | 10.631 | - |
| stress_2x | 459.19% | 16.33% | 0.923 | 0.881 | 0.484 | 33.73% | 49 | 10.625 | - |
| stress_3x | 458.69% | 16.32% | 0.923 | 0.880 | 0.484 | 33.73% | 49 | 10.619 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.977 | -0.531 | -0.496 | 21.20% | - |
| 1 | 126-882 | 882-1134 | 0.884 | 0.585 | 0.394 | 21.20% | - |
| 2 | 252-1008 | 1008-1260 | 0.635 | 2.303 | 4.486 | 6.87% | - |
| 3 | 378-1134 | 1134-1386 | 1.086 | 0.472 | 0.319 | 33.73% | - |
| 4 | 504-1260 | 1260-1512 | 1.058 | 0.702 | 0.598 | 33.73% | - |
| 5 | 630-1386 | 1386-1638 | 0.571 | 2.492 | 4.669 | 9.74% | - |
| 6 | 756-1512 | 1512-1764 | 0.716 | 2.206 | 6.216 | 5.52% | - |
| 7 | 882-1638 | 1638-1890 | 0.933 | -0.419 | -0.436 | 23.32% | - |
| 8 | 1008-1764 | 1764-2016 | 1.224 | -0.647 | -0.718 | 24.04% | - |
| 9 | 1134-1890 | 1890-2142 | 0.611 | 1.174 | 1.555 | 15.48% | - |
| 10 | 1260-2016 | 2016-2268 | 0.500 | 2.033 | 3.173 | 10.16% | - |
| 11 | 1386-2142 | 2142-2394 | 0.973 | 2.349 | 3.053 | 10.16% | - |
| 12 | 1512-2268 | 2268-2520 | 0.731 | 1.839 | 3.011 | 9.01% | - |
| 13 | 1638-2394 | 2394-2646 | 0.815 | 0.822 | 0.889 | 17.29% | - |
| 14 | 1764-2520 | 2520-2772 | 0.699 | 1.396 | 1.567 | 17.29% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ranker_proxy_h63_low | 1.211 | -0.315 | 1 | ranker_proxy_h63_low | -0.315 | 3 | false |
| 1 | ranker_proxy_h63_low | 1.139 | 0.722 | 2 | lowvol_sleeve_low | 0.779 | 3 | false |
| 2 | ranker_proxy_h63_low | 0.596 | 4.733 | 2 | lowvol_sleeve_low | 5.075 | 3 | false |
| 3 | ranker_proxy_h63_low | 0.923 | 0.322 | 1 | ranker_proxy_h63_low | 0.322 | 3 | false |
| 4 | ranker_proxy_h63_low | 0.883 | 0.523 | 2 | lgbm_ranker_h63_low | 0.598 | 3 | false |
| 5 | ranker_proxy_h63_low | 0.382 | 4.307 | 3 | lgbm_ranker_h63_low | 4.669 | 3 | true |
| 6 | lgbm_ranker_h63_low | 0.451 | 6.216 | 2 | ranker_proxy_h63_low | 6.343 | 3 | false |
| 7 | lgbm_ranker_h63_low | 0.624 | -0.436 | 2 | ranker_proxy_h63_low | -0.351 | 3 | false |
| 8 | lgbm_ranker_h63_low | 0.851 | -0.718 | 1 | lgbm_ranker_h63_low | -0.718 | 3 | false |
| 9 | lgbm_ranker_h63_low | 0.377 | 1.555 | 1 | lgbm_ranker_h63_low | 1.555 | 3 | false |
| 10 | lgbm_ranker_h63_low | 0.296 | 3.173 | 1 | lgbm_ranker_h63_low | 3.173 | 3 | false |
| 11 | lgbm_ranker_h63_low | 0.737 | 3.053 | 1 | lgbm_ranker_h63_low | 3.053 | 3 | false |
| 12 | lgbm_ranker_h63_low | 0.507 | 3.011 | 2 | lowvol_sleeve_low | 3.739 | 3 | false |
| 13 | lgbm_ranker_h63_low | 0.573 | 0.889 | 1 | lgbm_ranker_h63_low | 0.889 | 3 | false |
| 14 | lgbm_ranker_h63_low | 0.487 | 1.567 | 1 | lgbm_ranker_h63_low | 1.567 | 3 | false |

## lowvol_sleeve_low

- Family: `agent_catalog_low`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - turnover increases without return improvement

### Metadata Audit

| Key | Value |
|---|---|
| action | `hold_targets` |
| agent_strategy_deployment_status | `rejected_diagnostic` |
| agent_strategy_family | `benchmark_lowvol` |
| agent_strategy_key | `lowvol_sleeve_low` |
| agent_strategy_name | `Low-volatility sleeve low risk` |
| agent_strategy_paper_only | `true` |
| agent_strategy_promoted_checkpoint | `false` |
| agent_strategy_risk_profile | `low` |
| engine | `composite_momentum_sleeve` |
| rebalance | `false` |

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 335.94% | 13.81% | 0.843 | 0.795 | 0.412 | 33.52% | 328 | 22.563 | - |
| stress_2x | 334.96% | 13.79% | 0.842 | 0.794 | 0.411 | 33.52% | 328 | 22.532 | - |
| stress_3x | 333.98% | 13.77% | 0.841 | 0.792 | 0.411 | 33.53% | 328 | 22.502 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.011 | -0.355 | -0.385 | 18.15% | - |
| 1 | 126-882 | 882-1134 | 0.902 | 0.978 | 0.779 | 18.15% | - |
| 2 | 252-1008 | 1008-1260 | 0.755 | 2.436 | 5.075 | 6.14% | - |
| 3 | 378-1134 | 1134-1386 | 1.277 | 0.415 | 0.256 | 33.52% | - |
| 4 | 504-1260 | 1260-1512 | 1.200 | 0.591 | 0.458 | 33.52% | - |
| 5 | 630-1386 | 1386-1638 | 0.622 | 2.374 | 4.395 | 8.83% | - |
| 6 | 756-1512 | 1512-1764 | 0.670 | 2.194 | 6.145 | 4.91% | - |
| 7 | 882-1638 | 1638-1890 | 0.878 | -0.450 | -0.439 | 22.00% | - |
| 8 | 1008-1764 | 1764-2016 | 1.156 | -0.735 | -0.735 | 24.00% | - |
| 9 | 1134-1890 | 1890-2142 | 0.548 | 0.751 | 0.798 | 16.17% | - |
| 10 | 1260-2016 | 2016-2268 | 0.410 | 1.663 | 2.162 | 10.14% | - |
| 11 | 1386-2142 | 2142-2394 | 0.780 | 2.281 | 2.617 | 10.14% | - |
| 12 | 1512-2268 | 2268-2520 | 0.605 | 2.070 | 3.739 | 7.03% | - |
| 13 | 1638-2394 | 2394-2646 | 0.662 | 0.748 | 0.698 | 18.54% | - |
| 14 | 1764-2520 | 2520-2772 | 0.541 | 1.008 | 0.945 | 18.54% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ranker_proxy_h63_low | 1.211 | -0.315 | 1 | ranker_proxy_h63_low | -0.315 | 3 | false |
| 1 | ranker_proxy_h63_low | 1.139 | 0.722 | 2 | lowvol_sleeve_low | 0.779 | 3 | false |
| 2 | ranker_proxy_h63_low | 0.596 | 4.733 | 2 | lowvol_sleeve_low | 5.075 | 3 | false |
| 3 | ranker_proxy_h63_low | 0.923 | 0.322 | 1 | ranker_proxy_h63_low | 0.322 | 3 | false |
| 4 | ranker_proxy_h63_low | 0.883 | 0.523 | 2 | lgbm_ranker_h63_low | 0.598 | 3 | false |
| 5 | ranker_proxy_h63_low | 0.382 | 4.307 | 3 | lgbm_ranker_h63_low | 4.669 | 3 | true |
| 6 | lgbm_ranker_h63_low | 0.451 | 6.216 | 2 | ranker_proxy_h63_low | 6.343 | 3 | false |
| 7 | lgbm_ranker_h63_low | 0.624 | -0.436 | 2 | ranker_proxy_h63_low | -0.351 | 3 | false |
| 8 | lgbm_ranker_h63_low | 0.851 | -0.718 | 1 | lgbm_ranker_h63_low | -0.718 | 3 | false |
| 9 | lgbm_ranker_h63_low | 0.377 | 1.555 | 1 | lgbm_ranker_h63_low | 1.555 | 3 | false |
| 10 | lgbm_ranker_h63_low | 0.296 | 3.173 | 1 | lgbm_ranker_h63_low | 3.173 | 3 | false |
| 11 | lgbm_ranker_h63_low | 0.737 | 3.053 | 1 | lgbm_ranker_h63_low | 3.053 | 3 | false |
| 12 | lgbm_ranker_h63_low | 0.507 | 3.011 | 2 | lowvol_sleeve_low | 3.739 | 3 | false |
| 13 | lgbm_ranker_h63_low | 0.573 | 0.889 | 1 | lgbm_ranker_h63_low | 0.889 | 3 | false |
| 14 | lgbm_ranker_h63_low | 0.487 | 1.567 | 1 | lgbm_ranker_h63_low | 1.567 | 3 | false |

## ranker_proxy_h63_low

- Family: `agent_catalog_low`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `true`

### Metadata Audit

| Key | Value |
|---|---|
| action | `hold_targets` |
| agent_strategy_deployment_status | `conservative_variant` |
| agent_strategy_family | `benchmark_ranker_proxy_h63` |
| agent_strategy_key | `ranker_proxy_h63_low` |
| agent_strategy_name | `Deterministic h63 proxy low risk` |
| agent_strategy_paper_only | `true` |
| agent_strategy_promoted_checkpoint | `false` |
| agent_strategy_risk_profile | `low` |
| engine | `composite_momentum_sleeve` |
| rebalance | `false` |

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 420.45% | 15.60% | 0.907 | 0.854 | 0.441 | 35.39% | 103 | 16.034 | - |
| stress_2x | 419.70% | 15.58% | 0.906 | 0.853 | 0.440 | 35.39% | 103 | 16.020 | - |
| stress_3x | 418.95% | 15.57% | 0.905 | 0.852 | 0.440 | 35.39% | 103 | 16.006 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.170 | -0.266 | -0.315 | 18.46% | - |
| 1 | 126-882 | 882-1134 | 1.054 | 0.902 | 0.722 | 18.46% | - |
| 2 | 252-1008 | 1008-1260 | 0.869 | 2.338 | 4.733 | 6.60% | - |
| 3 | 378-1134 | 1134-1386 | 1.341 | 0.489 | 0.322 | 35.39% | - |
| 4 | 504-1260 | 1260-1512 | 1.248 | 0.662 | 0.523 | 35.39% | - |
| 5 | 630-1386 | 1386-1638 | 0.675 | 2.370 | 4.307 | 9.53% | - |
| 6 | 756-1512 | 1512-1764 | 0.699 | 2.296 | 6.343 | 5.31% | - |
| 7 | 882-1638 | 1638-1890 | 0.914 | -0.323 | -0.351 | 22.58% | - |
| 8 | 1008-1764 | 1764-2016 | 1.205 | -0.726 | -0.733 | 24.54% | - |
| 9 | 1134-1890 | 1890-2142 | 0.622 | 0.878 | 0.968 | 16.45% | - |
| 10 | 1260-2016 | 2016-2268 | 0.455 | 1.900 | 2.729 | 9.97% | - |
| 11 | 1386-2142 | 2142-2394 | 0.865 | 2.233 | 2.817 | 9.97% | - |
| 12 | 1512-2268 | 2268-2520 | 0.690 | 1.789 | 2.811 | 8.72% | - |
| 13 | 1638-2394 | 2394-2646 | 0.745 | 0.579 | 0.521 | 18.97% | - |
| 14 | 1764-2520 | 2520-2772 | 0.601 | 1.043 | 1.009 | 18.97% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ranker_proxy_h63_low | 1.211 | -0.315 | 1 | ranker_proxy_h63_low | -0.315 | 3 | false |
| 1 | ranker_proxy_h63_low | 1.139 | 0.722 | 2 | lowvol_sleeve_low | 0.779 | 3 | false |
| 2 | ranker_proxy_h63_low | 0.596 | 4.733 | 2 | lowvol_sleeve_low | 5.075 | 3 | false |
| 3 | ranker_proxy_h63_low | 0.923 | 0.322 | 1 | ranker_proxy_h63_low | 0.322 | 3 | false |
| 4 | ranker_proxy_h63_low | 0.883 | 0.523 | 2 | lgbm_ranker_h63_low | 0.598 | 3 | false |
| 5 | ranker_proxy_h63_low | 0.382 | 4.307 | 3 | lgbm_ranker_h63_low | 4.669 | 3 | true |
| 6 | lgbm_ranker_h63_low | 0.451 | 6.216 | 2 | ranker_proxy_h63_low | 6.343 | 3 | false |
| 7 | lgbm_ranker_h63_low | 0.624 | -0.436 | 2 | ranker_proxy_h63_low | -0.351 | 3 | false |
| 8 | lgbm_ranker_h63_low | 0.851 | -0.718 | 1 | lgbm_ranker_h63_low | -0.718 | 3 | false |
| 9 | lgbm_ranker_h63_low | 0.377 | 1.555 | 1 | lgbm_ranker_h63_low | 1.555 | 3 | false |
| 10 | lgbm_ranker_h63_low | 0.296 | 3.173 | 1 | lgbm_ranker_h63_low | 3.173 | 3 | false |
| 11 | lgbm_ranker_h63_low | 0.737 | 3.053 | 1 | lgbm_ranker_h63_low | 3.053 | 3 | false |
| 12 | lgbm_ranker_h63_low | 0.507 | 3.011 | 2 | lowvol_sleeve_low | 3.739 | 3 | false |
| 13 | lgbm_ranker_h63_low | 0.573 | 0.889 | 1 | lgbm_ranker_h63_low | 0.889 | 3 | false |
| 14 | lgbm_ranker_h63_low | 0.487 | 1.567 | 1 | lgbm_ranker_h63_low | 1.567 | 3 | false |

## lgbm_ranker_h63_medium

- Family: `agent_catalog_medium`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.333 above 0.200

### Metadata Audit

| Key | Value |
|---|---|
| active_weight | `0.15` |
| agent_strategy_deployment_status | `promoted_research_checkpoint` |
| agent_strategy_family | `benchmark_lgbm_ranker_h63` |
| agent_strategy_key | `lgbm_ranker_h63_medium` |
| agent_strategy_name | `LGBM h63 active sleeve medium risk` |
| agent_strategy_paper_only | `true` |
| agent_strategy_promoted_checkpoint | `true` |
| agent_strategy_risk_profile | `medium` |
| benchmark | `VOO` |
| benchmark_drawdown | `0` |
| benchmark_vol_20 | `0.0585979383` |
| benchmark_weight | `0.85` |
| candidate_count | `3` |
| engine | `daily_lgbm_ranker_sleeve` |
| point_in_time_universe | `false` |
| ranker_model_artifact_root | `../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts` |
| ranker_model_feature_count | `31` |
| ranker_model_feature_spec_version | `daily_ranker_v1` |
| ranker_model_loaded | `true` |
| ranker_model_path | `../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts/stocks_h63_s15_top3_reb63_z10/2018/model.txt` |
| ranker_model_sha256 | `dce2cb0d05a535e50158aeb0513a2e66b4f5afeee8f331b0c98965241e834bfe` |
| ranker_model_variant | `stocks_h63_s15_top3_reb63_z10` |
| ranker_model_year | `2018` |
| rebalance | `true` |
| sleeve_scale | `1` |
| turnover | `0.3` |
| turnover_band | `0.05` |

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 631.93% | 19.11% | 1.013 | 0.966 | 0.561 | 34.07% | 104 | 26.352 | - |
| stress_2x | 630.56% | 19.09% | 1.012 | 0.965 | 0.560 | 34.08% | 105 | 26.320 | - |
| stress_3x | 629.19% | 19.07% | 1.012 | 0.964 | 0.560 | 34.08% | 105 | 26.287 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.977 | -0.270 | -0.299 | 22.85% | - |
| 1 | 126-882 | 882-1134 | 0.894 | 0.799 | 0.570 | 22.85% | - |
| 2 | 252-1008 | 1008-1260 | 0.707 | 2.282 | 4.349 | 7.33% | - |
| 3 | 378-1134 | 1134-1386 | 1.159 | 0.538 | 0.399 | 34.07% | - |
| 4 | 504-1260 | 1260-1512 | 1.128 | 0.748 | 0.671 | 34.07% | - |
| 5 | 630-1386 | 1386-1638 | 0.683 | 2.425 | 4.396 | 10.47% | - |
| 6 | 756-1512 | 1512-1764 | 0.677 | 2.322 | 8.568 | 4.38% | - |
| 7 | 882-1638 | 1638-1890 | 0.923 | -0.284 | -0.360 | 22.54% | - |
| 8 | 1008-1764 | 1764-2016 | 1.250 | -0.521 | -0.636 | 24.19% | - |
| 9 | 1134-1890 | 1890-2142 | 0.725 | 1.325 | 1.754 | 16.58% | - |
| 10 | 1260-2016 | 2016-2268 | 0.538 | 2.198 | 3.854 | 9.66% | - |
| 11 | 1386-2142 | 2142-2394 | 1.086 | 2.435 | 3.560 | 9.66% | - |
| 12 | 1512-2268 | 2268-2520 | 0.834 | 1.950 | 3.254 | 9.81% | - |
| 13 | 1638-2394 | 2394-2646 | 0.902 | 0.952 | 1.063 | 18.84% | - |
| 14 | 1764-2520 | 2520-2772 | 0.821 | 1.598 | 1.823 | 18.84% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ranker_proxy_h63_medium | 1.319 | -0.307 | 3 | ranked_sleeve_medium | -0.275 | 3 | true |
| 1 | ranker_proxy_h63_medium | 1.211 | 0.774 | 1 | ranker_proxy_h63_medium | 0.774 | 3 | false |
| 2 | ranker_proxy_h63_medium | 0.686 | 4.713 | 2 | ranked_sleeve_medium | 5.362 | 3 | false |
| 3 | ranker_proxy_h63_medium | 0.966 | 0.325 | 3 | ranked_sleeve_medium | 0.529 | 3 | true |
| 4 | ranker_proxy_h63_medium | 0.902 | 0.551 | 3 | ranked_sleeve_medium | 0.762 | 3 | true |
| 5 | ranked_sleeve_medium | 0.427 | 4.425 | 2 | ranker_proxy_h63_medium | 4.698 | 3 | false |
| 6 | ranked_sleeve_medium | 0.494 | 5.090 | 3 | lgbm_ranker_h63_medium | 8.568 | 3 | true |
| 7 | ranked_sleeve_medium | 0.683 | -0.425 | 3 | ranker_proxy_h63_medium | -0.284 | 3 | true |
| 8 | lgbm_ranker_h63_medium | 0.898 | -0.636 | 1 | lgbm_ranker_h63_medium | -0.636 | 3 | false |
| 9 | lgbm_ranker_h63_medium | 0.484 | 1.754 | 1 | lgbm_ranker_h63_medium | 1.754 | 3 | false |
| 10 | lgbm_ranker_h63_medium | 0.335 | 3.854 | 1 | lgbm_ranker_h63_medium | 3.854 | 3 | false |
| 11 | lgbm_ranker_h63_medium | 0.871 | 3.560 | 2 | ranked_sleeve_medium | 3.581 | 3 | false |
| 12 | lgbm_ranker_h63_medium | 0.609 | 3.254 | 1 | lgbm_ranker_h63_medium | 3.254 | 3 | false |
| 13 | lgbm_ranker_h63_medium | 0.658 | 1.063 | 1 | lgbm_ranker_h63_medium | 1.063 | 3 | false |
| 14 | lgbm_ranker_h63_medium | 0.640 | 1.823 | 1 | lgbm_ranker_h63_medium | 1.823 | 3 | false |

## ranked_sleeve_medium

- Family: `agent_catalog_medium`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.333 above 0.200

### Metadata Audit

| Key | Value |
|---|---|
| action | `hold_targets` |
| agent_strategy_deployment_status | `rejected_diagnostic` |
| agent_strategy_family | `benchmark_ranked_sleeve` |
| agent_strategy_key | `ranked_sleeve_medium` |
| agent_strategy_name | `Risk-budgeted ranked sleeve medium risk` |
| agent_strategy_paper_only | `true` |
| agent_strategy_promoted_checkpoint | `false` |
| agent_strategy_risk_profile | `medium` |
| engine | `composite_momentum_sleeve` |
| rebalance | `false` |

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 474.41% | 16.60% | 0.944 | 0.888 | 0.482 | 34.45% | 561 | 104.657 | - |
| stress_2x | 469.61% | 16.52% | 0.940 | 0.883 | 0.479 | 34.46% | 561 | 104.105 | - |
| stress_3x | 464.85% | 16.43% | 0.936 | 0.879 | 0.477 | 34.47% | 561 | 103.557 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.030 | -0.197 | -0.275 | 19.54% | - |
| 1 | 126-882 | 882-1134 | 1.054 | 0.660 | 0.494 | 19.54% | - |
| 2 | 252-1008 | 1008-1260 | 0.751 | 2.586 | 5.362 | 6.34% | - |
| 3 | 378-1134 | 1134-1386 | 1.182 | 0.666 | 0.529 | 34.45% | - |
| 4 | 504-1260 | 1260-1512 | 1.211 | 0.843 | 0.762 | 34.45% | - |
| 5 | 630-1386 | 1386-1638 | 0.720 | 2.323 | 4.425 | 10.20% | - |
| 6 | 756-1512 | 1512-1764 | 0.784 | 1.923 | 5.090 | 5.97% | - |
| 7 | 882-1638 | 1638-1890 | 1.023 | -0.399 | -0.425 | 20.86% | - |
| 8 | 1008-1764 | 1764-2016 | 1.278 | -0.675 | -0.643 | 24.04% | - |
| 9 | 1134-1890 | 1890-2142 | 0.658 | 0.867 | 0.966 | 14.98% | - |
| 10 | 1260-2016 | 2016-2268 | 0.477 | 1.572 | 2.311 | 9.98% | - |
| 11 | 1386-2142 | 2142-2394 | 0.807 | 2.275 | 3.581 | 9.98% | - |
| 12 | 1512-2268 | 2268-2520 | 0.696 | 1.785 | 2.976 | 9.70% | - |
| 13 | 1638-2394 | 2394-2646 | 0.820 | 0.649 | 0.603 | 18.01% | - |
| 14 | 1764-2520 | 2520-2772 | 0.640 | 1.216 | 1.184 | 18.01% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ranker_proxy_h63_medium | 1.319 | -0.307 | 3 | ranked_sleeve_medium | -0.275 | 3 | true |
| 1 | ranker_proxy_h63_medium | 1.211 | 0.774 | 1 | ranker_proxy_h63_medium | 0.774 | 3 | false |
| 2 | ranker_proxy_h63_medium | 0.686 | 4.713 | 2 | ranked_sleeve_medium | 5.362 | 3 | false |
| 3 | ranker_proxy_h63_medium | 0.966 | 0.325 | 3 | ranked_sleeve_medium | 0.529 | 3 | true |
| 4 | ranker_proxy_h63_medium | 0.902 | 0.551 | 3 | ranked_sleeve_medium | 0.762 | 3 | true |
| 5 | ranked_sleeve_medium | 0.427 | 4.425 | 2 | ranker_proxy_h63_medium | 4.698 | 3 | false |
| 6 | ranked_sleeve_medium | 0.494 | 5.090 | 3 | lgbm_ranker_h63_medium | 8.568 | 3 | true |
| 7 | ranked_sleeve_medium | 0.683 | -0.425 | 3 | ranker_proxy_h63_medium | -0.284 | 3 | true |
| 8 | lgbm_ranker_h63_medium | 0.898 | -0.636 | 1 | lgbm_ranker_h63_medium | -0.636 | 3 | false |
| 9 | lgbm_ranker_h63_medium | 0.484 | 1.754 | 1 | lgbm_ranker_h63_medium | 1.754 | 3 | false |
| 10 | lgbm_ranker_h63_medium | 0.335 | 3.854 | 1 | lgbm_ranker_h63_medium | 3.854 | 3 | false |
| 11 | lgbm_ranker_h63_medium | 0.871 | 3.560 | 2 | ranked_sleeve_medium | 3.581 | 3 | false |
| 12 | lgbm_ranker_h63_medium | 0.609 | 3.254 | 1 | lgbm_ranker_h63_medium | 3.254 | 3 | false |
| 13 | lgbm_ranker_h63_medium | 0.658 | 1.063 | 1 | lgbm_ranker_h63_medium | 1.063 | 3 | false |
| 14 | lgbm_ranker_h63_medium | 0.640 | 1.823 | 1 | lgbm_ranker_h63_medium | 1.823 | 3 | false |

## ranker_proxy_h63_medium

- Family: `agent_catalog_medium`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.333 above 0.200

### Metadata Audit

| Key | Value |
|---|---|
| action | `hold_targets` |
| agent_strategy_deployment_status | `promoted_research_checkpoint` |
| agent_strategy_family | `benchmark_ranker_proxy_h63` |
| agent_strategy_key | `ranker_proxy_h63_medium` |
| agent_strategy_name | `Deterministic h63 proxy medium risk` |
| agent_strategy_paper_only | `true` |
| agent_strategy_promoted_checkpoint | `true` |
| agent_strategy_risk_profile | `medium` |
| engine | `composite_momentum_sleeve` |
| rebalance | `false` |

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 488.98% | 16.86% | 0.964 | 0.901 | 0.473 | 35.67% | 142 | 31.400 | - |
| stress_2x | 487.53% | 16.83% | 0.963 | 0.900 | 0.472 | 35.67% | 142 | 31.351 | - |
| stress_3x | 486.08% | 16.81% | 0.962 | 0.899 | 0.471 | 35.67% | 142 | 31.302 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.294 | -0.250 | -0.307 | 18.12% | - |
| 1 | 126-882 | 882-1134 | 1.148 | 0.951 | 0.774 | 18.12% | - |
| 2 | 252-1008 | 1008-1260 | 0.961 | 2.295 | 4.713 | 6.42% | - |
| 3 | 378-1134 | 1134-1386 | 1.366 | 0.493 | 0.325 | 35.67% | - |
| 4 | 504-1260 | 1260-1512 | 1.254 | 0.688 | 0.551 | 35.67% | - |
| 5 | 630-1386 | 1386-1638 | 0.683 | 2.449 | 4.698 | 9.48% | - |
| 6 | 756-1512 | 1512-1764 | 0.711 | 2.361 | 6.339 | 5.80% | - |
| 7 | 882-1638 | 1638-1890 | 0.944 | -0.231 | -0.284 | 21.51% | - |
| 8 | 1008-1764 | 1764-2016 | 1.238 | -0.700 | -0.715 | 23.76% | - |
| 9 | 1134-1890 | 1890-2142 | 0.684 | 0.904 | 1.021 | 16.01% | - |
| 10 | 1260-2016 | 2016-2268 | 0.509 | 1.919 | 2.925 | 9.68% | - |
| 11 | 1386-2142 | 2142-2394 | 0.942 | 2.344 | 3.205 | 9.68% | - |
| 12 | 1512-2268 | 2268-2520 | 0.742 | 1.888 | 2.984 | 8.91% | - |
| 13 | 1638-2394 | 2394-2646 | 0.836 | 0.560 | 0.500 | 18.91% | - |
| 14 | 1764-2520 | 2520-2772 | 0.666 | 1.070 | 1.047 | 18.91% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ranker_proxy_h63_medium | 1.319 | -0.307 | 3 | ranked_sleeve_medium | -0.275 | 3 | true |
| 1 | ranker_proxy_h63_medium | 1.211 | 0.774 | 1 | ranker_proxy_h63_medium | 0.774 | 3 | false |
| 2 | ranker_proxy_h63_medium | 0.686 | 4.713 | 2 | ranked_sleeve_medium | 5.362 | 3 | false |
| 3 | ranker_proxy_h63_medium | 0.966 | 0.325 | 3 | ranked_sleeve_medium | 0.529 | 3 | true |
| 4 | ranker_proxy_h63_medium | 0.902 | 0.551 | 3 | ranked_sleeve_medium | 0.762 | 3 | true |
| 5 | ranked_sleeve_medium | 0.427 | 4.425 | 2 | ranker_proxy_h63_medium | 4.698 | 3 | false |
| 6 | ranked_sleeve_medium | 0.494 | 5.090 | 3 | lgbm_ranker_h63_medium | 8.568 | 3 | true |
| 7 | ranked_sleeve_medium | 0.683 | -0.425 | 3 | ranker_proxy_h63_medium | -0.284 | 3 | true |
| 8 | lgbm_ranker_h63_medium | 0.898 | -0.636 | 1 | lgbm_ranker_h63_medium | -0.636 | 3 | false |
| 9 | lgbm_ranker_h63_medium | 0.484 | 1.754 | 1 | lgbm_ranker_h63_medium | 1.754 | 3 | false |
| 10 | lgbm_ranker_h63_medium | 0.335 | 3.854 | 1 | lgbm_ranker_h63_medium | 3.854 | 3 | false |
| 11 | lgbm_ranker_h63_medium | 0.871 | 3.560 | 2 | ranked_sleeve_medium | 3.581 | 3 | false |
| 12 | lgbm_ranker_h63_medium | 0.609 | 3.254 | 1 | lgbm_ranker_h63_medium | 3.254 | 3 | false |
| 13 | lgbm_ranker_h63_medium | 0.658 | 1.063 | 1 | lgbm_ranker_h63_medium | 1.063 | 3 | false |
| 14 | lgbm_ranker_h63_medium | 0.640 | 1.823 | 1 | lgbm_ranker_h63_medium | 1.823 | 3 | false |

## benchmark_tsmom_high

- Family: `agent_catalog_high`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.333 above 0.200

### Metadata Audit

| Key | Value |
|---|---|
| action | `hold_targets` |
| agent_strategy_deployment_status | `rejected_diagnostic` |
| agent_strategy_family | `benchmark_tsmom` |
| agent_strategy_key | `benchmark_tsmom_high` |
| agent_strategy_name | `Benchmark-funded TSMOM high risk` |
| agent_strategy_paper_only | `true` |
| agent_strategy_promoted_checkpoint | `false` |
| agent_strategy_risk_profile | `high` |
| engine | `composite_momentum_sleeve` |
| rebalance | `false` |

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 447.24% | 16.11% | 0.916 | 0.866 | 0.462 | 34.89% | 396 | 55.143 | - |
| stress_2x | 444.83% | 16.06% | 0.914 | 0.864 | 0.460 | 34.90% | 396 | 54.992 | - |
| stress_3x | 442.43% | 16.02% | 0.912 | 0.862 | 0.459 | 34.91% | 396 | 54.842 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.091 | -0.219 | -0.279 | 19.38% | - |
| 1 | 126-882 | 882-1134 | 1.064 | 0.754 | 0.569 | 19.38% | - |
| 2 | 252-1008 | 1008-1260 | 0.782 | 2.277 | 4.678 | 6.68% | - |
| 3 | 378-1134 | 1134-1386 | 1.258 | 0.448 | 0.284 | 34.89% | - |
| 4 | 504-1260 | 1260-1512 | 1.212 | 0.696 | 0.578 | 34.89% | - |
| 5 | 630-1386 | 1386-1638 | 0.615 | 2.165 | 3.990 | 10.46% | - |
| 6 | 756-1512 | 1512-1764 | 0.683 | 1.788 | 4.476 | 6.23% | - |
| 7 | 882-1638 | 1638-1890 | 0.857 | -0.171 | -0.259 | 18.61% | - |
| 8 | 1008-1764 | 1764-2016 | 1.144 | -0.500 | -0.615 | 20.72% | - |
| 9 | 1134-1890 | 1890-2142 | 0.636 | 0.956 | 1.125 | 15.49% | - |
| 10 | 1260-2016 | 2016-2268 | 0.496 | 1.809 | 2.809 | 9.83% | - |
| 11 | 1386-2142 | 2142-2394 | 0.943 | 2.458 | 3.705 | 9.83% | - |
| 12 | 1512-2268 | 2268-2520 | 0.815 | 1.583 | 2.198 | 11.37% | - |
| 13 | 1638-2394 | 2394-2646 | 0.943 | 0.444 | 0.385 | 18.62% | - |
| 14 | 1764-2520 | 2520-2772 | 0.596 | 1.196 | 1.230 | 18.62% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_high | 1.158 | -0.279 | 3 | lgbm_ranker_h63_high | -0.132 | 3 | true |
| 1 | benchmark_tsmom_high | 1.206 | 0.569 | 3 | lgbm_ranker_h63_high | 0.782 | 3 | true |
| 2 | composite_momentum_high | 0.561 | 4.209 | 2 | benchmark_tsmom_high | 4.678 | 3 | false |
| 3 | composite_momentum_high | 0.901 | 0.480 | 2 | lgbm_ranker_h63_high | 0.588 | 3 | false |
| 4 | lgbm_ranker_h63_high | 0.903 | 0.867 | 1 | lgbm_ranker_h63_high | 0.867 | 3 | false |
| 5 | lgbm_ranker_h63_high | 0.478 | 3.812 | 3 | composite_momentum_high | 4.219 | 3 | true |
| 6 | lgbm_ranker_h63_high | 0.564 | 4.968 | 2 | composite_momentum_high | 5.675 | 3 | false |
| 7 | lgbm_ranker_h63_high | 0.700 | -0.504 | 3 | composite_momentum_high | -0.216 | 3 | true |
| 8 | lgbm_ranker_h63_high | 0.909 | -0.702 | 2 | benchmark_tsmom_high | -0.615 | 3 | false |
| 9 | composite_momentum_high | 0.482 | 0.680 | 3 | lgbm_ranker_h63_high | 1.742 | 3 | true |
| 10 | lgbm_ranker_h63_high | 0.322 | 4.557 | 1 | lgbm_ranker_h63_high | 4.557 | 3 | false |
| 11 | benchmark_tsmom_high | 0.782 | 3.705 | 1 | benchmark_tsmom_high | 3.705 | 3 | false |
| 12 | benchmark_tsmom_high | 0.629 | 2.198 | 1 | benchmark_tsmom_high | 2.198 | 3 | false |
| 13 | benchmark_tsmom_high | 0.729 | 0.385 | 2 | lgbm_ranker_h63_high | 0.691 | 3 | false |
| 14 | lgbm_ranker_h63_high | 0.466 | 1.473 | 1 | lgbm_ranker_h63_high | 1.473 | 3 | false |

## composite_momentum_high

- Family: `agent_catalog_high`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.333 above 0.200

### Metadata Audit

| Key | Value |
|---|---|
| action | `hold_targets` |
| agent_strategy_deployment_status | `rejected_diagnostic` |
| agent_strategy_family | `composite_momentum` |
| agent_strategy_key | `composite_momentum_high` |
| agent_strategy_name | `Composite momentum high risk` |
| agent_strategy_paper_only | `true` |
| agent_strategy_promoted_checkpoint | `false` |
| agent_strategy_risk_profile | `high` |
| engine | `composite_momentum_sleeve` |
| rebalance | `false` |

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 421.90% | 15.63% | 0.916 | 0.854 | 0.462 | 33.83% | 703 | 150.574 | - |
| stress_2x | 415.56% | 15.50% | 0.910 | 0.848 | 0.458 | 33.85% | 703 | 149.441 | - |
| stress_3x | 409.30% | 15.38% | 0.904 | 0.843 | 0.454 | 33.86% | 703 | 148.317 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.025 | -0.133 | -0.216 | 17.19% | - |
| 1 | 126-882 | 882-1134 | 0.915 | 0.889 | 0.731 | 17.19% | - |
| 2 | 252-1008 | 1008-1260 | 0.757 | 2.217 | 4.209 | 6.77% | - |
| 3 | 378-1134 | 1134-1386 | 1.216 | 0.616 | 0.480 | 33.83% | - |
| 4 | 504-1260 | 1260-1512 | 1.187 | 0.835 | 0.763 | 33.83% | - |
| 5 | 630-1386 | 1386-1638 | 0.721 | 2.366 | 4.219 | 10.39% | - |
| 6 | 756-1512 | 1512-1764 | 0.767 | 2.143 | 5.675 | 5.91% | - |
| 7 | 882-1638 | 1638-1890 | 0.981 | -0.145 | -0.216 | 19.51% | - |
| 8 | 1008-1764 | 1764-2016 | 1.274 | -0.727 | -0.733 | 22.24% | - |
| 9 | 1134-1890 | 1890-2142 | 0.752 | 0.656 | 0.680 | 15.83% | - |
| 10 | 1260-2016 | 2016-2268 | 0.513 | 1.708 | 2.653 | 9.38% | - |
| 11 | 1386-2142 | 2142-2394 | 0.888 | 2.209 | 3.341 | 9.38% | - |
| 12 | 1512-2268 | 2268-2520 | 0.697 | 1.406 | 1.737 | 12.48% | - |
| 13 | 1638-2394 | 2394-2646 | 0.777 | 0.369 | 0.288 | 18.11% | - |
| 14 | 1764-2520 | 2520-2772 | 0.493 | 1.137 | 1.063 | 18.11% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_high | 1.158 | -0.279 | 3 | lgbm_ranker_h63_high | -0.132 | 3 | true |
| 1 | benchmark_tsmom_high | 1.206 | 0.569 | 3 | lgbm_ranker_h63_high | 0.782 | 3 | true |
| 2 | composite_momentum_high | 0.561 | 4.209 | 2 | benchmark_tsmom_high | 4.678 | 3 | false |
| 3 | composite_momentum_high | 0.901 | 0.480 | 2 | lgbm_ranker_h63_high | 0.588 | 3 | false |
| 4 | lgbm_ranker_h63_high | 0.903 | 0.867 | 1 | lgbm_ranker_h63_high | 0.867 | 3 | false |
| 5 | lgbm_ranker_h63_high | 0.478 | 3.812 | 3 | composite_momentum_high | 4.219 | 3 | true |
| 6 | lgbm_ranker_h63_high | 0.564 | 4.968 | 2 | composite_momentum_high | 5.675 | 3 | false |
| 7 | lgbm_ranker_h63_high | 0.700 | -0.504 | 3 | composite_momentum_high | -0.216 | 3 | true |
| 8 | lgbm_ranker_h63_high | 0.909 | -0.702 | 2 | benchmark_tsmom_high | -0.615 | 3 | false |
| 9 | composite_momentum_high | 0.482 | 0.680 | 3 | lgbm_ranker_h63_high | 1.742 | 3 | true |
| 10 | lgbm_ranker_h63_high | 0.322 | 4.557 | 1 | lgbm_ranker_h63_high | 4.557 | 3 | false |
| 11 | benchmark_tsmom_high | 0.782 | 3.705 | 1 | benchmark_tsmom_high | 3.705 | 3 | false |
| 12 | benchmark_tsmom_high | 0.629 | 2.198 | 1 | benchmark_tsmom_high | 2.198 | 3 | false |
| 13 | benchmark_tsmom_high | 0.729 | 0.385 | 2 | lgbm_ranker_h63_high | 0.691 | 3 | false |
| 14 | lgbm_ranker_h63_high | 0.466 | 1.473 | 1 | lgbm_ranker_h63_high | 1.473 | 3 | false |

## lgbm_ranker_h63_high

- Family: `agent_catalog_high`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.333 above 0.200

### Metadata Audit

| Key | Value |
|---|---|
| active_weight | `0.25` |
| agent_strategy_deployment_status | `experimental_variant` |
| agent_strategy_family | `benchmark_lgbm_ranker_h63` |
| agent_strategy_key | `lgbm_ranker_h63_high` |
| agent_strategy_name | `LGBM h63 active sleeve high risk` |
| agent_strategy_paper_only | `true` |
| agent_strategy_promoted_checkpoint | `false` |
| agent_strategy_risk_profile | `high` |
| benchmark | `VOO` |
| benchmark_drawdown | `0` |
| benchmark_vol_20 | `0.0585979383` |
| benchmark_weight | `0.75` |
| candidate_count | `5` |
| engine | `daily_lgbm_ranker_sleeve` |
| point_in_time_universe | `false` |
| ranker_model_artifact_root | `../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts` |
| ranker_model_feature_count | `31` |
| ranker_model_feature_spec_version | `daily_ranker_v1` |
| ranker_model_loaded | `true` |
| ranker_model_path | `../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts/stocks_h63_s15_top3_reb63_z10/2018/model.txt` |
| ranker_model_sha256 | `dce2cb0d05a535e50158aeb0513a2e66b4f5afeee8f331b0c98965241e834bfe` |
| ranker_model_variant | `stocks_h63_s15_top3_reb63_z10` |
| ranker_model_year | `2018` |
| rebalance | `true` |
| sleeve_scale | `1` |
| turnover | `0.5` |
| turnover_band | `0.05` |

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 624.09% | 19.00% | 0.996 | 0.938 | 0.556 | 34.17% | 236 | 63.470 | - |
| stress_2x | 621.14% | 18.96% | 0.994 | 0.936 | 0.555 | 34.18% | 236 | 63.307 | - |
| stress_3x | 618.19% | 18.91% | 0.992 | 0.934 | 0.553 | 34.19% | 236 | 63.145 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.977 | -0.034 | -0.132 | 20.35% | - |
| 1 | 126-882 | 882-1134 | 0.946 | 0.917 | 0.782 | 20.35% | - |
| 2 | 252-1008 | 1008-1260 | 0.793 | 2.306 | 4.111 | 8.22% | - |
| 3 | 378-1134 | 1134-1386 | 1.246 | 0.697 | 0.588 | 34.17% | - |
| 4 | 504-1260 | 1260-1512 | 1.232 | 0.894 | 0.867 | 34.17% | - |
| 5 | 630-1386 | 1386-1638 | 0.749 | 2.193 | 3.812 | 11.33% | - |
| 6 | 756-1512 | 1512-1764 | 0.826 | 1.949 | 4.968 | 6.19% | - |
| 7 | 882-1638 | 1638-1890 | 0.991 | -0.536 | -0.504 | 24.47% | - |
| 8 | 1008-1764 | 1764-2016 | 1.264 | -0.739 | -0.702 | 28.93% | - |
| 9 | 1134-1890 | 1890-2142 | 0.672 | 1.435 | 1.742 | 19.71% | - |
| 10 | 1260-2016 | 2016-2268 | 0.507 | 2.478 | 4.557 | 9.86% | - |
| 11 | 1386-2142 | 2142-2394 | 1.019 | 1.838 | 2.753 | 9.86% | - |
| 12 | 1512-2268 | 2268-2520 | 0.850 | 1.395 | 2.086 | 10.26% | - |
| 13 | 1638-2394 | 2394-2646 | 0.855 | 0.724 | 0.691 | 20.63% | - |
| 14 | 1764-2520 | 2520-2772 | 0.677 | 1.408 | 1.473 | 20.63% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_high | 1.158 | -0.279 | 3 | lgbm_ranker_h63_high | -0.132 | 3 | true |
| 1 | benchmark_tsmom_high | 1.206 | 0.569 | 3 | lgbm_ranker_h63_high | 0.782 | 3 | true |
| 2 | composite_momentum_high | 0.561 | 4.209 | 2 | benchmark_tsmom_high | 4.678 | 3 | false |
| 3 | composite_momentum_high | 0.901 | 0.480 | 2 | lgbm_ranker_h63_high | 0.588 | 3 | false |
| 4 | lgbm_ranker_h63_high | 0.903 | 0.867 | 1 | lgbm_ranker_h63_high | 0.867 | 3 | false |
| 5 | lgbm_ranker_h63_high | 0.478 | 3.812 | 3 | composite_momentum_high | 4.219 | 3 | true |
| 6 | lgbm_ranker_h63_high | 0.564 | 4.968 | 2 | composite_momentum_high | 5.675 | 3 | false |
| 7 | lgbm_ranker_h63_high | 0.700 | -0.504 | 3 | composite_momentum_high | -0.216 | 3 | true |
| 8 | lgbm_ranker_h63_high | 0.909 | -0.702 | 2 | benchmark_tsmom_high | -0.615 | 3 | false |
| 9 | composite_momentum_high | 0.482 | 0.680 | 3 | lgbm_ranker_h63_high | 1.742 | 3 | true |
| 10 | lgbm_ranker_h63_high | 0.322 | 4.557 | 1 | lgbm_ranker_h63_high | 4.557 | 3 | false |
| 11 | benchmark_tsmom_high | 0.782 | 3.705 | 1 | benchmark_tsmom_high | 3.705 | 3 | false |
| 12 | benchmark_tsmom_high | 0.629 | 2.198 | 1 | benchmark_tsmom_high | 2.198 | 3 | false |
| 13 | benchmark_tsmom_high | 0.729 | 0.385 | 2 | lgbm_ranker_h63_high | 0.691 | 3 | false |
| 14 | lgbm_ranker_h63_high | 0.466 | 1.473 | 1 | lgbm_ranker_h63_high | 1.473 | 3 | false |
