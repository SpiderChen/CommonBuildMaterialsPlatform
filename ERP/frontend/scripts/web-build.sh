#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT_DIR="$(cd "$PROJECT_DIR/.." && pwd)"
NODE_BIN="$("$ROOT_DIR/scripts/ensure-node.sh")"
export PATH="$NODE_BIN:$PATH"
export npm_config_cache="$ROOT_DIR/.tools/npm-cache"

cd "$PROJECT_DIR"
if [[ ! -d node_modules ]]; then
  npm install
fi
npm run build
