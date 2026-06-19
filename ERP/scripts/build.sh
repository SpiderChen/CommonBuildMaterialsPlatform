#!/usr/bin/env bash
set -euo pipefail

ERP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORKSPACE_DIR="$(cd "$ERP_DIR/.." && pwd)"

cd "$ERP_DIR"
mkdir -p dist
"$ERP_DIR/frontend/scripts/web-build.sh"
"$ERP_DIR/frontend/scripts/updater-build.sh"
"$ERP_DIR/backend/scripts/build.sh"
"$WORKSPACE_DIR/industrial-control-gateway/scripts/build.sh"
rm -rf dist/backend dist/frontend dist/industrial-control-gateway dist/gps-forwarder
mkdir -p dist/backend dist/frontend dist/industrial-control-gateway
cp backend/dist/cbmp-appliance dist/backend/cbmp-appliance
cp backend/dist/cbmp-server-updater dist/backend/cbmp-server-updater
cp -R frontend/dist dist/frontend/web
cp frontend/build/bin/cbmp-client-updater dist/frontend/cbmp-client-updater
cp "$WORKSPACE_DIR/industrial-control-gateway/dist/industrial-control-gateway" dist/industrial-control-gateway/industrial-control-gateway
cp "$WORKSPACE_DIR/industrial-control-gateway/README.md" dist/industrial-control-gateway/README.md
cp "$WORKSPACE_DIR/README.md" dist/README.md
if [[ -f "$WORKSPACE_DIR/docs/DELIVERY.md" ]]; then
  cp "$WORKSPACE_DIR/docs/DELIVERY.md" dist/DELIVERY.md
fi
if [[ -f "$WORKSPACE_DIR/docs/FUNCTION_AUDIT.md" ]]; then
  cp "$WORKSPACE_DIR/docs/FUNCTION_AUDIT.md" dist/FUNCTION_AUDIT.md
fi
if [[ -f "$WORKSPACE_DIR/docs/DEPLOYMENT.md" ]]; then
  cp "$WORKSPACE_DIR/docs/DEPLOYMENT.md" dist/DEPLOYMENT.md
fi
if [[ -d deploy ]]; then
  rm -rf dist/deploy
  mkdir -p dist/deploy
  cp deploy/* dist/deploy/
fi
echo "Built dist/backend, dist/frontend and dist/industrial-control-gateway"
