#!/usr/bin/env bash
# scripts/tests/lib-detect-test.sh — hermetic unit tests for detect_runtime()
# and ensure_podman_socket() in scripts/lib.sh.
#
# Strategy: for each case, build a stub bin directory containing fake
# `podman` / `podman-compose` / `docker` executables that record every
# invocation and answer exactly the probes lib.sh makes, then run
# detect_runtime inside `env -i` whose PATH holds only the stubs plus a
# symlink farm of the real coreutils lib.sh needs. No containers, no
# sockets, no network, no root — this runs in a second on any box with bash.
#
# What it pins (the CI failure this guards against): `podman compose` is only
# a dispatcher, and on GitHub's ubuntu runners it delegates to docker's
# compose plugin, which needs a podman API socket that is not running. So:
#   1. podman-compose on PATH must WIN over `podman compose`.
#   2. `podman compose` delegating to docker-compose must start the socket
#      and set COMPOSE_BAKE=false.
#   3. `podman compose` delegating to anything else must NOT touch sockets.
#   4. docker-only hosts keep using `docker compose`.
#   5. Preset CR/COMPOSE must bypass detection entirely.
#
# Usage: bash scripts/tests/lib-detect-test.sh

set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB="$SCRIPT_DIR/../lib.sh"
BASH_BIN="$(command -v bash)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

FAILED=0
pass() { printf 'PASS: %s\n' "$*"; }
fail() { printf 'FAIL: %s\n' "$*" >&2; FAILED=1; }

