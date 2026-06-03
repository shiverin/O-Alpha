# Agent Catalog Bucket Backtest Comparison

Generated from official `cmd/alpha-research` artifacts only. No metrics are hand-entered.

## Source Artifacts

- 2015 primary: `reports/batches/2026-06-03_alpha_validation_agent_catalog_buckets_2015_csv/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.json`
- 2016 shifted: `reports/batches/2026-06-03_alpha_validation_agent_catalog_buckets_2016_csv/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.json`
- Derived CSV: `reports/batches/2026-06-03_alpha_validation_agent_catalog_buckets_summary/agent_catalog_bucket_comparison.csv`

## Method

- Universe: Yahoo-adjusted 100-symbol research panel from exported `bars.csv`.
- Timeframe: `1Day`; train bars `756`, test bars `252`, step bars `126`, min OOS trades `30`.
- Strategy selector: `agent_catalog`, which runs all 9 agent-side low/medium/high catalog entries.
- PBO is estimated within each risk bucket using that bucket's three catalog entries as variants. This is a useful agent-catalog stability check, but it is not identical to the standalone h63-family PBO reports.
- All catalog entries are paper-only because the Yahoo100 panel is not survivorship-aware/PIT deployment-grade data.

## Benchmarks

| Window | Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Turnover |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| 2015 primary | buy_hold | 351.93% | 14.17% | 0.837 | 0.790 | 0.417 | 34.00% | 1.000 |
| 2015 primary | equal_weight | 1170.53% | 25.03% | 1.087 | 1.039 | 0.673 | 37.17% | 1.000 |
| 2015 primary | flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0.000 |
| 2016 shifted | buy_hold | 348.99% | 15.57% | 0.898 | 0.842 | 0.458 | 34.00% | 1.000 |
| 2016 shifted | equal_weight | 932.56% | 25.22% | 1.112 | 1.059 | 0.720 | 35.05% | 1.000 |
| 2016 shifted | flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0.000 |

## Strategy Results

### Low Risk

| Window | Strategy | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Turnover | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 2015 primary | lgbm_ranker_h63_low | true | 459.70% | 16.34% | 0.924 | 0.881 | 0.484 | 33.73% | 1.000 | 0.067 | 49 | 10.631 | pass |
| 2015 primary | lowvol_sleeve_low | false | 335.94% | 13.81% | 0.843 | 0.795 | 0.412 | 33.52% | 1.000 | 0.067 | 328 | 22.563 | turnover increases without return improvement |
| 2015 primary | ranker_proxy_h63_low | true | 420.45% | 15.60% | 0.907 | 0.854 | 0.441 | 35.39% | 1.000 | 0.067 | 103 | 16.034 | pass |
| 2016 shifted | lgbm_ranker_h63_low | true | 456.06% | 17.97% | 0.990 | 0.939 | 0.533 | 33.73% | 1.000 | 0.077 | 49 | 10.568 | pass |
| 2016 shifted | lowvol_sleeve_low | false | 329.99% | 15.09% | 0.901 | 0.843 | 0.450 | 33.52% | 1.000 | 0.077 | 296 | 21.209 | turnover increases without return improvement |
| 2016 shifted | ranker_proxy_h63_low | true | 397.15% | 16.71% | 0.953 | 0.890 | 0.472 | 35.39% | 1.000 | 0.077 | 94 | 14.704 | pass |

### Medium Risk

| Window | Strategy | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Turnover | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 2015 primary | lgbm_ranker_h63_medium | false | 631.93% | 19.11% | 1.013 | 0.966 | 0.561 | 34.07% | 1.000 | 0.333 | 104 | 26.352 | PBO 0.333 above 0.200 |
| 2015 primary | ranked_sleeve_medium | false | 474.41% | 16.60% | 0.944 | 0.888 | 0.482 | 34.45% | 1.000 | 0.333 | 561 | 104.657 | PBO 0.333 above 0.200 |
| 2015 primary | ranker_proxy_h63_medium | false | 488.98% | 16.86% | 0.964 | 0.901 | 0.473 | 35.67% | 1.000 | 0.333 | 142 | 31.400 | PBO 0.333 above 0.200 |
| 2016 shifted | lgbm_ranker_h63_medium | false | 627.17% | 21.06% | 1.084 | 1.027 | 0.618 | 34.07% | 1.000 | 0.308 | 104 | 26.187 | PBO 0.308 above 0.200 |
| 2016 shifted | ranked_sleeve_medium | false | 450.44% | 17.86% | 0.995 | 0.934 | 0.518 | 34.45% | 1.000 | 0.308 | 507 | 96.483 | PBO 0.308 above 0.200 |
| 2016 shifted | ranker_proxy_h63_medium | false | 451.32% | 17.87% | 1.006 | 0.934 | 0.501 | 35.67% | 1.000 | 0.308 | 129 | 28.372 | PBO 0.308 above 0.200 |

### High Risk

| Window | Strategy | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Turnover | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| 2015 primary | benchmark_tsmom_high | false | 447.24% | 16.11% | 0.916 | 0.866 | 0.462 | 34.89% | 1.000 | 0.333 | 396 | 55.143 | PBO 0.333 above 0.200 |
| 2015 primary | composite_momentum_high | false | 421.90% | 15.63% | 0.916 | 0.854 | 0.462 | 33.83% | 1.000 | 0.333 | 703 | 150.574 | PBO 0.333 above 0.200 |
| 2015 primary | lgbm_ranker_h63_high | false | 624.09% | 19.00% | 0.996 | 0.938 | 0.556 | 34.17% | 1.000 | 0.333 | 236 | 63.470 | PBO 0.333 above 0.200 |
| 2016 shifted | benchmark_tsmom_high | false | 423.35% | 17.28% | 0.960 | 0.903 | 0.495 | 34.89% | 1.000 | 0.231 | 360 | 51.151 | PBO 0.231 above 0.200 |
| 2016 shifted | composite_momentum_high | false | 394.06% | 16.64% | 0.957 | 0.888 | 0.492 | 33.83% | 1.000 | 0.231 | 645 | 139.327 | PBO 0.231 above 0.200 |
| 2016 shifted | lgbm_ranker_h63_high | false | 619.38% | 20.93% | 1.064 | 0.994 | 0.613 | 34.17% | 1.000 | 0.231 | 236 | 63.063 | PBO 0.231 above 0.200 |

## Readout

- Gate-passing catalog entries in both windows: `lgbm_ranker_h63_low` and `ranker_proxy_h63_low`.
- `lgbm_ranker_h63_medium` and `lgbm_ranker_h63_high` have the highest raw returns, but fail catalog-bucket PBO in both windows.
- `lowvol_sleeve_low` is defensive in concept but fails because turnover increases without enough return improvement versus VOO.
- High-risk entries all fail PBO; they remain diagnostics/challengers, not promoted agent choices.
