# Alpha Validation Report

- Generated: `2026-06-03T15:46:14Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2016-01-04` to `2026-06-01`
- Bars: `2617`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 348.99% | 15.57% | 0.898 | 0.842 | 0.458 | 34.00% | 0 | 1.000 |
| equal_weight | 932.56% | 25.22% | 1.112 | 1.059 | 0.720 | 35.05% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| lgbm_ranker_h63_low | buy_hold | true | 456.06% | 17.97% | 0.990 | 0.939 | 0.533 | 33.73% | 1.000 | 0.077 | 49 | pass |
| lowvol_sleeve_low | buy_hold | false | 329.99% | 15.09% | 0.901 | 0.843 | 0.450 | 33.52% | 1.000 | 0.077 | 296 | turnover increases without return improvement |
| ranker_proxy_h63_low | buy_hold | true | 397.15% | 16.71% | 0.953 | 0.890 | 0.472 | 35.39% | 1.000 | 0.077 | 94 | pass |
| lgbm_ranker_h63_medium | buy_hold | false | 627.17% | 21.06% | 1.084 | 1.027 | 0.618 | 34.07% | 1.000 | 0.308 | 104 | PBO 0.308 above 0.200 |
| ranked_sleeve_medium | buy_hold | false | 450.44% | 17.86% | 0.995 | 0.934 | 0.518 | 34.45% | 1.000 | 0.308 | 507 | PBO 0.308 above 0.200 |
| ranker_proxy_h63_medium | buy_hold | false | 451.32% | 17.87% | 1.006 | 0.934 | 0.501 | 35.67% | 1.000 | 0.308 | 129 | PBO 0.308 above 0.200 |
| benchmark_tsmom_high | buy_hold | false | 423.35% | 17.28% | 0.960 | 0.903 | 0.495 | 34.89% | 1.000 | 0.231 | 360 | PBO 0.231 above 0.200 |
| composite_momentum_high | buy_hold | false | 394.06% | 16.64% | 0.957 | 0.888 | 0.492 | 33.83% | 1.000 | 0.231 | 645 | PBO 0.231 above 0.200 |
| lgbm_ranker_h63_high | buy_hold | false | 619.38% | 20.93% | 1.064 | 0.994 | 0.613 | 34.17% | 1.000 | 0.231 | 236 | PBO 0.231 above 0.200 |

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
| normal | 456.06% | 17.97% | 0.990 | 0.939 | 0.533 | 33.73% | 49 | 10.568 | - |
| stress_2x | 455.55% | 17.96% | 0.990 | 0.939 | 0.533 | 33.73% | 49 | 10.562 | - |
| stress_3x | 455.05% | 17.95% | 0.989 | 0.938 | 0.532 | 33.73% | 49 | 10.556 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.635 | 2.303 | 4.486 | 6.87% | - |
| 1 | 126-882 | 882-1134 | 1.086 | 0.472 | 0.319 | 33.73% | - |
| 2 | 252-1008 | 1008-1260 | 1.058 | 0.702 | 0.598 | 33.73% | - |
| 3 | 378-1134 | 1134-1386 | 0.571 | 2.492 | 4.669 | 9.74% | - |
| 4 | 504-1260 | 1260-1512 | 0.716 | 2.206 | 6.216 | 5.52% | - |
| 5 | 630-1386 | 1386-1638 | 0.933 | -0.419 | -0.436 | 23.32% | - |
| 6 | 756-1512 | 1512-1764 | 1.224 | -0.647 | -0.718 | 24.04% | - |
| 7 | 882-1638 | 1638-1890 | 0.611 | 1.174 | 1.555 | 15.48% | - |
| 8 | 1008-1764 | 1764-2016 | 0.500 | 2.033 | 3.173 | 10.16% | - |
| 9 | 1134-1890 | 1890-2142 | 0.973 | 2.349 | 3.053 | 10.16% | - |
| 10 | 1260-2016 | 2016-2268 | 0.731 | 1.839 | 3.011 | 9.01% | - |
| 11 | 1386-2142 | 2142-2394 | 0.815 | 0.822 | 0.889 | 17.29% | - |
| 12 | 1512-2268 | 2268-2520 | 0.699 | 1.396 | 1.567 | 17.29% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ranker_proxy_h63_low | 0.596 | 4.733 | 2 | lowvol_sleeve_low | 5.075 | 3 | false |
| 1 | ranker_proxy_h63_low | 0.923 | 0.322 | 1 | ranker_proxy_h63_low | 0.322 | 3 | false |
| 2 | ranker_proxy_h63_low | 0.883 | 0.523 | 2 | lgbm_ranker_h63_low | 0.598 | 3 | false |
| 3 | ranker_proxy_h63_low | 0.382 | 4.307 | 3 | lgbm_ranker_h63_low | 4.669 | 3 | true |
| 4 | lgbm_ranker_h63_low | 0.451 | 6.216 | 2 | ranker_proxy_h63_low | 6.343 | 3 | false |
| 5 | lgbm_ranker_h63_low | 0.624 | -0.436 | 2 | ranker_proxy_h63_low | -0.351 | 3 | false |
| 6 | lgbm_ranker_h63_low | 0.851 | -0.718 | 1 | lgbm_ranker_h63_low | -0.718 | 3 | false |
| 7 | lgbm_ranker_h63_low | 0.377 | 1.555 | 1 | lgbm_ranker_h63_low | 1.555 | 3 | false |
| 8 | lgbm_ranker_h63_low | 0.296 | 3.173 | 1 | lgbm_ranker_h63_low | 3.173 | 3 | false |
| 9 | lgbm_ranker_h63_low | 0.737 | 3.053 | 1 | lgbm_ranker_h63_low | 3.053 | 3 | false |
| 10 | lgbm_ranker_h63_low | 0.507 | 3.011 | 2 | lowvol_sleeve_low | 3.739 | 3 | false |
| 11 | lgbm_ranker_h63_low | 0.573 | 0.889 | 1 | lgbm_ranker_h63_low | 0.889 | 3 | false |
| 12 | lgbm_ranker_h63_low | 0.487 | 1.567 | 1 | lgbm_ranker_h63_low | 1.567 | 3 | false |

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
| normal | 329.99% | 15.09% | 0.901 | 0.843 | 0.450 | 33.52% | 296 | 21.209 | - |
| stress_2x | 329.11% | 15.06% | 0.900 | 0.842 | 0.449 | 33.52% | 296 | 21.183 | - |
| stress_3x | 328.23% | 15.04% | 0.899 | 0.841 | 0.449 | 33.53% | 296 | 21.157 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.755 | 2.436 | 5.075 | 6.14% | - |
| 1 | 126-882 | 882-1134 | 1.277 | 0.415 | 0.256 | 33.52% | - |
| 2 | 252-1008 | 1008-1260 | 1.200 | 0.591 | 0.458 | 33.52% | - |
| 3 | 378-1134 | 1134-1386 | 0.622 | 2.374 | 4.395 | 8.83% | - |
| 4 | 504-1260 | 1260-1512 | 0.670 | 2.194 | 6.145 | 4.91% | - |
| 5 | 630-1386 | 1386-1638 | 0.878 | -0.450 | -0.439 | 22.00% | - |
| 6 | 756-1512 | 1512-1764 | 1.156 | -0.735 | -0.735 | 24.00% | - |
| 7 | 882-1638 | 1638-1890 | 0.548 | 0.751 | 0.798 | 16.17% | - |
| 8 | 1008-1764 | 1764-2016 | 0.410 | 1.663 | 2.162 | 10.14% | - |
| 9 | 1134-1890 | 1890-2142 | 0.780 | 2.281 | 2.617 | 10.14% | - |
| 10 | 1260-2016 | 2016-2268 | 0.605 | 2.070 | 3.739 | 7.03% | - |
| 11 | 1386-2142 | 2142-2394 | 0.662 | 0.748 | 0.698 | 18.54% | - |
| 12 | 1512-2268 | 2268-2520 | 0.541 | 1.008 | 0.945 | 18.54% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ranker_proxy_h63_low | 0.596 | 4.733 | 2 | lowvol_sleeve_low | 5.075 | 3 | false |
| 1 | ranker_proxy_h63_low | 0.923 | 0.322 | 1 | ranker_proxy_h63_low | 0.322 | 3 | false |
| 2 | ranker_proxy_h63_low | 0.883 | 0.523 | 2 | lgbm_ranker_h63_low | 0.598 | 3 | false |
| 3 | ranker_proxy_h63_low | 0.382 | 4.307 | 3 | lgbm_ranker_h63_low | 4.669 | 3 | true |
| 4 | lgbm_ranker_h63_low | 0.451 | 6.216 | 2 | ranker_proxy_h63_low | 6.343 | 3 | false |
| 5 | lgbm_ranker_h63_low | 0.624 | -0.436 | 2 | ranker_proxy_h63_low | -0.351 | 3 | false |
| 6 | lgbm_ranker_h63_low | 0.851 | -0.718 | 1 | lgbm_ranker_h63_low | -0.718 | 3 | false |
| 7 | lgbm_ranker_h63_low | 0.377 | 1.555 | 1 | lgbm_ranker_h63_low | 1.555 | 3 | false |
| 8 | lgbm_ranker_h63_low | 0.296 | 3.173 | 1 | lgbm_ranker_h63_low | 3.173 | 3 | false |
| 9 | lgbm_ranker_h63_low | 0.737 | 3.053 | 1 | lgbm_ranker_h63_low | 3.053 | 3 | false |
| 10 | lgbm_ranker_h63_low | 0.507 | 3.011 | 2 | lowvol_sleeve_low | 3.739 | 3 | false |
| 11 | lgbm_ranker_h63_low | 0.573 | 0.889 | 1 | lgbm_ranker_h63_low | 0.889 | 3 | false |
| 12 | lgbm_ranker_h63_low | 0.487 | 1.567 | 1 | lgbm_ranker_h63_low | 1.567 | 3 | false |

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
| normal | 397.15% | 16.71% | 0.953 | 0.890 | 0.472 | 35.39% | 94 | 14.704 | - |
| stress_2x | 396.50% | 16.69% | 0.952 | 0.890 | 0.472 | 35.39% | 94 | 14.692 | - |
| stress_3x | 395.85% | 16.68% | 0.951 | 0.889 | 0.471 | 35.39% | 94 | 14.681 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.869 | 2.338 | 4.733 | 6.60% | - |
| 1 | 126-882 | 882-1134 | 1.341 | 0.489 | 0.322 | 35.39% | - |
| 2 | 252-1008 | 1008-1260 | 1.248 | 0.662 | 0.523 | 35.39% | - |
| 3 | 378-1134 | 1134-1386 | 0.675 | 2.370 | 4.307 | 9.53% | - |
| 4 | 504-1260 | 1260-1512 | 0.699 | 2.296 | 6.343 | 5.31% | - |
| 5 | 630-1386 | 1386-1638 | 0.914 | -0.323 | -0.351 | 22.58% | - |
| 6 | 756-1512 | 1512-1764 | 1.205 | -0.726 | -0.733 | 24.54% | - |
| 7 | 882-1638 | 1638-1890 | 0.622 | 0.878 | 0.968 | 16.45% | - |
| 8 | 1008-1764 | 1764-2016 | 0.455 | 1.900 | 2.729 | 9.97% | - |
| 9 | 1134-1890 | 1890-2142 | 0.865 | 2.233 | 2.817 | 9.97% | - |
| 10 | 1260-2016 | 2016-2268 | 0.690 | 1.789 | 2.811 | 8.72% | - |
| 11 | 1386-2142 | 2142-2394 | 0.745 | 0.579 | 0.521 | 18.97% | - |
| 12 | 1512-2268 | 2268-2520 | 0.601 | 1.043 | 1.009 | 18.97% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ranker_proxy_h63_low | 0.596 | 4.733 | 2 | lowvol_sleeve_low | 5.075 | 3 | false |
| 1 | ranker_proxy_h63_low | 0.923 | 0.322 | 1 | ranker_proxy_h63_low | 0.322 | 3 | false |
| 2 | ranker_proxy_h63_low | 0.883 | 0.523 | 2 | lgbm_ranker_h63_low | 0.598 | 3 | false |
| 3 | ranker_proxy_h63_low | 0.382 | 4.307 | 3 | lgbm_ranker_h63_low | 4.669 | 3 | true |
| 4 | lgbm_ranker_h63_low | 0.451 | 6.216 | 2 | ranker_proxy_h63_low | 6.343 | 3 | false |
| 5 | lgbm_ranker_h63_low | 0.624 | -0.436 | 2 | ranker_proxy_h63_low | -0.351 | 3 | false |
| 6 | lgbm_ranker_h63_low | 0.851 | -0.718 | 1 | lgbm_ranker_h63_low | -0.718 | 3 | false |
| 7 | lgbm_ranker_h63_low | 0.377 | 1.555 | 1 | lgbm_ranker_h63_low | 1.555 | 3 | false |
| 8 | lgbm_ranker_h63_low | 0.296 | 3.173 | 1 | lgbm_ranker_h63_low | 3.173 | 3 | false |
| 9 | lgbm_ranker_h63_low | 0.737 | 3.053 | 1 | lgbm_ranker_h63_low | 3.053 | 3 | false |
| 10 | lgbm_ranker_h63_low | 0.507 | 3.011 | 2 | lowvol_sleeve_low | 3.739 | 3 | false |
| 11 | lgbm_ranker_h63_low | 0.573 | 0.889 | 1 | lgbm_ranker_h63_low | 0.889 | 3 | false |
| 12 | lgbm_ranker_h63_low | 0.487 | 1.567 | 1 | lgbm_ranker_h63_low | 1.567 | 3 | false |

