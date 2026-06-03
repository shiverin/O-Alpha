# Research Progress Report

Generated: `2026-06-03T07:49:10Z`

## Summary

The official Go alpha-validation harness found no promoted alpha yet.

Two important infrastructure fixes landed first:

- Portfolio backtest pending targets are now one-shot next-bar orders, covered by regression test.
- Walk-forward test folds now use the train window as warmup context and trim metrics to the OOS segment, covered by regression test.

This makes reports generated before this warmup fix stale for PBO/walk-forward claims.

## Data

Expanded daily Alpaca/IEX raw coverage from 42 to 102 symbols over `2020-07-27` through `2026-06-01`.

Coverage command:

```bash
cd backend && set -a; source ../.env.local; set +a; go run ./cmd/ml-meta-research -mode inventory -timeframe 1Day -from 2020-07-27 -to 2026-06-01 -output-dir ../reports/batches/2026-06-03_inventory_daily_coverage_expanded
```

## Corrected Results

`xsec_momentum`, expanded 102-symbol universe:

- Report: `reports/batches/2026-06-03_alpha_validation_xsec_expanded_warmupfix/`
- Return 46.93%, Sharpe 0.602, PBO 0.500.
- Verdict: rejected.

`composite_momentum`, expanded 102-symbol universe:

- Reports: `reports/batches/2026-06-03_alpha_validation_composite_expanded_warmupfix/`, `reports/batches/2026-06-03_alpha_validation_composite_expanded_warmupfix_test189/`, `reports/batches/2026-06-03_alpha_validation_composite_expanded_warmupfix_shifted_2021/`
- Primary: return 144.30%, Sharpe 1.004, PBO 0.250.
- Denser split: PBO 0.222.
- Shifted 2021: return 88.63%, Sharpe 0.799, PBO 0.286.
- Verdict: parked, not promoted.

`benchmark_tsmom`, expanded 102-symbol universe:

- Reports: `reports/batches/2026-06-03_alpha_validation_benchmark_tsmom_expanded_warmupfix/`, `reports/batches/2026-06-03_alpha_validation_benchmark_tsmom_expanded_warmupfix_test189/`, `reports/batches/2026-06-03_alpha_validation_benchmark_tsmom_expanded_warmupfix_shifted_2021/`
- Primary: return 156.90%, Sharpe 1.025, Calmar 0.849, max drawdown 20.75%, PBO 0.750.
- Denser split: PBO 0.444.
- Shifted 2021: return 90.64%, Sharpe 0.793, PBO 0.143, but underperforms VOO.
- Verdict: promising research checkpoint only, not promoted.

`benchmark_tsmom_blend`, expanded 102-symbol universe:

- Reports: `reports/batches/2026-06-03_alpha_validation_benchmark_tsmom_blend_expanded/`, `reports/batches/2026-06-03_alpha_validation_benchmark_tsmom_blend_expanded_test189/`, `reports/batches/2026-06-03_alpha_validation_benchmark_tsmom_blend_expanded_shifted_2021/`
- Primary: return 153.09%, Sharpe 1.030, PBO 0.250.
- Denser split: PBO 0.444.
- Shifted 2021: return 78.60%, Sharpe 0.714, PBO 0.000, but underperforms VOO.
- Verdict: rejected as start-date fragile.

`benchmark_lowvol`, expanded 102-symbol universe:

- Report: `reports/batches/2026-06-03_alpha_validation_benchmark_lowvol_expanded/`
- Primary: return 109.85%, Sharpe 0.927, PBO 0.000.
- Verdict: rejected because it does not improve return/risk enough versus VOO.

`benchmark_reversal`, expanded 102-symbol universe:

- Reports: `reports/batches/2026-06-03_alpha_validation_benchmark_reversal_expanded/`, `reports/batches/2026-06-03_alpha_validation_benchmark_reversal_expanded_test189/`, `reports/batches/2026-06-03_alpha_validation_benchmark_reversal_expanded_shifted_2021/`
- Primary: return 144.38%, Sharpe 0.999, PBO 0.500.
- Denser split: PBO 0.667.
- Shifted 2021: return 104.15%, Sharpe 0.873, PBO 0.143, but underperforms VOO.
- Verdict: rejected.

Deeper-history backfill:

- Command shape: `INGEST_RUN_ONCE=true INGEST_FORCE_BACKFILL=true INGEST_INTERVAL=1Day INGEST_LOOKBACK=100000h ... go run ./cmd/ingest/main.go`
- Result: most symbols still start around `2020-07-27`; this Alpaca/IEX dataset is not enough for robust pre-2020 validation.

## Next

Use the expanded universe and PBO diagnostics, but pivot away from more momentum-only tweaks unless pre-2020 history is ingested.
