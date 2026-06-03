# Alpha Validation Report

- Generated: `2026-06-03T08:39:16Z`
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
| xsec_momentum_top15 | equal_weight | false | 219.59% | 10.75% | 0.843 | 0.740 | 0.496 | 21.67% | 1.000 | 0.400 | 1227 | PBO 0.400 above 0.200 |
| composite_momentum_checkpoint | buy_hold | false | 407.29% | 15.34% | 0.897 | 0.843 | 0.449 | 34.17% | 1.000 | 0.533 | 658 | PBO 0.533 above 0.200 |
| benchmark_tsmom_checkpoint | buy_hold | false | 411.35% | 15.42% | 0.879 | 0.831 | 0.458 | 33.64% | 1.000 | 0.600 | 266 | PBO 0.600 above 0.200 |
| benchmark_tsmom_blend | buy_hold | false | 432.87% | 15.84% | 0.894 | 0.843 | 0.467 | 33.90% | 1.000 | 0.667 | 330 | PBO 0.667 above 0.200 |
| benchmark_lowvol_checkpoint | buy_hold | false | 331.91% | 13.72% | 0.862 | 0.817 | 0.420 | 32.68% | 1.000 | 0.533 | 653 | PBO 0.533 above 0.200 |
| benchmark_reversal_checkpoint | buy_hold | false | 387.62% | 14.94% | 0.865 | 0.816 | 0.424 | 35.24% | 1.000 | 0.533 | 4357 | PBO 0.533 above 0.200 |

## xsec_momentum_top15

- Family: `xsec_momentum`
- Benchmark: `equal_weight`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.400 above 0.200
  - turnover increases without return improvement

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 219.59% | 10.75% | 0.843 | 0.740 | 0.496 | 21.67% | 1227 | 117.990 | - |
| stress_2x | 215.38% | 10.62% | 0.834 | 0.732 | 0.490 | 21.69% | 1227 | 117.098 | - |
| stress_3x | 211.22% | 10.49% | 0.825 | 0.723 | 0.483 | 21.71% | 1228 | 116.215 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.331 | 0.020 | -0.076 | 15.17% | - |
| 1 | 126-882 | 882-1134 | 0.961 | 0.700 | 0.662 | 15.17% | - |
| 2 | 252-1008 | 1008-1260 | 0.734 | 1.596 | 3.793 | 4.63% | - |
| 3 | 378-1134 | 1134-1386 | 0.845 | 0.154 | 0.049 | 21.67% | - |
| 4 | 504-1260 | 1260-1512 | 0.463 | 0.334 | 0.223 | 21.67% | - |
| 5 | 630-1386 | 1386-1638 | 0.265 | 1.069 | 1.801 | 7.19% | - |
| 6 | 756-1512 | 1512-1764 | 0.524 | 0.829 | 1.440 | 7.19% | - |
| 7 | 882-1638 | 1638-1890 | 0.433 | -0.121 | -0.243 | 11.54% | - |
| 8 | 1008-1764 | 1764-2016 | 0.279 | -0.295 | -0.398 | 12.20% | - |
| 9 | 1134-1890 | 1890-2142 | 0.235 | 0.219 | 0.189 | 9.74% | - |
| 10 | 1260-2016 | 2016-2268 | 0.145 | 0.516 | 0.650 | 8.29% | - |
| 11 | 1386-2142 | 2142-2394 | -0.005 | 2.465 | 4.196 | 7.99% | - |
| 12 | 1512-2268 | 2268-2520 | 0.263 | 2.033 | 3.443 | 8.85% | - |
| 13 | 1638-2394 | 2394-2646 | 1.166 | 0.814 | 0.904 | 13.59% | - |
| 14 | 1764-2520 | 2520-2772 | 1.233 | 1.070 | 1.101 | 13.59% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | xsec_momentum_top25 | 1.802 | -0.023 | 1 | xsec_momentum_top25 | -0.023 | 3 | false |
| 1 | xsec_momentum_top10 | 1.134 | 0.449 | 3 | xsec_momentum_top25 | 1.101 | 3 | true |
| 2 | xsec_momentum_top25 | 0.670 | 5.276 | 1 | xsec_momentum_top25 | 5.276 | 3 | false |
| 3 | xsec_momentum_top25 | 0.791 | -0.030 | 3 | xsec_momentum_top10 | 0.178 | 3 | true |
| 4 | xsec_momentum_top25 | 0.472 | 0.164 | 3 | xsec_momentum_top10 | 0.318 | 3 | true |
| 5 | xsec_momentum_top25 | 0.150 | 2.252 | 1 | xsec_momentum_top25 | 2.252 | 3 | false |
| 6 | xsec_momentum_top10 | 0.326 | 0.993 | 3 | xsec_momentum_top25 | 1.869 | 3 | true |
| 7 | xsec_momentum_top25 | 0.245 | -0.064 | 1 | xsec_momentum_top25 | -0.064 | 3 | false |
| 8 | xsec_momentum_top25 | 0.163 | -0.298 | 1 | xsec_momentum_top25 | -0.298 | 3 | false |
| 9 | xsec_momentum_top25 | 0.342 | 0.528 | 1 | xsec_momentum_top25 | 0.528 | 3 | false |
| 10 | xsec_momentum_top25 | 0.192 | 0.463 | 3 | xsec_momentum_top15 | 0.650 | 3 | true |
| 11 | xsec_momentum_top25 | 0.057 | 2.725 | 3 | xsec_momentum_top10 | 5.874 | 3 | true |
| 12 | xsec_momentum_top15 | 0.201 | 3.443 | 2 | xsec_momentum_top10 | 4.655 | 3 | false |
| 13 | xsec_momentum_top10 | 1.152 | 1.014 | 1 | xsec_momentum_top10 | 1.014 | 3 | false |
| 14 | xsec_momentum_top10 | 1.878 | 1.080 | 2 | xsec_momentum_top15 | 1.101 | 3 | false |

