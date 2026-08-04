# O(Alpha) Solution Suitability Evaluation Pack

Date: 2026-07-05

Status: private working artifact. This file sits under `docs/`, which is ignored by the repository's current `.gitignore`.

## Integrity Note

This evaluation separates real evidence from simulated evidence.

Subagent and AI-assisted persona reviews are useful for product planning, but they are not actual users, interviews, surveys, or public usability tests. They should be reported as simulated user focus group / proxy persona review only.

Because the team has budget and privacy constraints around launching the app publicly, this document also includes ready-to-field private research protocols for real participants. Those methods should not be selected as "used" until real people complete them.

## What The Team Can Honestly Select Today

| Evaluation method | Status | Select now? | Evidence in this pack |
| --- | --- | --- | --- |
| Expert / self evaluation | Completed | Yes | Expert review of product docs, frontend flows, safety behavior, and validation story. |
| Cognitive walkthrough / heuristic evaluation | Completed | Yes | Task walkthrough across onboarding, dashboard, agent launch/terminate, settings, and evidence discovery. |
| Simulated user focus group | Completed | Yes, if labelled simulated | Ten proxy personas; six came from parallel subagent reviews, four were added as local synthetic proxy personas to complete the group. |
| Actual user focus group / interview | Not conducted | No | Private interview protocol is included for later use. |
| Usability testing with potential users on low-fidelity artifacts | Not conducted | No | Low-fidelity Figma/wireframe protocol is included for later use. |
| Survey of potential users | Not fielded | No | 50-person survey plan and item bank are included for later use. |
| Usability testing with potential users with high-fidelity artifacts | Not conducted with real users | No | Working-prototype usability protocol is included for later use. |

Recommended form selection today:

- Expert / self evaluation
- Cognitive walkthrough / heuristic evaluation
- Simulated user focus group

Do not select the "actual user", "potential users", or "survey" options until they have been run with real participants and anonymized notes are stored.

## Evidence Reviewed

The review was grounded in current repo artifacts and source files:

- `README.md`
- `docs/submission/OALPHA_README.md`
- `frontend/src/components/app/OnboardingOverlay.tsx`
- `frontend/src/app/app/dashboard/page.tsx`
- `frontend/src/app/app/agent-settings/page.tsx`
- `frontend/src/components/app/BeginnerTour.tsx`
- `frontend/src/components/sections/dashboard/BalanceCard.tsx`
- `frontend/src/components/sections/dashboard/StrategyControls.tsx`
- `frontend/src/app/app/activity/page.tsx`
- `reports/batches/2026-06-03_alpha_validation_agent_catalog_buckets_summary/agent_catalog_bucket_comparison.md`

This was not a live public test. No real participant account, brokerage account, financial identity, or private user data was used.

## Method 1: Expert / Self Evaluation

### Scope

The expert review assessed whether O(Alpha) is suitable as a safe, understandable paper-trading and quant research product for beginner users, risk-aware users, evaluators, and technical reviewers.

### Strengths

- The product scope is appropriately constrained to research and paper trading.
- The onboarding flow requires risk selection, compatible strategy choice, a backtest run, and explicit acceptance before unlocking the dashboard.
- The active-agent safety model is strong: settings are locked while a portfolio agent is starting or running, and the backend now remains the source of truth.
- The docs make an important distinction between user-facing backtests and research validation artifacts.
- Dashboard areas cover the right mental model: status, launch/terminate, portfolio value, allocation, execution log, alerts, and activity.

### Risks

- The UI does not yet surface enough validation evidence. DSR, PBO, benchmark comparison, report provenance, paper-only limitations, and model-artifact requirements are mostly in docs or reports rather than strategy cards.
- Some demo or fallback states can look too real, especially realistic P&L values and "real-time" language when backend state is missing.
- Beginner users may over-trust a five-year backtest unless the app explains that historical simulation is not proof or investment advice.
- Some copy is stylized rather than operationally precise, for example "Synchronizing Matrix" instead of "Saving".
- Strategy names and metrics such as Sharpe, Max DD, h63, LGBM, active sleeve, and regime need in-product explanations.

### Expert Verdict

O(Alpha) is suitable as a controlled paper-trading prototype and research-to-paper workflow demo. It is not yet suitable to present as a live-trading product or as self-evidently validated without stronger in-product evidence surfaces.

## Method 2: Cognitive Walkthrough / Heuristic Evaluation

### Walkthrough Tasks

1. Create/sign in to an account.
2. Choose a risk profile.
3. Choose a compatible catalog strategy.
4. Run and accept the onboarding backtest.
5. Open the dashboard and identify portfolio state.
6. Launch a paper portfolio agent.
7. Inspect execution log, allocation, positions, alerts, and activity.
8. Try to edit settings while a portfolio agent is active.
9. Stop the agent.
10. Edit settings after the agent is stopped.

