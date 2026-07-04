# GoTunnels

A fully self-contained demo stack that stands up a **Go API** and a **plain
HTML/CSS/JS web app**, each reachable only through its own **Cloudflare Quick
Tunnel**, with **PostgreSQL** and **OpenTelemetry** behind them. No service
binds to a host port — the only traffic leaving the machine is the two outbound
tunnel connections. Every secret is generated on first run, so you can stand up
ten independent instances side by side without configuring anything.

> ### Built with LLM assistance
> This project — its code, migrations, container and Compose files, shell
> scripts, tests, CI, and this documentation — was generated with substantial
> help from a large language model (Anthropic's Claude), working from a design
> conversation. It is a demonstration project. 
> Read it, run it, learn from it — but review it
> yourself before trusting it with anything that matters. See
> [Security & demo caveats](#security--demo-caveats).

## What it demonstrates

- **Passkeys (WebAuthn)** as the primary login, a **password** fallback
  (argon2id), and optional **TOTP** two-factor (RFC 6238, works with Aegis /
  Google Authenticator), with single-use recovery codes.
- **Server-side sessions** using an opaque Bearer token that works cross-origin
  (the frontend and API are on different tunnel domains).
- **Privacy-preserving audit log**: client IPs are never stored raw — only
  `sha256(pepper || ip)`, and the hash is shown to you plainly on your own
  activity page.
- **A shared plain-text notes feed** (create / delete, never edit) with
  visibility-aware polling for live updates, deferred "new notes" while you
  read, and server-side ownership checks in the SQL itself.
- **A CAPTCHA toy with server-backed stats**: the auto-solver can do hundreds
  of solves a second, so the browser batches deltas and syncs every few
  seconds; a `RANK() OVER` leaderboard and a per-user preferences endpoint
  round it out.
- **OpenTelemetry** traces, metrics, and logs over OTLP/HTTP — vendor-neutral,
  never tied to a specific backend.
- **Content-Security-Policy** in report-only mode by default; violations are
  reported back to the API, stored, and logged to telemetry.
- **No exposed host ports**; everything talks over the internal Compose network
  by service name.

## Architecture

```
                    ┌─────────────────────── your machine ────────────────────────┐
   browser ─┐       │                                                             │
            │ https │   ┌───────────┐        ┌───────────┐       ┌────────────┐   │
            ├─────────▶│ frontend  │        │    api    │─────▶│     db     │   │
            │       │   │  (Caddy)  │        │   (Go)    │       │ (Postgres) │   │
            │       │   └───────────┘        └───────────┘       └────────────┘   │
            │       │        ▲                     ▲                              │
            │ https │        │ tunnel              │ tunnel                       │
            └────────────────┘                     │                              │
                    │   cloudflared-frontend   cloudflared-api                    │
                    │        │                     │                              │
                    └────────┼─────────────────────┼──────────────────────────────┘
                             ▼                     ▼
                       Cloudflare edge        Cloudflare edge   ──▶ (OTLP) Uptrace
```

The browser loads the app from the frontend's tunnel URL, then calls the API's
tunnel URL directly (real cross-origin requests — CORS is load-bearing, not
decorative). The API trusts no caller: it authenticates and authorizes every
request regardless of origin, so it is a genuinely reusable service and not
secretly coupled to this frontend.

For the full design rationale (why two tunnels, why a Bearer token rather than a
cookie, how the frontend learns the API URL at runtime), see
[`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

## Quick start

```bash
./scripts/run.sh
```

That one command:

1. regenerates the LLM context dump (`export.sh`),
2. runs the full test suite (build + vet + unit tests), and
3. builds the images and brings the stack up, printing the two Quick Tunnel
   URLs when ready.

Open the **Web app** URL it prints, create an account, then add a passkey or
enable TOTP from Settings. Tear it down with:

```bash
./scripts/down.sh
```

### Prerequisites

- **Podman** (with `podman compose` or `podman-compose`) or **Docker** (with
  `docker compose`). Podman on Fedora is the primary target.
- Outbound internet access (to pull base images and, on the first API build, to
  resolve Go modules).
- `openssl` (for secret generation) — present on essentially every Linux box.

You do **not** need Go, Node, or Caddy installed on the host — everything builds
and runs in containers.

### Running several isolated instances

The Compose project name namespaces containers, network, and volumes, so just
give each instance a different name:

```bash
./scripts/up.sh alpha
./scripts/up.sh bravo
```

Each gets its own database, its own generated secrets (in its own `.env` — note
that a single repo checkout shares one `.env`; use separate checkouts for fully
independent secret sets), and its own pair of tunnel URLs.

## Telemetry (optional)

Point the stack at any OTLP/HTTP backend with a single DSN. For Uptrace Cloud:

```bash
export UPTRACE_DSN="https://<token>@api.uptrace.dev"
./scripts/run.sh
```

Or use the standard `OTEL_EXPORTER_OTLP_ENDPOINT` / `OTEL_EXPORTER_OTLP_HEADERS`
variables. With nothing set, the API still logs structured JSON to stdout
(`podman logs`) and installs no-op trace/metric providers. The exporter is the
vendor-neutral OpenTelemetry SDK — no Uptrace SDK is imported, so switching
backends is a config change, not a code change.

## Configuration

All configuration is centralized. On the Go side, **every** environment
variable is read in exactly one file, [`internal/config/config.go`](internal/config/config.go);
no other package calls `os.Getenv`. The CSP policy is centralized in the
frontend's [`Caddyfile`](frontend/Caddyfile) and mirrored to the API for its
info endpoint. The full list of variables, defaults, and meanings is in
[`docs/CONFIGURATION.md`](docs/CONFIGURATION.md) and
[`.env.example`](.env.example).

The **Content-Security-Policy** ships in report-only mode
(`Content-Security-Policy-Report-Only`) with a strict self-only policy — no
third-party scripts, styles, images, fonts, or frames; everything is
self-hosted. Flip a single variable (`GOTUNNELS_CSP_HEADER_NAME`) to enforce it.
Because the app uses only external scripts/styles (no inline `<script>`, no
inline event handlers, no inline styles), it is already compatible with the
enforcing policy.

## Testing and CI

Tests run through one script, locally and in CI:

```bash
./scripts/test.sh all         # build + vet + unit tests
./scripts/test.sh unit        # just unit tests
./scripts/test.sh vuln        # govulncheck (reachability-aware)
./scripts/test.sh freshness   # bump deps, tidy, retest (drift check)
```

If you have Go installed it uses it directly; otherwise it runs the suite in a
`golang` container. The GitHub Actions workflows are intentionally thin — they
call these script functions rather than embedding logic:

- `ci.yaml` — build, vet, unit tests on every push/PR.
- `govulncheck.yaml` — vulnerability scan on push/PR and weekly.
- `dependency-freshness.yaml` — weekly update-and-retest drift check.

Dependabot watches container base images and Actions versions; the Go module
graph is covered by `govulncheck` (call-graph aware, far less noise than
Dependabot's Go alerts).

## Repository layout

```
.
├── cmd/api/                     Go API entrypoint (wiring + graceful shutdown)
├── internal/
│   ├── config/                  the single source of configuration truth
│   ├── telemetry/               OpenTelemetry (traces/metrics/logs) setup
│   ├── database/                pgx pool + embedded migration runner
│   ├── store/                   all SQL data access
│   ├── auth/                    passwords, passkeys, TOTP, sessions, handlers
│   ├── activity/                audit logging + IP hashing
│   ├── captcha/                 CAPTCHA stats sync, reset, leaderboard
│   ├── notes/                   plain-text microblog (list/create/delete)
│   ├── prefs/                   per-user key/value preferences
│   ├── health/                  liveness / readiness / info
│   ├── csp/                     CSP report ingestion (3 wire formats)
│   ├── httpx/                   CORS, request-id, recovery, rate limit, JSON
│   └── server/                  route wiring + middleware chain
├── migrations/                  *.up.sql / *.down.sql (embedded)
├── frontend/                    plain HTML/CSS/JS + Caddyfile
├── Containerfile.api            multi-stage Go build → distroless
├── Containerfile.frontend       Caddy + static assets
├── compose.yaml                 db, api, frontend, 2× cloudflared
├── scripts/                     lib.sh, run.sh, up.sh, down.sh, test.sh
├── .github/workflows/           ci, govulncheck, dependency-freshness
├── export.sh                    dumps tracked files to docs/llm/dump.txt
├── LICENSE                      AGPL-3.0
└── README.md
```

## A note on `go.sum`

`go.sum` is intentionally **not** committed on first import: the environment
this was authored in could not reach the Go module proxy, so the checksum file
could not be generated there. Everything is set up so this does not matter — the
API container build and every `scripts/test.sh` target run `go mod tidy` first,
which resolves and locks the dependency graph on any machine with network
access. After your first successful run you may commit the generated `go.sum`
for reproducible builds.

## Security & demo caveats

This is a **demonstration**, deliberately simple in places:

- Signup has **no email/SMS verification** — one step by design; not suitable
  for real abuse prevention.
- **Cloudflare Quick Tunnels** are a testing/dev feature with no throughput
  guarantee and a URL that changes on restart. They are perfect for a demo and
  the wrong tool for sustained production traffic; the architecture is built so
  the ingress could later be swapped for a named tunnel or a real domain
  without touching the app.
- Argon2 parameters are modest to keep the demo snappy in a container.
- LLM-generated code should be **reviewed before you rely on it**.

## License

AGPL-3.0-or-later. See [`LICENSE`](LICENSE). If you run a modified version of
this software as a network service, the AGPL requires you to offer your users
the corresponding source.
