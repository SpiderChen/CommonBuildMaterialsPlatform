#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_ADDR="${CBMP_ADDR:-127.0.0.1:8088}"
BACKEND_PORT="${BACKEND_ADDR##*:}"
BACKEND_PID=""
BACKEND_WATCH_PID=""

is_wsl() {
  [[ -r /proc/version ]] && grep -qi "microsoft\\|wsl" /proc/version
}

dev_target_requires_windows_launcher() {
  local dev_target="${CBMP_DEV_TARGET:-auto}"
  if [[ "$dev_target" == "auto" ]] && is_wsl; then
    dev_target="windows-hot"
  fi
  [[ "$dev_target" == "windows-hot" || "$dev_target" == "windows-dev" || "$dev_target" == "windows" || "$dev_target" == "windows-build" ]]
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

port_is_open() {
  local port="$1"
  if command -v ss >/dev/null 2>&1; then
    local_port_is_listening "$port"
    return
  fi
  tcp_probe_port "$port"
}

wait_for_port() {
  local port="$1"
  local attempt
  for attempt in {1..120}; do
    if port_is_open "$port"; then
      return 0
    fi
    sleep 0.25
  done
  return 1
}

cleanup() {
  if [[ -n "${BACKEND_WATCH_PID:-}" ]] && kill -0 "$BACKEND_WATCH_PID" >/dev/null 2>&1; then
    kill "$BACKEND_WATCH_PID" >/dev/null 2>&1 || true
    wait "$BACKEND_WATCH_PID" >/dev/null 2>&1 || true
  fi
  if [[ -n "${BACKEND_PID:-}" ]] && kill -0 "$BACKEND_PID" >/dev/null 2>&1; then
    kill "$BACKEND_PID" >/dev/null 2>&1 || true
    wait "$BACKEND_PID" >/dev/null 2>&1 || true
  fi
}

trap cleanup EXIT INT TERM

if [[ "${CBMP_ALLOW_LOCAL_WAILS:-0}" != "1" ]] && dev_target_requires_windows_launcher; then
  "$ROOT_DIR/frontend/scripts/dev.sh" --check-windows-launcher
fi

if port_is_open "$BACKEND_PORT"; then
  echo "Using existing backend API at http://$BACKEND_ADDR"
else
  "$ROOT_DIR/backend/scripts/dev.sh" &
  BACKEND_PID=$!

  if [[ "${CBMP_WAIT_BACKEND:-0}" == "1" ]]; then
    if ! wait_for_port "$BACKEND_PORT"; then
      echo "Timed out waiting for backend API at http://$BACKEND_ADDR" >&2
      wait "$BACKEND_PID" || true
      exit 1
    fi
  else
    {
      if wait_for_port "$BACKEND_PORT"; then
        echo "Backend API is ready at http://$BACKEND_ADDR"
      else
        echo "Backend API is still starting at http://$BACKEND_ADDR" >&2
      fi
    } &
    BACKEND_WATCH_PID=$!
  fi
fi

"$ROOT_DIR/frontend/scripts/dev.sh"
