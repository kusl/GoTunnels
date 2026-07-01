#!/usr/bin/env bash
# scripts/run.sh — the single entrypoint.
#
# It does, in order:
#   1. regenerates the LLM context dump (export.sh), if present
#   2. runs the full test suite (build + vet + unit tests)
#   3. builds the images and brings the whole stack up, printing the two
#      Cloudflare Quick Tunnel URLs
#
# So the entire "make it real" flow is a single command:
#   ./scripts/run.sh
#
# Usage:
#   scripts/run.sh [project-name] [--skip-tests] [--skip-export]

set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"

PROJECT=""
SKIP_TESTS=0
SKIP_EXPORT=0
for arg in "$@"; do
  case "$arg" in
    --skip-tests)  SKIP_TESTS=1 ;;
    --skip-export) SKIP_EXPORT=1 ;;
    -*)            die "unknown flag: $arg" ;;
    *)             PROJECT="$arg" ;;
  esac
done

# 1) Refresh the code dump used to share status with the LLM.
if [ "$SKIP_EXPORT" -eq 0 ] && [ -f "$REPO_ROOT/export.sh" ]; then
  log "regenerating docs/llm/dump.txt via export.sh"
  bash "$REPO_ROOT/export.sh" >/dev/null 2>&1 || warn "export.sh failed (continuing)"
fi

# 2) Tests must pass before we stand anything up.
if [ "$SKIP_TESTS" -eq 0 ]; then
  log "running test suite before startup…"
  bash "$SCRIPT_DIR/test.sh" all
else
  warn "skipping tests (--skip-tests)"
fi

# 3) Build + run the stack.
log "bringing the stack up…"
bash "$SCRIPT_DIR/up.sh" "$PROJECT"
