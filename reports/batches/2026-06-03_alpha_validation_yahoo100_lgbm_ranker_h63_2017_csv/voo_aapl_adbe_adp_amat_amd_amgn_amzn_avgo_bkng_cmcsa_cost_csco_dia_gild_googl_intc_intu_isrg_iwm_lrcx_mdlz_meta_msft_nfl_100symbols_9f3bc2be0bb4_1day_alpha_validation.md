# Alpha Validation Report

- Generated: `2026-06-03T13:21:21Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
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
| equal_weight | 511.40% | 21.29% | 1.042 | 0.980 | 0.644 | 33.06% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_checkpoint | buy_hold | true | 534.28% | 21.76% | 1.089 | 1.028 | 0.639 | 34.07% | 1.000 | 0.091 | 104 | pass |

## benchmark_lgbm_ranker_h63_s15_checkpoint

- Family: `benchmark_lgbm_ranker_h63`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `true`

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 534.28% | 21.76% | 1.089 | 1.028 | 0.639 | 34.07% | 104 | 22.970 | - |
| stress_2x | 533.09% | 21.74% | 1.088 | 1.027 | 0.638 | 34.08% | 105 | 22.942 | - |
| stress_3x | 531.90% | 21.72% | 1.086 | 1.026 | 0.637 | 34.08% | 105 | 22.913 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.128 | 0.748 | 0.671 | 34.07% | - |
| 1 | 126-882 | 882-1134 | 0.683 | 2.425 | 4.396 | 10.47% | - |
| 2 | 252-1008 | 1008-1260 | 0.677 | 2.322 | 8.568 | 4.38% | - |
| 3 | 378-1134 | 1134-1386 | 0.923 | -0.284 | -0.360 | 22.54% | - |
| 4 | 504-1260 | 1260-1512 | 1.250 | -0.521 | -0.636 | 24.19% | - |
| 5 | 630-1386 | 1386-1638 | 0.725 | 1.325 | 1.754 | 16.58% | - |
| 6 | 756-1512 | 1512-1764 | 0.538 | 2.198 | 3.854 | 9.66% | - |
| 7 | 882-1638 | 1638-1890 | 1.086 | 2.435 | 3.560 | 9.66% | - |
| 8 | 1008-1764 | 1764-2016 | 0.834 | 1.950 | 3.254 | 9.81% | - |
| 9 | 1134-1890 | 1890-2142 | 0.902 | 0.952 | 1.063 | 18.84% | - |
| 10 | 1260-2016 | 2016-2268 | 0.821 | 1.598 | 1.823 | 18.84% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s10 | 0.700 | 0.581 | 3 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.671 | 4 | true |
| 1 | benchmark_lgbm_ranker_h63_s15_z125 | 0.423 | 4.396 | 1 | benchmark_lgbm_ranker_h63_s15_z125 | 4.396 | 4 | false |
| 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.430 | 8.568 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 8.568 | 4 | false |
| 3 | benchmark_lgbm_ranker_h63_s15_z125 | 0.637 | -0.360 | 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | -0.360 | 4 | false |
| 4 | benchmark_lgbm_ranker_h63_s15_z125 | 0.905 | -0.636 | 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | -0.636 | 4 | false |
| 5 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.484 | 1.754 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 1.754 | 4 | false |
| 6 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.335 | 3.854 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 3.854 | 4 | false |
| 7 | benchmark_lgbm_ranker_h63_s15_z125 | 0.871 | 3.560 | 1 | benchmark_lgbm_ranker_h63_s15_z125 | 3.560 | 4 | false |
| 8 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.609 | 3.254 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 3.254 | 4 | false |
| 9 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.658 | 1.063 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 1.063 | 4 | false |
| 10 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.640 | 1.823 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 1.823 | 4 | false |
