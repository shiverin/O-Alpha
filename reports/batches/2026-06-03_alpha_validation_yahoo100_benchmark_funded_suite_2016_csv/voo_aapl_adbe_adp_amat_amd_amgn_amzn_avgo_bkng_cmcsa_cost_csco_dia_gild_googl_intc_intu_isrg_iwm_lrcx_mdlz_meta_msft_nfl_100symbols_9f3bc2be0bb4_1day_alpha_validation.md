# Alpha Validation Report

- Generated: `2026-06-03T11:26:32Z`
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
| benchmark_rotation_defensive | buy_hold | false | 279.57% | 13.71% | 0.843 | 0.783 | 0.402 | 34.10% | 1.000 | 0.385 | 546 | PBO 0.385 above 0.200 |
| benchmark_tsmom_checkpoint | buy_hold | false | 392.87% | 16.61% | 0.925 | 0.869 | 0.494 | 33.64% | 1.000 | 0.538 | 241 | PBO 0.538 above 0.200 |
| benchmark_tsmom_blend | buy_hold | false | 418.28% | 17.17% | 0.945 | 0.888 | 0.507 | 33.90% | 1.000 | 0.615 | 298 | PBO 0.615 above 0.200 |
| benchmark_lowvol_checkpoint | buy_hold | false | 326.24% | 14.99% | 0.921 | 0.867 | 0.459 | 32.68% | 1.000 | 0.538 | 599 | PBO 0.538 above 0.200 |
| benchmark_reversal_checkpoint | buy_hold | false | 388.84% | 16.52% | 0.931 | 0.871 | 0.476 | 34.72% | 1.000 | 0.538 | 4023 | PBO 0.538 above 0.200 |
| benchmark_ranked_sleeve_checkpoint | buy_hold | false | 450.44% | 17.86% | 0.995 | 0.934 | 0.518 | 34.45% | 1.000 | 0.308 | 507 | PBO 0.308 above 0.200 |
| benchmark_ranker_proxy_h63_checkpoint | buy_hold | false | 451.32% | 17.87% | 1.006 | 0.934 | 0.501 | 35.67% | 1.000 | 0.231 | 129 | PBO 0.231 above 0.200 |
| benchmark_ranker_proxy_h84_checkpoint | buy_hold | false | 437.50% | 17.59% | 0.996 | 0.921 | 0.495 | 35.51% | 1.000 | 0.308 | 133 | PBO 0.308 above 0.200 |
| sector_ranked_sleeve_checkpoint | buy_hold | false | 373.04% | 16.15% | 0.926 | 0.871 | 0.476 | 33.92% | 1.000 | 0.615 | 283 | PBO 0.615 above 0.200 |

## benchmark_rotation_defensive

