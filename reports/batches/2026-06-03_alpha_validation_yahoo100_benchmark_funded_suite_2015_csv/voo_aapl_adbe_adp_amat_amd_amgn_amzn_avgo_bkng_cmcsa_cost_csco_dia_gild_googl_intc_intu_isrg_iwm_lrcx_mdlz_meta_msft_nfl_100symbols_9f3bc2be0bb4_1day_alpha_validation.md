# Alpha Validation Report

- Generated: `2026-06-03T11:26:32Z`
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
| benchmark_rotation_defensive | buy_hold | false | 283.63% | 12.54% | 0.789 | 0.735 | 0.368 | 34.10% | 1.000 | 0.400 | 587 | PBO 0.400 above 0.200 |
| benchmark_tsmom_checkpoint | buy_hold | false | 411.35% | 15.42% | 0.879 | 0.831 | 0.458 | 33.64% | 1.000 | 0.600 | 266 | PBO 0.600 above 0.200 |
| benchmark_tsmom_blend | buy_hold | false | 432.87% | 15.84% | 0.894 | 0.843 | 0.467 | 33.90% | 1.000 | 0.667 | 330 | PBO 0.667 above 0.200 |
| benchmark_lowvol_checkpoint | buy_hold | false | 331.91% | 13.72% | 0.862 | 0.817 | 0.420 | 32.68% | 1.000 | 0.533 | 653 | PBO 0.533 above 0.200 |
| benchmark_reversal_checkpoint | buy_hold | false | 387.62% | 14.94% | 0.865 | 0.816 | 0.424 | 35.24% | 1.000 | 0.533 | 4357 | PBO 0.533 above 0.200 |
| benchmark_ranked_sleeve_checkpoint | buy_hold | false | 474.41% | 16.60% | 0.944 | 0.888 | 0.482 | 34.45% | 1.000 | 0.400 | 561 | PBO 0.400 above 0.200 |
| benchmark_ranker_proxy_h63_checkpoint | buy_hold | true | 488.98% | 16.86% | 0.964 | 0.901 | 0.473 | 35.67% | 1.000 | 0.200 | 142 | pass |
| benchmark_ranker_proxy_h84_checkpoint | buy_hold | false | 461.86% | 16.38% | 0.945 | 0.879 | 0.461 | 35.51% | 1.000 | 0.267 | 147 | PBO 0.267 above 0.200 |
| sector_ranked_sleeve_checkpoint | buy_hold | false | 372.89% | 14.63% | 0.861 | 0.814 | 0.431 | 33.92% | 1.000 | 0.600 | 312 | PBO 0.600 above 0.200 |

## benchmark_rotation_defensive