## lgbm_ranker_h63_medium

- Family: `agent_catalog_medium`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.308 above 0.200

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
| normal | 627.17% | 21.06% | 1.084 | 1.027 | 0.618 | 34.07% | 104 | 26.187 | - |
| stress_2x | 625.81% | 21.04% | 1.083 | 1.027 | 0.617 | 34.08% | 105 | 26.155 | - |
| stress_3x | 624.44% | 21.02% | 1.082 | 1.026 | 0.617 | 34.08% | 105 | 26.122 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.707 | 2.282 | 4.349 | 7.33% | - |
| 1 | 126-882 | 882-1134 | 1.159 | 0.538 | 0.399 | 34.07% | - |
| 2 | 252-1008 | 1008-1260 | 1.128 | 0.748 | 0.671 | 34.07% | - |
| 3 | 378-1134 | 1134-1386 | 0.683 | 2.425 | 4.396 | 10.47% | - |
| 4 | 504-1260 | 1260-1512 | 0.677 | 2.322 | 8.568 | 4.38% | - |
| 5 | 630-1386 | 1386-1638 | 0.923 | -0.284 | -0.360 | 22.54% | - |
| 6 | 756-1512 | 1512-1764 | 1.250 | -0.521 | -0.636 | 24.19% | - |
| 7 | 882-1638 | 1638-1890 | 0.725 | 1.325 | 1.754 | 16.58% | - |
| 8 | 1008-1764 | 1764-2016 | 0.538 | 2.198 | 3.854 | 9.66% | - |
| 9 | 1134-1890 | 1890-2142 | 1.086 | 2.435 | 3.560 | 9.66% | - |
| 10 | 1260-2016 | 2016-2268 | 0.834 | 1.950 | 3.254 | 9.81% | - |
| 11 | 1386-2142 | 2142-2394 | 0.902 | 0.952 | 1.063 | 18.84% | - |
| 12 | 1512-2268 | 2268-2520 | 0.821 | 1.598 | 1.823 | 18.84% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ranker_proxy_h63_medium | 0.686 | 4.713 | 2 | ranked_sleeve_medium | 5.362 | 3 | false |
| 1 | ranker_proxy_h63_medium | 0.966 | 0.325 | 3 | ranked_sleeve_medium | 0.529 | 3 | true |
| 2 | ranker_proxy_h63_medium | 0.902 | 0.551 | 3 | ranked_sleeve_medium | 0.762 | 3 | true |
| 3 | ranked_sleeve_medium | 0.427 | 4.425 | 2 | ranker_proxy_h63_medium | 4.698 | 3 | false |
| 4 | ranked_sleeve_medium | 0.494 | 5.090 | 3 | lgbm_ranker_h63_medium | 8.568 | 3 | true |
| 5 | ranked_sleeve_medium | 0.683 | -0.425 | 3 | ranker_proxy_h63_medium | -0.284 | 3 | true |
| 6 | lgbm_ranker_h63_medium | 0.898 | -0.636 | 1 | lgbm_ranker_h63_medium | -0.636 | 3 | false |
| 7 | lgbm_ranker_h63_medium | 0.484 | 1.754 | 1 | lgbm_ranker_h63_medium | 1.754 | 3 | false |
| 8 | lgbm_ranker_h63_medium | 0.335 | 3.854 | 1 | lgbm_ranker_h63_medium | 3.854 | 3 | false |
| 9 | lgbm_ranker_h63_medium | 0.871 | 3.560 | 2 | ranked_sleeve_medium | 3.581 | 3 | false |
| 10 | lgbm_ranker_h63_medium | 0.609 | 3.254 | 1 | lgbm_ranker_h63_medium | 3.254 | 3 | false |
| 11 | lgbm_ranker_h63_medium | 0.658 | 1.063 | 1 | lgbm_ranker_h63_medium | 1.063 | 3 | false |
| 12 | lgbm_ranker_h63_medium | 0.640 | 1.823 | 1 | lgbm_ranker_h63_medium | 1.823 | 3 | false |

