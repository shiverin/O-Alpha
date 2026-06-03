# Alpha Validation Report

- Generated: `2026-06-03T07:42:23Z`
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
| xsec_momentum_top15 | equal_weight | false | 46.93% | 6.84% | 0.602 | 0.519 | 0.464 | 14.73% | 1.000 | 0.500 | 558 | PBO 0.500 above 0.200 |

## xsec_momentum_top15

- Family: `xsec_momentum`
- Benchmark: `equal_weight`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.500 above 0.200
  - turnover increases without return improvement

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 46.93% | 6.84% | 0.602 | 0.519 | 0.464 | 14.73% | 558 | 32.302 | - |
| stress_2x | 46.08% | 6.74% | 0.594 | 0.512 | 0.457 | 14.76% | 558 | 32.207 | - |
| stress_3x | 45.25% | 6.63% | 0.586 | 0.505 | 0.449 | 14.78% | 559 | 32.112 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.041 | 1.484 | 2.483 | 8.13% | - |
| 1 | 126-882 | 882-1134 | 0.473 | 1.450 | 2.380 | 8.85% | - |
| 2 | 252-1008 | 1008-1260 | 0.719 | 0.859 | 0.853 | 14.73% | - |
| 3 | 378-1134 | 1134-1386 | 1.130 | 0.323 | 0.257 | 14.73% | - |
