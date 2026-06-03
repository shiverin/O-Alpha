# Alpha Validation Report

- Generated: `2026-06-03T08:57:49Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2016-01-04` to `2026-06-01`
- Bars: `2617`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 348.99% | 15.57% | 0.898 | 0.842 | 0.458 | 34.00% | 0 | 1.000 |
| equal_weight | 932.56% | 25.22% | 1.112 | 1.059 | 0.720 | 35.05% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_ranked_sleeve_checkpoint | buy_hold | false | 450.44% | 17.86% | 0.995 | 0.934 | 0.518 | 34.45% | 1.000 | 0.308 | 507 | PBO 0.308 above 0.200 |

## benchmark_ranked_sleeve_checkpoint

- Family: `benchmark_ranked_sleeve`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.308 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 450.44% | 17.86% | 0.995 | 0.934 | 0.518 | 34.45% | 507 | 96.483 | - |
| stress_2x | 446.25% | 17.77% | 0.991 | 0.931 | 0.516 | 34.46% | 507 | 96.029 | - |
| stress_3x | 442.10% | 17.68% | 0.987 | 0.927 | 0.513 | 34.47% | 507 | 95.577 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.751 | 2.586 | 5.362 | 6.34% | - |
| 1 | 126-882 | 882-1134 | 1.182 | 0.666 | 0.529 | 34.45% | - |
| 2 | 252-1008 | 1008-1260 | 1.211 | 0.843 | 0.762 | 34.45% | - |
| 3 | 378-1134 | 1134-1386 | 0.720 | 2.323 | 4.425 | 10.20% | - |
| 4 | 504-1260 | 1260-1512 | 0.784 | 1.923 | 5.090 | 5.97% | - |
| 5 | 630-1386 | 1386-1638 | 1.023 | -0.399 | -0.425 | 20.86% | - |
| 6 | 756-1512 | 1512-1764 | 1.278 | -0.675 | -0.643 | 24.04% | - |
| 7 | 882-1638 | 1638-1890 | 0.658 | 0.867 | 0.966 | 14.98% | - |
| 8 | 1008-1764 | 1764-2016 | 0.477 | 1.572 | 2.311 | 9.98% | - |
| 9 | 1134-1890 | 1890-2142 | 0.807 | 2.275 | 3.581 | 9.98% | - |
| 10 | 1260-2016 | 2016-2268 | 0.696 | 1.785 | 2.976 | 9.70% | - |
| 11 | 1386-2142 | 2142-2394 | 0.820 | 0.649 | 0.603 | 18.01% | - |
| 12 | 1512-2268 | 2268-2520 | 0.640 | 1.216 | 1.184 | 18.01% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranked_sleeve_conservative | 0.655 | 3.934 | 4 | benchmark_ranked_sleeve_checkpoint | 5.362 | 4 | true |
| 1 | benchmark_ranked_sleeve_conservative | 1.015 | 0.358 | 3 | benchmark_ranked_sleeve_checkpoint | 0.529 | 4 | true |
| 2 | benchmark_ranked_sleeve_conservative | 0.933 | 0.910 | 1 | benchmark_ranked_sleeve_conservative | 0.910 | 4 | false |
| 3 | benchmark_ranked_sleeve_checkpoint | 0.427 | 4.425 | 2 | benchmark_ranked_sleeve_conservative | 4.944 | 4 | false |
| 4 | benchmark_ranked_sleeve_conservative | 0.510 | 5.453 | 1 | benchmark_ranked_sleeve_conservative | 5.453 | 4 | false |
| 5 | benchmark_ranked_sleeve_checkpoint | 0.683 | -0.425 | 2 | benchmark_ranked_sleeve_slow | -0.291 | 4 | false |
| 6 | benchmark_ranked_sleeve_checkpoint | 0.883 | -0.643 | 1 | benchmark_ranked_sleeve_checkpoint | -0.643 | 4 | false |
| 7 | benchmark_ranked_sleeve_conservative | 0.462 | 1.022 | 1 | benchmark_ranked_sleeve_conservative | 1.022 | 4 | false |
| 8 | benchmark_ranked_sleeve_conservative | 0.290 | 3.094 | 1 | benchmark_ranked_sleeve_conservative | 3.094 | 4 | false |
| 9 | benchmark_ranked_sleeve_slow | 0.635 | 4.426 | 1 | benchmark_ranked_sleeve_slow | 4.426 | 4 | false |
| 10 | benchmark_ranked_sleeve_slow | 0.454 | 3.946 | 1 | benchmark_ranked_sleeve_slow | 3.946 | 4 | false |
| 11 | benchmark_ranked_sleeve_slow | 0.571 | 0.428 | 4 | benchmark_ranked_sleeve_conservative | 0.673 | 4 | true |
| 12 | benchmark_ranked_sleeve_slow | 0.494 | 0.673 | 4 | benchmark_ranked_sleeve_conservative | 1.652 | 4 | true |
