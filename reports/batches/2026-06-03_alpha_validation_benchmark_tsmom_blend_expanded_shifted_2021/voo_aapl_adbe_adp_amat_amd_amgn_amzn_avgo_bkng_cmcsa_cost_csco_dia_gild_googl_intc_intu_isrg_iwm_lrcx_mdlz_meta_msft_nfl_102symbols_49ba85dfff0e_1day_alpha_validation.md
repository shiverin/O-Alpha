# Alpha Validation Report

- Generated: `2026-06-03T07:55:42Z`
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
| benchmark_tsmom_blend | buy_hold | false | 78.60% | 11.40% | 0.714 | 0.687 | 0.481 | 23.69% | 1.000 | 0.000 | 142 | turnover increases without return improvement |

## benchmark_tsmom_blend

- Family: `benchmark_tsmom_blend`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - turnover increases without return improvement

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 78.60% | 11.40% | 0.714 | 0.687 | 0.481 | 23.69% | 142 | 9.747 | - |
| stress_2x | 78.34% | 11.37% | 0.713 | 0.686 | 0.479 | 23.72% | 142 | 9.741 | - |
| stress_3x | 78.08% | 11.34% | 0.711 | 0.685 | 0.478 | 23.74% | 142 | 9.734 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-945 | 0.545 | 1.282 | 1.635 | 13.41% | - |
| 1 | 63-819 | 819-1008 | 0.556 | 0.464 | 0.485 | 13.41% | - |
| 2 | 126-882 | 882-1071 | 0.594 | -0.531 | -0.702 | 19.67% | - |
| 3 | 189-945 | 945-1134 | 0.498 | 0.399 | 0.324 | 19.67% | - |
| 4 | 252-1008 | 1008-1197 | 0.339 | 0.780 | 0.754 | 19.67% | - |
| 5 | 315-1071 | 1071-1260 | 0.257 | 1.523 | 2.348 | 9.58% | - |
| 6 | 378-1134 | 1134-1323 | 0.787 | 0.601 | 0.596 | 14.51% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_blend_reb42 | 0.392 | 2.171 | 1 | benchmark_tsmom_blend_reb42 | 2.171 | 4 | false |
| 1 | benchmark_tsmom_blend_reb42 | 0.364 | 0.928 | 1 | benchmark_tsmom_blend_reb42 | 0.928 | 4 | false |
| 2 | benchmark_tsmom_blend_reb42 | 0.422 | -0.617 | 2 | benchmark_tsmom_blend_etf_tilt | -0.610 | 4 | false |
| 3 | benchmark_tsmom_blend_reb42 | 0.382 | 0.470 | 1 | benchmark_tsmom_blend_reb42 | 0.470 | 4 | false |
| 4 | benchmark_tsmom_blend_reb42 | 0.237 | 1.004 | 1 | benchmark_tsmom_blend_reb42 | 1.004 | 4 | false |
| 5 | benchmark_tsmom_blend_reb42 | 0.228 | 3.835 | 1 | benchmark_tsmom_blend_reb42 | 3.835 | 4 | false |
| 6 | benchmark_tsmom_blend_reb42 | 0.839 | 1.763 | 1 | benchmark_tsmom_blend_reb42 | 1.763 | 4 | false |
