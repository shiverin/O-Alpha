# O(Alpha)

O(Alpha) is a paper-trading platform for running approved portfolio catalog strategies before they ever touch real execution.

It combines a Go trading backend, a Next.js dashboard, PostgreSQL/TimescaleDB market-data storage, and a catalog-driven paper workflow. The project is intentionally paper-only: approved catalog strategies run through persisted paper fills, positions, snapshots, and alerts.

## Highlights

- Portfolio-agent paper trading: one active agent per user, one chosen catalog strategy, daily-bar evaluation, deterministic fill idempotency, and DB-backed account state.
- Real dashboard state: portfolio summary, positions, allocation, execution log, alerts, and regime labels are read from backend state.
- Curated default universe: Yahoo100-style equity/ETF universe with `VOO` included as the portfolio benchmark anchor.
- Artifact-aware ranker strategies: LGBM rankers read private model artifacts and fail closed when required artifacts are missing.
- Local or containerized development: run against Supabase locally, or bring up TimescaleDB, Redis, API, ingest, and frontend with Docker Compose.

## What It Does

O(Alpha) has one production-facing loop:

1. Paper-trading loop: start a catalog strategy from the dashboard/API, evaluate on daily bars, reconcile target weights into paper fills, and persist the resulting state for the dashboard.

Strategy research, promotion evidence, and raw model-training artifacts are maintained outside this public repository. The public app consumes only approved catalog strategy definitions and private runtime artifacts.

The current production-facing paper flow is the portfolio catalog path:

- `GET /api/v1/strategies/catalog`
- `POST /api/v1/agent/portfolio/start`
- `POST /api/v1/agent/portfolio/stop`
- `GET /api/v1/agent/list`
- `GET /api/v1/user/portfolio/summary`
- `GET /api/v1/user/portfolio/positions`
- `GET /api/v1/user/portfolio/trades`
- `GET /api/v1/user/portfolio/alerts`

## Tech Stack

| Layer | Technology |
| --- | --- |
| Frontend | Next.js, React, TypeScript, Tailwind |
| Backend | Go, Gin, zerolog, golang-migrate |
| Database | PostgreSQL with TimescaleDB-compatible schema |
| Cache / queue | Redis |
| Market data | Alpaca and Yahoo daily ingestion paths |
| Orchestration | Docker Compose |

## Repository Map

```text
backend/                 Go API, portfolio agent, DB repositories
frontend/                Next.js dashboard
migrations/              SQL migrations
scripts/                 Utility scripts
docker-compose.yml       Local full-stack orchestration
```

## Quick Start

### Option 1: Local API + Frontend With Supabase

Use this when you want fast backend/frontend iteration while storing data in Supabase.

```bash
make setup-local
```

Edit `.env` and set at least:

```env
DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@db.YOUR_PROJECT_REF.supabase.co:5432/postgres?sslmode=require
REDIS_URL=redis://localhost:6379
ALPACA_API_KEY=YOUR_ALPACA_KEY
ALPACA_API_SECRET=YOUR_ALPACA_SECRET
```

Run migrations and start the services:

```bash
make migrate
make run-api
```

In another terminal:

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:3000`.

### Option 2: Full Docker Stack

Use this when you want a clean local stack with containerized TimescaleDB, Redis, API, ingest, and frontend.

```bash
make setup-docker
```

Edit `.env` and add Alpaca credentials if you want live data ingestion:

```env
ALPACA_API_KEY=YOUR_ALPACA_KEY
ALPACA_API_SECRET=YOUR_ALPACA_SECRET
```

Start everything:

```bash
make up
```

Useful follow-ups:

```bash
make logs
make db-shell
make down
```

## Environment

`.env` is the active runtime file and should not be committed.

| File | Purpose |
| --- | --- |
| `.env.example` | Safe reference template |
| `.env.local` | Local development template, typically Supabase |
| `.env.docker` | Docker Compose template |
| `.env` | Active runtime config |

Important variables:

| Variable | Purpose |
| --- | --- |
| `DATABASE_URL` | PostgreSQL/Supabase connection string |
| `REDIS_URL` | Redis connection string |
| `MIGRATIONS_PATH` | Migration path, usually `file://../migrations` locally |
| `HTTP_ADDR` | API bind address, usually `:8080` |
| `CORS_ALLOWED_ORIGINS` | Comma-separated frontend origins; defaults include `localhost`, `127.0.0.1`, `[::1]`, and the Vercel app |
| `NEXT_PUBLIC_API_URL` | Frontend API base URL |
| `INGEST_SYMBOLS` | Comma-separated market-data universe |
| `INGEST_INTERVAL` | Ingest interval, usually `1Day` for portfolio strategies |
| `INGEST_LOOKBACK` | Backfill lookback duration |
| `INGEST_RUN_ONCE` | Whether ingest exits after one pass |
| `OALPHA_DAILY_RANKER_ARTIFACT_ROOT` | Private root for LGBM ranker model artifacts |
| `OALPHA_DAILY_RANKER_PIT_UNIVERSE` | Optional point-in-time universe file |

