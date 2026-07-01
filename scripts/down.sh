#!/usr/bin/env bash
# scripts/down.sh — tear a stack down and remove its volumes.
#
# Usage: scripts/down.sh [project-name]
#
# The -v matters: throwaway instances would otherwise leave orphaned volumes.

set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"

detect_runtime
load_env

PROJECT="$(resolve_project "${1:-}")"
log "tearing down instance: $PROJECT"

dc -p "$PROJECT" down -v --remove-orphans || true

rm -f "$REPO_ROOT/tunnel-urls.txt"
ok "instance '$PROJECT' removed (containers + volumes)."