- Family: `composite_momentum_defensive`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.385 above 0.200
  - turnover increases without return improvement
  - no drawdown-adjusted improvement over benchmark

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 279.57% | 13.71% | 0.843 | 0.783 | 0.402 | 34.10% | 546 | 69.108 | - |
| stress_2x | 276.85% | 13.63% | 0.839 | 0.779 | 0.400 | 34.10% | 546 | 68.822 | - |
| stress_3x | 274.15% | 13.55% | 0.835 | 0.775 | 0.397 | 34.10% | 546 | 68.537 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.740 | 2.078 | 3.937 | 6.85% | - |
| 1 | 126-882 | 882-1134 | 1.189 | 0.189 | 0.025 | 34.10% | - |
| 2 | 252-1008 | 1008-1260 | 1.114 | 0.360 | 0.194 | 34.10% | - |
| 3 | 378-1134 | 1134-1386 | 0.459 | 2.244 | 4.061 | 9.58% | - |
| 4 | 504-1260 | 1260-1512 | 0.487 | 2.043 | 5.664 | 5.27% | - |
| 5 | 630-1386 | 1386-1638 | 0.749 | -0.352 | -0.349 | 22.63% | - |
| 6 | 756-1512 | 1512-1764 | 1.028 | -0.851 | -0.718 | 24.20% | - |
| 7 | 882-1638 | 1638-1890 | 0.527 | 0.618 | 0.573 | 14.73% | - |
| 8 | 1008-1764 | 1764-2016 | 0.422 | 1.536 | 1.956 | 9.97% | - |
| 9 | 1134-1890 | 1890-2142 | 0.750 | 2.218 | 2.896 | 9.97% | - |
| 10 | 1260-2016 | 2016-2268 | 0.592 | 1.805 | 2.965 | 8.70% | - |
| 11 | 1386-2142 | 2142-2394 | 0.674 | 0.360 | 0.276 | 18.00% | - |
| 12 | 1512-2268 | 2268-2520 | 0.482 | 0.772 | 0.668 | 18.00% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_rotation_defensive | 0.492 | 3.937 | 3 | benchmark_rotation_half_defensive | 4.152 | 4 | true |
| 1 | benchmark_rotation_defensive | 0.793 | 0.025 | 3 | benchmark_rotation_half_defensive | 0.128 | 4 | true |
| 2 | benchmark_rotation_defensive | 0.766 | 0.194 | 3 | benchmark_rotation_half_defensive | 0.298 | 4 | true |
| 3 | benchmark_rotation_half_defensive | 0.270 | 4.044 | 2 | benchmark_rotation_defensive | 4.061 | 4 | false |
| 4 | benchmark_rotation_half_defensive | 0.297 | 5.721 | 1 | benchmark_rotation_half_defensive | 5.721 | 4 | false |
| 5 | benchmark_rotation_half_defensive | 0.487 | -0.392 | 4 | benchmark_rotation_defensive | -0.349 | 4 | true |
| 6 | benchmark_rotation_half_defensive | 0.688 | -0.696 | 1 | benchmark_rotation_half_defensive | -0.696 | 4 | false |
| 7 | benchmark_rotation_half_defensive | 0.306 | 0.708 | 1 | benchmark_rotation_half_defensive | 0.708 | 4 | false |
| 8 | benchmark_rotation_defensive | 0.223 | 1.956 | 2 | benchmark_rotation_half_defensive | 2.148 | 4 | false |
| 9 | benchmark_rotation_half_defensive | 0.487 | 2.880 | 4 | benchmark_rotation_cash_defensive | 2.925 | 4 | true |
| 10 | benchmark_rotation_half_defensive | 0.360 | 2.997 | 1 | benchmark_rotation_half_defensive | 2.997 | 4 | false |
| 11 | benchmark_rotation_half_defensive | 0.408 | 0.456 | 1 | benchmark_rotation_half_defensive | 0.456 | 4 | false |
| 12 | benchmark_rotation_half_defensive | 0.274 | 0.841 | 1 | benchmark_rotation_half_defensive | 0.841 | 4 | false |

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
| normal | 392.87% | 16.61% | 0.925 | 0.869 | 0.494 | 33.64% | 241 | 31.831 | - |
| stress_2x | 391.57% | 16.58% | 0.923 | 0.868 | 0.493 | 33.64% | 241 | 31.780 | - |
| stress_3x | 390.27% | 16.55% | 0.922 | 0.867 | 0.492 | 33.65% | 241 | 31.729 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.729 | 2.176 | 4.307 | 6.66% | - |
| 1 | 126-882 | 882-1134 | 1.171 | 0.461 | 0.312 | 33.64% | - |
| 2 | 252-1008 | 1008-1260 | 1.087 | 0.646 | 0.541 | 33.64% | - |
| 3 | 378-1134 | 1134-1386 | 0.594 | 2.126 | 3.837 | 10.45% | - |
| 4 | 504-1260 | 1260-1512 | 0.616 | 1.946 | 5.221 | 5.84% | - |
| 5 | 630-1386 | 1386-1638 | 0.844 | -0.222 | -0.301 | 19.19% | - |
| 6 | 756-1512 | 1512-1764 | 1.122 | -0.543 | -0.649 | 20.70% | - |
| 7 | 882-1638 | 1638-1890 | 0.626 | 0.896 | 1.045 | 14.96% | - |
| 8 | 1008-1764 | 1764-2016 | 0.506 | 1.705 | 2.463 | 10.29% | - |
| 9 | 1134-1890 | 1890-2142 | 0.929 | 2.427 | 3.481 | 10.29% | - |
| 10 | 1260-2016 | 2016-2268 | 0.764 | 1.805 | 2.723 | 10.73% | - |
| 11 | 1386-2142 | 2142-2394 | 0.890 | 0.480 | 0.424 | 19.03% | - |
| 12 | 1512-2268 | 2268-2520 | 0.631 | 0.929 | 0.885 | 19.03% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_reb42 | 0.513 | 4.731 | 2 | benchmark_tsmom_slow | 5.187 | 4 | false |
| 1 | benchmark_tsmom_reb42 | 0.838 | 0.278 | 4 | benchmark_tsmom_slow | 0.358 | 4 | true |
| 2 | benchmark_tsmom_reb42 | 0.827 | 0.555 | 1 | benchmark_tsmom_reb42 | 0.555 | 4 | false |
| 3 | benchmark_tsmom_slow | 0.372 | 3.698 | 4 | benchmark_tsmom_reb42 | 4.063 | 4 | true |
| 4 | benchmark_tsmom_reb42 | 0.405 | 5.039 | 4 | benchmark_tsmom_medium | 5.482 | 4 | true |
| 5 | benchmark_tsmom_medium | 0.581 | -0.290 | 1 | benchmark_tsmom_medium | -0.290 | 4 | false |
| 6 | benchmark_tsmom_medium | 0.799 | -0.636 | 1 | benchmark_tsmom_medium | -0.636 | 4 | false |
| 7 | benchmark_tsmom_medium | 0.403 | 1.152 | 1 | benchmark_tsmom_medium | 1.152 | 4 | false |
| 8 | benchmark_tsmom_medium | 0.297 | 2.536 | 2 | benchmark_tsmom_reb42 | 2.752 | 4 | false |
| 9 | benchmark_tsmom_medium | 0.807 | 3.386 | 4 | benchmark_tsmom_reb42 | 3.503 | 4 | true |
| 10 | benchmark_tsmom_medium | 0.632 | 2.584 | 3 | benchmark_tsmom_slow | 2.831 | 4 | true |
| 11 | benchmark_tsmom_medium | 0.742 | 0.358 | 4 | benchmark_tsmom_reb42 | 0.484 | 4 | true |
| 12 | benchmark_tsmom_medium | 0.434 | 0.849 | 3 | benchmark_tsmom_reb42 | 1.184 | 4 | true |

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
| normal | 418.28% | 17.17% | 0.945 | 0.888 | 0.507 | 33.90% | 298 | 33.424 | - |
| stress_2x | 416.92% | 17.15% | 0.944 | 0.887 | 0.506 | 33.90% | 298 | 33.371 | - |
| stress_3x | 415.55% | 17.12% | 0.943 | 0.886 | 0.505 | 33.90% | 298 | 33.318 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.755 | 2.466 | 5.331 | 5.97% | - |
| 1 | 126-882 | 882-1134 | 1.185 | 0.542 | 0.403 | 33.90% | - |
| 2 | 252-1008 | 1008-1260 | 1.147 | 0.659 | 0.555 | 33.90% | - |
| 3 | 378-1134 | 1134-1386 | 0.634 | 2.048 | 3.687 | 10.61% | - |
| 4 | 504-1260 | 1260-1512 | 0.652 | 1.867 | 4.926 | 6.06% | - |
| 5 | 630-1386 | 1386-1638 | 0.881 | -0.189 | -0.277 | 18.66% | - |
| 6 | 756-1512 | 1512-1764 | 1.137 | -0.502 | -0.624 | 20.05% | - |
| 7 | 882-1638 | 1638-1890 | 0.619 | 0.961 | 1.154 | 14.62% | - |
| 8 | 1008-1764 | 1764-2016 | 0.505 | 1.731 | 2.498 | 10.45% | - |
| 9 | 1134-1890 | 1890-2142 | 0.950 | 2.427 | 3.540 | 10.45% | - |
| 10 | 1260-2016 | 2016-2268 | 0.797 | 1.759 | 2.630 | 11.23% | - |
| 11 | 1386-2142 | 2142-2394 | 0.909 | 0.420 | 0.352 | 19.38% | - |
| 12 | 1512-2268 | 2268-2520 | 0.654 | 0.838 | 0.768 | 19.38% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_tsmom_blend_reb42 | 0.530 | 4.907 | 4 | benchmark_tsmom_blend | 5.331 | 4 | true |
| 1 | benchmark_tsmom_blend_reb42 | 0.849 | 0.294 | 4 | benchmark_tsmom_blend | 0.403 | 4 | true |
| 2 | benchmark_tsmom_blend_reb42 | 0.826 | 0.561 | 1 | benchmark_tsmom_blend_reb42 | 0.561 | 4 | false |
| 3 | benchmark_tsmom_blend | 0.380 | 3.687 | 4 | benchmark_tsmom_blend_reb42 | 3.991 | 4 | true |
| 4 | benchmark_tsmom_blend_reb42 | 0.412 | 5.084 | 3 | benchmark_tsmom_blend_etf_tilt | 5.253 | 4 | true |
| 5 | benchmark_tsmom_blend_etf_tilt | 0.592 | -0.280 | 2 | benchmark_tsmom_blend | -0.277 | 4 | false |
| 6 | benchmark_tsmom_blend_etf_tilt | 0.799 | -0.637 | 2 | benchmark_tsmom_blend | -0.624 | 4 | false |
| 7 | benchmark_tsmom_blend_etf_tilt | 0.388 | 1.105 | 3 | benchmark_tsmom_blend | 1.154 | 4 | true |
| 8 | benchmark_tsmom_blend | 0.293 | 2.498 | 4 | benchmark_tsmom_blend_reb42 | 2.561 | 4 | true |
| 9 | benchmark_tsmom_blend | 0.806 | 3.540 | 1 | benchmark_tsmom_blend | 3.540 | 4 | false |
| 10 | benchmark_tsmom_blend | 0.633 | 2.630 | 2 | benchmark_tsmom_blend_slow_broad | 2.775 | 4 | false |
| 11 | benchmark_tsmom_blend | 0.680 | 0.352 | 4 | benchmark_tsmom_blend_reb42 | 0.441 | 4 | true |
| 12 | benchmark_tsmom_blend | 0.444 | 0.768 | 4 | benchmark_tsmom_blend_reb42 | 1.121 | 4 | true |

