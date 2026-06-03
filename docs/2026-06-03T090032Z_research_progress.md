# 2026-06-03T09:00:32Z Research Progress

## What changed

- Added research-only Yahoo adjusted daily ingest and `alpha-research -source` selection.
- Backfilled the existing 102-symbol panel to Yahoo adjusted daily bars from 2015-01-01 through 2026-06-01.
- Extended the VOO-core composite sleeve with risk-adjusted edge weighting and a turnover band.
- Registered `benchmark_ranked_sleeve` and `sector_ranked_sleeve` with PBO variants.

## Key reports

- Full 2015 old-family checkpoint: `reports/batches/2026-06-03_alpha_validation_yahoo100_longpanel_checkpoint/`
- Defensive family: `reports/batches/2026-06-03_alpha_validation_yahoo100_defensive_longpanel/`
- Ranked sleeves full panel: `reports/batches/2026-06-03_alpha_validation_yahoo100_ranked_sleeves_longpanel/`
- Ranked sleeve shifted 2016: `reports/batches/2026-06-03_alpha_validation_yahoo100_ranked_sleeve_shifted_2016/`
- Ranked sleeve shifted 2017: `reports/batches/2026-06-03_alpha_validation_yahoo100_ranked_sleeves_shifted_2017/`
- Ranked sleeve shifted 2018: `reports/batches/2026-06-03_alpha_validation_yahoo100_ranked_sleeve_shifted_2018/`

## Result

The best current checkpoint is `benchmark_ranked_sleeve_checkpoint`.

- Full 2015-2026: return 474.41%, Sharpe 0.944, PBO 0.400, promote false.
- Shifted 2016-2026: return 450.44%, Sharpe 0.995, PBO 0.308, promote false.
- Shifted 2017-2026: return 373.98%, Sharpe 0.988, PBO 0.182, promote true.
- Shifted 2018-2026: return 306.34%, Sharpe 0.964, PBO 0.222, promote false.

Verdict: parked as promising checkpoint. It is not robust enough to call durable alpha.

## Caveats

- The Yahoo panel is longer but still handpicked/current-universe data.
- `XLC` and `XLRE` were excluded from the 2015 full panel because of later inception dates.
- Equal-weight current large caps returned 1170.53% in the 2015 run, which highlights survivorship/selection bias.

## Verification

- `go test ./...`
- `go build ./...`
- `research/ml/.venv/bin/python -m py_compile research/ml/*.py`
- `git diff --check`
