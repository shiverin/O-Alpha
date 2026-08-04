# Alpha Validation Report

- Generated: `2026-06-03T13:23:54Z`
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
| benchmark_lgbm_ranker_h63_s15_checkpoint | buy_hold | true | 412.48% | 24.75% | 1.159 | 1.091 | 0.727 | 34.06% | 1.000 | 0.143 | 84 | pass |

## benchmark_lgbm_ranker_h63_s15_checkpoint

- Family: `benchmark_lgbm_ranker_h63`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `true`

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 412.48% | 24.75% | 1.159 | 1.091 | 0.727 | 34.06% | 84 | 17.484 | - |
| stress_2x | 411.70% | 24.73% | 1.158 | 1.090 | 0.726 | 34.07% | 84 | 17.467 | - |
| stress_3x | 410.92% | 24.70% | 1.157 | 1.089 | 0.725 | 34.07% | 84 | 17.450 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.245 | -0.502 | -0.617 | 23.60% | - |
| 1 | 126-882 | 882-1134 | 0.707 | 1.385 | 1.834 | 16.55% | - |
| 2 | 252-1008 | 1008-1260 | 0.543 | 2.106 | 3.706 | 9.43% | - |
| 3 | 378-1134 | 1134-1386 | 1.003 | 2.433 | 3.694 | 9.43% | - |
| 4 | 504-1260 | 1260-1512 | 0.817 | 2.111 | 3.536 | 10.00% | - |
| 5 | 630-1386 | 1386-1638 | 0.853 | 0.943 | 1.062 | 18.80% | - |
| 6 | 756-1512 | 1512-1764 | 0.786 | 1.282 | 1.426 | 18.80% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_lgbm_ranker_h63_s15_z125 | 0.893 | -0.617 | 1 | benchmark_lgbm_ranker_h63_s15_z125 | -0.617 | 4 | false |
| 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.462 | 1.834 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 1.834 | 4 | false |
| 2 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.336 | 3.706 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 3.706 | 4 | false |
| 3 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.809 | 3.694 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 3.694 | 4 | false |
| 4 | benchmark_lgbm_ranker_h63_s15_z125 | 0.612 | 3.536 | 3 | benchmark_lgbm_ranker_h63_s10 | 3.582 | 4 | true |
| 5 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.611 | 1.062 | 1 | benchmark_lgbm_ranker_h63_s15_checkpoint | 1.062 | 4 | false |
| 6 | benchmark_lgbm_ranker_h63_s15_checkpoint | 0.570 | 1.426 | 2 | benchmark_lgbm_ranker_h63_s15_z125 | 1.426 | 4 | false |
