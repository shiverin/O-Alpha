# Alpha Validation Report

- Generated: `2026-06-03T08:32:35Z`
- Symbols: `VOO, AAPL, SPY`
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
| equal_weight | 626.27% | 19.03% | 0.914 | 0.887 | 0.584 | 32.61% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| ma_crossover_20_50 | buy_hold | false | 17.58% | 1.43% | 0.822 | 0.764 | 0.368 | 3.89% | 1.000 | 0.375 | 1582 | PBO 0.375 above 0.200 |

## ma_crossover_20_50

- Family: `ma_crossover`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.375 above 0.200
  - turnover increases without return improvement

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 17.58% | 1.43% | 0.822 | 0.764 | 0.368 | 3.89% | 1582 | 7.587 | - |
| stress_2x | 17.42% | 1.42% | 0.815 | 0.757 | 0.365 | 3.90% | 1582 | 7.582 | - |
| stress_3x | 17.25% | 1.41% | 0.807 | 0.750 | 0.361 | 3.90% | 1582 | 7.576 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.028 | -0.448 | -0.373 | 2.09% | - |
| 1 | 252-1008 | 1008-1260 | 0.672 | 2.409 | 4.657 | 0.62% | - |
| 2 | 504-1260 | 1260-1512 | 0.985 | 0.535 | 0.459 | 3.89% | - |
| 3 | 756-1512 | 1512-1764 | 0.610 | 2.196 | 5.549 | 0.51% | - |
| 4 | 1008-1764 | 1764-2016 | 0.969 | -0.754 | -0.695 | 2.65% | - |
| 5 | 1260-2016 | 2016-2268 | 0.909 | 1.651 | 1.840 | 1.18% | - |
| 6 | 1512-2268 | 2268-2520 | 0.493 | 1.889 | 2.759 | 0.86% | - |
| 7 | 1764-2520 | 2520-2772 | 0.703 | 1.108 | 1.110 | 1.83% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | ma_crossover_20_50 | 0.865 | -0.373 | 2 | ma_crossover_50_100 | -0.280 | 3 | false |
| 1 | ma_crossover_50_100 | 0.449 | 4.054 | 2 | ma_crossover_20_50 | 4.657 | 3 | false |
| 2 | ma_crossover_50_100 | 0.709 | 0.425 | 2 | ma_crossover_20_50 | 0.459 | 3 | false |
| 3 | ma_crossover_50_100 | 0.375 | 5.376 | 2 | ma_crossover_20_50 | 5.549 | 3 | false |
| 4 | ma_crossover_50_100 | 0.561 | -0.630 | 1 | ma_crossover_50_100 | -0.630 | 3 | false |
| 5 | ma_crossover_20_50 | 0.715 | 1.840 | 3 | ma_crossover_10_30 | 2.639 | 3 | true |
| 6 | ma_crossover_10_30 | 0.404 | 2.757 | 3 | ma_crossover_50_100 | 2.786 | 3 | true |
| 7 | ma_crossover_50_100 | 0.850 | 1.034 | 3 | ma_crossover_10_30 | 1.473 | 3 | true |
