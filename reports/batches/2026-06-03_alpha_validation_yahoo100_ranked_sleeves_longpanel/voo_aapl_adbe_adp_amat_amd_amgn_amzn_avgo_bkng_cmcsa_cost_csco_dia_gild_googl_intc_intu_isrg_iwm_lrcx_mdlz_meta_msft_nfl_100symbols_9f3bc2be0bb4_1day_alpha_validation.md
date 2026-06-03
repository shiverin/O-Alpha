# Alpha Validation Report

- Generated: `2026-06-03T08:52:08Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2015-01-02` to `2026-06-01`
- Bars: `2869`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 351.93% | 14.17% | 0.837 | 0.790 | 0.417 | 34.00% | 0 | 1.000 |
| equal_weight | 1170.53% | 25.03% | 1.087 | 1.039 | 0.673 | 37.17% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_ranked_sleeve_checkpoint | buy_hold | false | 474.41% | 16.60% | 0.944 | 0.888 | 0.482 | 34.45% | 1.000 | 0.400 | 561 | PBO 0.400 above 0.200 |
| sector_ranked_sleeve_checkpoint | buy_hold | false | 372.89% | 14.63% | 0.861 | 0.814 | 0.431 | 33.92% | 1.000 | 0.600 | 312 | PBO 0.600 above 0.200 |

## benchmark_ranked_sleeve_checkpoint

- Family: `benchmark_ranked_sleeve`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.400 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 474.41% | 16.60% | 0.944 | 0.888 | 0.482 | 34.45% | 561 | 104.657 | - |
| stress_2x | 469.61% | 16.52% | 0.940 | 0.883 | 0.479 | 34.46% | 561 | 104.105 | - |
| stress_3x | 464.85% | 16.43% | 0.936 | 0.879 | 0.477 | 34.47% | 561 | 103.557 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.030 | -0.197 | -0.275 | 19.54% | - |
| 1 | 126-882 | 882-1134 | 1.054 | 0.660 | 0.494 | 19.54% | - |
| 2 | 252-1008 | 1008-1260 | 0.751 | 2.586 | 5.362 | 6.34% | - |
| 3 | 378-1134 | 1134-1386 | 1.182 | 0.666 | 0.529 | 34.45% | - |
| 4 | 504-1260 | 1260-1512 | 1.211 | 0.843 | 0.762 | 34.45% | - |
| 5 | 630-1386 | 1386-1638 | 0.720 | 2.323 | 4.425 | 10.20% | - |
| 6 | 756-1512 | 1512-1764 | 0.784 | 1.923 | 5.090 | 5.97% | - |
| 7 | 882-1638 | 1638-1890 | 1.023 | -0.399 | -0.425 | 20.86% | - |
| 8 | 1008-1764 | 1764-2016 | 1.278 | -0.675 | -0.643 | 24.04% | - |
| 9 | 1134-1890 | 1890-2142 | 0.658 | 0.867 | 0.966 | 14.98% | - |
| 10 | 1260-2016 | 2016-2268 | 0.477 | 1.572 | 2.311 | 9.98% | - |
| 11 | 1386-2142 | 2142-2394 | 0.807 | 2.275 | 3.581 | 9.98% | - |
| 12 | 1512-2268 | 2268-2520 | 0.696 | 1.785 | 2.976 | 9.70% | - |
| 13 | 1638-2394 | 2394-2646 | 0.820 | 0.649 | 0.603 | 18.01% | - |
| 14 | 1764-2520 | 2520-2772 | 0.640 | 1.216 | 1.184 | 18.01% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranked_sleeve_slow | 1.063 | -0.289 | 4 | benchmark_ranked_sleeve_conservative | -0.189 | 4 | true |
| 1 | benchmark_ranked_sleeve_slow | 1.370 | 0.422 | 4 | benchmark_ranked_sleeve_conservative | 0.889 | 4 | true |
| 2 | benchmark_ranked_sleeve_conservative | 0.655 | 3.934 | 4 | benchmark_ranked_sleeve_checkpoint | 5.362 | 4 | true |
| 3 | benchmark_ranked_sleeve_conservative | 1.015 | 0.358 | 3 | benchmark_ranked_sleeve_checkpoint | 0.529 | 4 | true |
| 4 | benchmark_ranked_sleeve_conservative | 0.933 | 0.910 | 1 | benchmark_ranked_sleeve_conservative | 0.910 | 4 | false |
| 5 | benchmark_ranked_sleeve_checkpoint | 0.427 | 4.425 | 2 | benchmark_ranked_sleeve_conservative | 4.944 | 4 | false |
| 6 | benchmark_ranked_sleeve_conservative | 0.510 | 5.453 | 1 | benchmark_ranked_sleeve_conservative | 5.453 | 4 | false |
| 7 | benchmark_ranked_sleeve_checkpoint | 0.683 | -0.425 | 2 | benchmark_ranked_sleeve_slow | -0.291 | 4 | false |
| 8 | benchmark_ranked_sleeve_checkpoint | 0.883 | -0.643 | 1 | benchmark_ranked_sleeve_checkpoint | -0.643 | 4 | false |
| 9 | benchmark_ranked_sleeve_conservative | 0.462 | 1.022 | 1 | benchmark_ranked_sleeve_conservative | 1.022 | 4 | false |
| 10 | benchmark_ranked_sleeve_conservative | 0.290 | 3.094 | 1 | benchmark_ranked_sleeve_conservative | 3.094 | 4 | false |
| 11 | benchmark_ranked_sleeve_slow | 0.635 | 4.426 | 1 | benchmark_ranked_sleeve_slow | 4.426 | 4 | false |
| 12 | benchmark_ranked_sleeve_slow | 0.454 | 3.946 | 1 | benchmark_ranked_sleeve_slow | 3.946 | 4 | false |
| 13 | benchmark_ranked_sleeve_slow | 0.571 | 0.428 | 4 | benchmark_ranked_sleeve_conservative | 0.673 | 4 | true |
| 14 | benchmark_ranked_sleeve_slow | 0.494 | 0.673 | 4 | benchmark_ranked_sleeve_conservative | 1.652 | 4 | true |

