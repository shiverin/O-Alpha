# Alpha Validation Report

- Generated: `2026-06-03T12:00:10Z`
- Symbols: `VOO, DIA, IWM, QQQ, SMH, SPY, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY`
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
| equal_weight | 441.81% | 16.01% | 0.886 | 0.846 | 0.466 | 34.35% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| composite_momentum_checkpoint | buy_hold | false | 346.95% | 14.06% | 0.840 | 0.789 | 0.414 | 33.98% | 1.000 | 0.467 | 225 | PBO 0.467 above 0.200 |
| benchmark_rotation_defensive | buy_hold | false | 260.21% | 11.92% | 0.759 | 0.707 | 0.351 | 33.99% | 1.000 | 0.267 | 205 | PBO 0.267 above 0.200 |
| benchmark_tsmom_checkpoint | buy_hold | false | 379.30% | 14.76% | 0.859 | 0.811 | 0.441 | 33.46% | 1.000 | 0.600 | 127 | PBO 0.600 above 0.200 |
| benchmark_tsmom_blend | buy_hold | false | 382.54% | 14.83% | 0.861 | 0.814 | 0.444 | 33.44% | 1.000 | 0.600 | 140 | PBO 0.600 above 0.200 |
| sector_ranked_sleeve_checkpoint | buy_hold | false | 372.89% | 14.63% | 0.861 | 0.814 | 0.431 | 33.92% | 1.000 | 0.600 | 312 | PBO 0.600 above 0.200 |

## composite_momentum_checkpoint

- Family: `composite_momentum`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.467 above 0.200
  - turnover increases without return improvement
  - no drawdown-adjusted improvement over benchmark

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 346.95% | 14.06% | 0.840 | 0.789 | 0.414 | 33.98% | 225 | 82.623 | - |
| stress_2x | 343.48% | 13.98% | 0.836 | 0.785 | 0.411 | 33.98% | 225 | 82.234 | - |
| stress_3x | 340.04% | 13.90% | 0.832 | 0.782 | 0.409 | 33.98% | 226 | 81.846 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.920 | -0.560 | -0.533 | 19.30% | - |
| 1 | 126-882 | 882-1134 | 0.789 | 0.698 | 0.505 | 19.30% | - |
| 2 | 252-1008 | 1008-1260 | 0.548 | 2.178 | 4.012 | 7.17% | - |
| 3 | 378-1134 | 1134-1386 | 1.081 | 0.415 | 0.254 | 33.98% | - |
| 4 | 504-1260 | 1260-1512 | 0.968 | 0.584 | 0.449 | 33.98% | - |
| 5 | 630-1386 | 1386-1638 | 0.526 | 2.321 | 4.283 | 9.50% | - |
| 6 | 756-1512 | 1512-1764 | 0.574 | 2.067 | 5.974 | 5.09% | - |
| 7 | 882-1638 | 1638-1890 | 0.853 | -0.559 | -0.508 | 22.99% | - |
| 8 | 1008-1764 | 1764-2016 | 1.109 | -0.902 | -0.780 | 26.79% | - |
| 9 | 1134-1890 | 1890-2142 | 0.531 | 0.679 | 0.653 | 17.91% | - |
| 10 | 1260-2016 | 2016-2268 | 0.366 | 1.874 | 2.998 | 8.92% | - |
| 11 | 1386-2142 | 2142-2394 | 0.725 | 2.291 | 3.381 | 8.92% | - |
| 12 | 1512-2268 | 2268-2520 | 0.561 | 1.803 | 3.003 | 8.58% | - |
| 13 | 1638-2394 | 2394-2646 | 0.637 | 0.869 | 0.896 | 17.22% | - |
| 14 | 1764-2520 | 2520-2772 | 0.539 | 1.262 | 1.299 | 17.22% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | composite_momentum_strict_etf | 0.966 | -0.408 | 1 | composite_momentum_strict_etf | -0.408 | 4 | false |
| 1 | composite_momentum_strict_etf | 0.942 | 0.601 | 1 | composite_momentum_strict_etf | 0.601 | 4 | false |
| 2 | composite_momentum_strict_etf | 0.467 | 4.602 | 1 | composite_momentum_strict_etf | 4.602 | 4 | false |
| 3 | composite_momentum_strict_etf | 0.805 | 0.272 | 1 | composite_momentum_strict_etf | 0.272 | 4 | false |
| 4 | composite_momentum_strict_etf | 0.772 | 0.491 | 1 | composite_momentum_strict_etf | 0.491 | 4 | false |
| 5 | composite_momentum_strict_etf | 0.345 | 4.158 | 4 | composite_momentum_sleeve20_broad5 | 4.285 | 4 | true |
| 6 | composite_momentum_strict_etf | 0.387 | 5.771 | 4 | composite_momentum_sleeve20_broad5 | 5.995 | 4 | true |
| 7 | composite_momentum_strict_etf | 0.550 | -0.488 | 1 | composite_momentum_strict_etf | -0.488 | 4 | false |
| 8 | composite_momentum_strict_etf | 0.755 | -0.762 | 1 | composite_momentum_strict_etf | -0.762 | 4 | false |
| 9 | composite_momentum_broader_core | 0.312 | 0.728 | 2 | composite_momentum_strict_etf | 0.910 | 4 | false |
| 10 | composite_momentum_strict_etf | 0.207 | 2.556 | 4 | composite_momentum_checkpoint | 2.998 | 4 | true |
| 11 | composite_momentum_strict_etf | 0.507 | 2.721 | 4 | composite_momentum_checkpoint | 3.381 | 4 | true |
| 12 | composite_momentum_strict_etf | 0.367 | 2.882 | 4 | composite_momentum_sleeve20_broad5 | 3.023 | 4 | true |
| 13 | composite_momentum_strict_etf | 0.399 | 0.672 | 4 | composite_momentum_checkpoint | 0.896 | 4 | true |
| 14 | composite_momentum_strict_etf | 0.325 | 1.036 | 4 | composite_momentum_checkpoint | 1.299 | 4 | true |

