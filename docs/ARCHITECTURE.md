# Architecture

This document explains the non-obvious design decisions. For a high-level tour
and how to run it, see the [README](../README.md).

## The core constraint: no exposed host ports

No container publishes a port to the host. Postgres, the API, and the frontend
talk to each other only on the internal Compose network, addressed by service
name (`db`, `api`, `frontend`). The single way anything reaches the outside
world is a `cloudflared` container opening an **outbound** connection to
Cloudflare's edge, which hands back a public `*.trycloudflare.com` URL. There
are no inbound firewall rules and nothing listening on the host's interfaces.

There are two tunnels — one for the frontend, one for the API — so each has its
own public URL. A human can browse to either directly (handy for poking the API
with `curl`). The browser, after loading the page from the frontend tunnel,
calls the **API tunnel URL** directly for its `fetch()` calls; it does not proxy
through the frontend. Note that "direct" still means over the public internet
through Cloudflare's edge, because the browser runs on a user's machine and has
no route to `api:8080` on the Compose network. That makes the API calls a real
cross-origin request, which is exactly why CORS matters here.

## The bootstrapping problem and staged startup

Quick Tunnel URLs are random and only known **after** `cloudflared` starts and
prints them to its log. Two things depend on the frontend's URL:

- the **WebAuthn Relying Party ID** must equal the frontend's registrable
  domain, and the RP origin must be the frontend's full origin; and
- the API's **CORS allow-list** must include the frontend origin.

And the frontend needs the **API's** URL to make calls. So `scripts/up.sh`
starts things in an order that avoids container restarts:

1. Start `db`; wait until healthy.
2. Start `frontend` + `cloudflared-frontend` with `--no-deps` (so the API is
   not dragged up yet).
3. Poll `cloudflared-frontend` logs until the frontend URL appears. Derive
   `GOTUNNELS_RP_ID` (the host) and set the RP origin and CORS origin to the
   full URL. Persist these to `.env` and export them.
4. Start `api` + `cloudflared-api`; the API now boots with the correct RP and
   CORS configuration.
5. Poll `cloudflared-api` logs for the API URL.
6. Write `config.json` (containing `apiBase`) into the running frontend
   container at `/srv/config.json` via `podman cp`.
7. Print both URLs.

The frontend's static files are therefore identical for every instance; only
the small `config.json` differs, and it is injected at run time. On page load,
the browser fetches same-origin `/config.json` to learn the API base.

## Sessions: Bearer token first, cookie second

Because the frontend and API are different origins, a session **cookie** would
be a third-party cookie from the API's perspective — increasingly blocked by
browsers. So the primary session transport is an opaque **Bearer token**
returned in the login/signup JSON body and stored in `sessionStorage`; every API
call sends `Authorization: Bearer <token>`. The server also sets a
`SameSite=None; Secure` cookie as a secondary path, but nothing depends on it.

The token is random 256-bit material. The database stores only `sha256(token)`
as the session's primary key, so a database leak does not expose usable tokens.

## Authentication model

Every account is created with a **password** (argon2id) so there is always a
working credential; signup then immediately issues a session. Once signed in a
user can:

- **Register passkeys** (WebAuthn). Registration and the sign-count update on
  login go through go-webauthn; the full `webauthn.Credential` is stored as JSON
  (the source of truth for reconstruction) alongside broken-out columns for
  indexing and uniqueness.
- **Enroll TOTP**. A secret is generated, **encrypted at rest with AES-256-GCM**
  (key derived from `GOTUNNELS_TOTP_ENCRYPTION_KEY`), and stored unconfirmed
  until the user proves possession with one code. Ten single-use recovery codes
  are generated and only their hashes are stored.

Login accepts password (plus a TOTP or recovery code if TOTP is enabled) or a
passkey assertion. WebAuthn ceremony state lives server-side in `webauthn_flows`
keyed by a random flow id echoed back on the finish request, so the ceremony
never depends on a cross-site cookie.

Authorization is a simple role model (`user` / `admin`) seeded by the first
migration.

## Privacy: hashed IPs

No raw IP is ever written. `internal/activity` computes `sha256(pepper || ip)`
as lowercase hex, where the pepper is a per-instance secret. A bare hash of an
IPv4 address would be trivially reversible with a rainbow table (only ~4 billion
values); the pepper defeats that while preserving the useful property that the
same visitor produces the same hash. The activity page shows users their own
hashes plainly, making the privacy design visible rather than hidden.

The same hashing keys the CSP endpoint's rate limiter, so abuse protection also
never needs a raw IP.

## Content-Security-Policy

Caddy emits the CSP header, centrally configured via environment variables. It
ships as `Content-Security-Policy-Report-Only` with a strict self-only policy.

