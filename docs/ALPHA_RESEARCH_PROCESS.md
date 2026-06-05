<<<<<<< HEAD
# Alpha Research Process Handoff

Last updated: 2026-06-03T14:00:35Z

This branch is an autonomous alpha-research checkpoint. The core rule is still:
do not call a strategy alpha unless the official Go validation gate returns
`Promote=true` with PBO actually estimated.

## Current State

The best current checkpoint is a benchmark-funded learned h63 LightGBM ranker:

- Hold VOO as the benchmark core.
- Allocate a small active sleeve into the top-ranked stock names.
- Use next-open execution, turnover bands, and normal/2x/3x cost stress.
- Estimate PBO from real sibling variants through `cmd/alpha-research`.
- Serve year-specific lambdarank artifacts through the Go `leaves` raw-score
  path with Python-vs-Go feature and raw-score parity already checked.

The prior deterministic proxy checkpoint was:

- `benchmark_ranker_proxy_h63_checkpoint`
- Primary official report:
  `reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_h63_longpanel_csv/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.md`
- Result: promote `true`, return `488.98%`, Sharpe `0.964`, DSR `1.000`,
  PBO `0.200`, trades `142`.
- Shifted official report:
  `reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_h63_shifted_2016_csv/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.md`
- Shifted result: promote `false`, return `451.32%`, Sharpe `1.006`,
  DSR `1.000`, PBO `0.231`, first reason `PBO 0.231 above 0.200`.

Interpretation: this is a real official promoted checkpoint on the primary
window, but it is not deployment-grade yet because the shifted official audit
fails the PBO gate.

The nearby 84-day checkpoint was tested and rejected:

- Primary report:
  `reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_h84_longpanel_csv/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.md`
- Result: promote `false`, return `461.86%`, Sharpe `0.945`, PBO `0.267`.
- Shifted report:
  `reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_h84_shifted_2016_csv/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.md`
- Result: promote `false`, return `437.50%`, Sharpe `0.996`, PBO `0.308`.

Two additional h63 robustness refinements were also rejected:

- `benchmark_ranker_proxy_blend_checkpoint`: primary PBO `0.533`, shifted PBO
  `0.538`.
- `benchmark_ranker_proxy_h63_riskcap_checkpoint`: primary return `466.10%`,
  Sharpe `0.946`, PBO `0.267`; shifted return `429.91%`, Sharpe `0.986`, PBO
  `0.308`.
- `benchmark_ranker_proxy_h63_trendguard_checkpoint`: primary return `460.66%`,
  Sharpe `0.949`, PBO `0.267`; shifted return `443.06%`, Sharpe `1.006`, PBO
  `0.231`.
- `benchmark_ranker_proxy_h63_liquidity_checkpoint`: primary return `392.39%`,
  Sharpe `0.867`, PBO `0.467`; shifted return `375.24%`, Sharpe `0.915`, PBO
  `0.538`.

Interpretation: cosmetic blends, simple volatility caps, a VOO-trend risk-state
overlay, and liquidity gating are not solving the variant-selection instability.
The next useful step is a better ranking target or a less biased validation
universe with a different economic mechanism.

A cleaner ETF/sector-only official suite was tested next:

- Primary report:
  `reports/batches/2026-06-03_alpha_validation_etf16_sector_suite_2015_csv/voo_dia_iwm_qqq_smh_spy_vti_xlb_xle_xlf_xli_xlk_xlp_xlu_xlv_xly_1day_alpha_validation.md`
- Shifted report:
  `reports/batches/2026-06-03_alpha_validation_etf16_sector_suite_2016_csv/voo_dia_iwm_qqq_smh_spy_vti_xlb_xle_xlf_xli_xlk_xlp_xlu_xlv_xly_1day_alpha_validation.md`
- Result: no promotions; best primary PBO `0.600`, best shifted PBO `0.615`.

Interpretation: reducing the universe to sector/broad ETFs lowers stock
survivorship concerns, but the existing ETF sleeve logic still fails the
official variant-selection test.

The broad benchmark-funded official suite confirms the same conclusion:

