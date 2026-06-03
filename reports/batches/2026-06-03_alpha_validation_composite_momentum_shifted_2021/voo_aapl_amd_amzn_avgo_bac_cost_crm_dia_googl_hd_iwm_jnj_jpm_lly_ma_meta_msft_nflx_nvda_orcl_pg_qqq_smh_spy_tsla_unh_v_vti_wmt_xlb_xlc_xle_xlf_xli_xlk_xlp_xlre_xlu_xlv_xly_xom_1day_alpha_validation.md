# Alpha Validation Report

- Generated: `2026-06-03T07:10:58Z`
- Symbols: `VOO, AAPL, AMD, AMZN, AVGO, BAC, COST, CRM, DIA, GOOGL, HD, IWM, JNJ, JPM, LLY, MA, META, MSFT, NFLX, NVDA, ORCL, PG, QQQ, SMH, SPY, TSLA, UNH, V, VTI, WMT, XLB, XLC, XLE, XLF, XLI, XLK, XLP, XLRE, XLU, XLV, XLY, XOM`
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
| buy_hold | 106.03% | 14.40% | 0.895 | 0.880 | 0.569 | 25.32% | 3 | 1.000 |
| equal_weight | 100.71% | 13.84% | 0.825 | 0.791 | 0.447 | 30.98% | 28067 | 19.557 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| composite_momentum_checkpoint | buy_hold | false | 89.71% | 12.66% | 0.803 | 0.766 | 0.439 | 28.85% | 1.000 | 0.286 | 4087 | PBO 0.286 above 0.200 |

## composite_momentum_checkpoint

- Family: `composite_momentum`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.286 above 0.200
  - turnover increases without return improvement
  - no drawdown-adjusted improvement over benchmark

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 89.71% | 12.66% | 0.803 | 0.766 | 0.439 | 28.85% | 4087 | 37.006 | - |
| stress_2x | 88.61% | 12.54% | 0.797 | 0.760 | 0.433 | 28.93% | 4089 | 36.892 | - |
| stress_3x | 87.53% | 12.41% | 0.790 | 0.754 | 0.428 | 29.00% | 4092 | 36.780 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-945 | 0.474 | 2.017 | 3.422 | 8.67% | - |
| 1 | 63-819 | 819-1008 | 0.530 | 1.198 | 1.961 | 8.34% | - |
| 2 | 126-882 | 882-1071 | 0.530 | -0.364 | -0.480 | 20.10% | - |
| 3 | 189-945 | 945-1134 | 0.604 | 0.267 | 0.177 | 19.67% | - |
| 4 | 252-1008 | 1008-1197 | 0.439 | 0.659 | 0.615 | 19.67% | - |
| 5 | 315-1071 | 1071-1260 | 0.211 | 2.557 | 6.669 | 5.53% | - |
| 6 | 378-1134 | 1134-1323 | 0.896 | 1.573 | 3.549 | 5.13% | - |
