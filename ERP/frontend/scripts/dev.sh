#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT_DIR="$(cd "$PROJECT_DIR/.." && pwd)"
LOCK_FILE="$PROJECT_DIR/.run/frontend-desktop.lock"
WINDOWS_DEV_WATCH_PID=""

cd "$PROJECT_DIR"

timestamp_seconds() {
  date +%s
}

elapsed_seconds() {
  local started_at="$1"
  local finished_at
  finished_at="$(timestamp_seconds)"
  printf '%ss' "$((finished_at - started_at))"
}

dev_status() {
  echo "[dev] $*"
}

acquire_single_instance_lock() {
  if ! command -v flock >/dev/null 2>&1; then
    echo "[WARN] flock is not available, skipping single-instance lock."
    return
  fi
  mkdir -p "$(dirname "$LOCK_FILE")"
  exec 9>"$LOCK_FILE"
  if ! flock -n 9; then
    echo "ERP frontend launcher already running; skip duplicate start."
    exit 0
  fi
}

wait_for_pids_to_exit() {
  local pid
  local attempt
  for attempt in {1..40}; do
    local alive=0
    for pid in "$@"; do
      if kill -0 "$pid" >/dev/null 2>&1; then
        alive=1
        break
      fi
    done
    if [[ "$alive" == "0" ]]; then
      return 0
    fi
    sleep 0.25
  done

  for pid in "$@"; do
    if kill -0 "$pid" >/dev/null 2>&1; then
      kill -KILL "$pid" >/dev/null 2>&1 || true
    fi
  done
}

stop_existing_frontend_launcher() {
  if [[ "${CBMP_RESTART_FRONTEND:-1}" != "1" ]]; then
    return 0
  fi
  if [[ ! -f "$LOCK_FILE" ]] || ! command -v fuser >/dev/null 2>&1; then
    return 0
  fi

  local pids=()
  local pid
  mapfile -t pids < <(fuser "$LOCK_FILE" 2>/dev/null | tr ' ' '\n' | grep -E '^[0-9]+$' || true)
  if [[ "${#pids[@]}" == "0" ]]; then
    return 0
  fi

  echo "Stopping existing ERP frontend launcher before restart."
  for pid in "${pids[@]}"; do
    if [[ "$pid" == "$$" || "$pid" == "${BASHPID:-}" ]]; then
      continue
    fi
    kill "$pid" >/dev/null 2>&1 || true
  done
  wait_for_pids_to_exit "${pids[@]}"
}

wails_tag_args=()
if [[ "$(uname -s)" == "Linux" ]]; then
  if ! pkg-config --exists webkit2gtk-4.0 2>/dev/null && pkg-config --exists webkit2gtk-4.1 2>/dev/null; then
    wails_tag_args=(-tags webkit2_41)
  fi
fi
if [[ -n "${WAILS_TAGS:-}" ]]; then
  wails_tag_args=(-tags "$WAILS_TAGS")
fi

is_wsl() {
  [[ -r /proc/version ]] && grep -qi "microsoft\\|wsl" /proc/version
}

windows_path_literal() {
  local value="$1"
  value="${value//\'/\'\'}"
  printf "'%s'" "$value"
}

find_windows_exe() {
  local exe_name="$1"

  case "$exe_name" in
    powershell.exe)
      local candidate
      for candidate in \
        /mnt/c/WINDOWS/System32/WindowsPowerShell/v1.0/powershell.exe \
        /mnt/c/Windows/System32/WindowsPowerShell/v1.0/powershell.exe; do
        if [[ -x "$candidate" ]]; then
          printf '%s\n' "$candidate"
          return 0
        fi
      done
      if is_wsl; then
        return 1
      fi
      ;;
    cmd.exe)
      local candidate
      for candidate in \
        /mnt/c/WINDOWS/System32/cmd.exe \
        /mnt/c/Windows/System32/cmd.exe; do
        if [[ -x "$candidate" ]]; then
          printf '%s\n' "$candidate"
          return 0
        fi
      done
      if is_wsl; then
        return 1
      fi
      ;;
  esac

  local found
  found="$(command -v "$exe_name" 2>/dev/null || true)"
  if [[ -n "$found" ]]; then
    printf '%s\n' "$found"
    return 0
  fi

  return 1
}

