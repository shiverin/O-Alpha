# Alpha Validation Report

- Generated: `2026-06-03T14:08:11Z`
- Symbols: `VOO, NKE, DIS, CMCSA, CVS, BMY, MDT, PFE, UPS, USB, VZ, BA, GILD, PEP, MDLZ, PG, T, COP, ABT, AMT, XOM, KO, ACN, CVX, AMGN, MO, SO, CI, JNJ, SBUX, MRK, RTX, GE, C, IBM, HON, SCHW, CRM, ADP, SYK, BAC, PM, ELV, BKNG, LOW, BLK`
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
| equal_weight | 158.22% | 9.57% | 0.640 | 0.612 | 0.264 | 36.29% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_checkpoint | buy_hold | true | 456.86% | 17.99% | 1.018 | 0.968 | 0.541 | 33.27% | 1.000 | 0.154 | 107 | pass |

## benchmark_lgbm_ranker_h63_s15_checkpoint

- Family: `benchmark_lgbm_ranker_h63`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `true`

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 456.86% | 17.99% | 1.018 | 0.968 | 0.541 | 33.27% | 107 | 24.060 | - |
| stress_2x | 455.82% | 17.97% | 1.017 | 0.967 | 0.540 | 33.27% | 107 | 24.032 | - |
| stress_3x | 454.79% | 17.95% | 1.016 | 0.966 | 0.539 | 33.27% | 107 | 24.004 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.700 | 2.375 | 4.486 | 7.38% | - |
| 1 | 126-882 | 882-1134 | 1.202 | 0.481 | 0.332 | 33.27% | - |
| 2 | 252-1008 | 1008-1260 | 1.165 | 0.598 | 0.472 | 33.27% | - |
| 3 | 378-1134 | 1134-1386 | 0.628 | 2.279 | 4.516 | 8.80% | - |
| 4 | 504-1260 | 1260-1512 | 0.650 | 2.205 | 8.124 | 3.92% | - |
| 5 | 630-1386 | 1386-1638 | 0.888 | -0.314 | -0.386 | 19.60% | - |
| 6 | 756-1512 | 1512-1764 | 1.164 | -0.356 | -0.495 | 21.83% | - |
| 7 | 882-1638 | 1638-1890 | 0.584 | 1.200 | 1.475 | 16.58% | - |
| 8 | 1008-1764 | 1764-2016 | 0.503 | 2.028 | 3.048 | 10.17% | - |
| 9 | 1134-1890 | 1890-2142 | 1.053 | 2.276 | 2.871 | 10.17% | - |
| 10 | 1260-2016 | 2016-2268 | 0.851 | 2.014 | 3.604 | 7.46% | - |
| 11 | 1386-2142 | 2142-2394 | 0.877 | 1.162 | 1.329 | 16.47% | - |
| 12 | 1512-2268 | 2268-2520 | 0.767 | 1.449 | 1.624 | 16.47% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s10_z125 | 0.439 | 4.564 | 2 | benchmark_lgbm_ranker_h63_s10 | 4.564 | 4 | false |
| 1 | benchmark_lgbm_ranker_h63_s10 | 0.761 | 0.294 | 4 | benchmark_lgbm_ranker_h63_s15_z125 | 0.332 | 4 | true |
| 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.765 | 0.472 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.472 | 4 | false |
| 3 | benchmark_lgbm_ranker_h63_s15_z125 | 0.370 | 4.516 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 4.516 | 4 | false |
| 4 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.399 | 8.124 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 8.124 | 4 | false |
| 5 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.587 | -0.386 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | -0.386 | 4 | false |
| 6 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.787 | -0.495 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | -0.495 | 4 | false |
| 7 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.354 | 1.475 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 1.475 | 4 | false |
| 8 | benchmark_lgbm_ranker_h63_s15_z125 | 0.298 | 3.048 | 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | 3.048 | 4 | false |
| 9 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.880 | 2.871 | 4 | benchmark_lgbm_ranker_h63_s10 | 2.883 | 4 | true |
| 10 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.660 | 3.604 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 3.604 | 4 | false |
| 11 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.680 | 1.329 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 1.329 | 4 | false |
| 12 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.535 | 1.624 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 1.624 | 4 | false |
