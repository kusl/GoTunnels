#!/usr/bin/env bash
# scripts/lib.sh — shared helpers sourced by the other scripts and by CI.
# Keeping logic here (not in the GitHub Actions YAML) is deliberate: the same
# code runs on a laptop and in CI.

# ---------------------------------------------------------------------------
# logging
# ---------------------------------------------------------------------------
if [ -t 2 ]; then
  _c_reset=$'\033[0m'; _c_dim=$'\033[2m'; _c_grn=$'\033[32m'; _c_ylw=$'\033[33m'; _c_red=$'\033[31m'; _c_cyn=$'\033[36m'
else
  _c_reset=; _c_dim=; _c_grn=; _c_ylw=; _c_red=; _c_cyn=
fi
log()  { printf '%s[gotunnels]%s %s\n' "$_c_cyn" "$_c_reset" "$*" >&2; }
ok()   { printf '%s[gotunnels]%s %s%s%s\n' "$_c_cyn" "$_c_reset" "$_c_grn" "$*" "$_c_reset" >&2; }
warn() { printf '%s[gotunnels]%s %s%s%s\n' "$_c_cyn" "$_c_reset" "$_c_ylw" "$*" "$_c_reset" >&2; }
err()  { printf '%s[gotunnels]%s %s%s%s\n' "$_c_cyn" "$_c_reset" "$_c_red" "$*" "$_c_reset" >&2; }
die()  { err "$*"; exit 1; }

# Repo root = parent of this scripts/ directory.
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$LIB_DIR/.." && pwd)"

# ---------------------------------------------------------------------------
# container runtime + compose detection
# ---------------------------------------------------------------------------
# Sets:
#   CR      -> container runtime binary (podman|docker) for logs/cp/inspect
#   COMPOSE -> the compose invocation (may be two words, used unquoted)
#
# Preset CR and COMPOSE in the environment to bypass detection entirely.
#
# podman branch ordering is deliberate: prefer the podman-compose BINARY over
# the `podman compose` dispatcher. `podman compose` is only a shim — it
# executes whichever external provider it finds first, and on GitHub's ubuntu
# runners that is docker's compose plugin
# (/usr/libexec/docker/cli-plugins/docker-compose), which speaks the Docker
# API to a podman socket that is not running by default in a rootless
# session. That is exactly the CI failure "Cannot connect to the Docker
# daemon at unix:///run/user/1001/podman/podman.sock" — the workflow's
# pipx-installed podman-compose was sitting unused because the old detection
# probed `podman compose` first. podman-compose drives the podman CLI
# directly and needs no socket. When only the dispatcher exists we still use
# it, but if its provider turns out to be docker-compose we make that path
# workable: start the podman API socket and disable compose Bake (a
# buildkit-only build path podman's compat socket cannot serve).
detect_runtime() {
  if [ -n "${CR:-}" ] && [ -n "${COMPOSE:-}" ]; then return 0; fi
  if command -v podman >/dev/null 2>&1; then
    CR=podman
    if command -v podman-compose >/dev/null 2>&1; then
      COMPOSE="podman-compose"
    elif podman compose version >/dev/null 2>&1; then
      COMPOSE="podman compose"
      local provider
      provider="$(podman_compose_provider)"
      if [ -n "$provider" ]; then
        log "podman compose delegates to: $provider"
      fi
      case "$provider" in
        *docker-compose*)
          export COMPOSE_BAKE="${COMPOSE_BAKE:-false}"
          ensure_podman_socket || die "podman compose delegates to docker-compose, which needs the podman API socket, and the socket could not be started. Install podman-compose (e.g. 'pipx install podman-compose') or start the socket yourself ('systemctl --user start podman.socket')."
          ;;
      esac
    else
      die "podman found but neither 'podman-compose' nor 'podman compose' is available"
    fi
  elif command -v docker >/dev/null 2>&1; then
    CR=docker
    if docker compose version >/dev/null 2>&1; then
      COMPOSE="docker compose"
    else
      die "docker found but 'docker compose' is not available"
    fi
  else
    die "no container runtime found (need podman or docker)"
  fi
  export CR COMPOSE
  log "using runtime: $CR / compose: $COMPOSE"
}

# podman_compose_provider — the path of the external provider that
# `podman compose` would execute, parsed from its
#   >>>> Executing external compose provider "<path>". ... <<<<
# stderr banner. Empty when the banner is absent (older podman). Probe only —
# never fails, so it is safe in `$(...)` under set -e/pipefail.
podman_compose_provider() {
  podman compose version 2>&1 \
    | sed -n 's/.*Executing external compose provider "\([^"]*\)".*/\1/p' \
    | sed -n '1p' || true
}

