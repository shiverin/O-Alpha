# Alpha Validation Report

- Generated: `2026-06-03T07:09:55Z`
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
| buy_hold | 135.68% | 15.89% | 0.972 | 0.950 | 0.628 | 25.32% | 3 | 1.000 |
| equal_weight | 128.60% | 15.28% | 0.892 | 0.851 | 0.493 | 30.98% | 30263 | 23.407 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| composite_momentum_checkpoint | buy_hold | false | 146.20% | 16.76% | 1.011 | 0.986 | 0.629 | 26.65% | 1.000 | 0.250 | 4420 | PBO 0.250 above 0.200 |

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
| normal | 146.20% | 16.76% | 1.011 | 0.986 | 0.629 | 26.65% | 4420 | 45.570 | - |
| stress_2x | 144.73% | 16.64% | 1.005 | 0.979 | 0.623 | 26.71% | 4424 | 45.420 | - |
| stress_3x | 143.26% | 16.52% | 0.999 | 0.973 | 0.617 | 26.77% | 4429 | 45.271 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.712 | 1.412 | 2.236 | 9.02% | - |
| 1 | 126-882 | 882-1134 | 0.446 | 1.507 | 2.015 | 11.00% | - |
| 2 | 252-1008 | 1008-1260 | 0.412 | 1.323 | 1.376 | 17.87% | - |
| 3 | 378-1134 | 1134-1386 | 0.616 | 0.955 | 0.933 | 17.55% | - |
