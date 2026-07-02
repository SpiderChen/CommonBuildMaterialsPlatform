#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT_DIR="$(cd "$PROJECT_DIR/.." && pwd)"
GO_BIN="$("$ROOT_DIR/ERP/scripts/ensure-go.sh")"
LOCK_FILE="$PROJECT_DIR/.run/industrial-gateway-dev.lock"

acquire_single_instance_lock() {
  if ! command -v flock >/dev/null 2>&1; then
    echo "[WARN] flock is not available, skipping single-instance lock."
    return
  fi
  mkdir -p "$(dirname "$LOCK_FILE")"
  exec 9>"$LOCK_FILE"
  if ! flock -n 9; then
    echo "Industrial gateway already running; skip duplicate start."
    exit 0
  fi
}

acquire_single_instance_lock
cd "$PROJECT_DIR"
exec "$GO_BIN" run ./cmd/industrial-control-gateway
