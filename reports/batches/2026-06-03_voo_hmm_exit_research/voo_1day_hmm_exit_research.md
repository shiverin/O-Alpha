# VOO HMM Exit Research

Generated: 2026-06-03T02:36:10Z

Data: 1254 1Day bars from 2021-06-03 to 2026-06-01. Close-to-close return: 81.08%.

## Method

- Benchmark is VOO buy-and-hold through the same backtest engine.
- All active variants start by buying VOO, then go to cash on HMM high-volatility stress, and re-enter VOO when the policy's calm condition is confirmed.
- Signals are generated after a bar closes and execute at the next bar open, matching the existing single-symbol backtest engine.
- HMM observation buckets are calibrated only from historical bars available at that point in the walk-forward timeline.
- The HMM state is high-volatility stress, not a pure bearish-direction classifier.
- No explicit transaction-cost or slippage model is added by this research runner.

## Summary

Benchmark buy-and-hold return is 80.27% with Sharpe 0.791 and max drawdown 25.32%.

Best HMM exit variant by total return is `prob55_sma200_2_3` at 63.66%, excess -16.61% versus buy-and-hold.

Lowest HMM exit max drawdown is `regime_1_1` at 18.48%, a 6.84% drawdown change versus buy-and-hold.

| Strategy | Total Return | Excess vs B&H | Ann. Return | Sharpe | Sortino | Calmar | Max DD | Trades | Exposure | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 80.27% | 0.00% | 12.58% | 0.791 | 0.779 | 0.497 | 25.32% | 2 | 99.92% | 2.80 |
| regime_1_1 | 30.53% | -49.74% | 5.51% | 0.530 | 0.390 | 0.298 | 18.48% | 16 | 62.44% | 18.25 |
| regime_2_3 | 31.44% | -48.83% | 5.65% | 0.546 | 0.404 | 0.306 | 18.48% | 12 | 61.96% | 14.20 |
| prob55_40_2_3 | 33.26% | -47.02% | 5.94% | 0.570 | 0.422 | 0.322 | 18.48% | 12 | 62.04% | 14.29 |
| prob45_30_2_3 | 29.44% | -50.84% | 5.33% | 0.519 | 0.383 | 0.288 | 18.48% | 12 | 61.40% | 14.06 |
| lowvol_reentry_2_3 | 32.59% | -47.68% | 5.84% | 0.564 | 0.414 | 0.316 | 18.48% | 10 | 60.69% | 11.77 |
| regime_sma200_1_3 | 61.93% | -18.35% | 10.18% | 0.856 | 0.698 | 0.551 | 18.48% | 6 | 74.48% | 7.24 |
| prob55_sma200_2_3 | 63.66% | -16.61% | 10.41% | 0.871 | 0.711 | 0.564 | 18.48% | 6 | 74.64% | 7.30 |
| prob55_trend50_2_3 | 40.73% | -39.54% | 7.11% | 0.657 | 0.497 | 0.385 | 18.48% | 10 | 64.35% | 12.24 |

## Policy Details

| Policy | Description | Re-entry rule | Filter | Sells | Buys | Calibration runs | Update errors |
|---|---|---|---|---:|---:|---:|---:|
| regime_1_1 | Exit after one High Vol Stress regime print; re-enter after one non-stress print. | non_stress | none/0 | 7 | 8 | 4 | 0 |
| regime_2_3 | Exit after two consecutive High Vol Stress prints; re-enter after three consecutive non-stress prints. | non_stress | none/0 | 5 | 6 | 4 | 0 |
| prob55_40_2_3 | Exit after high-stress posterior >= 55% for two bars; re-enter below 40% and non-stress for three bars. | probability_non_stress | none/0 | 5 | 6 | 4 | 0 |
| prob45_30_2_3 | Exit after high-stress posterior >= 45% for two bars; re-enter below 30% and non-stress for three bars. | probability_non_stress | none/0 | 5 | 6 | 4 | 0 |
| lowvol_reentry_2_3 | Exit after high-stress posterior >= 55% for two bars; re-enter only after three Low Vol Trend bars. | low_vol | none/0 | 4 | 5 | 4 | 0 |
| regime_sma200_1_3 | Exit after one High Vol Stress print only when price is below its 200-day SMA; re-enter after three non-stress bars above the 200-day SMA. | non_stress | below_sma/200 | 2 | 3 | 4 | 0 |
| prob55_sma200_2_3 | Exit after high-stress posterior >= 55% for two bars only when price is below its 200-day SMA; re-enter below 40% and above the 200-day SMA for three bars. | probability_non_stress | below_sma/200 | 2 | 3 | 4 | 0 |
| prob55_trend50_2_3 | Exit after high-stress posterior >= 55% for two bars only when the trailing 50-day return is negative; re-enter below 40% and positive 50-day return for three bars. | probability_non_stress | negative_return/50 | 4 | 5 | 4 | 0 |

## Trade Ledgers

### regime_1_1

| # | Entry | Exit | Entry Price | Exit Price | Return | PnL |
|---:|---|---|---:|---:|---:|---:|
| 1 | 2021-06-04 | 2022-06-03 | 386.81 | 379.08 | -2.00% | -1998.40 |
| 2 | 2023-04-24 | 2023-04-26 | 378.49 | 373.65 | -1.28% | -1253.21 |
| 3 | 2023-05-15 | 2024-08-05 | 378.73 | 470.15 | 24.14% | 23354.95 |
| 4 | 2024-10-25 | 2024-11-01 | 535.49 | 524.71 | -2.01% | -2416.71 |
| 5 | 2024-11-21 | 2024-11-27 | 545.55 | 551.76 | 1.14% | 1339.63 |
| 6 | 2024-11-29 | 2024-12-19 | 551.17 | 543.60 | -1.37% | -1634.76 |
| 7 | 2025-07-02 | 2026-04-09 | 567.41 | 620.63 | 9.38% | 11010.69 |
| 8 | 2026-05-22 | 2026-06-01 | 685.93 | 697.32 | 1.66% | 2132.14 |