Violation reporting is **native browser reporting**. The API's public URL is a
Quick Tunnel only discovered at runtime, so the policy cannot name it directly
— instead Caddy exposes a stable same-origin path, `/csp-report`, and
reverse-proxies it over the internal compose network to the API's
`POST /api/csp-reports` (upstream configurable via `GOTUNNELS_API_UPSTREAM`,
default `api:8080`). The Caddyfile appends
`report-uri /csp-report; report-to csp-endpoint` to whatever policy is
configured and sends a matching `Reporting-Endpoints: csp-endpoint="/csp-report"`
header, so browsers on the legacy `report-uri` path and on the modern
Reporting API both deliver reports, in report-only and enforcing modes alike.
Because the directives are appended **outside** the `GOTUNNELS_CSP_POLICY`
value, reporting keeps working even with a customised (or stale) policy in an
old `.env`. Same-origin also means no CORS preflight and no runtime URL
discovery; the earlier in-page `securitypolicyviolation` listener
(`frontend/js/csp.js`) is gone — native reporting also catches violations that
occur before any JavaScript loads.

`internal/csp` normalises the report shapes browsers actually send — the
legacy `application/csp-report` object and the Reporting API
`application/reports+json` array (plus the camelCase shape the former in-page
listener posted, kept for compatibility) — into one row. Each report is both
stored in `csp_reports` and logged through the OpenTelemetry-backed logger, so
violations show up in telemetry too.
`internal/config/csp_deployment_test.go` pins every duplicated copy of the
default policy to `config.DefaultCSPPolicy` and pins the Caddyfile's reporting
wiring, and the container smoke test posts both wire formats through the
proxy end to end.

## Telemetry

`internal/telemetry` configures the vendor-neutral OpenTelemetry Go SDK for all
three signals and exports them over OTLP/HTTP. Logs are always written to stdout
as JSON and, when an endpoint is configured, additionally shipped via the OTel
log bridge with trace/span correlation. When no endpoint is set, trace and
metric providers are no-ops and only stdout logging remains. Resource
attributes (`service.name`, `service.instance.id`) are environment-driven so
multiple instances reporting to one backend stay distinguishable.

## Request lifecycle

An API request passes through: OpenTelemetry HTTP instrumentation (span +
metrics) → CORS (which also answers `OPTIONS` preflight before routing) → panic
recovery → request-id assignment → the method-aware `ServeMux` → for protected
routes, the `RequireAuth` middleware that resolves the Bearer token to a session
and loads the user → the handler. Handlers speak only to `internal/store`, never
raw SQL.

## Migrations

`internal/database` embeds the `*.sql` files and applies pending `*.up.sql`
migrations on startup, each in its own transaction, recording applied versions
in `schema_migrations`. It is a deliberately tiny, dependency-free runner rather
than a third-party migration tool.

## User preferences

`user_prefs` is a per-user key/value table (`user_id`, `key`, `value`) exposed
through `GET/PUT /api/prefs/{key}`. It exists because some settings should
follow the *account* and some should stay with the *device*, and the split is
deliberate:

- **Account settings** (server-side prefs): things that describe the person's
  choice regardless of hardware — e.g. whether the CAPTCHA leaderboard is
  expanded (`captcha.leaderboard.open`). Log in anywhere and it looks the way
  you left it.
- **Device settings** (localStorage): things that depend on the machine —
  e.g. the Magic Solve speed, which is bounded by the display's refresh rate.
  Carrying a 144 Hz speed setting to a 60 Hz laptop would be wrong.

Keys are constrained to `[a-z0-9._-]`, must start alphanumeric, and are capped
at 64 characters; values are capped at 4 KiB. Prefs are plain strings — the
page owns the encoding. This keeps the endpoint generic without turning it
into an arbitrary blob store.

## CAPTCHA statistics: batched sync, not per-solve requests

The CAPTCHA page's auto-solver can finish *hundreds* of puzzles per second, so
one-request-per-solve was never on the table. Instead:

- The browser accumulates deltas locally: `{manual_delta, auto_delta,
  current_streak, best_streak}`.
- It flushes them to `POST /api/captcha/sync` every ~4 seconds while anything
  is pending, when the auto-solver stops, and — via `fetch(..., {keepalive:
  true})` — when the tab is hidden or unloaded. (`navigator.sendBeacon` cannot
  be used: it cannot carry the `Authorization` header this API relies on.)
- The server applies the batch as a single UPSERT into `captcha_stats`:
  totals are incremented, `best_streak` takes `GREATEST(stored, claimed)`, and
  `current_streak` is last-write-wins so a session can resume on another
  device.
