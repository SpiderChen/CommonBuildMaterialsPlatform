#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT_DIR="$(cd "$PROJECT_DIR/.." && pwd)"
GO_BIN="$("$ROOT_DIR/ERP/scripts/ensure-go.sh")"

cd "$PROJECT_DIR"
exec "$GO_BIN" run ./cmd/industrial-control-gateway
