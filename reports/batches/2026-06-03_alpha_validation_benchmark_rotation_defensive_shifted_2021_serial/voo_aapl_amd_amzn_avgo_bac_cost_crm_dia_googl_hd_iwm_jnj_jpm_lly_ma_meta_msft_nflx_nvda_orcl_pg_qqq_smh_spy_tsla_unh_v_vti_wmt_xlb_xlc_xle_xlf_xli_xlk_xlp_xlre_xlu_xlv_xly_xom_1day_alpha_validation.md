# Alpha Validation Report

- Generated: `2026-06-03T07:36:37Z`
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
| benchmark_rotation_defensive | buy_hold | false | 70.68% | 10.46% | 0.727 | 0.686 | 0.391 | 26.76% | 1.000 | 0.286 | 237 | PBO 0.286 above 0.200 |

## benchmark_rotation_defensive

- Family: `composite_momentum_defensive`
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
| normal | 70.68% | 10.46% | 0.727 | 0.686 | 0.391 | 26.76% | 237 | 24.178 | - |
| stress_2x | 69.99% | 10.38% | 0.722 | 0.681 | 0.386 | 26.85% | 237 | 24.129 | - |
| stress_3x | 69.29% | 10.29% | 0.717 | 0.676 | 0.382 | 26.95% | 237 | 24.080 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-945 | 0.398 | 1.982 | 3.189 | 8.54% | - |
| 1 | 63-819 | 819-1008 | 0.397 | 1.303 | 2.133 | 8.34% | - |
| 2 | 126-882 | 882-1071 | 0.417 | -0.275 | -0.395 | 19.69% | - |
| 3 | 189-945 | 945-1134 | 0.439 | 0.507 | 0.457 | 19.43% | - |
| 4 | 252-1008 | 1008-1197 | 0.282 | 0.854 | 0.867 | 19.43% | - |
| 5 | 315-1071 | 1071-1260 | 0.295 | 2.643 | 7.179 | 5.31% | - |
| 6 | 378-1134 | 1134-1323 | 0.838 | 1.475 | 3.085 | 5.69% | - |
