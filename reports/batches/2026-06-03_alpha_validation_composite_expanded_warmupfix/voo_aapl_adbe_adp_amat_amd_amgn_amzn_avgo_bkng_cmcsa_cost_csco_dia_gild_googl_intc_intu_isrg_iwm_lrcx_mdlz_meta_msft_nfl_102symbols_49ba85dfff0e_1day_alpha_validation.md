# Alpha Validation Report

- Generated: `2026-06-03T07:43:11Z`
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
| composite_momentum_checkpoint | buy_hold | false | 144.30% | 16.61% | 1.004 | 0.985 | 0.621 | 26.73% | 1.000 | 0.250 | 318 | PBO 0.250 above 0.200 |

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
| normal | 144.30% | 16.61% | 1.004 | 0.985 | 0.621 | 26.73% | 318 | 40.412 | - |
| stress_2x | 143.01% | 16.50% | 0.998 | 0.979 | 0.616 | 26.79% | 318 | 40.296 | - |
| stress_3x | 141.72% | 16.40% | 0.993 | 0.974 | 0.611 | 26.84% | 318 | 40.181 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.701 | 1.323 | 2.135 | 8.87% | - |
| 1 | 126-882 | 882-1134 | 0.416 | 1.431 | 1.769 | 12.45% | - |
| 2 | 252-1008 | 1008-1260 | 0.382 | 1.041 | 1.049 | 17.70% | - |
| 3 | 378-1134 | 1134-1386 | 0.577 | 0.979 | 0.945 | 17.67% | - |