## benchmark_rotation_defensive

- Family: `composite_momentum_defensive`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.267 above 0.200
  - turnover increases without return improvement
  - no drawdown-adjusted improvement over benchmark

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 260.21% | 11.92% | 0.759 | 0.707 | 0.351 | 33.99% | 205 | 66.033 | - |
| stress_2x | 257.58% | 11.85% | 0.756 | 0.703 | 0.349 | 33.99% | 205 | 65.754 | - |
| stress_3x | 254.96% | 11.77% | 0.752 | 0.700 | 0.346 | 33.99% | 207 | 65.475 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.959 | -0.432 | -0.447 | 18.23% | - |
| 1 | 126-882 | 882-1134 | 0.872 | 0.714 | 0.534 | 18.23% | - |
| 2 | 252-1008 | 1008-1260 | 0.686 | 2.099 | 4.096 | 6.63% | - |
| 3 | 378-1134 | 1134-1386 | 1.167 | 0.151 | -0.011 | 33.99% | - |
| 4 | 504-1260 | 1260-1512 | 1.091 | 0.310 | 0.144 | 33.99% | - |
| 5 | 630-1386 | 1386-1638 | 0.433 | 2.222 | 4.007 | 9.48% | - |
| 6 | 756-1512 | 1512-1764 | 0.462 | 2.050 | 5.803 | 5.09% | - |
| 7 | 882-1638 | 1638-1890 | 0.725 | -0.359 | -0.353 | 22.67% | - |
| 8 | 1008-1764 | 1764-2016 | 1.001 | -0.855 | -0.720 | 24.31% | - |
| 9 | 1134-1890 | 1890-2142 | 0.497 | 0.598 | 0.549 | 14.73% | - |
| 10 | 1260-2016 | 2016-2268 | 0.413 | 1.516 | 1.904 | 9.99% | - |
| 11 | 1386-2142 | 2142-2394 | 0.739 | 2.235 | 2.859 | 9.99% | - |
| 12 | 1512-2268 | 2268-2520 | 0.582 | 1.845 | 3.039 | 8.52% | - |
| 13 | 1638-2394 | 2394-2646 | 0.659 | 0.390 | 0.309 | 17.77% | - |
| 14 | 1764-2520 | 2520-2772 | 0.471 | 0.757 | 0.659 | 17.77% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_rotation_half_defensive | 0.952 | -0.461 | 2 | benchmark_rotation_defensive | -0.447 | 4 | false |
| 1 | benchmark_rotation_half_defensive | 0.933 | 0.545 | 1 | benchmark_rotation_half_defensive | 0.545 | 4 | false |
| 2 | benchmark_rotation_defensive | 0.452 | 4.096 | 2 | benchmark_rotation_half_defensive | 4.263 | 4 | false |
| 3 | benchmark_rotation_defensive | 0.778 | -0.011 | 3 | benchmark_rotation_half_defensive | 0.103 | 4 | true |
| 4 | benchmark_rotation_defensive | 0.750 | 0.144 | 3 | benchmark_rotation_half_defensive | 0.263 | 4 | true |
| 5 | benchmark_rotation_half_defensive | 0.258 | 4.009 | 1 | benchmark_rotation_half_defensive | 4.009 | 4 | false |
| 6 | benchmark_rotation_half_defensive | 0.285 | 5.814 | 2 | benchmark_rotation_trend126 | 5.839 | 4 | false |
| 7 | benchmark_rotation_half_defensive | 0.475 | -0.395 | 4 | benchmark_rotation_defensive | -0.353 | 4 | true |
| 8 | benchmark_rotation_half_defensive | 0.674 | -0.697 | 1 | benchmark_rotation_half_defensive | -0.697 | 4 | false |
| 9 | benchmark_rotation_half_defensive | 0.291 | 0.693 | 1 | benchmark_rotation_half_defensive | 0.693 | 4 | false |
| 10 | benchmark_rotation_half_defensive | 0.217 | 2.113 | 1 | benchmark_rotation_half_defensive | 2.113 | 4 | false |
| 11 | benchmark_rotation_half_defensive | 0.481 | 2.855 | 4 | benchmark_rotation_cash_defensive | 2.864 | 4 | true |
| 12 | benchmark_rotation_half_defensive | 0.354 | 3.046 | 1 | benchmark_rotation_half_defensive | 3.046 | 4 | false |
| 13 | benchmark_rotation_half_defensive | 0.400 | 0.481 | 1 | benchmark_rotation_half_defensive | 0.481 | 4 | false |
| 14 | benchmark_rotation_half_defensive | 0.268 | 0.836 | 1 | benchmark_rotation_half_defensive | 0.836 | 4 | false |

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
| normal | 379.30% | 14.76% | 0.859 | 0.811 | 0.441 | 33.46% | 127 | 26.188 | - |
| stress_2x | 378.18% | 14.74% | 0.858 | 0.810 | 0.440 | 33.46% | 127 | 26.152 | - |
| stress_3x | 377.07% | 14.72% | 0.857 | 0.809 | 0.440 | 33.47% | 127 | 26.116 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.032 | -0.386 | -0.410 | 19.87% | - |
| 1 | 126-882 | 882-1134 | 0.966 | 0.639 | 0.459 | 19.87% | - |
| 2 | 252-1008 | 1008-1260 | 0.696 | 2.181 | 4.161 | 6.84% | - |
| 3 | 378-1134 | 1134-1386 | 1.163 | 0.448 | 0.296 | 33.46% | - |
| 4 | 504-1260 | 1260-1512 | 1.087 | 0.651 | 0.544 | 33.46% | - |
| 5 | 630-1386 | 1386-1638 | 0.588 | 2.242 | 4.142 | 10.07% | - |
| 6 | 756-1512 | 1512-1764 | 0.621 | 1.992 | 5.758 | 5.34% | - |
| 7 | 882-1638 | 1638-1890 | 0.862 | -0.263 | -0.332 | 19.65% | - |
| 8 | 1008-1764 | 1764-2016 | 1.140 | -0.572 | -0.672 | 21.14% | - |
| 9 | 1134-1890 | 1890-2142 | 0.632 | 0.857 | 0.962 | 15.62% | - |
| 10 | 1260-2016 | 2016-2268 | 0.493 | 1.709 | 2.418 | 10.36% | - |
| 11 | 1386-2142 | 2142-2394 | 0.898 | 2.386 | 3.213 | 10.36% | - |
| 12 | 1512-2268 | 2268-2520 | 0.737 | 1.792 | 2.699 | 10.14% | - |
| 13 | 1638-2394 | 2394-2646 | 0.844 | 0.489 | 0.432 | 18.86% | - |
| 14 | 1764-2520 | 2520-2772 | 0.605 | 0.942 | 0.905 | 18.86% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_medium | 1.070 | -0.406 | 3 | benchmark_tsmom_reb42 | -0.378 | 4 | true |
| 1 | benchmark_tsmom_medium | 1.069 | 0.452 | 4 | benchmark_tsmom_reb42 | 0.544 | 4 | true |
| 2 | benchmark_tsmom_reb42 | 0.478 | 4.267 | 2 | benchmark_tsmom_slow | 4.969 | 4 | false |
| 3 | benchmark_tsmom_reb42 | 0.807 | 0.278 | 4 | benchmark_tsmom_slow | 0.338 | 4 | true |
| 4 | benchmark_tsmom_reb42 | 0.794 | 0.512 | 4 | benchmark_tsmom_checkpoint | 0.544 | 4 | true |
| 5 | benchmark_tsmom_slow | 0.361 | 4.077 | 3 | benchmark_tsmom_checkpoint | 4.142 | 4 | true |
| 6 | benchmark_tsmom_slow | 0.402 | 5.804 | 2 | benchmark_tsmom_medium | 5.843 | 4 | false |
| 7 | benchmark_tsmom_slow | 0.588 | -0.372 | 4 | benchmark_tsmom_medium | -0.310 | 4 | true |
| 8 | benchmark_tsmom_medium | 0.796 | -0.665 | 1 | benchmark_tsmom_medium | -0.665 | 4 | false |
| 9 | benchmark_tsmom_medium | 0.405 | 1.019 | 1 | benchmark_tsmom_medium | 1.019 | 4 | false |
| 10 | benchmark_tsmom_medium | 0.293 | 2.477 | 2 | benchmark_tsmom_reb42 | 2.478 | 4 | false |
| 11 | benchmark_tsmom_medium | 0.755 | 3.214 | 1 | benchmark_tsmom_medium | 3.214 | 4 | false |
| 12 | benchmark_tsmom_medium | 0.580 | 2.622 | 3 | benchmark_tsmom_slow | 2.806 | 4 | true |
| 13 | benchmark_tsmom_medium | 0.681 | 0.407 | 4 | benchmark_tsmom_slow | 0.483 | 4 | true |
| 14 | benchmark_tsmom_medium | 0.405 | 0.896 | 3 | benchmark_tsmom_reb42 | 1.057 | 4 | true |

