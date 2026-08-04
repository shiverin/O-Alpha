# Alpha Validation Report

- Generated: `2026-06-03T13:57:31Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2016-01-04` to `2026-06-01`
- Bars: `2617`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 348.99% | 15.57% | 0.898 | 0.842 | 0.458 | 34.00% | 0 | 1.000 |
| equal_weight | 932.56% | 25.22% | 1.112 | 1.059 | 0.720 | 35.05% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_exmegacap | buy_hold | false | 528.39% | 19.37% | 1.032 | 0.982 | 0.580 | 33.40% | 1.000 | 0.385 | 109 | PBO 0.385 above 0.200 |

## benchmark_lgbm_ranker_h63_s15_exmegacap

- Family: `benchmark_lgbm_ranker_h63_exmegacap`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.385 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 528.39% | 19.37% | 1.032 | 0.982 | 0.580 | 33.40% | 109 | 23.753 | - |
| stress_2x | 527.23% | 19.35% | 1.031 | 0.981 | 0.579 | 33.41% | 110 | 23.726 | - |
| stress_3x | 526.07% | 19.33% | 1.030 | 0.980 | 0.579 | 33.41% | 110 | 23.699 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.729 | 2.284 | 4.356 | 7.33% | - |
| 1 | 126-882 | 882-1134 | 1.185 | 0.497 | 0.355 | 33.40% | - |
| 2 | 252-1008 | 1008-1260 | 1.152 | 0.674 | 0.576 | 33.40% | - |
| 3 | 378-1134 | 1134-1386 | 0.685 | 2.351 | 4.873 | 8.78% | - |
| 4 | 504-1260 | 1260-1512 | 0.698 | 2.191 | 7.524 | 4.55% | - |
| 5 | 630-1386 | 1386-1638 | 0.931 | -0.376 | -0.406 | 22.61% | - |
| 6 | 756-1512 | 1512-1764 | 1.209 | -0.535 | -0.631 | 23.95% | - |
| 7 | 882-1638 | 1638-1890 | 0.605 | 1.048 | 1.282 | 16.58% | - |
| 8 | 1008-1764 | 1764-2016 | 0.491 | 1.942 | 3.065 | 9.90% | - |
| 9 | 1134-1890 | 1890-2142 | 0.991 | 2.210 | 3.033 | 9.90% | - |
| 10 | 1260-2016 | 2016-2268 | 0.726 | 1.488 | 2.115 | 10.07% | - |
| 11 | 1386-2142 | 2142-2394 | 0.782 | 0.842 | 0.882 | 18.48% | - |
| 12 | 1512-2268 | 2268-2520 | 0.606 | 1.705 | 1.949 | 18.48% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s10_z125_exmegacap | 0.439 | 4.410 | 2 | benchmark_lgbm_ranker_h63_s10_exmegacap | 4.410 | 4 | false |
| 1 | benchmark_lgbm_ranker_h63_s10_exmegacap | 0.745 | 0.304 | 3 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | 0.355 | 4 | true |
| 2 | benchmark_lgbm_ranker_h63_s10_z125_exmegacap | 0.744 | 0.523 | 4 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.576 | 4 | true |
| 3 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | 0.428 | 4.873 | 2 | benchmark_lgbm_ranker_h63_s15_exmegacap | 4.873 | 4 | false |
| 4 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.447 | 7.524 | 3 | benchmark_lgbm_ranker_h63_s10_exmegacap | 7.962 | 4 | true |
| 5 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.642 | -0.406 | 2 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | -0.406 | 4 | false |
| 6 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.857 | -0.631 | 1 | benchmark_lgbm_ranker_h63_s15_exmegacap | -0.631 | 4 | false |
| 7 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.369 | 1.282 | 1 | benchmark_lgbm_ranker_h63_s15_exmegacap | 1.282 | 4 | false |
| 8 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | 0.292 | 3.065 | 2 | benchmark_lgbm_ranker_h63_s15_exmegacap | 3.065 | 4 | false |
| 9 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | 0.753 | 3.033 | 1 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | 3.033 | 4 | false |
| 10 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.497 | 2.115 | 4 | benchmark_lgbm_ranker_h63_s10_z125_exmegacap | 2.436 | 4 | true |
| 11 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.573 | 0.882 | 1 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.882 | 4 | false |
| 12 | benchmark_lgbm_ranker_h63_s10_exmegacap | 0.422 | 1.680 | 4 | benchmark_lgbm_ranker_h63_s15_exmegacap | 1.949 | 4 | true |
