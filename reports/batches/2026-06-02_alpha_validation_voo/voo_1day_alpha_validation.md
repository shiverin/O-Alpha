# Alpha Validation Report

- Generated: `2026-06-02T06:15:55Z`
- Symbols: `VOO`
- Timeframe: `1Day`
- Period: `2021-06-02` to `2026-06-01`
- Bars: `1255`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 81.44% | 12.72% | 0.799 | 0.786 | 0.502 | 25.32% | 3 | 1.000 |
| equal_weight | 81.44% | 12.72% | 0.799 | 0.786 | 0.502 | 25.32% | 3 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| ma_crossover_20_50 | buy_hold | false | 6.09% | 1.20% | 0.724 | 0.697 | 0.433 | 2.76% | 1.000 | 0.333 | 648 | PBO 0.333 above 0.200 |
| kalman_z2 | buy_hold | false | 6.50% | 1.27% | 0.781 | 0.727 | 0.511 | 2.49% | 1.000 | 0.333 | 604 | PBO 0.333 above 0.200 |

## ma_crossover_20_50

- Family: `ma_crossover`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.333 above 0.200
  - turnover increases without return improvement

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 6.09% | 1.20% | 0.724 | 0.697 | 0.433 | 2.76% | 648 | 3.359 | - |
| stress_2x | 6.02% | 1.18% | 0.716 | 0.689 | 0.426 | 2.77% | 648 | 3.358 | - |
| stress_3x | 5.95% | 1.17% | 0.708 | 0.681 | 0.419 | 2.79% | 648 | 3.356 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-882 | 0.399 | 1.983 | 4.356 | 0.42% | - |
| 1 | 126-882 | 882-1008 | 0.622 | -0.001 | -0.015 | 1.87% | - |
| 2 | 252-1008 | 1008-1134 | 0.727 | 1.549 | 2.794 | 0.51% | - |

## kalman_z2

- Family: `kalman`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.333 above 0.200
  - turnover increases without return improvement

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 6.50% | 1.27% | 0.781 | 0.727 | 0.511 | 2.49% | 604 | 14.109 | - |
| stress_2x | 6.21% | 1.22% | 0.747 | 0.696 | 0.488 | 2.49% | 604 | 14.089 | - |
| stress_3x | 5.91% | 1.16% | 0.713 | 0.665 | 0.465 | 2.50% | 604 | 14.070 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-882 | 0.601 | 1.834 | 3.206 | 0.74% | - |
| 1 | 126-882 | 882-1008 | 0.670 | -0.001 | -0.016 | 2.01% | - |
| 2 | 252-1008 | 1008-1134 | 0.876 | 2.401 | 5.680 | 0.46% | - |