## benchmark_lowvol_checkpoint

- Family: `benchmark_lowvol`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.538 above 0.200
  - turnover increases without return improvement

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 326.24% | 14.99% | 0.921 | 0.867 | 0.459 | 32.68% | 599 | 85.091 | - |
| stress_2x | 322.92% | 14.90% | 0.917 | 0.863 | 0.456 | 32.68% | 600 | 84.690 | - |
| stress_3x | 319.64% | 14.82% | 0.912 | 0.859 | 0.453 | 32.69% | 600 | 84.292 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.778 | 2.387 | 5.125 | 5.75% | - |
| 1 | 126-882 | 882-1134 | 1.286 | 0.464 | 0.315 | 32.68% | - |
| 2 | 252-1008 | 1008-1260 | 1.174 | 0.657 | 0.543 | 32.68% | - |
| 3 | 378-1134 | 1134-1386 | 0.643 | 2.391 | 4.402 | 8.42% | - |
| 4 | 504-1260 | 1260-1512 | 0.697 | 2.264 | 6.190 | 4.79% | - |
| 5 | 630-1386 | 1386-1638 | 0.897 | -0.262 | -0.300 | 19.78% | - |
| 6 | 756-1512 | 1512-1764 | 1.195 | -0.617 | -0.662 | 21.76% | - |
| 7 | 882-1638 | 1638-1890 | 0.612 | 0.756 | 0.794 | 15.62% | - |
| 8 | 1008-1764 | 1764-2016 | 0.460 | 1.533 | 1.889 | 10.11% | - |
| 9 | 1134-1890 | 1890-2142 | 0.844 | 2.061 | 2.247 | 10.11% | - |
| 10 | 1260-2016 | 2016-2268 | 0.672 | 2.049 | 4.363 | 5.61% | - |
| 11 | 1386-2142 | 2142-2394 | 0.680 | 0.709 | 0.611 | 18.83% | - |
| 12 | 1512-2268 | 2268-2520 | 0.554 | 1.007 | 0.876 | 18.83% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lowvol_voladj | 0.633 | 5.234 | 2 | benchmark_lowvol_reb42 | 5.341 | 4 | false |
| 1 | benchmark_lowvol_wider | 0.914 | 0.309 | 2 | benchmark_lowvol_checkpoint | 0.315 | 4 | false |
| 2 | benchmark_lowvol_voladj | 0.875 | 0.444 | 4 | benchmark_lowvol_checkpoint | 0.543 | 4 | true |
| 3 | benchmark_lowvol_wider | 0.382 | 4.275 | 4 | benchmark_lowvol_voladj | 4.499 | 4 | true |
| 4 | benchmark_lowvol_checkpoint | 0.429 | 6.190 | 1 | benchmark_lowvol_checkpoint | 6.190 | 4 | false |
| 5 | benchmark_lowvol_checkpoint | 0.580 | -0.300 | 2 | benchmark_lowvol_wider | -0.296 | 4 | false |
| 6 | benchmark_lowvol_checkpoint | 0.802 | -0.662 | 2 | benchmark_lowvol_wider | -0.620 | 4 | false |
| 7 | benchmark_lowvol_checkpoint | 0.368 | 0.794 | 2 | benchmark_lowvol_wider | 0.811 | 4 | false |
| 8 | benchmark_lowvol_checkpoint | 0.251 | 1.889 | 3 | benchmark_lowvol_reb42 | 1.969 | 4 | true |
| 9 | benchmark_lowvol_wider | 0.621 | 2.259 | 3 | benchmark_lowvol_reb42 | 2.492 | 4 | true |
| 10 | benchmark_lowvol_wider | 0.473 | 3.887 | 3 | benchmark_lowvol_checkpoint | 4.363 | 4 | true |
| 11 | benchmark_lowvol_wider | 0.462 | 0.512 | 4 | benchmark_lowvol_reb42 | 0.708 | 4 | true |
| 12 | benchmark_lowvol_checkpoint | 0.341 | 0.876 | 3 | benchmark_lowvol_voladj | 1.020 | 4 | true |