## composite_momentum_checkpoint

- Family: `composite_momentum`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.533 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 407.29% | 15.34% | 0.897 | 0.843 | 0.449 | 34.17% | 658 | 98.828 | - |
| stress_2x | 402.97% | 15.25% | 0.893 | 0.840 | 0.446 | 34.17% | 658 | 98.308 | - |
| stress_3x | 398.68% | 15.16% | 0.888 | 0.836 | 0.444 | 34.17% | 658 | 97.792 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.025 | -0.456 | -0.458 | 19.38% | - |
| 1 | 126-882 | 882-1134 | 0.893 | 0.717 | 0.526 | 19.38% | - |
| 2 | 252-1008 | 1008-1260 | 0.646 | 2.142 | 3.783 | 7.54% | - |
| 3 | 378-1134 | 1134-1386 | 1.126 | 0.488 | 0.332 | 34.17% | - |
| 4 | 504-1260 | 1260-1512 | 1.016 | 0.675 | 0.555 | 34.17% | - |
| 5 | 630-1386 | 1386-1638 | 0.580 | 2.354 | 4.370 | 9.66% | - |
| 6 | 756-1512 | 1512-1764 | 0.624 | 2.052 | 5.739 | 5.39% | - |
| 7 | 882-1638 | 1638-1890 | 0.897 | -0.555 | -0.513 | 22.69% | - |
| 8 | 1008-1764 | 1764-2016 | 1.157 | -0.879 | -0.762 | 26.61% | - |
| 9 | 1134-1890 | 1890-2142 | 0.577 | 0.746 | 0.743 | 17.55% | - |
| 10 | 1260-2016 | 2016-2268 | 0.388 | 1.905 | 3.103 | 8.89% | - |
| 11 | 1386-2142 | 2142-2394 | 0.754 | 2.262 | 3.451 | 8.89% | - |
| 12 | 1512-2268 | 2268-2520 | 0.589 | 1.737 | 2.882 | 8.88% | - |
| 13 | 1638-2394 | 2394-2646 | 0.672 | 0.812 | 0.810 | 17.63% | - |
| 14 | 1764-2520 | 2520-2772 | 0.560 | 1.276 | 1.294 | 17.60% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | composite_momentum_broader_core | 1.174 | -0.382 | 2 | composite_momentum_strict_etf | -0.331 | 4 | false |
| 1 | composite_momentum_broader_core | 1.088 | 0.560 | 2 | composite_momentum_strict_etf | 0.622 | 4 | false |
| 2 | composite_momentum_strict_etf | 0.544 | 4.321 | 1 | composite_momentum_strict_etf | 4.321 | 4 | false |
| 3 | composite_momentum_strict_etf | 0.848 | 0.350 | 2 | composite_momentum_broader_core | 0.388 | 4 | false |
| 4 | composite_momentum_strict_etf | 0.815 | 0.598 | 2 | composite_momentum_broader_core | 0.640 | 4 | false |
| 5 | composite_momentum_strict_etf | 0.385 | 4.246 | 4 | composite_momentum_broader_core | 4.414 | 4 | true |
| 6 | composite_momentum_strict_etf | 0.426 | 5.547 | 4 | composite_momentum_sleeve20_broad5 | 5.851 | 4 | true |
| 7 | composite_momentum_broader_core | 0.606 | -0.504 | 3 | composite_momentum_strict_etf | -0.492 | 4 | true |
| 8 | composite_momentum_broader_core | 0.811 | -0.742 | 1 | composite_momentum_broader_core | -0.742 | 4 | false |
| 9 | composite_momentum_broader_core | 0.369 | 0.881 | 2 | composite_momentum_strict_etf | 1.017 | 4 | false |
| 10 | composite_momentum_strict_etf | 0.225 | 2.648 | 4 | composite_momentum_checkpoint | 3.103 | 4 | true |
| 11 | composite_momentum_strict_etf | 0.531 | 2.782 | 4 | composite_momentum_checkpoint | 3.451 | 4 | true |
| 12 | composite_momentum_strict_etf | 0.389 | 2.766 | 4 | composite_momentum_sleeve20_broad5 | 2.949 | 4 | true |
| 13 | composite_momentum_broader_core | 0.437 | 0.692 | 3 | composite_momentum_sleeve20_broad5 | 0.811 | 4 | true |
| 14 | composite_momentum_broader_core | 0.348 | 1.226 | 3 | composite_momentum_checkpoint | 1.294 | 4 | true |