# ensure_podman_socket — make sure the podman API socket is accepting
# connections; start it when it is not. Only needed when a Docker-API client
# (docker's compose plugin) fronts podman. Prefers the systemd user unit;
# falls back to a backgrounded `podman system service` with a 10-minute idle
# timeout so it reaps itself — no EXIT trap needed here, which matters
# because ci-container-test.sh already owns the (single) EXIT trap.
ensure_podman_socket() {
  local exists sock
  exists="$(podman info --format '{{.Host.RemoteSocket.Exists}}' 2>/dev/null || true)"
  if [ "$exists" = "true" ]; then return 0; fi

  # Path may or may not carry a unix:// scheme depending on podman version.
  sock="$(podman info --format '{{.Host.RemoteSocket.Path}}' 2>/dev/null || true)"
  sock="${sock#unix://}"
  log "podman API socket not running — starting it for the docker-compose provider"

  if command -v systemctl >/dev/null 2>&1 \
     && systemctl --user start podman.socket >/dev/null 2>&1; then
    log "started podman.socket (systemd user unit)"
  elif [ -n "$sock" ]; then
    mkdir -p "$(dirname "$sock")" 2>/dev/null || true
    (podman system service --time=600 "unix://$sock" >/dev/null 2>&1 &)
    log "started 'podman system service' on unix://$sock (10 min idle timeout)"
  else
    (podman system service --time=600 >/dev/null 2>&1 &)
    log "started 'podman system service' on the default socket (10 min idle timeout)"
  fi

  for _ in $(seq 1 40); do
    exists="$(podman info --format '{{.Host.RemoteSocket.Exists}}' 2>/dev/null || true)"
    if [ "$exists" = "true" ]; then return 0; fi
    sleep 0.25
  done
  return 1
}

# dc: run a compose command against the repo compose file. COMPOSE may be two
# words ("podman compose" / "docker compose"); it is expanded unquoted so it
# splits into separate arguments. These scripts run under a strict
# IFS=$'\n\t' (no space), which would otherwise keep "podman compose" as one
# nonexistent command — so restore a normal IFS locally just for the split.
dc() {
  local IFS=$' \t\n'
  # shellcheck disable=SC2086
  $COMPOSE -f "$REPO_ROOT/compose.yaml" "$@"
}

# ---------------------------------------------------------------------------
# secrets + env
# ---------------------------------------------------------------------------
gen_secret() {
  local bytes="${1:-32}"
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -base64 "$bytes" | tr -d '\n' | tr '+/' '-_' | tr -d '='
  else
    head -c "$bytes" /dev/urandom | base64 | tr -d '\n' | tr '+/' '-_' | tr -d '='
  fi
}

ENV_FILE="$REPO_ROOT/.env"

# set_env_var KEY VALUE — idempotently upsert KEY=VALUE in .env.
set_env_var() {
  local key="$1" val="$2"
  touch "$ENV_FILE"
  if grep -q "^${key}=" "$ENV_FILE" 2>/dev/null; then
    # Replace in place; use a temp file to stay portable across sed variants.
    grep -v "^${key}=" "$ENV_FILE" > "$ENV_FILE.tmp"
    mv "$ENV_FILE.tmp" "$ENV_FILE"
  fi
  printf '%s=%s\n' "$key" "$val" >> "$ENV_FILE"
}