## benchmark_reversal_checkpoint

- Family: `benchmark_reversal`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.538 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 388.84% | 16.52% | 0.931 | 0.871 | 0.476 | 34.72% | 4023 | 241.839 | - |
| stress_2x | 378.89% | 16.29% | 0.920 | 0.861 | 0.469 | 34.73% | 4027 | 238.782 | - |
| stress_3x | 369.16% | 16.06% | 0.909 | 0.851 | 0.462 | 34.75% | 4034 | 235.771 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.820 | 2.307 | 4.289 | 7.33% | - |
| 1 | 126-882 | 882-1134 | 1.281 | 0.390 | 0.223 | 34.65% | - |
| 2 | 252-1008 | 1008-1260 | 1.235 | 0.617 | 0.472 | 35.01% | - |
| 3 | 378-1134 | 1134-1386 | 0.639 | 2.362 | 4.404 | 9.33% | - |
| 4 | 504-1260 | 1260-1512 | 0.673 | 2.131 | 5.927 | 5.37% | - |
| 5 | 630-1386 | 1386-1638 | 0.877 | -0.321 | -0.383 | 21.54% | - |
| 6 | 756-1512 | 1512-1764 | 1.177 | -0.661 | -0.734 | 24.13% | - |
| 7 | 882-1638 | 1638-1890 | 0.585 | 0.692 | 0.704 | 17.59% | - |
| 8 | 1008-1764 | 1764-2016 | 0.409 | 1.813 | 2.713 | 9.34% | - |
| 9 | 1134-1890 | 1890-2142 | 0.804 | 2.332 | 2.913 | 9.75% | - |
| 10 | 1260-2016 | 2016-2268 | 0.702 | 1.807 | 2.615 | 9.43% | - |
| 11 | 1386-2142 | 2142-2394 | 0.689 | 0.755 | 0.802 | 17.44% | - |
| 12 | 1512-2268 | 2268-2520 | 0.547 | 1.248 | 1.377 | 17.55% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_reversal_slow | 0.606 | 4.459 | 1 | benchmark_reversal_slow | 4.459 | 4 | false |
| 1 | benchmark_reversal_slow | 0.887 | 0.260 | 2 | benchmark_reversal_fast | 0.263 | 4 | false |
| 2 | benchmark_reversal_slow | 0.881 | 0.452 | 3 | benchmark_reversal_fast | 0.529 | 4 | true |
| 3 | benchmark_reversal_fast | 0.372 | 4.288 | 4 | benchmark_reversal_medium | 4.648 | 4 | true |
| 4 | benchmark_reversal_medium | 0.443 | 6.032 | 3 | benchmark_reversal_slow | 6.810 | 4 | true |
| 5 | benchmark_reversal_slow | 0.564 | -0.406 | 2 | benchmark_reversal_checkpoint | -0.383 | 4 | false |
| 6 | benchmark_reversal_checkpoint | 0.781 | -0.734 | 3 | benchmark_reversal_medium | -0.700 | 4 | true |
| 7 | benchmark_reversal_fast | 0.357 | 0.794 | 3 | benchmark_reversal_slow | 1.014 | 4 | true |
| 8 | benchmark_reversal_slow | 0.257 | 3.192 | 2 | benchmark_reversal_medium | 3.218 | 4 | false |
| 9 | benchmark_reversal_slow | 0.677 | 3.008 | 2 | benchmark_reversal_medium | 3.319 | 4 | false |
| 10 | benchmark_reversal_slow | 0.540 | 3.642 | 1 | benchmark_reversal_slow | 3.642 | 4 | false |
| 11 | benchmark_reversal_slow | 0.517 | 0.765 | 3 | benchmark_reversal_medium | 0.859 | 4 | true |
| 12 | benchmark_reversal_slow | 0.473 | 1.001 | 3 | benchmark_reversal_checkpoint | 1.377 | 4 | true |

