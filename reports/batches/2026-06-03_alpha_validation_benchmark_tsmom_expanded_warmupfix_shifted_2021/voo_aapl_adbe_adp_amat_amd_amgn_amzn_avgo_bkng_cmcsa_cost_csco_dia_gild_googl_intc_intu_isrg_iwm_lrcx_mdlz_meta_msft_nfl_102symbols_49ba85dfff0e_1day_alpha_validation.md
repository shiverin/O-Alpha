# Alpha Validation Report

- Generated: `2026-06-03T07:48:39Z`
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
| benchmark_tsmom_checkpoint | buy_hold | false | 90.64% | 12.76% | 0.793 | 0.770 | 0.555 | 22.97% | 1.000 | 0.143 | 118 | turnover increases without return improvement |

## benchmark_tsmom_checkpoint

- Family: `benchmark_tsmom`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - turnover increases without return improvement

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 90.64% | 12.76% | 0.793 | 0.770 | 0.555 | 22.97% | 118 | 9.512 | - |
| stress_2x | 90.37% | 12.73% | 0.791 | 0.768 | 0.554 | 22.99% | 118 | 9.505 | - |
| stress_3x | 90.10% | 12.70% | 0.790 | 0.767 | 0.552 | 23.01% | 118 | 9.499 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-945 | 0.573 | 1.444 | 1.936 | 12.60% | - |
| 1 | 63-819 | 819-1008 | 0.580 | 0.593 | 0.671 | 12.60% | - |
| 2 | 126-882 | 882-1071 | 0.627 | -0.479 | -0.642 | 19.49% | - |
| 3 | 189-945 | 945-1134 | 0.529 | 0.422 | 0.351 | 19.49% | - |
| 4 | 252-1008 | 1008-1197 | 0.368 | 0.846 | 0.846 | 19.49% | - |
| 5 | 315-1071 | 1071-1260 | 0.295 | 2.012 | 4.278 | 6.86% | - |
| 6 | 378-1134 | 1134-1323 | 0.820 | 0.937 | 1.164 | 11.88% | - |
