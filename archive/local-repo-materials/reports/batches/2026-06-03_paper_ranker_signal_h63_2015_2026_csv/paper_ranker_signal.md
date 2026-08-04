# Paper Ranker Signal

- Generated at: `2026-06-03T14:41:32Z`
- Strategy: `daily_lgbm_ranker_sleeve`
- Paper only: `true`
- Orders enabled: `false`
- Orders submitted: `0`
- Broker connected: `false`
- Panel: `2015-01-02` to `2026-06-01` (`2869` bars)
- Latest signal time: `2026-06-01T13:30:00Z`
- Last target refresh: `2026-04-14T13:30:00Z`
- Target source: `last_non_empty_target`
- Model variant: `stocks_h63_s15_top3_reb63_z10`
- Model root: `../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts`
- Benchmark/core: `VOO`

## Warnings

- `research_simulation_only`
- `no_orders_submitted`
- `broker_client_not_loaded`
- `static_symbol_panel_external_validity_not_cleared`

## Targets

| Symbol | Weight | Role | Rank | Model Score | Score Z | Vol 20 | Confidence | Reason |
| --- | ---: | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `VOO` | 0.850000 | `benchmark_core` |  |  |  |  | 1.000 | `` |
| `AMAT` | 0.050000 | `active_sleeve` | 2 | 0.380384 | 2.867736 | 0.562253 | 1.000 | `` |
| `INTC` | 0.050000 | `active_sleeve` | 1 | 0.608913 | 4.395054 | 0.761797 | 1.000 | `` |
| `LRCX` | 0.050000 | `active_sleeve` | 3 | 0.233915 | 1.888850 | 0.670332 | 1.000 | `` |

## Latest Metadata

| Key | Value |
| --- | --- |
| `action` | `hold_targets` |
| `engine` | `daily_lgbm_ranker_sleeve` |
| `ranker_model_artifact_root` | `../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts` |
| `ranker_model_feature_count` | `31` |
| `ranker_model_feature_spec_version` | `daily_ranker_v1` |
| `ranker_model_loaded` | `true` |
| `ranker_model_path` | `../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts/stocks_h63_s15_top3_reb63_z10/2026/model.txt` |
| `ranker_model_sha256` | `396eae05e39c61f650ebabc6a0c8c3f209a107b7484419dacd6de4802165f02a` |
| `ranker_model_variant` | `stocks_h63_s15_top3_reb63_z10` |
| `ranker_model_year` | `2026` |
| `rebalance` | `false` |

## Last Target Metadata

| Key | Value |
| --- | --- |
| `active_weight` | `0.15` |
| `benchmark` | `VOO` |
| `benchmark_drawdown` | `0` |
| `benchmark_vol_20` | `0.19121441532026492` |
| `benchmark_weight` | `0.85` |
| `candidate_count` | `3` |
| `engine` | `daily_lgbm_ranker_sleeve` |
| `point_in_time_universe` | `false` |
| `ranker_model_artifact_root` | `../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts` |
| `ranker_model_feature_count` | `31` |
| `ranker_model_feature_spec_version` | `daily_ranker_v1` |
| `ranker_model_loaded` | `true` |
| `ranker_model_path` | `../reports/batches/2026-06-03_yahoo100_daily_ranker_walkforward_slow_horizons_2018_2026/fold_artifacts/stocks_h63_s15_top3_reb63_z10/2026/model.txt` |
| `ranker_model_sha256` | `396eae05e39c61f650ebabc6a0c8c3f209a107b7484419dacd6de4802165f02a` |
| `ranker_model_variant` | `stocks_h63_s15_top3_reb63_z10` |
| `ranker_model_year` | `2026` |
| `rebalance` | `true` |
| `risk_reasons` | `` |
| `selection_rows` | `3 rows` |
| `sleeve_scale` | `1` |
| `turnover` | `0.10000000000000003` |
| `turnover_band` | `0.05` |
