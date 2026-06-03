# Research Blockers

## Current

- `benchmark_lgbm_ranker_h63_s15_checkpoint` is promoted by the official Go
  gate on Yahoo100 starts 2015, 2016, 2017, 2018, 2019, and 2020, with PBO
  estimated in every run. The remaining blocker is not the repo gate on this
  panel; it is external validity on survivorship-aware or otherwise defensible
  data.
- The current 100-symbol panel is handpicked/current large caps and ETFs;
  equal-weight buy-and-hold returns of 1170.53% over 2015-2026 show material
  selection/survivorship bias. This remains the dominant blocker before live
  deployment claims.
- The h63-only learned-ranker family was tested after broad-family diagnostics
  showed h126/h63 variant instability. Treat the all-window promotion as a
  strong research checkpoint, but still require independent validation because
  the narrowing was informed by observed diagnostics.
- Equal-weight benchmark stress rejects the h63 learned ranker on 2015 and
  2016 starts: `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_equal_2015_csv/`
  and `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_equal_2016_csv/`
  fail with `turnover increases without return improvement`. The same stress
  promotes on starts 2017 through 2020. Treat this as evidence that the
  full-period Yahoo100 result is not an equal-weight selected-universe alpha.
- Ex-megacap stress rejects the h63 learned ranker on every start from 2015
  through 2020: `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_exmegacap_2015_csv/`
  through
  `reports/batches/2026-06-03_alpha_validation_yahoo100_lgbm_ranker_h63_exmegacap_2020_csv/`
  all fail with PBO above `0.200` after excluding `AAPL`, `AMZN`, `AVGO`,
  `GOOG`, `GOOGL`, `LLY`, `META`, `MSFT`, `NVDA`, and `TSLA` from active
  candidates. This strengthens the external-validity blocker: the current
  checkpoint is too dependent on the selected mega-cap winner cohort for
  deployment-grade claims.
- Attribution audit `reports/batches/2026-06-03_ranker_attribution_h63_s15/ranker_attribution_report.md`
  shows the h63 ranker is not a single-name trade (`35` unique selected symbols,
  top symbol `TSLA` at `9.52%`, target-weight HHI `0.0471`), but mega-cap
  exposure is still material (`35.24%` target-weight share overall and `66.67%`
  in both 2023 and 2024 folds). This supports the ex-megacap blocker rather
  than resolving it.
- Bottom-performer holdout stress on VOO plus the 45 lowest full-sample
  buy-and-hold names from the current Yahoo100 stock panel promoted only on
  starts 2016 and 2017, and rejected starts 2015, 2018, 2019, and 2020:
  `reports/batches/2026-06-03_alpha_validation_yahoo_bottom45_lgbm_ranker_h63_2015_csv/`
  through
  `reports/batches/2026-06-03_alpha_validation_yahoo_bottom45_lgbm_ranker_h63_2020_csv/`.
  This diagnostic is intentionally future-selected and cannot be traded; it
  reinforces that selected-panel robustness remains incomplete.
- `benchmark_ranker_proxy_blend_checkpoint` was tested as a naive h63/h84/h126
  horizon blend and rejected: primary PBO `0.533`, shifted PBO `0.538`.
- `benchmark_ranker_proxy_h63_riskcap_checkpoint` was tested as a simple
  risk-capped h63 refinement and rejected: primary PBO `0.267`, shifted PBO
  `0.308`.
- `benchmark_ranker_proxy_h63_trendguard_checkpoint` was tested as a VOO-trend
  risk-state overlay and rejected: primary PBO `0.267`, shifted PBO `0.231`.
  It lowered turnover but did not improve max drawdown enough and gave up
  primary return versus the original h63 checkpoint.
- `benchmark_ranker_proxy_h63_liquidity_checkpoint` was tested as a
  point-in-time median-dollar-volume eligibility filter and rejected: primary
  PBO `0.467`, shifted PBO `0.538`. Liquidity gating materially reduced return
  versus the original h63 checkpoint.
- `benchmark_ranker_proxy_h84_checkpoint` was tested as a nearby sibling and
  failed both primary and shifted official audits because PBO was `0.267` and
  `0.308`.
- The cleaner 16-symbol ETF/sector official suite rejected all current ETF
  sleeve families. Best primary ETF16 candidate PBO was `0.600`; best shifted
  ETF16 candidate PBO was `0.615`. Equal-weight ETF16 also beat the active
  sleeves on return in both windows.