resolve_windows_launcher_tools() {
  WINDOWS_POWERSHELL_BIN="${CBMP_POWERSHELL_EXE:-$(find_windows_exe powershell.exe || true)}"
  WINDOWS_CMD_BIN="${CBMP_CMD_EXE:-$(find_windows_exe cmd.exe || true)}"
  WINDOWS_WSLPATH_BIN="${CBMP_WSLPATH_EXE:-$(command -v wslpath 2>/dev/null || true)}"
  [[ -n "$WINDOWS_POWERSHELL_BIN" && -n "$WINDOWS_CMD_BIN" && -n "$WINDOWS_WSLPATH_BIN" ]]
}

windows_launcher_tools_available() {
  resolve_windows_launcher_tools
}

run_windows_powershell() {
  "$WINDOWS_POWERSHELL_BIN" -NoProfile -ExecutionPolicy Bypass "$@"
}

windows_path_from_wsl() {
  "$WINDOWS_WSLPATH_BIN" -w "$1"
}

windows_start_exe() {
  local exe_path="$1"
  local dev_url="${2:-}"
  local ps_exe
  local ps_dev_url
  ps_exe="$(windows_path_literal "$exe_path")"
  ps_dev_url="$(windows_path_literal "$dev_url")"
  run_windows_powershell -Command "\$ErrorActionPreference='Stop'; \$exe=$ps_exe; \$devUrl=$ps_dev_url; \$env:CBMP_FRONTEND_DEV_URL=\$devUrl; \$work=Split-Path -Parent \$exe; Set-Location \$env:SystemRoot; \$cmd='start \"\" /D \"' + \$work + '\" \"' + \$exe + '\"'; & cmd.exe /d /s /c \$cmd"
}

windows_dev_health_url() {
  if [[ -n "${CBMP_WAILS_HEALTH_URL:-}" ]]; then
    printf '%s\n' "$CBMP_WAILS_HEALTH_URL"
  fi
}

report_windows_launcher_unavailable() {
  cat >&2 <<'EOF'
Windows desktop app launch is not available from this WSL session.
Required WSL tools: powershell.exe, cmd.exe, and wslpath.

In this session the Windows drive mounts appear unhealthy; /mnt/c returning
Input/output error prevents WSL from finding powershell.exe.

Fix WSL Windows interop first, then rerun:
  npm run dev

For browser/Linux Wails dev only, run:
  CBMP_ALLOW_LOCAL_WAILS=1 npm run dev
EOF
}

