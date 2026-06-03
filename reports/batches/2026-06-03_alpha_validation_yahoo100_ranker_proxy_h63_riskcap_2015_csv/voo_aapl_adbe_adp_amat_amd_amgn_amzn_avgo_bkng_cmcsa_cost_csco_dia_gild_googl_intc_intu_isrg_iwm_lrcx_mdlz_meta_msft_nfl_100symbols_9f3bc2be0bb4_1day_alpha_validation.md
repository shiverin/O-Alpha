# Alpha Validation Report

- Generated: `2026-06-03T11:54:28Z`
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
| benchmark_ranker_proxy_h63_riskcap_checkpoint | buy_hold | false | 466.10% | 16.45% | 0.946 | 0.883 | 0.461 | 35.67% | 1.000 | 0.267 | 144 | PBO 0.267 above 0.200 |

## benchmark_ranker_proxy_h63_riskcap_checkpoint

- Family: `benchmark_ranker_proxy_h63_riskcap`
- Benchmark: `buy_hold`
- PBO estimated: `true`
- Promotion: `false`
- Rejection reasons:
  - PBO 0.267 above 0.200

### Cost Stress

| Scenario | Return | Ann Ret | Sharpe | Sortino | Calmar | Max DD | Trades | Turnover | Error |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---|
| normal | 466.10% | 16.45% | 0.946 | 0.883 | 0.461 | 35.67% | 144 | 30.821 | - |
| stress_2x | 464.71% | 16.43% | 0.945 | 0.882 | 0.461 | 35.67% | 144 | 30.774 | - |
| stress_3x | 463.32% | 16.40% | 0.943 | 0.881 | 0.460 | 35.67% | 144 | 30.727 | - |

### Walk Forward

| Fold | Train Bars | Test Bars | Train Sharpe | Test Sharpe | Test Calmar | Test Max DD | Error |
|---:|---:|---:|---:|---:|---:|---:|---|
| 0 | 0-756 | 756-1008 | 1.294 | -0.250 | -0.307 | 18.12% | - |
| 1 | 126-882 | 882-1134 | 1.148 | 0.951 | 0.774 | 18.12% | - |
| 2 | 252-1008 | 1008-1260 | 0.961 | 2.295 | 4.713 | 6.42% | - |
| 3 | 378-1134 | 1134-1386 | 1.366 | 0.493 | 0.325 | 35.67% | - |
| 4 | 504-1260 | 1260-1512 | 1.254 | 0.685 | 0.548 | 35.67% | - |
| 5 | 630-1386 | 1386-1638 | 0.683 | 2.434 | 4.468 | 9.94% | - |
| 6 | 756-1512 | 1512-1764 | 0.710 | 2.361 | 6.339 | 5.80% | - |
| 7 | 882-1638 | 1638-1890 | 0.942 | -0.231 | -0.284 | 21.51% | - |
| 8 | 1008-1764 | 1764-2016 | 1.236 | -0.700 | -0.715 | 23.76% | - |
| 9 | 1134-1890 | 1890-2142 | 0.682 | 0.851 | 0.945 | 16.01% | - |
| 10 | 1260-2016 | 2016-2268 | 0.508 | 1.865 | 2.801 | 9.68% | - |
| 11 | 1386-2142 | 2142-2394 | 0.924 | 2.204 | 2.903 | 9.68% | - |
| 12 | 1512-2268 | 2268-2520 | 0.724 | 1.725 | 2.627 | 8.91% | - |
| 13 | 1638-2394 | 2394-2646 | 0.775 | 0.540 | 0.476 | 18.91% | - |
| 14 | 1764-2520 | 2520-2772 | 0.601 | 1.053 | 1.023 | 18.91% | - |

### PBO Diagnostics

| Fold | Train Winner | Train Score | Winner Test Score | Winner Test Rank | Test Winner | Test Winner Score | Variants | Overfit |
|---:|---|---:|---:|---:|---|---:|---:|---:|
| 0 | benchmark_ranker_proxy_h63_riskcap_checkpoint | 1.319 | -0.307 | 3 | benchmark_ranker_proxy_h63_riskcap_vol30 | -0.296 | 5 | false |
| 1 | benchmark_ranker_proxy_h63_riskcap_vol40 | 1.211 | 0.774 | 1 | benchmark_ranker_proxy_h63_riskcap_checkpoint | 0.774 | 5 | false |
| 2 | benchmark_ranker_proxy_h63_riskcap_checkpoint | 0.686 | 4.713 | 1 | benchmark_ranker_proxy_h63_riskcap_checkpoint | 4.713 | 5 | false |
| 3 | benchmark_ranker_proxy_h63_riskcap_vol40 | 0.966 | 0.325 | 3 | benchmark_ranker_proxy_h63_riskcap_strict | 0.325 | 5 | false |
| 4 | benchmark_ranker_proxy_h63_riskcap_vol30 | 0.906 | 0.510 | 5 | benchmark_ranker_proxy_h63_riskcap_vol40 | 0.551 | 5 | true |
| 5 | benchmark_ranker_proxy_h63_riskcap_vol30 | 0.397 | 4.714 | 1 | benchmark_ranker_proxy_h63_riskcap_vol30 | 4.714 | 5 | false |
| 6 | benchmark_ranker_proxy_h63_riskcap_vol30 | 0.425 | 6.243 | 4 | benchmark_ranker_proxy_h63_riskcap_checkpoint | 6.339 | 5 | true |
| 7 | benchmark_ranker_proxy_h63_riskcap_vol30 | 0.608 | -0.245 | 1 | benchmark_ranker_proxy_h63_riskcap_vol30 | -0.245 | 5 | false |
| 8 | benchmark_ranker_proxy_h63_riskcap_vol30 | 0.838 | -0.717 | 4 | benchmark_ranker_proxy_h63_riskcap_vol40 | -0.715 | 5 | true |
| 9 | benchmark_ranker_proxy_h63_riskcap_vol30 | 0.427 | 0.727 | 5 | benchmark_ranker_proxy_h63_riskcap_vol40 | 1.021 | 5 | true |
| 10 | benchmark_ranker_proxy_h63_riskcap_vol40 | 0.298 | 2.925 | 1 | benchmark_ranker_proxy_h63_riskcap_vol40 | 2.925 | 5 | false |
| 11 | benchmark_ranker_proxy_h63_riskcap_vol40 | 0.686 | 3.205 | 1 | benchmark_ranker_proxy_h63_riskcap_vol40 | 3.205 | 5 | false |
| 12 | benchmark_ranker_proxy_h63_riskcap_vol40 | 0.497 | 2.984 | 1 | benchmark_ranker_proxy_h63_riskcap_vol40 | 2.984 | 5 | false |
| 13 | benchmark_ranker_proxy_h63_riskcap_vol40 | 0.565 | 0.500 | 2 | benchmark_ranker_proxy_h63_riskcap_sleeve10 | 0.528 | 5 | false |
| 14 | benchmark_ranker_proxy_h63_riskcap_vol40 | 0.434 | 1.047 | 1 | benchmark_ranker_proxy_h63_riskcap_vol40 | 1.047 | 5 | false |
