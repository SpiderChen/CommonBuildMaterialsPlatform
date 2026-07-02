#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOCK_FILE="$ROOT/.run/operations-dev.lock"

is_port_in_use() {
  local port="$1"

  if ss -ltn 2>/dev/null | awk 'NR > 1 {print $4}' | awk -F: '{print $NF}' | grep -qx "$port"; then
    return 0
  fi

  if timeout 1 bash -c 'echo >/dev/tcp/127.0.0.1/"$1"' _ "$port" >/dev/null 2>&1; then
    return 0
  fi

  return 1
}

acquire_single_instance_lock() {
  if ! command -v flock >/dev/null 2>&1; then
    echo "[WARN] flock is not available, skipping single-instance lock."
    return
  fi
  mkdir -p "$(dirname "$LOCK_FILE")"
  exec 9>"$LOCK_FILE"
  if ! flock -n 9; then
    echo "Operations platform backend already running; skip duplicate start."
    exit 0
  fi
}

export CBM_OPS_ADDR="${CBM_OPS_ADDR:-:8095}"
export CBM_OPS_DATA="${CBM_OPS_DATA:-$ROOT/backend/data/ops.json}"
export CBM_OPS_FRONTEND_DIR="${CBM_OPS_FRONTEND_DIR:-$ROOT/frontend}"

OPS_PORT="${CBM_OPS_ADDR##*:}"
if is_port_in_use "$OPS_PORT"; then
  echo "Operations platform already running on $CBM_OPS_ADDR; skip duplicate start."
  exit 0
fi

acquire_single_instance_lock
cd "$ROOT/backend"
go run ./cmd/server
