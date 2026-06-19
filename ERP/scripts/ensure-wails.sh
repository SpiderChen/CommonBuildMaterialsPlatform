#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_BIN="$("$ROOT_DIR/scripts/ensure-go.sh")"
TOOL_BIN="$ROOT_DIR/.tools/bin"
WAILS_BIN="$TOOL_BIN/wails"

if command -v wails >/dev/null 2>&1; then
  command -v wails
  exit 0
fi

if [[ -x "$WAILS_BIN" ]]; then
  echo "$WAILS_BIN"
  exit 0
fi

mkdir -p "$TOOL_BIN"
echo "Installing Wails CLI to project-local .tools/bin..." >&2
GOBIN="$TOOL_BIN" "$GO_BIN" install github.com/wailsapp/wails/v2/cmd/wails@latest
echo "$WAILS_BIN"

