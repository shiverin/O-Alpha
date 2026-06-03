# Alpha Validation Report

- Generated: `2026-06-03T08:11:21Z`
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
| benchmark_reversal_checkpoint | buy_hold | false | 144.38% | 16.61% | 0.999 | 0.979 | 0.684 | 24.27% | 1.000 | 0.500 | 2263 | PBO 0.500 above 0.200 |

## benchmark_reversal_checkpoint

- Family: `benchmark_reversal`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.500 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 144.38% | 16.61% | 0.999 | 0.979 | 0.684 | 24.27% | 2263 | 91.915 | - |
| stress_2x | 141.53% | 16.38% | 0.987 | 0.968 | 0.672 | 24.39% | 2265 | 91.315 | - |
| stress_3x | 138.71% | 16.14% | 0.975 | 0.955 | 0.659 | 24.50% | 2270 | 90.721 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.779 | 1.688 | 2.562 | 8.16% | - |
| 1 | 126-882 | 882-1134 | 0.479 | 1.802 | 2.881 | 8.65% | - |
| 2 | 252-1008 | 1008-1260 | 0.442 | 1.019 | 1.116 | 17.80% | - |
| 3 | 378-1134 | 1134-1386 | 0.700 | 0.794 | 0.753 | 18.62% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_reversal_medium | 0.598 | 2.322 | 3 | benchmark_reversal_checkpoint | 2.562 | 4 | true |
| 1 | benchmark_reversal_slow | 0.397 | 3.502 | 1 | benchmark_reversal_slow | 3.502 | 4 | false |
| 2 | benchmark_reversal_slow | 0.355 | 0.808 | 4 | benchmark_reversal_checkpoint | 1.116 | 4 | true |
| 3 | benchmark_reversal_fast | 0.569 | 0.893 | 1 | benchmark_reversal_fast | 0.893 | 4 | false |
