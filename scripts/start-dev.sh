#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 加载 .env 文件
if [ -f "$PROJECT_ROOT/.env" ]; then
  echo "Loading .env file..."
  set -a
  source "$PROJECT_ROOT/.env"
  set +a
fi

echo "Stopping any existing Cogniforge server..."
pkill -f "server" 2>/dev/null || true
sleep 1

cd "$PROJECT_ROOT"

echo "Starting Cogniforge server..."
go run ./cmd/server/main.go