# ---------------------------------------------------------------------------
# hermetic PATH: symlink only the external tools lib.sh (and the stubs, which
# use nothing but bash builtins) actually need. Builtins resolve inside bash
# and need no entry; `command -v` prints a bare name for those, which the /*
# guard skips.
# ---------------------------------------------------------------------------
REALBIN="$TMP/realbin"
mkdir -p "$REALBIN"
for tool in bash dirname sed sleep seq mkdir rm cat grep head cut tr; do
  p="$(command -v "$tool" || true)"
  case "$p" in
    /*) ln -sf "$p" "$REALBIN/$tool" ;;
  esac
done

# The harness every case runs: source lib.sh under the same strict shell
# options the real scripts use, detect, and print the outcome as KEY=VALUE.
HARNESS="$TMP/harness.sh"
cat > "$HARNESS" <<'EOF'
set -euo pipefail
IFS=$'\n\t'
. "$1"
detect_runtime
printf 'CR=%s\n' "$CR"
printf 'COMPOSE=%s\n' "$COMPOSE"
printf 'COMPOSE_BAKE=%s\n' "${COMPOSE_BAKE:-}"
EOF

# run_detect STUBDIR [VAR=VAL ...] — run the harness hermetically; stdout is
# the KEY=VALUE outcome, lib.sh's own logging stays on stderr.
run_detect() {
  local stubdir="$1"
  shift
  env -i HOME="$TMP" TERM=dumb PATH="$stubdir:$REALBIN" "$@" \
    "$BASH_BIN" "$HARNESS" "$LIB"
}

# make_stub DIR NAME — install an executable stub from stdin.
make_stub() {
  mkdir -p "$1"
  cat > "$1/$2"
  chmod +x "$1/$2"
}

expect_line() { # OUTPUT LINE LABEL — grep must drain stdin (no -q + pipefail)
  if printf '%s\n' "$1" | grep -Fx -- "$2" >/dev/null; then
    pass "$3"
  else
    fail "$3 (wanted line '$2'; got: $(printf '%s' "$1" | tr '\n' '|'))"
  fi
}

calls_contain() { # CALLS_FILE NEEDLE LABEL
  if grep -F -- "$2" "$1" >/dev/null 2>&1; then
    pass "$3"
  else
    fail "$3 (no '$2' in $(tr '\n' '|' < "$1" 2>/dev/null))"
  fi
}

calls_lack() { # CALLS_FILE NEEDLE LABEL
  if grep -F -- "$2" "$1" >/dev/null 2>&1; then
    fail "$3 (found unwanted '$2' in $(tr '\n' '|' < "$1" 2>/dev/null))"
  else
    pass "$3"
  fi
}

# A recording podman stub that answers every probe lib.sh makes. Behaviour is
# steered by env: CALLS (log file), STATE (dir; socket-up marker = socket
# exists), PROVIDER (path echoed in the dispatcher banner). `system service`
# creates the marker, simulating a socket that comes up after being started.
# shellcheck disable=SC2016  # single quotes are the point: this is a template
PODMAN_STUB='#!/usr/bin/env bash
printf "podman %s\n" "$*" >> "$CALLS"
if [ "${1:-}" = "compose" ] && [ "${2:-}" = "version" ]; then
  if [ -n "${PROVIDER:-}" ]; then
    printf ">>>> Executing external compose provider \"%s\". Please refer to the documentation for details. <<<<\n" "$PROVIDER" >&2
  fi
  exit 0
fi
if [ "${1:-}" = "info" ]; then
  case "$*" in
    *RemoteSocket.Exists*)
      if [ -e "$STATE/socket-up" ]; then echo true; else echo false; fi
      exit 0 ;;
    *RemoteSocket.Path*)
      echo "unix://$STATE/podman.sock"
      exit 0 ;;
  esac
  exit 0
fi
if [ "${1:-}" = "system" ] && [ "${2:-}" = "service" ]; then
  : > "$STATE/socket-up"
  exit 0
fi
exit 0
'

# ---------------------------------------------------------------------------
# case 1: podman + podman-compose both present -> podman-compose wins, and
# nothing ever goes near the socket machinery.
# ---------------------------------------------------------------------------
c1="$TMP/c1"; mkdir -p "$c1/state"; : > "$c1/calls.log"
printf '%s' "$PODMAN_STUB" | make_stub "$c1/bin" podman
make_stub "$c1/bin" podman-compose <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
if out="$(run_detect "$c1/bin" CALLS="$c1/calls.log" STATE="$c1/state" PROVIDER=/usr/libexec/docker/cli-plugins/docker-compose)"; then
  expect_line "$out" "CR=podman" "case1: CR is podman"
  expect_line "$out" "COMPOSE=podman-compose" "case1: podman-compose preferred over the dispatcher"
  expect_line "$out" "COMPOSE_BAKE=" "case1: COMPOSE_BAKE untouched"
  calls_lack "$c1/calls.log" "system service" "case1: no socket started"
else
  fail "case1: detect_runtime exited non-zero"
fi

# ---------------------------------------------------------------------------
# case 2: podman only, dispatcher delegates to docker-compose, socket down,
# no systemctl -> falls back to `podman system service`, exports
# COMPOSE_BAKE=false, and succeeds once the socket is up.
# ---------------------------------------------------------------------------
c2="$TMP/c2"; mkdir -p "$c2/state"; : > "$c2/calls.log"
printf '%s' "$PODMAN_STUB" | make_stub "$c2/bin" podman
if out="$(run_detect "$c2/bin" CALLS="$c2/calls.log" STATE="$c2/state" PROVIDER=/usr/libexec/docker/cli-plugins/docker-compose)"; then
  expect_line "$out" "CR=podman" "case2: CR is podman"
  expect_line "$out" "COMPOSE=podman compose" "case2: falls back to the dispatcher"
  expect_line "$out" "COMPOSE_BAKE=false" "case2: COMPOSE_BAKE=false for the docker-compose provider"
  calls_contain "$c2/calls.log" "system service" "case2: socket self-heal started 'podman system service'"
  if [ -e "$c2/state/socket-up" ]; then
    pass "case2: socket came up"
  else
    fail "case2: socket marker missing"
  fi
else
  fail "case2: detect_runtime exited non-zero"
fi

# ---------------------------------------------------------------------------
# case 3: podman only, dispatcher delegates to a NON-docker provider ->
# no socket calls, COMPOSE_BAKE untouched.
# ---------------------------------------------------------------------------
c3="$TMP/c3"; mkdir -p "$c3/state"; : > "$c3/calls.log"
printf '%s' "$PODMAN_STUB" | make_stub "$c3/bin" podman
if out="$(run_detect "$c3/bin" CALLS="$c3/calls.log" STATE="$c3/state" PROVIDER=/usr/libexec/podman/podman-compose-provider)"; then
  expect_line "$out" "COMPOSE=podman compose" "case3: uses the dispatcher"
  expect_line "$out" "COMPOSE_BAKE=" "case3: COMPOSE_BAKE untouched for non-docker provider"
  calls_lack "$c3/calls.log" "system service" "case3: no socket started"
else
  fail "case3: detect_runtime exited non-zero"
fi

# ---------------------------------------------------------------------------
# case 4: docker only -> docker compose.
# ---------------------------------------------------------------------------
c4="$TMP/c4"
make_stub "$c4/bin" docker <<'EOF'
#!/usr/bin/env bash
if [ "${1:-}" = "compose" ] && [ "${2:-}" = "version" ]; then exit 0; fi
exit 0
EOF
if out="$(run_detect "$c4/bin")"; then
  expect_line "$out" "CR=docker" "case4: CR is docker"
  expect_line "$out" "COMPOSE=docker compose" "case4: docker compose selected"
else
  fail "case4: detect_runtime exited non-zero"
fi

# ---------------------------------------------------------------------------
# case 5: preset CR/COMPOSE bypasses detection (no runtimes on PATH at all).
# ---------------------------------------------------------------------------
c5="$TMP/c5"; mkdir -p "$c5/bin"
if out="$(run_detect "$c5/bin" CR=docker COMPOSE="docker compose")"; then
  expect_line "$out" "CR=docker" "case5: preset CR respected"
  expect_line "$out" "COMPOSE=docker compose" "case5: preset COMPOSE respected"
else
  fail "case5: detect_runtime exited non-zero with preset CR/COMPOSE"
fi

# ---------------------------------------------------------------------------
# verdict
# ---------------------------------------------------------------------------
if [ "$FAILED" -ne 0 ]; then
  printf 'lib-detect-test: FAILURES above\n' >&2
  exit 1
fi
printf 'lib-detect-test: all cases passed.\n'