## benchmark_tsmom_checkpoint

- Family: `benchmark_tsmom`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.600 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 411.35% | 15.42% | 0.879 | 0.831 | 0.458 | 33.64% | 266 | 33.993 | - |
| stress_2x | 409.89% | 15.39% | 0.878 | 0.830 | 0.457 | 33.64% | 266 | 33.933 | - |
| stress_3x | 408.45% | 15.36% | 0.877 | 0.828 | 0.457 | 33.65% | 266 | 33.874 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.082 | -0.314 | -0.363 | 20.16% | - |
| 1 | 126-882 | 882-1134 | 1.038 | 0.643 | 0.474 | 20.16% | - |
| 2 | 252-1008 | 1008-1260 | 0.729 | 2.176 | 4.307 | 6.66% | - |
| 3 | 378-1134 | 1134-1386 | 1.171 | 0.461 | 0.312 | 33.64% | - |
| 4 | 504-1260 | 1260-1512 | 1.087 | 0.646 | 0.541 | 33.64% | - |
| 5 | 630-1386 | 1386-1638 | 0.594 | 2.126 | 3.837 | 10.45% | - |
| 6 | 756-1512 | 1512-1764 | 0.616 | 1.946 | 5.221 | 5.84% | - |
| 7 | 882-1638 | 1638-1890 | 0.844 | -0.222 | -0.301 | 19.19% | - |
| 8 | 1008-1764 | 1764-2016 | 1.122 | -0.543 | -0.649 | 20.70% | - |
| 9 | 1134-1890 | 1890-2142 | 0.626 | 0.896 | 1.045 | 14.96% | - |
| 10 | 1260-2016 | 2016-2268 | 0.506 | 1.705 | 2.463 | 10.29% | - |
| 11 | 1386-2142 | 2142-2394 | 0.929 | 2.427 | 3.481 | 10.29% | - |
| 12 | 1512-2268 | 2268-2520 | 0.764 | 1.805 | 2.723 | 10.73% | - |
| 13 | 1638-2394 | 2394-2646 | 0.890 | 0.480 | 0.424 | 19.03% | - |
| 14 | 1764-2520 | 2520-2772 | 0.631 | 0.929 | 0.885 | 19.03% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_medium | 1.134 | -0.388 | 4 | benchmark_tsmom_reb42 | -0.310 | 4 | true |
| 1 | benchmark_tsmom_checkpoint | 1.163 | 0.474 | 3 | benchmark_tsmom_reb42 | 0.582 | 4 | true |
| 2 | benchmark_tsmom_reb42 | 0.513 | 4.731 | 2 | benchmark_tsmom_slow | 5.187 | 4 | false |
| 3 | benchmark_tsmom_reb42 | 0.838 | 0.278 | 4 | benchmark_tsmom_slow | 0.358 | 4 | true |
| 4 | benchmark_tsmom_reb42 | 0.827 | 0.555 | 1 | benchmark_tsmom_reb42 | 0.555 | 4 | false |
| 5 | benchmark_tsmom_slow | 0.372 | 3.698 | 4 | benchmark_tsmom_reb42 | 4.063 | 4 | true |
| 6 | benchmark_tsmom_reb42 | 0.405 | 5.039 | 4 | benchmark_tsmom_medium | 5.482 | 4 | true |
| 7 | benchmark_tsmom_medium | 0.581 | -0.290 | 1 | benchmark_tsmom_medium | -0.290 | 4 | false |
| 8 | benchmark_tsmom_medium | 0.799 | -0.636 | 1 | benchmark_tsmom_medium | -0.636 | 4 | false |
| 9 | benchmark_tsmom_medium | 0.403 | 1.152 | 1 | benchmark_tsmom_medium | 1.152 | 4 | false |
| 10 | benchmark_tsmom_medium | 0.297 | 2.536 | 2 | benchmark_tsmom_reb42 | 2.752 | 4 | false |
| 11 | benchmark_tsmom_medium | 0.807 | 3.386 | 4 | benchmark_tsmom_reb42 | 3.503 | 4 | true |
| 12 | benchmark_tsmom_medium | 0.632 | 2.584 | 3 | benchmark_tsmom_slow | 2.831 | 4 | true |
| 13 | benchmark_tsmom_medium | 0.742 | 0.358 | 4 | benchmark_tsmom_reb42 | 0.484 | 4 | true |
| 14 | benchmark_tsmom_medium | 0.434 | 0.849 | 3 | benchmark_tsmom_reb42 | 1.184 | 4 | true |

