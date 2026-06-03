# Research Plan

## Active Hypothesis

Find a benchmark-funded active sleeve that survives the fixed Go alpha-validation
harness across both primary and shifted windows and matches the consulting
reports' product definition:

- VOO benchmark core.
- Multi-asset ranking over a diversified universe.
- Risk-budgeted top-k active sleeve, not a full-book one-name switch.
- Turnover band and next-open, costed execution through the Go harness.
- Multiple parameter variants so PBO can be estimated after warmup-aware walk-forward splits.

## Current Checkpoint

`benchmark_lgbm_ranker_h63_s15_checkpoint` is the best current official
research checkpoint. It uses year-specific lambdarank artifacts, Go
daily-ranker feature parity, raw-score `leaves` inference, and a
benchmark-funded h63 active sleeve.

It promoted in all official Yahoo100 CSV windows tested so far:

- 2015 start: return `631.93%`, Sharpe `1.013`, DSR `1.000`, PBO `0.133`.
- 2016 start: return `627.17%`, Sharpe `1.084`, DSR `1.000`, PBO `0.154`.
- 2017 start: return `534.28%`, Sharpe `1.089`, DSR `1.000`, PBO `0.091`.
- 2018 start: return `408.48%`, Sharpe `1.035`, DSR `1.000`, PBO `0.000`.
- 2019 start: return `412.48%`, Sharpe `1.159`, DSR `1.000`, PBO `0.143`.
- 2020 start: return `275.51%`, Sharpe `1.060`, DSR `1.000`, PBO `0.200`.

Report directories:
`reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2015_csv/`
through
`reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_2020_csv/`.

Verdict: promoted research checkpoint on the available Yahoo100 panel. The next
blocker is not the gate; it is independent data quality and survivorship-aware
validation.

Equal-weight benchmark stress:

- Same strategy evaluated against `equal_weight` rejects the 2015 and 2016
  starts, because the current Yahoo100 equal-weight portfolio returned
  `1170.53%` and `932.56%` in those windows.
- The same equal-weight stress promotes on starts 2017, 2018, 2019, and 2020.
- Interpretation: the h63 learned ranker is a strong VOO-core active-sleeve
  checkpoint, but not a full-period selected-universe equal-weight beater.

Ex-megacap active-sleeve stress:

- Same h63 model/feature path, but excluding `AAPL`, `AMZN`, `AVGO`, `GOOG`,
  `GOOGL`, `LLY`, `META`, `MSFT`, `NVDA`, and `TSLA` from active-sleeve
  candidates, rejects all starts from 2015 through 2020.
- PBO rises to `0.333`, `0.385`, `0.364`, `0.444`, `0.429`, and `0.400`
  across the six start windows.
- Interpretation: the promoted Yahoo100 result is materially dependent on the
  obvious mega-cap winner cohort. This does not invalidate it as a VOO-core
  research checkpoint, but it weakens generalization confidence.

Attribution/concentration diagnostic:

- `reports/batches/2026-06-03_ranker_attribution_h63_s15/ranker_attribution_report.md`
  audits the Python walk-forward artifacts for `stocks_h63_s15_top3_reb63_z10`.
- It found `35` unique selected symbols across 2018-2026, top symbol `TSLA`
  at `9.52%` of target-weight selections, top-5 target-weight share `38.10%`,
  target-weight HHI `0.0471`, and mega-cap target-weight share `35.24%`.
- Interpretation: the checkpoint is not a single-name trade, but the
  mega-cap winner cohort is still important, especially in 2023-2024. This
  reinforces the ex-megacap blocker rather than clearing it.

Bottom-performer holdout stress:

- Same h63 family tested on VOO plus the 45 current-panel stocks with the
  weakest full-sample 2015-2026 returns promoted only in starts 2016 and 2017.
- It rejected starts 2015, 2018, 2019, and 2020 with PBO `0.267`, `0.222`,
  `0.286`, and `0.400`.
- Interpretation: this future-selected diagnostic shows the ranker is not only
  a mega-cap story, but still lacks enough stability to offset the selected
  universe blocker.

Two nearby robustness refinements were tested after this checkpoint and
rejected:

- `benchmark_ranker_proxy_blend_checkpoint` failed primary and shifted PBO
  (`0.533` and `0.538`).
- `benchmark_ranker_proxy_h63_riskcap_checkpoint` failed primary and shifted
  PBO (`0.267` and `0.308`).
- `benchmark_ranker_proxy_h63_trendguard_checkpoint` failed primary and shifted
  PBO (`0.267` and `0.231`).
- `benchmark_ranker_proxy_h63_liquidity_checkpoint` failed primary and shifted
  PBO (`0.467` and `0.538`).

These results argue against further cosmetic horizon/risk-cap tweaks around the
same signal unless the change is economically distinct and pre-declared.