## benchmark_tsmom_blend

- Family: `benchmark_tsmom_blend`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.600 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 382.54% | 14.83% | 0.861 | 0.814 | 0.444 | 33.44% | 140 | 24.691 | - |
| stress_2x | 381.49% | 14.81% | 0.860 | 0.813 | 0.443 | 33.44% | 140 | 24.660 | - |
| stress_3x | 380.45% | 14.79% | 0.859 | 0.812 | 0.442 | 33.44% | 140 | 24.629 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.015 | -0.374 | -0.402 | 20.01% | - |
| 1 | 126-882 | 882-1134 | 0.949 | 0.646 | 0.464 | 20.01% | - |
| 2 | 252-1008 | 1008-1260 | 0.712 | 2.385 | 4.735 | 6.45% | - |
| 3 | 378-1134 | 1134-1386 | 1.173 | 0.493 | 0.348 | 33.44% | - |
| 4 | 504-1260 | 1260-1512 | 1.135 | 0.654 | 0.548 | 33.44% | - |
| 5 | 630-1386 | 1386-1638 | 0.614 | 2.196 | 4.057 | 10.13% | - |
| 6 | 756-1512 | 1512-1764 | 0.645 | 1.946 | 5.559 | 5.44% | - |
| 7 | 882-1638 | 1638-1890 | 0.885 | -0.246 | -0.319 | 19.39% | - |
| 8 | 1008-1764 | 1764-2016 | 1.145 | -0.562 | -0.660 | 21.11% | - |
| 9 | 1134-1890 | 1890-2142 | 0.626 | 0.856 | 0.950 | 15.77% | - |
| 10 | 1260-2016 | 2016-2268 | 0.492 | 1.705 | 2.411 | 10.41% | - |
| 11 | 1386-2142 | 2142-2394 | 0.897 | 2.358 | 3.202 | 10.41% | - |
| 12 | 1512-2268 | 2268-2520 | 0.739 | 1.766 | 2.663 | 10.30% | - |
| 13 | 1638-2394 | 2394-2646 | 0.825 | 0.466 | 0.406 | 18.92% | - |
| 14 | 1764-2520 | 2520-2772 | 0.607 | 0.879 | 0.825 | 18.92% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_blend_etf_tilt | 1.047 | -0.401 | 2 | benchmark_tsmom_blend_reb42 | -0.377 | 4 | false |
| 1 | benchmark_tsmom_blend_etf_tilt | 1.048 | 0.458 | 4 | benchmark_tsmom_blend_reb42 | 0.572 | 4 | true |
| 2 | benchmark_tsmom_blend_reb42 | 0.494 | 4.624 | 4 | benchmark_tsmom_blend_etf_tilt | 4.754 | 4 | true |
| 3 | benchmark_tsmom_blend_reb42 | 0.817 | 0.288 | 4 | benchmark_tsmom_blend_etf_tilt | 0.352 | 4 | true |
| 4 | benchmark_tsmom_blend_reb42 | 0.801 | 0.517 | 4 | benchmark_tsmom_blend | 0.548 | 4 | true |
| 5 | benchmark_tsmom_blend_etf_tilt | 0.363 | 4.054 | 4 | benchmark_tsmom_blend_slow_broad | 4.118 | 4 | true |
| 6 | benchmark_tsmom_blend | 0.398 | 5.559 | 4 | benchmark_tsmom_blend_slow_broad | 5.928 | 4 | true |
| 7 | benchmark_tsmom_blend_etf_tilt | 0.593 | -0.306 | 1 | benchmark_tsmom_blend_etf_tilt | -0.306 | 4 | false |
| 8 | benchmark_tsmom_blend_etf_tilt | 0.800 | -0.656 | 1 | benchmark_tsmom_blend_etf_tilt | -0.656 | 4 | false |
| 9 | benchmark_tsmom_blend_etf_tilt | 0.390 | 0.977 | 1 | benchmark_tsmom_blend_etf_tilt | 0.977 | 4 | false |
| 10 | benchmark_tsmom_blend_etf_tilt | 0.287 | 2.455 | 1 | benchmark_tsmom_blend_etf_tilt | 2.455 | 4 | false |
| 11 | benchmark_tsmom_blend_etf_tilt | 0.741 | 3.196 | 2 | benchmark_tsmom_blend | 3.202 | 4 | false |
| 12 | benchmark_tsmom_blend_etf_tilt | 0.571 | 2.625 | 3 | benchmark_tsmom_blend_slow_broad | 2.793 | 4 | true |
| 13 | benchmark_tsmom_blend_etf_tilt | 0.606 | 0.397 | 4 | benchmark_tsmom_blend_slow_broad | 0.488 | 4 | true |
| 14 | benchmark_tsmom_blend_etf_tilt | 0.400 | 0.815 | 4 | benchmark_tsmom_blend_reb42 | 1.063 | 4 | true |

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
