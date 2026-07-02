#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT_DIR="$(cd "$PROJECT_DIR/.." && pwd)"
GO_BIN="$("$ROOT_DIR/scripts/ensure-go.sh")"

LOCK_FILE="$PROJECT_DIR/.run/backend-dev.lock"

local_port_is_listening() {
  local port="$1"
  command -v ss >/dev/null 2>&1 || return 1
  ss -ltn 2>/dev/null | awk -v port="$port" '
    NR > 1 {
      n = split($4, parts, ":")
      if (parts[n] == port) {
        found = 1
      }
    }
    END { exit found ? 0 : 1 }
  '
}

tcp_probe_port() {
  local port="$1"
  if command -v timeout >/dev/null 2>&1; then
    timeout 1s bash -c 'echo >/dev/tcp/127.0.0.1/$1' _ "$port" >/dev/null 2>&1
    return
  fi
  return 1
}

is_port_in_use() {
  local port="$1"
  if command -v ss >/dev/null 2>&1; then
    local_port_is_listening "$port"
    return
  fi
  tcp_probe_port "$port"
}

acquire_single_instance_lock() {
  if ! command -v flock >/dev/null 2>&1; then
    echo "[WARN] flock is not available, skipping single-instance lock."
    return
  fi
  mkdir -p "$(dirname "$LOCK_FILE")"
  exec 9>"$LOCK_FILE"
  if ! flock -n 9; then
    echo "ERP backend already has an active launcher/session; skip duplicate start."
    exit 0
  fi
}

cd "$PROJECT_DIR"
export CBMP_ADDR="${CBMP_ADDR:-127.0.0.1:8088}"
export CBMP_DATA="${CBMP_DATA:-data/app.vault}"
BACKEND_PORT="${CBMP_ADDR##*:}"

if is_port_in_use "$BACKEND_PORT"; then
  echo "ERP backend already running on http://$CBMP_ADDR; skip duplicate start."
  exit 0
fi

acquire_single_instance_lock
echo "Starting backend API at http://$CBMP_ADDR"
exec "$GO_BIN" run ./cmd/appliance
