# ✅ Globally load and export the .env file if it exists in the root directory
ifneq (,$(wildcard .env))
    include .env
    export
endif

.PHONY: help up down logs db-shell migrate test-backend run-api run-ingest setup-local setup-docker

help:
	@echo "O(Alpha) dev commands:"
	@echo ""
	@echo "Setup:"
	@echo "  make setup-local   - setup for local development (Supabase)"
	@echo "  make setup-docker  - setup for Docker deployment"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make up            - start all Docker services"
	@echo "  make down          - stop services"
	@echo "  make logs          - tail compose logs"
	@echo "  make db-shell      - psql into TimescaleDB"
	@echo ""
	@echo "Local Development:"
	@echo "  make migrate       - run migrations locally"
	@echo "  make test-backend  - go test ./..."
	@echo "  make run-api       - run API locally (requires .env)"
	@echo "  make run-ingest    - run ingest locally (requires .env)"

setup-local:
	@echo "Setting up for local development with Supabase..."
	@cp .env.local .env
	@echo "✅ Copied .env.local to .env"
	@echo ""
	@echo "⚠️  Next steps:"
	@echo "1. Update DATABASE_URL in .env with your Supabase credentials"
	@echo "2. Run: make migrate"
	@echo "3. Run: make run-api"
	@echo "4. Run: npm run dev (from frontend folder)"
	@echo ""
	@echo "📖 See DATABASE_CONFIG.md for detailed instructions"

setup-docker:
	@echo "Setting up for Docker deployment..."
	@cp .env.docker .env
	@echo "✅ Copied .env.docker to .env"
	@echo ""
	@echo "⚠️  Next steps:"
	@echo "1. Run: make up"
	@echo "2. Services will start:"
	@echo "   - API: http://localhost:8080"
	@echo "   - Frontend: http://localhost:3000"
	@echo "   - Database: localhost:5432 (oalpha:dev)"
	@echo ""
	@echo "📖 See DATABASE_CONFIG.md for detailed instructions"

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

db-shell:
	docker compose exec timescale psql -U oalpha -d oalpha

migrate:
	cd backend && MIGRATIONS_PATH=file://../migrations go run ./cmd/migrate/main.go

test-backend:
	cd backend && go test ./...

run-api:
	cd backend && MIGRATIONS_PATH=file://../migrations go run ./cmd/api/main.go

run-ingest:
	cd backend && MIGRATIONS_PATH=file://../migrations go run ./cmd/ingest/main.go