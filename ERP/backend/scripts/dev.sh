#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT_DIR="$(cd "$PROJECT_DIR/.." && pwd)"
GO_BIN="$("$ROOT_DIR/scripts/ensure-go.sh")"

cd "$PROJECT_DIR"
export CBMP_ADDR="${CBMP_ADDR:-127.0.0.1:8088}"
export CBMP_DATA="${CBMP_DATA:-data/app.vault}"

echo "Starting backend API at http://$CBMP_ADDR"
exec "$GO_BIN" run ./cmd/appliance
