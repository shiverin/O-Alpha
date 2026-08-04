# O(Alpha)

O(Alpha) is a full-stack paper-trading platform for evaluating approved portfolio strategies in a controlled environment. It combines catalog-driven strategy selection, historical backtesting, simulated execution, portfolio monitoring, and safety controls in one application.

## Features

- Guided onboarding with risk-profile selection and strategy backtesting
- Catalog-based portfolio agents with persisted run state
- Simulated fills, positions, portfolio snapshots, alerts, and activity history
- Live dashboard views for allocation, performance, execution, and market regime
- Server-enforced settings lock while an agent is starting or running
- Deterministic execution identifiers for retry-safe paper fills
- PostgreSQL-backed state with optional TimescaleDB and Redis services
- Backend, frontend, integration, and browser test suites

## Architecture

```text
Next.js dashboard
       |
       v
Go REST API ---- Redis
       |
       v
PostgreSQL / TimescaleDB
       ^
       |
Market-data ingest worker ---- Alpaca / Yahoo data
```

| Layer               | Technology                                         |
| ------------------- | -------------------------------------------------- |
| Frontend            | Next.js 14, React, TypeScript, Tailwind CSS        |
| Backend             | Go 1.23, Gin, pgx, zerolog                         |
| Data                | PostgreSQL, TimescaleDB-compatible schema, Redis   |
| Testing             | Go test, Vitest, React Testing Library, Playwright |
| Local orchestration | Docker Compose, Make                               |

## Prerequisites

Choose either the containerized setup or the local-development setup.

- Docker Desktop with Docker Compose, or
- Go 1.23+, Node.js 20+, npm, PostgreSQL, and Redis
- Optional Alpaca paper credentials for external market-data ingestion

## Quick Start

### Docker Compose

This is the simplest way to run the complete stack locally.

```bash
git clone https://github.com/shiverin/O-Alpha.git
cd O-Alpha
make setup-docker
make up
```

Open:

- Frontend: `http://localhost:3000`
- API health check: `http://localhost:8080/health`

View logs or stop the stack with:

```bash
make logs
make down
```

### Local Development

Use this workflow when running the API and frontend directly on your machine. PostgreSQL and Redis must already be reachable.

```bash
git clone https://github.com/shiverin/O-Alpha.git
cd O-Alpha
cp .env.example .env
```

Update `.env` with your local database, Redis, JWT, and optional Alpaca configuration, then run:

```bash
make migrate
make run-api
```

In a second terminal:

```bash
cd frontend
npm ci
npm run dev
```

The frontend runs at `http://localhost:3000` and uses `NEXT_PUBLIC_API_URL` to reach the API.

## Configuration

Never commit `.env` files or credentials. Start from `.env.example` and configure only the services you use.

| Variable                            | Description                               | Typical local value                                           |
| ----------------------------------- | ----------------------------------------- | ------------------------------------------------------------- |
| `DATABASE_URL`                      | PostgreSQL connection string              | `postgres://oalpha:dev@localhost:5432/oalpha?sslmode=disable` |
| `REDIS_URL`                         | Redis connection string                   | `redis://localhost:6379`                                      |
| `JWT_SECRET`                        | Secret used to sign authentication tokens | Set a private random value                                    |
| `HTTP_ADDR`                         | API bind address                          | `:8080`                                                       |
| `MIGRATIONS_PATH`                   | SQL migration location                    | `file://../migrations`                                        |
| `NEXT_PUBLIC_API_URL`               | API URL used by the frontend              | `http://localhost:8080`                                       |
| `PORTFOLIO_EXECUTION_MODE`          | Paper execution adapter                   | `internal`                                                    |
| `ALPACA_API_KEY`                    | Optional Alpaca paper/data key            | Your paper-account key                                        |
| `ALPACA_API_SECRET`                 | Optional Alpaca paper/data secret         | Your paper-account secret                                     |
| `INGEST_SYMBOLS`                    | Comma-separated ingest universe           | See `.env.example`                                            |
| `OALPHA_DAILY_RANKER_ARTIFACT_ROOT` | Private ranker artifact directory         | Local or mounted private path                                 |

The API defaults to the internal multi-user paper ledger. Private ranker artifacts must be supplied at runtime and are intentionally excluded from the repository.

## Common Commands

```bash
make help              # List development commands
make up                # Build and start the Docker stack
make down              # Stop the Docker stack
make logs              # Follow container logs
make db-shell          # Open psql in the database container
make migrate           # Apply database migrations
make run-api           # Run the API locally
make run-ingest        # Run the market-data ingest worker locally
```

## Testing

Run the fast backend and frontend suites:

```bash
make test-all
```

Run static frontend checks and browser tests:

```bash
cd frontend
npm run lint
npm run typecheck
npm run test:e2e
```

Database integration tests require a disposable PostgreSQL database:

```bash
OALPHA_TEST_DATABASE_URL='postgres://...' make test-integration
```

See [docs/TEST_STRATEGY.md](docs/TEST_STRATEGY.md) for test tiers, fixtures, CI gates, and safety invariants.

## Project Structure

```text
.
|-- backend/           # Go API, agent runtime, repositories, and workers
|-- frontend/          # Next.js dashboard and browser tests
|-- migrations/        # Versioned PostgreSQL migrations
|-- docs/              # Engineering and testing documentation
|-- scripts/           # Development and operational utilities
|-- docker-compose.yml # Local full-stack services
`-- Makefile           # Common setup, run, and test commands
```

## Safety Model

- Only approved catalog strategies can start portfolio agents.
- Each user can have at most one active portfolio agent.
- Saved settings cannot change while an agent is `starting` or `running`.
- Backend guards are authoritative; disabled frontend controls are an additional UX safeguard.
- Missing required strategy artifacts cause execution to fail closed.
- Paper fills use deterministic identifiers to prevent duplicate execution during retries.

## Troubleshooting

**The API cannot connect to PostgreSQL**

Check `DATABASE_URL`, confirm the database is running, and use `sslmode=require` for providers that require TLS.

**Redis fails inside Docker**

Use `redis://redis:6379`. Inside a container, `localhost` refers to that container.

**The dashboard has no portfolio activity**

Confirm onboarding is complete, an approved strategy was accepted, daily bars exist, and `/api/v1/agent/list` reports an active run.

**A ranker strategy will not start**

Confirm `OALPHA_DAILY_RANKER_ARTIFACT_ROOT` points to a mounted private artifact directory containing the required files.

## Maintainers

- Tan Jia Jun
- Zhao Shi Zhen
