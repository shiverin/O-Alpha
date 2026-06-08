# Research Plan

## Current State

The active checkpoint is `benchmark_lgbm_ranker_h63_s15_checkpoint`. It promotes
on Yahoo100 VOO-benchmark starts from 2015 through 2020, but remains blocked by
external validity:

- The panel is not survivorship-aware.
- Equal-weight stress rejects early starts.
- Ex-megacap stress rejects all starts.

Use `docs/RESEARCH_LOG.md` for the compact evidence table and exact report
directories. Use `docs/BLOCKERS.md` for blockers and data requests.

## Default Paper-Agent Choice

For paper-only agent catalog selection, prefer:

- `lgbm_ranker_h63_low`
- `ranker_proxy_h63_low`

Reason: the catalog bucket comparison promoted only those low-risk entries in
both the 2015 and shifted 2016 catalog-bucket windows. Medium/high entries have
higher raw returns but failed bucket-level PBO and should stay research-only.

Evidence: `reports/batches/2026-06-03_alpha_validation_agent_catalog_buckets_summary/agent_catalog_bucket_comparison.md`.

## Next Research Unit

1. Acquire or build deployment-grade point-in-time constituents and adjusted
   daily prices.
2. Build a PIT manifest with `research/ml/build_pit_universe.py`.
3. Audit coverage with `research/ml/audit_pit_universe_coverage.py`; do not run
   performance tests until coverage passes.
4. Re-run the official h63 family through `cmd/alpha-research` on the PIT panel.
5. Repeat equal-weight and ex-megacap style stresses on the PIT panel.
6. Log only report-backed results in `docs/RESEARCH_LOG.md`.

## Do Not Spend More Time On

- Cosmetic h63 overlays such as simple blends, risk caps, trend guards, or
  liquidity gates.
- ETF-only learned-ranker presets unless the economic signal changes materially.
- Old MA-event meta-labeling, which had weak CPCV/AUC evidence.
- Any Python-only prescreen as a promotable alpha claim.

## Required Verification

Before any new checkpoint is called promotable:

- Official Go harness report exists under `reports/batches/`.
- Promotion gate returns `Promote=true`.
- PBO is actually estimated and at or below 0.20.
- DSR is at or above 0.95.
- OOS trades meet the configured minimum.
- Data-quality and no-lookahead assumptions are explicitly logged.
- At least one shifted-date official audit also promotes.
