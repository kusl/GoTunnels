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
#   GOTUNNELS_TUNNEL_LOG_WAIT                  seconds to wait before printing
#                                              the tunnel URL log lines (def 60)

set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"

detect_runtime
ensure_env
# Quick Tunnel URLs are ephemeral: whatever last run wrote into .env for
# RP_ID / RP_ORIGINS / CORS is stale now. Reset them to bootstrap defaults so
# nothing (including a dependency-started api container, see step 6) can boot
# against a dead tunnel's origin. This removes the need to delete .env.
reset_runtime_env

PROJECT="$(resolve_project "${1:-}")"
export GOTUNNELS_INSTANCE_ID="${GOTUNNELS_INSTANCE_ID:-$PROJECT}"
set_env_var GOTUNNELS_INSTANCE_ID "$GOTUNNELS_INSTANCE_ID"
log "project (instance): $PROJECT"

# Telemetry host.name: the OTel SDK does not detect the host by default, so
# without this the backend shows host_name as an empty string. Export the
# machine's hostname (e.g. "virginia") so every span/metric/log says which
# box the stack runs on. Override by exporting GOTUNNELS_HOST_NAME yourself.
GOTUNNELS_HOST_NAME="${GOTUNNELS_HOST_NAME:-$(hostname)}"
export GOTUNNELS_HOST_NAME
set_env_var GOTUNNELS_HOST_NAME "$GOTUNNELS_HOST_NAME"
log "telemetry host.name: $GOTUNNELS_HOST_NAME"

# 1) Build images (API multi-stage build runs go mod tidy + go build).
log "building images…"
dc -p "$PROJECT" build

# 2) Database first, wait until it actually accepts connections.
#    We pass an explicit `pg_isready` probe so readiness does NOT hinge on
#    podman's health *timer* firing (it often doesn't in a rootless shell) — see
#    the long note on wait_healthy in lib.sh. 60s is plenty for a fresh volume.
log "starting database…"
dc -p "$PROJECT" up -d db
wait_healthy "$PROJECT" db 60 \
  pg_isready -U "${POSTGRES_USER:-gotunnels}" -d "${POSTGRES_DB:-gotunnels}" -q

# 2b) Clear any stale api container from a previous run of THIS project,
#     BEFORE the frontend and tunnels come up. Doing it this early matters:
#     right now nothing has been discovered yet, so if removing the api drags
#     dependent containers with it (instances created before the compose fix
#     carry podman --requires links from frontend/cloudflared-api onto the
#     api container), those dependents are simply recreated fresh by the
#     steps below. Waiting until after tunnel discovery to do this — as an
#     earlier version of this script did — is what produced the virginia
#     failure: `podman rm -f api` failed with "has dependent containers",
#     the `|| true` swallowed it, `up -d api` then hit "name already in use",
#     and the API silently kept its bootstrap env (RP_ID=localhost, CORS=*):
#     everything worked except passkeys.
_stale_api_cid="$(cid_of "$PROJECT" api)"
if [ -n "$_stale_api_cid" ]; then
  warn "found an api container from a previous run; removing it (and any podman-level dependents) so it can be recreated with this run's environment"
  "$CR" rm -f "$_stale_api_cid" >/dev/null 2>&1 || true
  if [ -n "$(cid_of "$PROJECT" api)" ]; then
    # Old-format instances: dependents hold --requires on the api container.
    # podman's --depend removes them too; they are recreated fresh below.
    "$CR" rm -f --depend "$_stale_api_cid" >/dev/null 2>&1 || true
  fi
  if [ -n "$(cid_of "$PROJECT" api)" ]; then
    die "could not remove the stale api container ($_stale_api_cid). Tear the instance down first: scripts/down.sh $PROJECT"
  fi
  ok "stale api container removed"
fi

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
#
#    compose.yaml deliberately has no depends_on -> api edges anymore, and
#    step 2b removed any leftover api container, so nothing can have
#    pre-created the api with stale env by this point. Verify that instead of
#    assuming it: if an api container somehow exists here, recreating it is
#    NOT safe to skip — a container created before the exports above would
#    run with RP_ID=localhost and CORS=*, the exact state where everything
#    works except passkeys. Fail loudly rather than continue into that trap.
if [ -n "$(cid_of "$PROJECT" api)" ]; then
  die "an api container already exists before 'up -d api' — it would keep pre-discovery env. Tear down and retry: scripts/down.sh $PROJECT && scripts/up.sh $PROJECT"
fi
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

# 10) Final check. Cloudflare Quick Tunnels can take a little while after the
#     containers start to finish registering and become reachable. Give them a
#     moment, then re-read the cloudflared container logs with the runtime
#     ($CR, e.g. podman) and print the raw log line(s) that announce each
#     https://<subdomain>.trycloudflare.com URL. That line is the signal the
#     tunnel is live — and it is the exact line to copy/paste.
#
#     Override the pause with GOTUNNELS_TUNNEL_LOG_WAIT (seconds); set it to 0
#     to skip waiting entirely.
FINAL_WAIT="${GOTUNNELS_TUNNEL_LOG_WAIT:-60}"
if [ "$FINAL_WAIT" -gt 0 ] 2>/dev/null; then
  log "waiting ${FINAL_WAIT}s for the Quick Tunnels to settle, then printing their URL log lines…"
  sleep "$FINAL_WAIT"
fi

echo >&2
# "service|human label" pairs. An explicit for-loop list is not subject to IFS
# word splitting, and the ${var%%|*} / ${var##*|} expansions do not use IFS
# either, so this is safe under IFS=$'\n\t'.
for _pair in "cloudflared-frontend|Web app" "cloudflared-api|API"; do
  _svc="${_pair%%|*}"
  _label="${_pair##*|}"
  _cid="$(cid_of "$PROJECT" "$_svc")"
  if [ -z "$_cid" ]; then
    warn "$_label: container '$_svc' not found; cannot read its logs"
    continue
  fi
  # grep exits non-zero when it matches nothing; '|| true' keeps that from
  # tripping 'set -e' on this final, informational step.
  _lines="$("$CR" logs "$_cid" 2>&1 | grep -E 'https://[a-z0-9._-]+\.trycloudflare\.com' || true)"
  if [ -n "$_lines" ]; then
    ok "$_label — tunnel URL (from: $CR logs $_svc):"
    printf '%s\n' "$_lines" >&2
  else
    warn "$_label: no trycloudflare.com URL in '$_svc' logs yet (check: dc -p $PROJECT logs $_svc)"
  fi
  echo >&2
done