## benchmark_tsmom_blend

- Family: `benchmark_tsmom_blend`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.667 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 432.87% | 15.84% | 0.894 | 0.843 | 0.467 | 33.90% | 330 | 35.365 | - |
| stress_2x | 431.36% | 15.81% | 0.893 | 0.842 | 0.466 | 33.90% | 330 | 35.303 | - |
| stress_3x | 429.85% | 15.78% | 0.891 | 0.841 | 0.465 | 33.90% | 330 | 35.241 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.089 | -0.301 | -0.353 | 20.63% | - |
| 1 | 126-882 | 882-1134 | 1.051 | 0.626 | 0.456 | 20.63% | - |
| 2 | 252-1008 | 1008-1260 | 0.755 | 2.466 | 5.331 | 5.97% | - |
| 3 | 378-1134 | 1134-1386 | 1.185 | 0.542 | 0.403 | 33.90% | - |
| 4 | 504-1260 | 1260-1512 | 1.147 | 0.659 | 0.555 | 33.90% | - |
| 5 | 630-1386 | 1386-1638 | 0.634 | 2.048 | 3.687 | 10.61% | - |
| 6 | 756-1512 | 1512-1764 | 0.652 | 1.867 | 4.926 | 6.06% | - |
| 7 | 882-1638 | 1638-1890 | 0.881 | -0.189 | -0.277 | 18.66% | - |
| 8 | 1008-1764 | 1764-2016 | 1.137 | -0.502 | -0.624 | 20.05% | - |
| 9 | 1134-1890 | 1890-2142 | 0.619 | 0.961 | 1.154 | 14.62% | - |
| 10 | 1260-2016 | 2016-2268 | 0.505 | 1.731 | 2.498 | 10.45% | - |
| 11 | 1386-2142 | 2142-2394 | 0.950 | 2.427 | 3.540 | 10.45% | - |
| 12 | 1512-2268 | 2268-2520 | 0.797 | 1.759 | 2.630 | 11.23% | - |
| 13 | 1638-2394 | 2394-2646 | 0.909 | 0.420 | 0.352 | 19.38% | - |
| 14 | 1764-2520 | 2520-2772 | 0.654 | 0.838 | 0.768 | 19.38% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_blend_etf_tilt | 1.077 | -0.372 | 4 | benchmark_tsmom_blend_reb42 | -0.311 | 4 | true |
| 1 | benchmark_tsmom_blend | 1.181 | 0.456 | 3 | benchmark_tsmom_blend_reb42 | 0.582 | 4 | true |
| 2 | benchmark_tsmom_blend_reb42 | 0.530 | 4.907 | 4 | benchmark_tsmom_blend | 5.331 | 4 | true |
| 3 | benchmark_tsmom_blend_reb42 | 0.849 | 0.294 | 4 | benchmark_tsmom_blend | 0.403 | 4 | true |
| 4 | benchmark_tsmom_blend_reb42 | 0.826 | 0.561 | 1 | benchmark_tsmom_blend_reb42 | 0.561 | 4 | false |
| 5 | benchmark_tsmom_blend | 0.380 | 3.687 | 4 | benchmark_tsmom_blend_reb42 | 3.991 | 4 | true |
| 6 | benchmark_tsmom_blend_reb42 | 0.412 | 5.084 | 3 | benchmark_tsmom_blend_etf_tilt | 5.253 | 4 | true |
| 7 | benchmark_tsmom_blend_etf_tilt | 0.592 | -0.280 | 2 | benchmark_tsmom_blend | -0.277 | 4 | false |
| 8 | benchmark_tsmom_blend_etf_tilt | 0.799 | -0.637 | 2 | benchmark_tsmom_blend | -0.624 | 4 | false |
| 9 | benchmark_tsmom_blend_etf_tilt | 0.388 | 1.105 | 3 | benchmark_tsmom_blend | 1.154 | 4 | true |
| 10 | benchmark_tsmom_blend | 0.293 | 2.498 | 4 | benchmark_tsmom_blend_reb42 | 2.561 | 4 | true |
| 11 | benchmark_tsmom_blend | 0.806 | 3.540 | 1 | benchmark_tsmom_blend | 3.540 | 4 | false |
| 12 | benchmark_tsmom_blend | 0.633 | 2.630 | 2 | benchmark_tsmom_blend_slow_broad | 2.775 | 4 | false |
| 13 | benchmark_tsmom_blend | 0.680 | 0.352 | 4 | benchmark_tsmom_blend_reb42 | 0.441 | 4 | true |
| 14 | benchmark_tsmom_blend | 0.444 | 0.768 | 4 | benchmark_tsmom_blend_reb42 | 1.121 | 4 | true |

