# 2026-06-03T09:18:35Z Feature Parity Progress

## Work Completed

- Added deterministic Python fixture export: `research/ml/export_feature_fixture.py`.
- Added Go runtime parity validator: `backend/cmd/validate-feature-parity`.
- Added unit coverage for signal parsing, feature-spec YAML loading, and UTC key normalization.
- Ran parity on 500 labeled events from the Yahoo100 live-core dataset.

## Artifact

- Report: `reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2024/dataset_live_core/feature_parity_report.json`
- Fixture: `reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2024/dataset_live_core/feature_fixture.csv`
- Manifest: `reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2024/dataset_live_core/feature_fixture.manifest.json`

## Result

- Status: `passed`
- Rows checked: 500
- Feature count: 40
- Max absolute error: `5.810907310888069e-13`
- Failures: 0
- Missing rows: 0

## Verification

- `cd backend && go test ./...`
- `cd backend && go build ./...`
- `research/ml/.venv/bin/python -m py_compile research/ml/*.py`
- `git diff --check`

## Next

- Add costed next-open active-sleeve comparison for Python research allocation tools.
- Move from sparse MA-event meta-labels to a daily ranker or direct-weight active-sleeve target after execution alignment.
