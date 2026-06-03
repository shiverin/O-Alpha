# S&P 500 Historical Constituent Source Probe

Generated: 2026-06-03T15:08:00Z

## Purpose

Test whether a free point-in-time S&P 500 data stack is sufficient to replace
the current static Yahoo100 current-large-cap panel for official learned-ranker
validation.

## Candidate Sources Tested

| Source | Local artifact | Result |
|---|---|---|
| `fja05680/sp500` | `free_sources/fja05680_sp500_ticker_start_end.csv` | Downloaded successfully. Interval file built into a PIT manifest with `755` symbols and `764` intervals for 2015-2026. |
| `hanshof/sp500_constituents` | `free_sources/hanshof_sp_500_historical_components.csv` | Downloaded successfully. Snapshot file built into a PIT manifest with `712` symbols and `721` intervals for 2015-2026. |

Source URLs:

- https://github.com/fja05680/sp500/blob/master/sp500_ticker_start_end.csv
- https://github.com/hanshof/sp500_constituents/blob/main/sp_500_historical_components.csv

## Commands Run

```bash
curl -L --fail --retry 3 --retry-delay 2 \
  -o reports/batches/2026-06-03_sp500_constituent_source_probe/free_sources/fja05680_sp500_ticker_start_end.csv \
  https://raw.githubusercontent.com/fja05680/sp500/master/sp500_ticker_start_end.csv

curl -L --fail --retry 3 --retry-delay 2 \
  -o reports/batches/2026-06-03_sp500_constituent_source_probe/free_sources/hanshof_sp_500_historical_components.csv \
  https://raw.githubusercontent.com/hanshof/sp500_constituents/main/sp_500_historical_components.csv

research/ml/.venv/bin/python research/ml/build_pit_universe.py \
  --input reports/batches/2026-06-03_sp500_constituent_source_probe/free_sources/fja05680_sp500_ticker_start_end.csv \
  --format interval \
  --source-name fja05680_sp500 \
  --source-url https://github.com/fja05680/sp500/blob/master/sp500_ticker_start_end.csv \
  --min-date 2015-01-01 \
  --max-date 2026-06-01 \
  --out reports/batches/2026-06-03_sp500_constituent_source_probe/free_sources/fja05680_pit_universe_2015_2026.json

research/ml/.venv/bin/python research/ml/build_pit_universe.py \
  --input reports/batches/2026-06-03_sp500_constituent_source_probe/free_sources/hanshof_sp_500_historical_components.csv \
  --format snapshot \
  --source-name hanshof_sp500_constituents \
  --source-url https://github.com/hanshof/sp500_constituents/blob/main/sp_500_historical_components.csv \
  --min-date 2015-01-01 \
  --max-date 2026-06-01 \
  --out reports/batches/2026-06-03_sp500_constituent_source_probe/free_sources/hanshof_pit_universe_2015_2026.json
```

## Constituent Cross-Check

The two community sources are broadly similar but not identical:

- `fja05680`: `755` symbols, active count `499` on 2015-01-02 and `503` on 2026-06-01.
- `hanshof`: `712` symbols, active count `464` on 2015-01-02 and `503` on 2026-06-01.
- Intersection: `710` symbols.
- `fja05680` has `45` symbols not in `hanshof`; `hanshof` has `2` symbols not in `fja05680`.

This is good enough for research cross-checks, but the disagreement is one more
reason not to call either community file deployment-grade without vendor
reconciliation.

## Price Coverage Audit Against Current Bars

Current available bars:

`reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2026/bars.csv`

Coverage commands used `--min-active-symbols 450` and
`--min-coverage-ratio 0.95`.

| PIT source | Coverage report | Status | PIT symbols | Bars symbols | Observed PIT symbols | Missing PIT symbols | Coverage ratio | Min covered symbols |
|---|---|---|---:|---:|---:|---:|---:|---:|
| `fja05680` | `free_sources/fja05680_coverage_on_yahoo100/pit_coverage_report.md` | failed | `755` | `100` | `84` | `671` | `0.1602` | `77` |
| `hanshof` | `free_sources/hanshof_coverage_on_yahoo100/pit_coverage_report.md` | failed | `712` | `100` | `84` | `628` | `0.1635` | `77` |

This means the current Yahoo100 bars cannot support a true S&P 500 PIT
validation. Running the learned ranker on these PIT manifests with the current
bars would silently discard most historical constituents or fail alignment,
which would not be credible alpha evidence.

## Free Bar Source Probe

Probe artifacts:

- `free_sources/free_bar_probe.json`
- `free_sources/free_bar_probe.csv`

Sampled current and delisted/renamed symbols:

`A`, `AAL`, `AABA`, `ABC`, `ABMD`, `AGN`, `ALTR`, `EMC`, `MON`, `PX`, `RTN`,
`YHOO`, `BRK-B`, `BF-B`, `AAPL`.

Findings:

- Stooq CSV endpoint returned an API-key instruction rather than usable price
  CSVs for the sampled symbols.
- Yahoo chart returned usable bars for current symbols such as `A`, `AAL`,
  `BRK-B`, `BF-B`, and `AAPL`, and partial bars for `EMC`, but returned HTTP
  404 for many delisted/renamed names including `AABA`, `ABC`, `ABMD`, `AGN`,
  `ALTR`, `MON`, `PX`, `RTN`, and `YHOO`.
- FMP demo endpoint returned HTTP 401 for every sampled symbol, including
  `AAPL`.

## Verdict

The free constituent side can be used for research-only PIT experiments:

- Prefer `fja05680/sp500` as the primary interval source because it is already
  shaped as `ticker,start_date,end_date`.
- Use `hanshof/sp500_constituents` as a cross-check snapshot source.

The free stack is not deployment-grade and is not sufficient for an official
survivorship-aware learned-ranker validation yet, because adjusted daily bars
for removed/delisted/symbol-changed constituents are missing.

The next viable paths are:

1. Use a paid/institutional source with PIT constituents and adjusted historical
   bars, such as CRSP/WRDS, Norgate, Bloomberg/Refinitiv, Polygon with
   delisted/corporate-action coverage, or another licensed vendor.
2. If staying free/research-only, build a reduced PIT experiment around the
   subset of historical constituents that Yahoo can currently fetch, but mark
   that result as coverage-biased and non-promotable until
   `audit_pit_universe_coverage.py` passes.

No alpha claim should be made from the current static Yahoo100 panel as though
it were a true historical S&P 500 universe.