## benchmark_ranked_sleeve_checkpoint

- Family: `benchmark_ranked_sleeve`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.308 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 450.44% | 17.86% | 0.995 | 0.934 | 0.518 | 34.45% | 507 | 96.483 | - |
| stress_2x | 446.25% | 17.77% | 0.991 | 0.931 | 0.516 | 34.46% | 507 | 96.029 | - |
| stress_3x | 442.10% | 17.68% | 0.987 | 0.927 | 0.513 | 34.47% | 507 | 95.577 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.751 | 2.586 | 5.362 | 6.34% | - |
| 1 | 126-882 | 882-1134 | 1.182 | 0.666 | 0.529 | 34.45% | - |
| 2 | 252-1008 | 1008-1260 | 1.211 | 0.843 | 0.762 | 34.45% | - |
| 3 | 378-1134 | 1134-1386 | 0.720 | 2.323 | 4.425 | 10.20% | - |
| 4 | 504-1260 | 1260-1512 | 0.784 | 1.923 | 5.090 | 5.97% | - |
| 5 | 630-1386 | 1386-1638 | 1.023 | -0.399 | -0.425 | 20.86% | - |
| 6 | 756-1512 | 1512-1764 | 1.278 | -0.675 | -0.643 | 24.04% | - |
| 7 | 882-1638 | 1638-1890 | 0.658 | 0.867 | 0.966 | 14.98% | - |
| 8 | 1008-1764 | 1764-2016 | 0.477 | 1.572 | 2.311 | 9.98% | - |
| 9 | 1134-1890 | 1890-2142 | 0.807 | 2.275 | 3.581 | 9.98% | - |
| 10 | 1260-2016 | 2016-2268 | 0.696 | 1.785 | 2.976 | 9.70% | - |
| 11 | 1386-2142 | 2142-2394 | 0.820 | 0.649 | 0.603 | 18.01% | - |
| 12 | 1512-2268 | 2268-2520 | 0.640 | 1.216 | 1.184 | 18.01% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranked_sleeve_conservative | 0.655 | 3.934 | 4 | benchmark_ranked_sleeve_checkpoint | 5.362 | 4 | true |
| 1 | benchmark_ranked_sleeve_conservative | 1.015 | 0.358 | 3 | benchmark_ranked_sleeve_checkpoint | 0.529 | 4 | true |
| 2 | benchmark_ranked_sleeve_conservative | 0.933 | 0.910 | 1 | benchmark_ranked_sleeve_conservative | 0.910 | 4 | false |
| 3 | benchmark_ranked_sleeve_checkpoint | 0.427 | 4.425 | 2 | benchmark_ranked_sleeve_conservative | 4.944 | 4 | false |
| 4 | benchmark_ranked_sleeve_conservative | 0.510 | 5.453 | 1 | benchmark_ranked_sleeve_conservative | 5.453 | 4 | false |
| 5 | benchmark_ranked_sleeve_checkpoint | 0.683 | -0.425 | 2 | benchmark_ranked_sleeve_slow | -0.291 | 4 | false |
| 6 | benchmark_ranked_sleeve_checkpoint | 0.883 | -0.643 | 1 | benchmark_ranked_sleeve_checkpoint | -0.643 | 4 | false |
| 7 | benchmark_ranked_sleeve_conservative | 0.462 | 1.022 | 1 | benchmark_ranked_sleeve_conservative | 1.022 | 4 | false |
| 8 | benchmark_ranked_sleeve_conservative | 0.290 | 3.094 | 1 | benchmark_ranked_sleeve_conservative | 3.094 | 4 | false |
| 9 | benchmark_ranked_sleeve_slow | 0.635 | 4.426 | 1 | benchmark_ranked_sleeve_slow | 4.426 | 4 | false |
| 10 | benchmark_ranked_sleeve_slow | 0.454 | 3.946 | 1 | benchmark_ranked_sleeve_slow | 3.946 | 4 | false |
| 11 | benchmark_ranked_sleeve_slow | 0.571 | 0.428 | 4 | benchmark_ranked_sleeve_conservative | 0.673 | 4 | true |
| 12 | benchmark_ranked_sleeve_slow | 0.494 | 0.673 | 4 | benchmark_ranked_sleeve_conservative | 1.652 | 4 | true |

