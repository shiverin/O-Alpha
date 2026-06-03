# Alpha Validation Report

- Generated: `2026-06-03T07:45:02Z`
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
| composite_momentum_checkpoint | buy_hold | false | 88.63% | 12.54% | 0.799 | 0.764 | 0.439 | 28.55% | 1.000 | 0.286 | 297 | PBO 0.286 above 0.200 |

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
| normal | 88.63% | 12.54% | 0.799 | 0.764 | 0.439 | 28.55% | 297 | 32.867 | - |
| stress_2x | 87.66% | 12.43% | 0.793 | 0.759 | 0.434 | 28.61% | 297 | 32.777 | - |
| stress_3x | 86.69% | 12.32% | 0.787 | 0.753 | 0.430 | 28.66% | 298 | 32.688 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-945 | 0.464 | 1.716 | 2.680 | 9.55% | - |
| 1 | 63-819 | 819-1008 | 0.521 | 0.657 | 0.888 | 9.55% | - |
| 2 | 126-882 | 882-1071 | 0.505 | -0.421 | -0.538 | 19.99% | - |
| 3 | 189-945 | 945-1134 | 0.559 | 0.234 | 0.138 | 19.99% | - |
| 4 | 252-1008 | 1008-1197 | 0.389 | 0.668 | 0.613 | 19.99% | - |
| 5 | 315-1071 | 1071-1260 | 0.172 | 2.589 | 6.440 | 5.48% | - |
| 6 | 378-1134 | 1134-1323 | 0.846 | 1.704 | 3.897 | 5.48% | - |
