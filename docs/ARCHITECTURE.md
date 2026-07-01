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
The header carries **no** `report-uri`/`report-to`, because the API's URL is
only known at runtime. Instead, `frontend/js/csp.js` listens for the in-page
`securitypolicyviolation` event (which fires for both report-only and enforcing
policies, regardless of any reporting endpoint) and POSTs a compact JSON report
to the API.

`internal/csp` normalises the three report shapes browsers actually send — the
legacy `application/csp-report` object, the Reporting API array, and the custom
body from the in-page listener — into one row. Each report is both stored in
`csp_reports` and logged through the OpenTelemetry-backed logger, so violations
show up in telemetry too.

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
