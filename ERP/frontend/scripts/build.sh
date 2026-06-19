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
"$WAILS_BIN" build