- Family: `composite_momentum_defensive`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.400 above 0.200
  - turnover increases without return improvement
  - no drawdown-adjusted improvement over benchmark

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 283.63% | 12.54% | 0.789 | 0.735 | 0.368 | 34.10% | 587 | 72.630 | - |
| stress_2x | 280.67% | 12.46% | 0.785 | 0.731 | 0.365 | 34.10% | 587 | 72.302 | - |
| stress_3x | 277.73% | 12.39% | 0.781 | 0.727 | 0.363 | 34.10% | 587 | 71.976 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.010 | -0.381 | -0.405 | 18.45% | - |
| 1 | 126-882 | 882-1134 | 0.926 | 0.710 | 0.529 | 18.45% | - |
| 2 | 252-1008 | 1008-1260 | 0.740 | 2.078 | 3.937 | 6.85% | - |
| 3 | 378-1134 | 1134-1386 | 1.189 | 0.189 | 0.025 | 34.10% | - |
| 4 | 504-1260 | 1260-1512 | 1.114 | 0.360 | 0.194 | 34.10% | - |
| 5 | 630-1386 | 1386-1638 | 0.459 | 2.244 | 4.061 | 9.58% | - |
| 6 | 756-1512 | 1512-1764 | 0.487 | 2.043 | 5.664 | 5.27% | - |
| 7 | 882-1638 | 1638-1890 | 0.749 | -0.352 | -0.349 | 22.63% | - |
| 8 | 1008-1764 | 1764-2016 | 1.028 | -0.851 | -0.718 | 24.20% | - |
| 9 | 1134-1890 | 1890-2142 | 0.527 | 0.618 | 0.573 | 14.73% | - |
| 10 | 1260-2016 | 2016-2268 | 0.422 | 1.536 | 1.956 | 9.97% | - |
| 11 | 1386-2142 | 2142-2394 | 0.750 | 2.218 | 2.896 | 9.97% | - |
| 12 | 1512-2268 | 2268-2520 | 0.592 | 1.805 | 2.965 | 8.70% | - |
| 13 | 1638-2394 | 2394-2646 | 0.674 | 0.360 | 0.276 | 18.00% | - |
| 14 | 1764-2520 | 2520-2772 | 0.482 | 0.772 | 0.668 | 18.00% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_rotation_cash_defensive | 1.019 | -0.427 | 3 | benchmark_rotation_defensive | -0.405 | 4 | true |
| 1 | benchmark_rotation_defensive | 0.999 | 0.529 | 2 | benchmark_rotation_half_defensive | 0.542 | 4 | false |
| 2 | benchmark_rotation_defensive | 0.492 | 3.937 | 3 | benchmark_rotation_half_defensive | 4.152 | 4 | true |
| 3 | benchmark_rotation_defensive | 0.793 | 0.025 | 3 | benchmark_rotation_half_defensive | 0.128 | 4 | true |
| 4 | benchmark_rotation_defensive | 0.766 | 0.194 | 3 | benchmark_rotation_half_defensive | 0.298 | 4 | true |
| 5 | benchmark_rotation_half_defensive | 0.270 | 4.044 | 2 | benchmark_rotation_defensive | 4.061 | 4 | false |
| 6 | benchmark_rotation_half_defensive | 0.297 | 5.721 | 1 | benchmark_rotation_half_defensive | 5.721 | 4 | false |
| 7 | benchmark_rotation_half_defensive | 0.487 | -0.392 | 4 | benchmark_rotation_defensive | -0.349 | 4 | true |
| 8 | benchmark_rotation_half_defensive | 0.688 | -0.696 | 1 | benchmark_rotation_half_defensive | -0.696 | 4 | false |
| 9 | benchmark_rotation_half_defensive | 0.306 | 0.708 | 1 | benchmark_rotation_half_defensive | 0.708 | 4 | false |
| 10 | benchmark_rotation_defensive | 0.223 | 1.956 | 2 | benchmark_rotation_half_defensive | 2.148 | 4 | false |
| 11 | benchmark_rotation_half_defensive | 0.487 | 2.880 | 4 | benchmark_rotation_cash_defensive | 2.925 | 4 | true |
| 12 | benchmark_rotation_half_defensive | 0.360 | 2.997 | 1 | benchmark_rotation_half_defensive | 2.997 | 4 | false |
| 13 | benchmark_rotation_half_defensive | 0.408 | 0.456 | 1 | benchmark_rotation_half_defensive | 0.456 | 4 | false |
| 14 | benchmark_rotation_half_defensive | 0.274 | 0.841 | 1 | benchmark_rotation_half_defensive | 0.841 | 4 | false |

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

## benchmark_ranker_proxy_h63_checkpoint

