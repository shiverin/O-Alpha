# Alpha Validation Report

- Generated: `2026-06-03T07:15:28Z`
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
| composite_momentum_checkpoint | buy_hold | false | 146.20% | 16.76% | 1.011 | 0.987 | 0.628 | 26.67% | 1.000 | 0.222 | 294 | PBO 0.222 above 0.200 |

## composite_momentum_checkpoint

- Family: `composite_momentum`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.222 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 146.20% | 16.76% | 1.011 | 0.987 | 0.628 | 26.67% | 294 | 40.500 | - |
| stress_2x | 144.91% | 16.66% | 1.006 | 0.981 | 0.623 | 26.73% | 294 | 40.384 | - |
| stress_3x | 143.63% | 16.55% | 1.000 | 0.976 | 0.618 | 26.78% | 295 | 40.269 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-945 | 0.711 | 1.265 | 1.912 | 9.02% | - |
| 1 | 63-819 | 819-1008 | 0.385 | 2.622 | 6.100 | 6.80% | - |
| 2 | 126-882 | 882-1071 | 0.444 | 1.611 | 2.121 | 10.91% | - |
| 3 | 189-945 | 945-1134 | 0.369 | 1.488 | 2.095 | 10.95% | - |
| 4 | 252-1008 | 1008-1197 | 0.414 | 0.835 | 0.859 | 17.90% | - |
| 5 | 315-1071 | 1071-1260 | 0.449 | 0.878 | 0.921 | 17.55% | - |
| 6 | 378-1134 | 1134-1323 | 0.621 | 1.198 | 1.322 | 17.55% | - |
| 7 | 441-1197 | 1197-1386 | 0.612 | 2.270 | 5.332 | 5.92% | - |
| 8 | 504-1260 | 1260-1449 | 0.902 | 1.831 | 3.789 | 6.62% | - |
