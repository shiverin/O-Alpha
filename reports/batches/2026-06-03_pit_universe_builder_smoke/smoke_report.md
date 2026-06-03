# Point-In-Time Universe Builder Smoke Report

Generated: 2026-06-03T14:14:43Z

## Purpose

Validate that `research/ml/build_pit_universe.py` can produce the JSON manifest consumed by the Go daily ranker point-in-time eligibility gate.

## Commands

Interval-row format:

```bash
research/ml/.venv/bin/python research/ml/build_pit_universe.py \
  --input reports/batches/2026-06-03_pit_universe_builder_smoke/interval_input.csv \
  --format interval \
  --source-name synthetic_interval \
  --out reports/batches/2026-06-03_pit_universe_builder_smoke/interval_manifest.json
```

Snapshot-row format:

```bash
research/ml/.venv/bin/python research/ml/build_pit_universe.py \
  --input reports/batches/2026-06-03_pit_universe_builder_smoke/snapshot_input.csv \
  --format snapshot \
  --source-name synthetic_snapshot \
  --out reports/batches/2026-06-03_pit_universe_builder_smoke/snapshot_manifest.json
```

Targeted Go tests:

```bash
cd backend && export GOCACHE=/tmp/oalpha-gocache
go test ./internal/alpha/ranker ./internal/research/alphavalidation \
  -run 'PointInTime|PIT|DailyRankerSleeveConfigUses'
```

## Result

- `interval_manifest.json`: 3 symbols, 3 intervals.
- `snapshot_manifest.json`: 3 symbols, 3 intervals.
- Targeted Go tests passed.

## Runtime Hook

Official ranker validations can now use a point-in-time manifest by setting:

```bash
export OALPHA_DAILY_RANKER_PIT_UNIVERSE=/path/to/pit_universe.json
```

Then run the usual official command:

```bash
cd backend && GOCACHE=/tmp/oalpha-gocache go run ./cmd/alpha-research \
  -bars-csv ../reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2026/bars.csv \
  -symbols "<VOO plus all possible historical symbols present in bars>" \
  -strategies benchmark_lgbm_ranker_h63 \
  -timeframe 1Day -from 2015-01-01 -to 2026-06-01 \
  -train-bars 756 -test-bars 252 -step-bars 126 -min-trades 30 \
  -output-dir ../reports/batches/<pit_validation_batch>
```

This only gates active-sleeve candidate eligibility. It does not solve missing prices for delisted or symbol-changed names; that still requires a real adjusted historical price source.
