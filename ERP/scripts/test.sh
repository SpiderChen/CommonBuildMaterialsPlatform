#!/usr/bin/env bash
set -euo pipefail

ERP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKSPACE_DIR="$(cd "$ERP_DIR/.." && pwd)"
"$ERP_DIR/backend/scripts/test.sh"
"$ERP_DIR/frontend/scripts/test.sh"
"$WORKSPACE_DIR/industrial-control-gateway/scripts/test.sh"