- Family: `benchmark_ranker_proxy_h63`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `true`

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 488.98% | 16.86% | 0.964 | 0.901 | 0.473 | 35.67% | 142 | 31.400 | - |
| stress_2x | 487.53% | 16.83% | 0.963 | 0.900 | 0.472 | 35.67% | 142 | 31.351 | - |
| stress_3x | 486.08% | 16.81% | 0.962 | 0.899 | 0.471 | 35.67% | 142 | 31.302 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.294 | -0.250 | -0.307 | 18.12% | - |
| 1 | 126-882 | 882-1134 | 1.148 | 0.951 | 0.774 | 18.12% | - |
| 2 | 252-1008 | 1008-1260 | 0.961 | 2.295 | 4.713 | 6.42% | - |
| 3 | 378-1134 | 1134-1386 | 1.366 | 0.493 | 0.325 | 35.67% | - |
| 4 | 504-1260 | 1260-1512 | 1.254 | 0.688 | 0.551 | 35.67% | - |
| 5 | 630-1386 | 1386-1638 | 0.683 | 2.449 | 4.698 | 9.48% | - |
| 6 | 756-1512 | 1512-1764 | 0.711 | 2.361 | 6.339 | 5.80% | - |
| 7 | 882-1638 | 1638-1890 | 0.944 | -0.231 | -0.284 | 21.51% | - |
| 8 | 1008-1764 | 1764-2016 | 1.238 | -0.700 | -0.715 | 23.76% | - |
| 9 | 1134-1890 | 1890-2142 | 0.684 | 0.904 | 1.021 | 16.01% | - |
| 10 | 1260-2016 | 2016-2268 | 0.509 | 1.919 | 2.925 | 9.68% | - |
| 11 | 1386-2142 | 2142-2394 | 0.942 | 2.344 | 3.205 | 9.68% | - |
| 12 | 1512-2268 | 2268-2520 | 0.742 | 1.888 | 2.984 | 8.91% | - |
| 13 | 1638-2394 | 2394-2646 | 0.836 | 0.560 | 0.500 | 18.91% | - |
| 14 | 1764-2520 | 2520-2772 | 0.666 | 1.070 | 1.047 | 18.91% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranker_proxy_h63_checkpoint | 1.319 | -0.307 | 2 | benchmark_ranker_proxy_h126 | -0.282 | 5 | false |
| 1 | benchmark_ranker_proxy_h63_checkpoint | 1.211 | 0.774 | 2 | benchmark_ranker_proxy_h126 | 0.792 | 5 | false |
| 2 | benchmark_ranker_proxy_h63_checkpoint | 0.686 | 4.713 | 1 | benchmark_ranker_proxy_h63_checkpoint | 4.713 | 5 | false |
| 3 | benchmark_ranker_proxy_h63_checkpoint | 0.966 | 0.325 | 3 | benchmark_ranker_proxy_h84 | 0.339 | 5 | false |
| 4 | benchmark_ranker_proxy_h63_checkpoint | 0.902 | 0.551 | 2 | benchmark_ranker_proxy_h63_strict | 0.552 | 5 | false |
| 5 | benchmark_ranker_proxy_h63_checkpoint | 0.386 | 4.698 | 2 | benchmark_ranker_proxy_h63_strict | 4.703 | 5 | false |
| 6 | benchmark_ranker_proxy_h63_checkpoint | 0.420 | 6.339 | 1 | benchmark_ranker_proxy_h63_checkpoint | 6.339 | 5 | false |
| 7 | benchmark_ranker_proxy_h63_checkpoint | 0.596 | -0.284 | 1 | benchmark_ranker_proxy_h63_checkpoint | -0.284 | 5 | false |
| 8 | benchmark_ranker_proxy_h63_checkpoint | 0.818 | -0.715 | 2 | benchmark_ranker_proxy_h84 | -0.687 | 5 | false |
| 9 | benchmark_ranker_proxy_h63_checkpoint | 0.411 | 1.021 | 4 | benchmark_ranker_proxy_h84 | 1.145 | 5 | true |
| 10 | benchmark_ranker_proxy_h84 | 0.304 | 3.441 | 1 | benchmark_ranker_proxy_h84 | 3.441 | 5 | false |
| 11 | benchmark_ranker_proxy_h84 | 0.794 | 3.888 | 1 | benchmark_ranker_proxy_h84 | 3.888 | 5 | false |
| 12 | benchmark_ranker_proxy_h84 | 0.614 | 3.038 | 1 | benchmark_ranker_proxy_h84 | 3.038 | 5 | false |
| 13 | benchmark_ranker_proxy_h84 | 0.708 | 0.385 | 5 | benchmark_ranker_proxy_h126 | 0.634 | 5 | true |
| 14 | benchmark_ranker_proxy_h84 | 0.442 | 0.914 | 5 | benchmark_ranker_proxy_h126 | 1.163 | 5 | true |

## benchmark_ranker_proxy_h84_checkpoint