- Primary suite report:
  `reports/batches/2026-06-03_alpha_validation_yahoo100_benchmark_funded_suite_2015_csv/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.md`
- Primary suite result: h63 is the only promoted candidate; all other
  benchmark-funded families fail mostly on PBO.
- Shifted suite report:
  `reports/batches/2026-06-03_alpha_validation_yahoo100_benchmark_funded_suite_2016_csv/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.md`
- Shifted suite result: no candidate promotes; h63 has the strongest Sharpe
  (`1.006`) but fails with PBO `0.231`.

## Data

The current long-panel research artifact is:

`reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2026/bars.csv`

This is a Yahoo-adjusted, current-large-cap/ETF research panel, not a
survivorship-aware historical S&P 500 constituent panel. Treat it as useful for
research discovery only. The equal-weight benchmark's very high return in prior
reports is a reminder that the panel is materially selection-biased.

`cmd/alpha-research` now supports `-bars-csv`, so official validation can run
offline against an exported CSV when the database is unreachable. If a database
is available, the original DB path still works.

## Official Validation Tool

Run from `backend/`:

```bash
symbols=$(jq -r '.symbols | join(",")' ../reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_longpanel/*_alpha_validation.json)
GOCACHE=/tmp/oalpha-gocache go run ./cmd/alpha-research \
  -bars-csv ../reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2026/bars.csv \
  -symbols "$symbols" \
  -strategies benchmark_ranker_proxy_h63 \
  -timeframe 1Day -from 2015-01-01 -to 2026-06-01 \
  -train-bars 756 -test-bars 252 -step-bars 126 -min-trades 30 \
  -output-dir ../reports/batches/<new_batch_name>
```

Use this tool for any promotion claim. It writes JSON and Markdown reports with:

- buy-hold/equal-weight/flat benchmarks
- net performance after costs
- stress costs
- walk-forward train/test windows
- DSR
- PBO diagnostics
- final `PromotionDecision`

## Python Prescreen Tools

Python tools are useful for hypothesis search, but they are not promotion
evidence by themselves.

Daily LightGBM ranker walk-forward:

```bash
research/ml/.venv/bin/python research/ml/walkforward_daily_ranker.py \
  --bars-csv reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2026/bars.csv \
  --benchmark VOO --start-year 2018 --end-year 2026 \
  --variants stocks_h63_s10_top3_reb63_z10,stocks_h63_s15_top3_reb63_z10,stocks_h126_s10_top3_reb63_z10,stocks_h126_s15_top3_reb63_z10 \
  --out-dir reports/batches/<new_python_prescreen_batch>
```

The best Python prescreen was `stocks_h63_s15_top3_reb63_z10`:

- 2018-2026 report:
  `reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/daily_ranker_walkforward.md`
- Result: return `281.95%` versus VOO `225.07%`, folds beating `9/9`.
- Decision remained `research_only_weak_validation`, because validation was
  positive in only `6/9` folds.

## Implementation Map

Key Go files:

- `backend/cmd/alpha-research/main.go`: official CLI; now supports `-bars-csv`.
- `backend/cmd/validate-daily-ranker-parity/main.go`: validates Go runtime
  daily-ranker features against Python training features.
- `backend/internal/research/alphavalidation/runner.go`: official validation,
  DSR/PBO, warmup-aware walk-forward, promotion gate.
- `backend/internal/research/alphavalidation/strategies.go`: strategy registry
  and PBO variants.
- `backend/internal/alpha/momentum/composite_momentum.go`: benchmark-funded
  active-sleeve strategy engine.
- `backend/internal/ml/daily_ranker_features.go`: Go implementation of the
  31-feature daily LightGBM ranker vector.
- `backend/internal/backtest/portfolio_engine.go`: next-open portfolio execution.

Key Python files:

- `research/ml/backtest_daily_ranker_sleeve.py`: daily LightGBM ranker and
  costed benchmark-funded sleeve simulation.
- `research/ml/walkforward_daily_ranker.py`: annual model-per-fold ranker
  prescreen with per-fold artifacts.
- `research/ml/compare_active_sleeves.py`: deterministic active-sleeve
  comparison.
