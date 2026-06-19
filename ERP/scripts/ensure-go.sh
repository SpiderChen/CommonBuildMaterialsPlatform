#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_VERSION="${GO_VERSION:-1.25.0}"

if command -v go >/dev/null 2>&1; then
  command -v go
  exit 0
fi

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac

case "$os" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $os" >&2; exit 1 ;;
esac

tool_root="$ROOT_DIR/.tools"
go_root="$tool_root/go"
go_bin="$go_root/bin/go"

if [[ -x "$go_bin" ]]; then
  echo "$go_bin"
  exit 0
fi

mkdir -p "$tool_root/cache"
archive="$tool_root/cache/go${GO_VERSION}.${os}-${arch}.tar.gz"
url="https://go.dev/dl/go${GO_VERSION}.${os}-${arch}.tar.gz"

echo "Downloading Go ${GO_VERSION} to project-local .tools directory..." >&2
if [[ -f "$archive" ]]; then
  if ! tar -tzf "$archive" >/dev/null 2>&1; then
    rm -f "$archive"
  fi
fi

if [[ ! -f "$archive" ]]; then
  curl -L --fail "$url" -o "$archive"
fi

rm -rf "$go_root"
tar -C "$tool_root" -xzf "$archive"
echo "$go_bin"