## ranked_sleeve_medium

- Family: `agent_catalog_medium`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.308 above 0.200

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
| normal | 450.44% | 17.86% | 0.995 | 0.934 | 0.518 | 34.45% | 507 | 96.483 | - |
| stress_2x | 446.25% | 17.77% | 0.991 | 0.931 | 0.516 | 34.46% | 507 | 96.029 | - |
| stress_3x | 442.10% | 17.68% | 0.987 | 0.927 | 0.513 | 34.47% | 507 | 95.577 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.751 | 2.586 | 5.362 | 6.34% | - |
| 1 | 126-882 | 882-1134 | 1.182 | 0.666 | 0.529 | 34.45% | - |
| 2 | 252-1008 | 1008-1260 | 1.211 | 0.843 | 0.762 | 34.45% | - |
| 3 | 378-1134 | 1134-1386 | 0.720 | 2.323 | 4.425 | 10.20% | - |
| 4 | 504-1260 | 1260-1512 | 0.784 | 1.923 | 5.090 | 5.97% | - |
| 5 | 630-1386 | 1386-1638 | 1.023 | -0.399 | -0.425 | 20.86% | - |
| 6 | 756-1512 | 1512-1764 | 1.278 | -0.675 | -0.643 | 24.04% | - |
| 7 | 882-1638 | 1638-1890 | 0.658 | 0.867 | 0.966 | 14.98% | - |
| 8 | 1008-1764 | 1764-2016 | 0.477 | 1.572 | 2.311 | 9.98% | - |
| 9 | 1134-1890 | 1890-2142 | 0.807 | 2.275 | 3.581 | 9.98% | - |
| 10 | 1260-2016 | 2016-2268 | 0.696 | 1.785 | 2.976 | 9.70% | - |
| 11 | 1386-2142 | 2142-2394 | 0.820 | 0.649 | 0.603 | 18.01% | - |
| 12 | 1512-2268 | 2268-2520 | 0.640 | 1.216 | 1.184 | 18.01% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ranker_proxy_h63_medium | 0.686 | 4.713 | 2 | ranked_sleeve_medium | 5.362 | 3 | false |
| 1 | ranker_proxy_h63_medium | 0.966 | 0.325 | 3 | ranked_sleeve_medium | 0.529 | 3 | true |
| 2 | ranker_proxy_h63_medium | 0.902 | 0.551 | 3 | ranked_sleeve_medium | 0.762 | 3 | true |
| 3 | ranked_sleeve_medium | 0.427 | 4.425 | 2 | ranker_proxy_h63_medium | 4.698 | 3 | false |
| 4 | ranked_sleeve_medium | 0.494 | 5.090 | 3 | lgbm_ranker_h63_medium | 8.568 | 3 | true |
| 5 | ranked_sleeve_medium | 0.683 | -0.425 | 3 | ranker_proxy_h63_medium | -0.284 | 3 | true |
| 6 | lgbm_ranker_h63_medium | 0.898 | -0.636 | 1 | lgbm_ranker_h63_medium | -0.636 | 3 | false |
| 7 | lgbm_ranker_h63_medium | 0.484 | 1.754 | 1 | lgbm_ranker_h63_medium | 1.754 | 3 | false |
| 8 | lgbm_ranker_h63_medium | 0.335 | 3.854 | 1 | lgbm_ranker_h63_medium | 3.854 | 3 | false |
| 9 | lgbm_ranker_h63_medium | 0.871 | 3.560 | 2 | ranked_sleeve_medium | 3.581 | 3 | false |
| 10 | lgbm_ranker_h63_medium | 0.609 | 3.254 | 1 | lgbm_ranker_h63_medium | 3.254 | 3 | false |
| 11 | lgbm_ranker_h63_medium | 0.658 | 1.063 | 1 | lgbm_ranker_h63_medium | 1.063 | 3 | false |
| 12 | lgbm_ranker_h63_medium | 0.640 | 1.823 | 1 | lgbm_ranker_h63_medium | 1.823 | 3 | false |

