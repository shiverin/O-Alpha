# Alpha Validation Report

- Generated: `2026-06-03T13:58:42Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2018-01-02` to `2026-06-01`
- Bars: `2114`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 222.00% | 14.97% | 0.820 | 0.767 | 0.440 | 34.00% | 0 | 1.000 |
| equal_weight | 340.32% | 19.34% | 0.953 | 0.908 | 0.587 | 32.94% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_exmegacap | buy_hold | false | 320.88% | 18.70% | 0.951 | 0.906 | 0.551 | 33.95% | 1.000 | 0.444 | 106 | PBO 0.444 above 0.200 |

## benchmark_lgbm_ranker_h63_s15_exmegacap

- Family: `benchmark_lgbm_ranker_h63_exmegacap`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.444 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 320.88% | 18.70% | 0.951 | 0.906 | 0.551 | 33.95% | 106 | 16.767 | - |
| stress_2x | 320.11% | 18.67% | 0.950 | 0.905 | 0.550 | 33.95% | 106 | 16.749 | - |
| stress_3x | 319.34% | 18.64% | 0.949 | 0.904 | 0.549 | 33.95% | 106 | 16.730 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.688 | 2.309 | 8.366 | 4.29% | - |
| 1 | 126-882 | 882-1134 | 0.931 | -0.345 | -0.386 | 22.29% | - |
| 2 | 252-1008 | 1008-1260 | 1.219 | -0.580 | -0.664 | 24.14% | - |
| 3 | 378-1134 | 1134-1386 | 0.555 | 1.140 | 1.428 | 16.57% | - |
| 4 | 504-1260 | 1260-1512 | 0.511 | 1.840 | 2.739 | 10.48% | - |
| 5 | 630-1386 | 1386-1638 | 1.022 | 2.197 | 2.866 | 10.48% | - |
| 6 | 756-1512 | 1512-1764 | 0.745 | 1.599 | 2.309 | 10.07% | - |
| 7 | 882-1638 | 1638-1890 | 0.722 | 0.813 | 0.833 | 18.44% | - |
| 8 | 1008-1764 | 1764-2016 | 0.580 | 1.468 | 1.610 | 18.44% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.430 | 8.366 | 1 | benchmark_lgbm_ranker_h63_s15_exmegacap | 8.366 | 4 | false |
| 1 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.627 | -0.386 | 2 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | -0.386 | 4 | false |
| 2 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | 0.853 | -0.664 | 2 | benchmark_lgbm_ranker_h63_s15_exmegacap | -0.664 | 4 | false |
| 3 | benchmark_lgbm_ranker_h63_s10_exmegacap | 0.335 | 1.270 | 4 | benchmark_lgbm_ranker_h63_s15_exmegacap | 1.428 | 4 | true |
| 4 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.306 | 2.739 | 1 | benchmark_lgbm_ranker_h63_s15_exmegacap | 2.739 | 4 | false |
| 5 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | 0.788 | 2.866 | 1 | benchmark_lgbm_ranker_h63_s15_exmegacap | 2.866 | 4 | false |
| 6 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.518 | 2.309 | 3 | benchmark_lgbm_ranker_h63_s10_exmegacap | 2.656 | 4 | true |
| 7 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.500 | 0.833 | 3 | benchmark_lgbm_ranker_h63_s10_z125_exmegacap | 0.862 | 4 | true |
| 8 | benchmark_lgbm_ranker_h63_s10_exmegacap | 0.398 | 1.504 | 3 | benchmark_lgbm_ranker_h63_s15_exmegacap | 1.610 | 4 | true |
