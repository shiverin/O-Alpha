# Alpha Validation Report

- Generated: `2026-06-03T12:00:10Z`
- Symbols: `VOO, DIA, IWM, QQQ, SMH, SPY, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY`
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
| equal_weight | 441.39% | 17.67% | 0.951 | 0.901 | 0.509 | 34.69% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| composite_momentum_checkpoint | buy_hold | false | 336.03% | 15.24% | 0.890 | 0.829 | 0.448 | 33.98% | 1.000 | 0.538 | 218 | PBO 0.538 above 0.200 |
| benchmark_rotation_defensive | buy_hold | false | 258.28% | 13.08% | 0.814 | 0.756 | 0.385 | 33.99% | 1.000 | 0.308 | 198 | PBO 0.308 above 0.200 |
| benchmark_tsmom_checkpoint | buy_hold | false | 368.54% | 16.04% | 0.910 | 0.853 | 0.479 | 33.46% | 1.000 | 0.538 | 119 | PBO 0.538 above 0.200 |
| benchmark_tsmom_blend | buy_hold | false | 377.15% | 16.25% | 0.919 | 0.863 | 0.486 | 33.44% | 1.000 | 0.615 | 129 | PBO 0.615 above 0.200 |
| sector_ranked_sleeve_checkpoint | buy_hold | false | 373.04% | 16.15% | 0.926 | 0.871 | 0.476 | 33.92% | 1.000 | 0.615 | 283 | PBO 0.615 above 0.200 |

## composite_momentum_checkpoint

- Family: `composite_momentum`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.538 above 0.200
  - turnover increases without return improvement
  - no drawdown-adjusted improvement over benchmark

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 336.03% | 15.24% | 0.890 | 0.829 | 0.448 | 33.98% | 218 | 79.659 | - |
| stress_2x | 332.73% | 15.16% | 0.886 | 0.826 | 0.446 | 33.98% | 218 | 79.294 | - |
| stress_3x | 329.45% | 15.07% | 0.881 | 0.822 | 0.444 | 33.98% | 219 | 78.931 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.548 | 2.178 | 4.012 | 7.17% | - |
| 1 | 126-882 | 882-1134 | 1.081 | 0.415 | 0.254 | 33.98% | - |
| 2 | 252-1008 | 1008-1260 | 0.968 | 0.584 | 0.449 | 33.98% | - |
| 3 | 378-1134 | 1134-1386 | 0.526 | 2.321 | 4.283 | 9.50% | - |
| 4 | 504-1260 | 1260-1512 | 0.574 | 2.067 | 5.974 | 5.09% | - |
| 5 | 630-1386 | 1386-1638 | 0.853 | -0.559 | -0.508 | 22.99% | - |
| 6 | 756-1512 | 1512-1764 | 1.109 | -0.902 | -0.780 | 26.79% | - |
| 7 | 882-1638 | 1638-1890 | 0.531 | 0.679 | 0.653 | 17.91% | - |
| 8 | 1008-1764 | 1764-2016 | 0.366 | 1.874 | 2.998 | 8.92% | - |
| 9 | 1134-1890 | 1890-2142 | 0.725 | 2.291 | 3.381 | 8.92% | - |
| 10 | 1260-2016 | 2016-2268 | 0.561 | 1.803 | 3.003 | 8.58% | - |
| 11 | 1386-2142 | 2142-2394 | 0.637 | 0.869 | 0.896 | 17.22% | - |
| 12 | 1512-2268 | 2268-2520 | 0.539 | 1.262 | 1.299 | 17.22% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | composite_momentum_strict_etf | 0.467 | 4.602 | 1 | composite_momentum_strict_etf | 4.602 | 4 | false |
| 1 | composite_momentum_strict_etf | 0.805 | 0.272 | 1 | composite_momentum_strict_etf | 0.272 | 4 | false |
| 2 | composite_momentum_strict_etf | 0.772 | 0.491 | 1 | composite_momentum_strict_etf | 0.491 | 4 | false |
| 3 | composite_momentum_strict_etf | 0.345 | 4.158 | 4 | composite_momentum_sleeve20_broad5 | 4.285 | 4 | true |
| 4 | composite_momentum_strict_etf | 0.387 | 5.771 | 4 | composite_momentum_sleeve20_broad5 | 5.995 | 4 | true |
| 5 | composite_momentum_strict_etf | 0.550 | -0.488 | 1 | composite_momentum_strict_etf | -0.488 | 4 | false |
| 6 | composite_momentum_strict_etf | 0.755 | -0.762 | 1 | composite_momentum_strict_etf | -0.762 | 4 | false |
| 7 | composite_momentum_broader_core | 0.312 | 0.728 | 2 | composite_momentum_strict_etf | 0.910 | 4 | false |
| 8 | composite_momentum_strict_etf | 0.207 | 2.556 | 4 | composite_momentum_checkpoint | 2.998 | 4 | true |
| 9 | composite_momentum_strict_etf | 0.507 | 2.721 | 4 | composite_momentum_checkpoint | 3.381 | 4 | true |
| 10 | composite_momentum_strict_etf | 0.367 | 2.882 | 4 | composite_momentum_sleeve20_broad5 | 3.023 | 4 | true |
| 11 | composite_momentum_strict_etf | 0.399 | 0.672 | 4 | composite_momentum_checkpoint | 0.896 | 4 | true |
| 12 | composite_momentum_strict_etf | 0.325 | 1.036 | 4 | composite_momentum_checkpoint | 1.299 | 4 | true |

