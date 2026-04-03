#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=========================================="
echo "  Cogniforge Microservices"
echo "=========================================="
echo ""

echo "[1/1] Building server..."
cd "$PROJECT_ROOT"
go build -o server ./cmd/server/main.go

echo ""
echo "=========================================="
echo "  Build complete: ./server"
echo ""
echo "  Current mode: MONOLITH (all in one)"
echo "  Port: 8080"
echo ""
echo "  To run:"
echo "    POSTGRES_PORT=5433 ./server"
echo "=========================================="
