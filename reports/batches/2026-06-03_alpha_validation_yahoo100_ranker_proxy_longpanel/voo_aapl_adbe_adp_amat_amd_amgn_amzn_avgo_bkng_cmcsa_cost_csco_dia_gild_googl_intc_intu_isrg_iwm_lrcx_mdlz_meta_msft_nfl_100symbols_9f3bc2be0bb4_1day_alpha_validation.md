# Alpha Validation Report

- Generated: `2026-06-03T09:30:44Z`
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
| benchmark_ranker_proxy_checkpoint | buy_hold | false | 433.00% | 15.84% | 0.915 | 0.862 | 0.462 | 34.26% | 1.000 | 0.600 | 202 | PBO 0.600 above 0.200 |

## benchmark_ranker_proxy_checkpoint

- Family: `benchmark_ranker_proxy`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.600 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 433.00% | 15.84% | 0.915 | 0.862 | 0.462 | 34.26% | 202 | 21.951 | - |
| stress_2x | 431.97% | 15.82% | 0.914 | 0.861 | 0.462 | 34.27% | 202 | 21.925 | - |
| stress_3x | 430.95% | 15.80% | 0.913 | 0.861 | 0.461 | 34.27% | 202 | 21.899 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.111 | -0.359 | -0.392 | 19.39% | - |
| 1 | 126-882 | 882-1134 | 1.043 | 0.748 | 0.568 | 19.39% | - |
| 2 | 252-1008 | 1008-1260 | 0.770 | 2.351 | 5.086 | 6.09% | - |
| 3 | 378-1134 | 1134-1386 | 1.190 | 0.433 | 0.271 | 34.26% | - |
| 4 | 504-1260 | 1260-1512 | 1.150 | 0.694 | 0.579 | 34.26% | - |
| 5 | 630-1386 | 1386-1638 | 0.599 | 2.385 | 4.206 | 10.22% | - |
| 6 | 756-1512 | 1512-1764 | 0.685 | 2.020 | 5.874 | 5.08% | - |
| 7 | 882-1638 | 1638-1890 | 0.891 | -0.309 | -0.368 | 20.73% | - |
| 8 | 1008-1764 | 1764-2016 | 1.168 | -0.598 | -0.680 | 22.65% | - |
| 9 | 1134-1890 | 1890-2142 | 0.618 | 0.878 | 1.006 | 15.65% | - |
| 10 | 1260-2016 | 2016-2268 | 0.458 | 1.823 | 2.673 | 9.61% | - |
| 11 | 1386-2142 | 2142-2394 | 0.877 | 2.459 | 3.481 | 9.61% | - |
| 12 | 1512-2268 | 2268-2520 | 0.711 | 2.036 | 3.460 | 8.65% | - |
| 13 | 1638-2394 | 2394-2646 | 0.771 | 0.719 | 0.688 | 18.43% | - |
| 14 | 1764-2520 | 2520-2772 | 0.639 | 1.043 | 1.013 | 18.43% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranker_proxy_medium | 1.032 | -0.409 | 4 | benchmark_ranker_proxy_fast | -0.308 | 4 | true |
| 1 | benchmark_ranker_proxy_medium | 1.116 | 0.511 | 4 | benchmark_ranker_proxy_fast | 0.732 | 4 | true |
| 2 | benchmark_ranker_proxy_fast | 0.551 | 4.262 | 3 | benchmark_ranker_proxy_checkpoint | 5.086 | 4 | true |
| 3 | benchmark_ranker_proxy_fast | 0.890 | 0.312 | 1 | benchmark_ranker_proxy_fast | 0.312 | 4 | false |
| 4 | benchmark_ranker_proxy_fast | 0.846 | 0.691 | 1 | benchmark_ranker_proxy_fast | 0.691 | 4 | false |
| 5 | benchmark_ranker_proxy_fast | 0.384 | 4.616 | 1 | benchmark_ranker_proxy_fast | 4.616 | 4 | false |
| 6 | benchmark_ranker_proxy_fast | 0.446 | 5.732 | 3 | benchmark_ranker_proxy_checkpoint | 5.874 | 4 | true |
| 7 | benchmark_ranker_proxy_fast | 0.608 | -0.443 | 3 | benchmark_ranker_proxy_slow | -0.347 | 4 | true |
| 8 | benchmark_ranker_proxy_fast | 0.820 | -0.759 | 4 | benchmark_ranker_proxy_checkpoint | -0.680 | 4 | true |
| 9 | benchmark_ranker_proxy_fast | 0.392 | 0.951 | 2 | benchmark_ranker_proxy_checkpoint | 1.006 | 4 | false |
| 10 | benchmark_ranker_proxy_medium | 0.258 | 2.617 | 3 | benchmark_ranker_proxy_fast | 2.820 | 4 | true |
| 11 | benchmark_ranker_proxy_checkpoint | 0.664 | 3.481 | 2 | benchmark_ranker_proxy_medium | 3.767 | 4 | false |
| 12 | benchmark_ranker_proxy_checkpoint | 0.498 | 3.460 | 3 | benchmark_ranker_proxy_slow | 3.642 | 4 | true |
| 13 | benchmark_ranker_proxy_checkpoint | 0.509 | 0.688 | 1 | benchmark_ranker_proxy_checkpoint | 0.688 | 4 | false |
| 14 | benchmark_ranker_proxy_slow | 0.422 | 0.854 | 4 | benchmark_ranker_proxy_fast | 1.318 | 4 | true |