- ETF-only learned-ranker presets were tested as a cleaner-universe prescreen
  and rejected before official promotion:
  `reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_etf_presets_2018_2026/daily_ranker_walkforward.md`
  shows champion `etfs_h126_s15_top3_reb63_z10` returned `229.91%` versus VOO
  `225.07%`, but beat in only `5/9` folds and had minimum active breadth of
  `2`. This confirms that simply moving the same ranker machinery to sector
  ETFs does not solve robustness.
- `benchmark_ranked_sleeve_checkpoint` promoted on the 2017-start Yahoo-adjusted audit, but failed the full 2015 panel and failed/narrowly failed 2016 and 2018 shifted audits. Treat it as a promising checkpoint, not durable alpha.
- The composite momentum strategy is now represented as a Go `PortfolioStrategy`, but the corrected promotion gate rejects it because PBO remains above 0.20 and shifted-start validation underperforms VOO.
- The benchmark-funded TSMOM checkpoint has better full-period return/drawdown than VOO, but PBO remains too high on primary/denser splits and the shifted 2021 run underperforms VOO.
- The blended TSMOM family lowers primary PBO versus the unblended TSMOM but fails denser PBO and performs worse on the shifted 2021 run.
- The low-volatility sleeve passes PBO but does not improve return/risk enough versus VOO.
- The short-term reversal sleeve is not stable enough by PBO and does not beat VOO on shifted-start validation.
- `xsec_momentum` is now tested on a 102-symbol daily universe and passes the size guard, but its corrected return/Sharpe are too weak.
- Alpaca/IEX daily backfill appears capped for most symbols around the 2020-07-27 start in this dataset, even when requesting `INGEST_LOOKBACK=100000h`.
- Yahoo chart adjusted data now provides a longer 2015-2026 research panel, but it is not a licensed/survivorship-aware S&P 500 constituent history. `XLC` and `XLRE` have later inception dates and were excluded from 2015 full-panel validation.
- Stooq daily CSV and bulk endpoints were tested but were blocked by API-key/captcha/unauthorized responses in this environment.
- Candidate free historical S&P 500 constituent sources were downloaded and
  logged in
  `reports/batches/2026-06-03_sp500_constituent_source_probe/source_probe.md`.
  `fja05680/sp500` and `hanshof/sp500_constituents` built PIT manifests
  successfully, but the current Yahoo100 bars cover only about `16%` of PIT
  member-days. The blocker has moved from constituent download to adjusted
  price coverage for removed/delisted/symbol-changed constituents.
- Free bar-source probing did not clear the price blocker: Stooq CSV endpoints
  returned API-key instructions, Yahoo chart fetched current names but 404ed
  many delisted/renamed sample names, and FMP demo returned HTTP 401 even for
  `AAPL`.
- Point-in-time universe plumbing is now available: `research/ml/build_pit_universe.py`
  builds manifests, and `OALPHA_DAILY_RANKER_PIT_UNIVERSE` activates
  membership-date gating in the Go daily ranker. This removes the code-path
  blocker, but not the data blocker: we still need a real constituent CSV and
  historical adjusted prices for removed/delisted/symbol-changed names.
- Daily ranker model provenance is now available in runtime outputs: artifact
  root, variant, year, path, SHA-256, feature spec version/count, and loaded
  state. Verification artifact:
  `reports/batches/2026-06-03_daily_ranker_model_metadata_hardening/go_test_ranker_metadata.json`.
- Daily ranker model provenance is also now visible in official alpha-validation
  reports. Verification artifact:
  `reports/batches/2026-06-03_daily_ranker_report_metadata_audit/alpha_research_smoke_2020/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.md`
  includes `ranker_model_sha256`, model path, variant, year, feature spec, and
  active candidate count. This clears the model-audit metadata blocker.
- Paper-only daily ranker signal emission is now available through
  `cmd/paper-ranker-signal`. Verification artifact:
  `reports/batches/2026-06-03_paper_ranker_signal_h63_2015_2026_csv/paper_ranker_signal.md`
  records `paper_only=true`, `orders_enabled=false`, `orders_submitted=0`,
  `broker_connected=false`, the 2026 model SHA, and the current paper target
  book. This clears the paper-wrapper blocker, but it is intentionally not a
  broker/live-trading path and does not clear the external data-validity blocker.
