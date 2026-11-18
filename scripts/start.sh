#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

export NEXO_DATA_DIR="${NEXO_DATA_DIR:-$HOME/.local/share/nexo}"
export NEXO_HTTP_ADDR="${NEXO_HTTP_ADDR:-:8080}"
export NEXO_AIP_ADDR="${NEXO_AIP_ADDR:-:8081}"
export NEXO_ADMIN_TOKEN="${NEXO_ADMIN_TOKEN:-nexo-admin-change-me}"
export NEXO_NODE_NAME="${NEXO_NODE_NAME:-nexo-node}"
export NEXO_WEB_DIR="${NEXO_WEB_DIR:-$ROOT/web}"
export GOPROXY="${GOPROXY:-https://proxy.golang.org,direct}"

mkdir -p "$NEXO_DATA_DIR"
echo "NEXO · data=$NEXO_DATA_DIR http=$NEXO_HTTP_ADDR aip=$NEXO_AIP_ADDR"
exec go run ./cmd/nexo
