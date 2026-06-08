# Research Blockers

## Deployment Blockers

- The current Yahoo100 panel is a handpicked/current large-cap and ETF panel, not
  a survivorship-aware point-in-time universe. This is the dominant blocker.
- Equal-weight Yahoo100 stress rejects 2015 and 2016 starts, showing the selected
  universe itself is a very strong benchmark.
- Ex-megacap stress rejects every 2015-2020 start, so the promoted h63 checkpoint
  is too dependent on the mega-cap winner cohort for deployment-grade claims.
- Free constituent sources were found, but adjusted price coverage for removed,
  delisted, or renamed historical constituents is still insufficient.
- Paper-only signal emission exists and records `orders_enabled=false`; do not
  connect the ranker to broker execution without explicit live-trading approval
  and independent point-in-time validation.

## Research Blockers

- Any future alpha claim must cite JSON/MD artifacts under `reports/batches/`.
- Runtime model artifacts under `reports/batches/` are currently ignored by Git
  but are required by the daily ranker defaults. Keep the restored export and
  fold-artifact directories available locally, or move production model
  artifacts to a tracked/managed model-artifact location before deployment.
- PBO must be estimated from real sibling variants. Missing PBO fails closed.
- New learned-ranker features need Python-vs-Go parity fixtures before official
  promotion runs.
- New PIT validation must include a passing coverage audit before performance
  reports are considered credible.

## Data Requests

- Provide a survivorship-aware constituent history if the target is true S&P 500
  generalization.
- Provide or authorize a production-grade adjusted daily bar source covering
  delisted and symbol-changed names.
- After new data arrives, build a PIT universe manifest and run
  `research/ml/audit_pit_universe_coverage.py` before any performance test.
