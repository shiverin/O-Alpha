# Alpha Validation Report

- Generated: `2026-06-03T11:25:17Z`
- Symbols: `VOO, AAPL, ADBE, ADP, AMAT, AMD, AMGN, AMZN, AVGO, BKNG, CMCSA, COST, CSCO, DIA, GILD, GOOGL, INTC, INTU, ISRG, IWM, LRCX, MDLZ, META, MSFT, NFLX, NVDA, PEP, QCOM, QQQ, SBUX, SMH, SPY, TSLA, TXN, VTI, XLB, XLE, XLF, XLI, XLK, XLP, XLU, XLV, XLY, HON, ABBV, ABT, ACN, GE, KO, SCHW, AMT, AXP, BA, BAC, BLK, BMY, C, CAT, CI, COP, CRM, CVS, CVX, DE, DIS, ELV, GS, HD, IBM, JNJ, JPM, LIN, LLY, LOW, MA, MCD, MDT, MO, MRK, NEE, NKE, NOW, ORCL, PFE, PG, PLD, PM, RTX, SO, SYK, T, TMO, UNH, UPS, USB, V, VZ, WMT, XOM`
- Timeframe: `1Day`
- Period: `2015-01-02` to `2026-06-01`
- Bars: `2869`

## Notes

- All candidate runs use target-weight execution at next-bar open.
- Promotion is intentionally strict: if PBO cannot be estimated from variants and walk-forward splits, the candidate is not promotable.
- Gross-only performance is not used for promotion; reported metrics are net of the selected cost scenario.

## Benchmarks

| Benchmark | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| buy_hold | 351.93% | 14.17% | 0.837 | 0.790 | 0.417 | 34.00% | 0 | 1.000 |
| equal_weight | 1170.53% | 25.03% | 1.087 | 1.039 | 0.673 | 37.17% | 0 | 1.000 |
| flat_cash | 0.00% | 0.00% | 0.000 | 0.000 | 0.000 | 0.00% | 0 | 0.000 |

## Candidates

| Candidate | Benchmark | Promote | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | DSR | PBO | Trades | Main Reason |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| benchmark_ranker_proxy_h84_checkpoint | buy_hold | false | 461.86% | 16.38% | 0.945 | 0.879 | 0.461 | 35.51% | 1.000 | 0.267 | 147 | PBO 0.267 above 0.200 |

## benchmark_ranker_proxy_h84_checkpoint

- Family: `benchmark_ranker_proxy_h84`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.267 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 461.86% | 16.38% | 0.945 | 0.879 | 0.461 | 35.51% | 147 | 29.623 | - |
| stress_2x | 460.50% | 16.35% | 0.944 | 0.878 | 0.460 | 35.52% | 147 | 29.578 | - |
| stress_3x | 459.15% | 16.33% | 0.943 | 0.877 | 0.460 | 35.52% | 147 | 29.533 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.173 | -0.289 | -0.346 | 18.09% | - |
| 1 | 126-882 | 882-1134 | 1.017 | 0.908 | 0.732 | 18.09% | - |
| 2 | 252-1008 | 1008-1260 | 0.867 | 2.282 | 4.496 | 6.67% | - |
| 3 | 378-1134 | 1134-1386 | 1.228 | 0.507 | 0.339 | 35.51% | - |
| 4 | 504-1260 | 1260-1512 | 1.202 | 0.677 | 0.539 | 35.51% | - |
| 5 | 630-1386 | 1386-1638 | 0.653 | 2.365 | 4.293 | 10.15% | - |
| 6 | 756-1512 | 1512-1764 | 0.700 | 2.210 | 6.001 | 5.77% | - |
| 7 | 882-1638 | 1638-1890 | 0.916 | -0.227 | -0.292 | 20.02% | - |
| 8 | 1008-1764 | 1764-2016 | 1.213 | -0.592 | -0.687 | 21.04% | - |
| 9 | 1134-1890 | 1890-2142 | 0.682 | 0.963 | 1.145 | 15.34% | - |
| 10 | 1260-2016 | 2016-2268 | 0.518 | 2.029 | 3.441 | 8.71% | - |
| 11 | 1386-2142 | 2142-2394 | 0.983 | 2.569 | 3.888 | 8.71% | - |
| 12 | 1512-2268 | 2268-2520 | 0.817 | 1.944 | 3.038 | 9.01% | - |
| 13 | 1638-2394 | 2394-2646 | 0.925 | 0.465 | 0.385 | 19.31% | - |
| 14 | 1764-2520 | 2520-2772 | 0.678 | 0.977 | 0.914 | 19.31% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranker_proxy_h63 | 1.319 | -0.307 | 2 | benchmark_ranker_proxy_h126 | -0.282 | 5 | false |
| 1 | benchmark_ranker_proxy_h63 | 1.211 | 0.774 | 2 | benchmark_ranker_proxy_h126 | 0.792 | 5 | false |
| 2 | benchmark_ranker_proxy_h63 | 0.686 | 4.713 | 1 | benchmark_ranker_proxy_h63 | 4.713 | 5 | false |
| 3 | benchmark_ranker_proxy_h63 | 0.966 | 0.325 | 3 | benchmark_ranker_proxy_h84_checkpoint | 0.339 | 5 | false |
| 4 | benchmark_ranker_proxy_h63 | 0.902 | 0.551 | 1 | benchmark_ranker_proxy_h63 | 0.551 | 5 | false |
| 5 | benchmark_ranker_proxy_h63 | 0.386 | 4.698 | 1 | benchmark_ranker_proxy_h63 | 4.698 | 5 | false |
| 6 | benchmark_ranker_proxy_h63 | 0.420 | 6.339 | 1 | benchmark_ranker_proxy_h63 | 6.339 | 5 | false |
| 7 | benchmark_ranker_proxy_h63 | 0.596 | -0.284 | 1 | benchmark_ranker_proxy_h63 | -0.284 | 5 | false |
| 8 | benchmark_ranker_proxy_h63 | 0.818 | -0.715 | 4 | benchmark_ranker_proxy_h84_checkpoint | -0.687 | 5 | true |
| 9 | benchmark_ranker_proxy_h63 | 0.411 | 1.021 | 5 | benchmark_ranker_proxy_h84_strict | 1.145 | 5 | true |
| 10 | benchmark_ranker_proxy_h84_checkpoint | 0.304 | 3.441 | 1 | benchmark_ranker_proxy_h84_checkpoint | 3.441 | 5 | false |
| 11 | benchmark_ranker_proxy_h84_checkpoint | 0.794 | 3.888 | 1 | benchmark_ranker_proxy_h84_checkpoint | 3.888 | 5 | false |
| 12 | benchmark_ranker_proxy_h84_checkpoint | 0.614 | 3.038 | 3 | benchmark_ranker_proxy_h84_strict | 3.075 | 5 | false |
| 13 | benchmark_ranker_proxy_h84_checkpoint | 0.708 | 0.385 | 5 | benchmark_ranker_proxy_h126 | 0.634 | 5 | true |
| 14 | benchmark_ranker_proxy_h84_strict | 0.444 | 0.918 | 4 | benchmark_ranker_proxy_h126 | 1.163 | 5 | true |
