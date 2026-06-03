# Alpha Validation Report

- Generated: `2026-06-03T08:59:24Z`
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
| benchmark_ranked_sleeve_checkpoint | buy_hold | false | 306.34% | 18.20% | 0.964 | 0.912 | 0.551 | 33.01% | 1.000 | 0.222 | 394 | PBO 0.222 above 0.200 |

## benchmark_ranked_sleeve_checkpoint

- Family: `benchmark_ranked_sleeve`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.222 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 306.34% | 18.20% | 0.964 | 0.912 | 0.551 | 33.01% | 394 | 62.636 | - |
| stress_2x | 303.87% | 18.11% | 0.960 | 0.909 | 0.549 | 33.02% | 396 | 62.409 | - |
| stress_3x | 301.41% | 18.03% | 0.956 | 0.905 | 0.546 | 33.03% | 396 | 62.184 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 0.885 | 1.811 | 4.701 | 6.01% | - |
| 1 | 126-882 | 882-1134 | 1.063 | -0.343 | -0.391 | 20.20% | - |
| 2 | 252-1008 | 1008-1260 | 1.327 | -0.642 | -0.629 | 23.74% | - |
| 3 | 378-1134 | 1134-1386 | 0.642 | 0.932 | 1.048 | 15.09% | - |
| 4 | 504-1260 | 1260-1512 | 0.468 | 1.421 | 2.050 | 10.04% | - |
| 5 | 630-1386 | 1386-1638 | 0.806 | 2.356 | 3.698 | 10.04% | - |
| 6 | 756-1512 | 1512-1764 | 0.717 | 2.239 | 4.270 | 8.67% | - |
| 7 | 882-1638 | 1638-1890 | 0.816 | 0.793 | 0.741 | 18.41% | - |
| 8 | 1008-1764 | 1764-2016 | 0.712 | 1.223 | 1.163 | 18.41% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranked_sleeve_checkpoint | 0.601 | 4.701 | 2 | benchmark_ranked_sleeve_conservative | 5.432 | 4 | false |
| 1 | benchmark_ranked_sleeve_checkpoint | 0.752 | -0.391 | 2 | benchmark_ranked_sleeve_slow | -0.195 | 4 | false |
| 2 | benchmark_ranked_sleeve_checkpoint | 0.973 | -0.629 | 1 | benchmark_ranked_sleeve_checkpoint | -0.629 | 4 | false |
| 3 | benchmark_ranked_sleeve_conservative | 0.486 | 1.143 | 1 | benchmark_ranked_sleeve_conservative | 1.143 | 4 | false |
| 4 | benchmark_ranked_sleeve_conservative | 0.309 | 2.922 | 1 | benchmark_ranked_sleeve_conservative | 2.922 | 4 | false |
| 5 | benchmark_ranked_sleeve_slow | 0.771 | 4.544 | 1 | benchmark_ranked_sleeve_slow | 4.544 | 4 | false |
| 6 | benchmark_ranked_sleeve_slow | 0.573 | 5.074 | 1 | benchmark_ranked_sleeve_slow | 5.074 | 4 | false |
| 7 | benchmark_ranked_sleeve_slow | 0.600 | 0.607 | 4 | benchmark_ranked_sleeve_conservative | 0.781 | 4 | true |
| 8 | benchmark_ranked_sleeve_slow | 0.543 | 0.656 | 4 | benchmark_ranked_sleeve_conservative | 1.527 | 4 | true |
