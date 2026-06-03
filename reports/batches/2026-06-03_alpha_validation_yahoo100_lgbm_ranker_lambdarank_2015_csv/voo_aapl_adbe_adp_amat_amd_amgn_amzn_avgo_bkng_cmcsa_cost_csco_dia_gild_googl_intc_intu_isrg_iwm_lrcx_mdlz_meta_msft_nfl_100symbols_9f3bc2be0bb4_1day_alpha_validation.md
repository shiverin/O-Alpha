# Alpha Validation Report

- Generated: `2026-06-03T13:17:12Z`
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
| benchmark_lgbm_ranker_h63_s15 | buy_hold | false | 631.93% | 19.11% | 1.013 | 0.966 | 0.561 | 34.07% | 1.000 | 0.333 | 104 | PBO 0.333 above 0.200 |

## benchmark_lgbm_ranker_h63_s15

- Family: `benchmark_lgbm_ranker`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.333 above 0.200

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
| 0 | benchmark_lgbm_ranker_h63_s15 | 0.919 | -0.299 | 3 | benchmark_lgbm_ranker_h126_s10 | -0.292 | 4 | true |
| 1 | benchmark_lgbm_ranker_h126_s10 | 0.923 | 0.619 | 1 | benchmark_lgbm_ranker_h126_s10 | 0.619 | 4 | false |
| 2 | benchmark_lgbm_ranker_h126_s10 | 0.445 | 4.453 | 2 | benchmark_lgbm_ranker_h126_s15 | 4.681 | 4 | false |
| 3 | benchmark_lgbm_ranker_h126_s10 | 0.748 | 0.507 | 2 | benchmark_lgbm_ranker_h126_s15 | 0.647 | 4 | false |
| 4 | benchmark_lgbm_ranker_h126_s10 | 0.760 | 0.764 | 2 | benchmark_lgbm_ranker_h126_s15 | 0.960 | 4 | false |
| 5 | benchmark_lgbm_ranker_h126_s15 | 0.498 | 4.132 | 3 | benchmark_lgbm_ranker_h63_s15 | 4.396 | 4 | true |
| 6 | benchmark_lgbm_ranker_h126_s15 | 0.597 | 5.882 | 4 | benchmark_lgbm_ranker_h63_s15 | 8.568 | 4 | true |
| 7 | benchmark_lgbm_ranker_h126_s15 | 0.685 | -0.329 | 1 | benchmark_lgbm_ranker_h126_s15 | -0.329 | 4 | false |
| 8 | benchmark_lgbm_ranker_h126_s15 | 0.995 | -0.778 | 4 | benchmark_lgbm_ranker_h63_s15 | -0.636 | 4 | true |
| 9 | benchmark_lgbm_ranker_h63_s15 | 0.484 | 1.754 | 1 | benchmark_lgbm_ranker_h63_s15 | 1.754 | 4 | false |
| 10 | benchmark_lgbm_ranker_h63_s15 | 0.335 | 3.854 | 1 | benchmark_lgbm_ranker_h63_s15 | 3.854 | 4 | false |
| 11 | benchmark_lgbm_ranker_h63_s15 | 0.871 | 3.560 | 1 | benchmark_lgbm_ranker_h63_s15 | 3.560 | 4 | false |
| 12 | benchmark_lgbm_ranker_h63_s15 | 0.609 | 3.254 | 3 | benchmark_lgbm_ranker_h126_s15 | 4.055 | 4 | true |
| 13 | benchmark_lgbm_ranker_h63_s15 | 0.658 | 1.063 | 1 | benchmark_lgbm_ranker_h63_s15 | 1.063 | 4 | false |
| 14 | benchmark_lgbm_ranker_h63_s15 | 0.640 | 1.823 | 1 | benchmark_lgbm_ranker_h63_s15 | 1.823 | 4 | false |