- The server clamps everything (per-sync deltas to 100 000, streaks to 10⁹)
  and both layers enforce non-negativity — the DB with CHECK constraints, the
  handler before the query. Scores are self-reported by client code; the
  clamps bound the damage of a dishonest client without pretending this is an
  anti-cheat system.

Displayed solves are `server base + un-flushed pending + in-flight`, so the
number never moves backwards while a request is in the air. If a sync fails,
its deltas are restored to the pending pile and retried; nothing is lost short
of the process dying mid-flight, which costs at most a few seconds of play.

The leaderboard (`GET /api/captcha/leaderboard`) ranks by `best_streak` (ties:
`total_solves`, then earliest update) with SQL `RANK() OVER`, returns the top
10 plus the caller's own ranked row, and is collapsed by default behind a
`<details>` element whose open state is the server-side pref above.

`POST /api/captcha/reset` deletes the caller's row entirely — the "reset my
stats" button is honest: the server forgets you, and you leave the
leaderboard.

## Notes: a deliberately minimal shared microblog

`notes` is a plain-text, everyone-sees-everything feed with three verbs: list,
create, delete. The constraints are the feature:

- **Plain text only.** Bodies are validated server-side (valid UTF-8, CRLF
  normalized to LF, no control characters besides newline and tab, 1–500
  runes — the DB enforces the same range with a `char_length` CHECK) and
  rendered client-side exclusively via `textContent`. Nothing in a note can
  become markup; URLs are visibly text and not clickable, by design. The
  per-note **Copy** button exists precisely because links do not work — you
  copy, you inspect, you decide.
- **Delete, never edit.** A note is either exactly what its author wrote or it
  is gone. No edit endpoint exists, so no "edited after people replied"
  ambiguity can exist either. Deletes are hard `DELETE`s (the row is gone, not
  flagged), matching the privacy posture of the rest of the project, and the
  table uses `ON DELETE CASCADE` from `users` — content leaves with the
  account, unlike `activity_log`'s `SET NULL`, which is an *audit* record.
- **Ownership in the query.** `DELETE FROM notes WHERE id = $1 AND user_id =
  $2` — authorization is part of the statement, not a separate read followed
  by a write. Deleting someone else's note and deleting a nonexistent note are
  the same uniform 404, so the endpoint is not an existence oracle.
- **Rate-limited creation.** `POST /api/notes` sits behind the existing token
  bucket (0.5 rps sustained, burst 5) keyed by the *authenticated user id* —
  inside `RequireAuth`, so the key is server-derived, not client-controlled.

## Live updates: polling, on purpose

The notes feed auto-refreshes by **polling** (`GET /api/notes`, newest 50)
rather than SSE or WebSockets, and this is a considered choice, not a
shortcut:

- The API's `http.Server` sets `WriteTimeout: 30s`, which kills any
  long-lived streaming response. Streams would also punch through the
  otelhttp instrumentation (one span per infinite response) and interact
  badly with the Cloudflare tunnel's buffering.
- Polling makes every update a normal, small, instrumented, rate-limitable
  request — the same shape as everything else in the API.

The client polls every 5 s **only while the tab is visible**, refreshes
immediately on becoming visible, and backs off exponentially to 60 s while
the API is unreachable. Each response is reconciled against the rendered feed
keyed by note id: notes that disappear from the newest-50 window are removed
(deletions propagate live), your own new notes appear immediately, and other
people's new notes appear immediately *unless you have scrolled down to
read* — then they wait behind a "New notes ↑" pill so the feed never jumps
under your thumb. A deletion older than the newest-50 window is not detected
live; it surfaces on the next full page load, which is an accepted trade-off
of window-based reconciliation.

## Page gating: shells are public, data is not

`/notes` and `/captcha` are "login-only" the same way `/activity` and
`/settings` are: Caddy happily serves the static HTML shell to anyone, the
page's script calls `requireAuth()` and redirects anonymous visitors to
`/login`, and — the part that actually matters — **every API route behind
those pages requires a session server-side** (`RequireAuth`). The redirect is
UX; the enforcement is the 401. There is nothing sensitive in the static
shells, so protecting them would add a server-side rendering dependency
without adding security.

## Domain metrics

Beyond the automatic HTTP metrics, the new features export a few
domain-level OpenTelemetry instruments: `gotunnels.captcha.solves` (counter,
attribute `mode=manual|auto`), `gotunnels.captcha.syncs`,
`gotunnels.captcha.streak` (a histogram of streak values reported at sync
time), `gotunnels.notes.created`, and `gotunnels.notes.deleted`. Handlers
also annotate their active span (batch sizes, note ids) so traces answer
"what did this request actually do" without log-diving.