The ETF/sector-only official suite was also tested as a less stock-biased
universe. It did not promote: the best primary ETF16 candidate had PBO `0.600`,
and the best shifted ETF16 candidate had PBO `0.615`.

An ETF-only learned-ranker prescreen was also rejected before official Go
promotion: `etfs_h126_s15_top3_reb63_z10` returned `229.91%` versus VOO
`225.07%`, but beat VOO in only `5/9` annual folds and had minimum active
breadth of `2`. Do not spend more cycles on ETF ranker presets unless the
economic signal changes materially.

## Next Steps

1. Move beyond the current deterministic h63 refinement family and the current
   ETF sleeve family. The h63/h84 blend, simple risk-capped h63, h63 trend
   guard, h63 liquidity gate, ETF16 sector-suite tests, and ETF-only learned
   ranker prescreen are rejected. Prefer the promoted learned h63 ranker
   checkpoint or a less biased universe with a different economic mechanism.
2. Require primary and shifted official windows to promote before calling a
   checkpoint deployment-grade. The learned h63 ranker now clears this on
   Yahoo100 versus VOO; repeat it on independent/survivorship-aware data before
   live use. The equal-weight stress failure in 2015/2016 makes this urgent.
3. Harden learned-ranker deployment plumbing. Model-root configuration and
   model/version metadata in strategy outputs are now implemented and verified
   by `reports/batches/2026-06-03_daily_ranker_model_metadata_hardening/`.
   Report-level model provenance is also implemented and verified by
   `reports/batches/2026-06-03_daily_ranker_report_metadata_audit/`.
   Paper-only signal emission is implemented and verified by
   `reports/batches/2026-06-03_paper_ranker_signal_h63_2015_2026_csv/`.
   Do not connect this to broker execution until independent/PIT validation
   clears and explicit live-trading approval is given.
4. Reduce data bias: move from the current Yahoo100 large-cap research panel to
   a survivorship-aware constituent history or a defensible ETF/sector universe.
   This is now the main research lane. The ex-megacap stress makes more
   same-panel h63 tuning low value.
5. Keep artifact/run manifests, dataset diagnostics, and exact report-path
   citations mandatory for every future ML or ranker artifact.
6. Acquire deployment-grade point-in-time constituents plus adjusted prices.
   Free constituent files from `fja05680/sp500` and
   `hanshof/sp500_constituents` were downloaded and converted successfully, but
   price coverage against the current Yahoo100 bars failed at only about `16%`
   member-day coverage. The updated source probe is documented in
   `reports/batches/2026-06-03_sp500_constituent_source_probe/source_probe.md`.
7. Use the new PIT validation plumbing once deployment-grade data arrives:
   `research/ml/build_pit_universe.py` builds the manifest, and
   `OALPHA_DAILY_RANKER_PIT_UNIVERSE=/path/to/pit_universe.json` activates
   point-in-time candidate gating in the official daily-ranker family.
8. Before any PIT performance claim, run `research/ml/audit_pit_universe_coverage.py`
   and require a passing `pit_coverage_report.json`. Static Yahoo100 smoke
   passes at coverage ratio `1.0000`; synthetic missing-symbol smoke fails at
   `0.0000`, proving the auditor catches unpriceable constituents.
9. Use `research/ml/audit_ranker_attribution.py` after any learned-ranker
   prescreen to log selected-symbol breadth, fold concentration, and mega-cap
   dependence before spending an official promotion run on a new variant.
10. Preserve ranker model audit metadata in every future runtime/backtest path:
    `ranker_model_artifact_root`, `ranker_model_variant`, `ranker_model_year`,
    `ranker_model_path`, `ranker_model_sha256`, feature spec version/count, and
    loaded/missing-model state.
11. Require future learned-ranker promotion reports to include the Markdown
    `Metadata Audit` section and JSON `audit_metadata` fields before using them
    as deployment evidence.
12. Use `docs/AGENT_STRATEGY_CATALOG.md` and
    `PortfolioAgentManager.StartCatalogPortfolioAgent(...)` for agent-side
    paper strategy selection. Keep all catalog entries paper-only until PIT
    adjusted-price coverage and explicit live-trading approval are available.
13. For the current paper-agent catalog, prefer the low-risk entries
    `lgbm_ranker_h63_low` and `ranker_proxy_h63_low` as defaults. The official
    catalog-bucket validation promoted both in 2015 and shifted 2016 windows:
    `reports/batches/2026-06-03_alpha_validation_agent_catalog_buckets_summary/agent_catalog_bucket_comparison.md`.
    Keep medium/high catalog entries as research challengers because they failed
    bucket-level PBO even when raw returns were higher.

## Promotion Standard

No strategy is called alpha unless the Go promotion gate returns `Promote=true`
with PBO actually estimated. For deployment-grade confidence, require the same
family to clear at least one shifted-date official audit.