## ranker_proxy_h63_medium

- Family: `agent_catalog_medium`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.308 above 0.200

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
| normal | 451.32% | 17.87% | 1.006 | 0.934 | 0.501 | 35.67% | 129 | 28.372 | - |
| stress_2x | 450.08% | 17.85% | 1.004 | 0.933 | 0.500 | 35.67% | 129 | 28.332 | - |
| stress_3x | 448.85% | 17.82% | 1.003 | 0.932 | 0.500 | 35.67% | 129 | 28.292 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.961 | 2.295 | 4.713 | 6.42% | - |
| 1 | 126-882 | 882-1134 | 1.366 | 0.493 | 0.325 | 35.67% | - |
| 2 | 252-1008 | 1008-1260 | 1.254 | 0.688 | 0.551 | 35.67% | - |
| 3 | 378-1134 | 1134-1386 | 0.683 | 2.449 | 4.698 | 9.48% | - |
| 4 | 504-1260 | 1260-1512 | 0.711 | 2.361 | 6.339 | 5.80% | - |
| 5 | 630-1386 | 1386-1638 | 0.944 | -0.231 | -0.284 | 21.51% | - |
| 6 | 756-1512 | 1512-1764 | 1.238 | -0.700 | -0.715 | 23.76% | - |
| 7 | 882-1638 | 1638-1890 | 0.684 | 0.904 | 1.021 | 16.01% | - |
| 8 | 1008-1764 | 1764-2016 | 0.509 | 1.919 | 2.925 | 9.68% | - |
| 9 | 1134-1890 | 1890-2142 | 0.942 | 2.344 | 3.205 | 9.68% | - |
| 10 | 1260-2016 | 2016-2268 | 0.742 | 1.888 | 2.984 | 8.91% | - |
| 11 | 1386-2142 | 2142-2394 | 0.836 | 0.560 | 0.500 | 18.91% | - |
| 12 | 1512-2268 | 2268-2520 | 0.666 | 1.070 | 1.047 | 18.91% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ranker_proxy_h63_medium | 0.686 | 4.713 | 2 | ranked_sleeve_medium | 5.362 | 3 | false |
| 1 | ranker_proxy_h63_medium | 0.966 | 0.325 | 3 | ranked_sleeve_medium | 0.529 | 3 | true |
| 2 | ranker_proxy_h63_medium | 0.902 | 0.551 | 3 | ranked_sleeve_medium | 0.762 | 3 | true |
| 3 | ranked_sleeve_medium | 0.427 | 4.425 | 2 | ranker_proxy_h63_medium | 4.698 | 3 | false |
| 4 | ranked_sleeve_medium | 0.494 | 5.090 | 3 | lgbm_ranker_h63_medium | 8.568 | 3 | true |
| 5 | ranked_sleeve_medium | 0.683 | -0.425 | 3 | ranker_proxy_h63_medium | -0.284 | 3 | true |
| 6 | lgbm_ranker_h63_medium | 0.898 | -0.636 | 1 | lgbm_ranker_h63_medium | -0.636 | 3 | false |
| 7 | lgbm_ranker_h63_medium | 0.484 | 1.754 | 1 | lgbm_ranker_h63_medium | 1.754 | 3 | false |
| 8 | lgbm_ranker_h63_medium | 0.335 | 3.854 | 1 | lgbm_ranker_h63_medium | 3.854 | 3 | false |
| 9 | lgbm_ranker_h63_medium | 0.871 | 3.560 | 2 | ranked_sleeve_medium | 3.581 | 3 | false |
| 10 | lgbm_ranker_h63_medium | 0.609 | 3.254 | 1 | lgbm_ranker_h63_medium | 3.254 | 3 | false |
| 11 | lgbm_ranker_h63_medium | 0.658 | 1.063 | 1 | lgbm_ranker_h63_medium | 1.063 | 3 | false |
| 12 | lgbm_ranker_h63_medium | 0.640 | 1.823 | 1 | lgbm_ranker_h63_medium | 1.823 | 3 | false |

