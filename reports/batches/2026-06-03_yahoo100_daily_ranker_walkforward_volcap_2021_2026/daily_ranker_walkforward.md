# Daily Ranker Walk-Forward

- Benchmark: `VOO`
- Status: `python_prescreen_only`
- Promotion note: official alpha promotion still requires cmd/alpha-research DSR/PBO gate
- Folds: `6`
- Variants: `4`
- Champion: `stocks_h21_s15_top3_reb42_z10`
- Champion decision: `reject_weak_fold_repeatability`

| Variant | Decision | Return | Benchmark | Excess | Folds Beat | Validation Positive | Candidate Folds | Max DD | Benchmark Max DD | Turnover |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| stocks_h21_s15_top3_reb42_z10 | reject_weak_fold_repeatability | 133.51% | 122.72% | 10.79% | 3/6 | 5/6 | 1/6 | 25.23% | 24.52% | 0.680 |
| stocks_h21_s15_top3_reb42_z10_vol35 | reject_weak_fold_repeatability | 125.98% | 122.72% | 3.26% | 3/6 | 5/6 | 0/6 | 24.80% | 24.52% | 0.738 |
| stocks_h21_s15_top3_reb42_z10_vol35_riskoff | reject_weak_fold_repeatability | 123.21% | 122.72% | 0.48% | 2/6 | 5/6 | 0/6 | 24.52% | 24.52% | 0.582 |
| stocks_h21_s15_top3_reb42_z10_vol30 | reject_under_benchmark | 120.94% | 122.72% | -1.79% | 2/6 | 5/6 | 0/6 | 24.85% | 24.52% | 0.669 |
