# HMM Regime Model Implementation and Backtest Research Report

Date: 2026-06-02  
Branch: `feat/HMM_ensemble_decision`  
Primary asset evaluated: `VOO`  
Primary artifact directory: `reports/`

## Abstract

This report documents the end-to-end implementation and empirical evaluation of a hidden Markov model (HMM) regime layer for the O-Alpha ensemble trading worker. The PRD required three comparable decision modes: no HMM gating, calibrated HMM gating, and Baum-Welch fitted HMM gating. The implementation added a shared observation encoder, a flat 3-state by 9-symbol HMM representation, calibrated and fitted HMM model paths behind the same ensemble interface, walk-forward backtesting, fitted-model promotion gates, worker-parity backtests, and trade ledgers.

The main empirical conclusion is conservative: the Baum-Welch model should not be promoted. It improved validation likelihood on the hourly experiment, but it failed the trading promotion gates. On both daily and hourly VOO experiments, no-HMM gating produced better active-strategy trading metrics than the HMM-gated variants. In the worker-parity experiment, the current worker-like HMM path also underperformed the same worker logic with the HMM layer removed.

## Research Questions

1. Does Baum-Welch fitting improve the calibrated HMM out-of-sample?
2. Does any HMM regime layer improve trading performance versus no HMM gating?
3. Does the current live worker logic differ materially from the walk-forward experiments?
4. If the current worker logic is mirrored, does a worker without HMM, using only MA and Kalman, perform better?
5. Are the fitted HMMs numerically stable enough to be considered for promotion?

## PRD Requirements Interpreted

The PRD specified a safe comparison framework rather than immediate replacement of the calibrated HMM. The core requirements were:

- Implement calibrated HMM and Baum-Welch fitted HMM behind a shared interface.
- Use a shared observation encoder to prevent mismatched feature definitions.
- Represent emissions as a flat 9-symbol model: volatility bucket times trend bucket.
- Use Baum-Welch with scaled forward-backward inference, warm start from calibrated parameters, smoothing, sticky prior, state relabeling, and degeneration guards.
- Run all variants under identical walk-forward backtests.
- Promote Baum-Welch only if it wins out-of-sample on trading metrics, not only likelihood.
- Include a no-HMM baseline because regime gating itself may hurt performance.

## Implementation Inventory

| Area | Files |
|---|---|
| Observation encoding | `backend/internal/agent/hmm_observation_encoder.go` |
| Fitted and calibrated model struct | `backend/internal/agent/hmm_model.go` |
| Shared HMM inference | `backend/internal/agent/hmm_inference.go` |
| Baum-Welch trainer | `backend/internal/agent/hmm_baum_welch.go` |
| Regime detector integration | `backend/internal/agent/hmm_regime_detector.go`, `backend/internal/agent/hmm_calibration.go` |
| Ensemble mode selection | `backend/internal/agent/regime_mode.go`, `backend/internal/agent/ensemble_decision_layer.go` |
| Walk-forward comparison | `backend/internal/agent/hmm_regime_comparison.go` |
| Worker-parity comparison | `backend/internal/agent/worker_parity_backtest.go` |
| Backtest engine and metrics | `backend/internal/backtest/engine.go`, `backend/internal/backtest/metrics.go` |
| Backtest CLI | `backend/cmd/backtest/main.go` |
| Models and trade ledger schema | `backend/pkg/models/bar.go` |
| Data ingest force backfill | `backend/cmd/ingest/main.go`, `backend/internal/config/config.go`, `.env.example` |
| Tests | `backend/internal/agent/hmm_model_test.go`, `backend/internal/backtest/engine_test.go` |

## Model Specification

### Hidden States

The HMM uses three hidden states:

| State | Canonical regime |
|---:|---|
| 0 | Low Vol Trend |
| 1 | Medium |
| 2 | High Vol Stress |

For Baum-Welch, state labels are arbitrary during fitting. After fitting, states are relabeled by expected volatility so that state 0 remains lowest volatility, state 1 middle volatility, and state 2 highest volatility.

### Observations

Each rolling window is encoded into one of nine symbols:

`symbol = volBucket * 3 + trendBucket`

| Symbol | Volatility bucket | Trend bucket |
|---:|---|---|
| 0 | Low | Down |
| 1 | Low | Neutral |
| 2 | Low | Up |
| 3 | Medium | Down |
| 4 | Medium | Neutral |
| 5 | Medium | Up |
| 6 | High | Down |
| 7 | High | Neutral |
| 8 | High | Up |

The encoder computes:

- Realized volatility: square root of mean squared log returns over the rolling window.
- Rolling trend: `(lastClose - firstClose) / firstClose`.
- Bucket thresholds: fitted from training bars only to prevent future leakage.

Implementation note: the encoder stores 25th, 50th, and 75th percentile thresholds. The current discretization uses the first two volatility thresholds and the latter two trend thresholds. This is the implementation actually evaluated in this report.

### Calibrated HMM

The calibrated model uses hand-tuned priors:

`Pi = [0.33, 0.34, 0.33]`

Transition matrix:

| From / To | Low | Medium | Stress |
|---|---:|---:|---:|
| Low | 0.80 | 0.15 | 0.05 |
| Medium | 0.25 | 0.50 | 0.25 |
| Stress | 0.10 | 0.30 | 0.60 |

Volatility emission priors:

| State | Low vol | Medium vol | High vol |
|---|---:|---:|---:|
| Low | 0.70 | 0.25 | 0.05 |
| Medium | 0.20 | 0.60 | 0.20 |
| Stress | 0.05 | 0.25 | 0.70 |

Trend emission priors:

| State | Down | Neutral | Up |
|---|---:|---:|---:|
| Low | 0.10 | 0.20 | 0.70 |
| Medium | 0.33 | 0.34 | 0.33 |
| Stress | 0.60 | 0.30 | 0.10 |

The final flat emission matrix is computed as:

`B[state][volBucket * 3 + trendBucket] = volEmission[state][volBucket] * trendEmission[state][trendBucket]`

Rows are normalized after construction.

### Baum-Welch HMM

The fitted model uses the same `Pi`, `A`, and `B` shape as the calibrated model:

- Number of states: 3
- Number of symbols: 9
- Initialization: calibrated model warm start
- Max iterations: 50
- Tolerance: `1e-4`
- Emission smoothing: `1e-3`
- Transition smoothing: `1e-3`
- Minimum state occupancy: `0.02`
- Sticky prior: enabled
- Sticky strength: `0.05`

The trainer uses scaled forward-backward to avoid underflow:

`alpha_t(j) = B_j(o_t) * sum_i alpha_{t-1}(i) A_{ij}`

Each alpha row is scaled. The log likelihood is recovered from the scale factors. The backward pass uses matching scaling, then gamma and xi statistics are computed for the M-step.

The M-step applies smoothed updates:

`A_ij = (xiSum_ij + transitionSmoothing * priorA_ij + stickyPrior_ij) / denominator`

`B_ik = (emissionCount_ik + emissionSmoothing * priorB_ik) / denominator`

Rows are normalized after each update.

### Degenerate Fit Guards

A fitted model is rejected if:

- Any state occupancy is below the configured minimum.
- Any emission row collapses to one symbol with probability greater than 0.97.
- Any probability row is invalid or non-normalized.
- Final training log likelihood decreases from the calibrated initialization.
- Validation log likelihood is lower than the calibrated baseline when validation observations exist.

Rejected Baum-Welch folds fall back to the calibrated HMM strategy for trading.

## Ensemble Decision Layer

The ensemble combines:

- MA crossover: default fast period 20, slow period 50.
- Kalman mean reversion: `q=0.001`, `r=0.01`, lookback 20, z-threshold 2.0.
- Optional HMM regime detector.

Voting:

| Regime | MA weight | Kalman weight |
|---|---:|---:|
| Low Vol Trend | 0.70 | 0.30 |
| Medium | 0.50 | 0.50 |
| High Vol Stress | 0.20 | 0.80 |

Signal thresholds:

- Buy if weighted score is at least `0.5`.
- Sell if weighted score is at most `-0.5`.
- In high-volatility stress, buys are suppressed and only sell exits are allowed.

Position sizing:

| Risk profile | Base size |
|---|---:|
| Conservative | 5 percent |
| Moderate | 10 percent |
| Aggressive | 20 percent |

Regime multipliers:

| Regime | Multiplier |
|---|---:|
| Low Vol Trend | 1.00 |
| Medium | 0.75 |
| High Vol Stress | 0.25 |

The experiment used the moderate profile.

## Backtest Harness

### Walk-Forward Harness

The PRD comparison uses `RunWalkForwardRegimeComparison`.

For each fold:

1. Split bars into a train segment and a test segment.
2. Fit encoder buckets on train only.
3. Build one of:
   - `none`: MA plus Kalman, no HMM regime gating.
   - `calibrated`: calibrated HMM with train-fitted buckets.
   - `baumwelch`: train-fitted Baum-Welch HMM, with calibrated fallback on rejection.