## benchmark_tsmom_high

- Family: `agent_catalog_high`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.231 above 0.200

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
| normal | 423.35% | 17.28% | 0.960 | 0.903 | 0.495 | 34.89% | 360 | 51.151 | - |
| stress_2x | 421.22% | 17.24% | 0.958 | 0.902 | 0.494 | 34.90% | 360 | 51.024 | - |
| stress_3x | 419.09% | 17.19% | 0.956 | 0.899 | 0.492 | 34.91% | 360 | 50.897 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.782 | 2.277 | 4.678 | 6.68% | - |
| 1 | 126-882 | 882-1134 | 1.258 | 0.448 | 0.284 | 34.89% | - |
| 2 | 252-1008 | 1008-1260 | 1.212 | 0.696 | 0.578 | 34.89% | - |
| 3 | 378-1134 | 1134-1386 | 0.615 | 2.165 | 3.990 | 10.46% | - |
| 4 | 504-1260 | 1260-1512 | 0.683 | 1.788 | 4.476 | 6.23% | - |
| 5 | 630-1386 | 1386-1638 | 0.857 | -0.171 | -0.259 | 18.61% | - |
| 6 | 756-1512 | 1512-1764 | 1.144 | -0.500 | -0.615 | 20.72% | - |
| 7 | 882-1638 | 1638-1890 | 0.636 | 0.956 | 1.125 | 15.49% | - |
| 8 | 1008-1764 | 1764-2016 | 0.496 | 1.809 | 2.809 | 9.83% | - |
| 9 | 1134-1890 | 1890-2142 | 0.943 | 2.458 | 3.705 | 9.83% | - |
| 10 | 1260-2016 | 2016-2268 | 0.815 | 1.583 | 2.198 | 11.37% | - |
| 11 | 1386-2142 | 2142-2394 | 0.943 | 0.444 | 0.385 | 18.62% | - |
| 12 | 1512-2268 | 2268-2520 | 0.596 | 1.196 | 1.230 | 18.62% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | composite_momentum_high | 0.561 | 4.209 | 2 | benchmark_tsmom_high | 4.678 | 3 | false |
| 1 | composite_momentum_high | 0.901 | 0.480 | 2 | lgbm_ranker_h63_high | 0.588 | 3 | false |
| 2 | lgbm_ranker_h63_high | 0.903 | 0.867 | 1 | lgbm_ranker_h63_high | 0.867 | 3 | false |
| 3 | lgbm_ranker_h63_high | 0.478 | 3.812 | 3 | composite_momentum_high | 4.219 | 3 | true |
| 4 | lgbm_ranker_h63_high | 0.564 | 4.968 | 2 | composite_momentum_high | 5.675 | 3 | false |
| 5 | lgbm_ranker_h63_high | 0.700 | -0.504 | 3 | composite_momentum_high | -0.216 | 3 | true |
| 6 | lgbm_ranker_h63_high | 0.909 | -0.702 | 2 | benchmark_tsmom_high | -0.615 | 3 | false |
| 7 | composite_momentum_high | 0.482 | 0.680 | 3 | lgbm_ranker_h63_high | 1.742 | 3 | true |
| 8 | lgbm_ranker_h63_high | 0.322 | 4.557 | 1 | lgbm_ranker_h63_high | 4.557 | 3 | false |
| 9 | benchmark_tsmom_high | 0.782 | 3.705 | 1 | benchmark_tsmom_high | 3.705 | 3 | false |
| 10 | benchmark_tsmom_high | 0.629 | 2.198 | 1 | benchmark_tsmom_high | 2.198 | 3 | false |
| 11 | benchmark_tsmom_high | 0.729 | 0.385 | 2 | lgbm_ranker_h63_high | 0.691 | 3 | false |
| 12 | lgbm_ranker_h63_high | 0.466 | 1.473 | 1 | lgbm_ranker_h63_high | 1.473 | 3 | false |

