# Parallel Alpha Research Plan

You are a parallel research agent working in the O-Alpha repo. Read
`AGENTS.md`, `docs/RESEARCH_LOG.md`, `docs/PLAN.md`, `docs/BLOCKERS.md`, and
`docs/ALPHA_RESEARCH_PROCESS.md` before running anything.

## Objective

Find a benchmark-funded active sleeve that survives the official Go validation
gate across primary and shifted windows. The current best checkpoint is
`benchmark_lgbm_ranker_h63_s15_checkpoint`: it promotes on the Yahoo100 CSV
starts 2015, 2016, 2017, 2018, 2019, and 2020, but still needs independent or
survivorship-aware validation before any deployment-grade claim.

Do not repeat two rejected refinements unless you have a materially different
reason:

- `benchmark_ranker_proxy_blend_checkpoint`: primary/shifted PBO
  `0.533`/`0.538`.
- `benchmark_ranker_proxy_h63_riskcap_checkpoint`: primary/shifted PBO
  `0.267`/`0.308`.
- `benchmark_ranker_proxy_h63_trendguard_checkpoint`: primary/shifted PBO
  `0.267`/`0.231`.
- `benchmark_ranker_proxy_h63_liquidity_checkpoint`: primary/shifted PBO
  `0.467`/`0.538`.
- The ETF16 sector-suite official check also failed: best primary/shifted PBO
  `0.600`/`0.615`.

Your job is to improve robustness, not merely maximize return.

Latest infrastructure checkpoint: daily-ranker Python-vs-Go feature parity is
cleared. See
`reports/batches/2026-06-03_daily_ranker_feature_parity/parity_report.json`
for the passed 500-row, 31-feature report.

Serving checkpoint: regression-objective and lambdarank rankers both pass Go
`leaves` raw-score parity on sampled fixtures, but the current regression
walk-forward is rejected. The h63 lambdarank path is the only official promoted
checkpoint on the current Yahoo100 panel.

## Non-Negotiable Rules

- Use `cmd/alpha-research` for promotion evidence.
- Every metric you report must cite an exact JSON/MD report path.
- Do not call Python prescreen output alpha.
- Do not hand-edit report metrics.
- Do not use live trading or broker order paths.
- If PBO is not estimated, promotion fails closed.
- Treat the Yahoo100 panel as research-only because it is not
  survivorship-aware.

## Setup

Run from repo root unless specified.

```bash
export GOCACHE=/tmp/oalpha-gocache
symbols=$(jq -r '.symbols | join(",")' reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_longpanel/*_alpha_validation.json)
bars=reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2026/bars.csv
```

Use the offline CSV path unless your environment has working database access.

## Lane A: Stabilize H63 Official PBO

The broad official suite already shows h63 is the best current candidate on the
Yahoo100 panel: it is the only primary-window promotion, and no broad-suite
candidate promotes on the shifted 2016 window. Naive h63/h84 blending and
simple h63 risk caps have now also failed. The h63 trend guard and ETF16
sector-suite checks did not promote either. The h63 liquidity gate was also
rejected. Serving-compatible regression ranker sweeps are mechanically
validated, but the current narrow and broad 2018-2026 walk-forwards are both
rejected for weak fold repeatability. LambdaRank raw-score serving parity is
now cleared. The h63-only lambdarank ranker has become the best checkpoint: it
promoted in official Yahoo100 starts 2015, 2016, 2017, 2018, 2019, and 2020.
The equal-weight benchmark stress rejects 2015 and 2016, then promotes 2017
through 2020. The ex-megacap stress rejects every start from 2015 through 2020
with PBO between `0.333` and `0.444`. Next work should validate external data
quality and survivorship, not simply tune more variants on the same panel. The
ETF-only learned-ranker prescreen also failed: champion
`etfs_h126_s15_top3_reb63_z10` returned `229.91%` versus VOO `225.07%` but beat
in only `5/9` folds, so there is no ETF-ranker promotion attempt to continue.
The bottom-45 current-panel holdout stress promoted only starts 2016 and 2017
and rejected starts 2015, 2018, 2019, and 2020. This diagnostic is
future-selected, so do not treat it as a tradable universe or survivorship
solution.

1. Inspect the promoted h63 learned-ranker reports:

```bash
jq '.candidates[0].promotion_decision, .candidates[0].primary.metrics' \
  reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2015_csv/*_alpha_validation.json
```

2. Avoid further same-panel h63 tuning unless diagnostics show a new,
   pre-declared reason. Better examples:

- independent/survivorship-aware validation of the promoted h63 learned ranker
- equal-weight or stronger benchmark comparisons on any new universe
- paper-trading wrapper with explicit model-root/version metadata
- survivorship-aware or ETF-only universe with a different economic mechanism,
  not the current rejected ETF sleeve family
- active-sleeve diagnostics that reduce selected-winner dependence, but only
  when the universe or economic signal changes materially
- do not rerun ETF learned-ranker presets unless the feature/label economics
  change; the 2018-2026 prescreen was already rejected
