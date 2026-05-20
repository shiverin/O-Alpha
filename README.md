# O(Alpha)
Quant research and paper-trading platform with a Go backend and Next.js frontend.

## Architecture
- Frontend: Next.js (React/TypeScript)
- Backend: Go (Gin) API + ingest service
- Data: PostgreSQL with TimescaleDB
- Cache/Queue: Redis
- Orchestration: Docker Compose

## Tech Stack
- Go 1.23
- Gin, zerolog, golang-migrate
- PostgreSQL + TimescaleDB
- Next.js + TypeScript

## Getting Started
### Prerequisites
- Docker and Docker Compose
- Alpaca API Keys (paper trading)

### Quick Start
1) Copy `.env.example` to `.env` and set values:
   - `ALPACA_API_KEY`
   - `ALPACA_API_SECRET`
   - `DATABASE_URL` (optional if using the compose defaults)

2) Build and run:
```
docker compose up --build
```

3) Ingest data:
```
docker compose up ingest
```

4) Verify data (example):
```
psql "postgres://oalpha:dev@localhost:5432/oalpha" -f scripts/verify_data.sql
```

## Development Notes
- Database migrations are run automatically on API and ingest startup.
- Use `INGEST_SYMBOLS`, `INGEST_INTERVAL`, and `INGEST_LOOKBACK` to control ingest.

Project maintained by Tan Jia Jun and Zhao Shi Zhen.
