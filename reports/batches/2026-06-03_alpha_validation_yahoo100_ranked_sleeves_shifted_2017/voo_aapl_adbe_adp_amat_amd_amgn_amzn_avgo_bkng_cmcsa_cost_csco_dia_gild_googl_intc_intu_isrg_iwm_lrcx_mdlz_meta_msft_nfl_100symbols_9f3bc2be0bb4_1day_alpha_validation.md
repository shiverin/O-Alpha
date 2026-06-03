# Alpha Validation Report

- Generated: `2026-06-03T08:55:13Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2017-01-03` to `2026-06-01`
- Bars: `2365`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 291.62% | 15.66% | 0.885 | 0.827 | 0.461 | 34.00% | 0 | 1.000 |
| equal_weight | 511.40% | 21.29% | 1.042 | 0.980 | 0.644 | 33.06% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_ranked_sleeve_checkpoint | buy_hold | true | 373.98% | 18.04% | 0.988 | 0.925 | 0.524 | 34.45% | 1.000 | 0.182 | 451 | pass |
| sector_ranked_sleeve_checkpoint | buy_hold | false | 315.30% | 16.39% | 0.921 | 0.862 | 0.483 | 33.92% | 1.000 | 0.636 | 249 | PBO 0.636 above 0.200 |

## benchmark_ranked_sleeve_checkpoint

- Family: `benchmark_ranked_sleeve`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `true`

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 373.98% | 18.04% | 0.988 | 0.925 | 0.524 | 34.45% | 451 | 78.627 | - |
| stress_2x | 370.80% | 17.96% | 0.984 | 0.921 | 0.521 | 34.46% | 451 | 78.308 | - |
| stress_3x | 367.63% | 17.87% | 0.980 | 0.917 | 0.519 | 34.47% | 451 | 77.991 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.211 | 0.843 | 0.762 | 34.45% | - |
| 1 | 126-882 | 882-1134 | 0.720 | 2.323 | 4.425 | 10.20% | - |
| 2 | 252-1008 | 1008-1260 | 0.784 | 1.923 | 5.090 | 5.97% | - |
| 3 | 378-1134 | 1134-1386 | 1.023 | -0.399 | -0.425 | 20.86% | - |
| 4 | 504-1260 | 1260-1512 | 1.278 | -0.675 | -0.643 | 24.04% | - |
| 5 | 630-1386 | 1386-1638 | 0.658 | 0.867 | 0.966 | 14.98% | - |
| 6 | 756-1512 | 1512-1764 | 0.477 | 1.572 | 2.311 | 9.98% | - |
| 7 | 882-1638 | 1638-1890 | 0.807 | 2.275 | 3.581 | 9.98% | - |
| 8 | 1008-1764 | 1764-2016 | 0.696 | 1.785 | 2.976 | 9.70% | - |
| 9 | 1134-1890 | 1890-2142 | 0.820 | 0.649 | 0.603 | 18.01% | - |
| 10 | 1260-2016 | 2016-2268 | 0.640 | 1.216 | 1.184 | 18.01% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranked_sleeve_conservative | 0.933 | 0.910 | 1 | benchmark_ranked_sleeve_conservative | 0.910 | 4 | false |
| 1 | benchmark_ranked_sleeve_checkpoint | 0.427 | 4.425 | 2 | benchmark_ranked_sleeve_conservative | 4.944 | 4 | false |
| 2 | benchmark_ranked_sleeve_conservative | 0.510 | 5.453 | 1 | benchmark_ranked_sleeve_conservative | 5.453 | 4 | false |
| 3 | benchmark_ranked_sleeve_checkpoint | 0.683 | -0.425 | 2 | benchmark_ranked_sleeve_slow | -0.291 | 4 | false |
| 4 | benchmark_ranked_sleeve_checkpoint | 0.883 | -0.643 | 1 | benchmark_ranked_sleeve_checkpoint | -0.643 | 4 | false |
| 5 | benchmark_ranked_sleeve_conservative | 0.462 | 1.022 | 1 | benchmark_ranked_sleeve_conservative | 1.022 | 4 | false |
| 6 | benchmark_ranked_sleeve_conservative | 0.290 | 3.094 | 1 | benchmark_ranked_sleeve_conservative | 3.094 | 4 | false |
| 7 | benchmark_ranked_sleeve_slow | 0.635 | 4.426 | 1 | benchmark_ranked_sleeve_slow | 4.426 | 4 | false |
| 8 | benchmark_ranked_sleeve_slow | 0.454 | 3.946 | 1 | benchmark_ranked_sleeve_slow | 3.946 | 4 | false |
| 9 | benchmark_ranked_sleeve_slow | 0.571 | 0.428 | 4 | benchmark_ranked_sleeve_conservative | 0.673 | 4 | true |
| 10 | benchmark_ranked_sleeve_slow | 0.494 | 0.673 | 4 | benchmark_ranked_sleeve_conservative | 1.652 | 4 | true |

