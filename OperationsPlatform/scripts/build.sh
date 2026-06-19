#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

mkdir -p "$ROOT/backend/dist"
cd "$ROOT/backend"
go build -o "$ROOT/backend/dist/cbm-ops" ./cmd/server
go build -o "$ROOT/backend/dist/gps-forwarder" ./cmd/gps-forwarder
