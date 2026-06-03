# Alpha Validation Report

- Generated: `2026-06-03T12:08:24Z`
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
| benchmark_ranker_proxy_h63_liquidity_checkpoint | buy_hold | false | 392.39% | 15.04% | 0.867 | 0.815 | 0.424 | 35.43% | 1.000 | 0.467 | 143 | PBO 0.467 above 0.200 |

## benchmark_ranker_proxy_h63_liquidity_checkpoint

- Family: `benchmark_ranker_proxy_h63_liquidity`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.467 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 392.39% | 15.04% | 0.867 | 0.815 | 0.424 | 35.43% | 143 | 25.161 | - |
| stress_2x | 391.29% | 15.01% | 0.865 | 0.814 | 0.424 | 35.43% | 143 | 25.127 | - |
| stress_3x | 390.20% | 14.99% | 0.864 | 0.813 | 0.423 | 35.43% | 143 | 25.092 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.034 | -0.490 | -0.471 | 21.22% | - |
| 1 | 126-882 | 882-1134 | 0.986 | 0.434 | 0.279 | 21.22% | - |
| 2 | 252-1008 | 1008-1260 | 0.654 | 2.246 | 3.660 | 8.44% | - |
| 3 | 378-1134 | 1134-1386 | 1.079 | 0.548 | 0.387 | 35.43% | - |
| 4 | 504-1260 | 1260-1512 | 1.067 | 0.691 | 0.561 | 35.43% | - |
| 5 | 630-1386 | 1386-1638 | 0.618 | 2.373 | 4.111 | 10.51% | - |
| 6 | 756-1512 | 1512-1764 | 0.628 | 2.256 | 6.070 | 5.55% | - |
| 7 | 882-1638 | 1638-1890 | 0.885 | -0.473 | -0.463 | 23.39% | - |
| 8 | 1008-1764 | 1764-2016 | 1.209 | -0.770 | -0.755 | 24.92% | - |
| 9 | 1134-1890 | 1890-2142 | 0.627 | 1.015 | 1.228 | 15.49% | - |
| 10 | 1260-2016 | 2016-2268 | 0.443 | 1.920 | 3.014 | 9.63% | - |
| 11 | 1386-2142 | 2142-2394 | 0.855 | 2.269 | 3.141 | 9.63% | - |
| 12 | 1512-2268 | 2268-2520 | 0.687 | 1.822 | 2.884 | 8.91% | - |
| 13 | 1638-2394 | 2394-2646 | 0.765 | 0.469 | 0.405 | 18.55% | - |
| 14 | 1764-2520 | 2520-2772 | 0.648 | 1.057 | 1.052 | 18.55% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranker_proxy_h63_liquidity_500m | 1.246 | -0.401 | 1 | benchmark_ranker_proxy_h63_liquidity_500m | -0.401 | 5 | false |
| 1 | benchmark_ranker_proxy_h63_liquidity_500m | 1.161 | 0.492 | 1 | benchmark_ranker_proxy_h63_liquidity_500m | 0.492 | 5 | false |
| 2 | benchmark_ranker_proxy_h63_liquidity_500m | 0.544 | 4.277 | 1 | benchmark_ranker_proxy_h63_liquidity_500m | 4.277 | 5 | false |
| 3 | benchmark_ranker_proxy_h63_liquidity_500m | 0.855 | 0.331 | 5 | benchmark_ranker_proxy_h63_liquidity_checkpoint | 0.387 | 5 | true |
| 4 | benchmark_ranker_proxy_h63_liquidity_500m | 0.791 | 0.521 | 5 | benchmark_ranker_proxy_h63_liquidity_1500m | 0.589 | 5 | true |
| 5 | benchmark_ranker_proxy_h63_liquidity_500m | 0.356 | 4.357 | 1 | benchmark_ranker_proxy_h63_liquidity_500m | 4.357 | 5 | false |
| 6 | benchmark_ranker_proxy_h63_liquidity_500m | 0.378 | 5.508 | 5 | benchmark_ranker_proxy_h63_liquidity_2000m | 6.601 | 5 | true |
| 7 | benchmark_ranker_proxy_h63_liquidity_checkpoint | 0.560 | -0.463 | 5 | benchmark_ranker_proxy_h63_liquidity_500m | -0.330 | 5 | true |
| 8 | benchmark_ranker_proxy_h63_liquidity_2000m | 0.819 | -0.808 | 4 | benchmark_ranker_proxy_h63_liquidity_500m | -0.711 | 5 | true |
| 9 | benchmark_ranker_proxy_h63_liquidity_2000m | 0.405 | 0.990 | 4 | benchmark_ranker_proxy_h63_liquidity_checkpoint | 1.228 | 5 | true |
| 10 | benchmark_ranker_proxy_h63_liquidity_2000m | 0.272 | 2.899 | 4 | benchmark_ranker_proxy_h63_liquidity_500m | 3.016 | 5 | true |
| 11 | benchmark_ranker_proxy_h63_liquidity_2000m | 0.659 | 3.164 | 3 | benchmark_ranker_proxy_h63_liquidity_500m | 3.454 | 5 | false |
| 12 | benchmark_ranker_proxy_h63_liquidity_500m | 0.494 | 3.055 | 3 | benchmark_ranker_proxy_h63_liquidity_2000m | 3.275 | 5 | false |
| 13 | benchmark_ranker_proxy_h63_liquidity_500m | 0.588 | 0.467 | 3 | benchmark_ranker_proxy_h63_liquidity_1500m | 0.482 | 5 | false |
| 14 | benchmark_ranker_proxy_h63_liquidity_500m | 0.454 | 1.083 | 1 | benchmark_ranker_proxy_h63_liquidity_500m | 1.083 | 5 | false |
