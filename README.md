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

### Option 1: Local Development (VS Code + Supabase)
```bash
make setup-local
# Update .env with your Supabase credentials
make run-api                   # Terminal 1
npm run dev                    # Terminal 2
```

### Option 2: Docker Deployment (Local Database)
```bash
make setup-docker
make up
```


### Option 3: Docker + Supabase
```bash
cp .env.local .env
# Update DATABASE_URL in .env with Supabase credentials
docker compose up
```


---

## Database Configuration

### Quick Reference

| Scenario | Command | Database | Config File |
|----------|---------|----------|-----------|
| Local development | `go run` + `npm run dev` | Supabase | `.env.local` |
| Docker full stack | `docker-compose up` | Local TimescaleDB | `.env.docker` |
| Docker + Supabase | `docker-compose up` | Supabase | Custom `.env` |

### Environment Variables

**Copy the template that matches your setup:**

```bash
make setup-local   # For local development with Supabase
# or
make setup-docker  # For Docker with local database
```

### Database URL Formats

**PostgreSQL (Local):**
```
postgres://user:password@localhost:5432/dbname?sslmode=disable
```

**Supabase (Cloud):**
```
postgresql://postgres:PASSWORD@db.PROJECT.supabase.co:5432/postgres?sslmode=require
```

**Docker Internal:**
```
postgres://user:password@timescale:5432/dbname?sslmode=disable
```

### Important Variables

| Variable | Local Dev | Docker |
|----------|-----------|--------|
| `DATABASE_URL` | Supabase URL or local | `postgres://oalpha:dev@timescale:5432/oalpha?sslmode=disable` |
| `REDIS_URL` | `redis://localhost:6379` | `redis://redis:6379` |
| `MIGRATIONS_PATH` | `file://migrations` | `file:///migrations` |
| `HTTP_ADDR` | `:8080` | `:8080` |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | `http://localhost:8080` |

### Setup Supabase for Local Development

1. **Create Supabase Project:**
   - Go to https://supabase.com
   - Click "New Project"
   - Save your password (you'll need it)

2. **Get Connection String:**
   - Dashboard → Settings → Database
   - Under "Connection string", select "PostgreSQL"
   - Copy the connection string

3. **Update .env:**
   ```bash
   # Paste the Supabase URL
   DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@db.YOUR_PROJECT.supabase.co:5432/postgres?sslmode=require
   ```

4. **Run Migrations:**
   ```bash
   cd backend
   MIGRATIONS_PATH=file://../migrations go run ./cmd/migrate
   ```

5. **Start Backend:**
   ```bash
   cd backend
   MIGRATIONS_PATH=file://../migrations go run ./cmd/api
   ```

6. **Start Frontend** (in separate terminal):
   ```bash
   cd frontend
   npm run dev
   ```

### Docker with Environment Variables

The `docker-compose.yml` uses smart variable substitution:

```yaml
DATABASE_URL: ${DATABASE_URL:-postgres://oalpha:dev@timescale:5432/oalpha?sslmode=disable}
REDIS_URL: ${REDIS_URL:-redis://redis:6379}
```

**This means:**
- If you set `DATABASE_URL` in `.env` → Docker uses it (e.g., Supabase)
- If not set → Falls back to local TimescaleDB container
- Same pattern for Redis and other services

### Troubleshooting

**"Can't connect to database"**
- Check DATABASE_URL is correct
- For Supabase: verify `sslmode=require` is set
- For local: ensure PostgreSQL/Redis is running

**"Connection refused"**
- Local dev: Start Redis: `redis-server`
- Docker: Run `docker-compose down && docker-compose up`

**"SSL certificate error"**
- Supabase requires SSL. Use `sslmode=require`
- Local dev doesn't need SSL: use `sslmode=disable`

**"Wrong host/port"**
- VS Code: Use `localhost` (not `127.0.0.1`)
- Docker: Use service name (`timescale`, `redis`)

---

## Quick Start (Docker Only)

### Prerequisites
- Docker and Docker Compose
- Alpaca API Keys (paper trading)

### Steps
1) Copy `.env.example` to `.env` and set values:
   - `ALPACA_API_KEY`
   - `ALPACA_API_SECRET`

2) Build and run:
```bash
docker compose up --build
```

3) Ingest data:
```bash
docker compose run ingest
```

4) Verify data (example):
```bash
psql "postgres://oalpha:dev@localhost:5432/oalpha" -f scripts/verify_data.sql
```

## Development Notes
- Database migrations are run automatically on API and ingest startup.
- Use `INGEST_SYMBOLS`, `INGEST_INTERVAL`, and `INGEST_LOOKBACK` to control ingest.

Project maintained by Tan Jia Jun and Zhao Shi Zhen.
