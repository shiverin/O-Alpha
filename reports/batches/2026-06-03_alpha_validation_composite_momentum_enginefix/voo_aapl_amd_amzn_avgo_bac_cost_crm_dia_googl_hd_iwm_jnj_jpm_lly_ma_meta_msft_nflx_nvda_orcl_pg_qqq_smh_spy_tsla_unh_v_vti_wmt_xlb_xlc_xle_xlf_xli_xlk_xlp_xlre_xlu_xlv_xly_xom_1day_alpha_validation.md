# Alpha Validation Report

- Generated: `2026-06-03T07:14:41Z`
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
| composite_momentum_checkpoint | buy_hold | false | 146.20% | 16.76% | 1.011 | 0.987 | 0.628 | 26.67% | 1.000 | 0.250 | 294 | PBO 0.250 above 0.200 |

## composite_momentum_checkpoint

- Family: `composite_momentum`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.250 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 146.20% | 16.76% | 1.011 | 0.987 | 0.628 | 26.67% | 294 | 40.500 | - |
| stress_2x | 144.91% | 16.66% | 1.006 | 0.981 | 0.623 | 26.73% | 294 | 40.384 | - |
| stress_3x | 143.63% | 16.55% | 1.000 | 0.976 | 0.618 | 26.78% | 295 | 40.269 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.711 | 1.427 | 2.258 | 9.02% | - |
| 1 | 126-882 | 882-1134 | 0.444 | 1.505 | 2.028 | 10.91% | - |
| 2 | 252-1008 | 1008-1260 | 0.414 | 1.321 | 1.372 | 17.90% | - |
| 3 | 378-1134 | 1134-1386 | 0.621 | 0.944 | 0.920 | 17.55% | - |
