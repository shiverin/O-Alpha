# 2026-06-03T09:13:08Z ML Manifest Progress

## What changed

- `cmd/ml-meta-research` now supports `-feed`, `-adjustment`, and `-source` for export, compare, and inventory.
- `train_meta_label.py` now accepts `--export-manifest` and `--diagnostics-dir` and writes those into the artifact manifest.
- Exported a Yahoo-adjusted 100-symbol MA-signal panel from 2015-01-01 through 2024-12-31.
- Built live-core features and triple-barrier labels.
- Ran dataset diagnostics and trained a provenance-backed LightGBM artifact.

## Artifacts

- Export: `reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2024/`
- Dataset: `reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2024/dataset_live_core/`
- Diagnostics: `reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2024/dataset_live_core/diagnostics/`
- Training artifact: `reports/batches/2026-06-03_yahoo100_ml_manifest/train_live_core_artifact_v2/`

## Result

- Features: 251,600 rows, 40 live-core features, zero constant features.
- Labels: 2,665 long-side events across 100 symbols.
- Overlap: 2,665/2,665 labels matched feature rows.
- Model status: rejected.
- Rejection reason: `cpcv_auc_below_random`.
- Validation AUC: 0.5049.
- CPCV AUC mean: 0.4992.

## Verdict

Infrastructure improved, model rejected. The consulting-report warning is confirmed: sparse MA-event meta-labeling is not a reliable primary alpha engine here.

## Verification

- `go test ./...`
- `go build ./...`
- `research/ml/.venv/bin/python -m py_compile research/ml/*.py`
- `git diff --check`
