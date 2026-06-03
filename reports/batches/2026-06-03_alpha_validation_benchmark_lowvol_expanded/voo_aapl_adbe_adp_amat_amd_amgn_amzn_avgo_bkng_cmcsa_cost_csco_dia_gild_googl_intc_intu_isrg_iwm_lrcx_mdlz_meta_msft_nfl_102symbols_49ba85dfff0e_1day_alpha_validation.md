# Alpha Validation Report

- Generated: `2026-06-03T07:59:15Z`
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
| benchmark_lowvol_checkpoint | buy_hold | false | 109.85% | 13.60% | 0.927 | 0.893 | 0.573 | 23.72% | 1.000 | 0.000 | 332 | turnover increases without return improvement |

## benchmark_lowvol_checkpoint

- Family: `benchmark_lowvol`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - turnover increases without return improvement

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 109.85% | 13.60% | 0.927 | 0.893 | 0.573 | 23.72% | 332 | 31.366 | - |
| stress_2x | 108.94% | 13.51% | 0.922 | 0.888 | 0.569 | 23.77% | 332 | 31.296 | - |
| stress_3x | 108.05% | 13.43% | 0.917 | 0.884 | 0.564 | 23.81% | 333 | 31.225 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.719 | 1.678 | 2.105 | 9.04% | - |
| 1 | 126-882 | 882-1134 | 0.524 | 1.717 | 3.248 | 6.28% | - |
| 2 | 252-1008 | 1008-1260 | 0.493 | 0.653 | 0.564 | 18.14% | - |
| 3 | 378-1134 | 1134-1386 | 0.631 | 0.731 | 0.613 | 18.14% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lowvol_voladj | 0.721 | 2.171 | 2 | benchmark_lowvol_wider | 2.218 | 4 | false |
| 1 | benchmark_lowvol_voladj | 0.503 | 3.597 | 1 | benchmark_lowvol_voladj | 3.597 | 4 | false |
| 2 | benchmark_lowvol_voladj | 0.463 | 0.573 | 2 | benchmark_lowvol_reb42 | 0.635 | 4 | false |
| 3 | benchmark_lowvol_voladj | 0.614 | 0.652 | 1 | benchmark_lowvol_voladj | 0.652 | 4 | false |
