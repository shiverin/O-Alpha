# Alpha Validation Report

- Generated: `2026-06-03T11:22:57Z`
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
| benchmark_ranker_proxy_h63_checkpoint | buy_hold | true | 488.98% | 16.86% | 0.964 | 0.901 | 0.473 | 35.67% | 1.000 | 0.200 | 142 | pass |

## benchmark_ranker_proxy_h63_checkpoint

- Family: `benchmark_ranker_proxy_h63`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `true`

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 488.98% | 16.86% | 0.964 | 0.901 | 0.473 | 35.67% | 142 | 31.400 | - |
| stress_2x | 487.53% | 16.83% | 0.963 | 0.900 | 0.472 | 35.67% | 142 | 31.351 | - |
| stress_3x | 486.08% | 16.81% | 0.962 | 0.899 | 0.471 | 35.67% | 142 | 31.302 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.294 | -0.250 | -0.307 | 18.12% | - |
| 1 | 126-882 | 882-1134 | 1.148 | 0.951 | 0.774 | 18.12% | - |
| 2 | 252-1008 | 1008-1260 | 0.961 | 2.295 | 4.713 | 6.42% | - |
| 3 | 378-1134 | 1134-1386 | 1.366 | 0.493 | 0.325 | 35.67% | - |
| 4 | 504-1260 | 1260-1512 | 1.254 | 0.688 | 0.551 | 35.67% | - |
| 5 | 630-1386 | 1386-1638 | 0.683 | 2.449 | 4.698 | 9.48% | - |
| 6 | 756-1512 | 1512-1764 | 0.711 | 2.361 | 6.339 | 5.80% | - |
| 7 | 882-1638 | 1638-1890 | 0.944 | -0.231 | -0.284 | 21.51% | - |
| 8 | 1008-1764 | 1764-2016 | 1.238 | -0.700 | -0.715 | 23.76% | - |
| 9 | 1134-1890 | 1890-2142 | 0.684 | 0.904 | 1.021 | 16.01% | - |
| 10 | 1260-2016 | 2016-2268 | 0.509 | 1.919 | 2.925 | 9.68% | - |
| 11 | 1386-2142 | 2142-2394 | 0.942 | 2.344 | 3.205 | 9.68% | - |
| 12 | 1512-2268 | 2268-2520 | 0.742 | 1.888 | 2.984 | 8.91% | - |
| 13 | 1638-2394 | 2394-2646 | 0.836 | 0.560 | 0.500 | 18.91% | - |
| 14 | 1764-2520 | 2520-2772 | 0.666 | 1.070 | 1.047 | 18.91% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranker_proxy_h63_checkpoint | 1.319 | -0.307 | 2 | benchmark_ranker_proxy_h126 | -0.282 | 5 | false |
| 1 | benchmark_ranker_proxy_h63_checkpoint | 1.211 | 0.774 | 2 | benchmark_ranker_proxy_h126 | 0.792 | 5 | false |
| 2 | benchmark_ranker_proxy_h63_checkpoint | 0.686 | 4.713 | 1 | benchmark_ranker_proxy_h63_checkpoint | 4.713 | 5 | false |
| 3 | benchmark_ranker_proxy_h63_checkpoint | 0.966 | 0.325 | 3 | benchmark_ranker_proxy_h84 | 0.339 | 5 | false |
| 4 | benchmark_ranker_proxy_h63_checkpoint | 0.902 | 0.551 | 2 | benchmark_ranker_proxy_h63_strict | 0.552 | 5 | false |
| 5 | benchmark_ranker_proxy_h63_checkpoint | 0.386 | 4.698 | 2 | benchmark_ranker_proxy_h63_strict | 4.703 | 5 | false |
| 6 | benchmark_ranker_proxy_h63_checkpoint | 0.420 | 6.339 | 1 | benchmark_ranker_proxy_h63_checkpoint | 6.339 | 5 | false |
| 7 | benchmark_ranker_proxy_h63_checkpoint | 0.596 | -0.284 | 1 | benchmark_ranker_proxy_h63_checkpoint | -0.284 | 5 | false |
| 8 | benchmark_ranker_proxy_h63_checkpoint | 0.818 | -0.715 | 2 | benchmark_ranker_proxy_h84 | -0.687 | 5 | false |
| 9 | benchmark_ranker_proxy_h63_checkpoint | 0.411 | 1.021 | 4 | benchmark_ranker_proxy_h84 | 1.145 | 5 | true |
| 10 | benchmark_ranker_proxy_h84 | 0.304 | 3.441 | 1 | benchmark_ranker_proxy_h84 | 3.441 | 5 | false |
| 11 | benchmark_ranker_proxy_h84 | 0.794 | 3.888 | 1 | benchmark_ranker_proxy_h84 | 3.888 | 5 | false |
| 12 | benchmark_ranker_proxy_h84 | 0.614 | 3.038 | 1 | benchmark_ranker_proxy_h84 | 3.038 | 5 | false |
| 13 | benchmark_ranker_proxy_h84 | 0.708 | 0.385 | 5 | benchmark_ranker_proxy_h126 | 0.634 | 5 | true |
| 14 | benchmark_ranker_proxy_h84 | 0.442 | 0.914 | 5 | benchmark_ranker_proxy_h126 | 1.163 | 5 | true |
