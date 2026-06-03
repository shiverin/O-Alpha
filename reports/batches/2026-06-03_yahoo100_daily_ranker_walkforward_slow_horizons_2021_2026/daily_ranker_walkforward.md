# Daily Ranker Walk-Forward

- Benchmark: `VOO`
- Status: `python_prescreen_only`
- Promotion note: official alpha promotion still requires cmd/alpha-research DSR/PBO gate
- Folds: `6`
- Variants: `7`
- Champion: `stocks_h63_s15_top3_reb63_z10`
- Champion decision: `research_only_weak_validation`

| Variant | Decision | Return | Benchmark | Excess | Folds Beat | Validation Positive | Candidate Folds | Max DD | Benchmark Max DD | Turnover |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| stocks_h63_s15_top3_reb63_z10 | research_only_weak_validation | 155.23% | 122.72% | 32.50% | 6/6 | 4/6 | 2/6 | 23.65% | 24.52% | 0.389 |
| stocks_h63_s10_top3_reb63_z10 | research_only_weak_validation | 143.58% | 122.72% | 20.85% | 6/6 | 4/6 | 2/6 | 23.69% | 24.52% | 0.260 |
| stocks_h126_s15_top3_reb63_z10 | reject_weak_fold_repeatability | 137.37% | 122.72% | 14.65% | 5/6 | 5/6 | 1/6 | 25.95% | 24.52% | 0.422 |
| stocks_h126_s10_top3_reb63_z10 | reject_weak_fold_repeatability | 133.08% | 122.72% | 10.36% | 5/6 | 5/6 | 0/6 | 25.47% | 24.52% | 0.259 |
| stocks_h21_s15_top3_reb42_z10 | reject_weak_fold_repeatability | 133.51% | 122.72% | 10.79% | 3/6 | 5/6 | 1/6 | 25.23% | 24.52% | 0.680 |
| stocks_h42_s15_top3_reb42_z10 | reject_weak_fold_repeatability | 132.52% | 122.72% | 9.79% | 3/6 | 4/6 | 1/6 | 25.48% | 24.52% | 0.626 |
| stocks_h42_s10_top3_reb42_z10 | reject_under_benchmark | 122.00% | 122.72% | -0.72% | 2/6 | 4/6 | 0/6 | 25.36% | 24.52% | 0.374 |