- Family: `benchmark_ranker_proxy_h84`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.267 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 461.86% | 16.38% | 0.945 | 0.879 | 0.461 | 35.51% | 147 | 29.623 | - |
| stress_2x | 460.50% | 16.35% | 0.944 | 0.878 | 0.460 | 35.52% | 147 | 29.578 | - |
| stress_3x | 459.15% | 16.33% | 0.943 | 0.877 | 0.460 | 35.52% | 147 | 29.533 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.173 | -0.289 | -0.346 | 18.09% | - |
| 1 | 126-882 | 882-1134 | 1.017 | 0.908 | 0.732 | 18.09% | - |
| 2 | 252-1008 | 1008-1260 | 0.867 | 2.282 | 4.496 | 6.67% | - |
| 3 | 378-1134 | 1134-1386 | 1.228 | 0.507 | 0.339 | 35.51% | - |
| 4 | 504-1260 | 1260-1512 | 1.202 | 0.677 | 0.539 | 35.51% | - |
| 5 | 630-1386 | 1386-1638 | 0.653 | 2.365 | 4.293 | 10.15% | - |
| 6 | 756-1512 | 1512-1764 | 0.700 | 2.210 | 6.001 | 5.77% | - |
| 7 | 882-1638 | 1638-1890 | 0.916 | -0.227 | -0.292 | 20.02% | - |
| 8 | 1008-1764 | 1764-2016 | 1.213 | -0.592 | -0.687 | 21.04% | - |
| 9 | 1134-1890 | 1890-2142 | 0.682 | 0.963 | 1.145 | 15.34% | - |
| 10 | 1260-2016 | 2016-2268 | 0.518 | 2.029 | 3.441 | 8.71% | - |
| 11 | 1386-2142 | 2142-2394 | 0.983 | 2.569 | 3.888 | 8.71% | - |
| 12 | 1512-2268 | 2268-2520 | 0.817 | 1.944 | 3.038 | 9.01% | - |
| 13 | 1638-2394 | 2394-2646 | 0.925 | 0.465 | 0.385 | 19.31% | - |
| 14 | 1764-2520 | 2520-2772 | 0.678 | 0.977 | 0.914 | 19.31% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranker_proxy_h63 | 1.319 | -0.307 | 2 | benchmark_ranker_proxy_h126 | -0.282 | 5 | false |
| 1 | benchmark_ranker_proxy_h63 | 1.211 | 0.774 | 2 | benchmark_ranker_proxy_h126 | 0.792 | 5 | false |
| 2 | benchmark_ranker_proxy_h63 | 0.686 | 4.713 | 1 | benchmark_ranker_proxy_h63 | 4.713 | 5 | false |
| 3 | benchmark_ranker_proxy_h63 | 0.966 | 0.325 | 3 | benchmark_ranker_proxy_h84_checkpoint | 0.339 | 5 | false |
| 4 | benchmark_ranker_proxy_h63 | 0.902 | 0.551 | 1 | benchmark_ranker_proxy_h63 | 0.551 | 5 | false |
| 5 | benchmark_ranker_proxy_h63 | 0.386 | 4.698 | 1 | benchmark_ranker_proxy_h63 | 4.698 | 5 | false |
| 6 | benchmark_ranker_proxy_h63 | 0.420 | 6.339 | 1 | benchmark_ranker_proxy_h63 | 6.339 | 5 | false |
| 7 | benchmark_ranker_proxy_h63 | 0.596 | -0.284 | 1 | benchmark_ranker_proxy_h63 | -0.284 | 5 | false |
| 8 | benchmark_ranker_proxy_h63 | 0.818 | -0.715 | 4 | benchmark_ranker_proxy_h84_checkpoint | -0.687 | 5 | true |
| 9 | benchmark_ranker_proxy_h63 | 0.411 | 1.021 | 5 | benchmark_ranker_proxy_h84_strict | 1.145 | 5 | true |
| 10 | benchmark_ranker_proxy_h84_checkpoint | 0.304 | 3.441 | 1 | benchmark_ranker_proxy_h84_checkpoint | 3.441 | 5 | false |
| 11 | benchmark_ranker_proxy_h84_checkpoint | 0.794 | 3.888 | 1 | benchmark_ranker_proxy_h84_checkpoint | 3.888 | 5 | false |
| 12 | benchmark_ranker_proxy_h84_checkpoint | 0.614 | 3.038 | 3 | benchmark_ranker_proxy_h84_strict | 3.075 | 5 | false |
| 13 | benchmark_ranker_proxy_h84_checkpoint | 0.708 | 0.385 | 5 | benchmark_ranker_proxy_h126 | 0.634 | 5 | true |
| 14 | benchmark_ranker_proxy_h84_strict | 0.444 | 0.918 | 4 | benchmark_ranker_proxy_h126 | 1.163 | 5 | true |

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
