# O(Alpha)

Quantitative research and paper-trading platform featuring a Go backend and a Next.js frontend.

---

# Architecture & Tech Stack

| Layer         | Technology                             |
| ------------- | -------------------------------------- |
| Frontend      | Next.js (React / TypeScript)           |
| Backend       | Go 1.23 (Gin, zerolog, golang-migrate) |
| Database      | PostgreSQL + TimescaleDB               |
| Cache / Queue | Redis                                  |
| Orchestration | Docker Compose                         |

---

# Database & Deployment Scenarios

Choose the setup scenario that matches your development workflow.

| Scenario             | Database                    | Configuration Template | Run Command                    |
| -------------------- | --------------------------- | ---------------------- | ------------------------------ |
| 1. Local Development | Cloud (Supabase)            | `.env.local`           | `make run-api` + `npm run dev` |
| 2. Docker Full Stack | Local Container (Timescale) | `.env.docker`          | `make up`                      |
| 3. Docker + Cloud DB | Cloud (Supabase)            | Custom `.env`          | `docker compose up -d --build` |

---

# Getting Started: Step-by-Step

## Option 1: Local Development (Go + Next.js with Supabase)

Best for rapid code changes, debugging, and utilizing the VS Code debugger.

### 1. Initialize the configuration file

```bash
make setup-local
```

This copies `.env.local` to `.env`.

### 2. Configure your Supabase URL

Open the newly generated `.env` file and update the `DATABASE_URL` with your Supabase connection string:

```env
DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@db.YOUR_PROJECT_REF.supabase.co:5432/postgres?sslmode=require
```

### 3. Run database migrations

```bash
make migrate
```

Executes migrations against your cloud database.

### 4. Start the services

#### Terminal 1 — Backend API

```bash
make run-api
```

#### Terminal 2 — Frontend UI

```bash
cd frontend && npm run dev
```

---

## Option 2: Docker Deployment (Full Local Stack)

Best for evaluating the environment cleanly without needing local database installations.

### 1. Initialize the configuration file

```bash
make setup-docker
```

This copies `.env.docker` to `.env`.

### 2. Add your Alpaca Paper Trading Credentials

Open `.env` and fill out:

```env
ALPACA_API_KEY=YOUR_KEY
ALPACA_API_SECRET=YOUR_SECRET
```

### 3. Spin up the entire stack

```bash
make up
```

Launches isolated containers for:

- TimescaleDB
- Redis
- API
- Ingest
- Frontend

---

## Option 3: Docker Orchestration with Supabase

Best for reproducing staging environments locally while pointing to persistent cloud storage.

### 1. Initialize the template

```bash
make setup-local
```

### 2. Modify your `.env` variables for container cross-communication

Update `DATABASE_URL` to point to your Supabase connection string.

### 3. Critical Redis configuration change

⚠️ **IMPORTANT:** Change `REDIS_URL` from `localhost` to the Docker network service name:

```env
REDIS_URL=redis://redis:6379
```

If left as `localhost`, the containerized API will search for Redis internally within its own loopback interface instead of routing to the Redis container.

### 4. Spin up the containers

```bash
docker compose up -d --build
```

---

# Environment File & Variables Guide

## File Precedence Rules

```text
.env
  ↓
.env.local
.env.docker
```

### Explanation

| File          | Purpose                                             |
| ------------- | --------------------------------------------------- |
| `.env`        | Active runtime file — never commit to Git           |
| `.env.local`  | Template for Local Development / Supabase           |
| `.env.docker` | Template for Local Docker Containerized Development |

---

# Connection Reference Table

| Variable              | Local Dev Setup (Option 1)                   | Docker Container Setup (Option 2 & 3)                         |
| --------------------- | -------------------------------------------- | ------------------------------------------------------------- |
| `DATABASE_URL`        | Cloud URL or `postgres://localhost:5432/...` | `postgres://oalpha:dev@timescale:5432/oalpha?sslmode=disable` |
| `REDIS_URL`           | `redis://localhost:6379`                     | `redis://redis:6379`                                          |
| `MIGRATIONS_PATH`     | `file://migrations`                          | `file:///migrations`                                          |
| `HTTP_ADDR`           | `:8080`                                      | `:8080`                                                       |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080`                      | `http://localhost:8080`                                       |

---

# Smart Variable Substitutions

Your `docker-compose.yml` includes flexible variable defaults.

If `DATABASE_URL` or `REDIS_URL` are not explicitly defined inside the active `.env` file, Docker automatically falls back to the internal local service architecture.

---

# Data Ingestion & Verification

Database migrations are designed to run safely during container or service initialization.

## Triggering the Data Ingestion Routine

Financial market ingestion behavior is controlled through:

- `INGEST_SYMBOLS`
- `INGEST_INTERVAL`
- `INGEST_LOOKBACK`

inside the `.env` file.

### Local Execution

```bash
make run-ingest
```

### Docker Execution

```bash
docker compose run ingest
```

---

# Verifying Active Table Content

To inspect your local Timescale database container directly:

```bash
make db-shell
```

---

# Troubleshooting Checklist

## "Can't connect to database / Connection Refused"

### Supabase

Ensure `sslmode=require` is appended to the end of your database connection string.

### Local Engine

Verify that:

- Your local Redis daemon is running:

```bash
redis-server
```

- Or confirm containers are healthy:

```bash
make logs
```

---

## "Wrong Host / Port Configurations"

### VS Code (Local Engine)

Use explicit `localhost` endpoints instead of strict loopback aliases like `127.0.0.1`.

### Docker

Ensure services communicate using internal Docker service names:

- `timescale`
- `redis`

instead of `localhost`.

---

# Maintainers

Project maintained by:

- Tan Jia Jun
- Zhao Shi Zhen
