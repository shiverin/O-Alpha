# Agent Strategy Catalog

Generated: 2026-06-03

The portfolio agent can now instantiate researched strategies by catalog key via
`portfolio.NewStrategyFromCatalog(...)` or
`PortfolioAgentManager.StartCatalogPortfolioAgent(...)`.

Every catalog strategy is paper/research-only by default and stamps these fields
into `PortfolioOutput.EngineMetadata`:

- `agent_strategy_key`
- `agent_strategy_name`
- `agent_strategy_family`
- `agent_strategy_risk_profile`
- `agent_strategy_deployment_status`
- `agent_strategy_promoted_checkpoint`
- `agent_strategy_paper_only`

## Risk Buckets

| Risk | Strategy key | Status | Notes |
|---|---|---|---|
| low | `lgbm_ranker_h63_low` | conservative variant | 5% learned-ranker sleeve, stricter score threshold; derived from the promoted h63 checkpoint but not separately promoted. |
| low | `ranker_proxy_h63_low` | conservative variant | 8% deterministic h63 sleeve. |
| low | `lowvol_sleeve_low` | rejected diagnostic | Defensive low-vol sleeve; useful comparator, not promoted alpha. |
| medium | `lgbm_ranker_h63_medium` | promoted research checkpoint | Current best: 15% h63 learned-ranker sleeve, top 3, 63-bar rebalance. |
| medium | `ranker_proxy_h63_medium` | promoted research checkpoint | Deterministic h63 proxy; primary window promoted but shifted PBO was weaker. |
| medium | `ranked_sleeve_medium` | rejected diagnostic | Risk-budgeted ranked sleeve; one shifted window promoted, full 2015 window rejected. |
| high | `lgbm_ranker_h63_high` | experimental variant | 25% learned-ranker sleeve, wider top-k, lower score threshold; paper comparison only. |
| high | `benchmark_tsmom_high` | rejected diagnostic | Larger ETF/broad time-series momentum sleeve; useful challenger, failed PBO. |
| high | `composite_momentum_high` | rejected diagnostic | Higher active-weight composite momentum; raw return was attractive, PBO failed. |

## Usage

```go
cfg := portfolio.DefaultStrategyCatalogConfig()
strategy, spec, err := portfolio.NewStrategyFromCatalog(
    "lgbm_ranker_h63_medium",
    symbols,
    cfg,
)
```

or start the paper portfolio worker directly:

```go
worker, spec, err := manager.StartCatalogPortfolioAgent(
    ctx,
    "paper-h63-medium",
    "lgbm_ranker_h63_medium",
    symbols,
    "1Day",
    100000,
    portfolio.DefaultStrategyCatalogConfig(),
    nil,
)
```

## Boundary

This catalog does not approve live trading. The strongest strategy is still
blocked from deployment-grade claims by the unresolved survivorship-aware/PIT
adjusted-price data issue documented in `docs/BLOCKERS.md`.