4. Generate signals over train plus test history.
5. Keep only test outputs for out-of-sample scoring.
6. Concatenate all test folds into one out-of-sample active-strategy backtest.
7. Compare against buy-and-hold on the same out-of-sample bars.

Execution model:

- Walk-forward strategy signals generated at bar `t` execute at bar `t+1` open.
- Open active positions are liquidated at final bar close.
- Initial cash: 100,000.
- Metrics include total return, annualized return, Sharpe, Sortino, Calmar, maximum drawdown, trade count, profit factor, win rate, exposure, turnover, and trade ledger.

### Worker-Parity Harness

The worker-parity simulator mirrors the live worker more closely:

- Initial warmup history: 51 bars.
- Rolling history max: 10,000 bars.
- Recalibration cadence: every 500 bars.
- Signal evaluated using `EvaluateLatest`.
- Buy execution at current bar close using `AvailableCash * PositionSizePct`.
- Sell execution at current bar close for half the current position.
- If the half-sell size is below 1 share, exit the remaining position.
- No forced final liquidation, matching worker mark-to-market behavior.
- Closed trade ledgers use FIFO lot matching because worker scale-in and half-exit behavior can close fractional pieces of multiple entries.

Two worker modes were compared:

- `worker_calibrated`: current worker-like MA plus Kalman plus calibrated HMM.
- `worker_none`: same worker-like execution with MA plus Kalman only.

## Data and Ingestion

The user provided Alpaca credentials, then VOO bars were ingested into the configured database. The ingest path was extended with `INGEST_FORCE_BACKFILL` so data could be backfilled even when recent bars already existed.

Relevant environment variables:

- `INGEST_SYMBOLS=VOO`
- `INGEST_LOOKBACK=...`
- `INGEST_FORCE_BACKFILL=true`

The ingest command path is:

`make run-ingest`

Final VOO datasets used:

| Timeframe | Bars | First bar | Last bar |
|---|---:|---|---|
| 1Day | 1,469 | 2020-07-27 12:00:00 +08 | 2026-06-01 12:00:00 +08 |
| 1Hour | 10,530 | 2020-07-27 21:00:00 +08 | 2026-06-02 04:00:00 +08 |

## Experiment Configuration

Daily run:

```bash
cd backend
set -a; source ../.env.local; set +a
go run ./cmd/backtest \
  --symbol VOO \
  --timeframe 1Day \
  --from 2020-06-01 \
  --to 2026-06-02 \
  --regime-modes none,calibrated,baumwelch \
  --train-bars 1260 \
  --test-bars 21 \
  --step-bars 21 \
  --min-trades 3 \
  --worker-warmup-bars 51 \
  --output ../reports/voo_hmm_regime_comparison_5ytrain_1day.json \
  --trades-output-dir ../reports/trades
```

Hourly run:

```bash
cd backend
set -a; source ../.env.local; set +a
go run ./cmd/backtest \
  --symbol VOO \
  --timeframe 1Hour \
  --from 2020-06-01 \
  --to 2026-06-02 \
  --regime-modes none,calibrated,baumwelch \
  --train-bars 8820 \
  --test-bars 147 \
  --step-bars 147 \
  --min-trades 10 \
  --worker-warmup-bars 51 \
  --output ../reports/voo_hmm_regime_comparison_5ytrain_1hour.json \
  --trades-output-dir ../reports/trades
```

Artifacts:

- `reports/voo_hmm_regime_comparison_5ytrain_1day.json`
- `reports/voo_hmm_regime_comparison_5ytrain_1hour.json`
- `reports/trades/*.csv`

## Test Coverage

Implemented tests include:

- Encoder produces valid flat 9-symbol observations.
- Calibrated model emits normalized flat rows.
- Baum-Welch fitting improves likelihood on synthetic clustered observations.
- Baum-Welch relabels states by expected volatility.
- Fitted model JSON persistence round-trips.
- Walk-forward comparison runs all modes.
- Worker-parity backtests run HMM and no-HMM modes.
- Backtest engine preserves idle cash.
- Buy-and-hold uses full allocation.
- Trade ledger records expected entry and exit prices.

Validation command:

```bash
cd backend
go test ./...
```

Status: passing as of 2026-06-02.

## Model Diagnostics

### Daily Baum-Welch Diagnostics

| Item | Value |
|---|---:|
| Folds | 9 |
| Rejected fitted folds | 0 |
| Converged folds | 8 |
| Average iterations | 24.44 |
| Min iterations | 16 |
| Max iterations | 50 |
| Average training log likelihood | -1402.77 |
| Min state occupancy | [0.242, 0.223, 0.428] |
| Average state occupancy | [0.257, 0.275, 0.468] |
| Max state occupancy | [0.338, 0.322, 0.501] |

