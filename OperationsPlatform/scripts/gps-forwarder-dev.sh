#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

export GPSF_ADDR="${GPSF_ADDR:-0.0.0.0:19102}"

cd "$ROOT/backend"
go run ./cmd/gps-forwarder
