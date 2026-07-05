#!/usr/bin/env bash
# scripts/ci-container-test.sh — build the Containerfiles and exercise the real
# stack (db + api + frontend, no Cloudflare tunnels) end to end over HTTP.
#
# This is the piece plain `scripts/test.sh` cannot cover: that the images
# actually build, that compose wiring + env substitution is right, that
# migrations run against real Postgres, and that the HTTP surface behaves —
# including a regression test for the captcha sync 500 (`operator is not
# unique: unknown + unknown`), which only ever reproduced against a real
# Postgres because unit tests never hit the pgx extended protocol.
#
# Runs identically on a laptop and in GitHub Actions:
#   bash scripts/ci-container-test.sh [project-name]     (default: gotunnels-ci)
#
# No host ports are published; all HTTP assertions run from a throwaway curl
# container attached to the instance's compose network. The api image is
# distroless (no shell), so an external HTTP driver is the only option anyway.

set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"

detect_runtime
ensure_env
reset_runtime_env

PROJECT="$(resolve_project "${1:-gotunnels-ci}")"
export GOTUNNELS_INSTANCE_ID="${GOTUNNELS_INSTANCE_ID:-$PROJECT}"
NET="${PROJECT}_default"
# Pinned, fully qualified (podman does not assume docker.io). busybox wget in
# the caddy image can't do PUT/DELETE, hence a real curl.
CURL_IMAGE="docker.io/curlimages/curl:8.10.1"

log "container test instance: $PROJECT"

