# Alpha Validation Report

- Generated: `2026-06-03T14:08:57Z`
- Symbols: `VOO, NKE, DIS, CMCSA, CVS, BMY, MDT, PFE, UPS, USB, VZ, BA, GILD, PEP, MDLZ, PG, T, COP, ABT, AMT, XOM, KO, ACN, CVX, AMGN, MO, SO, CI, JNJ, SBUX, MRK, RTX, GE, C, IBM, HON, SCHW, CRM, ADP, SYK, BAC, PM, ELV, BKNG, LOW, BLK`
- Timeframe: `1Day`
- Period: `2019-01-02` to `2026-06-01`
- Bars: `1863`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 242.40% | 18.13% | 0.950 | 0.890 | 0.533 | 34.00% | 0 | 1.000 |
| equal_weight | 119.38% | 11.22% | 0.692 | 0.660 | 0.314 | 35.69% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_checkpoint | buy_hold | false | 326.55% | 21.69% | 1.100 | 1.048 | 0.629 | 34.50% | 1.000 | 0.286 | 87 | PBO 0.286 above 0.200 |

## benchmark_lgbm_ranker_h63_s15_checkpoint

- Family: `benchmark_lgbm_ranker_h63`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.286 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 326.55% | 21.69% | 1.100 | 1.048 | 0.629 | 34.50% | 87 | 17.011 | - |
| stress_2x | 325.89% | 21.67% | 1.099 | 1.046 | 0.628 | 34.50% | 87 | 16.995 | - |
| stress_3x | 325.23% | 21.64% | 1.098 | 1.045 | 0.627 | 34.50% | 87 | 16.980 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.219 | -0.396 | -0.533 | 21.79% | - |
| 1 | 126-882 | 882-1134 | 0.598 | 1.314 | 1.653 | 16.55% | - |
| 2 | 252-1008 | 1008-1260 | 0.511 | 1.931 | 2.616 | 11.26% | - |
| 3 | 378-1134 | 1134-1386 | 1.077 | 2.175 | 2.458 | 11.26% | - |
| 4 | 504-1260 | 1260-1512 | 0.827 | 2.075 | 3.692 | 7.77% | - |
| 5 | 630-1386 | 1386-1638 | 0.777 | 1.212 | 1.464 | 16.43% | - |
| 6 | 756-1512 | 1512-1764 | 0.733 | 1.345 | 1.551 | 16.43% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.820 | -0.533 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | -0.533 | 4 | false |
| 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.351 | 1.653 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 1.653 | 4 | false |
| 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.303 | 2.616 | 3 | benchmark_lgbm_ranker_h63_s10_z125 | 2.623 | 4 | true |
| 3 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.900 | 2.458 | 3 | benchmark_lgbm_ranker_h63_s10_z125 | 2.656 | 4 | true |
| 4 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.636 | 3.692 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 3.692 | 4 | false |
| 5 | benchmark_lgbm_ranker_h63_s15_z125 | 0.553 | 1.464 | 1 | benchmark_lgbm_ranker_h63_s15_z125 | 1.464 | 4 | false |
| 6 | benchmark_lgbm_ranker_h63_s15_z125 | 0.507 | 1.551 | 1 | benchmark_lgbm_ranker_h63_s15_z125 | 1.551 | 4 | false |
