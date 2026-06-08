# Research Log

This is the compact evidence ledger for O(Alpha). A result is only considered
real if it points to a harness-written artifact under `reports/batches/`.

## Current Best Checkpoint

`benchmark_lgbm_ranker_h63_s15_checkpoint` is the strongest current research
checkpoint. It uses the official Go `cmd/alpha-research` harness, year-specific
LightGBM ranker artifacts, Go feature/parity validation, and a VOO-funded h63
active sleeve over the Yahoo100 research panel.

Official VOO-benchmark promotion reports:

| Start | Report directory | Promote | Return | Sharpe | DSR | PBO | Trades |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: |
| 2015 | `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2015_csv/` | true | 631.93% | 1.013 | 1.000 | 0.133 | 104 |
| 2016 | `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2016_csv/` | true | 627.17% | 1.084 | 1.000 | 0.154 | 104 |
| 2017 | `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2017_csv/` | true | 534.28% | 1.089 | 1.000 | 0.091 | 104 |
| 2018 | `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2018_csv/` | true | 408.48% | 1.035 | 1.000 | 0.000 | 99 |
| 2019 | `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2019_csv/` | true | 412.48% | 1.159 | 1.000 | 0.143 | 84 |
| 2020 | `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2020_csv/` | true | 275.51% | 1.060 | 1.000 | 0.200 | 74 |

Verdict: promoted research checkpoint on the available Yahoo100 panel, but not
deployment-grade alpha until external data validity is solved.

## Essential Stress Tests

Equal-weight benchmark stress:

- Reports: `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_equal_2015_csv/` and `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_equal_2016_csv/`.
- Result: rejects in 2015 and 2016 because the selected Yahoo100 equal-weight benchmark is extremely strong. The 2015 equal-weight benchmark returned 1170.53%; the candidate returned 631.93% and failed with `turnover increases without return improvement`.
- Interpretation: the checkpoint is a VOO-core active-sleeve result, not an equal-weight selected-universe beater.

Ex-megacap stress:

- Reports: `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_exmegacap_2015_csv/` through `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_exmegacap_2020_csv/`.
- Result: rejects every start from 2015 through 2020 after excluding `AAPL`, `AMZN`, `AVGO`, `GOOG`, `GOOGL`, `LLY`, `META`, `MSFT`, `NVDA`, and `TSLA` from active candidates.
- Interpretation: the checkpoint depends materially on the mega-cap winner cohort.

Attribution and concentration:

- Report: `reports/batches/2026-06-03_ranker_attribution_h63_s15/ranker_attribution_report.md`.
- Result: 35 unique selected symbols, top symbol `TSLA` at 9.52% of target-weight selections, top-5 target-weight share 38.10%, HHI 0.0471, mega-cap target-weight share 35.24%.
- Interpretation: not a single-name trade, but mega-cap exposure remains important.

Agent catalog summary:

- Report: `reports/batches/2026-06-03_alpha_validation_agent_catalog_buckets_summary/agent_catalog_bucket_comparison.md`.
- Result: only `lgbm_ranker_h63_low` and `ranker_proxy_h63_low` promote in both 2015 and shifted 2016 catalog-bucket windows. Medium/high entries remain research challengers.

## Infrastructure Evidence

- Daily ranker feature parity: `reports/batches/2026-06-03_daily_ranker_feature_parity/parity_report.json`.
- LightGBM raw-score parity: `reports/batches/2026-06-03_daily_ranker_leaves_parity/stocks_h63_s15_2026_leaves_parity.json`.
- Report metadata audit: `reports/batches/2026-06-03_daily_ranker_report_metadata_audit/`.
- Paper-only signal wrapper: `reports/batches/2026-06-03_paper_ranker_signal_h63_2015_2026_csv/`.
- Restored daily ranker runtime artifacts: `reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2026/`,
  `reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/`,
  and smoke report `reports/batches/2026-06-08_restored_paper_ranker_signal_h63/`.
- PIT coverage audit smoke: `reports/batches/2026-06-03_pit_coverage_audit_smoke/`.
- S&P 500 constituent/source probe: `reports/batches/2026-06-03_sp500_constituent_source_probe/source_probe.md`.

## 2026-06-08T09:41:34Z — Restored ranker runtime artifacts

- Commands run: regenerated ML export with `cmd/ml-meta-research -mode export`
  over the retained Yahoo100 universe, then rebuilt
  `walkforward_daily_ranker.py --start-year 2018 --end-year 2026` with the four
  h63/h126 slow-horizon variants.
- Restored artifacts:
  `reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2026/bars.csv`,
  `reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2026/signals.csv`,
  and 36 model files under
  `reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts/`.
- Verification: `go test -count=1 ./internal/alpha/ranker ./internal/research/alphavalidation ./cmd/paper-ranker-signal`
  passed from `backend/`.
- Paper-only smoke:
  `reports/batches/2026-06-08_restored_paper_ranker_signal_h63/paper_ranker_signal.md`.
  It wrote targets from the restored model root with `orders_enabled=false` and
  `orders_submitted=0`.
- Verdict: RESTORED runtime artifacts only. This is not a new promotion claim;
  official promotion evidence remains the Go alpha-validation reports above.

## Superseded Research

Momentum, TSMOM, reversal, low-vol, deterministic h63 overlays, ETF-only rankers,
MA-event meta-labeling, and early Python-only ranker prescreens were all either
rejected by PBO, weaker than the current checkpoint, or pre-official-harness
work. Their bulky artifacts were removed during the cleanup; the decision state
above is the source of truth.