### Heuristic Findings

| Heuristic | Finding | Severity | Recommendation |
| --- | --- | --- | --- |
| Visibility of system status | Dashboard shows active/terminate state, but freshness, last evaluation time, stream connection, and latest bar date are not prominent. | Medium | Add last updated, active run ID, last heartbeat, latest bar date, and API/stream status. |
| Match between system and real world | Finance and quant terms are accurate but not beginner-friendly. | Medium | Add plain-language tooltips for Sharpe, Max DD, PBO, DSR, active sleeve, regime, LGBM, and strategy cadence. |
| User control and freedom | Launch/terminate and settings lock are clear and safety-oriented. | Low | Keep the lock; add a short explanation of why settings cannot change mid-run. |
| Error prevention | Backend and frontend settings guardrails are strong. | Low | Continue treating backend checks as the safety boundary; preserve fail-closed behavior. |
| Recognition rather than recall | Users must leave the app to verify report artifacts and promotion reasoning. | High | Add evidence drawers/cards that link strategy choices to report paths and validation status. |
| Trust and reassurance | Paper-only scope is documented, but not repeated enough near risky-feeling actions. | High | Add "Paper only / no brokerage orders" badges near Launch Agent, P&L, backtest acceptance, and settings save. |
| Empty and fallback states | Realistic fallback data may mask missing backend state. | High | Replace realistic fallback values with explicit demo, unavailable, or empty states. |
| Accessibility and clarity | Static code review cannot prove keyboard/focus behavior or contrast compliance. | Medium | Run browser accessibility checks and keyboard-only smoke tests before final submission. |

### Walkthrough Verdict

The core task path is coherent and safety-aware. The biggest suitability gap is not workflow completion; it is user trust calibration. Users can do the actions, but they need clearer proof, clearer limits, and clearer status.

## Method 3: Simulated User Focus Group

### Method

This was a simulated 10-person proxy focus group. It should not be described as actual user research.

Composition:

- Personas 1-6 came from parallel subagent reviews grounded in repo docs and frontend code.
- Personas 7-10 were added locally as synthetic proxy personas to complete the planned 10-person group.
- No participant was a real person.
- No survey was fielded.
- No public launch or private user data was used.

### Simulated Persona Summary

| Persona | Background | Ease | Confidence | Transparency | Primary concern |
| --- | --- | ---: | ---: | ---: | --- |
| 1. Beginner paper trader | Knows ETFs and basic diversification, new to quant terms. | 4.0 | 3.0 | 3.0 | Paper/demo/live distinction is not repeated near key actions. |
| 2. Risk-aware retail investor | Comfortable with drawdowns and wants evidence before automation. | 3.0 | 3.0 | 2.0 | Strategy evidence and artifact provenance are hidden from UI. |
| 3. Quant hobbyist | Understands Sharpe, walk-forward tests, and overfitting risk. | 4.0 | 3.0 | 2.0 | Wants DSR, PBO, benchmark, costs, and report links on cards. |
| 4. Skeptical product manager | Evaluates onboarding, proof, and product story. | 3.0 | 3.0 | 3.0 | Proof is mostly in docs rather than obvious in product. |
| 5. Reliability-focused engineer | Looks for lifecycle safety, freshness, and failure modes. | 4.0 | 3.0 | 3.0 | Needs explicit stale/error/heartbeat indicators. |
| 6. Finance student | Learning systematic trading and backtest limitations. | 3.5 | 3.0 | 2.5 | Terms and acceptance semantics need education. |
| 7. Compliance-minded reviewer | Checks claims, disclaimers, and audit trail. | 3.0 | 2.5 | 2.0 | Wants stronger "paper-only", no advice, and artifact lineage labels. |
| 8. Active trader / power user | Wants fast controls, trade explanations, and exports. | 3.0 | 3.0 | 2.5 | Wants trade rationale and position-change drilldowns. |
| 9. Cautious nontechnical saver | Wants safety and plain language before trying automation. | 3.5 | 2.5 | 2.0 | May over-trust the backtest without clearer warnings. |
| 10. Privacy-conscious evaluator | Wants to review without exposing personal financial data. | 3.0 | 3.0 | 3.0 | Needs a clean seeded demo script and anonymized sample data. |

Ratings use a 1-5 scale, where 5 is best. These are simulated ratings, not survey statistics.

### Simulated Theme Counts

These counts summarize proxy persona reactions only.

