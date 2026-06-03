# Costed Active-Sleeve Comparison

- Benchmark: `VOO`
- Status: `python_prescreen_only`
- Promotion note: official alpha promotion still requires cmd/alpha-research DSR/PBO gate
- Variant count: `8`
- Champion: `benchmark_ranked_sleeve_conservative`
- Champion decision: `reject_weak_fold_repeatability`

| Variant | Decision | Return | Benchmark | Excess | Folds Beating | Max DD | Benchmark Max DD | Mean Turnover |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| benchmark_ranked_sleeve_conservative | reject_weak_fold_repeatability | 330.02% | 293.11% | 36.91% | 6/10 | 33.30% | 33.99% | 0.936 |
| benchmark_ranked_sleeve_medium | reject_weak_fold_repeatability | 334.66% | 293.11% | 41.54% | 5/10 | 33.28% | 33.99% | 2.552 |
| benchmark_ranked_sleeve_checkpoint | reject_weak_fold_repeatability | 334.43% | 293.11% | 41.32% | 5/10 | 33.28% | 33.99% | 2.551 |
| sector_ranked_sleeve_conservative | reject_under_benchmark | 291.65% | 293.11% | -1.46% | 6/10 | 33.99% | 33.99% | 0.648 |
| benchmark_ranked_sleeve_slow | reject_under_benchmark | 290.07% | 293.11% | -3.04% | 5/10 | 33.99% | 33.99% | 0.674 |
| sector_ranked_sleeve_slow | reject_under_benchmark | 286.98% | 293.11% | -6.14% | 5/10 | 33.99% | 33.99% | 0.530 |
| sector_ranked_sleeve_checkpoint | reject_under_benchmark | 289.91% | 293.11% | -3.21% | 4/10 | 33.81% | 33.99% | 1.594 |
| sector_ranked_sleeve_medium | reject_under_benchmark | 287.11% | 293.11% | -6.00% | 4/10 | 33.81% | 33.99% | 1.576 |
