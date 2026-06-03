# Alpha Validation Report

- Generated: `2026-06-03T14:08:44Z`
- Symbols: `VOO, NKE, DIS, CMCSA, CVS, BMY, MDT, PFE, UPS, USB, VZ, BA, GILD, PEP, MDLZ, PG, T, COP, ABT, AMT, XOM, KO, ACN, CVX, AMGN, MO, SO, CI, JNJ, SBUX, MRK, RTX, GE, C, IBM, HON, SCHW, CRM, ADP, SYK, BAC, PM, ELV, BKNG, LOW, BLK`
- Timeframe: `1Day`
- Period: `2018-01-02` to `2026-06-01`
- Bars: `2114`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 222.00% | 14.97% | 0.820 | 0.767 | 0.440 | 34.00% | 0 | 1.000 |
| equal_weight | 97.85% | 8.48% | 0.560 | 0.533 | 0.241 | 35.21% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_checkpoint | buy_hold | false | 277.29% | 17.16% | 0.916 | 0.863 | 0.490 | 35.04% | 1.000 | 0.222 | 104 | PBO 0.222 above 0.200 |

## benchmark_lgbm_ranker_h63_s15_checkpoint

- Family: `benchmark_lgbm_ranker_h63`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.222 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 277.29% | 17.16% | 0.916 | 0.863 | 0.490 | 35.04% | 104 | 16.914 | - |
| stress_2x | 276.60% | 17.13% | 0.914 | 0.862 | 0.489 | 35.04% | 104 | 16.896 | - |
| stress_3x | 275.91% | 17.11% | 0.913 | 0.861 | 0.488 | 35.04% | 104 | 16.877 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.607 | 2.378 | 8.697 | 3.95% | - |
| 1 | 126-882 | 882-1134 | 0.830 | -0.305 | -0.375 | 19.69% | - |
| 2 | 252-1008 | 1008-1260 | 1.114 | -0.402 | -0.538 | 21.86% | - |
| 3 | 378-1134 | 1134-1386 | 0.511 | 1.283 | 1.609 | 16.57% | - |
| 4 | 504-1260 | 1260-1512 | 0.505 | 1.937 | 2.873 | 10.27% | - |
| 5 | 630-1386 | 1386-1638 | 1.057 | 2.308 | 2.885 | 10.27% | - |
| 6 | 756-1512 | 1512-1764 | 0.880 | 2.164 | 3.908 | 7.47% | - |
| 7 | 882-1638 | 1638-1890 | 0.872 | 1.162 | 1.324 | 16.44% | - |
| 8 | 1008-1764 | 1764-2016 | 0.763 | 1.366 | 1.506 | 16.44% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.350 | 8.697 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 8.697 | 4 | false |
| 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.519 | -0.375 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | -0.375 | 4 | false |
| 2 | benchmark_lgbm_ranker_h63_s10_z125 | 0.733 | -0.619 | 4 | benchmark_lgbm_ranker_h63_s15_checkpoint | -0.538 | 4 | true |
| 3 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.280 | 1.609 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 1.609 | 4 | false |
| 4 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.297 | 2.873 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 2.873 | 4 | false |
| 5 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.876 | 2.885 | 4 | benchmark_lgbm_ranker_h63_s10_z125 | 2.909 | 4 | true |
| 6 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.685 | 3.908 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 3.908 | 4 | false |
| 7 | benchmark_lgbm_ranker_h63_s15_z125 | 0.679 | 1.324 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 1.324 | 4 | false |
| 8 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.521 | 1.506 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 1.506 | 4 | false |
