#!/bin/bash

# Load environment variables from .env (from repo root)
REPO_ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
export $(cat "$REPO_ROOT/.env" | grep -v '^#' | xargs)

# Run the Alpaca test
cd "$REPO_ROOT/backend"
go run ./cmd/alpaca-test
