# Alpha Validation Report

- Generated: `2026-06-03T07:15:23Z`
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
| buy_hold | 106.05% | 14.40% | 0.895 | 0.880 | 0.569 | 25.32% | 0 | 1.000 |
| equal_weight | 75.31% | 11.01% | 0.704 | 0.667 | 0.382 | 28.84% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| composite_momentum_checkpoint | buy_hold | false | 89.68% | 12.65% | 0.804 | 0.766 | 0.443 | 28.56% | 1.000 | 0.286 | 276 | PBO 0.286 above 0.200 |

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
| normal | 89.68% | 12.65% | 0.804 | 0.766 | 0.443 | 28.56% | 276 | 32.595 | - |
| stress_2x | 88.73% | 12.55% | 0.798 | 0.761 | 0.438 | 28.63% | 276 | 32.508 | - |
| stress_3x | 87.77% | 12.44% | 0.792 | 0.756 | 0.434 | 28.70% | 277 | 32.421 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-945 | 0.480 | 2.009 | 3.411 | 8.67% | - |
| 1 | 63-819 | 819-1008 | 0.535 | 1.200 | 1.965 | 8.34% | - |
| 2 | 126-882 | 882-1071 | 0.530 | -0.365 | -0.480 | 20.12% | - |
| 3 | 189-945 | 945-1134 | 0.603 | 0.271 | 0.181 | 19.69% | - |
| 4 | 252-1008 | 1008-1197 | 0.438 | 0.659 | 0.615 | 19.69% | - |
| 5 | 315-1071 | 1071-1260 | 0.209 | 2.554 | 6.693 | 5.51% | - |
| 6 | 378-1134 | 1134-1323 | 0.889 | 1.567 | 3.526 | 5.13% | - |
