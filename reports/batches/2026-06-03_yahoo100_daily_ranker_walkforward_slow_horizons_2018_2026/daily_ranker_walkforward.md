# Daily Ranker Walk-Forward

- Benchmark: `VOO`
- Status: `python_prescreen_only`
- Promotion note: official alpha promotion still requires cmd/alpha-research DSR/PBO gate
- Folds: `9`
- Variants: `4`
- Champion: `stocks_h63_s15_top3_reb63_z10`
- Champion decision: `research_only_weak_validation`

| Variant | Decision | Return | Benchmark | Excess | Folds Beat | Validation Positive | Candidate Folds | Max DD | Benchmark Max DD | Turnover |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| stocks_h63_s15_top3_reb63_z10 | research_only_weak_validation | 281.95% | 225.07% | 56.88% | 9/9 | 6/9 | 3/9 | 33.99% | 33.99% | 0.410 |
| stocks_h126_s10_top3_reb63_z10 | reject_weak_fold_repeatability | 245.58% | 225.07% | 20.51% | 8/9 | 7/9 | 0/9 | 33.99% | 33.99% | 0.272 |
| stocks_h63_s10_top3_reb63_z10 | reject_weak_fold_repeatability | 257.62% | 225.07% | 32.55% | 8/9 | 6/9 | 2/9 | 33.99% | 33.99% | 0.272 |
| stocks_h126_s15_top3_reb63_z10 | reject_weak_fold_repeatability | 253.36% | 225.07% | 28.29% | 7/9 | 7/9 | 1/9 | 33.99% | 33.99% | 0.434 |