- `research/ml/export_feature_fixture.py`: feature parity fixture export.
- `research/ml/export_daily_ranker_feature_fixture.py`: daily-ranker parity
  fixture export.

Ranker feature parity checkpoint:

- Report: `reports/batches/2026-06-03_daily_ranker_feature_parity/parity_report.json`
- Result: `passed`, 500 rows, 31 features, max absolute error
  `2.3363533330211794e-12`.
- Interpretation: Go and Python now agree on the daily-ranker feature formulas;
  model promotion still requires official out-of-sample validation.

Ranker serving checkpoint:

- Regression raw-score leaves parity report:
  `reports/batches/2026-06-03_daily_ranker_leaves_parity/regression_slow_h63_s10_2026_leaves_parity.json`
- Result: `passed`, 500 rows, max absolute raw-score error `0`.
- LambdaRank raw-score leaves parity report:
  `reports/batches/2026-06-03_daily_ranker_leaves_parity/stocks_h63_s15_2026_leaves_parity.json`
- Result: `passed`, 500 rows, max absolute raw-score error `0`.
- Interpretation: Go can now serve raw ranker scores for sampled lambdarank
  artifacts; official model-per-fold validation is still required before any
  promotion claim.
- Regression walk-forward report:
  `reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_regression_slow_horizons_2018_2026/daily_ranker_walkforward.md`
- Result: champion returned `241.74%` versus VOO `225.07%`, but decision was
  `reject_weak_fold_repeatability`.
- Broad regression preset sweep:
  `reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_regression_broad_presets_2018_2026/daily_ranker_walkforward.md`
- Result: champion again returned `241.74%` versus VOO `225.07%`, with
  `7/9` folds beating, `4/9` validation-positive folds, `0/9` candidate folds,
  and decision `reject_weak_fold_repeatability`.

Current promoted research checkpoint:

- Strategy: `benchmark_lgbm_ranker_h63_s15_checkpoint`
- Report directories:
  `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2015_csv/`
  through
  `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2020_csv/`
- Result: all six official starts from 2015 through 2020 promoted with PBO
  estimated. Returns ranged from `275.51%` to `631.93%`; PBO ranged from
  `0.000` to `0.200`; DSR was `1.000` in all six reports.
- Caveat: this is still Yahoo100/current-large-cap research data. The next
  standard is independent, survivorship-aware validation.

Equal-weight benchmark stress:

- Diagnostic family: `benchmark_lgbm_ranker_h63_equal`
- Reports:
  `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_equal_2015_csv/`
  through
  `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_equal_2020_csv/`
- Result: starts 2015 and 2016 reject against equal-weight Yahoo100; starts
  2017, 2018, 2019, and 2020 promote.
- Interpretation: the current checkpoint is a VOO-core active-sleeve result,
  not proof of dominance over the handpicked Yahoo100 equal-weight portfolio.

Ex-megacap active-sleeve stress:

- Diagnostic family: `benchmark_lgbm_ranker_h63_exmegacap`
- Reports:
  `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_exmegacap_2015_csv/`
  through
  `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_exmegacap_2020_csv/`
- Result: all six starts reject. PBO values are `0.333`, `0.385`, `0.364`,
  `0.444`, `0.429`, and `0.400` after excluding `AAPL`, `AMZN`, `AVGO`,
  `GOOG`, `GOOGL`, `LLY`, `META`, `MSFT`, `NVDA`, and `TSLA` from active
  candidates.
- Interpretation: the current checkpoint is materially dependent on the
  obvious mega-cap winner cohort in the selected Yahoo100 panel. Do not spend
  more effort tuning the same h63 family on the same panel unless the
  validation universe changes.

ETF-only learned-ranker prescreen:

- Python prescreen report:
  `reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_etf_presets_2018_2026/daily_ranker_walkforward.md`
- Result: champion `etfs_h126_s15_top3_reb63_z10` returned `229.91%` versus
  VOO `225.07%`, but beat VOO in only `5/9` annual folds and had minimum
  active breadth of `2`.
- Interpretation: the cleaner ETF universe does not rescue the current learned
  ranker idea. No official Go promotion run was attempted because the prescreen
  already failed repeatability.