## sector_ranked_sleeve_checkpoint

- Family: `sector_ranked_sleeve`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.600 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 372.89% | 14.63% | 0.861 | 0.814 | 0.431 | 33.92% | 312 | 57.933 | - |
| stress_2x | 370.51% | 14.58% | 0.859 | 0.811 | 0.430 | 33.93% | 312 | 57.753 | - |
| stress_3x | 368.13% | 14.53% | 0.856 | 0.809 | 0.428 | 33.94% | 312 | 57.575 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.957 | -0.385 | -0.428 | 18.65% | - |
| 1 | 126-882 | 882-1134 | 0.864 | 0.733 | 0.550 | 18.65% | - |
| 2 | 252-1008 | 1008-1260 | 0.702 | 2.496 | 5.499 | 5.84% | - |
| 3 | 378-1134 | 1134-1386 | 1.159 | 0.540 | 0.390 | 33.92% | - |
| 4 | 504-1260 | 1260-1512 | 1.186 | 0.700 | 0.587 | 33.92% | - |
| 5 | 630-1386 | 1386-1638 | 0.646 | 2.206 | 4.019 | 10.18% | - |
| 6 | 756-1512 | 1512-1764 | 0.708 | 1.766 | 5.580 | 4.82% | - |
| 7 | 882-1638 | 1638-1890 | 0.927 | -0.304 | -0.383 | 19.22% | - |
| 8 | 1008-1764 | 1764-2016 | 1.152 | -0.565 | -0.691 | 20.56% | - |
| 9 | 1134-1890 | 1890-2142 | 0.611 | 0.701 | 0.754 | 15.76% | - |
| 10 | 1260-2016 | 2016-2268 | 0.458 | 1.543 | 2.134 | 10.30% | - |
| 11 | 1386-2142 | 2142-2394 | 0.796 | 2.291 | 3.124 | 10.30% | - |
| 12 | 1512-2268 | 2268-2520 | 0.671 | 1.728 | 2.601 | 9.79% | - |
| 13 | 1638-2394 | 2394-2646 | 0.725 | 0.518 | 0.465 | 18.28% | - |
| 14 | 1764-2520 | 2520-2772 | 0.520 | 1.068 | 1.066 | 18.28% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | sector_ranked_sleeve_conservative | 1.055 | -0.402 | 1 | sector_ranked_sleeve_conservative | -0.402 | 4 | false |
| 1 | sector_ranked_sleeve_conservative | 1.038 | 0.521 | 3 | sector_ranked_sleeve_checkpoint | 0.550 | 4 | true |
| 2 | sector_ranked_sleeve_checkpoint | 0.475 | 5.499 | 1 | sector_ranked_sleeve_checkpoint | 5.499 | 4 | false |
| 3 | sector_ranked_sleeve_conservative | 0.784 | 0.282 | 4 | sector_ranked_sleeve_checkpoint | 0.390 | 4 | true |
| 4 | sector_ranked_sleeve_checkpoint | 0.831 | 0.587 | 2 | sector_ranked_sleeve_slow | 0.594 | 4 | false |
| 5 | sector_ranked_sleeve_checkpoint | 0.371 | 4.019 | 4 | sector_ranked_sleeve_slow | 4.216 | 4 | true |
| 6 | sector_ranked_sleeve_checkpoint | 0.433 | 5.580 | 2 | sector_ranked_sleeve_medium | 5.693 | 4 | false |
| 7 | sector_ranked_sleeve_checkpoint | 0.605 | -0.383 | 4 | sector_ranked_sleeve_medium | -0.379 | 4 | true |
| 8 | sector_ranked_sleeve_slow | 0.798 | -0.664 | 1 | sector_ranked_sleeve_slow | -0.664 | 4 | false |
| 9 | sector_ranked_sleeve_medium | 0.371 | 0.779 | 3 | sector_ranked_sleeve_conservative | 0.969 | 4 | true |
| 10 | sector_ranked_sleeve_slow | 0.269 | 2.157 | 3 | sector_ranked_sleeve_conservative | 2.555 | 4 | true |
| 11 | sector_ranked_sleeve_conservative | 0.663 | 2.792 | 4 | sector_ranked_sleeve_slow | 3.162 | 4 | true |
| 12 | sector_ranked_sleeve_conservative | 0.524 | 2.288 | 4 | sector_ranked_sleeve_slow | 2.830 | 4 | true |
| 13 | sector_ranked_sleeve_conservative | 0.578 | 0.523 | 1 | sector_ranked_sleeve_conservative | 0.523 | 4 | false |
| 14 | sector_ranked_sleeve_slow | 0.349 | 0.841 | 4 | sector_ranked_sleeve_conservative | 1.103 | 4 | true |