## benchmark_rotation_defensive

- Family: `composite_momentum_defensive`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.308 above 0.200
  - turnover increases without return improvement
  - no drawdown-adjusted improvement over benchmark

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 258.28% | 13.08% | 0.814 | 0.756 | 0.385 | 33.99% | 198 | 63.081 | - |
| stress_2x | 255.84% | 13.01% | 0.810 | 0.752 | 0.383 | 33.99% | 198 | 62.836 | - |
| stress_3x | 253.42% | 12.93% | 0.806 | 0.749 | 0.380 | 33.99% | 200 | 62.591 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.686 | 2.099 | 4.096 | 6.63% | - |
| 1 | 126-882 | 882-1134 | 1.167 | 0.151 | -0.011 | 33.99% | - |
| 2 | 252-1008 | 1008-1260 | 1.091 | 0.310 | 0.144 | 33.99% | - |
| 3 | 378-1134 | 1134-1386 | 0.433 | 2.222 | 4.007 | 9.48% | - |
| 4 | 504-1260 | 1260-1512 | 0.462 | 2.050 | 5.803 | 5.09% | - |
| 5 | 630-1386 | 1386-1638 | 0.725 | -0.359 | -0.353 | 22.67% | - |
| 6 | 756-1512 | 1512-1764 | 1.001 | -0.855 | -0.720 | 24.31% | - |
| 7 | 882-1638 | 1638-1890 | 0.497 | 0.598 | 0.549 | 14.73% | - |
| 8 | 1008-1764 | 1764-2016 | 0.413 | 1.516 | 1.904 | 9.99% | - |
| 9 | 1134-1890 | 1890-2142 | 0.739 | 2.235 | 2.859 | 9.99% | - |
| 10 | 1260-2016 | 2016-2268 | 0.582 | 1.845 | 3.039 | 8.52% | - |
| 11 | 1386-2142 | 2142-2394 | 0.659 | 0.390 | 0.309 | 17.77% | - |
| 12 | 1512-2268 | 2268-2520 | 0.471 | 0.757 | 0.659 | 17.77% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_rotation_defensive | 0.452 | 4.096 | 2 | benchmark_rotation_half_defensive | 4.263 | 4 | false |
| 1 | benchmark_rotation_defensive | 0.778 | -0.011 | 3 | benchmark_rotation_half_defensive | 0.103 | 4 | true |
| 2 | benchmark_rotation_defensive | 0.750 | 0.144 | 3 | benchmark_rotation_half_defensive | 0.263 | 4 | true |
| 3 | benchmark_rotation_half_defensive | 0.258 | 4.009 | 1 | benchmark_rotation_half_defensive | 4.009 | 4 | false |
| 4 | benchmark_rotation_half_defensive | 0.285 | 5.814 | 2 | benchmark_rotation_trend126 | 5.839 | 4 | false |
| 5 | benchmark_rotation_half_defensive | 0.475 | -0.395 | 4 | benchmark_rotation_defensive | -0.353 | 4 | true |
| 6 | benchmark_rotation_half_defensive | 0.674 | -0.697 | 1 | benchmark_rotation_half_defensive | -0.697 | 4 | false |
| 7 | benchmark_rotation_half_defensive | 0.291 | 0.693 | 1 | benchmark_rotation_half_defensive | 0.693 | 4 | false |
| 8 | benchmark_rotation_half_defensive | 0.217 | 2.113 | 1 | benchmark_rotation_half_defensive | 2.113 | 4 | false |
| 9 | benchmark_rotation_half_defensive | 0.481 | 2.855 | 4 | benchmark_rotation_cash_defensive | 2.864 | 4 | true |
| 10 | benchmark_rotation_half_defensive | 0.354 | 3.046 | 1 | benchmark_rotation_half_defensive | 3.046 | 4 | false |
| 11 | benchmark_rotation_half_defensive | 0.400 | 0.481 | 1 | benchmark_rotation_half_defensive | 0.481 | 4 | false |
| 12 | benchmark_rotation_half_defensive | 0.268 | 0.836 | 1 | benchmark_rotation_half_defensive | 0.836 | 4 | false |

