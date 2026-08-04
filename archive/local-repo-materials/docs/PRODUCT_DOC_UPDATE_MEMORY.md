# O(Alpha) Product Documentation Update Memory

Use this as the source-context note for future Codex sessions that need to update
the O(Alpha) product documentation.

## Current Source Of Truth

- Latest completed DOCX:
  `/Users/shizhen/Downloads/Quant in your pocket_copy_toc_updated.docx`
- Previous important intermediate:
  `/Users/shizhen/Downloads/Quant in your pocket_copy_database_updated.docx`
- Regenerated database images:
  `/Users/shizhen/Downloads/oalpha_doc_screenshots/db_schema_updated/`

Always start from the latest completed DOCX above unless the user explicitly
provides a newer file.

## User Preference

- Preserve the old document's exact format and tone.
- Do not rewrite the whole document unless asked.
- Make local, surgical edits in the existing structure.
- For review requests, first point out exactly what is stale/wrong and where to
  touch up.
- For implementation requests, edit the DOCX directly and provide a new copy.
- Use screenshots plus frontend/backend explanation tables where the existing
  feature-section format uses that pattern.
- Keep technical claims grounded in the actual codebase/migrations, not guesses.

## Documentation Skill Workflow

For DOCX edits, use the Documents skill.

Preferred QA:

1. Edit with `python-docx`.
2. Export/render through Microsoft Word when page numbers matter.
3. Verify text with `python-docx` and/or `textutil -convert html`.
4. Verify DOCX package integrity with `unzip -t`.
5. Render/check key pages visually when possible.

Known environment note:

- The bundled `render_docx.py` currently fails locally because `pdf2image` is
  missing.
- Microsoft Word is installed and `docx2pdf` works for DOCX -> PDF export.
- `pypdf` and `pymupdf` were installed with `python3 -m pip install --user` for
  PDF text/page extraction and page image rendering.

## Latest App/Codebase Facts Captured In The Doc

O(Alpha) is now a full-stack research and paper-trading system, not just a
frontend proof of concept.

Backend highlights:

- Authenticated Go/Gin REST APIs.
- Portfolio backtests and NDJSON streaming backtests.
- One active `PORTFOLIO_CATALOG` portfolio agent per user.
- Paper fills, positions, cash ledger entries, portfolio snapshots, agent run
  state, and system alerts persisted in Postgres.
- Dashboard state is driven from backend state, not local/mock state.
- HMM regime runtime state is written into `agent_runs.parameters.runtime_state`
  for dashboard display.
- LGBM ranker catalog strategies require artifacts and fail closed if artifacts
  are absent.

Frontend highlights:

- Onboarding is now: welcome -> risk-profile select -> catalog backtest
  acceptance.
- Risk cards select only; explicit Next button advances.
- Onboarding backtest uses `POST /api/v1/backtest/stream` with
  `PORTFOLIO_CATALOG`.
- The onboarding graph progressively builds from NDJSON progress events.
- User must accept the backtest before onboarding completion.
- Dashboard strategy controls are minimal/display-only: risk profile and asset
  universe only.
- Portfolio allocation reads `/api/v1/user/portfolio/positions`.
- Activity reads persisted fills and system alerts.

## Current Database Documentation Facts

The migrated schema has 23 logical application tables:

- Baseline/runtime:
  `users`, `sessions`, `assets`, `accounts`, `agent_settings`,
  `strategy_configs`, `backtest_runs`, `backtest_trades`, `agent_runs`,
  `orders`, `fills`, `positions`, `cash_ledger`, `system_alerts`,
  `portfolio_snapshots`, `bars`
- Alpha-foundation additions:
  `universes`, `universe_members`, `portfolio_backtest_runs`,
  `sleeve_returns`, `ml_model_artifacts`, `pair_candidates`,
  `strategy_trials`

If counting physical partition tables separately, also mention:

- `bars_y2024`, `bars_y2025`, `bars_y2026`, `bars_default`

Important schema nuance:

- `PORTFOLIO_CATALOG` is added to `agent_runs.strategy_type`.
- Do not say `PORTFOLIO_CATALOG` is part of `strategy_configs` or
  `backtest_runs` constraints.
- Concrete catalog key lives in `agent_runs.parameters.strategy_key`.
- `agent_settings` stores only risk/settings controls, not `strategy_key` and
  not `backtest_accepted`.
- `ml_model_artifacts` stores metadata and `artifact_uri`, not model blobs.
- `estimated_annual_yield` exists in `portfolio_snapshots` but is currently
  written as `0`; do not market it as live computed yield.
- Current streaming portfolio backtests return live results; persistence into
  `portfolio_backtest_runs` is schema-ready but should not be overstated unless
  the code is wired.

## Latest TOC Update

The table of contents is a manually built 3-column table:

- `Section`
- `Subsection`
- `Page`

It is not a Word-native auto-TOC field.

For page numbers:

1. Export DOCX to PDF with Word via `docx2pdf`.
2. Use `pypdf` to extract page text.
3. Ignore TOC pages when matching headings.
4. Write literal page numbers into the existing table.
5. Re-export and verify the table values match final rendered pages.

Last verified state:

- Word-rendered PDF had 78 pages.
- TOC spans pages 2-3.
- All entries resolved with zero page mismatches.

## Final Deliverable Pattern

When done, return only the final DOCX link and a compact QA note.

Example:

`[Quant in your pocket_copy_toc_updated.docx](</Users/shizhen/Downloads/Quant in your pocket_copy_toc_updated.docx>)`

Mention:

- What was updated.
- Whether Word/PDF render was used.
- Whether ZIP integrity passed.
- Any render/tool limitation if relevant.