| Theme | Simulated count |
| --- | ---: |
| Wanted clearer paper-only / no live orders messaging near Launch, P&L, and backtest acceptance. | 10/10 |
| Wanted validation evidence visible in the app, not only in docs or reports. | 9/10 |
| Found at least one quant term or strategy name unclear. | 8/10 |
| Wanted dashboard freshness, stale-state, or backend health indicators. | 7/10 |
| Wanted trade rationale or "why did this happen?" drilldowns. | 6/10 |
| Found the active-agent settings lock reassuring. | 10/10 |

### Key Persona Findings

1. Beginner paper trader
   - Can complete onboarding and launch a paper agent.
   - Confused by Sharpe, Max DD, active sleeve, h63, and LGBM.
   - Recommendation: add tooltips and visible paper-only badges.

2. Risk-aware retail investor
   - Likes the safety boundaries and audit categories.
   - Cannot fully verify strategy quality from the UI alone.
   - Recommendation: add evidence-aware strategy cards and trade audit drawers.

3. Quant hobbyist
   - Understands the workflow but wants deeper validation metadata.
   - Concerned if fallback catalog data appears to contradict report evidence.
   - Recommendation: show Promote, DSR, PBO, benchmark, OOS trades, costs, and report path.

4. Skeptical product manager
   - Sees a coherent product story but not enough proof in the screens.
   - Some copy reduces clarity.
   - Recommendation: add a proof panel after onboarding and use direct operational labels.

5. Reliability-focused engineer
   - Trusts the running-agent settings lock.
   - Wants last refresh, active run ID, heartbeat, stream state, and failure reason.
   - Recommendation: add dashboard health metadata.

6. Finance student
   - Finds onboarding approachable and educational.
   - May treat "Accept Backtest" as "this is good" instead of "I understand simulation limits".
   - Recommendation: change acceptance copy to an explicit historical simulation acknowledgement.

7. Compliance-minded reviewer
   - Wants user-facing claims to match artifact-backed evidence.
   - Looks for no investment advice, paper-only, and no brokerage execution language.
   - Recommendation: add compliance-safe disclaimers near high-risk actions.

8. Active trader / power user
   - Wants quicker inspection of trades, alerts, and allocation changes.
   - Needs richer export and drilldown behavior.
   - Recommendation: add a trade detail drawer with target weight, previous position, fill price, and related alert.

9. Cautious nontechnical saver
   - Likes guardrails but feels the product is technical.
   - Needs reassurance that no real money is connected.
   - Recommendation: add plain-language safety mode and empty states before showing realistic values.

10. Privacy-conscious evaluator
   - Wants a reviewable seeded demo that does not require personal financial data.
   - Needs a stable path for reviewers to reproduce onboarding and dashboard state.
   - Recommendation: create a deterministic demo script with seeded users and paper-only sample portfolios.

## Not Yet Conducted: Actual User Focus Group / Interview

### Private Protocol To Field Later

Recommended sample: 10 real participants.

Suggested participant mix:

- 3 beginner investors or paper traders.
- 2 risk-aware retail investors.
- 2 technically inclined users.
- 1 product/evaluator persona.
- 1 compliance or risk reviewer.
- 1 finance student or learner.

Session length: 45-60 minutes.

Privacy setup:

- Use seeded demo accounts only.
- Do not collect brokerage credentials.
- Do not collect real portfolio data.
- Do not ask for personal financial position details.
- Record only anonymized notes, task outcomes, and ratings.

Interview tasks:

1. Explain what you think O(Alpha) does after reading the first screen.
2. Pick a risk profile and say why.
3. Run and accept a backtest.
4. Explain what the backtest does and does not prove.
5. Launch the paper agent.
6. Identify whether any real money is involved.
7. Find the latest trade or activity.
8. Try to edit settings while the agent is running.
9. Stop the agent and edit one setting.
10. Explain what evidence would make you trust or distrust the strategy.

Questions:

- What felt safe?
- What felt risky?
- What terms were unclear?
- Where did you expect to see proof?
- Would you trust this as a paper-trading tool? Why or why not?
- What would need to change before you recommended it to a beginner?

Output template:

| Participant | Segment | Task success | Main confusion | Trust rating | Safety rating | Recommendation |
| --- | --- | --- | --- | ---: | ---: | --- |

## Not Yet Conducted: Low-Fidelity Usability Testing

### Protocol To Field Later

Artifact: Figma or static wireframes for onboarding, strategy card evidence drawer, dashboard status panel, and settings lock state.

Recommended sample: 5-8 participants.

Tasks:

1. Point to where you would start.
2. Choose a risk profile.
3. Identify which strategy is safest for your profile.
4. Find proof that a strategy is paper-only and evidence-backed.
5. Explain what happens if settings are edited while an agent is running.

Success metrics:

