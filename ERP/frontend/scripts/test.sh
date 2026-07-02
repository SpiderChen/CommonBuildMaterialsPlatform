#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT_DIR="$(cd "$PROJECT_DIR/.." && pwd)"
GO_BIN="$("$ROOT_DIR/scripts/ensure-go.sh")"

cd "$PROJECT_DIR"
go_tag_args=()
if [[ "$(uname -s)" == "Linux" ]]; then
  if ! pkg-config --exists webkit2gtk-4.0 2>/dev/null && pkg-config --exists webkit2gtk-4.1 2>/dev/null; then
    go_tag_args=(-tags webkit2_41)
  fi
fi
if [[ -n "${WAILS_TAGS:-}" ]]; then
  go_tag_args=(-tags "$WAILS_TAGS")
fi

"$GO_BIN" test "${go_tag_args[@]}" ./...