## benchmark_tsmom_checkpoint

- Family: `benchmark_tsmom`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.538 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 368.54% | 16.04% | 0.910 | 0.853 | 0.479 | 33.46% | 119 | 24.966 | - |
| stress_2x | 367.51% | 16.02% | 0.909 | 0.852 | 0.479 | 33.46% | 119 | 24.934 | - |
| stress_3x | 366.48% | 15.99% | 0.908 | 0.851 | 0.478 | 33.47% | 119 | 24.902 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.696 | 2.181 | 4.161 | 6.84% | - |
| 1 | 126-882 | 882-1134 | 1.163 | 0.448 | 0.296 | 33.46% | - |
| 2 | 252-1008 | 1008-1260 | 1.087 | 0.651 | 0.544 | 33.46% | - |
| 3 | 378-1134 | 1134-1386 | 0.588 | 2.242 | 4.142 | 10.07% | - |
| 4 | 504-1260 | 1260-1512 | 0.621 | 1.992 | 5.758 | 5.34% | - |
| 5 | 630-1386 | 1386-1638 | 0.862 | -0.263 | -0.332 | 19.65% | - |
| 6 | 756-1512 | 1512-1764 | 1.140 | -0.572 | -0.672 | 21.14% | - |
| 7 | 882-1638 | 1638-1890 | 0.632 | 0.857 | 0.962 | 15.62% | - |
| 8 | 1008-1764 | 1764-2016 | 0.493 | 1.709 | 2.418 | 10.36% | - |
| 9 | 1134-1890 | 1890-2142 | 0.898 | 2.386 | 3.213 | 10.36% | - |
| 10 | 1260-2016 | 2016-2268 | 0.737 | 1.792 | 2.699 | 10.14% | - |
| 11 | 1386-2142 | 2142-2394 | 0.844 | 0.489 | 0.432 | 18.86% | - |
| 12 | 1512-2268 | 2268-2520 | 0.605 | 0.942 | 0.905 | 18.86% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_reb42 | 0.478 | 4.267 | 2 | benchmark_tsmom_slow | 4.969 | 4 | false |
| 1 | benchmark_tsmom_reb42 | 0.807 | 0.278 | 4 | benchmark_tsmom_slow | 0.338 | 4 | true |
| 2 | benchmark_tsmom_reb42 | 0.794 | 0.512 | 4 | benchmark_tsmom_checkpoint | 0.544 | 4 | true |
| 3 | benchmark_tsmom_slow | 0.361 | 4.077 | 3 | benchmark_tsmom_checkpoint | 4.142 | 4 | true |
| 4 | benchmark_tsmom_slow | 0.402 | 5.804 | 2 | benchmark_tsmom_medium | 5.843 | 4 | false |
| 5 | benchmark_tsmom_slow | 0.588 | -0.372 | 4 | benchmark_tsmom_medium | -0.310 | 4 | true |
| 6 | benchmark_tsmom_medium | 0.796 | -0.665 | 1 | benchmark_tsmom_medium | -0.665 | 4 | false |
| 7 | benchmark_tsmom_medium | 0.405 | 1.019 | 1 | benchmark_tsmom_medium | 1.019 | 4 | false |
| 8 | benchmark_tsmom_medium | 0.293 | 2.477 | 2 | benchmark_tsmom_reb42 | 2.478 | 4 | false |
| 9 | benchmark_tsmom_medium | 0.755 | 3.214 | 1 | benchmark_tsmom_medium | 3.214 | 4 | false |
| 10 | benchmark_tsmom_medium | 0.580 | 2.622 | 3 | benchmark_tsmom_slow | 2.806 | 4 | true |
| 11 | benchmark_tsmom_medium | 0.681 | 0.407 | 4 | benchmark_tsmom_slow | 0.483 | 4 | true |
| 12 | benchmark_tsmom_medium | 0.405 | 0.896 | 3 | benchmark_tsmom_reb42 | 1.057 | 4 | true |

