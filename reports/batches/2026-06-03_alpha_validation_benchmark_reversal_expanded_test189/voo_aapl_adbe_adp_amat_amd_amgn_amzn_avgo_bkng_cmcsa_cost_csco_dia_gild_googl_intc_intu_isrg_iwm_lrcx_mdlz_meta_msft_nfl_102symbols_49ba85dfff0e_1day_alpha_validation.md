# Alpha Validation Report

- Generated: `2026-06-03T08:12:30Z`
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
| benchmark_reversal_checkpoint | buy_hold | false | 144.38% | 16.61% | 0.999 | 0.979 | 0.684 | 24.27% | 1.000 | 0.667 | 2263 | PBO 0.667 above 0.200 |

## benchmark_reversal_checkpoint

- Family: `benchmark_reversal`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.667 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 144.38% | 16.61% | 0.999 | 0.979 | 0.684 | 24.27% | 2263 | 91.915 | - |
| stress_2x | 141.53% | 16.38% | 0.987 | 0.968 | 0.672 | 24.39% | 2265 | 91.315 | - |
| stress_3x | 138.71% | 16.14% | 0.975 | 0.955 | 0.659 | 24.50% | 2270 | 90.721 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-945 | 0.779 | 1.313 | 1.937 | 8.16% | - |
| 1 | 63-819 | 819-1008 | 0.482 | 3.036 | 7.079 | 5.74% | - |
| 2 | 126-882 | 882-1071 | 0.479 | 2.064 | 3.321 | 8.65% | - |
| 3 | 189-945 | 945-1134 | 0.446 | 1.842 | 2.799 | 9.32% | - |
| 4 | 252-1008 | 1008-1197 | 0.442 | 0.608 | 0.642 | 17.80% | - |
| 5 | 315-1071 | 1071-1260 | 0.603 | 0.736 | 0.796 | 17.85% | - |
| 6 | 378-1134 | 1134-1323 | 0.700 | 0.901 | 0.962 | 18.62% | - |
| 7 | 441-1197 | 1197-1386 | 0.669 | 2.453 | 5.505 | 5.57% | - |
| 8 | 504-1260 | 1260-1449 | 0.941 | 1.603 | 2.045 | 10.34% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_reversal_medium | 0.598 | 1.700 | 3 | benchmark_reversal_checkpoint | 1.937 | 4 | true |
| 1 | benchmark_reversal_slow | 0.336 | 7.336 | 3 | benchmark_reversal_medium | 7.698 | 4 | true |
| 2 | benchmark_reversal_slow | 0.397 | 4.191 | 1 | benchmark_reversal_slow | 4.191 | 4 | false |
| 3 | benchmark_reversal_slow | 0.284 | 4.189 | 1 | benchmark_reversal_slow | 4.189 | 4 | false |
| 4 | benchmark_reversal_slow | 0.355 | 0.398 | 3 | benchmark_reversal_checkpoint | 0.642 | 4 | true |
| 5 | benchmark_reversal_slow | 0.399 | 0.544 | 4 | benchmark_reversal_medium | 0.923 | 4 | true |
| 6 | benchmark_reversal_fast | 0.569 | 1.091 | 1 | benchmark_reversal_fast | 1.091 | 4 | false |
| 7 | benchmark_reversal_medium | 0.646 | 5.266 | 3 | benchmark_reversal_checkpoint | 5.505 | 4 | true |
| 8 | benchmark_reversal_medium | 0.922 | 2.324 | 3 | benchmark_reversal_slow | 2.912 | 4 | true |
