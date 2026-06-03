# Alpha Validation Report

- Generated: `2026-06-03T13:59:09Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2019-01-02` to `2026-06-01`
- Bars: `1863`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 242.40% | 18.13% | 0.950 | 0.890 | 0.533 | 34.00% | 0 | 1.000 |
| equal_weight | 373.26% | 23.41% | 1.085 | 1.031 | 0.695 | 33.67% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_exmegacap | buy_hold | false | 343.13% | 22.32% | 1.087 | 1.033 | 0.658 | 33.94% | 1.000 | 0.429 | 88 | PBO 0.429 above 0.200 |

## benchmark_lgbm_ranker_h63_s15_exmegacap

- Family: `benchmark_lgbm_ranker_h63_exmegacap`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.429 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 343.13% | 22.32% | 1.087 | 1.033 | 0.658 | 33.94% | 88 | 15.569 | - |
| stress_2x | 342.47% | 22.30% | 1.085 | 1.032 | 0.657 | 33.94% | 88 | 15.555 | - |
| stress_3x | 341.80% | 22.27% | 1.084 | 1.031 | 0.656 | 33.94% | 88 | 15.542 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.227 | -0.574 | -0.656 | 24.21% | - |
| 1 | 126-882 | 882-1134 | 0.569 | 1.157 | 1.454 | 16.55% | - |
| 2 | 252-1008 | 1008-1260 | 0.514 | 1.855 | 2.845 | 10.28% | - |
| 3 | 378-1134 | 1134-1386 | 0.983 | 2.090 | 2.764 | 10.28% | - |
| 4 | 504-1260 | 1260-1512 | 0.707 | 1.591 | 2.298 | 10.09% | - |
| 5 | 630-1386 | 1386-1638 | 0.684 | 0.843 | 0.904 | 18.44% | - |
| 6 | 756-1512 | 1512-1764 | 0.556 | 1.375 | 1.555 | 18.44% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.853 | -0.656 | 2 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | -0.656 | 4 | false |
| 1 | benchmark_lgbm_ranker_h63_s10_exmegacap | 0.338 | 1.354 | 4 | benchmark_lgbm_ranker_h63_s15_exmegacap | 1.454 | 4 | true |
| 2 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.309 | 2.845 | 2 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | 2.845 | 4 | false |
| 3 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.736 | 2.764 | 2 | benchmark_lgbm_ranker_h63_s15_z125_exmegacap | 2.764 | 4 | false |
| 4 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.485 | 2.298 | 4 | benchmark_lgbm_ranker_h63_s10_z125_exmegacap | 2.623 | 4 | true |
| 5 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.460 | 0.904 | 1 | benchmark_lgbm_ranker_h63_s15_exmegacap | 0.904 | 4 | false |
| 6 | benchmark_lgbm_ranker_h63_s10_exmegacap | 0.392 | 1.421 | 3 | benchmark_lgbm_ranker_h63_s15_exmegacap | 1.555 | 4 | true |
