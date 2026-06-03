# 2026-06-03T09:36:16Z Daily Ranker Walk-Forward

## Work Completed

- Added `research/ml/walkforward_daily_ranker.py`.
- Ran expanding-history annual OOS folds for 2021 through partial 2026.
- Wrote per-fold model artifacts, training reports, feature importances, and aggregate fold reports.

## Artifact

- Report: `reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_2021_2026/daily_ranker_walkforward.md`
- JSON: `reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_2021_2026/daily_ranker_walkforward.json`
- Fold models: `reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_2021_2026/fold_artifacts/`

## Champion

- Variant: `stocks_h21_s15_top3_reb42`
- Decision: `reject_weak_fold_repeatability`
- Compounded return: 131.44%
- VOO compounded return: 122.72%
- Excess: 8.72%
- Folds beating VOO: 3/6
- Validation-positive folds: 5/6
- Candidate folds: 1/6
- Mean turnover: 0.688

## Verdict

Rejected for fold repeatability. The ranker has some validation signal, but the active-sleeve portfolio does not convert it into reliable annual benchmark outperformance.

## Verification

- `cd backend && go test ./...`
- `cd backend && go build ./...`
- `research/ml/.venv/bin/python -m py_compile research/ml/*.py`
- `git diff --check`
