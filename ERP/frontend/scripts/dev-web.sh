#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOCK_FILE="$PROJECT_DIR/.run/frontend-web.lock"

cd "$PROJECT_DIR"

acquire_single_instance_lock() {
  if ! command -v flock >/dev/null 2>&1; then
    echo "[WARN] flock is not available, skipping single-instance lock."
    return
  fi
  mkdir -p "$(dirname "$LOCK_FILE")"
  exec 9>"$LOCK_FILE"
  if ! flock -n 9; then
    echo "Vite web dev server already running; skip duplicate start."
    exit 0
  fi
}

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

VITE_HOST="127.0.0.1"
VITE_PORT="${VITE_DEV_PORT:-5173}"
EXTRA_ARGS=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --host)
      VITE_HOST="${2:-$VITE_HOST}"
      shift 2
      ;;
    --host=*)
      VITE_HOST="${1#*=}"
      shift
      ;;
    --port)
      VITE_PORT="${2:-$VITE_PORT}"
      shift 2
      ;;
    --port=*)
      VITE_PORT="${1#*=}"
      shift
      ;;
    *)
      EXTRA_ARGS+=("$1")
      shift
      ;;
  esac
done

if is_port_in_use "$VITE_PORT"; then
  echo "Vite web dev server already running on http://$VITE_HOST:$VITE_PORT"
  exit 0
fi

acquire_single_instance_lock
exec "$PROJECT_DIR/node_modules/.bin/vite" --host "$VITE_HOST" --port "$VITE_PORT" --strictPort "${EXTRA_ARGS[@]}"
