# Alpha Validation Report

- Generated: `2026-06-03T14:08:29Z`
- Symbols: `VOO, NKE, DIS, CMCSA, CVS, BMY, MDT, PFE, UPS, USB, VZ, BA, GILD, PEP, MDLZ, PG, T, COP, ABT, AMT, XOM, KO, ACN, CVX, AMGN, MO, SO, CI, JNJ, SBUX, MRK, RTX, GE, C, IBM, HON, SCHW, CRM, ADP, SYK, BAC, PM, ELV, BKNG, LOW, BLK`
- Timeframe: `1Day`
- Period: `2017-01-03` to `2026-06-01`
- Bars: `2365`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 291.62% | 15.66% | 0.885 | 0.827 | 0.461 | 34.00% | 0 | 1.000 |
| equal_weight | 129.79% | 9.27% | 0.617 | 0.586 | 0.259 | 35.78% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_checkpoint | buy_hold | true | 385.72% | 18.35% | 1.014 | 0.962 | 0.552 | 33.27% | 1.000 | 0.091 | 107 | pass |

## benchmark_lgbm_ranker_h63_s15_checkpoint

- Family: `benchmark_lgbm_ranker_h63`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `true`

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 385.72% | 18.35% | 1.014 | 0.962 | 0.552 | 33.27% | 107 | 21.114 | - |
| stress_2x | 384.82% | 18.33% | 1.013 | 0.961 | 0.551 | 33.27% | 107 | 21.090 | - |
| stress_3x | 383.92% | 18.30% | 1.012 | 0.960 | 0.550 | 33.27% | 107 | 21.066 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.165 | 0.598 | 0.472 | 33.27% | - |
| 1 | 126-882 | 882-1134 | 0.628 | 2.279 | 4.516 | 8.80% | - |
| 2 | 252-1008 | 1008-1260 | 0.650 | 2.205 | 8.124 | 3.92% | - |
| 3 | 378-1134 | 1134-1386 | 0.888 | -0.314 | -0.386 | 19.60% | - |
| 4 | 504-1260 | 1260-1512 | 1.164 | -0.356 | -0.495 | 21.83% | - |
| 5 | 630-1386 | 1386-1638 | 0.584 | 1.200 | 1.475 | 16.58% | - |
| 6 | 756-1512 | 1512-1764 | 0.503 | 2.028 | 3.048 | 10.17% | - |
| 7 | 882-1638 | 1638-1890 | 1.053 | 2.276 | 2.871 | 10.17% | - |
| 8 | 1008-1764 | 1764-2016 | 0.851 | 2.014 | 3.604 | 7.46% | - |
| 9 | 1134-1890 | 1890-2142 | 0.877 | 1.162 | 1.329 | 16.47% | - |
| 10 | 1260-2016 | 2016-2268 | 0.767 | 1.449 | 1.624 | 16.47% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.765 | 0.472 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.472 | 4 | false |
| 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.370 | 4.516 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 4.516 | 4 | false |
| 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.399 | 8.124 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 8.124 | 4 | false |
| 3 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.587 | -0.386 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | -0.386 | 4 | false |
| 4 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.787 | -0.495 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | -0.495 | 4 | false |
| 5 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.354 | 1.475 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 1.475 | 4 | false |
| 6 | benchmark_lgbm_ranker_h63_s15_z125 | 0.298 | 3.048 | 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | 3.048 | 4 | false |
| 7 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.880 | 2.871 | 4 | benchmark_lgbm_ranker_h63_s10_z125 | 2.883 | 4 | true |
| 8 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.660 | 3.604 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 3.604 | 4 | false |
| 9 | benchmark_lgbm_ranker_h63_s15_z125 | 0.680 | 1.329 | 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | 1.329 | 4 | false |
| 10 | benchmark_lgbm_ranker_h63_s15_z125 | 0.535 | 1.624 | 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | 1.624 | 4 | false |