- Agent-catalog bucket validation promotes only the low-risk entries
  `lgbm_ranker_h63_low` and `ranker_proxy_h63_low` on the current Yahoo100
  panel. The medium learned ranker remains the strongest standalone h63-family
  checkpoint, but it fails catalog-bucket PBO when compared against the other
  medium-risk catalog entries. Treat bucket-level medium/high catalog choices
  as research challengers, not default paper-agent settings. Summary artifact:
  `reports/batches/2026-06-03_alpha_validation_agent_catalog_buckets_summary/agent_catalog_bucket_comparison.md`.
- PIT price coverage auditing is now available:
  `research/ml/audit_pit_universe_coverage.py`. Static Yahoo100 coverage smoke
  passed with coverage ratio `1.0000`, while synthetic missing-symbol smoke
  failed with coverage ratio `0.0000`. Any future PIT alpha claim must cite a
  passing coverage report before performance reports are considered credible.
- Pre-fix reports generated before `2026-06-03T07:49:10Z` used walk-forward test windows without train-window warmup context; those PBO/walk-forward claims are stale.
- Consulting-report P0s partially complete: export/train manifests, dataset diagnostics, and Python-vs-Go feature parity now exist for the Yahoo100 live-core ML artifact. Remaining P0 blocker is costed next-open Python/Go rotation alignment.
- Python-vs-Go feature parity is cleared only for `ml_meta_label_live_core_v1`: 500 labeled-event rows, 40 features, max absolute error `5.810907310888069e-13`. Any new feature spec still needs its own parity fixture before promotion.
- The Yahoo100 live-core MA-event meta-label artifact is rejected: validation AUC 0.5049 but CPCV AUC mean 0.4992, status `rejected`, reason `cpcv_auc_below_random`.
- Python daily LightGBM ranker sleeves now show promising costed OOS checkpoints, especially 2022-2026 `stocks, horizon 21, sleeve 15%, top 3, rebalance 42`, but this is a Python pre-screen only. It is not an official promotable alpha result.
- The official Go `benchmark_ranker_proxy` deterministic approximation returned 433.00% versus VOO 351.93% on the 2015-2026 Yahoo100 long panel, but failed promotion because PBO was 0.600. The newer h63 proxy improves this materially, but shifted PBO still needs work.
- Daily ranker feature parity is cleared: `reports/batches/2026-06-03_daily_ranker_feature_parity/parity_report.json` passed on 500 rows, 31 features, max absolute error `2.3363533330211794e-12`. Daily ranker deployment is still blocked by model-per-fold official validation and promotion-gate integration.
- LambdaRank raw-score serving is no longer blocked: `reports/batches/2026-06-03_daily_ranker_leaves_parity/stocks_h63_s15_2026_leaves_parity.json` passed with 500 rows and max absolute raw-score error `0`. The remaining blocker is official model-per-fold validation with DSR/PBO or an equivalent ranker promotion gate.
- A broader deployable regression ranker sweep over 10 variants was also rejected: `reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_regression_broad_presets_2018_2026/daily_ranker_walkforward.md` shows champion `stocks_h63_s10_top3_reb63_z10` at `241.74%` versus VOO `225.07%`, but only `7/9` folds beat VOO, `4/9` validation folds were positive, `0/9` candidate folds passed, and the decision remained `reject_weak_fold_repeatability`.
- Ranker walk-forward conversion is weak: the best model-per-fold Python walk-forward variant returned 131.44% versus VOO 122.72%, but beat VOO in only 3/6 annual OOS folds. Validation liveness is not enough; portfolio selection/risk budgeting still needs work.
- Docker is not running in this session, so the Docker `.env` database host `timescale` is not reachable from local commands. The successful harness runs used `.env.local`.
- `reports/` is ignored by `.gitignore`, while `AGENTS.md` says report artifacts should be committed. Force-add selected JSON/MD artifacts that support committed claims, or adjust the ignore policy in a future cleanup.

## Data Requests

- Provide survivorship-aware constituents if the target is true S&P 500 generalization rather than a handpicked current large-cap panel.
- Provide or authorize a production-grade historical data source for adjusted daily bars if Yahoo research data is not acceptable for promotion evidence.
- If using a free constituent file, place it under
  `reports/batches/2026-06-03_sp500_constituent_source_probe/` so the next
  session can build a point-in-time universe manifest and rerun the official
  h63 validation on that universe.
- After placing the file, run `research/ml/build_pit_universe.py --format auto`
  and set `OALPHA_DAILY_RANKER_PIT_UNIVERSE` for `cmd/alpha-research`.
- Also run `research/ml/audit_pit_universe_coverage.py`; if it fails, ingest or
  source the missing adjusted bars before running official performance tests.
