#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NODE_VERSION="${NODE_VERSION:-20.19.4}"

if command -v node >/dev/null 2>&1 && command -v npm >/dev/null 2>&1; then
  dirname "$(command -v node)"
  exit 0
fi

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="x64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac

case "$os" in
  darwin) platform="darwin-${arch}" ;;
  linux) platform="linux-${arch}" ;;
  *) echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

tool_root="$ROOT_DIR/.tools"
node_root="$tool_root/node-v${NODE_VERSION}-${platform}"
node_bin="$node_root/bin"

if [[ -x "$node_bin/node" && -x "$node_bin/npm" ]]; then
  echo "$node_bin"
  exit 0
fi

mkdir -p "$tool_root/cache"
archive="$tool_root/cache/node-v${NODE_VERSION}-${platform}.tar.xz"
url="https://nodejs.org/dist/v${NODE_VERSION}/node-v${NODE_VERSION}-${platform}.tar.xz"

echo "Downloading Node ${NODE_VERSION} to project-local .tools directory..." >&2
if [[ -f "$archive" ]]; then
  if ! tar -tJf "$archive" >/dev/null 2>&1; then
    rm -f "$archive"
  fi
fi

if [[ ! -f "$archive" ]]; then
  curl -L --fail "$url" -o "$archive"
fi

rm -rf "$node_root"
tar -C "$tool_root" -xJf "$archive"
echo "$node_bin"