# ensure_env — create .env with fresh secrets if absent, then load it.
ensure_env() {
  if [ ! -f "$ENV_FILE" ]; then
    log "generating fresh .env with per-instance secrets"
    cat > "$ENV_FILE" <<EOF
GOTUNNELS_INSTANCE_ID=${GOTUNNELS_INSTANCE_ID:-default}
GOTUNNELS_VERSION=${GOTUNNELS_VERSION:-dev}
POSTGRES_USER=gotunnels
POSTGRES_DB=gotunnels
POSTGRES_PASSWORD=$(gen_secret 24)
GOTUNNELS_IP_HASH_PEPPER=$(gen_secret 32)
GOTUNNELS_TOTP_ENCRYPTION_KEY=$(gen_secret 32)
GOTUNNELS_RP_ID=localhost
GOTUNNELS_RP_DISPLAY_NAME=GoTunnels
GOTUNNELS_RP_ORIGINS=http://localhost:8080
GOTUNNELS_CORS_ALLOWED_ORIGINS=*
GOTUNNELS_CSP_HEADER_NAME=Content-Security-Policy-Report-Only
GOTUNNELS_CSP_MODE=report-only
GOTUNNELS_CSP_POLICY="default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self' https:; media-src 'self'; object-src 'none'; frame-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
UPTRACE_DSN=${UPTRACE_DSN:-}
OTEL_EXPORTER_OTLP_ENDPOINT=${OTEL_EXPORTER_OTLP_ENDPOINT:-}
OTEL_EXPORTER_OTLP_HEADERS=${OTEL_EXPORTER_OTLP_HEADERS:-}
OTEL_EXPORTER_OTLP_COMPRESSION=${OTEL_EXPORTER_OTLP_COMPRESSION:-gzip}
OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=${OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE:-delta}
OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION=${OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION:-base2_exponential_bucket_histogram}
OTEL_SERVICE_NAME=gotunnels-api
GOTUNNELS_DEV=false
EOF
    ok "wrote $ENV_FILE (gitignored)"
  else
    log "using existing $ENV_FILE"
  fi

  # Persist any telemetry settings the CALLER exported in their shell into
  # .env BEFORE load_env runs. Order matters: an existing .env with an empty
  # `UPTRACE_DSN=` line would otherwise clobber `export UPTRACE_DSN=...` from
  # the invoking shell (load_env re-exports the empty value over it) and
  # telemetry silently turns off. That is exactly the trap in
  # `export UPTRACE_DSN=… ; bash scripts/up.sh` against an older .env — the
  # only visible symptom is nothing ever arriving at the backend. Persisting
  # here also means the DSN survives into future runs without re-exporting.
  local _k _v
  for _k in UPTRACE_DSN \
            OTEL_EXPORTER_OTLP_ENDPOINT \
            OTEL_EXPORTER_OTLP_HEADERS \
            OTEL_EXPORTER_OTLP_COMPRESSION \
            OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE \
            OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION; do
    _v="${!_k:-}"
    if [ -n "$_v" ]; then
      set_env_var "$_k" "$_v"
    fi
  done

  load_env
}

# load_env — export every non-comment KEY=VALUE from .env into the environment
# so compose ${VAR} substitution and the scripts both see them.
load_env() {
  [ -f "$ENV_FILE" ] || return 0
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
}

# reset_runtime_env — return the tunnel-derived keys to their bootstrap
# defaults at the start of every run. Quick Tunnel hostnames are ephemeral, so
# whatever up.sh wrote into .env on the LAST run is guaranteed stale on this
# one. Left in place, those stale values become the API's CORS allow-list and
# passkey RP if anything starts the api container before step 6 re-derives
# them — which is precisely the CORS-on-signup failure that deleting .env
# "fixed" (a fresh .env happens to default to `*`). Resetting the three keys
# to the same permissive bootstrap values a fresh .env would carry makes
# deleting .env unnecessary and, unlike deletion, preserves the generated
# secrets (Postgres password, pepper, TOTP key) and any persisted DSN.
reset_runtime_env() {
  set_env_var GOTUNNELS_RP_ID localhost
  set_env_var GOTUNNELS_RP_ORIGINS http://localhost:8080
  set_env_var GOTUNNELS_CORS_ALLOWED_ORIGINS '*'
  # load_env has usually already run by the time this is called; re-export so
  # the current shell (and compose var substitution) sees the fresh values,
  # not the stale ones .env held a moment ago.
  export GOTUNNELS_RP_ID=localhost
  export GOTUNNELS_RP_ORIGINS=http://localhost:8080
  export GOTUNNELS_CORS_ALLOWED_ORIGINS='*'
  log "reset tunnel-derived env (RP_ID / RP_ORIGINS / CORS) to bootstrap defaults"
}

