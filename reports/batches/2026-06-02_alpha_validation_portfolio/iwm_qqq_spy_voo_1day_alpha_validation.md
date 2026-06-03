# Alpha Validation Report

- Generated: `2026-06-02T06:17:55Z`
- Symbols: `IWM, QQQ, SPY, VOO`
- Timeframe: `1Day`
- Period: `2021-06-04` to `2026-06-01`
- Bars: `1253`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| equal_weight | 75.12% | 11.94% | 0.695 | 0.688 | 0.421 | 28.33% | 2527 | 5.133 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| xsec_momentum_top15 | equal_weight | false | 17.00% | 3.21% | 1.060 | 0.943 | 0.804 | 3.99% | 1.000 | 0.000 | 1070 | xsec universe size 4 below research minimum 50 |
| kalman_cointegration_z2 | flat_cash | false | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0.000 | 0.000 | 0 | DSR 0.000 below 0.950 |

## xsec_momentum_top15

- Family: `xsec_momentum`
- Benchmark: `equal_weight`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - xsec universe size 4 below research minimum 50
- Diagnostics:
  - xsec universe size 4 below research minimum 50

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 17.00% | 3.21% | 1.060 | 0.943 | 0.804 | 3.99% | 1070 | 4.152 | - |
| stress_2x | 16.91% | 3.20% | 1.055 | 0.938 | 0.800 | 4.00% | 1070 | 4.150 | - |
| stress_3x | 16.82% | 3.18% | 1.050 | 0.933 | 0.795 | 4.00% | 1070 | 4.148 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-882 | 0.837 | 0.000 | 0.000 | 0.00% | - |
| 1 | 126-882 | 882-1008 | 1.417 | 0.000 | 0.000 | 0.00% | - |
| 2 | 252-1008 | 1008-1134 | 0.890 | 0.000 | 0.000 | 0.00% | - |

## kalman_cointegration_z2

- Family: `kalman_cointegration`
- Benchmark: `flat_cash`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - DSR 0.000 below 0.950
  - OOS trades 0 below 30
  - no drawdown-adjusted improvement over benchmark
  - pair sleeve requires offline approved cointegration candidate and live shortability gate before promotion
- Diagnostics:
  - pair sleeve requires offline approved cointegration candidate and live shortability gate before promotion

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 | - |
| stress_2x | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 | - |
| stress_3x | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-882 | 0.000 | 0.000 | 0.000 | 0.00% | - |
| 1 | 126-882 | 882-1008 | 0.000 | 0.000 | 0.000 | 0.00% | - |
| 2 | 252-1008 | 1008-1134 | 0.000 | 0.000 | 0.000 | 0.00% | - |