## benchmark_tsmom_blend

- Family: `benchmark_tsmom_blend`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.615 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 377.15% | 16.25% | 0.919 | 0.863 | 0.486 | 33.44% | 129 | 23.755 | - |
| stress_2x | 376.19% | 16.22% | 0.918 | 0.862 | 0.485 | 33.44% | 129 | 23.727 | - |
| stress_3x | 375.22% | 16.20% | 0.916 | 0.861 | 0.484 | 33.44% | 129 | 23.699 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.712 | 2.385 | 4.735 | 6.45% | - |
| 1 | 126-882 | 882-1134 | 1.173 | 0.493 | 0.348 | 33.44% | - |
| 2 | 252-1008 | 1008-1260 | 1.135 | 0.654 | 0.548 | 33.44% | - |
| 3 | 378-1134 | 1134-1386 | 0.614 | 2.196 | 4.057 | 10.13% | - |
| 4 | 504-1260 | 1260-1512 | 0.645 | 1.946 | 5.559 | 5.44% | - |
| 5 | 630-1386 | 1386-1638 | 0.885 | -0.246 | -0.319 | 19.39% | - |
| 6 | 756-1512 | 1512-1764 | 1.145 | -0.562 | -0.660 | 21.11% | - |
| 7 | 882-1638 | 1638-1890 | 0.626 | 0.856 | 0.950 | 15.77% | - |
| 8 | 1008-1764 | 1764-2016 | 0.492 | 1.705 | 2.411 | 10.41% | - |
| 9 | 1134-1890 | 1890-2142 | 0.897 | 2.358 | 3.202 | 10.41% | - |
| 10 | 1260-2016 | 2016-2268 | 0.739 | 1.766 | 2.663 | 10.30% | - |
| 11 | 1386-2142 | 2142-2394 | 0.825 | 0.466 | 0.406 | 18.92% | - |
| 12 | 1512-2268 | 2268-2520 | 0.607 | 0.879 | 0.825 | 18.92% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_blend_reb42 | 0.494 | 4.624 | 4 | benchmark_tsmom_blend_etf_tilt | 4.754 | 4 | true |
| 1 | benchmark_tsmom_blend_reb42 | 0.817 | 0.288 | 4 | benchmark_tsmom_blend_etf_tilt | 0.352 | 4 | true |
| 2 | benchmark_tsmom_blend_reb42 | 0.801 | 0.517 | 4 | benchmark_tsmom_blend | 0.548 | 4 | true |
| 3 | benchmark_tsmom_blend_etf_tilt | 0.363 | 4.054 | 4 | benchmark_tsmom_blend_slow_broad | 4.118 | 4 | true |
| 4 | benchmark_tsmom_blend | 0.398 | 5.559 | 4 | benchmark_tsmom_blend_slow_broad | 5.928 | 4 | true |
| 5 | benchmark_tsmom_blend_etf_tilt | 0.593 | -0.306 | 1 | benchmark_tsmom_blend_etf_tilt | -0.306 | 4 | false |
| 6 | benchmark_tsmom_blend_etf_tilt | 0.800 | -0.656 | 1 | benchmark_tsmom_blend_etf_tilt | -0.656 | 4 | false |
| 7 | benchmark_tsmom_blend_etf_tilt | 0.390 | 0.977 | 1 | benchmark_tsmom_blend_etf_tilt | 0.977 | 4 | false |
| 8 | benchmark_tsmom_blend_etf_tilt | 0.287 | 2.455 | 1 | benchmark_tsmom_blend_etf_tilt | 2.455 | 4 | false |
| 9 | benchmark_tsmom_blend_etf_tilt | 0.741 | 3.196 | 2 | benchmark_tsmom_blend | 3.202 | 4 | false |
| 10 | benchmark_tsmom_blend_etf_tilt | 0.571 | 2.625 | 3 | benchmark_tsmom_blend_slow_broad | 2.793 | 4 | true |
| 11 | benchmark_tsmom_blend_etf_tilt | 0.606 | 0.397 | 4 | benchmark_tsmom_blend_slow_broad | 0.488 | 4 | true |
| 12 | benchmark_tsmom_blend_etf_tilt | 0.400 | 0.815 | 4 | benchmark_tsmom_blend_reb42 | 1.063 | 4 | true |