# ---------------------------------------------------------------------------
# project name (instance isolation)
# ---------------------------------------------------------------------------
# Resolves a compose project name. Precedence: $1 arg > $GOTUNNELS_PROJECT >
# $GOTUNNELS_INSTANCE_ID (if not 'default') > generated random.
resolve_project() {
  local p="${1:-}"
  if [ -z "$p" ]; then p="${GOTUNNELS_PROJECT:-}"; fi
  if [ -z "$p" ] && [ -n "${GOTUNNELS_INSTANCE_ID:-}" ] && [ "${GOTUNNELS_INSTANCE_ID}" != "default" ]; then
    p="gotunnels-${GOTUNNELS_INSTANCE_ID}"
  fi
  if [ -z "$p" ]; then p="gotunnels-$(gen_secret 4 | tr '[:upper:]' '[:lower:]' | tr -cd 'a-z0-9' | cut -c1-6)"; fi
  # compose project names must be lowercase alnum/dash/underscore.
  p="$(printf '%s' "$p" | tr '[:upper:]' '[:lower:]' | tr -cd 'a-z0-9_-')"
  echo "$p"
}

# ---------------------------------------------------------------------------
# health + tunnel discovery
# ---------------------------------------------------------------------------
# cid_of project service — the container id for <service> in <project>, or ""
# if there is none.
#
# We resolve the id straight from the container runtime ($CR) by label, NOT via
# `<compose> ps -q <service>`, because that compose path is broken here: on
# Fedora `podman compose` shells out to the external `podman-compose` provider
# (that's the ">>>> Executing external compose provider" banner), and
# podman-compose's `ps` subcommand is not docker-compatible — it accepts NO
# service argument and filters only by the project label. So
# `<compose> ps -q db` either errors on the stray `db` token (non-zero, empty
# stdout) or lists *every* container in the project; it never returns db's id
# specifically. That is the "no container id resolved for 'db'" we hit even
# though the container had been created.
#
# Every compose implementation we support — docker compose, podman compose, and
# podman-compose — stamps each container with the docker-compat labels
# `com.docker.compose.project` and `com.docker.compose.service`, so filtering on
# both with the runtime's own `ps` returns exactly the one container we mean.
#
# BEST-EFFORT / MUST NEVER FAIL. Callers run under `set -euo pipefail` and
# capture this with a bare assignment (`cid="$(cid_of …)"`), whose exit status
# is that of this pipeline. `sed -n '1p'` prints the first id while still
# draining the rest of the stream (no early pipe close → no SIGPIPE back to
# `ps`), and the trailing `|| true` swallows any non-zero exit so errexit can't
# abort the caller. `-a` is intentional: it also matches a crashed/exited
# container so the error paths below can still read its logs.
cid_of() { # project service
  "$CR" ps -aq \
    --filter "label=com.docker.compose.project=$1" \
    --filter "label=com.docker.compose.service=$2" \
    2>/dev/null | sed -n '1p' || true
}

# health_status cid — the container's healthcheck status, or "" when it has
# none. The `{{if .State.Health}}` guard stops Go's template engine from
# printing the literal "<no value>" for a container without a healthcheck, so
# "no healthcheck" and "not reported yet" both read as the empty string.
health_status() { # cid
  [ -n "${1:-}" ] || return 0
  $CR inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{end}}' "$1" 2>/dev/null || true
}

