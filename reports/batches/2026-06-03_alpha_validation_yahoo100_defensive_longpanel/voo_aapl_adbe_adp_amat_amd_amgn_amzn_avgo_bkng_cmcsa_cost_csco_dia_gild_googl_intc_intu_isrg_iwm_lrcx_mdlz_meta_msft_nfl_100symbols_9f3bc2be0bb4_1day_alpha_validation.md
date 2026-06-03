# Alpha Validation Report

- Generated: `2026-06-03T08:46:14Z`
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
| benchmark_rotation_defensive | buy_hold | false | 283.63% | 12.54% | 0.789 | 0.735 | 0.368 | 34.10% | 1.000 | 0.400 | 587 | PBO 0.400 above 0.200 |

## benchmark_rotation_defensive

- Family: `composite_momentum_defensive`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.400 above 0.200
  - turnover increases without return improvement
  - no drawdown-adjusted improvement over benchmark

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 283.63% | 12.54% | 0.789 | 0.735 | 0.368 | 34.10% | 587 | 72.630 | - |
| stress_2x | 280.67% | 12.46% | 0.785 | 0.731 | 0.365 | 34.10% | 587 | 72.302 | - |
| stress_3x | 277.73% | 12.39% | 0.781 | 0.727 | 0.363 | 34.10% | 587 | 71.976 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.010 | -0.381 | -0.405 | 18.45% | - |
| 1 | 126-882 | 882-1134 | 0.926 | 0.710 | 0.529 | 18.45% | - |
| 2 | 252-1008 | 1008-1260 | 0.740 | 2.078 | 3.937 | 6.85% | - |
| 3 | 378-1134 | 1134-1386 | 1.189 | 0.189 | 0.025 | 34.10% | - |
| 4 | 504-1260 | 1260-1512 | 1.114 | 0.360 | 0.194 | 34.10% | - |
| 5 | 630-1386 | 1386-1638 | 0.459 | 2.244 | 4.061 | 9.58% | - |
| 6 | 756-1512 | 1512-1764 | 0.487 | 2.043 | 5.664 | 5.27% | - |
| 7 | 882-1638 | 1638-1890 | 0.749 | -0.352 | -0.349 | 22.63% | - |
| 8 | 1008-1764 | 1764-2016 | 1.028 | -0.851 | -0.718 | 24.20% | - |
| 9 | 1134-1890 | 1890-2142 | 0.527 | 0.618 | 0.573 | 14.73% | - |
| 10 | 1260-2016 | 2016-2268 | 0.422 | 1.536 | 1.956 | 9.97% | - |
| 11 | 1386-2142 | 2142-2394 | 0.750 | 2.218 | 2.896 | 9.97% | - |
| 12 | 1512-2268 | 2268-2520 | 0.592 | 1.805 | 2.965 | 8.70% | - |
| 13 | 1638-2394 | 2394-2646 | 0.674 | 0.360 | 0.276 | 18.00% | - |
| 14 | 1764-2520 | 2520-2772 | 0.482 | 0.772 | 0.668 | 18.00% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_rotation_cash_defensive | 1.019 | -0.427 | 3 | benchmark_rotation_defensive | -0.405 | 4 | true |
| 1 | benchmark_rotation_defensive | 0.999 | 0.529 | 2 | benchmark_rotation_half_defensive | 0.542 | 4 | false |
| 2 | benchmark_rotation_defensive | 0.492 | 3.937 | 3 | benchmark_rotation_half_defensive | 4.152 | 4 | true |
| 3 | benchmark_rotation_defensive | 0.793 | 0.025 | 3 | benchmark_rotation_half_defensive | 0.128 | 4 | true |
| 4 | benchmark_rotation_defensive | 0.766 | 0.194 | 3 | benchmark_rotation_half_defensive | 0.298 | 4 | true |
| 5 | benchmark_rotation_half_defensive | 0.270 | 4.044 | 2 | benchmark_rotation_defensive | 4.061 | 4 | false |
| 6 | benchmark_rotation_half_defensive | 0.297 | 5.721 | 1 | benchmark_rotation_half_defensive | 5.721 | 4 | false |
| 7 | benchmark_rotation_half_defensive | 0.487 | -0.392 | 4 | benchmark_rotation_defensive | -0.349 | 4 | true |
| 8 | benchmark_rotation_half_defensive | 0.688 | -0.696 | 1 | benchmark_rotation_half_defensive | -0.696 | 4 | false |
| 9 | benchmark_rotation_half_defensive | 0.306 | 0.708 | 1 | benchmark_rotation_half_defensive | 0.708 | 4 | false |
| 10 | benchmark_rotation_defensive | 0.223 | 1.956 | 2 | benchmark_rotation_half_defensive | 2.148 | 4 | false |
| 11 | benchmark_rotation_half_defensive | 0.487 | 2.880 | 4 | benchmark_rotation_cash_defensive | 2.925 | 4 | true |
| 12 | benchmark_rotation_half_defensive | 0.360 | 2.997 | 1 | benchmark_rotation_half_defensive | 2.997 | 4 | false |
| 13 | benchmark_rotation_half_defensive | 0.408 | 0.456 | 1 | benchmark_rotation_half_defensive | 0.456 | 4 | false |
| 14 | benchmark_rotation_half_defensive | 0.274 | 0.841 | 1 | benchmark_rotation_half_defensive | 0.841 | 4 | false |
