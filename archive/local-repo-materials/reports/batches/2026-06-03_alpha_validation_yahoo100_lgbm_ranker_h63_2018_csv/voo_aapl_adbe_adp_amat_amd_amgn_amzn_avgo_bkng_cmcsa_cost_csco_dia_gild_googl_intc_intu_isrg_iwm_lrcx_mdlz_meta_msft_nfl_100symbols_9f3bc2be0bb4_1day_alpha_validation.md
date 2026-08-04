# Alpha Validation Report

- Generated: `2026-06-03T13:21:21Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2018-01-02` to `2026-06-01`
- Bars: `2114`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 222.00% | 14.97% | 0.820 | 0.767 | 0.440 | 34.00% | 0 | 1.000 |
| equal_weight | 340.32% | 19.34% | 0.953 | 0.908 | 0.587 | 32.94% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_checkpoint | buy_hold | true | 408.48% | 21.40% | 1.035 | 0.979 | 0.628 | 34.07% | 1.000 | 0.000 | 99 | pass |

## benchmark_lgbm_ranker_h63_s15_checkpoint

- Family: `benchmark_lgbm_ranker_h63`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `true`

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 408.48% | 21.40% | 1.035 | 0.979 | 0.628 | 34.07% | 99 | 18.287 | - |
| stress_2x | 407.56% | 21.38% | 1.034 | 0.978 | 0.627 | 34.07% | 99 | 18.265 | - |
| stress_3x | 406.64% | 21.35% | 1.033 | 0.977 | 0.627 | 34.07% | 99 | 18.243 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.693 | 2.249 | 8.234 | 4.37% | - |
| 1 | 126-882 | 882-1134 | 0.955 | -0.244 | -0.332 | 21.87% | - |
| 2 | 252-1008 | 1008-1260 | 1.253 | -0.550 | -0.663 | 24.09% | - |
| 3 | 378-1134 | 1134-1386 | 0.711 | 1.552 | 2.166 | 16.57% | - |
| 4 | 504-1260 | 1260-1512 | 0.543 | 2.275 | 4.194 | 9.56% | - |
| 5 | 630-1386 | 1386-1638 | 1.155 | 2.579 | 3.891 | 9.56% | - |
| 6 | 756-1512 | 1512-1764 | 0.918 | 2.251 | 3.886 | 9.81% | - |
| 7 | 882-1638 | 1638-1890 | 0.978 | 0.937 | 1.030 | 18.81% | - |
| 8 | 1008-1764 | 1764-2016 | 0.909 | 1.372 | 1.504 | 18.81% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.442 | 8.234 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 8.234 | 4 | false |
| 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.662 | -0.332 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | -0.332 | 4 | false |
| 2 | benchmark_lgbm_ranker_h63_s15_z125 | 0.906 | -0.663 | 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | -0.663 | 4 | false |
| 3 | benchmark_lgbm_ranker_h63_s15_z125 | 0.470 | 2.166 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 2.166 | 4 | false |
| 4 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.336 | 4.194 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 4.194 | 4 | false |
| 5 | benchmark_lgbm_ranker_h63_s15_z125 | 0.965 | 3.891 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 3.891 | 4 | false |
| 6 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.704 | 3.886 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 3.886 | 4 | false |
| 7 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.735 | 1.030 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 1.030 | 4 | false |
| 8 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.698 | 1.504 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 1.504 | 4 | false |
