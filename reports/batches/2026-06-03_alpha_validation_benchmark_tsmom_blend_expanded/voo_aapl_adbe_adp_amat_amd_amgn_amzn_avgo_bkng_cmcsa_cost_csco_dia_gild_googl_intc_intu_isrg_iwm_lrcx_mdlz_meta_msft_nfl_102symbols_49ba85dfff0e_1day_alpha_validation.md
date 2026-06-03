# Alpha Validation Report

- Generated: `2026-06-03T07:54:53Z`
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
| benchmark_tsmom_blend | buy_hold | false | 153.09% | 17.32% | 1.030 | 1.008 | 0.857 | 20.21% | 1.000 | 0.250 | 164 | PBO 0.250 above 0.200 |

## benchmark_tsmom_blend

- Family: `benchmark_tsmom_blend`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.250 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 153.09% | 17.32% | 1.030 | 1.008 | 0.857 | 20.21% | 164 | 12.622 | - |
| stress_2x | 152.71% | 17.29% | 1.028 | 1.007 | 0.855 | 20.23% | 164 | 12.613 | - |
| stress_3x | 152.33% | 17.26% | 1.026 | 1.005 | 0.853 | 20.24% | 164 | 12.604 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.920 | 1.634 | 2.631 | 8.87% | - |
| 1 | 126-882 | 882-1134 | 0.681 | 1.471 | 2.162 | 9.89% | - |
| 2 | 252-1008 | 1008-1260 | 0.539 | 0.876 | 0.842 | 19.53% | - |
| 3 | 378-1134 | 1134-1386 | 0.673 | 0.543 | 0.466 | 19.53% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_blend | 0.780 | 2.631 | 1 | benchmark_tsmom_blend | 2.631 | 4 | false |
| 1 | benchmark_tsmom_blend | 0.512 | 2.162 | 3 | benchmark_tsmom_blend_etf_tilt | 2.322 | 4 | true |
| 2 | benchmark_tsmom_blend | 0.341 | 0.842 | 2 | benchmark_tsmom_blend_slow_broad | 0.880 | 4 | false |
| 3 | benchmark_tsmom_blend_slow_broad | 0.497 | 0.543 | 1 | benchmark_tsmom_blend_slow_broad | 0.543 | 4 | false |
