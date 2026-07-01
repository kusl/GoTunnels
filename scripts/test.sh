#!/usr/bin/env bash
# scripts/test.sh — run the Go test suite, vet, and vulnerability checks.
#
# All CI logic lives here so the GitHub Actions workflow stays a thin caller.
# Runs directly with the host Go toolchain when present; otherwise runs inside
# a golang container so no host Go install is required.
#
# Usage:
#   scripts/test.sh [unit|vet|vuln|build|all]   (default: all)

set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"

GO_IMAGE="${GO_IMAGE:-docker.io/library/golang:1.23-bookworm}"

# run_go "<go shell snippet>" — execute a snippet with a Go toolchain, on host
# if available else in a container mounting the repo.
run_go() {
  local snippet="$1"
  if command -v go >/dev/null 2>&1; then
    ( cd "$REPO_ROOT" && GOFLAGS=-mod=mod bash -c "$snippet" )
  else
    detect_runtime
    local zflag=""
    # SELinux relabel on Fedora/RHEL when using podman.
    if [ "$CR" = "podman" ]; then zflag=":Z"; fi
    "$CR" run --rm \
      -v "$REPO_ROOT":/src${zflag} \
      -w /src \
      -e GOFLAGS=-mod=mod \
      -e GOTOOLCHAIN=local \
      "$GO_IMAGE" \
      bash -c "$snippet"
  fi
}

cmd_unit() {
  log "running unit tests…"
  run_go "go mod tidy && go test ./... -count=1"
  ok "unit tests passed"
}

cmd_vet() {
  log "running go vet…"
  run_go "go mod tidy && go vet ./..."
  ok "go vet clean"
}

cmd_build() {
  log "verifying the API builds…"
  run_go "go mod tidy && go build ./..."
  ok "build ok"
}

cmd_vuln() {
  log "running govulncheck (reachability-aware)…"
  run_go "go install golang.org/x/vuln/cmd/govulncheck@latest && \$(go env GOPATH)/bin/govulncheck ./..."
  ok "govulncheck clean"
}

cmd_freshness() {
  log "checking dependency freshness (update + tidy + test)…"
  run_go "go get -u ./... && go mod tidy && go build ./... && go test ./... -count=1"
  ok "dependencies update cleanly and tests still pass"
}

main() {
  local target="${1:-all}"
  case "$target" in
    unit)      cmd_unit ;;
    vet)       cmd_vet ;;
    build)     cmd_build ;;
    vuln)      cmd_vuln ;;
    freshness) cmd_freshness ;;
    all)       cmd_build; cmd_vet; cmd_unit ;;
    *)         die "unknown target '$target' (use: unit|vet|vuln|build|freshness|all)" ;;
  esac
}

main "$@"
