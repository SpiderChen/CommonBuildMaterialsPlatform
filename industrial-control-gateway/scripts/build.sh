#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT_DIR="$(cd "$PROJECT_DIR/.." && pwd)"
GO_BIN="$("$ROOT_DIR/ERP/scripts/ensure-go.sh")"

cd "$PROJECT_DIR"
mkdir -p dist
"$GO_BIN" build -trimpath -ldflags="-s -w" -o dist/industrial-control-gateway ./cmd/industrial-control-gateway
echo "Built industrial-control-gateway/dist/industrial-control-gateway"
