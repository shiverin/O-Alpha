# Consulting Alignment Continuation

Generated: 2026-06-03

## Objective

Continue the consulting-report execution after the initial ML meta-label alignment work. The reports' main correction is to stop treating the model as an all-in single-name switch and instead run a VOO benchmark core with a capped active sleeve, broad selection, next-open execution, explicit costs, and strict fold-level promotion gates.

## Implemented Research Tools

1. `research/ml/portfolio_sleeve.py`
   - Shared next-open target-weight simulator.
   - Applies explicit spread/slippage/commission/SEC-fee cost assumptions.
   - Tracks equity, benchmark equity, active weight, orders, turnover, and drawdown.

2. `research/ml/backtest_meta_label_sleeve.py`
   - Reuses the existing LightGBM meta-label artifact.
   - Keeps VOO as the core and allocates only a capped active sleeve to accepted events.

3. `research/ml/backtest_daily_ranker_sleeve.py`
   - Trains a daily LightGBM cross-sectional ranker.
   - Uses purged chronological validation.
   - Backtests top-k active-sleeve allocation.

4. `research/ml/backtest_momentum_sleeve.py`
   - Deterministic cross-sectional relative-momentum sleeve.
   - Supports lookback, universe, threshold, and allocation-mode variants.

5. `research/ml/momentum_sleeve_search.py`
   - Grid-searches momentum sleeve variants across 2023, 2024, and 2025-2026 walk-forward folds.

6. `research/ml/backtest_composite_momentum_sleeve.py`
   - Combines multiple momentum legs under one benchmark-core active sleeve.
   - Supports point-in-time volatility caps per leg.

## Rejected Or Research-Only Results

| Experiment | Result | Decision |
|---|---:|---|
| Existing ML meta-label as 30% active sleeve, 2025-2026 | 31.75% vs VOO 29.84%, excess 1.91% | Better architecture, weak edge |
| Daily LightGBM ranker, all universe, 21-bar horizon, 2025-2026 | 27.72% vs VOO 29.84%, excess -2.12% | Rejected |
| Best single momentum sleeve search checkpoint | 120.26% compounded vs VOO 101.08%, excess 19.19%, beats 2/3 folds | Research-only, not all-fold |

The daily ranker did not find a robust selection edge on the 42-name panel. The strongest simple momentum sleeve beat compounded VOO but missed the 2023 fold by 6 bps after costs, so it was not promoted.

## Current Research Checkpoint

Composite benchmark-core momentum sleeve:

- Core: VOO residual allocation.
- Total active sleeve: 29%.
- Rebalance: every 21 daily bars.
- Execution: next open after close-time signal.
- Costs: 2 bps spread, 1 bp slippage, no commissions.

Leg 1:

- Universe: ETFs only, excluding VOO/SPY benchmark proxies.
- Signal: 21-bar relative log momentum vs VOO.
- Minimum relative momentum: 5%.
- Volatility cap: 20-day annualized vol must be <= 25%.
- Allocation: top 1 ETF, max 24% weight.

Leg 2:

- Universe: all non-benchmark candidates.
- Signal: 126-bar relative log momentum vs VOO.
- Minimum relative momentum: 10%.
- Allocation: top 5 names, 5% sleeve total, max 1% per name.

## Walk-Forward Results

| Fold | Strategy | VOO | Excess | Sharpe | VOO Sharpe | Max DD | VOO Max DD | Turnover | Costs | Alpha Symbols |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 2023 | 25.12% | 24.81% | 0.31% | 1.909 | 1.928 | 9.77% | 10.35% | 3.509 | $70.18 | 18 |
| 2024 | 25.56% | 24.08% | 1.48% | 1.888 | 1.916 | 10.34% | 8.34% | 3.145 | $62.91 | 18 |
| 2025-2026 | 34.48% | 29.84% | 4.64% | 1.421 | 1.193 | 18.30% | 19.01% | 4.180 | $83.59 | 24 |

Compounded across folds:

- Strategy return: 111.27%
- VOO return: 101.08%
- Excess return: 10.19%
- Max fold drawdown: 18.30%
- Benchmark max fold drawdown: 19.01%
- Mean turnover: 3.611
- Total explicit costs: $216.68
- Minimum alpha symbols per fold: 18

## Decision

This is the first clean research checkpoint that beats VOO in every tested fold after costs while preserving broad participation and staying inside the 30% active-sleeve constraint.

It should still be treated as a research candidate, not production-promoted alpha, because it was selected after searching the same walk-forward folds. The next promotion step should be a locked-rule test on fresh unseen data, a larger universe, and broker-realistic live/paper execution telemetry.

## Artifacts

- Meta-label sleeve: `reports/batches/2026-06-03_consulting_alignment/meta_label_active_sleeve_oos_21_126/`
- Daily ranker attempts: `reports/batches/2026-06-03_consulting_alignment/daily_ranker_sleeve_oos_21_126/`
- Momentum search: `reports/batches/2026-06-03_consulting_alignment/momentum_sleeve_search/`
- Threshold momentum search: `reports/batches/2026-06-03_consulting_alignment/momentum_sleeve_threshold_search/`
- Composite checkpoint: `reports/batches/2026-06-03_consulting_alignment/composite_momentum_sleeve/etf24v25_all126_5/`
