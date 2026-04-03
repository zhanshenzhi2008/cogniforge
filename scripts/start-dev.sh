#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Stopping any existing Cogniforge server..."
pkill -f "server" 2>/dev/null || true
sleep 1

cd "$PROJECT_ROOT"

echo "Starting Cogniforge server..."
PORT=8080 POSTGRES_PORT=5433 go run ./cmd/server/main.go
