# Alpha Validation Report

- Generated: `2026-06-03T14:07:51Z`
- Symbols: `VOO, NKE, DIS, CMCSA, CVS, BMY, MDT, PFE, UPS, USB, VZ, BA, GILD, PEP, MDLZ, PG, T, COP, ABT, AMT, XOM, KO, ACN, CVX, AMGN, MO, SO, CI, JNJ, SBUX, MRK, RTX, GE, C, IBM, HON, SCHW, CRM, ADP, SYK, BAC, PM, ELV, BKNG, LOW, BLK`
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
| equal_weight | 167.76% | 9.04% | 0.613 | 0.586 | 0.250 | 36.13% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_checkpoint | buy_hold | false | 460.51% | 16.35% | 0.947 | 0.906 | 0.492 | 33.27% | 1.000 | 0.267 | 107 | PBO 0.267 above 0.200 |

## benchmark_lgbm_ranker_h63_s15_checkpoint

- Family: `benchmark_lgbm_ranker_h63`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.267 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 460.51% | 16.35% | 0.947 | 0.906 | 0.492 | 33.27% | 107 | 24.211 | - |
| stress_2x | 459.47% | 16.33% | 0.946 | 0.905 | 0.491 | 33.27% | 107 | 24.183 | - |
| stress_3x | 458.42% | 16.31% | 0.945 | 0.904 | 0.490 | 33.27% | 107 | 24.155 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.977 | -0.401 | -0.406 | 20.37% | - |
| 1 | 126-882 | 882-1134 | 0.906 | 0.817 | 0.619 | 20.37% | - |
| 2 | 252-1008 | 1008-1260 | 0.700 | 2.375 | 4.486 | 7.38% | - |
| 3 | 378-1134 | 1134-1386 | 1.202 | 0.481 | 0.332 | 33.27% | - |
| 4 | 504-1260 | 1260-1512 | 1.165 | 0.598 | 0.472 | 33.27% | - |
| 5 | 630-1386 | 1386-1638 | 0.628 | 2.279 | 4.516 | 8.80% | - |
| 6 | 756-1512 | 1512-1764 | 0.650 | 2.205 | 8.124 | 3.92% | - |
| 7 | 882-1638 | 1638-1890 | 0.888 | -0.314 | -0.386 | 19.60% | - |
| 8 | 1008-1764 | 1764-2016 | 1.164 | -0.356 | -0.495 | 21.83% | - |
| 9 | 1134-1890 | 1890-2142 | 0.584 | 1.200 | 1.475 | 16.58% | - |
| 10 | 1260-2016 | 2016-2268 | 0.503 | 2.028 | 3.048 | 10.17% | - |
| 11 | 1386-2142 | 2142-2394 | 1.053 | 2.276 | 2.871 | 10.17% | - |
| 12 | 1512-2268 | 2268-2520 | 0.851 | 2.014 | 3.604 | 7.46% | - |
| 13 | 1638-2394 | 2394-2646 | 0.877 | 1.162 | 1.329 | 16.47% | - |
| 14 | 1764-2520 | 2520-2772 | 0.767 | 1.449 | 1.624 | 16.47% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.919 | -0.406 | 3 | benchmark_lgbm_ranker_h63_s10 | -0.403 | 4 | true |
| 1 | benchmark_lgbm_ranker_h63_s10 | 0.896 | 0.605 | 3 | benchmark_lgbm_ranker_h63_s15_z125 | 0.619 | 4 | true |
| 2 | benchmark_lgbm_ranker_h63_s10 | 0.439 | 4.564 | 1 | benchmark_lgbm_ranker_h63_s10 | 4.564 | 4 | false |
| 3 | benchmark_lgbm_ranker_h63_s10 | 0.761 | 0.294 | 3 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.332 | 4 | true |
| 4 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.765 | 0.472 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 0.472 | 4 | false |
| 5 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.370 | 4.516 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 4.516 | 4 | false |
| 6 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.399 | 8.124 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 8.124 | 4 | false |
| 7 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.587 | -0.386 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | -0.386 | 4 | false |
| 8 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.787 | -0.495 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | -0.495 | 4 | false |
| 9 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.354 | 1.475 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 1.475 | 4 | false |
| 10 | benchmark_lgbm_ranker_h63_s15_z125 | 0.298 | 3.048 | 1 | benchmark_lgbm_ranker_h63_s15_z125 | 3.048 | 4 | false |
| 11 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.880 | 2.871 | 4 | benchmark_lgbm_ranker_h63_s10 | 2.883 | 4 | true |
| 12 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.660 | 3.604 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 3.604 | 4 | false |
| 13 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.680 | 1.329 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 1.329 | 4 | false |
| 14 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.535 | 1.624 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 1.624 | 4 | false |
