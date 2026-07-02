#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT_DIR="$(cd "$PROJECT_DIR/.." && pwd)"
NODE_BIN="$("$ROOT_DIR/scripts/ensure-node.sh")"
WAILS_BIN="$("$ROOT_DIR/scripts/ensure-wails.sh")"
GO_BIN="$("$ROOT_DIR/scripts/ensure-go.sh")"
export PATH="$NODE_BIN:$(dirname "$WAILS_BIN"):$(dirname "$GO_BIN"):$PATH"
export VITE_API_BASE_URL="${VITE_API_BASE_URL:-http://127.0.0.1:8088/api}"
export npm_config_cache="$ROOT_DIR/.tools/npm-cache"

cd "$PROJECT_DIR"
if [[ ! -d node_modules ]]; then
  npm install
fi

wails_tag_args=()
if [[ "$(uname -s)" == "Linux" ]]; then
  if ! pkg-config --exists webkit2gtk-4.0 2>/dev/null && pkg-config --exists webkit2gtk-4.1 2>/dev/null; then
    wails_tag_args=(-tags webkit2_41)
  fi
fi
if [[ -n "${WAILS_TAGS:-}" ]]; then
  wails_tag_args=(-tags "$WAILS_TAGS")
fi

"$WAILS_BIN" build "${wails_tag_args[@]}"
