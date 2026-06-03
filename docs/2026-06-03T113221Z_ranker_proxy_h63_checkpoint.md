# Ranker Proxy H63 Checkpoint

Timestamp: 2026-06-03T11:32:21Z

## Summary

The h63 benchmark-funded ranker proxy produced the first official primary-window
promotion on this branch, but the shifted official audit failed PBO. Treat it as
a serious research checkpoint, not as a deployment-ready alpha.

## Official Primary Run

Command:

```bash
cd backend
symbols=$(jq -r '.symbols | join(",")' ../reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_longpanel/*_alpha_validation.json)
GOCACHE=/tmp/oalpha-gocache go run ./cmd/alpha-research \
  -bars-csv ../reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2026/bars.csv \
  -symbols "$symbols" \
  -strategies benchmark_ranker_proxy_h63 \
  -timeframe 1Day -from 2015-01-01 -to 2026-06-01 \
  -train-bars 756 -test-bars 252 -step-bars 126 -min-trades 30 \
  -output-dir ../reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_h63_longpanel_csv
```

Report:

`reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_h63_longpanel_csv/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.md`

Result:

- promote: `true`
- total return: `488.98%`
- buy-hold return: `351.93%`
- Sharpe: `0.964`
- DSR: `1.000`
- PBO: `0.200`
- trades: `142`
- main reason: `pass`

## Shifted Official Audit

Command:

```bash
cd backend
symbols=$(jq -r '.symbols | join(",")' ../reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_longpanel/*_alpha_validation.json)
GOCACHE=/tmp/oalpha-gocache go run ./cmd/alpha-research \
  -bars-csv ../reports/batches/2026-06-03_yahoo100_ml_manifest/export_2015_2026/bars.csv \
  -symbols "$symbols" \
  -strategies benchmark_ranker_proxy_h63 \
  -timeframe 1Day -from 2016-01-01 -to 2026-06-01 \
  -train-bars 756 -test-bars 252 -step-bars 126 -min-trades 30 \
  -output-dir ../reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_h63_shifted_2016_csv
```

Report:

`reports/batches/2026-06-03_alpha_validation_yahoo100_ranker_proxy_h63_shifted_2016_csv/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.md`

Result:

- promote: `false`
- total return: `451.32%`
- buy-hold return: `348.99%`
- Sharpe: `1.006`
- DSR: `1.000`
- PBO: `0.231`
- trades: `129`
- first rejection reason: `PBO 0.231 above 0.200`

## Interpretation

The economics are directionally strong: the active sleeve improves total return,
Sharpe, Sortino, and Calmar versus VOO on both windows. The statistical
stability is not yet sufficient because shifted PBO breaches the hard gate. The
next research unit should focus on PBO stability, not more raw return.

## Broad Suite Check

Primary broad suite:

`reports/batches/2026-06-03_alpha_validation_yahoo100_benchmark_funded_suite_2015_csv/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.md`

Result: h63 is the only promoted candidate in the broad benchmark-funded suite.

Shifted broad suite:

`reports/batches/2026-06-03_alpha_validation_yahoo100_benchmark_funded_suite_2016_csv/voo_aapl_adbe_adp_amat_amd_amgn_amzn_avgo_bkng_cmcsa_cost_csco_dia_gild_googl_intc_intu_isrg_iwm_lrcx_mdlz_meta_msft_nfl_100symbols_9f3bc2be0bb4_1day_alpha_validation.md`

Result: no candidate promotes on the shifted 2016 window; h63 fails on PBO
`0.231`.