# wait_healthy project service timeout [probe-cmd...] — block until <service> is
# actually ready, up to <timeout> seconds.
#
# Why this is more than a poll of `.State.Health.Status`:
#   podman-compose *does* translate compose's `healthcheck:` into podman
#   `--health-cmd` / `--health-interval` flags, so the container is created WITH
#   a healthcheck. But podman drives the *periodic* re-check from a per-container
#   systemd timer, and in a rootless / plain-shell session that timer frequently
#   never fires — so `.State.Health.Status` sits at "starting" forever even
#   though Postgres accepted connections within a second or two. Passively
#   waiting on that field therefore always times out here (this is exactly the
#   "'db' did not become healthy in 120s" we kept hitting).
#
# So each second we accept readiness from whichever of these fires first:
#   1. passive status == "healthy"      — docker (its daemon runs the checks),
#                                          or podman if the timer *is* firing.
#   2. `podman healthcheck run` exit 0   — runs the container's OWN healthcheck
#                                          command once, on demand, with no timer
#                                          involved (podman only). Bonus: it also
#                                          updates the recorded status, so
#                                          `podman ps` shows "healthy" afterward.
#   3. the caller's probe, exit 0        — run as `$CR exec <cid> <probe...>`;
#                                          the ultimate fallback that depends on
#                                          nothing but the container running. For
#                                          db we pass `pg_isready …`.
# If no probe is given, only 1 and 2 are used.
wait_healthy() { # project service timeout_seconds [probe-cmd...]
  local project="$1" svc="$2" timeout="${3:-90}"
  shift 3 2>/dev/null || true
  local cid status st hs
  log "waiting for '$svc' to become ready (up to ${timeout}s)"
  for _ in $(seq 1 "$timeout"); do
    cid="$(cid_of "$project" "$svc")"
    if [ -n "$cid" ]; then
      status="$(health_status "$cid")"
      if [ "$status" = "healthy" ]; then ok "'$svc' is healthy"; return 0; fi
      if [ "$CR" = "podman" ] && "$CR" healthcheck run "$cid" >/dev/null 2>&1; then
        ok "'$svc' is healthy"; return 0
      fi
      if [ "$#" -gt 0 ] && "$CR" exec "$cid" "$@" >/dev/null 2>&1; then
        ok "'$svc' is ready"; return 0
      fi
    fi
    sleep 1
  done
  # Don't fail with a bare timeout — surface the container's real state and its
  # recent logs so a genuine Postgres problem (bad config, crash loop, wrong
  # password) is visible instead of being hidden behind "did not become ready".
  err "'$svc' did not become ready in ${timeout}s"
  if [ -n "${cid:-}" ]; then
    st="$($CR inspect --format '{{.State.Status}}' "$cid" 2>/dev/null || echo '?')"
    hs="$(health_status "$cid")"; [ -n "$hs" ] || hs='(none)'
    err "  container: state=$st health=$hs id=$(printf '%.12s' "$cid")"
    err "  recent '$svc' logs ($CR logs --tail 40 $svc):"
    "$CR" logs --tail 40 "$cid" 2>&1 | sed 's/^/    /' >&2 || true
  else
    err "  no container id resolved for '$svc' — was it created? (check: dc -p $project ps)"
  fi
  return 1
}

poll_tunnel_url() { # project service timeout_seconds
  local project="$1" svc="$2" timeout="${3:-60}" cid url
  for _ in $(seq 1 "$timeout"); do
    cid="$(cid_of "$project" "$svc")"
    if [ -n "$cid" ]; then
      # `|| true` is required: before the URL is logged, `grep` matches nothing
      # and exits 1; once it matches, `head -n1` SIGPIPE-s `grep`. Under
      # `pipefail` either would make this bare assignment abort the caller via
      # errexit — so neutralize the pipeline's exit status here.
      url="$($CR logs "$cid" 2>&1 | grep -Eo 'https://[a-z0-9._-]+\.trycloudflare\.com' | head -n1 || true)"
      if [ -n "$url" ]; then echo "$url"; return 0; fi
    fi
    sleep 1
  done
  return 1
}

# wait_for_log project service pattern timeout — poll container logs for a regex.
#
# The grep must DRAIN the whole log stream: `grep -E … >/dev/null`, NOT
# `grep -Eq`. With -q grep exits at the first match, the runtime's `logs`
# process takes SIGPIPE (exit 141), and under `pipefail` the pipeline — and so
# this `if` condition — reports failure even though the pattern WAS present.
# That inverted every success into a miss, which is why up.sh printed "did not
# observe API listening log yet" on every run while the API was in fact
# serving. Same failure family as cid_of / poll_tunnel_url above; keeping grep
# reading to EOF lets `logs` exit 0. A genuine no-match still exits 1 and the
# loop just polls again.
wait_for_log() {
  local project="$1" svc="$2" pat="$3" timeout="${4:-60}" cid
  for _ in $(seq 1 "$timeout"); do
    cid="$(cid_of "$project" "$svc")"
    if [ -n "$cid" ] && "$CR" logs "$cid" 2>&1 | grep -E -- "$pat" >/dev/null; then
      return 0
    fi
    sleep 1
  done
  return 1
}

host_of_url() { # https://x.y.z/... -> x.y.z
  printf '%s' "$1" | sed -E 's#^[a-z]+://##; s#/.*$##'
}

write_frontend_config() { # project api_url
  local project="$1" api_url="$2" cid tmp
  cid="$(cid_of "$project" frontend)"
  [ -n "$cid" ] || { err "frontend container not found; cannot write config.json"; return 1; }
  tmp="$(mktemp)"
  cat > "$tmp" <<EOF
{
  "apiBase": "${api_url}",
  "instanceId": "${GOTUNNELS_INSTANCE_ID:-default}",
  "generatedAt": "$(date --iso-8601=seconds 2>/dev/null || date)"
}
EOF
  $CR cp "$tmp" "${cid}:/srv/config.json"
  rm -f "$tmp"
  ok "wrote config.json into frontend container"
}