Bottom-performer holdout stress:

- Official reports:
  `reports/batches/2026-06-03_alpha_validation_yahoo_bottom45_lgbm_ranker_h63_2015_csv/`
  through
  `reports/batches/2026-06-03_alpha_validation_yahoo_bottom45_lgbm_ranker_h63_2020_csv/`
- Result: VOO plus the 45 lowest-return current-panel stocks promoted only in
  starts 2016 and 2017. Starts 2015, 2018, 2019, and 2020 rejected on PBO.
- Interpretation: this future-selected diagnostic is not tradable and does not
  prove survivorship robustness. It shows the ranker can still add value
  outside the obvious winner cohort in some windows, but the family remains too
  unstable for deployment-grade claims.

Historical-constituent source probe:

- Report:
  `reports/batches/2026-06-03_sp500_constituent_source_probe/source_probe.md`
- Candidate sources found: `hanshof/sp500_constituents`,
  `fja05680/sp500`, History of Market's recent changes endpoint, and paid
  CRSP/Norgate-style data.
- Local blocker: shell DNS cannot resolve `raw.githubusercontent.com`, so the
  historical constituent CSV was not downloaded in this session.

Point-in-time universe plumbing:

- Builder: `research/ml/build_pit_universe.py`
- Smoke report:
  `reports/batches/2026-06-03_pit_universe_builder_smoke/smoke_report.md`
- Runtime hook:
  `OALPHA_DAILY_RANKER_PIT_UNIVERSE=/path/to/pit_universe.json`
- Interpretation: once a historical constituent CSV and matching adjusted bars
  are available, the official Go daily-ranker family can gate active-sleeve
  candidates by membership date. This does not solve missing delisted prices by
  itself.

PIT coverage audit:

- Auditor: `research/ml/audit_pit_universe_coverage.py`
- Smoke reports:
  `reports/batches/2026-06-03_pit_coverage_audit_smoke/synthetic_missing/pit_coverage_report.md`
  and
  `reports/batches/2026-06-03_pit_coverage_audit_smoke/static_yahoo100/pit_coverage_report.md`
- Result: synthetic missing-symbol audit failed with coverage ratio `0.0000`;
  static Yahoo100 audit passed with coverage ratio `1.0000` and minimum covered
  symbols `100`.
- Rule: a future point-in-time alpha claim must cite a passing coverage report
  before citing `cmd/alpha-research` performance.

## Guardrails

- Do not use Python prescreen results as alpha claims.
- Do not hand-edit report metrics.
- Use report artifacts and exact paths for every number.
- Keep PBO variants real; a single variant makes PBO fail closed.
- Use `GOCACHE=/tmp/oalpha-gocache` in this sandbox.
- `reports/` is gitignored. Force-add selected report artifacts when they are
  required for the committed research claim.
- If using `.env`, note that `.env` points at Docker host `timescale`; this
  session could not resolve it. `.env.local` points at Supabase, but network DNS
  is blocked in this sandbox. Prefer `-bars-csv` unless DB access is available.

## Verification

Before committing a research checkpoint, run:

```bash
cd backend
GOCACHE=/tmp/oalpha-gocache go test ./...
GOCACHE=/tmp/oalpha-gocache go build ./...
cd ..
research/ml/.venv/bin/python -m py_compile research/ml/*.py
git diff --check
```

Current verification note: `go test ./...` fails in this sandbox only because
`internal/alpaca` uses `httptest.NewServer`, and the sandbox cannot bind a
localhost port. The rest of the Go suite passed with:

```bash
cd backend
export GOCACHE=/tmp/oalpha-gocache
go test $(go list ./... | grep -v '/internal/alpaca$')
```

`go build ./...`, Python `py_compile`, and `git diff --check` passed.
=======
# Alpha Research Process

1. Run `cmd/alpha-research` against DB or offline CSV bars.
2. Use report JSON/Markdown artifacts as the only promotion evidence.
3. Require PBO to be estimated from real sibling variants.
4. Log exact commands, report paths, blockers, and next steps after each session.
>>>>>>> 3ea6d428 (Alpha research)
