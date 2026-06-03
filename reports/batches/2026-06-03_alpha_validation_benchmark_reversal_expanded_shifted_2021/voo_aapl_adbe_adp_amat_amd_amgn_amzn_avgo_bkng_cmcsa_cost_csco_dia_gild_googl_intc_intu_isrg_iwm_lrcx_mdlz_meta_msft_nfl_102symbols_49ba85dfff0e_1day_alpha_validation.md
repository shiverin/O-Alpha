# Alpha Validation Report

- Generated: `2026-06-03T08:12:31Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLC, XLE, XLF, XLI, XLK, XLP, XLRE, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2021-01-04` to `2026-06-01`
- Bars: `1355`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 106.05% | 14.40% | 0.895 | 0.880 | 0.569 | 25.32% | 0 | 1.000 |
| equal_weight | 89.75% | 12.66% | 0.829 | 0.819 | 0.499 | 25.37% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_reversal_checkpoint | buy_hold | false | 104.15% | 14.21% | 0.873 | 0.860 | 0.518 | 27.42% | 1.000 | 0.143 | 2103 | turnover increases without return improvement |

## benchmark_reversal_checkpoint

- Family: `benchmark_reversal`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - turnover increases without return improvement
  - no drawdown-adjusted improvement over benchmark

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 104.15% | 14.21% | 0.873 | 0.860 | 0.518 | 27.42% | 2103 | 73.893 | - |
| stress_2x | 101.93% | 13.97% | 0.861 | 0.848 | 0.508 | 27.53% | 2104 | 73.452 | - |
| stress_3x | 99.74% | 13.74% | 0.849 | 0.836 | 0.497 | 27.65% | 2105 | 73.014 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-945 | 0.519 | 2.226 | 3.689 | 8.65% | - |
| 1 | 63-819 | 819-1008 | 0.565 | 1.170 | 1.687 | 9.32% | - |
| 2 | 126-882 | 882-1071 | 0.543 | -0.033 | -0.167 | 17.80% | - |
| 3 | 189-945 | 945-1134 | 0.705 | 0.634 | 0.660 | 17.85% | - |
| 4 | 252-1008 | 1008-1197 | 0.500 | 1.010 | 1.114 | 18.62% | - |
| 5 | 315-1071 | 1071-1260 | 0.416 | 2.940 | 7.427 | 5.57% | - |
| 6 | 378-1134 | 1134-1323 | 1.032 | 1.176 | 1.440 | 10.34% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_reversal_slow | 0.533 | 3.899 | 1 | benchmark_reversal_slow | 3.899 | 4 | false |
| 1 | benchmark_reversal_medium | 0.622 | 1.710 | 2 | benchmark_reversal_slow | 1.982 | 4 | false |
| 2 | benchmark_reversal_fast | 0.365 | -0.400 | 4 | benchmark_reversal_checkpoint | -0.167 | 4 | true |
| 3 | benchmark_reversal_checkpoint | 0.474 | 0.660 | 1 | benchmark_reversal_checkpoint | 0.660 | 4 | false |
| 4 | benchmark_reversal_medium | 0.367 | 1.273 | 1 | benchmark_reversal_medium | 1.273 | 4 | false |
| 5 | benchmark_reversal_checkpoint | 0.276 | 7.427 | 1 | benchmark_reversal_checkpoint | 7.427 | 4 | false |
| 6 | benchmark_reversal_fast | 0.952 | 1.959 | 1 | benchmark_reversal_fast | 1.959 | 4 | false |
