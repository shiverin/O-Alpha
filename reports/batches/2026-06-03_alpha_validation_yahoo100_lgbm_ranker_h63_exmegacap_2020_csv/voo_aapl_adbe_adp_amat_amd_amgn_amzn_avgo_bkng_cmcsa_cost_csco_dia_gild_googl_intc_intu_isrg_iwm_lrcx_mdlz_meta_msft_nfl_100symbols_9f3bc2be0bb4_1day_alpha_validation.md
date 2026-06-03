# Alpha Validation Report

- Generated: `2026-06-03T13:59:33Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2020-01-02` to `2026-06-01`
- Bars: `1611`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 159.21% | 16.08% | 0.831 | 0.783 | 0.473 | 34.00% | 0 | 1.000 |
| equal_weight | 219.23% | 19.92% | 0.953 | 0.904 | 0.594 | 33.53% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_exmegacap | buy_hold | false | 232.37% | 20.68% | 0.982 | 0.940 | 0.608 | 34.00% | 1.000 | 0.400 | 77 | PBO 0.400 above 0.200 |

## benchmark_lgbm_ranker_h63_s15_exmegacap

- Family: `benchmark_lgbm_ranker_h63_exmegacap`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.400 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 232.37% | 20.68% | 0.982 | 0.940 | 0.608 | 34.00% | 77 | 11.063 | - |
| stress_2x | 231.90% | 20.66% | 0.981 | 0.939 | 0.607 | 34.01% | 77 | 11.054 | - |
| stress_3x | 231.44% | 20.63% | 0.980 | 0.938 | 0.607 | 34.01% | 77 | 11.045 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.514 | 1.855 | 2.845 | 10.28% | - |
| 1 | 126-882 | 882-1134 | 0.983 | 2.090 | 2.764 | 10.28% | - |
| 2 | 252-1008 | 1008-1260 | 0.707 | 1.591 | 2.298 | 10.09% | - |
| 3 | 378-1134 | 1134-1386 | 0.684 | 0.843 | 0.904 | 18.44% | - |
| 4 | 504-1260 | 1260-1512 | 0.556 | 1.375 | 1.555 | 18.44% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.309 | 2.845 | 1 | benchmark_lgbm_ranker_h63_s15_exmegacap | 2.845 | 4 | false |
| 1 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.736 | 2.764 | 2 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | 2.764 | 4 | false |
| 2 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.485 | 2.298 | 3 | benchmark_lgbm_ranker_h63_s10_exmegacap | 2.623 | 4 | true |
| 3 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.460 | 0.904 | 1 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.904 | 4 | false |
| 4 | benchmark_lgbm_ranker_h63_s10_exmegacap | 0.392 | 1.421 | 3 | benchmark_lgbm_ranker_h63_s15_exmegacap | 1.555 | 4 | true |