## benchmark_lowvol_checkpoint

- Family: `benchmark_lowvol`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.533 above 0.200
  - turnover increases without return improvement

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 331.91% | 13.72% | 0.862 | 0.817 | 0.420 | 32.68% | 653 | 89.580 | - |
| stress_2x | 328.27% | 13.63% | 0.858 | 0.812 | 0.417 | 32.68% | 654 | 89.116 | - |
| stress_3x | 324.67% | 13.55% | 0.853 | 0.808 | 0.414 | 32.69% | 654 | 88.656 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.034 | -0.333 | -0.366 | 17.88% | - |
| 1 | 126-882 | 882-1134 | 0.901 | 1.004 | 0.807 | 17.88% | - |
| 2 | 252-1008 | 1008-1260 | 0.778 | 2.387 | 5.125 | 5.75% | - |
| 3 | 378-1134 | 1134-1386 | 1.286 | 0.464 | 0.315 | 32.68% | - |
| 4 | 504-1260 | 1260-1512 | 1.174 | 0.657 | 0.543 | 32.68% | - |
| 5 | 630-1386 | 1386-1638 | 0.643 | 2.391 | 4.402 | 8.42% | - |
| 6 | 756-1512 | 1512-1764 | 0.697 | 2.264 | 6.190 | 4.79% | - |
| 7 | 882-1638 | 1638-1890 | 0.897 | -0.262 | -0.300 | 19.78% | - |
| 8 | 1008-1764 | 1764-2016 | 1.195 | -0.617 | -0.662 | 21.76% | - |
| 9 | 1134-1890 | 1890-2142 | 0.612 | 0.756 | 0.794 | 15.62% | - |
| 10 | 1260-2016 | 2016-2268 | 0.460 | 1.533 | 1.889 | 10.11% | - |
| 11 | 1386-2142 | 2142-2394 | 0.844 | 2.061 | 2.247 | 10.11% | - |
| 12 | 1512-2268 | 2268-2520 | 0.672 | 2.049 | 4.363 | 5.61% | - |
| 13 | 1638-2394 | 2394-2646 | 0.680 | 0.709 | 0.611 | 18.83% | - |
| 14 | 1764-2520 | 2520-2772 | 0.554 | 1.007 | 0.876 | 18.83% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lowvol_voladj | 1.229 | -0.320 | 1 | benchmark_lowvol_voladj | -0.320 | 4 | false |
| 1 | benchmark_lowvol_voladj | 1.179 | 0.697 | 4 | benchmark_lowvol_reb42 | 0.885 | 4 | true |
| 2 | benchmark_lowvol_voladj | 0.633 | 5.234 | 2 | benchmark_lowvol_reb42 | 5.341 | 4 | false |
| 3 | benchmark_lowvol_wider | 0.914 | 0.309 | 2 | benchmark_lowvol_checkpoint | 0.315 | 4 | false |
| 4 | benchmark_lowvol_voladj | 0.875 | 0.444 | 4 | benchmark_lowvol_checkpoint | 0.543 | 4 | true |
| 5 | benchmark_lowvol_wider | 0.382 | 4.275 | 4 | benchmark_lowvol_voladj | 4.499 | 4 | true |
| 6 | benchmark_lowvol_checkpoint | 0.429 | 6.190 | 1 | benchmark_lowvol_checkpoint | 6.190 | 4 | false |
| 7 | benchmark_lowvol_checkpoint | 0.580 | -0.300 | 2 | benchmark_lowvol_wider | -0.296 | 4 | false |
| 8 | benchmark_lowvol_checkpoint | 0.802 | -0.662 | 2 | benchmark_lowvol_wider | -0.620 | 4 | false |
| 9 | benchmark_lowvol_checkpoint | 0.368 | 0.794 | 2 | benchmark_lowvol_wider | 0.811 | 4 | false |
| 10 | benchmark_lowvol_checkpoint | 0.251 | 1.889 | 3 | benchmark_lowvol_reb42 | 1.969 | 4 | true |
| 11 | benchmark_lowvol_wider | 0.621 | 2.259 | 3 | benchmark_lowvol_reb42 | 2.492 | 4 | true |
| 12 | benchmark_lowvol_wider | 0.473 | 3.887 | 3 | benchmark_lowvol_checkpoint | 4.363 | 4 | true |
| 13 | benchmark_lowvol_wider | 0.462 | 0.512 | 4 | benchmark_lowvol_reb42 | 0.708 | 4 | true |
| 14 | benchmark_lowvol_checkpoint | 0.341 | 0.876 | 3 | benchmark_lowvol_voladj | 1.020 | 4 | true |