### regime_2_3

| # | Entry | Exit | Entry Price | Exit Price | Return | PnL |
|---:|---|---|---:|---:|---:|---:|
| 1 | 2021-06-04 | 2022-06-06 | 386.81 | 381.25 | -1.44% | -1437.40 |
| 2 | 2023-05-17 | 2024-08-06 | 378.83 | 477.25 | 25.98% | 25605.25 |
| 3 | 2024-10-29 | 2024-11-04 | 532.95 | 525.03 | -1.49% | -1845.22 |
| 4 | 2024-11-25 | 2024-12-20 | 551.59 | 536.62 | -2.71% | -3319.83 |
| 5 | 2025-07-07 | 2026-04-10 | 573.00 | 626.59 | 9.35% | 11129.77 |
| 6 | 2026-05-27 | 2026-06-01 | 690.38 | 697.32 | 1.01% | 1308.15 |

### prob55_40_2_3

| # | Entry | Exit | Entry Price | Exit Price | Return | PnL |
|---:|---|---|---:|---:|---:|---:|
| 1 | 2021-06-04 | 2022-06-06 | 386.81 | 381.25 | -1.44% | -1437.40 |
| 2 | 2023-05-17 | 2024-08-06 | 378.83 | 477.25 | 25.98% | 25605.25 |
| 3 | 2024-10-29 | 2024-11-04 | 532.95 | 525.03 | -1.49% | -1845.22 |
| 4 | 2024-11-25 | 2024-12-23 | 551.59 | 544.02 | -1.37% | -1676.55 |
| 5 | 2025-07-07 | 2026-04-10 | 573.00 | 626.59 | 9.35% | 11283.46 |
| 6 | 2026-05-27 | 2026-06-01 | 690.38 | 697.32 | 1.01% | 1326.21 |

### prob45_30_2_3

| # | Entry | Exit | Entry Price | Exit Price | Return | PnL |
|---:|---|---|---:|---:|---:|---:|
| 1 | 2021-06-04 | 2022-06-06 | 386.81 | 381.25 | -1.44% | -1437.40 |
| 2 | 2023-05-18 | 2024-08-06 | 381.31 | 477.25 | 25.16% | 24797.68 |
| 3 | 2024-10-30 | 2024-11-04 | 534.37 | 525.03 | -1.75% | -2156.16 |
| 4 | 2024-12-03 | 2024-12-20 | 555.01 | 536.62 | -3.31% | -4017.13 |
| 5 | 2025-07-07 | 2026-04-10 | 573.00 | 626.59 | 9.35% | 10959.95 |
| 6 | 2026-05-27 | 2026-06-01 | 690.38 | 697.32 | 1.01% | 1288.19 |

### lowvol_reentry_2_3

| # | Entry | Exit | Entry Price | Exit Price | Return | PnL |
|---:|---|---|---:|---:|---:|---:|
| 1 | 2021-06-04 | 2022-06-06 | 386.81 | 381.25 | -1.44% | -1437.40 |
| 2 | 2023-05-23 | 2024-08-06 | 383.20 | 477.25 | 24.54% | 24189.25 |
| 3 | 2024-12-05 | 2024-12-23 | 558.70 | 544.02 | -2.63% | -3224.24 |
| 4 | 2025-07-08 | 2026-04-10 | 571.03 | 626.59 | 9.73% | 11628.63 |
| 5 | 2026-05-28 | 2026-06-01 | 689.76 | 697.32 | 1.10% | 1437.52 |

### regime_sma200_1_3

| # | Entry | Exit | Entry Price | Exit Price | Return | PnL |
|---:|---|---|---:|---:|---:|---:|
| 1 | 2021-06-04 | 2022-06-03 | 386.81 | 379.08 | -2.00% | -1998.40 |
| 2 | 2023-05-17 | 2025-03-11 | 378.83 | 514.34 | 35.77% | 35055.82 |
| 3 | 2025-07-07 | 2026-06-01 | 573.00 | 697.32 | 21.70% | 28868.58 |

### prob55_sma200_2_3

| # | Entry | Exit | Entry Price | Exit Price | Return | PnL |
|---:|---|---|---:|---:|---:|---:|
| 1 | 2021-06-04 | 2022-06-06 | 386.81 | 381.25 | -1.44% | -1437.40 |
| 2 | 2023-05-17 | 2025-03-12 | 378.83 | 516.89 | 36.44% | 35919.95 |
| 3 | 2025-07-07 | 2026-06-01 | 573.00 | 697.32 | 21.70% | 29177.78 |

### prob55_trend50_2_3

| # | Entry | Exit | Entry Price | Exit Price | Return | PnL |
|---:|---|---|---:|---:|---:|---:|
| 1 | 2021-06-04 | 2022-06-06 | 386.81 | 381.25 | -1.44% | -1437.40 |
| 2 | 2023-05-17 | 2024-08-07 | 378.83 | 485.76 | 28.23% | 27820.66 |
| 3 | 2024-10-29 | 2025-01-14 | 532.95 | 537.29 | 0.82% | 1030.37 |
| 4 | 2025-07-07 | 2026-04-10 | 573.00 | 626.59 | 9.35% | 11916.40 |
| 5 | 2026-05-27 | 2026-06-01 | 690.38 | 697.32 | 1.01% | 1400.61 |
