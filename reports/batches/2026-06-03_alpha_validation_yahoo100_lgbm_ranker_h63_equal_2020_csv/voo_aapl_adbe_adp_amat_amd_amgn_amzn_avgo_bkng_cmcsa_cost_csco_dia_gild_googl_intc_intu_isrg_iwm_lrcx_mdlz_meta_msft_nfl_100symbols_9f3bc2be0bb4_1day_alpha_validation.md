# Alpha Validation Report

- Generated: `2026-06-03T13:31:07Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2020-01-02` to `2026-06-01`
- Bars: `1611`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 159.21% | 16.08% | 0.831 | 0.783 | 0.473 | 34.00% | 0 | 1.000 |
| equal_weight | 219.23% | 19.92% | 0.953 | 0.904 | 0.594 | 33.53% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_lgbm_ranker_h63_s15_equal_benchmark | equal_weight | true | 275.51% | 23.01% | 1.060 | 1.001 | 0.677 | 34.00% | 1.000 | 0.200 | 74 | pass |

## benchmark_lgbm_ranker_h63_s15_equal_benchmark

- Family: `benchmark_lgbm_ranker_h63_equal`
- Benchmark: `equal_weight`
- PBO estimated: `true`
- Promotion: `true`

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 275.51% | 23.01% | 1.060 | 1.001 | 0.677 | 34.00% | 74 | 12.310 | - |
| stress_2x | 274.97% | 22.98% | 1.059 | 1.000 | 0.676 | 34.01% | 74 | 12.299 | - |
| stress_3x | 274.43% | 22.95% | 1.057 | 0.999 | 0.675 | 34.01% | 74 | 12.288 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.543 | 2.106 | 3.706 | 9.43% | - |
| 1 | 126-882 | 882-1134 | 1.003 | 2.433 | 3.694 | 9.43% | - |
| 2 | 252-1008 | 1008-1260 | 0.817 | 2.111 | 3.536 | 10.00% | - |
| 3 | 378-1134 | 1134-1386 | 0.853 | 0.943 | 1.062 | 18.80% | - |
| 4 | 504-1260 | 1260-1512 | 0.786 | 1.282 | 1.426 | 18.80% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s15_equal_benchmark | 0.336 | 3.706 | 1 | benchmark_lgbm_ranker_h63_s15_equal_benchmark | 3.706 | 4 | false |
| 1 | benchmark_lgbm_ranker_h63_s15_equal_benchmark | 0.809 | 3.694 | 1 | benchmark_lgbm_ranker_h63_s15_equal_benchmark | 3.694 | 4 | false |
| 2 | benchmark_lgbm_ranker_h63_s15_z125_equal_benchmark | 0.612 | 3.536 | 3 | benchmark_lgbm_ranker_h63_s10_equal_benchmark | 3.582 | 4 | true |
| 3 | benchmark_lgbm_ranker_h63_s15_equal_benchmark | 0.611 | 1.062 | 2 | benchmark_lgbm_ranker_h63_s15_z125_equal_benchmark | 1.062 | 4 | false |
| 4 | benchmark_lgbm_ranker_h63_s15_equal_benchmark | 0.570 | 1.426 | 1 | benchmark_lgbm_ranker_h63_s15_equal_benchmark | 1.426 | 4 | false |
