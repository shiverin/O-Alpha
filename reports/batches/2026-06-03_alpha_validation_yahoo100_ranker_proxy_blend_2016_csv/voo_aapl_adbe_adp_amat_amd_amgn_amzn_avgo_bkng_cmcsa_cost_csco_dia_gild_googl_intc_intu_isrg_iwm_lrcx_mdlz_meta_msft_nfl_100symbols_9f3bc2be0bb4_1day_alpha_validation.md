# Alpha Validation Report

- Generated: `2026-06-03T11:50:42Z`
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
| benchmark_ranker_proxy_blend_checkpoint | buy_hold | false | 433.18% | 17.50% | 0.989 | 0.920 | 0.490 | 35.69% | 1.000 | 0.538 | 233 | PBO 0.538 above 0.200 |

## benchmark_ranker_proxy_blend_checkpoint

- Family: `benchmark_ranker_proxy_blend`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.538 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 433.18% | 17.50% | 0.989 | 0.920 | 0.490 | 35.69% | 233 | 26.567 | - |
| stress_2x | 432.03% | 17.47% | 0.988 | 0.919 | 0.489 | 35.70% | 233 | 26.533 | - |
| stress_3x | 430.88% | 17.45% | 0.987 | 0.918 | 0.489 | 35.70% | 233 | 26.499 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.892 | 2.313 | 4.878 | 6.22% | - |
| 1 | 126-882 | 882-1134 | 1.311 | 0.488 | 0.320 | 35.69% | - |
| 2 | 252-1008 | 1008-1260 | 1.235 | 0.683 | 0.547 | 35.69% | - |
| 3 | 378-1134 | 1134-1386 | 0.674 | 2.374 | 4.377 | 9.93% | - |
| 4 | 504-1260 | 1260-1512 | 0.703 | 2.304 | 6.295 | 5.77% | - |
| 5 | 630-1386 | 1386-1638 | 0.923 | -0.238 | -0.289 | 21.45% | - |
| 6 | 756-1512 | 1512-1764 | 1.221 | -0.686 | -0.710 | 23.03% | - |
| 7 | 882-1638 | 1638-1890 | 0.674 | 0.938 | 1.087 | 15.46% | - |
| 8 | 1008-1764 | 1764-2016 | 0.509 | 1.940 | 3.078 | 9.38% | - |
| 9 | 1134-1890 | 1890-2142 | 0.950 | 2.407 | 3.411 | 9.38% | - |
| 10 | 1260-2016 | 2016-2268 | 0.779 | 1.827 | 2.760 | 9.25% | - |
| 11 | 1386-2142 | 2142-2394 | 0.852 | 0.518 | 0.455 | 18.50% | - |
| 12 | 1512-2268 | 2268-2520 | 0.651 | 1.045 | 1.013 | 18.50% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranker_proxy_blend_h63_h84 | 0.656 | 4.643 | 5 | benchmark_ranker_proxy_blend_checkpoint | 4.878 | 5 | true |
| 1 | benchmark_ranker_proxy_blend_h63_h126 | 0.955 | 0.305 | 5 | benchmark_ranker_proxy_blend_h63_h84 | 0.332 | 5 | true |
| 2 | benchmark_ranker_proxy_blend_h63_h126 | 0.900 | 0.551 | 1 | benchmark_ranker_proxy_blend_h63_h126 | 0.551 | 5 | false |
| 3 | benchmark_ranker_proxy_blend_h63_h126 | 0.386 | 4.408 | 2 | benchmark_ranker_proxy_blend_h63_h84 | 4.501 | 5 | false |
| 4 | benchmark_ranker_proxy_blend_h63_h84 | 0.420 | 6.166 | 5 | benchmark_ranker_proxy_blend_h63_h126 | 6.440 | 5 | true |
| 5 | benchmark_ranker_proxy_blend_h63_h84 | 0.589 | -0.298 | 4 | benchmark_ranker_proxy_blend_h63_h126 | -0.285 | 5 | true |
| 6 | benchmark_ranker_proxy_blend_h63_h84 | 0.810 | -0.705 | 1 | benchmark_ranker_proxy_blend_h63_h84 | -0.705 | 5 | false |
| 7 | benchmark_ranker_proxy_blend_h63_h84 | 0.409 | 1.088 | 2 | benchmark_ranker_proxy_blend_h84_h126 | 1.114 | 5 | false |
| 8 | benchmark_ranker_proxy_blend_h63_h84 | 0.298 | 3.210 | 1 | benchmark_ranker_proxy_blend_h63_h84 | 3.210 | 5 | false |
| 9 | benchmark_ranker_proxy_blend_h63_h84 | 0.721 | 3.538 | 1 | benchmark_ranker_proxy_blend_h63_h84 | 3.538 | 5 | false |
| 10 | benchmark_ranker_proxy_blend_h84_h126 | 0.559 | 2.647 | 4 | benchmark_ranker_proxy_blend_h63_h84 | 2.908 | 5 | true |
| 11 | benchmark_ranker_proxy_blend_h63_h84 | 0.625 | 0.420 | 5 | benchmark_ranker_proxy_blend_h63_h126 | 0.524 | 5 | true |
| 12 | benchmark_ranker_proxy_blend_h63_h84 | 0.437 | 0.988 | 5 | benchmark_ranker_proxy_blend_h63_h126 | 1.073 | 5 | true |
