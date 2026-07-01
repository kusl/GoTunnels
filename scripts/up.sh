#!/usr/bin/env bash
# scripts/up.sh — bring the whole stack up, in the staged order that lets the
# frontend and API each get a Quick Tunnel URL and lets the API be configured
# with the correct WebAuthn RP ID / CORS origin (both derived from the
# frontend's runtime URL). Safe to run for multiple instances concurrently by
# passing a distinct project name.
#
# Usage:
#   scripts/up.sh [project-name]
#
# Environment:
#   GOTUNNELS_PROJECT / GOTUNNELS_INSTANCE_ID  alternative ways to name the run
#   UPTRACE_DSN                                optional telemetry DSN

set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"

detect_runtime
ensure_env

PROJECT="$(resolve_project "${1:-}")"
export GOTUNNELS_INSTANCE_ID="${GOTUNNELS_INSTANCE_ID:-$PROJECT}"
set_env_var GOTUNNELS_INSTANCE_ID "$GOTUNNELS_INSTANCE_ID"
log "project (instance): $PROJECT"

# 1) Build images (API multi-stage build runs go mod tidy + go build).
log "building images…"
dc -p "$PROJECT" build

# 2) Database first, wait until healthy.
log "starting database…"
dc -p "$PROJECT" up -d db
wait_healthy "$PROJECT" db 120

# 3) Frontend + its tunnel, WITHOUT pulling in the api dependency yet.
log "starting frontend and its tunnel…"
dc -p "$PROJECT" up -d --no-deps frontend cloudflared-frontend

# 4) Discover the frontend's public URL.
log "waiting for the frontend Quick Tunnel URL…"
FRONTEND_URL="$(poll_tunnel_url "$PROJECT" cloudflared-frontend 90)" \
  || die "timed out waiting for frontend tunnel URL (check: dc -p $PROJECT logs cloudflared-frontend)"
FRONTEND_HOST="$(host_of_url "$FRONTEND_URL")"
ok "frontend: $FRONTEND_URL"

# 5) Configure WebAuthn RP + CORS from the frontend origin, persist, and export.
set_env_var GOTUNNELS_RP_ID "$FRONTEND_HOST"
set_env_var GOTUNNELS_RP_ORIGINS "$FRONTEND_URL"
set_env_var GOTUNNELS_CORS_ALLOWED_ORIGINS "$FRONTEND_URL"
export GOTUNNELS_RP_ID="$FRONTEND_HOST"
export GOTUNNELS_RP_ORIGINS="$FRONTEND_URL"
export GOTUNNELS_CORS_ALLOWED_ORIGINS="$FRONTEND_URL"

# 6) Now start the API (with correct RP/CORS) and its tunnel.
log "starting API and its tunnel…"
dc -p "$PROJECT" up -d --no-deps api cloudflared-api

# 7) API readiness (best-effort) then discover its public URL.
wait_for_log "$PROJECT" api 'http server listening' 60 || warn "did not observe API listening log yet"
log "waiting for the API Quick Tunnel URL…"
API_URL="$(poll_tunnel_url "$PROJECT" cloudflared-api 90)" \
  || die "timed out waiting for API tunnel URL (check: dc -p $PROJECT logs cloudflared-api)"
ok "api: $API_URL"

# 8) Tell the frontend where the API lives (runtime config.json).
write_frontend_config "$PROJECT" "$API_URL"

# 9) Report.
URLS_FILE="$REPO_ROOT/tunnel-urls.txt"
{
  echo "instance=$PROJECT"
  echo "frontend=$FRONTEND_URL"
  echo "api=$API_URL"
} > "$URLS_FILE"

echo >&2
ok "GoTunnels is up."
printf '  %sWeb app :%s %s\n' "$_c_grn" "$_c_reset" "$FRONTEND_URL" >&2
printf '  %sAPI     :%s %s\n' "$_c_grn" "$_c_reset" "$API_URL" >&2
printf '  %sInstance:%s %s (urls saved to tunnel-urls.txt)\n' "$_c_dim" "$_c_reset" "$PROJECT" >&2
echo >&2
log "tear down with: scripts/down.sh $PROJECT"