Daily validation likelihood caveat: the daily test fold length was 21 bars, while the HMM observation window was 50 bars. Therefore the daily validation observation sequence was empty and validation likelihood was not informative for this configuration.

### Hourly Baum-Welch Diagnostics

| Item | Value |
|---|---:|
| Folds | 11 |
| Rejected fitted folds | 0 |
| Converged folds | 6 |
| Average iterations | 44.73 |
| Min iterations | 30 |
| Max iterations | 50 |
| Average training log likelihood | -10051.43 |
| Average fitted validation log likelihood | -116.80 |
| Average calibrated validation log likelihood | -192.20 |
| Min state occupancy | [0.309, 0.329, 0.312] |
| Average state occupancy | [0.314, 0.348, 0.338] |
| Max state occupancy | [0.319, 0.374, 0.359] |

Interpretation: hourly Baum-Welch produced stable, non-collapsed state occupancies and improved validation likelihood relative to calibrated HMM. However, the PRD promotion rule is based on out-of-sample trading performance, not likelihood alone.

## Walk-Forward Trading Results

### Daily 5-Year-Train Walk-Forward

| Mode | Total return | Annualized return | Sharpe | Sortino | Calmar | Max drawdown | Trades | Exposure | Rejected |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Baum-Welch HMM | 0.60% | 0.80% | 1.200 | 0.794 | 1.400 | 0.57% | 4 | 34.92% | 0 |
| Calibrated HMM | 0.60% | 0.80% | 1.200 | 0.794 | 1.400 | 0.57% | 4 | 34.92% | 0 |
| No HMM | 1.51% | 2.03% | 2.121 | 1.591 | 2.654 | 0.76% | 8 | 44.44% | 0 |
| Buy-and-hold | 15.19% | 20.88% | 1.581 | 1.555 | 2.281 | 9.15% | 2 | 99.47% | n/a |

Daily promotion decision:

- Promote Baum-Welch: false.
- Reason: Sharpe improvement below 10 percent.
- Additional issue: Baum-Welch did not beat the no-HMM baseline.

### Hourly 5-Year-Train Walk-Forward

| Mode | Total return | Annualized return | Sharpe | Sortino | Calmar | Max drawdown | Trades | Exposure | Rejected |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Baum-Welch HMM | 0.55% | 0.09% | 0.366 | 0.212 | 0.239 | 0.36% | 26 | 36.92% | 0 |
| Calibrated HMM | 0.52% | 0.08% | 0.345 | 0.202 | 0.226 | 0.36% | 28 | 37.66% | 0 |
| No HMM | 1.76% | 0.27% | 0.715 | 0.543 | 0.802 | 0.34% | 60 | 51.08% | 0 |
| Buy-and-hold | 19.00% | 2.75% | 0.593 | 0.578 | 0.292 | 9.43% | 2 | 99.94% | n/a |

Hourly promotion decision:

- Promote Baum-Welch: false.
- Sharpe improvement over calibrated: 5.84 percent.
- Required Sharpe improvement: at least 10 percent.
- Additional issue: Baum-Welch did not beat the no-HMM baseline.

## Worker-Parity Trading Results

The live worker path is not identical to the walk-forward engine. The worker evaluates the latest rolling history, executes at the current close, buys with available cash times position size, and sells half the current position. Therefore a separate worker-parity experiment was run.

### Daily Worker-Parity

| Mode | Total return | Annualized return | Sharpe | Sortino | Calmar | Max drawdown | Fill count | Closed trades | Exposure |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Current worker HMM | 6.43% | 1.11% | 0.605 | 0.477 | 0.243 | 4.58% | 36 | 33 | 59.66% |
| Worker no-HMM | 31.07% | 4.93% | 0.975 | 1.021 | 0.559 | 8.82% | 134 | 129 | 95.56% |
| Worker buy-and-hold | 122.76% | 15.31% | 0.944 | 0.931 | 0.605 | 25.32% | 1 | 0 | 100.00% |

### Hourly Worker-Parity

| Mode | Total return | Annualized return | Sharpe | Sortino | Calmar | Max drawdown | Fill count | Closed trades | Exposure |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Current worker HMM | 8.40% | 0.19% | 0.261 | 0.212 | 0.075 | 2.59% | 341 | 328 | 69.89% |
| Worker no-HMM | 36.68% | 0.75% | 0.378 | 0.385 | 0.089 | 8.49% | 1051 | 1024 | 92.61% |
| Worker buy-and-hold | 128.61% | 2.01% | 0.351 | 0.341 | 0.075 | 26.88% | 1 | 0 | 100.00% |

