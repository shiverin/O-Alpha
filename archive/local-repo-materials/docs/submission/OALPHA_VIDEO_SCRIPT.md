# O(Alpha) Five-Minute Project Showcase Script

**Target duration:** 5:00  
**Format:** Screen recording with voice-over  
**Recording tip:** Use a demo account with onboarding reset. Preload the strategy catalog and historical data so the backtest finishes quickly.

## 0:00-0:25 - Opening and Problem

**On screen**

- Begin on the O(Alpha) landing page.
- Scroll briefly through the hero and product features, then select **Launch App**.

**Narration**

> This is O(Alpha, or O-Alpha), a quantitative strategy research and simulated-execution platform. Many retail tools make automation look simple, but hide the difficult questions: how was a strategy tested, how is risk controlled, and can every portfolio change be explained? O(Alpha turns that process into one guided workflow, from choosing a risk profile to backtesting, running an agent, and auditing its decisions.

## 0:25-1:20 - Guided Onboarding and Backtesting

**On screen**

- Open the onboarding flow.
- Select **Moderate** and briefly flip or inspect one risk-profile card.
- Continue to the strategy selection step.
- Select a compatible catalog strategy and click **Start Backtest**.
- Let the progress indicator and equity curve appear, then click **Accept Backtest**.

**Narration**

> A new user starts with guided onboarding instead of a wall of technical settings. I first choose a risk profile. This filters the strategy catalog so the interface only presents compatible choices and applies sensible defaults for leverage, position limits, exits, and rebalancing.
>
> Before the application unlocks, O(Alpha requires a backtest. The frontend streams progress from the Go backend while it loads aligned daily bars, runs the portfolio simulation, and returns an equity curve with return, Sharpe ratio, drawdown, and trade count. Accepting the result is an explicit validation step; the system does not silently activate a strategy from a default setting.
>
> Architecturally, the interface is built with Next.js and TypeScript, while the Go service owns validation and domain logic. This separation keeps presentation concerns out of the backtest engine and ensures the same backend rules can be reused by the UI, command-line research tools, and automated tests.

## 1:20-2:05 - Overview and Agent Lifecycle

**On screen**

- Arrive at **Overview**.
- Point to agent status, regime, balance/equity chart, strategy controls, execution log, and portfolio allocation.
- Click **Launch Agent** and show the status changing to running.

**Narration**

> The Overview is the operational command centre. At a glance, it combines account value, the equity curve, current strategy, market regime, recent execution events, and portfolio allocation. The dashboard receives persisted snapshots and incremental price events, so positions and unrealised profit and loss can update without rebuilding the entire page.
>
> Launching the agent starts a controlled lifecycle on the backend. Each user can have only one active portfolio agent, preventing two workers from competing over the same account. The worker evaluates market bars, calculates target allocations, reconciles them against current holdings, and records the resulting actions. Start, running, stopping, and stopped are explicit states, which also allows interrupted workers to resume predictably.

## 2:05-3:00 - Configuration and Safety Controls

**On screen**

- Open **Agent Settings**.
- Show the three risk profiles and catalog strategies.
- Toggle **Advanced Tuning**.
- Adjust leverage or max positions briefly, then restore the original value.
- Return to Overview, terminate the agent, then revisit settings to demonstrate that editing becomes available.

**Narration**

> Agent Settings separates simple choices from advanced controls. Most users can stay with a risk profile and recommended strategy. Advanced users can inspect leverage, maximum active positions, stop-loss and take-profit exits, and rebalance cadence.
>
> The important engineering feature is not just that these controls exist, but when they are allowed to change. Settings are locked while an agent is running. The user must stop it before editing, and changing risk profile requires a fresh backtest before the new configuration can be saved. These are fail-closed rules: when state or evidence is missing, O(Alpha blocks the action instead of guessing.
>
> Strategy research follows the same principle. Candidate strategies are evaluated with no-lookahead checks, walk-forward validation, transaction-cost stress, and overfitting measures such as the Deflated Sharpe Ratio and Probability of Backtest Overfitting. A visually impressive return alone is not enough to promote a strategy.

## 3:00-3:50 - Portfolio Observability

**On screen**

- Open **Portfolio**.
- Highlight total asset value and the historical chart.
- Move to composition, key metrics, and the positions table.
- Show a live price-row flash if available.
- Click **Export** but cancel or close the download prompt if needed.

**Narration**

> The Portfolio page explains the resulting state. It shows total asset value over time, cash versus invested exposure, portfolio composition, performance and risk metrics, and every open position. Incoming price events update the affected position, exposure, unrealised profit and loss, and portfolio history together.
>
> Underneath this view is a normalised PostgreSQL data model. Users, accounts, orders, fills, positions, cash-ledger entries, market bars, strategy configurations, and research artifacts are stored separately. Foreign keys and constraints protect relationships, while the append-only cash ledger preserves an auditable history instead of overwriting previous balances. This design makes the displayed portfolio reconstructable and easier to test.

## 3:50-4:30 - Activity and Audit Trail

**On screen**

- Open **Activity**.
- Filter the execution table by one symbol or status.
- Highlight order side, quantity, price, status, timestamps, and **System Alerts**.
- Select **Export CSV**.

**Narration**

> Activity provides the audit trail behind the summary charts. Trades can be filtered by asset and status, inspected with timestamps, quantities, prices, and execution state, and exported as CSV. System alerts expose lifecycle and risk events alongside trades. This is important because an automated system should not only act; it should make its actions inspectable after the fact.

## 4:30-5:00 - Engineering Quality and Close

**On screen**

- Cut to a simple architecture diagram or README section showing: Next.js frontend, Go API, strategy and agent services, PostgreSQL, and market-data provider.
- Briefly show the GitHub Actions page or test-output screenshot.
- End on the Overview with the O(Alpha logo visible.

**Narration**

> O(Alpha is engineered as distinct frontend, API, domain, execution, persistence, and research layers. Automated Go tests cover backtesting, validation, agent behaviour, market-data parsing, and strategy construction. TypeScript checks and browser tests protect key frontend workflows, while GitHub Actions supports repeatable quality checks and scheduled jobs.
>
> The project is deliberately focused on research and simulated execution, not real-money trading. Its goal is to make quantitative automation understandable, testable, and auditable. O(Alpha brings strategy validation, lifecycle safety, portfolio observability, and evidence-driven engineering into one coherent product: a quant in your pocket.

## Recording Checklist

- Reset onboarding before recording.
- Use one consistent demo account and risk profile.
- Pre-run the backtest once to warm caches.
- Avoid presenting illustrative dashboard values as actual investment performance.
- Keep the cursor still while speaking and move only when introducing the next feature.
- Record at 1080p or higher; zoom the browser to keep tables readable.
- Replace long loading periods with a clean cut, not accelerated cursor movement.
- Show test or CI evidence only if it is current and actually passed.
