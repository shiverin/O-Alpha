# Git Commit Handoff

The working tree is ready for a checkpoint commit, but this Codex sandbox has
read-only access to `.git`, so `git add` cannot create `.git/index.lock`.

Run these commands from the repo root in a normal terminal:

```bash
git add -A
find reports -type f \( -name '*.md' -o -name '*.json' \) -print0 | xargs -0 git add -f --
cat > /tmp/oalpha_commit_message.txt <<'EOF'
research: Add alpha research handoff

The alpha search has a benchmark-funded ranker checkpoint that promotes on
the primary Yahoo100 window, while shifted validation still fails the PBO
gate.

Parallel research needs a reproducible map of the harness, scripts, report
artifacts, and current blockers so another agent can continue without
reconstructing the conversation.

Let's add the official CSV-backed alpha-research path, ranker proxy variants,
Yahoo/feature-parity tooling, report-backed documentation, and a parallel
agent plan.

This keeps promotion evidence in the existing Go validation harness instead
of a parallel backtester, while using Python only for prescreening and
hypothesis generation.

The h63 checkpoint is a serious research result, not deployment-grade alpha:
the primary report promotes with PBO 0.200, but the shifted 2016 report fails
with PBO 0.231.
EOF
git commit -F /tmp/oalpha_commit_message.txt
git push
```

Suggested commit message:

```text
research: Add alpha research handoff

The alpha search has a benchmark-funded ranker checkpoint that promotes on
the primary Yahoo100 window, while shifted validation still fails the PBO
gate.

Parallel research needs a reproducible map of the harness, scripts, report
artifacts, and current blockers so another agent can continue without
reconstructing the conversation.

Let's add the official CSV-backed alpha-research path, ranker proxy variants,
Yahoo/feature-parity tooling, report-backed documentation, and a parallel
agent plan.

This keeps promotion evidence in the existing Go validation harness instead
of a parallel backtester, while using Python only for prescreening and
hypothesis generation.

The h63 checkpoint is a serious research result, not deployment-grade alpha:
the primary report promotes with PBO 0.200, but the shifted 2016 report fails
with PBO 0.231.
```
