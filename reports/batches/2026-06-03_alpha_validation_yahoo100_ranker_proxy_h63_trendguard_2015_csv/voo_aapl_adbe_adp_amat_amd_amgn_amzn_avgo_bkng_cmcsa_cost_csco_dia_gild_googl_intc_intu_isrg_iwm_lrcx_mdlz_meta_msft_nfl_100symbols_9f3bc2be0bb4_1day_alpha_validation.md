# Alpha Validation Report

- Generated: `2026-06-03T12:02:59Z`
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
| benchmark_ranker_proxy_h63_trendguard_checkpoint | buy_hold | false | 460.66% | 16.36% | 0.949 | 0.888 | 0.459 | 35.67% | 1.000 | 0.267 | 128 | PBO 0.267 above 0.200 |

## benchmark_ranker_proxy_h63_trendguard_checkpoint

- Family: `benchmark_ranker_proxy_h63_trendguard`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.267 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 460.66% | 16.36% | 0.949 | 0.888 | 0.459 | 35.67% | 128 | 26.237 | - |
| stress_2x | 459.46% | 16.33% | 0.947 | 0.887 | 0.458 | 35.67% | 128 | 26.201 | - |
| stress_3x | 458.25% | 16.31% | 0.946 | 0.885 | 0.457 | 35.67% | 128 | 26.165 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.221 | -0.250 | -0.307 | 18.12% | - |
| 1 | 126-882 | 882-1134 | 1.142 | 0.933 | 0.743 | 18.12% | - |
| 2 | 252-1008 | 1008-1260 | 0.961 | 2.319 | 4.836 | 6.13% | - |
| 3 | 378-1134 | 1134-1386 | 1.364 | 0.453 | 0.281 | 35.67% | - |
| 4 | 504-1260 | 1260-1512 | 1.251 | 0.602 | 0.445 | 35.67% | - |
| 5 | 630-1386 | 1386-1638 | 0.661 | 2.433 | 4.861 | 8.61% | - |
| 6 | 756-1512 | 1512-1764 | 0.667 | 2.361 | 6.339 | 5.80% | - |
| 7 | 882-1638 | 1638-1890 | 0.904 | -0.231 | -0.284 | 21.51% | - |
| 8 | 1008-1764 | 1764-2016 | 1.208 | -0.681 | -0.733 | 22.62% | - |
| 9 | 1134-1890 | 1890-2142 | 0.646 | 0.991 | 1.138 | 15.99% | - |
| 10 | 1260-2016 | 2016-2268 | 0.492 | 1.997 | 3.060 | 9.68% | - |
| 11 | 1386-2142 | 2142-2394 | 0.973 | 2.344 | 3.205 | 9.68% | - |
| 12 | 1512-2268 | 2268-2520 | 0.773 | 1.888 | 2.984 | 8.91% | - |
| 13 | 1638-2394 | 2394-2646 | 0.868 | 0.590 | 0.533 | 18.91% | - |
| 14 | 1764-2520 | 2520-2772 | 0.697 | 1.104 | 1.084 | 18.91% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranker_proxy_h63_trendguard_slow_defensive | 1.281 | -0.307 | 3 | benchmark_ranker_proxy_h63_trendguard_voo_only | -0.307 | 5 | false |
| 1 | benchmark_ranker_proxy_h63_trendguard_checkpoint | 1.278 | 0.743 | 4 | benchmark_ranker_proxy_h63_trendguard_voo_only | 0.798 | 5 | true |
| 2 | benchmark_ranker_proxy_h63_trendguard_checkpoint | 0.686 | 4.836 | 1 | benchmark_ranker_proxy_h63_trendguard_checkpoint | 4.836 | 5 | false |
| 3 | benchmark_ranker_proxy_h63_trendguard_voo_only | 0.974 | 0.323 | 1 | benchmark_ranker_proxy_h63_trendguard_voo_only | 0.323 | 5 | false |
| 4 | benchmark_ranker_proxy_h63_trendguard_voo_only | 0.908 | 0.515 | 1 | benchmark_ranker_proxy_h63_trendguard_voo_only | 0.515 | 5 | false |
| 5 | benchmark_ranker_proxy_h63_trendguard_voo_only | 0.389 | 4.551 | 5 | benchmark_ranker_proxy_h63_trendguard_checkpoint | 4.861 | 5 | true |
| 6 | benchmark_ranker_proxy_h63_trendguard_voo_only | 0.412 | 6.339 | 3 | benchmark_ranker_proxy_h63_trendguard_checkpoint | 6.339 | 5 | false |
| 7 | benchmark_ranker_proxy_h63_trendguard_voo_only | 0.587 | -0.284 | 2 | benchmark_ranker_proxy_h63_trendguard_slow_defensive | -0.284 | 5 | false |
| 8 | benchmark_ranker_proxy_h63_trendguard_voo_only | 0.805 | -0.728 | 2 | benchmark_ranker_proxy_h63_trendguard_light | -0.723 | 5 | false |
| 9 | benchmark_ranker_proxy_h63_trendguard_voo_only | 0.399 | 1.134 | 4 | benchmark_ranker_proxy_h63_trendguard_slow_defensive | 1.167 | 5 | true |
| 10 | benchmark_ranker_proxy_h63_trendguard_slow_defensive | 0.303 | 3.010 | 4 | benchmark_ranker_proxy_h63_trendguard_voo_only | 3.151 | 5 | true |
| 11 | benchmark_ranker_proxy_h63_trendguard_light | 0.753 | 3.205 | 3 | benchmark_ranker_proxy_h63_trendguard_slow_defensive | 3.205 | 5 | false |
| 12 | benchmark_ranker_proxy_h63_trendguard_light | 0.554 | 2.984 | 2 | benchmark_ranker_proxy_h63_trendguard_slow_defensive | 2.984 | 5 | false |
| 13 | benchmark_ranker_proxy_h63_trendguard_light | 0.626 | 0.582 | 2 | benchmark_ranker_proxy_h63_trendguard_voo_only | 0.631 | 5 | false |
| 14 | benchmark_ranker_proxy_h63_trendguard_voo_only | 0.487 | 1.194 | 1 | benchmark_ranker_proxy_h63_trendguard_voo_only | 1.194 | 5 | false |