# ---------------------------------------------------------------------------
# teardown + failure diagnostics
# ---------------------------------------------------------------------------
cleanup() {
  status=$?
  if [ "$status" -ne 0 ] && [ -n "${CR:-}" ]; then
    err "container test FAILED (exit $status) — dumping service logs"
    for _svc in db api frontend; do
      _cid="$(cid_of "$PROJECT" "$_svc")"
      if [ -n "$_cid" ]; then
        err "----- logs: $_svc -----"
        "$CR" logs --tail 120 "$_cid" 2>&1 | sed 's/^/    /' >&2 || true
      else
        err "----- logs: $_svc ----- (no container)"
      fi
    done
  fi
  if [ -n "${CR:-}" ]; then
    log "tearing down CI instance $PROJECT"
    dc -p "$PROJECT" down -v --remove-orphans >/dev/null 2>&1 || true
  fi
  exit "$status"
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# assertion helpers
# ---------------------------------------------------------------------------
FAILED=0

# ccurl — run curl inside a one-shot container on the compose network. -sS so
# transport errors surface; --max-time caps a wedged request, not the test.
ccurl() {
  "$CR" run --rm --network "$NET" "$CURL_IMAGE" -sS --max-time 30 "$@"
}

# assert_contains BODY NEEDLE LABEL — grep must DRAIN stdin (>/dev/null, not
# -q): with -q it exits at first match, printf takes SIGPIPE, and under
# pipefail a PASS would read as FAIL. Same trap as wait_for_log in lib.sh.
assert_contains() {
  if printf '%s' "$1" | grep -F -- "$2" >/dev/null; then
    ok "PASS: $3"
  else
    err "FAIL: $3"
    err "  expected to find: $2"
    printf '%s\n' "$1" | head -20 | sed 's/^/    got: /' >&2
    FAILED=1
  fi
}

assert_icontains() { # case-insensitive variant (headers, HTML)
  if printf '%s' "$1" | grep -Fi -- "$2" >/dev/null; then
    ok "PASS: $3"
  else
    err "FAIL: $3"
    err "  expected to find (case-insensitive): $2"
    printf '%s\n' "$1" | head -20 | sed 's/^/    got: /' >&2
    FAILED=1
  fi
}

assert_not_contains() {
  if printf '%s' "$1" | grep -F -- "$2" >/dev/null; then
    err "FAIL: $3"
    err "  expected NOT to find: $2"
    FAILED=1
  else
    ok "PASS: $3"
  fi
}

assert_eq() { # GOT WANT LABEL
  if [ "$1" = "$2" ]; then
    ok "PASS: $3"
  else
    err "FAIL: $3 (got '$1', want '$2')"
    FAILED=1
  fi
}

# ---------------------------------------------------------------------------
# 1) build both images
# ---------------------------------------------------------------------------
log "building images (Containerfile.api + Containerfile.frontend)…"
dc -p "$PROJECT" build

# ---------------------------------------------------------------------------
# 2) database, then api + frontend (no tunnels in CI)
# ---------------------------------------------------------------------------
log "starting database…"
dc -p "$PROJECT" up -d db
wait_healthy "$PROJECT" db 90 \
  pg_isready -U "${POSTGRES_USER:-gotunnels}" -d "${POSTGRES_DB:-gotunnels}" -q

log "starting api + frontend…"
# podman-compose ignores --no-deps and will re-process db; it prints a
# "name already in use" for the existing container but exits 0, so this is
# safe under both runtimes (docker compose honors --no-deps outright).
dc -p "$PROJECT" up -d --no-deps api frontend

# Unlike up.sh (best-effort warn), CI must hard-fail if the API never comes up.
wait_for_log "$PROJECT" api 'http server listening' 90 \
  || die "api never logged 'http server listening' — startup failed"
ok "api is listening"

# ---------------------------------------------------------------------------
# 3) migrations actually ran
# ---------------------------------------------------------------------------
db_cid="$(cid_of "$PROJECT" db)"
[ -n "$db_cid" ] || die "no db container id"
mig="$("$CR" exec "$db_cid" psql -U "${POSTGRES_USER:-gotunnels}" -d "${POSTGRES_DB:-gotunnels}" \
  -tAc 'SELECT COALESCE(max(version), 0) FROM schema_migrations' | tr -d '[:space:]')"
log "schema_migrations max(version) = $mig"
if [ "${mig:-0}" -ge 7 ] 2>/dev/null; then
  ok "PASS: migrations applied (>= 7)"
else
  err "FAIL: expected migration version >= 7, got '$mig'"
  FAILED=1
fi

# ---------------------------------------------------------------------------
# 4) health endpoints
# ---------------------------------------------------------------------------
ccurl -f http://api:8080/healthz >/dev/null && ok "PASS: /healthz"
ccurl -f http://api:8080/readyz  >/dev/null && ok "PASS: /readyz (db reachable)"

# ---------------------------------------------------------------------------
# 5) signup -> bearer token
# ---------------------------------------------------------------------------
signup_resp="$(ccurl -X POST http://api:8080/api/signup \
  -H 'Content-Type: application/json' \
  -d '{"username":"ciuser","password":"ci-password-123","display_name":"CI User"}')"
TOKEN="$(printf '%s' "$signup_resp" | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')"
if [ -n "$TOKEN" ]; then
  ok "PASS: signup issued a session token"
else
  err "FAIL: signup did not return a token"
  printf '%s\n' "$signup_resp" | head -5 | sed 's/^/    got: /' >&2
  die "cannot continue without a session token"
fi
AUTH="Authorization: Bearer $TOKEN"

# ---------------------------------------------------------------------------
# 6) captcha sync — REGRESSION for the 'unknown + unknown' 500.
#    Before the ::bigint casts in store.SyncCaptchaStats, every call to this
#    endpoint failed with {"error":"internal server error"}; both the insert
#    and the update paths below would have 500ed.
# ---------------------------------------------------------------------------
sync1="$(ccurl -X POST http://api:8080/api/captcha/sync \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"manual_delta":3,"auto_delta":5,"current_streak":2,"best_streak":4}')"
assert_contains "$sync1" '"total_solves":8'   "captcha sync insert path (3 manual + 5 auto = 8)"
assert_contains "$sync1" '"best_streak":4'    "captcha sync insert path records best streak"
assert_not_contains "$sync1" 'internal server error' "captcha sync does not 500 (insert)"

sync2="$(ccurl -X POST http://api:8080/api/captcha/sync \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"manual_delta":4,"auto_delta":8,"current_streak":1,"best_streak":2}')"
assert_contains "$sync2" '"total_solves":20'  "captcha sync update path accumulates (8 + 12 = 20)"
assert_contains "$sync2" '"best_streak":4'    "captcha sync keeps GREATEST best streak (4 > 2)"
assert_contains "$sync2" '"current_streak":1' "captcha sync current streak is last-write-wins"
assert_not_contains "$sync2" 'internal server error' "captcha sync does not 500 (update)"

stats="$(ccurl -H "$AUTH" http://api:8080/api/captcha/stats)"
assert_contains "$stats" '"total_solves":20' "captcha stats reads back the synced totals"

lb="$(ccurl -H "$AUTH" http://api:8080/api/captcha/leaderboard)"
assert_contains "$lb" '"username":"ciuser"' "leaderboard ranks the CI user"

# ---------------------------------------------------------------------------
# 7) prefs round trip
# ---------------------------------------------------------------------------
put_pref="$(ccurl -X PUT http://api:8080/api/prefs/theme \
  -H "$AUTH" -H 'Content-Type: application/json' -d '{"value":"dark"}')"
assert_contains "$put_pref" '"status":"saved"' "prefs PUT saves"

get_pref="$(ccurl -H "$AUTH" http://api:8080/api/prefs/theme)"
assert_contains "$get_pref" '"value":"dark"' "prefs GET returns the stored value"
assert_contains "$get_pref" '"exists":true'  "prefs GET reports exists=true"

# ---------------------------------------------------------------------------
# 8) notes create / list / delete
# ---------------------------------------------------------------------------
note_resp="$(ccurl -X POST http://api:8080/api/notes \
  -H "$AUTH" -H 'Content-Type: application/json' \
  -d '{"body":"hello from the CI smoke test"}')"
NOTE_ID="$(printf '%s' "$note_resp" | sed -n 's/.*"note":{"id":\([0-9][0-9]*\).*/\1/p')"
if [ -n "$NOTE_ID" ]; then
  ok "PASS: note created (id=$NOTE_ID)"
else
  err "FAIL: note creation did not return a numeric id"
  printf '%s\n' "$note_resp" | head -5 | sed 's/^/    got: /' >&2
  FAILED=1
fi

notes_list="$(ccurl -H "$AUTH" http://api:8080/api/notes)"
assert_contains "$notes_list" 'hello from the CI smoke test' "notes list contains the new note"

if [ -n "$NOTE_ID" ]; then
  del_resp="$(ccurl -X DELETE "http://api:8080/api/notes/$NOTE_ID" -H "$AUTH")"
  assert_contains "$del_resp" '"status":"deleted"' "note delete (hard delete semantics)"
  notes_after="$(ccurl -H "$AUTH" http://api:8080/api/notes)"
  assert_not_contains "$notes_after" 'hello from the CI smoke test' "deleted note is gone from the list"
fi

# ---------------------------------------------------------------------------
# 9) CORS preflight (the middleware must answer OPTIONS itself with 204)
# ---------------------------------------------------------------------------
pre_code="$(ccurl -o /dev/null -w '%{http_code}' -X OPTIONS http://api:8080/api/signup \
  -H 'Origin: https://example.com' -H 'Access-Control-Request-Method: POST')"
assert_eq "$pre_code" "204" "CORS preflight answered with 204"

# ---------------------------------------------------------------------------
# 10) frontend serves the app with the CSP header
# ---------------------------------------------------------------------------
front="$(ccurl -D - http://frontend:8080/)"
assert_icontains "$front" 'content-security-policy' "frontend sends a CSP header"
assert_icontains "$front" '<html' "frontend serves the HTML app"

# ---------------------------------------------------------------------------
# verdict
# ---------------------------------------------------------------------------
if [ "$FAILED" -ne 0 ]; then
  die "container smoke test had failures (see FAIL lines above)"
fi
ok "container smoke test passed."