## benchmark_ranker_proxy_h63_checkpoint

- Family: `benchmark_ranker_proxy_h63`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.231 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 451.32% | 17.87% | 1.006 | 0.934 | 0.501 | 35.67% | 129 | 28.372 | - |
| stress_2x | 450.08% | 17.85% | 1.004 | 0.933 | 0.500 | 35.67% | 129 | 28.332 | - |
| stress_3x | 448.85% | 17.82% | 1.003 | 0.932 | 0.500 | 35.67% | 129 | 28.292 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.961 | 2.295 | 4.713 | 6.42% | - |
| 1 | 126-882 | 882-1134 | 1.366 | 0.493 | 0.325 | 35.67% | - |
| 2 | 252-1008 | 1008-1260 | 1.254 | 0.688 | 0.551 | 35.67% | - |
| 3 | 378-1134 | 1134-1386 | 0.683 | 2.449 | 4.698 | 9.48% | - |
| 4 | 504-1260 | 1260-1512 | 0.711 | 2.361 | 6.339 | 5.80% | - |
| 5 | 630-1386 | 1386-1638 | 0.944 | -0.231 | -0.284 | 21.51% | - |
| 6 | 756-1512 | 1512-1764 | 1.238 | -0.700 | -0.715 | 23.76% | - |
| 7 | 882-1638 | 1638-1890 | 0.684 | 0.904 | 1.021 | 16.01% | - |
| 8 | 1008-1764 | 1764-2016 | 0.509 | 1.919 | 2.925 | 9.68% | - |
| 9 | 1134-1890 | 1890-2142 | 0.942 | 2.344 | 3.205 | 9.68% | - |
| 10 | 1260-2016 | 2016-2268 | 0.742 | 1.888 | 2.984 | 8.91% | - |
| 11 | 1386-2142 | 2142-2394 | 0.836 | 0.560 | 0.500 | 18.91% | - |
| 12 | 1512-2268 | 2268-2520 | 0.666 | 1.070 | 1.047 | 18.91% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranker_proxy_h63_checkpoint | 0.686 | 4.713 | 1 | benchmark_ranker_proxy_h63_checkpoint | 4.713 | 5 | false |
| 1 | benchmark_ranker_proxy_h63_checkpoint | 0.966 | 0.325 | 3 | benchmark_ranker_proxy_h84 | 0.339 | 5 | false |
| 2 | benchmark_ranker_proxy_h63_checkpoint | 0.902 | 0.551 | 2 | benchmark_ranker_proxy_h63_strict | 0.552 | 5 | false |
| 3 | benchmark_ranker_proxy_h63_checkpoint | 0.386 | 4.698 | 2 | benchmark_ranker_proxy_h63_strict | 4.703 | 5 | false |
| 4 | benchmark_ranker_proxy_h63_checkpoint | 0.420 | 6.339 | 1 | benchmark_ranker_proxy_h63_checkpoint | 6.339 | 5 | false |
| 5 | benchmark_ranker_proxy_h63_checkpoint | 0.596 | -0.284 | 1 | benchmark_ranker_proxy_h63_checkpoint | -0.284 | 5 | false |
| 6 | benchmark_ranker_proxy_h63_checkpoint | 0.818 | -0.715 | 2 | benchmark_ranker_proxy_h84 | -0.687 | 5 | false |
| 7 | benchmark_ranker_proxy_h63_checkpoint | 0.411 | 1.021 | 4 | benchmark_ranker_proxy_h84 | 1.145 | 5 | true |
| 8 | benchmark_ranker_proxy_h84 | 0.304 | 3.441 | 1 | benchmark_ranker_proxy_h84 | 3.441 | 5 | false |
| 9 | benchmark_ranker_proxy_h84 | 0.794 | 3.888 | 1 | benchmark_ranker_proxy_h84 | 3.888 | 5 | false |
| 10 | benchmark_ranker_proxy_h84 | 0.614 | 3.038 | 1 | benchmark_ranker_proxy_h84 | 3.038 | 5 | false |
| 11 | benchmark_ranker_proxy_h84 | 0.708 | 0.385 | 5 | benchmark_ranker_proxy_h126 | 0.634 | 5 | true |
| 12 | benchmark_ranker_proxy_h84 | 0.442 | 0.914 | 5 | benchmark_ranker_proxy_h126 | 1.163 | 5 | true |