## composite_momentum_high

- Family: `agent_catalog_high`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.231 above 0.200

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
| normal | 394.06% | 16.64% | 0.957 | 0.888 | 0.492 | 33.83% | 645 | 139.327 | - |
| stress_2x | 388.38% | 16.51% | 0.951 | 0.883 | 0.488 | 33.85% | 645 | 138.346 | - |
| stress_3x | 382.77% | 16.38% | 0.944 | 0.877 | 0.484 | 33.86% | 645 | 137.373 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.757 | 2.217 | 4.209 | 6.77% | - |
| 1 | 126-882 | 882-1134 | 1.216 | 0.616 | 0.480 | 33.83% | - |
| 2 | 252-1008 | 1008-1260 | 1.187 | 0.835 | 0.763 | 33.83% | - |
| 3 | 378-1134 | 1134-1386 | 0.721 | 2.366 | 4.219 | 10.39% | - |
| 4 | 504-1260 | 1260-1512 | 0.767 | 2.143 | 5.675 | 5.91% | - |
| 5 | 630-1386 | 1386-1638 | 0.981 | -0.145 | -0.216 | 19.51% | - |
| 6 | 756-1512 | 1512-1764 | 1.274 | -0.727 | -0.733 | 22.24% | - |
| 7 | 882-1638 | 1638-1890 | 0.752 | 0.656 | 0.680 | 15.83% | - |
| 8 | 1008-1764 | 1764-2016 | 0.513 | 1.708 | 2.653 | 9.38% | - |
| 9 | 1134-1890 | 1890-2142 | 0.888 | 2.209 | 3.341 | 9.38% | - |
| 10 | 1260-2016 | 2016-2268 | 0.697 | 1.406 | 1.737 | 12.48% | - |
| 11 | 1386-2142 | 2142-2394 | 0.777 | 0.369 | 0.288 | 18.11% | - |
| 12 | 1512-2268 | 2268-2520 | 0.493 | 1.137 | 1.063 | 18.11% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | composite_momentum_high | 0.561 | 4.209 | 2 | benchmark_tsmom_high | 4.678 | 3 | false |
| 1 | composite_momentum_high | 0.901 | 0.480 | 2 | lgbm_ranker_h63_high | 0.588 | 3 | false |
| 2 | lgbm_ranker_h63_high | 0.903 | 0.867 | 1 | lgbm_ranker_h63_high | 0.867 | 3 | false |
| 3 | lgbm_ranker_h63_high | 0.478 | 3.812 | 3 | composite_momentum_high | 4.219 | 3 | true |
| 4 | lgbm_ranker_h63_high | 0.564 | 4.968 | 2 | composite_momentum_high | 5.675 | 3 | false |
| 5 | lgbm_ranker_h63_high | 0.700 | -0.504 | 3 | composite_momentum_high | -0.216 | 3 | true |
| 6 | lgbm_ranker_h63_high | 0.909 | -0.702 | 2 | benchmark_tsmom_high | -0.615 | 3 | false |
| 7 | composite_momentum_high | 0.482 | 0.680 | 3 | lgbm_ranker_h63_high | 1.742 | 3 | true |
| 8 | lgbm_ranker_h63_high | 0.322 | 4.557 | 1 | lgbm_ranker_h63_high | 4.557 | 3 | false |
| 9 | benchmark_tsmom_high | 0.782 | 3.705 | 1 | benchmark_tsmom_high | 3.705 | 3 | false |
| 10 | benchmark_tsmom_high | 0.629 | 2.198 | 1 | benchmark_tsmom_high | 2.198 | 3 | false |
| 11 | benchmark_tsmom_high | 0.729 | 0.385 | 2 | lgbm_ranker_h63_high | 0.691 | 3 | false |
| 12 | lgbm_ranker_h63_high | 0.466 | 1.473 | 1 | lgbm_ranker_h63_high | 1.473 | 3 | false |

