# Database Configuration Guide

This document explains how to configure your database for different deployment scenarios.

## Quick Start

### Option 1: Local Development (VS Code) with Supabase

**Setup:**
```bash
# 1. Copy .env.local to .env
cp .env.local .env

# 2. Update DATABASE_URL in .env with your Supabase credentials
# Get this from: Supabase Dashboard → Settings → Database
DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@db.YOUR_PROJECT_REF.supabase.co:5432/postgres?sslmode=require

# 3. Run services separately
go run ./cmd/api/main.go         # In terminal 1
npm run dev                       # In terminal 2 (from frontend folder)
```

**Benefits:**
- ✅ Uses your Supabase cloud database
- ✅ Easy to debug with separate processes
- ✅ Can use VS Code debugger
- ✅ Database persists across restarts

---

### Option 2: Docker Deployment with Local Database

**Setup:**
```bash
# 1. Use .env.docker (or just use default .env)
cp .env.docker .env
# or just use the default .env

# 2. Run everything in Docker
docker-compose up

# The api container will connect to:
# - TimescaleDB: postgres://oalpha:dev@timescale:5432/oalpha
# - Redis: redis://redis:6379
```

**Benefits:**
- ✅ Everything isolated in containers
- ✅ Easy to spin up/tear down
- ✅ Matches production environment
- ✅ No local PostgreSQL needed

---

### Option 3: Docker with External Database

**Setup:**
```bash
# 1. Create .env with your Supabase/external database
cat > .env << EOF
ALPACA_API_KEY=your_key_id
ALPACA_API_SECRET=your_secret_key
DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@db.YOUR_PROJECT_REF.supabase.co:5432/postgres?sslmode=require
REDIS_URL=redis://redis:6379
MIGRATIONS_PATH=file:///migrations
HTTP_ADDR=:8080
EOF

# 2. Run docker-compose (API will use external DB, Redis runs in container)
docker-compose up
```

**Benefits:**
- ✅ Docker containers for reproducibility
- ✅ Uses cloud database for persistence
- ✅ Great for CI/CD pipelines

---

## Environment File Precedence

```
.env (takes priority)
  ↓
.env.local (local development)
.env.docker (docker deployment)
.env.example (reference)
```

**Rules:**
1. Always use `.env` for your actual setup
2. Copy from `.env.local` or `.env.docker` as a template
3. Update `DATABASE_URL` with your credentials
4. Never commit `.env` to git (already in .gitignore)

---

## Configuration Variables

### Database URL Format

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

### Other Important Variables

| Variable | Local Dev | Docker |
|----------|-----------|--------|
| `REDIS_URL` | `redis://localhost:6379` | `redis://redis:6379` |
| `MIGRATIONS_PATH` | `file://migrations` | `file:///migrations` |
| `HTTP_ADDR` | `:8080` | `:8080` |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | `http://localhost:8080` |

---

## Step-by-Step: Set Up Supabase for Local Development

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
   go run ./cmd/migrate/main.go
   ```

5. **Start Backend:**
   ```bash
   go run ./cmd/api/main.go
   ```

---

## Troubleshooting

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

## Docker Compose Environment Variables

The `docker-compose.yml` now uses variable substitution:

```yaml
# These values come from your .env file
DATABASE_URL: ${DATABASE_URL:-postgres://oalpha:dev@timescale:5432/oalpha?sslmode=disable}
REDIS_URL: ${REDIS_URL:-redis://redis:6379}
```

**Syntax:**
- `${VAR}` - Use VAR from .env
- `${VAR:-default}` - Use VAR, or fall back to "default"

This means you can override any value in docker-compose by setting it in `.env`.

---

## Summary

| Scenario | Command | Database | .env File |
|----------|---------|----------|-----------|
| Local development | `go run` + `npm run dev` | Supabase | `.env.local` |
| Docker full stack | `docker-compose up` | Local TimescaleDB | `.env.docker` |
| Docker + Supabase | `docker-compose up` | Supabase | Custom `.env` |

Choose what works for you! 🚀
