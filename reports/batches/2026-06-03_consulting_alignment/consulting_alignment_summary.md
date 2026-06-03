# Consulting Alignment Summary

Generated: 2026-06-03

## What The Reports Say Is Wrong

The current project is not failing because it lacks another indicator. It is misaligned:

- Research has been testing a benchmark-funded rotation allocator.
- Go runtime is mainly a single-symbol ML BUY filter with pass-through exits.
- Earlier rotation reports used close-to-close returns and no explicit cost model.
- Training artifacts did not preserve enough provenance to prove symbol universe, command, fold, cost, and feature assumptions.
- Several nominal features are not live in the current dataset and are silently zero-filled.
- The best ML result is still driven by one accepted alpha trade, so it is not promotable.

## Changes Implemented

1. Go runtime now applies artifact calibration before ML thresholding and sizing.
   - Raw probability is preserved as `p_success_raw`.
   - Calibrated probability is stored as `p_success`.

2. Go model metadata now supports an immutable `manifest` block.
   - It can record commands, symbols, context symbols, data snapshot, hashes, folds, status, benchmark, and cost model.

3. Research status semantics are now explicit.
   - Research guard accepts `candidate` and `promoted`.
   - Runtime registry still requires `promoted` plus leaf parity.

4. Go ML export no longer silently falls back to single-symbol export.
   - Omitting `--symbols` now fails unless `--allow-single-symbol-export` is passed.
   - Export writes `export_manifest.json`.

5. Python benchmark rotation now defaults to next-open execution with explicit costs.
   - Default cost model: 2 bps spread, 1 bp slippage.
   - `close_to_close` remains available only as an ablation.
   - Rotation and policy-search artifacts now write manifests.

6. Python rotation now applies metadata calibration before thresholding.

7. Dataset diagnostics script added.
   - Writes `feature_stats.csv`, `label_stats.csv`, and `overlap_report.json`.

8. Default label config is now deployment-aligned.
   - 21-bar horizon.
   - next-open entry.
   - 5-bar event spacing.

9. Live-core feature spec added.
   - Removes known-dead or non-live columns: HMM probabilities, Kalman metadata, order-book imbalance, and sector ETF fields.

## Dataset Diagnostics

Current full feature training dataset:

- Feature rows: 46,833
- Labeled events: 226
- Symbols with labels: 42
- Feature-label overlap: 226 / 226

Dead or non-live columns in the current full feature set:

- `hmm_regime_probability_low`
- `hmm_regime_probability_medium`
- `hmm_regime_probability_high`
- `kalman_residual`
- `kalman_zscore`
- `order_book_imbalance`
- `sector_etf_ret_1`
- `sector_etf_ret_5`

The live-core rebuilt dataset removes those dead columns and keeps 40 features.

## Aligned Retrain

Fresh artifact:

- Feature spec: `research/ml/feature_spec_live_core.yaml`
- Label config: `research/ml/label_config.yaml`
- Status: `candidate`
- Status reason: `passes_structural_training_gates`
- Train events: 174
- Validation events: 46
- Validation AUC: 0.558
- Validation log loss: 0.683
- CPCV AUC mean: 0.562
- Leaf parity: passed, max absolute error `5.55e-17`

This is structurally acceptable as a research candidate, but not production-promoted.

## Costed Next-Open Rotation Result

OOS period: 2025 to 2026 panel.

Fresh live-core artifact:

| Strategy | Return | Benchmark | Excess | Sharpe | Max DD | Trades | Costs |
|---|---:|---:|---:|---:|---:|---:|---:|
| ML rotation | 43.32% | 29.84% | 13.48% | 1.548 | 19.01% | 1 | $43.24 |

Trade:

- 2025-07-23: VOO -> JNJ
- 2026-06-01: end of test
- Trade return: 32.63%

Interpretation:

The aligned implementation still beats matched VOO on this OOS fold, but the entire excess remains concentrated in one JNJ trade. This confirms the consulting reports: it is a provisional research checkpoint, not a robust deployable alpha engine.

## Artifacts

- Dataset diagnostics: `reports/batches/2026-06-03_consulting_alignment/dataset_diagnostics_train_21_126/`
- Live-core train dataset: `reports/batches/2026-06-03_consulting_alignment/live_core_next_open/train_21_126/dataset/`
- Live-core artifact: `reports/batches/2026-06-03_consulting_alignment/live_core_next_open/train_21_126/artifact/`
- Live-core OOS dataset: `reports/batches/2026-06-03_consulting_alignment/live_core_next_open/oos_2025_2026_panel/dataset/`
- Live-core costed rotation: `reports/batches/2026-06-03_consulting_alignment/live_core_next_open/costed_rotation_oos_21_126/`
- Costed policy search: `reports/batches/2026-06-03_consulting_alignment/costed_rotation_policy_search/`

## Next Required Move

The correct next implementation is not another single-symbol filter tweak. It is a benchmark-core active sleeve or daily cross-sectional ranker:

- Keep VOO as the core allocation.
- Allocate only a capped active sleeve, for example 15% to 30%.
- Select top-k names by calibrated edge divided by risk.
- Add turnover bands and cost checks.
- Require breadth before promotion: multiple OOS folds, multiple alpha trades, and no single trade contributing more than 25% of cumulative excess.