For local ranker artifacts, mount or point to a private, ignored directory:

```env
OALPHA_DAILY_RANKER_ARTIFACT_ROOT=/var/lib/oalpha/models/fold_artifacts
```

Do not store large model blobs in Postgres. Use local mounts for development and object storage plus `ml_model_artifacts.artifact_uri` as the deployment registry path.

## Private Artifacts

Research reports, validation runs, model files, feature fixtures, and promotion evidence are intentionally excluded from this public repository. Keep them in a private artifact store, encrypted bucket, private model registry, or ignored local mount.

Runtime code expects approved LGBM ranker artifacts at `OALPHA_DAILY_RANKER_ARTIFACT_ROOT`. The default container path is `/opt/oalpha/ranker-fold-artifacts`, but deployment must provide that directory privately. The Docker image does not copy artifacts from the source tree.

## Market Data

The portfolio agent expects daily bars for the curated universe. Ingest is controlled by `INGEST_SYMBOLS`, `INGEST_INTERVAL`, and `INGEST_LOOKBACK`.

Run one local ingest worker:

```bash
make run-ingest
```

Run the Docker ingest service:

```bash
docker compose run ingest
```

Verify local container data:

```bash
make db-shell
```

For Supabase or another remote database, connect with your normal SQL client and inspect the bars, positions, fills, snapshots, and alerts tables.

## Paper Trading Flow

1. Ingest or sync daily bars for the curated universe.
2. Start the API and frontend.
3. Open the dashboard and complete onboarding by selecting a risk profile and approved catalog strategy.
4. Start the portfolio agent.
5. The agent warms up on daily bars, evaluates immediately, writes target-weight paper fills, updates positions/snapshots, and records alerts.
6. The dashboard reads state from the database-backed API endpoints.

The v1 execution router is long-only. It sells reductions before buys, uses deterministic `client_order_id` keys for idempotency, and writes a portfolio snapshot on every evaluation.

## Testing

The canonical test policy, tiers, and safety invariants live in `docs/TEST_STRATEGY.md`.

Backend:

```bash
make test-backend
```

Frontend:

```bash
cd frontend
npm run lint
npm run typecheck
npm test
npm run test:e2e
```

All fast unit suites:

```bash
make test-all
```

Opt-in database integration tests:

```bash
OALPHA_TEST_DATABASE_URL=postgres://... make test-integration
```

Full-stack smoke checklist:

```text
1. Run migrations.
2. Ingest daily bars for the configured universe.
3. Start API and frontend.
4. Confirm /api/v1/strategies/catalog returns the expected catalog and universe.
5. Start a low-risk catalog strategy.
6. Confirm /api/v1/agent/list shows an active run.
7. Confirm positions, fills, portfolio summary, allocation, and alerts update.
8. Restart the API and confirm stale active runs are reconciled.
```

## Troubleshooting

### Database connection refused

- Supabase URLs usually need `sslmode=require`.
- Local Docker services should use internal hostnames such as `timescale` and `redis`.
- Local non-Docker services usually use `localhost`.

### Redis connection fails in Docker

Use:

```env
REDIS_URL=redis://redis:6379
```

not `localhost`, because `localhost` inside a container points at that container.

### LGBM ranker will not start

Check that `OALPHA_DAILY_RANKER_ARTIFACT_ROOT` points to the mounted artifact directory and that the required model files exist. The API intentionally fails closed rather than silently falling back to the proxy strategy.

### Dashboard stays idle

Check:

- the user is authenticated,
- onboarding is complete,
- a strategy was accepted after backtest,
- `/api/v1/agent/list` returns an active run,
- daily bars exist for the configured universe.

## Maintainers

- Tan Jia Jun
- Zhao Shi Zhen
