#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

export CBM_OPS_ADDR="${CBM_OPS_ADDR:-:8095}"
export CBM_OPS_DATA="${CBM_OPS_DATA:-$ROOT/backend/data/ops.json}"
export CBM_OPS_FRONTEND_DIR="${CBM_OPS_FRONTEND_DIR:-$ROOT/frontend}"

cd "$ROOT/backend"
go run ./cmd/server