- Task completion without moderator help.
- Misinterpretation of paper-only scope.
- Correct explanation of backtest limits.
- Ability to find strategy evidence.
- Confidence rating after each task.

## Not Yet Fielded: Survey Of Potential Users

### 50-Person Survey Plan

Target sample: 50 potential users, recruited privately through classmates, colleagues, university channels, or paper-trading communities that permit research requests.

Eligibility:

- Age 18 or above.
- Interested in investing, paper trading, systematic strategies, or finance education.
- Comfortable reviewing a prototype or screenshots.
- Does not need to share personal portfolio data.

Recommended survey sections:

1. Background and experience.
2. Trust and safety expectations.
3. Reaction to paper-only quant assistant concept.
4. Comprehension of strategy evidence.
5. Feature priority ranking.
6. Open-ended concerns.

Example survey items:

| Item | Response type |
| --- | --- |
| I understand that this product is for paper trading only. | 1-5 agreement |
| I would want to see validation evidence before launching an automated strategy. | 1-5 agreement |
| A five-year backtest alone is enough for me to trust a strategy. | 1-5 agreement, reverse-coded |
| I can explain what "paper trading" means. | 1-5 agreement |
| I can explain what "maximum drawdown" means. | 1-5 agreement |
| I would find strategy evidence links useful. | 1-5 agreement |
| I would use a guided beginner mode. | 1-5 agreement |
| I am concerned about accidentally placing real trades. | 1-5 agreement |
| Rank the most important dashboard information: P&L, positions, trades, alerts, strategy proof, data freshness. | Ranking |
| What would make you trust this product more? | Open text |
| What would make you stop using this product? | Open text |
| What term or screen felt most confusing? | Open text |

Analysis plan:

- Report medians and distributions, not just averages.
- Segment beginners versus experienced users.
- Summarize open-text themes with representative anonymized quotes.
- Do not claim statistical significance from 50 respondents.

## Not Yet Conducted: High-Fidelity Usability Testing

### Working Prototype Protocol To Field Later

Recommended sample: 8-12 real participants.

Test setup:

- Local or private deployed environment.
- Seeded test account and deterministic paper portfolio.
- No brokerage integration.
- No real personal financial data.
- Logging limited to task events, errors, and anonymized ratings.

Tasks:

1. Complete onboarding.
2. Run and accept a backtest.
3. Launch the paper agent.
4. Confirm whether any real money is at risk.
5. Locate portfolio value, allocation, alerts, and recent trades.
6. Attempt to change settings while running and explain what happens.
7. Stop the agent.
8. Save a valid settings change.
9. Find strategy evidence or explain where you expected it.

Metrics:

- Task success.
- Time on task.
- Number of moderator assists.
- Errors and backtracks.
- Single Ease Question after each task.
- SUS score after the session.
- Trust rating before and after.
- Safety comprehension: can the participant state that the product is paper-only?

Pass/fail thresholds:

- At least 80% of participants correctly identify paper-only scope.
- At least 80% complete onboarding without moderator intervention.
- 100% are prevented from changing settings while an agent is running.
- At least 70% can find or correctly request strategy evidence.

## Prioritized Product Recommendations

1. Add persistent "Paper only / no brokerage orders" labels near Launch Agent, P&L, onboarding backtest acceptance, and settings.
2. Add evidence-aware strategy cards or a drawer with promotion status, DSR, PBO, benchmark, OOS trades, cost stress, report path, and model artifact requirements.
3. Replace realistic fallback values with explicit demo, unavailable, or empty states.
4. Add dashboard health metadata: latest bar date, last portfolio snapshot, last agent evaluation, active run ID, heartbeat, and stream connection state.
5. Add a trade audit drawer showing triggering strategy, target weight change, previous/current position, simulated fill price, timestamp, and related alert.
6. Add tooltips or small "learn more" copy for Sharpe, Max DD, DSR, PBO, active sleeve, regime, LGBM, h63, leverage, stop-loss, and take-profit.
7. Rename stylized operational copy to direct labels: "Save settings", "Saving", "Settings locked while agent is running", and "Run backtest".
8. Change "Accept Backtest" to an explicit acknowledgement that the result is historical paper simulation and not investment advice.
9. Create a deterministic private demo script with seeded data so evaluators can test without exposing private financial information.
10. Add keyboard and screen-reader checks for onboarding, Launch/Terminate, settings controls, disabled controls, and error messages.

## Suggested Next Evidence To Collect

The next highest-value research step is a private high-fidelity usability test with 5 real participants using seeded accounts. It gives stronger evidence than more simulated personas while preserving privacy and avoiding a public launch.

After that, run the 50-person survey if the team needs broader market-suitability evidence.

