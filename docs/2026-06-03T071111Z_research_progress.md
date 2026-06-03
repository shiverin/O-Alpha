# Research Progress Report

Timestamp: 2026-06-03T07:11:11Z

## Focus

Moved the composite benchmark-core momentum sleeve from Python research artifacts into the official Go `cmd/alpha-research` validation harness required by `AGENTS.md`.

## Implemented

- Added `backend/internal/alpha/momentum/composite_momentum.go`.
- Registered `composite_momentum` in `backend/internal/research/alphavalidation/strategies.go`.
- Added PBO variant factories for the composite momentum family.
- Added multi-symbol `buy_hold` benchmark availability so VOO-core strategies can be benchmarked against VOO, not only equal-weight.
- Added tests for composite momentum behavior and validation factory registration.

## Official Harness Results

Primary report:

- Path: `reports/batches/2026-06-03_alpha_validation_composite_momentum/voo_aapl_amd_amzn_avgo_bac_cost_crm_dia_googl_hd_iwm_jnj_jpm_lly_ma_meta_msft_nflx_nvda_orcl_pg_qqq_smh_spy_tsla_unh_v_vti_wmt_xlb_xlc_xle_xlf_xli_xlk_xlp_xlre_xlu_xlv_xly_xom_1day_alpha_validation.md`
- Period: 2020-07-27 to 2026-06-01.
- Symbols: 42.
- Strategy return: 146.20%.
- VOO buy-hold return: 135.68%.
- Sharpe: 1.011.
- DSR: 1.000.
- PBO: 0.250.
- Trades: 4,420.
- Promotion: false.
- First rejection reason: PBO 0.250 above 0.200.

Robustness report:

- Path: `reports/batches/2026-06-03_alpha_validation_composite_momentum_test189/`
- Validation change: test-bars 189, step-bars 63.
- PBO improved to 0.222 but still failed the 0.200 gate.

Shifted-date report:

- Path: `reports/batches/2026-06-03_alpha_validation_composite_momentum_shifted_2021/`
- Period: 2021-01-04 to 2026-06-01.
- Strategy return: 89.71%.
- VOO buy-hold return: 106.03%.
- PBO: 0.286.
- Promotion: false.

## Verdict

PARKED. The strategy is promising but not promotable. The official gate rejected it in all harness runs, and the shifted-date audit exposed instability versus VOO buy-and-hold.

## Verification

- `go test ./internal/alpha/momentum ./internal/research/alphavalidation`
- `go test ./...`
- `go build ./...`
- `git diff --check`

All passed before the harness runs.

## Next Research Step

Search for a lower-turnover composite family or a broader-universe effect that can pass PBO <= 0.20 on shifted date ranges. Do not call this alpha until the promotion gate returns true.
