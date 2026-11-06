#!/usr/bin/env bash
# scripts/tests/compose-defaults-test.sh — hermetic unit tests for
# seed_compose_defaults() in scripts/lib.sh.
#
# What it pins: podman-compose before 1.6 turns an environment entry such as
#   GOTUNNELS_SIGNUP_IP_INTERVAL: "${GOTUNNELS_SIGNUP_IP_INTERVAL:-5m}"
# into the literal "${GOTUNNELS_SIGNUP_IP_INTERVAL:-5m}" when the variable is
# unset on the host, which crash-looped the API. dc() therefore seeds every
# compose default into the environment first. These tests check that:
#   1. every ${NAME:-default} in compose.yaml ends up set after seeding,
#      with exactly the default compose.yaml declares;
#   2. variables already set (including set-but-empty) are never overridden;
#   3. defaults containing spaces, quotes, '*' and ':' survive intact;
#   4. references inside YAML comments are ignored.
#
# Usage: bash scripts/tests/compose-defaults-test.sh

set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB="$SCRIPT_DIR/../lib.sh"
COMPOSE_FILE="$SCRIPT_DIR/../../compose.yaml"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

FAILED=0
pass() { printf 'PASS: %s\n' "$*"; }
fail() { printf 'FAIL: %s\n' "$*" >&2; FAILED=1; }

run_seed() { # file, then NAME=VALUE presets; prints the resulting env
  local file="$1"
  shift
  # shellcheck disable=SC2016
  env -i PATH="$PATH" "$@" bash -c '
    set -euo pipefail
    . "$1" >/dev/null 2>&1
    seed_compose_defaults "$2"
    env
  ' _ "$LIB" "$file"
}

value_of() { # name, env dump
  printf '%s\n' "$2" | sed -n "s/^$1=//p"
}

# --- 1. every default in the real compose.yaml gets seeded -----------------
out="$(run_seed "$COMPOSE_FILE")"
count=0
while IFS= read -r ref; do
  name="$(printf '%s' "$ref" | sed -E 's/^\$\{([A-Za-z_][A-Za-z0-9_]*):?-.*$/\1/')"
  def="$(printf '%s' "$ref" | sed -E 's/^\$\{[A-Za-z_][A-Za-z0-9_]*:?-(.*)\}$/\1/')"
  if ! printf '%s\n' "$out" | grep -q "^${name}="; then
    fail "compose.yaml: $name was not seeded"
  elif [ "$(value_of "$name" "$out")" != "$def" ]; then
    fail "compose.yaml: $name seeded as '$(value_of "$name" "$out")', want '$def'"
  fi
  count=$((count + 1))
done < <(grep -v '^[[:space:]]*#' "$COMPOSE_FILE" | grep -oE '\$\{[A-Za-z_][A-Za-z0-9_]*:?-[^}]*\}' | sort -u -t: -k1,1)
if [ "$count" -lt 10 ]; then
  fail "expected at least 10 defaulted references in compose.yaml, found $count"
else
  pass "all $count defaulted compose.yaml references are seeded with their defaults"
fi
for name in GOTUNNELS_SIGNUP_IP_INTERVAL GOTUNNELS_AUTH_ATTEMPT_BURST GOTUNNELS_MAINTENANCE_INTERVAL GOTUNNELS_RP_ORIGIN_PATTERNS GOTUNNELS_API_UPSTREAM; do
  if printf '%s\n' "$out" | grep -q "^${name}=" && [ "$(value_of "$name" "$out")" != "" ]; then
    pass "$name seeded ($(value_of "$name" "$out"))"
  else
    fail "$name not seeded"
  fi
done

# --- 2..4. synthetic fixture ------------------------------------------------
cat > "$TMP/compose.yaml" <<'YAML'
services:
  api:
    environment:
      # COMMENTED: "${COMMENTED_VAR:-nope}"
      PLAIN: "${PLAIN:-5m}"
      PRESET: "${PRESET:-default}"
      EMPTYSET: "${EMPTYSET:-default}"
      SPACED: "${SPACED:-default-src 'self'; connect-src 'self' https:}"
      STAR: "${STAR:-https://*.trycloudflare.com}"
      DASHONLY: "${DASHONLY-dash}"
      TWO: "${TWO_A:-a}:${TWO_B:-b}"
      EMPTYDEF: "${EMPTYDEF:-}"
      NODEF: "${NODEF}"
YAML
out="$(run_seed "$TMP/compose.yaml" PRESET=mine EMPTYSET=)"

check() { # name want description
  if [ "$(value_of "$1" "$out")" = "$2" ]; then pass "$3"; else fail "$3 (got '$(value_of "$1" "$out")', want '$2')"; fi
}
check PLAIN 5m "unset variable gets its default"
check PRESET mine "set variable is not overridden"
if printf '%s\n' "$out" | grep -q '^EMPTYSET=$'; then pass "set-but-empty variable is not overridden"; else fail "set-but-empty variable was overridden"; fi
check SPACED "default-src 'self'; connect-src 'self' https:" "default with spaces and quotes survives"
check STAR "https://*.trycloudflare.com" "default with '*' and ':' survives"
check DASHONLY dash "\${NAME-default} form is seeded"
check TWO_A a "first of two references on one line is seeded"
check TWO_B b "second of two references on one line is seeded"
if printf '%s\n' "$out" | grep -q '^EMPTYDEF=$'; then pass "empty default is seeded as empty"; else fail "empty default not seeded"; fi
if printf '%s\n' "$out" | grep -q '^COMMENTED_VAR='; then fail "reference inside a comment was seeded"; else pass "references in comments are ignored"; fi
if printf '%s\n' "$out" | grep -q '^NODEF='; then fail "reference without a default was seeded"; else pass "references without a default are left alone"; fi

if [ "$FAILED" -ne 0 ]; then
  printf '\nsome compose-defaults tests FAILED\n' >&2
  exit 1
fi
printf '\nall compose-defaults tests passed\n'
