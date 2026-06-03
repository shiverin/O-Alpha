# Alpha Validation Report

- Generated: `2026-06-03T07:56:06Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLC, XLE, XLF, XLI, XLK, XLP, XLRE, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2020-07-27` to `2026-06-01`
- Bars: `1466`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 135.70% | 15.89% | 0.972 | 0.950 | 0.628 | 25.32% | 0 | 1.000 |
| equal_weight | 128.33% | 15.26% | 0.927 | 0.925 | 0.581 | 26.25% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_tsmom_blend | buy_hold | false | 153.09% | 17.32% | 1.030 | 1.008 | 0.857 | 20.21% | 1.000 | 0.444 | 164 | PBO 0.444 above 0.200 |

## benchmark_tsmom_blend

- Family: `benchmark_tsmom_blend`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.444 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 153.09% | 17.32% | 1.030 | 1.008 | 0.857 | 20.21% | 164 | 12.622 | - |
| stress_2x | 152.71% | 17.29% | 1.028 | 1.007 | 0.855 | 20.23% | 164 | 12.613 | - |
| stress_3x | 152.33% | 17.26% | 1.026 | 1.005 | 0.853 | 20.24% | 164 | 12.604 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-945 | 0.920 | 1.381 | 2.250 | 8.87% | - |
| 1 | 63-819 | 819-1008 | 0.542 | 2.920 | 6.889 | 6.58% | - |
| 2 | 126-882 | 882-1071 | 0.681 | 1.662 | 2.472 | 9.89% | - |
| 3 | 189-945 | 945-1134 | 0.536 | 1.624 | 2.360 | 9.89% | - |
| 4 | 252-1008 | 1008-1197 | 0.539 | 0.435 | 0.376 | 19.53% | - |
| 5 | 315-1071 | 1071-1260 | 0.532 | 0.644 | 0.616 | 19.53% | - |
| 6 | 378-1134 | 1134-1323 | 0.673 | 0.949 | 0.986 | 19.53% | - |
| 7 | 441-1197 | 1197-1386 | 0.635 | 1.735 | 3.285 | 7.39% | - |
| 8 | 504-1260 | 1260-1449 | 0.882 | 1.202 | 1.412 | 12.71% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_blend | 0.780 | 2.250 | 1 | benchmark_tsmom_blend | 2.250 | 4 | false |
| 1 | benchmark_tsmom_blend | 0.395 | 6.889 | 4 | benchmark_tsmom_blend_reb42 | 7.266 | 4 | true |
| 2 | benchmark_tsmom_blend | 0.512 | 2.472 | 3 | benchmark_tsmom_blend_etf_tilt | 2.746 | 4 | true |
| 3 | benchmark_tsmom_blend | 0.365 | 2.360 | 3 | benchmark_tsmom_blend_etf_tilt | 2.551 | 4 | true |
| 4 | benchmark_tsmom_blend | 0.341 | 0.376 | 2 | benchmark_tsmom_blend_slow_broad | 0.425 | 4 | false |
| 5 | benchmark_tsmom_blend_slow_broad | 0.337 | 0.661 | 2 | benchmark_tsmom_blend_reb42 | 0.668 | 4 | false |
| 6 | benchmark_tsmom_blend_slow_broad | 0.497 | 0.941 | 3 | benchmark_tsmom_blend_etf_tilt | 1.043 | 4 | true |
| 7 | benchmark_tsmom_blend_slow_broad | 0.574 | 4.334 | 2 | benchmark_tsmom_blend_reb42 | 4.518 | 4 | false |
| 8 | benchmark_tsmom_blend_slow_broad | 0.782 | 1.833 | 1 | benchmark_tsmom_blend_slow_broad | 1.833 | 4 | false |
