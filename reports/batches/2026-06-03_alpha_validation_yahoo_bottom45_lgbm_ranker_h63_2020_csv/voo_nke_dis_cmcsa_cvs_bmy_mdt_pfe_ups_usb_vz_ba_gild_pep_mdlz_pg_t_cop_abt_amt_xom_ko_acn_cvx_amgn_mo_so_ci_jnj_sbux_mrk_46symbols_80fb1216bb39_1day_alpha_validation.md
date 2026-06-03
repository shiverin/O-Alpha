# Alpha Validation Report

- Generated: `2026-06-03T14:09:07Z`
- Symbols: `VOO, NKE, DIS, CMCSA, CVS, BMY, MDT, PFE, UPS, USB, VZ, BA, GILD, PEP, MDLZ, PG, T, COP, ABT, AMT, XOM, KO, ACN, CVX, AMGN, MO, SO, CI, JNJ, SBUX, MRK, RTX, GE, C, IBM, HON, SCHW, CRM, ADP, SYK, BAC, PM, ELV, BKNG, LOW, BLK`
- Timeframe: `1Day`
- Period: `2020-01-02` to `2026-06-01`
- Bars: `1611`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 159.21% | 16.08% | 0.831 | 0.783 | 0.473 | 34.00% | 0 | 1.000 |
| equal_weight | 71.24% | 8.78% | 0.553 | 0.532 | 0.248 | 35.49% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_checkpoint | buy_hold | false | 207.02% | 19.19% | 0.956 | 0.913 | 0.565 | 34.00% | 1.000 | 0.400 | 75 | PBO 0.400 above 0.200 |

## benchmark_lgbm_ranker_h63_s15_checkpoint

- Family: `benchmark_lgbm_ranker_h63`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.400 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 207.02% | 19.19% | 0.956 | 0.913 | 0.565 | 34.00% | 75 | 11.477 | - |
| stress_2x | 206.59% | 19.17% | 0.954 | 0.912 | 0.564 | 34.01% | 75 | 11.468 | - |
| stress_3x | 206.17% | 19.14% | 0.953 | 0.911 | 0.563 | 34.01% | 75 | 11.459 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.511 | 1.931 | 2.616 | 11.26% | - |
| 1 | 126-882 | 882-1134 | 1.077 | 2.175 | 2.458 | 11.26% | - |
| 2 | 252-1008 | 1008-1260 | 0.827 | 2.075 | 3.692 | 7.77% | - |
| 3 | 378-1134 | 1134-1386 | 0.777 | 1.212 | 1.464 | 16.43% | - |
| 4 | 504-1260 | 1260-1512 | 0.733 | 1.345 | 1.551 | 16.43% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.303 | 2.616 | 4 | benchmark_lgbm_ranker_h63_s10_z125 | 2.623 | 4 | true |
| 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.900 | 2.458 | 3 | benchmark_lgbm_ranker_h63_s10 | 2.656 | 4 | true |
| 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.636 | 3.692 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 3.692 | 4 | false |
| 3 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.553 | 1.464 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 1.464 | 4 | false |
| 4 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.507 | 1.551 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 1.551 | 4 | false |