launch_windows_build_from_wsl() {
  local output_name="${WAILS_OUTPUT:-CommonBuildMaterialsPlatform.exe}"
  local process_name="${output_name%.exe}"
  local exe_path="$PROJECT_DIR/build/bin/$output_name"
  local windows_exe_path
  local ps_exe
  local ps_process

  if ! windows_launcher_tools_available; then
    echo "Windows launcher tools were not found in WSL." >&2
    return 1
  fi

  ps_process="$(windows_path_literal "$process_name")"
  run_windows_powershell -Command "Get-Process -Name $ps_process -ErrorAction SilentlyContinue | Stop-Process -Force" >/dev/null 2>&1 || true

  "$WAILS_BIN" build "${wails_tag_args[@]}" -clean -platform windows/amd64 -webview2 "${WAILS_WEBVIEW2:-browser}" -o "$output_name"

  windows_exe_path="$(windows_path_from_wsl "$exe_path")"
  windows_start_exe "$windows_exe_path" "${CBMP_FRONTEND_DEV_URL:-}"
  echo "Started Windows build: $windows_exe_path"
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

wait_for_port_to_close() {
  local port="$1"
  local attempt
  for attempt in {1..40}; do
    if ! port_is_open "$port"; then
      return 0
    fi
    sleep 0.25
  done
  return 1
}

pids_for_listening_port() {
  local port="$1"
  command -v ss >/dev/null 2>&1 || return 1
  ss -ltnp 2>/dev/null | awk -v port="$port" '
    $4 ~ ":" port "$" {
      line = $0
      while (match(line, /pid=[0-9]+/)) {
        print substr(line, RSTART + 4, RLENGTH - 4)
        line = substr(line, RSTART + RLENGTH)
      }
    }
  ' | sort -u
}

stop_project_vite_on_port() {
  local port="$1"
  if [[ "${CBMP_RESTART_FRONTEND:-1}" != "1" ]]; then
    return 0
  fi

  local pids=()
  local pid
  mapfile -t pids < <(pids_for_listening_port "$port" || true)
  if [[ "${#pids[@]}" == "0" ]]; then
    return 0
  fi

  local killed=()
  local command_line
  for pid in "${pids[@]}"; do
    command_line="$(ps -p "$pid" -o args= 2>/dev/null || true)"
    if [[ "$command_line" == *"$PROJECT_DIR"* && "$command_line" == *"vite"* ]]; then
      killed+=("$pid")
    fi
  done

  if [[ "${#killed[@]}" == "0" ]]; then
    return 0
  fi

  echo "Stopping existing ERP Vite dev server on http://127.0.0.1:$port before restart."
  for pid in "${killed[@]}"; do
    kill "$pid" >/dev/null 2>&1 || true
  done
  wait_for_pids_to_exit "${killed[@]}"
  wait_for_port_to_close "$port" || true
}

stop_windows_process_from_wsl() {
  local process_name="$1"
  local ps_process
  ps_process="$(windows_path_literal "$process_name")"
  run_windows_powershell -Command "Get-Process -Name $ps_process -ErrorAction SilentlyContinue | Stop-Process -Force" >/dev/null 2>&1 || true
}

windows_dev_shell_process_is_running() {
  local process_name="$1"
  local ps_process
  ps_process="$(windows_path_literal "$process_name")"
  run_windows_powershell -Command "\$process=$ps_process; if (Get-Process -Name \$process -ErrorAction SilentlyContinue) { exit 0 } else { exit 1 }"
}

windows_dev_shell_http_is_ready() {
  local process_name="$1"
  local health_url
  health_url="$(windows_dev_health_url)"
  if [[ -z "$health_url" ]]; then
    return 0
  fi

  local ps_process
  local ps_url
  ps_process="$(windows_path_literal "$process_name")"
  ps_url="$(windows_path_literal "$health_url")"
  run_windows_powershell -Command "\$ErrorActionPreference='Stop'; \$process=$ps_process; \$url=$ps_url; \$procs=@(Get-Process -Name \$process -ErrorAction SilentlyContinue); if (\$procs.Count -eq 0) { exit 1 }; try { \$response=Invoke-WebRequest -UseBasicParsing -Uri \$url -TimeoutSec 2; if (\$response.StatusCode -ge 200 -and \$response.StatusCode -lt 500) { exit 0 } } catch {}; try { \$ids=@(\$procs | ForEach-Object { \$_.Id }); \$listeners=@(Get-NetTCPConnection -State Listen -ErrorAction SilentlyContinue | Where-Object { \$ids -contains \$_.OwningProcess }); if (\$listeners.Count -gt 0) { exit 0 } } catch {}; exit 1"
}

windows_dev_shell_is_healthy() {
  local process_name="$1"
  windows_dev_shell_process_is_running "$process_name" && windows_dev_shell_http_is_ready "$process_name"
}

wait_for_windows_dev_shell_ready() {
  local process_name="$1"
  local attempt
  for attempt in {1..45}; do
    if windows_dev_shell_is_healthy "$process_name"; then
      return 0
    fi
    sleep 1
  done
  return 1
}

ensure_windows_dev_shell_running() {
  local process_name="$1"
  local exe_path="$2"
  local dev_url="$3"
  local reason="${4:-}"

  if windows_dev_shell_is_healthy "$process_name"; then
    echo "Windows dev shell already running: $process_name"
    return 0
  fi

  if windows_dev_shell_process_is_running "$process_name"; then
    if [[ -n "$reason" ]]; then
      dev_status "$reason; restarting stale Windows dev shell."
    else
      dev_status "Windows dev shell is running but wails.localhost is not healthy; restarting."
    fi
    stop_windows_process_from_wsl "$process_name"
  elif [[ -n "$reason" ]]; then
    dev_status "$reason; relaunching Windows dev shell."
  fi

  windows_start_exe "$exe_path" "$dev_url"
  echo "Started Windows dev shell: $exe_path -> $dev_url"
  if ! wait_for_windows_dev_shell_ready "$process_name"; then
    echo "Windows dev shell launched, but wails.localhost was not ready within 45s." >&2
  fi
}

monitor_windows_dev_shell() {
  local process_name="$1"
  local exe_path="$2"
  local dev_url="$3"
  local interval="${CBMP_WINDOWS_DEV_WATCH_INTERVAL:-5}"

  while true; do
    sleep "$interval"
    if windows_dev_shell_is_healthy "$process_name"; then
      continue
    fi
    ensure_windows_dev_shell_running "$process_name" "$exe_path" "$dev_url" "Windows dev shell stopped serving"
  done
}

cleanup_hot_dev_children() {
  if [[ -n "${WINDOWS_DEV_WATCH_PID:-}" ]]; then
    kill "$WINDOWS_DEV_WATCH_PID" >/dev/null 2>&1 || true
  fi
  if [[ -n "${VITE_DEV_PID:-}" ]]; then
    kill "$VITE_DEV_PID" >/dev/null 2>&1 || true
  fi
}

install_hot_dev_cleanup_trap() {
  trap cleanup_hot_dev_children EXIT INT TERM
}

windows_dev_needs_rebuild() {
  local exe_path="$1"
  if [[ ! -f "$exe_path" || "${CBMP_FORCE_REBUILD:-0}" == "1" ]]; then
    return 0
  fi

  if find . \
    \( -path ./node_modules -o -path ./dist -o -path ./build \) -prune -o \
    \( -name "*.go" -o -name "go.mod" -o -name "go.sum" -o -name "wails.json" \) \
    -newer "$exe_path" -print -quit | grep -q .; then
    return 0
  fi

  if find build \
    \( -path build/appicon.png -o -path build/windows/icon.ico \) \
    -newer "$exe_path" -print -quit 2>/dev/null | grep -q .; then
    return 0
  fi

  return 1
}

ensure_embedded_dist_exists() {
  if [[ -f "$PROJECT_DIR/dist/index.html" ]]; then
    return 0
  fi

  local started_at
  started_at="$(timestamp_seconds)"
  dev_status "Embedded dist is missing; building web assets before Windows shell build."
  npm run build:web
  dev_status "Embedded dist is ready after $(elapsed_seconds "$started_at")."
}

start_vite_dev_server() {
  local port="$1"
  local started_at
  VITE_DEV_PID=""

  if port_is_open "$port"; then
    dev_status "Using existing Vite dev server on http://127.0.0.1:$port"
    return 0
  fi

  started_at="$(timestamp_seconds)"
  dev_status "Starting Vite dev server on http://127.0.0.1:$port"
  VITE_DEV_PORT="$port" npm run dev:web -- --host 127.0.0.1 --port "$port" --strictPort &
  VITE_DEV_PID=$!
  install_hot_dev_cleanup_trap

  if ! wait_for_port "$port"; then
    echo "Timed out waiting for Vite dev server on http://127.0.0.1:$port" >&2
    wait "$VITE_DEV_PID"
    return 1
  fi
  dev_status "Vite dev server is ready after $(elapsed_seconds "$started_at")."
}

launch_windows_hot_from_wsl() {
  local port="${VITE_DEV_PORT:-9245}"
  local dev_url="${CBMP_FRONTEND_DEV_URL:-http://127.0.0.1:$port}"
  local base_output="${WAILS_OUTPUT:-CommonBuildMaterialsPlatform.exe}"
  local output_name="${WAILS_DEV_OUTPUT:-${base_output%.exe}-dev.exe}"
  local process_name="${output_name%.exe}"
  local exe_path="$PROJECT_DIR/build/bin/$output_name"
  local windows_exe_path

  if ! windows_launcher_tools_available; then
    echo "Windows launcher tools were not found in WSL." >&2
    return 1
  fi

  start_vite_dev_server "$port"
  ensure_embedded_dist_exists

  if windows_dev_needs_rebuild "$exe_path"; then
    local build_started_at
    build_started_at="$(timestamp_seconds)"
    dev_status "Windows dev shell rebuild required: $output_name"
    dev_status "Building Windows dev shell; the app window opens after this step."
    stop_windows_process_from_wsl "$process_name"
    "$WAILS_BIN" build "${wails_tag_args[@]}" -s -debug -devtools -platform windows/amd64 -webview2 "${WAILS_WEBVIEW2:-browser}" -o "$output_name"
    dev_status "Windows dev shell build finished after $(elapsed_seconds "$build_started_at")."
  elif [[ "${CBMP_KEEP_DESKTOP:-0}" != "1" ]]; then
    dev_status "Restarting Windows dev shell so it uses $dev_url"
    stop_windows_process_from_wsl "$process_name"
  else
    dev_status "Keeping existing Windows dev shell process."
  fi

  windows_exe_path="$(windows_path_from_wsl "$exe_path")"
  dev_status "Launching Windows app with frontend: $dev_url"
  ensure_windows_dev_shell_running "$process_name" "$windows_exe_path" "$dev_url"
  dev_status "Windows app launch request sent."

  monitor_windows_dev_shell "$process_name" "$windows_exe_path" "$dev_url" &
  WINDOWS_DEV_WATCH_PID=$!
  install_hot_dev_cleanup_trap

  if [[ -n "${VITE_DEV_PID:-}" ]]; then
    wait "$VITE_DEV_PID"
  else
    wait "$WINDOWS_DEV_WATCH_PID"
  fi
}

run_local_wails_dev() {
  "$WAILS_BIN" dev "${wails_tag_args[@]}"
}

if [[ "${1:-}" == "--check-windows-launcher" ]]; then
  if windows_launcher_tools_available; then
    exit 0
  fi
  report_windows_launcher_unavailable
  exit 1
fi

NODE_BIN="$("$ROOT_DIR/scripts/ensure-node.sh")"
WAILS_BIN="$("$ROOT_DIR/scripts/ensure-wails.sh")"
GO_BIN="$("$ROOT_DIR/scripts/ensure-go.sh")"
export PATH="$NODE_BIN:$(dirname "$WAILS_BIN"):$(dirname "$GO_BIN"):$PATH"
export VITE_API_BASE_URL="${VITE_API_BASE_URL:-http://127.0.0.1:8088/api}"
export npm_config_cache="$ROOT_DIR/.tools/npm-cache"

dev_target="${CBMP_DEV_TARGET:-auto}"
if [[ "$dev_target" == "auto" ]] && is_wsl; then
  dev_target="windows-hot"
fi

if [[ ! -d node_modules ]]; then
  npm install
fi

stop_existing_frontend_launcher
if [[ "$dev_target" == "windows-hot" || "$dev_target" == "windows-dev" ]]; then
  stop_project_vite_on_port "${VITE_DEV_PORT:-9245}"
fi

acquire_single_instance_lock

if [[ "$dev_target" == "windows-hot" || "$dev_target" == "windows-dev" ]]; then
  if windows_launcher_tools_available; then
    launch_windows_hot_from_wsl
  elif [[ "${CBMP_ALLOW_LOCAL_WAILS:-0}" == "1" ]]; then
    report_windows_launcher_unavailable
    run_local_wails_dev
  else
    report_windows_launcher_unavailable
    exit 1
  fi
  exit 0
fi

if [[ "$dev_target" == "windows" || "$dev_target" == "windows-build" ]]; then
  if windows_launcher_tools_available; then
    launch_windows_build_from_wsl
  elif [[ "${CBMP_ALLOW_LOCAL_WAILS:-0}" == "1" ]]; then
    report_windows_launcher_unavailable
    run_local_wails_dev
  else
    report_windows_launcher_unavailable
    exit 1
  fi
  exit 0
fi

run_local_wails_dev
