#!/bin/bash
set -e

echo "Stopping Cogniforge services..."

pkill -f "server" 2>/dev/null || true

echo "All services stopped."
