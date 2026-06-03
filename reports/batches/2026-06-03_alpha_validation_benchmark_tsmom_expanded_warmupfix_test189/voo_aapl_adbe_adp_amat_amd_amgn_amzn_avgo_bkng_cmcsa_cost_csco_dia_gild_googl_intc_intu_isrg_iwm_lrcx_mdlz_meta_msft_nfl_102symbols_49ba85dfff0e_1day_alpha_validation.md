# Alpha Validation Report

- Generated: `2026-06-03T07:47:33Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLC, XLE, XLF, XLI, XLK, XLP, XLRE, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
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
| equal_weight | 128.33% | 15.26% | 0.927 | 0.925 | 0.581 | 26.25% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_tsmom_checkpoint | buy_hold | false | 156.90% | 17.62% | 1.025 | 1.025 | 0.849 | 20.75% | 1.000 | 0.444 | 129 | PBO 0.444 above 0.200 |

## benchmark_tsmom_checkpoint

- Family: `benchmark_tsmom`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.444 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 156.90% | 17.62% | 1.025 | 1.025 | 0.849 | 20.75% | 129 | 13.372 | - |
| stress_2x | 156.50% | 17.59% | 1.023 | 1.023 | 0.847 | 20.76% | 129 | 13.362 | - |
| stress_3x | 156.09% | 17.56% | 1.022 | 1.021 | 0.845 | 20.78% | 129 | 13.351 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-945 | 0.927 | 1.451 | 2.348 | 8.92% | - |
| 1 | 63-819 | 819-1008 | 0.497 | 2.911 | 6.907 | 6.59% | - |
| 2 | 126-882 | 882-1071 | 0.645 | 1.625 | 2.238 | 10.67% | - |
| 3 | 189-945 | 945-1134 | 0.521 | 1.540 | 2.021 | 10.67% | - |
| 4 | 252-1008 | 1008-1197 | 0.527 | 0.383 | 0.313 | 19.57% | - |
| 5 | 315-1071 | 1071-1260 | 0.482 | 0.636 | 0.604 | 19.57% | - |
| 6 | 378-1134 | 1134-1323 | 0.639 | 0.956 | 0.994 | 19.57% | - |
| 7 | 441-1197 | 1197-1386 | 0.596 | 1.843 | 3.801 | 6.76% | - |
| 8 | 504-1260 | 1260-1449 | 0.820 | 1.205 | 1.443 | 12.09% | - |