- build a point-in-time universe from a historical constituent CSV. Candidate
  sources are logged in
  `reports/batches/2026-06-03_sp500_constituent_source_probe/source_probe.md`,
  but local DNS currently blocks `raw.githubusercontent.com`

3. Register the family in
   `backend/internal/research/alphavalidation/strategies.go`.

4. Add at least three real `VariantFactories` for PBO.

5. Add tests in
   `backend/internal/research/alphavalidation/strategies_test.go`.

6. Run the official primary and shifted validations:

```bash
cd backend
GOCACHE=/tmp/oalpha-gocache go run ./cmd/alpha-research \
  -bars-csv ../$bars -symbols "$symbols" \
  -strategies <new_family> -timeframe 1Day \
  -from 2015-01-01 -to 2026-06-01 \
  -train-bars 756 -test-bars 252 -step-bars 126 -min-trades 30 \
  -output-dir ../reports/batches/<date>_<new_family>_2015_csv

GOCACHE=/tmp/oalpha-gocache go run ./cmd/alpha-research \
  -bars-csv ../$bars -symbols "$symbols" \
  -strategies <new_family> -timeframe 1Day \
  -from 2016-01-01 -to 2026-06-01 \
  -train-bars 756 -test-bars 252 -step-bars 126 -min-trades 30 \
  -output-dir ../reports/batches/<date>_<new_family>_2016_csv
```

Promotion target:

- primary promote `true`
- shifted promote `true`
- PBO `<= 0.20` in both
- Sharpe/Sortino/Calmar improve versus buy-hold
- max drawdown does not regress materially

## Lane B: Python Prescreen for Better Economic Signal

Use this only to generate hypotheses for Lane A.

1. Run slower-horizon ranker sweeps:

```bash
research/ml/.venv/bin/python research/ml/walkforward_daily_ranker.py \
  --bars-csv "$bars" --benchmark VOO \
  --start-year 2018 --end-year 2026 \
  --variants stocks_h63_s10_top3_reb63_z10,stocks_h63_s15_top3_reb63_z10,stocks_h126_s10_top3_reb63_z10,stocks_h126_s15_top3_reb63_z10 \
  --out-dir reports/batches/<date>_ranker_prescreen
```

2. If a variant beats VOO in all annual folds and has validation positive in at
   least 8/9 folds, translate it to a deterministic Go proxy or implement a
   ranker artifact parity path. Broad regression preset sweeps have already
   failed this standard; prefer the lambdarank raw-score path or a materially
   different signal before another large regression preset run.

3. Do not promote from Python.

## Lane C: Data Quality and Survivorship

The current Yahoo100 universe is biased. A valuable parallel task is to reduce
data bias:

- build a survivorship-aware historical constituent universe, or
- add an external adjusted daily panel with historical membership metadata, or
- create sector/ETF-only universes where survivorship risk is smaller.

Candidate constituent sources are documented in
`reports/batches/2026-06-03_sp500_constituent_source_probe/source_probe.md`.
If a CSV is downloaded/provided, first create a point-in-time universe manifest
and only then rerun the official Go gate.

Use the new plumbing:

```bash
research/ml/.venv/bin/python research/ml/build_pit_universe.py \
  --input reports/batches/2026-06-03_sp500_constituent_source_probe/<constituents.csv> \
  --format auto \
  --out reports/batches/2026-06-03_sp500_constituent_source_probe/pit_universe.json

export OALPHA_DAILY_RANKER_PIT_UNIVERSE=../reports/batches/2026-06-03_sp500_constituent_source_probe/pit_universe.json
```

Then run `cmd/alpha-research` normally. The PIT manifest only gates candidate
eligibility; price coverage for removed/delisted/symbol-changed names is still
required.

Audit coverage before performance:

```bash
research/ml/.venv/bin/python research/ml/audit_pit_universe_coverage.py \
  --pit-universe reports/batches/2026-06-03_sp500_constituent_source_probe/pit_universe.json \
  --bars-csv reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2026/bars.csv \
  --from-date 2015-01-01 \
  --to-date 2026-06-01 \
  --min-active-symbols 50 \
  --out-dir reports/batches/<date>_pit_coverage
```

Only proceed to official performance validation if `pit_coverage_report.json`
has status `passed`.

Then rerun the official gate on the same candidate families.

## Logging

Append every session to `docs/RESEARCH_LOG.md`:

```markdown
## <UTC timestamp> - <hypothesis>

- Command run: `<exact command>`
- Report path: `<exact JSON/MD path>`
- Universe/timeframe/date range:
- Result: net Sharpe | DSR | PBO | OOS trades | promote? | first gate reason
- Leakage/data issues:
- Verdict: PROMOTED / PARKED / REJECTED
- Next step:
```

Update `docs/PLAN.md` and `docs/BLOCKERS.md` after each run.

## Verification Before Commit

```bash
cd backend
GOCACHE=/tmp/oalpha-gocache go test ./...
GOCACHE=/tmp/oalpha-gocache go build ./...
cd ..
research/ml/.venv/bin/python -m py_compile research/ml/*.py
git diff --check
```
