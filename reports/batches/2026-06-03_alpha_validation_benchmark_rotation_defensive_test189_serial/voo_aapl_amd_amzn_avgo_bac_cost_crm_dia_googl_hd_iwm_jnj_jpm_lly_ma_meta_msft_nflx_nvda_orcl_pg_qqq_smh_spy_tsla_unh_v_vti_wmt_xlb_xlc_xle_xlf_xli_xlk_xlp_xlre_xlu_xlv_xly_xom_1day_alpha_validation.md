# Alpha Validation Report

- Generated: `2026-06-03T07:36:03Z`
- Symbols: `VOO, AAPL, AMD, AMZN, AVGO, BAC, COST, CRM, DIA, GOOGL, HD, IWM, JNJ, JPM, LLY, MA, META, MSFT, NFLX, NVDA, ORCL, PG, QQQ, SMH, SPY, TSLA, UNH, V, VTI, WMT, XLB, XLC, XLE, XLF, XLI, XLK, XLP, XLRE, XLU, XLV, XLY, XOM`
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
| equal_weight | 96.74% | 12.35% | 0.759 | 0.711 | 0.427 | 28.88% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_rotation_defensive | buy_hold | false | 108.81% | 13.50% | 0.921 | 0.871 | 0.545 | 24.78% | 1.000 | 0.667 | 253 | PBO 0.667 above 0.200 |

## benchmark_rotation_defensive

- Family: `composite_momentum_defensive`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.667 above 0.200
  - turnover increases without return improvement
  - no drawdown-adjusted improvement over benchmark

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 108.81% | 13.50% | 0.921 | 0.871 | 0.545 | 24.78% | 253 | 33.750 | - |
| stress_2x | 107.79% | 13.41% | 0.915 | 0.865 | 0.539 | 24.88% | 254 | 33.665 | - |
| stress_3x | 106.77% | 13.31% | 0.909 | 0.860 | 0.533 | 24.98% | 255 | 33.580 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-945 | 0.632 | 1.115 | 1.508 | 9.02% | - |
| 1 | 63-819 | 819-1008 | 0.276 | 2.934 | 7.953 | 5.03% | - |
| 2 | 126-882 | 882-1071 | 0.364 | 1.971 | 3.098 | 8.67% | - |
| 3 | 189-945 | 945-1134 | 0.212 | 1.774 | 2.974 | 8.69% | - |
| 4 | 252-1008 | 1008-1197 | 0.305 | 0.791 | 0.809 | 18.28% | - |
| 5 | 315-1071 | 1071-1260 | 0.339 | 0.816 | 0.840 | 18.10% | - |
| 6 | 378-1134 | 1134-1323 | 0.523 | 1.107 | 1.190 | 18.10% | - |
| 7 | 441-1197 | 1197-1386 | 0.490 | 2.355 | 5.610 | 5.61% | - |
| 8 | 504-1260 | 1260-1449 | 0.781 | 1.756 | 3.222 | 7.31% | - |