## lgbm_ranker_h63_high

- Family: `agent_catalog_high`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.231 above 0.200

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
| normal | 619.38% | 20.93% | 1.064 | 0.994 | 0.613 | 34.17% | 236 | 63.063 | - |
| stress_2x | 616.43% | 20.89% | 1.062 | 0.993 | 0.611 | 34.18% | 236 | 62.901 | - |
| stress_3x | 613.50% | 20.84% | 1.060 | 0.991 | 0.610 | 34.19% | 236 | 62.740 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.793 | 2.306 | 4.111 | 8.22% | - |
| 1 | 126-882 | 882-1134 | 1.246 | 0.697 | 0.588 | 34.17% | - |
| 2 | 252-1008 | 1008-1260 | 1.232 | 0.894 | 0.867 | 34.17% | - |
| 3 | 378-1134 | 1134-1386 | 0.749 | 2.193 | 3.812 | 11.33% | - |
| 4 | 504-1260 | 1260-1512 | 0.826 | 1.949 | 4.968 | 6.19% | - |
| 5 | 630-1386 | 1386-1638 | 0.991 | -0.536 | -0.504 | 24.47% | - |
| 6 | 756-1512 | 1512-1764 | 1.264 | -0.739 | -0.702 | 28.93% | - |
| 7 | 882-1638 | 1638-1890 | 0.672 | 1.435 | 1.742 | 19.71% | - |
| 8 | 1008-1764 | 1764-2016 | 0.507 | 2.478 | 4.557 | 9.86% | - |
| 9 | 1134-1890 | 1890-2142 | 1.019 | 1.838 | 2.753 | 9.86% | - |
| 10 | 1260-2016 | 2016-2268 | 0.850 | 1.395 | 2.086 | 10.26% | - |
| 11 | 1386-2142 | 2142-2394 | 0.855 | 0.724 | 0.691 | 20.63% | - |
| 12 | 1512-2268 | 2268-2520 | 0.677 | 1.408 | 1.473 | 20.63% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | composite_momentum_high | 0.561 | 4.209 | 2 | benchmark_tsmom_high | 4.678 | 3 | false |
| 1 | composite_momentum_high | 0.901 | 0.480 | 2 | lgbm_ranker_h63_high | 0.588 | 3 | false |
| 2 | lgbm_ranker_h63_high | 0.903 | 0.867 | 1 | lgbm_ranker_h63_high | 0.867 | 3 | false |
| 3 | lgbm_ranker_h63_high | 0.478 | 3.812 | 3 | composite_momentum_high | 4.219 | 3 | true |
| 4 | lgbm_ranker_h63_high | 0.564 | 4.968 | 2 | composite_momentum_high | 5.675 | 3 | false |
| 5 | lgbm_ranker_h63_high | 0.700 | -0.504 | 3 | composite_momentum_high | -0.216 | 3 | true |
| 6 | lgbm_ranker_h63_high | 0.909 | -0.702 | 2 | benchmark_tsmom_high | -0.615 | 3 | false |
| 7 | composite_momentum_high | 0.482 | 0.680 | 3 | lgbm_ranker_h63_high | 1.742 | 3 | true |
| 8 | lgbm_ranker_h63_high | 0.322 | 4.557 | 1 | lgbm_ranker_h63_high | 4.557 | 3 | false |
| 9 | benchmark_tsmom_high | 0.782 | 3.705 | 1 | benchmark_tsmom_high | 3.705 | 3 | false |
| 10 | benchmark_tsmom_high | 0.629 | 2.198 | 1 | benchmark_tsmom_high | 2.198 | 3 | false |
| 11 | benchmark_tsmom_high | 0.729 | 0.385 | 2 | lgbm_ranker_h63_high | 0.691 | 3 | false |
| 12 | lgbm_ranker_h63_high | 0.466 | 1.473 | 1 | lgbm_ranker_h63_high | 1.473 | 3 | false |