## benchmark_ranker_proxy_h84_checkpoint

- Family: `benchmark_ranker_proxy_h84`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.308 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 437.50% | 17.59% | 0.996 | 0.921 | 0.495 | 35.51% | 133 | 27.288 | - |
| stress_2x | 436.31% | 17.56% | 0.994 | 0.920 | 0.494 | 35.52% | 133 | 27.251 | - |
| stress_3x | 435.13% | 17.54% | 0.993 | 0.919 | 0.494 | 35.52% | 133 | 27.214 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.867 | 2.282 | 4.496 | 6.67% | - |
| 1 | 126-882 | 882-1134 | 1.228 | 0.507 | 0.339 | 35.51% | - |
| 2 | 252-1008 | 1008-1260 | 1.202 | 0.677 | 0.539 | 35.51% | - |
| 3 | 378-1134 | 1134-1386 | 0.653 | 2.365 | 4.293 | 10.15% | - |
| 4 | 504-1260 | 1260-1512 | 0.700 | 2.210 | 6.001 | 5.77% | - |
| 5 | 630-1386 | 1386-1638 | 0.916 | -0.227 | -0.292 | 20.02% | - |
| 6 | 756-1512 | 1512-1764 | 1.213 | -0.592 | -0.687 | 21.04% | - |
| 7 | 882-1638 | 1638-1890 | 0.682 | 0.963 | 1.145 | 15.34% | - |
| 8 | 1008-1764 | 1764-2016 | 0.518 | 2.029 | 3.441 | 8.71% | - |
| 9 | 1134-1890 | 1890-2142 | 0.983 | 2.569 | 3.888 | 8.71% | - |
| 10 | 1260-2016 | 2016-2268 | 0.817 | 1.944 | 3.038 | 9.01% | - |
| 11 | 1386-2142 | 2142-2394 | 0.925 | 0.465 | 0.385 | 19.31% | - |
| 12 | 1512-2268 | 2268-2520 | 0.678 | 0.977 | 0.914 | 19.31% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranker_proxy_h63 | 0.686 | 4.713 | 1 | benchmark_ranker_proxy_h63 | 4.713 | 5 | false |
| 1 | benchmark_ranker_proxy_h63 | 0.966 | 0.325 | 3 | benchmark_ranker_proxy_h84_checkpoint | 0.339 | 5 | false |
| 2 | benchmark_ranker_proxy_h63 | 0.902 | 0.551 | 1 | benchmark_ranker_proxy_h63 | 0.551 | 5 | false |
| 3 | benchmark_ranker_proxy_h63 | 0.386 | 4.698 | 1 | benchmark_ranker_proxy_h63 | 4.698 | 5 | false |
| 4 | benchmark_ranker_proxy_h63 | 0.420 | 6.339 | 1 | benchmark_ranker_proxy_h63 | 6.339 | 5 | false |
| 5 | benchmark_ranker_proxy_h63 | 0.596 | -0.284 | 1 | benchmark_ranker_proxy_h63 | -0.284 | 5 | false |
| 6 | benchmark_ranker_proxy_h63 | 0.818 | -0.715 | 4 | benchmark_ranker_proxy_h84_checkpoint | -0.687 | 5 | true |
| 7 | benchmark_ranker_proxy_h63 | 0.411 | 1.021 | 5 | benchmark_ranker_proxy_h84_strict | 1.145 | 5 | true |
| 8 | benchmark_ranker_proxy_h84_checkpoint | 0.304 | 3.441 | 1 | benchmark_ranker_proxy_h84_checkpoint | 3.441 | 5 | false |
| 9 | benchmark_ranker_proxy_h84_checkpoint | 0.794 | 3.888 | 1 | benchmark_ranker_proxy_h84_checkpoint | 3.888 | 5 | false |
| 10 | benchmark_ranker_proxy_h84_checkpoint | 0.614 | 3.038 | 3 | benchmark_ranker_proxy_h84_strict | 3.075 | 5 | false |
| 11 | benchmark_ranker_proxy_h84_checkpoint | 0.708 | 0.385 | 5 | benchmark_ranker_proxy_h126 | 0.634 | 5 | true |
| 12 | benchmark_ranker_proxy_h84_strict | 0.444 | 0.918 | 4 | benchmark_ranker_proxy_h126 | 1.163 | 5 | true |

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