## sector_ranked_sleeve_checkpoint

- Family: `sector_ranked_sleeve`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.615 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 373.04% | 16.15% | 0.926 | 0.871 | 0.476 | 33.92% | 283 | 56.231 | - |
| stress_2x | 370.81% | 16.10% | 0.924 | 0.869 | 0.474 | 33.93% | 283 | 56.070 | - |
| stress_3x | 368.58% | 16.04% | 0.921 | 0.866 | 0.473 | 33.94% | 283 | 55.910 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.702 | 2.496 | 5.499 | 5.84% | - |
| 1 | 126-882 | 882-1134 | 1.159 | 0.540 | 0.390 | 33.92% | - |
| 2 | 252-1008 | 1008-1260 | 1.186 | 0.700 | 0.587 | 33.92% | - |
| 3 | 378-1134 | 1134-1386 | 0.646 | 2.206 | 4.019 | 10.18% | - |
| 4 | 504-1260 | 1260-1512 | 0.708 | 1.766 | 5.580 | 4.82% | - |
| 5 | 630-1386 | 1386-1638 | 0.927 | -0.304 | -0.383 | 19.22% | - |
| 6 | 756-1512 | 1512-1764 | 1.152 | -0.565 | -0.691 | 20.56% | - |
| 7 | 882-1638 | 1638-1890 | 0.611 | 0.701 | 0.754 | 15.76% | - |
| 8 | 1008-1764 | 1764-2016 | 0.458 | 1.543 | 2.134 | 10.30% | - |
| 9 | 1134-1890 | 1890-2142 | 0.796 | 2.291 | 3.124 | 10.30% | - |
| 10 | 1260-2016 | 2016-2268 | 0.671 | 1.728 | 2.601 | 9.79% | - |
| 11 | 1386-2142 | 2142-2394 | 0.725 | 0.518 | 0.465 | 18.28% | - |
| 12 | 1512-2268 | 2268-2520 | 0.520 | 1.068 | 1.066 | 18.28% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | sector_ranked_sleeve_checkpoint | 0.475 | 5.499 | 1 | sector_ranked_sleeve_checkpoint | 5.499 | 4 | false |
| 1 | sector_ranked_sleeve_conservative | 0.784 | 0.282 | 4 | sector_ranked_sleeve_checkpoint | 0.390 | 4 | true |
| 2 | sector_ranked_sleeve_checkpoint | 0.831 | 0.587 | 2 | sector_ranked_sleeve_slow | 0.594 | 4 | false |
| 3 | sector_ranked_sleeve_checkpoint | 0.371 | 4.019 | 4 | sector_ranked_sleeve_slow | 4.216 | 4 | true |
| 4 | sector_ranked_sleeve_checkpoint | 0.433 | 5.580 | 2 | sector_ranked_sleeve_medium | 5.693 | 4 | false |
| 5 | sector_ranked_sleeve_checkpoint | 0.605 | -0.383 | 4 | sector_ranked_sleeve_medium | -0.379 | 4 | true |
| 6 | sector_ranked_sleeve_slow | 0.798 | -0.664 | 1 | sector_ranked_sleeve_slow | -0.664 | 4 | false |
| 7 | sector_ranked_sleeve_medium | 0.371 | 0.779 | 3 | sector_ranked_sleeve_conservative | 0.969 | 4 | true |
| 8 | sector_ranked_sleeve_slow | 0.269 | 2.157 | 3 | sector_ranked_sleeve_conservative | 2.555 | 4 | true |
| 9 | sector_ranked_sleeve_conservative | 0.663 | 2.792 | 4 | sector_ranked_sleeve_slow | 3.162 | 4 | true |
| 10 | sector_ranked_sleeve_conservative | 0.524 | 2.288 | 4 | sector_ranked_sleeve_slow | 2.830 | 4 | true |
| 11 | sector_ranked_sleeve_conservative | 0.578 | 0.523 | 1 | sector_ranked_sleeve_conservative | 0.523 | 4 | false |
| 12 | sector_ranked_sleeve_slow | 0.349 | 0.841 | 4 | sector_ranked_sleeve_conservative | 1.103 | 4 | true |