## sector_ranked_sleeve_checkpoint

- Family: `sector_ranked_sleeve`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.636 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 315.30% | 16.39% | 0.921 | 0.862 | 0.483 | 33.92% | 249 | 47.026 | - |
| stress_2x | 313.54% | 16.34% | 0.918 | 0.860 | 0.481 | 33.93% | 249 | 46.908 | - |
| stress_3x | 311.78% | 16.28% | 0.916 | 0.857 | 0.480 | 33.94% | 249 | 46.790 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.186 | 0.700 | 0.587 | 33.92% | - |
| 1 | 126-882 | 882-1134 | 0.646 | 2.206 | 4.019 | 10.18% | - |
| 2 | 252-1008 | 1008-1260 | 0.708 | 1.766 | 5.580 | 4.82% | - |
| 3 | 378-1134 | 1134-1386 | 0.927 | -0.304 | -0.383 | 19.22% | - |
| 4 | 504-1260 | 1260-1512 | 1.152 | -0.565 | -0.691 | 20.56% | - |
| 5 | 630-1386 | 1386-1638 | 0.611 | 0.701 | 0.754 | 15.76% | - |
| 6 | 756-1512 | 1512-1764 | 0.458 | 1.543 | 2.134 | 10.30% | - |
| 7 | 882-1638 | 1638-1890 | 0.796 | 2.291 | 3.124 | 10.30% | - |
| 8 | 1008-1764 | 1764-2016 | 0.671 | 1.728 | 2.601 | 9.79% | - |
| 9 | 1134-1890 | 1890-2142 | 0.725 | 0.518 | 0.465 | 18.28% | - |
| 10 | 1260-2016 | 2016-2268 | 0.520 | 1.068 | 1.066 | 18.28% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | sector_ranked_sleeve_checkpoint | 0.831 | 0.587 | 2 | sector_ranked_sleeve_slow | 0.594 | 4 | false |
| 1 | sector_ranked_sleeve_checkpoint | 0.371 | 4.019 | 4 | sector_ranked_sleeve_slow | 4.216 | 4 | true |
| 2 | sector_ranked_sleeve_checkpoint | 0.433 | 5.580 | 2 | sector_ranked_sleeve_medium | 5.693 | 4 | false |
| 3 | sector_ranked_sleeve_checkpoint | 0.605 | -0.383 | 4 | sector_ranked_sleeve_medium | -0.379 | 4 | true |
| 4 | sector_ranked_sleeve_slow | 0.798 | -0.664 | 1 | sector_ranked_sleeve_slow | -0.664 | 4 | false |
| 5 | sector_ranked_sleeve_medium | 0.371 | 0.779 | 3 | sector_ranked_sleeve_conservative | 0.969 | 4 | true |
| 6 | sector_ranked_sleeve_slow | 0.269 | 2.157 | 3 | sector_ranked_sleeve_conservative | 2.555 | 4 | true |
| 7 | sector_ranked_sleeve_conservative | 0.663 | 2.792 | 4 | sector_ranked_sleeve_slow | 3.162 | 4 | true |
| 8 | sector_ranked_sleeve_conservative | 0.524 | 2.288 | 4 | sector_ranked_sleeve_slow | 2.830 | 4 | true |
| 9 | sector_ranked_sleeve_conservative | 0.578 | 0.523 | 1 | sector_ranked_sleeve_conservative | 0.523 | 4 | false |
| 10 | sector_ranked_sleeve_slow | 0.349 | 0.841 | 4 | sector_ranked_sleeve_conservative | 1.103 | 4 | true |