## benchmark_reversal_checkpoint

- Family: `benchmark_reversal`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.533 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 387.62% | 14.94% | 0.865 | 0.816 | 0.424 | 35.24% | 4357 | 249.915 | - |
| stress_2x | 376.79% | 14.71% | 0.854 | 0.806 | 0.417 | 35.26% | 4363 | 246.430 | - |
| stress_3x | 366.20% | 14.48% | 0.843 | 0.795 | 0.411 | 35.27% | 4367 | 243.004 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.022 | -0.264 | -0.322 | 18.32% | - |
| 1 | 126-882 | 882-1134 | 1.003 | 0.852 | 0.656 | 19.84% | - |
| 2 | 252-1008 | 1008-1260 | 0.820 | 2.307 | 4.289 | 7.33% | - |
| 3 | 378-1134 | 1134-1386 | 1.281 | 0.390 | 0.223 | 34.65% | - |
| 4 | 504-1260 | 1260-1512 | 1.235 | 0.617 | 0.472 | 35.01% | - |
| 5 | 630-1386 | 1386-1638 | 0.639 | 2.362 | 4.404 | 9.33% | - |
| 6 | 756-1512 | 1512-1764 | 0.673 | 2.131 | 5.927 | 5.37% | - |
| 7 | 882-1638 | 1638-1890 | 0.877 | -0.321 | -0.383 | 21.54% | - |
| 8 | 1008-1764 | 1764-2016 | 1.177 | -0.661 | -0.734 | 24.13% | - |
| 9 | 1134-1890 | 1890-2142 | 0.585 | 0.692 | 0.704 | 17.59% | - |
| 10 | 1260-2016 | 2016-2268 | 0.409 | 1.813 | 2.713 | 9.34% | - |
| 11 | 1386-2142 | 2142-2394 | 0.804 | 2.332 | 2.913 | 9.75% | - |
| 12 | 1512-2268 | 2268-2520 | 0.702 | 1.807 | 2.615 | 9.43% | - |
| 13 | 1638-2394 | 2394-2646 | 0.689 | 0.755 | 0.802 | 17.44% | - |
| 14 | 1764-2520 | 2520-2772 | 0.547 | 1.248 | 1.377 | 17.55% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_reversal_slow | 1.096 | -0.285 | 2 | benchmark_reversal_medium | -0.277 | 4 | false |
| 1 | benchmark_reversal_slow | 1.143 | 0.644 | 4 | benchmark_reversal_medium | 0.694 | 4 | true |
| 2 | benchmark_reversal_slow | 0.606 | 4.459 | 1 | benchmark_reversal_slow | 4.459 | 4 | false |
| 3 | benchmark_reversal_slow | 0.887 | 0.260 | 2 | benchmark_reversal_fast | 0.263 | 4 | false |
| 4 | benchmark_reversal_slow | 0.881 | 0.452 | 3 | benchmark_reversal_fast | 0.529 | 4 | true |
| 5 | benchmark_reversal_fast | 0.372 | 4.288 | 4 | benchmark_reversal_medium | 4.648 | 4 | true |
| 6 | benchmark_reversal_medium | 0.443 | 6.032 | 3 | benchmark_reversal_slow | 6.810 | 4 | true |
| 7 | benchmark_reversal_slow | 0.564 | -0.406 | 2 | benchmark_reversal_checkpoint | -0.383 | 4 | false |
| 8 | benchmark_reversal_checkpoint | 0.781 | -0.734 | 3 | benchmark_reversal_medium | -0.700 | 4 | true |
| 9 | benchmark_reversal_fast | 0.357 | 0.794 | 3 | benchmark_reversal_slow | 1.014 | 4 | true |
| 10 | benchmark_reversal_slow | 0.257 | 3.192 | 2 | benchmark_reversal_medium | 3.218 | 4 | false |
| 11 | benchmark_reversal_slow | 0.677 | 3.008 | 2 | benchmark_reversal_medium | 3.319 | 4 | false |
| 12 | benchmark_reversal_slow | 0.540 | 3.642 | 1 | benchmark_reversal_slow | 3.642 | 4 | false |
| 13 | benchmark_reversal_slow | 0.517 | 0.765 | 3 | benchmark_reversal_medium | 0.859 | 4 | true |
| 14 | benchmark_reversal_slow | 0.473 | 1.001 | 3 | benchmark_reversal_checkpoint | 1.377 | 4 | true |
