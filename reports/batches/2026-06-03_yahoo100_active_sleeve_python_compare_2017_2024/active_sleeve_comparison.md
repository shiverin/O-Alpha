# Costed Active-Sleeve Comparison

- Benchmark: `VOO`
- Status: `python_prescreen_only`
- Promotion note: official alpha promotion still requires cmd/alpha-research DSR/PBO gate
- Variant count: `8`
- Champion: `benchmark_ranked_sleeve_medium`
- Champion decision: `reject_weak_fold_repeatability`

| Variant | Decision | Return | Benchmark | Excess | Folds Beating | Max DD | Benchmark Max DD | Mean Turnover |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| benchmark_ranked_sleeve_medium | reject_weak_fold_repeatability | 238.17% | 198.93% | 39.23% | 5/8 | 33.28% | 33.99% | 2.789 |
| benchmark_ranked_sleeve_checkpoint | reject_weak_fold_repeatability | 238.04% | 198.93% | 39.11% | 5/8 | 33.28% | 33.99% | 2.788 |
| benchmark_ranked_sleeve_conservative | reject_weak_fold_repeatability | 228.76% | 198.93% | 29.83% | 5/8 | 33.30% | 33.99% | 0.996 |
| benchmark_ranked_sleeve_slow | reject_weak_fold_repeatability | 207.74% | 198.93% | 8.81% | 5/8 | 33.99% | 33.99% | 0.728 |
| sector_ranked_sleeve_checkpoint | reject_weak_fold_repeatability | 201.00% | 198.93% | 2.06% | 4/8 | 33.81% | 33.99% | 1.687 |
| sector_ranked_sleeve_slow | reject_weak_fold_repeatability | 199.57% | 198.93% | 0.63% | 4/8 | 33.99% | 33.99% | 0.533 |
| sector_ranked_sleeve_medium | reject_weak_fold_repeatability | 199.23% | 198.93% | 0.30% | 4/8 | 33.81% | 33.99% | 1.653 |
| sector_ranked_sleeve_conservative | reject_under_benchmark | 195.62% | 198.93% | -3.31% | 5/8 | 33.99% | 33.99% | 0.673 |
