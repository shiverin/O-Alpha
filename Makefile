.PHONY: help up down logs db-shell migrate test-backend run-api run-ingest

help:
	@echo "O(Alpha) dev commands:"
	@echo "  make up            - start all Docker services"
	@echo "  make down          - stop services"
	@echo "  make logs          - tail compose logs"
	@echo "  make db-shell      - psql into TimescaleDB"
	@echo "  make migrate       - run migrations locally"
	@echo "  make test-backend  - go test ./..."
	@echo "  make run-api       - run API locally"
	@echo "  make run-ingest    - run ingest locally"

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

db-shell:
	docker compose exec timescale psql -U oalpha -d oalpha

migrate:
	cd backend && MIGRATIONS_PATH=file://../migrations go run ./cmd/migrate

test-backend:
	cd backend && go test ./...

run-api:
	cd backend && MIGRATIONS_PATH=file://../migrations go run ./cmd/api

run-ingest:
	cd backend && MIGRATIONS_PATH=file://../migrations go run ./cmd/ingest