Interpretation:

- The no-HMM worker variant outperformed the current HMM worker variant on both daily and hourly total return and Sharpe.
- The HMM layer reduced exposure and drawdown but also reduced participation in a strongly rising asset.
- Buy-and-hold still dominated total return, with much larger drawdown.

## Trade Ledgers

Trade ledgers were added to `BacktestResult` and exported to CSV. Each row contains:

`symbol,timeframe,mode,entry_time,exit_time,entry_price,exit_price,quantity,entry_value,exit_value,pnl,return_pct`

Key ledgers:

| Ledger | Closed trades |
|---|---:|
| `reports/trades/voo_1day_worker_none_trades.csv` | 129 |
| `reports/trades/voo_1day_worker_calibrated_trades.csv` | 33 |
| `reports/trades/voo_1hour_worker_none_trades.csv` | 1024 |
| `reports/trades/voo_1hour_worker_calibrated_trades.csv` | 328 |
| `reports/trades/voo_1day_none_trades.csv` | 4 |
| `reports/trades/voo_1day_calibrated_trades.csv` | 2 |
| `reports/trades/voo_1day_baumwelch_trades.csv` | 2 |
| `reports/trades/voo_1hour_none_trades.csv` | 30 |
| `reports/trades/voo_1hour_calibrated_trades.csv` | 14 |
| `reports/trades/voo_1hour_baumwelch_trades.csv` | 13 |

Worker ledgers use FIFO closed round trips. This is necessary because the worker can scale into positions and later sell half the aggregate position. One sell fill can therefore close pieces of multiple earlier entries. Open residual positions are not listed as closed trades because they do not have an exit price.

## Final Decision Against PRD Gates

| Gate | Requirement | Daily result | Hourly result | Pass |
|---|---|---|---|---|
| Sharpe improvement | Baum-Welch at least 10% above calibrated | 0.00% | 5.84% | No |
| Drawdown control | Max drawdown increase no more than 5% | 0.00% | 0.00% | Yes |
| Trade count | Statistically meaningful minimum | 4 vs daily min 3 | 26 vs hourly min 10 | Yes |
| Out-of-sample validation | Better held-out likelihood where available | Not available under 21-bar daily test fold | Improved vs calibrated | Partial |
| State stability | No collapsed state after relabeling | Stable | Stable | Yes |
| No-HMM baseline | Baum-Welch should beat no-HMM | No | No | No |

Decision: do not promote Baum-Welch.

Secondary decision: no-HMM MA plus Kalman should be treated as a serious candidate for the worker path, because it outperformed HMM-gated worker parity in this experiment.

## Limitations and Caveats

1. The experiment evaluated VOO only. Results may not transfer to other ETFs, individual equities, or non-US market data.
2. VOO was in a long upward regime over much of the sample. HMM stress gating may look worse in such periods because it reduces participation.
3. Hourly annualized metrics currently use the same `252` period annualization convention as daily metrics. Total return, drawdown, trade count, and exposure are still directly interpretable, but hourly annualized return, Sharpe, Sortino, and Calmar should be treated as convention-dependent until timeframe-aware annualization is added.
4. Daily validation likelihood was unavailable because the 21-bar test fold is shorter than the 50-bar observation window.
5. Worker-parity uses close-price execution and no final liquidation, matching the current worker more closely but not matching the walk-forward engine.
6. Transaction costs, slippage, tax effects, and borrow constraints were not modeled.
7. The current encoder threshold usage should be reviewed. It stores three percentile thresholds but uses an asymmetric subset for volatility and trend discretization.

## Recommended Next Steps

1. Do not replace calibrated HMM with Baum-Welch in production.
2. Add timeframe-aware metric annualization before relying on hourly annualized statistics.
3. Run the same framework across SPY, IVV, QQQ, sector ETFs, and stress-heavy historical windows.
4. Run a worker no-HMM paper-trading shadow mode against current worker HMM.
5. Add transaction cost and slippage assumptions to the worker-parity simulator.
6. Consider a longer daily test fold, at least 50 bars, if daily validation likelihood is required.
7. Review the observation encoder bucket threshold convention and decide whether the asymmetric discretization is intentional.

## Bottom Line

The implementation satisfies the PRD: calibrated HMM, Baum-Welch HMM, and no-HMM modes are implemented under a shared interface, tested, backtested, compared, and gated by promotion logic. Baum-Welch is numerically viable and improves hourly validation likelihood, but it does not improve trading performance enough to pass promotion. The strongest active strategy result in these experiments is the no-HMM MA plus Kalman variant, and the strongest absolute return remains buy-and-hold.
