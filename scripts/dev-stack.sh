#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENGINE_PORT=${ENGINE_PORT:-50051}
DATA_DIR=${CARTOMIX_DATA_DIR:-"$ROOT/.data"}

mkdir -p "$DATA_DIR"

echo "[algiers] starting engine on :$ENGINE_PORT (data: $DATA_DIR)"
CARTOMIX_DATA_DIR="$DATA_DIR" PORT="$ENGINE_PORT" go run ./cmd/engine &
ENGINE_PID=$!

echo "[algiers] installing web deps (once)"
npm install --prefix "$ROOT/web" >/dev/null 2>&1 || true

echo "[algiers] starting web dev server (Vite)"
npm run dev --prefix "$ROOT/web" -- --host &
WEB_PID=$!

echo "✔ stack running — press Ctrl+C to stop"
trap 'kill $ENGINE_PID $WEB_PID 2>/dev/null || true' EXIT
wait
