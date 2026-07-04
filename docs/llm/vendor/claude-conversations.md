00
54
19

I have added a new /captcha page 
however, right now it is frontend only
as a demonstration of the website's ability to handle user information
we need to store this in the backend on our database
but also we need to add opentelemetry here as well 
also we need to have a leaderboard, but only for people who are signed in
on the same page, collapsed by default 
but we should remember the user setting 
if they expand it / collapse it again
this captcha page will only be available to people who are logged in

also the index.html is a little strange
it says 
Get Started 
Sign up / Log in 
even though I am logged in 
and the settings page correctly displays the logout button 
and even the index / home page clearly identifies 
earlier in the page that 
I am `signed in as x`

another big feature we can add is a new page where everyone can add a new short note
something like a microblog or a tweet, plain text only, no attachments allowed. 
and everyone can see everyone else's posts. 
there are no images, no image previews, no link previews, 
and even the links in posts are not actually clickable 
however, it should be very easy to copy paste any post 
also any user should be able to delete an old post but NOT edit them
you should not be able to delete any one else's posts 
this page should also be responsive and mobile friendly 
see what I changed with the /activity page for inspiration (basically cards)
also this page should somehow autorefresh 
and this might be the most difficult part of the whole operation
this feed is only available to people who are logged in

please document all architectural decisions 
and please update all readme as necessary as we continue to improve this application 

please return full files as well as full path for all files that need to change


/
Claude finished the response
I have added a new /captcha page 
however, right now it is frontend only
as a demonstration of the website's ability to handle user information
we need to store this in the backend on our database
but also we need to add opentelemetry here as well 
also we need to have a leaderboard, but only for people who are signed in
on the same page, collapsed by default 
but we should remember the user setting 
if they expand it / collapse it again
this captcha page will only be available to people who are logged in
also the index.html is a little strange
it says 
Get Started 
Sign up / Log in 
even though I am logged in 
and the settings page correctly displays the logout button 
and even the index / home page clearly identifies 
earlier in the page that 
I am signed in as x
another big feature we can add is a new page where everyone can add a new short note
something like a microblog or a tweet, plain text only, no attachments allowed. 
and everyone can see everyone else's posts. 
there are no images, no image previews, no link previews, 
and even the links in posts are not actually clickable 
however, it should be very easy to copy paste any post 
also any user should be able to delete an old post but NOT edit them
you should not be able to delete any one else's posts 
this page should also be responsive and mobile friendly 
see what I changed with the /activity page for inspiration (basically cards)
also this page should somehow autorefresh 
and this might be the most difficult part of the whole operation
this feed is only available to people who are logged in
please document all architectural decisions 
and please update all readme as necessary as we continue to improve this application 
please return full files as well as full path for all files that need to change

Show less

Architected backend persistence, leaderboard system, notes microblog, and frontend integration
Architected backend persistence, leaderboard system, notes microblog, and frontend integration
I'm reviewing the project structure from the dump file to understand the current codebase, then I'll work on implementing the three main features: a captcha backend with database storage and OpenTelemetry instrumentation, a leaderboard component on the captcha page that's only visible to signed-in users and collapsed by default with persistent state, and what looks like additional captcha-related functionality that got cut off.

The user also wants me to fix the index.html to properly reflect login state, create a new microblog/notes page with plain text posts that auto-refresh, document architectural decisions, and update READMEs. I should be careful with the GitHub PAT they provided—I'll only use it for reading public endpoints if necessary and won't commit it to any files.

Let me start by exploring the project structure and reading the dump.txt file to understand what I'm working with.

Show less

List project files and get the file index from dump.txt
bash

ls -la /mnt/project/ && wc -l /mnt/project/dump.txt && grep -n '^# FILE:' /mnt/project/dump.txt | head -100
Output

total 412
drwxr-xr-x 2 root root   4096 Jul  4 14:07 .
drwxr-xr-x 6 root root   4096 Jul  4 14:07 ..
-rw-r--r-- 1 root root 411001 Jan  1  1980 dump.txt
10889 /mnt/project/dump.txt
35:# FILE: export.sh  [THIS SCRIPT — included for full context]
295:# FILE: ${SCRIPT_NAME}  [THIS SCRIPT — included for full context]
534:# FILE: cmd/api/main.go
704:# FILE: compose.yaml
815:# FILE: Containerfile.api
865:# FILE: Containerfile.frontend
911:# FILE: .containerignore
949:# FILE: docs/ARCHITECTURE.md
1110:# FILE: docs/CONFIGURATION.md
1244:# FILE: .env.example
1322:# FILE: export.sh
1587:# FILE: ${SCRIPT_NAME}  [THIS SCRIPT — included for full context]
1721:# FILE: frontend/activity.html
1796:# FILE: frontend/Caddyfile
1869:# FILE: frontend/captcha.html
2564:# FILE: frontend/css/style.css
2861:# FILE: frontend/index.html
2947:# FILE: frontend/js/api.js
3054:# FILE: frontend/js/common.js
3136:# FILE: frontend/js/config.js
3174:# FILE: frontend/js/csp.js
3237:# FILE: frontend/js/page-activity.js
3314:# FILE: frontend/js/page-index.js
3366:# FILE: frontend/js/page-login.js
3448:# FILE: frontend/js/page-settings.js
3582:# FILE: frontend/js/page-signup.js
3640:# FILE: frontend/js/webauthn.js
3760:# FILE: frontend/login.html
3840:# FILE: frontend/settings.html
3963:# FILE: frontend/signup.html
4040:# FILE: .github/dependabot.yml
4077:# FILE: .github/workflows/ci.yaml
4129:# FILE: .github/workflows/dependency-freshness.yaml
4176:# FILE: .github/workflows/govulncheck.yaml
4226:# FILE: .gitignore
4275:# FILE: go.mod
4332:# FILE: internal/activity/activity.go
4441:# FILE: internal/activity/activity_test.go
4524:# FILE: internal/auth/handlers.go
5294:# FILE: internal/auth/handlers_test.go
5396:# FILE: internal/auth/passkey.go
5460:# FILE: internal/auth/passkey_test.go
5521:# FILE: internal/auth/password.go
5646:# FILE: internal/auth/password_test.go
5754:# FILE: internal/auth/session.go
5805:# FILE: internal/auth/session_test.go
5871:# FILE: internal/auth/totp.go
6051:# FILE: internal/auth/totp_test.go
6212:# FILE: internal/config/config.go
6571:# FILE: internal/config/config_test.go
6770:# FILE: internal/csp/csp.go
7068:# FILE: internal/csp/csp_test.go
7227:# FILE: internal/database/database.go
7418:# FILE: internal/database/database_test.go
7509:# FILE: internal/health/health.go
7616:# FILE: internal/health/health_test.go
7695:# FILE: internal/httpx/middleware.go
7955:# FILE: internal/httpx/middleware_test.go
8144:# FILE: internal/server/server.go
8254:# FILE: internal/store/store.go
8875:# FILE: internal/telemetry/multihandler.go
8954:# FILE: internal/telemetry/multihandler_test.go
9029:# FILE: internal/telemetry/telemetry.go
9247:# FILE: LICENSE
9503:# FILE: migrations/0001_init.down.sql
9528:# FILE: migrations/0001_init.up.sql
9581:# FILE: migrations/0002_auth_credentials.down.sql
9608:# FILE: migrations/0002_auth_credentials.up.sql
9695:# FILE: migrations/0003_sessions.down.sql
9718:# FILE: migrations/0003_sessions.up.sql
9755:# FILE: migrations/0004_activity_and_health.down.sql
9779:# FILE: migrations/0004_activity_and_health.up.sql
9833:# FILE: migrations/0005_csp_reports.down.sql
9856:# FILE: migrations/0005_csp_reports.up.sql
9906:# FILE: migrations/embed.go
9937:# FILE: README.md
10180:# FILE: scripts/down.sh
10225:# FILE: scripts/lib.sh
10552:# FILE: scripts/run.sh
10624:# FILE: scripts/test.sh
10729:# FILE: scripts/up.sh
I'm going to review the key files to understand the current codebase structure, starting with the entry point and working through the routing, database layer, and frontend components to see what's already implemented.


Read cmd/api/main.go

Script

Read internal/server/server.go

Script

Read internal/store/store.go

Script
I'm examining the frontend structure now, looking through the HTML pages for the captcha and main interface, along with the JavaScript modules that handle page logic, API communication, and authentication.


Read frontend/captcha.html

Script
I need to examine the stylesheet, markup, and JavaScript files to understand the project structure.


Read frontend/css/style.css, index.html, api.js, common.js, config.js, csp.js

Script

Read frontend/activity.html, page-activity.js, page-index.js, page-login.js

Script
I need to examine the auth handlers to understand the authentication patterns, then look at the httpx middleware instrumentation, activity tracking, telemetry setup, and database migrations before moving on to configuration, compose files, and documentation.


Read internal/auth/handlers.go

Script
I need to go through the remaining files to get the full picture of the codebase.


Read internal/activity/activity.go and internal/httpx/middleware.go

Script

Read internal/telemetry/telemetry.go and migration 0004/0005 up files

Script
I'm going to examine the README, architecture documentation, Go module file, settings page patterns, health check implementation, and database migration files to understand the codebase structure and how metrics are being tracked.


Read README.md

Script
I'm going to examine the architecture documentation, module dependencies, and health check implementation to understand the patterns being used. Then I'll look at the settings pages and authentication gating to see how they're structured.


Read docs/ARCHITECTURE.md and go.mod

Script
I'm going to examine the health handler patterns, review the settings page structure and JavaScript, check the config setup, look at the login and signup templates, and see how the database migrations are handled.


Read internal/health/health.go and frontend/settings.html

Script
I should review the config.go file to understand what configuration options are available, check the database migration logic in database.go, and look at the compose.yaml alongside the Caddyfile to see how everything fits together.


Read first part of internal/config/config.go

Script
I've got a good overview now, so let me check the database migration function and the compose configuration to understand the setup, then I'll review the JavaScript patterns in the existing files to keep things consistent before planning out the changes.


Read internal/database/database.go

Script
I'm mapping out the implementation plan for the captcha system, starting with the database migrations to create tables for tracking user statistics like streaks and solve counts, along with detailed solve records.

The key architectural decision is that since the Magic Solve can generate hundreds of solves per second, I need to avoid per-solve database rows and HTTP requests. Instead, the client will batch up solve deltas and sync them periodically to the server, which accumulates the aggregate stats. For the leaderboard collapse/expand state, I'm considering a server-side user preferences table to keep settings consistent across devices rather than relying on localStorage. I'll keep it simple with server-side prefs only and graceful degradation if the call fails. I'm sketching out the storage methods: UpsertCaptchaStats to accumulate solves and manage streaks, GetCaptchaStats for retrieval, CaptchaLeaderboard to rank users by best streak with tiebreakers, and generic GetUserPref/SetUserPref for the collapse state. Then I'm building a new captcha package with API handlers for fetching stats and syncing updates from the client, validating the incoming deltas to prevent absurd values.

For observability, I'm adding OpenTelemetry metrics like a counter for solves tagged by mode and a histogram for streaks, plus span attributes within the handlers to track what's happening.

For preferences, I'm setting up two endpoints to get and update user prefs by key, validating the key format and capping value size at 4KB.

I'm splitting the migrations into separate files—one for captcha stats and another for the notes feature to keep concerns isolated.

For the notes table, I need to decide on the body character limit (thinking tweet-like at 500 chars) and whether to use soft or hard deletes—leaning toward hard delete for simplicity since the user just wants to remove old posts.

The store layer needs CreateNote, ListNotes with pagination support (accepting an after ID and limit for efficient polling), and DeleteNote with ownership verification. For the API handlers, I'm building a GET /api/notes endpoint that accepts after and limit query parameters, returning notes ordered newest-first with the latest ID included in the response. I'll add a per-user rate limiter for note creation (0.5 notes/sec with burst of 5) using the existing RateLimiter keyed by user ID from context after authentication. The middleware stack will be authed(noteLimiter.LimitByIP(userKeyFn)(handler)) so the user context is available for the rate limit key function. Auto-refresh is next and might be the trickier part.

For auto-refresh, I'm choosing polling with an after cursor over WebSocket or SSE because SSE connections would get killed by the 30-second WriteTimeout and inflate span durations, while polling every 5 seconds with a simple indexed query is stateless and robust across the proxy layers.

For the frontend notes page, I'm building a card-based layout with a composer textarea, character counter, and a list of note cards showing author info, timestamps, and body text. Each card gets a Copy button using the clipboard API with a fallback, and a Delete button that only appears on your own notes. New notes get prepended with a subtle highlight animation, and deleted notes are handled through polling.

The key insight is that polling for the latest 50 notes each time and reconciling client-side is the simplest correct approach — it handles deletions and edits cleanly at a 5-second interval without much overhead. Rather than rebuild the DOM every poll (which causes flicker), I'm diffing by ID to map existing notes, remove missing ones, and insert new ones in the right positions. Since the ordering is stable by ID descending, the reconciliation logic stays straightforward.

For pagination, I'm adding a before parameter to support loading older notes with a "Load older" button — it's genuinely useful and keeps the API simple with just ListNotes(before, limit). The polling will refresh only the latest 50, so the scope stays manageable for the demo.

Now I need to handle reconciliation carefully: when polling fetches the latest 50, I'll upsert them into the in-memory map and detect deletions by checking if any currently rendered note with an id greater than the minimum fetched id is missing from the new fetch — those get removed. Notes older than the fetched window stay untouched, so deletions of very old notes won't trigger unnecessary updates.

For loading older notes, I'll use a separate request with a before parameter to append earlier entries. The in-memory state stays as a sorted map by id in descending order.

On the auth side, I'm gating both the captcha and notes endpoints at the API layer while keeping the HTML shells public (consistent with the activity and settings pages) — the client-side redirect is purely for UX, not security. For the captcha puzzle itself, I need to think through how to handle that...

Since the puzzle is client-side rendered, an unauthenticated user could technically access it if they had the HTML, but the requireAuth() check at boot prevents that, and any stats sync would fail with a 401 anyway — same model as the rest of the app.

I also need to add navigation links for CAPTCHA and Notes across all pages, which means updating the nav in every HTML file. And for the index page's Get Started card, I'm adding conditional rendering based on auth state — showing sign-up/login buttons when logged out, and links to Notes/CAPTCHA/Settings when logged in.

There's a flash issue to handle: the signed-out content is visible by default, then hidden after the /api/me check completes. I'll keep this consistent with the existing nav behavior. The card will have two divs — one for logged-out state with signup/login buttons, and one for logged-in state (hidden by default) with a confirmation message and links to the other features.

For the CAPTCHA stats, I'm moving them server-side so they sync on boot via GET /api/captcha/stats, but keeping the speed slider in localStorage since that's device-specific. The "Clear my local data" button should be renamed to "Reset my stats" and call a new POST /api/captcha/reset endpoint to reflect what it actually does now.

On the leaderboard toggle — the user asked for backend storage of captcha results, not necessarily UI state. The leaderboard expanded/collapsed preference could stay in localStorage to keep things simpler, since remembering that doesn't require server infrastructure.

But a user_prefs table is minimal (one table, two methods, two handlers) and demonstrates handling user information across devices, which fits the project's theme. I'll store the toggle server-side with no local cache — the page is already auth-gated and fetches stats from the server anyway, so one more GET request is negligible. This keeps the principle clean: user settings live server-side, device settings (like the speed slider) live locally.

For syncing, I'll accumulate changes on the client and flush every 4 seconds if there are updates, plus on visibility change when the tab comes back into focus. and rename the button to "Reset my CAPTCHA data" so it clears both server and local stats, which aligns with the original intent. For the leaderboard, I'll use a native <details> element for accessible collapsing, fetch the leaderboard when opened, and re-fetch after each stats flush so the user's row updates in real time.

For OpenTelemetry instrumentation, I'm adding package-level tracer and meter with counters for CAPTCHA syncs and solves (tagged by mode), plus a histogram to record the current streak on each sync, and a counter for notes created.

For notes, I'll track creation and deletion with counters, and record body length via histogram or just set it as a span attribute. I'm using trace.SpanFromContext(ctx).SetAttributes(...) to attach metadata to the request span rather than creating child spans everywhere—keeping it focused with counters per operation and attributes for context. When constructing the meter and counters, I need to handle any errors that come back and log a warning if instrumentation setup fails. Creating instrument handlers in NewHandlers where the logger is available, with nil checks guarding their usage. Then wiring up the server routes for the captcha and notes endpoints with authentication middleware.

For the dependencies, I'm adding Captcha and Notes handler structs to Deps, plus a NotesRateLimiter keyed by user ID for consistency with the existing CSPRateLimiter pattern. The rate limiter gets passed into notes.NewHandlers so it can be applied at the handler level. Still deciding whether the prefs handler belongs in the captcha package or elsewhere. Adding API wrapper functions... Wait, there's a CSP conflict here — the existing captcha.html has inline scripts and styles, but the README claims the app uses only external scripts and styles. That inline code would violate the script-src 'self' and style-src 'self' policies, even if it's currently in report-only mode and still functioning.

To fix this properly, I need to extract the inline <style> block and inline <script> into separate files — moving the styles into the main stylesheet with appropriate class scoping, and moving the JavaScript into a dedicated module file in the frontend directory. I should also replace any inline style attributes with CSS classes. This keeps the architecture consistent with the project's principle of external-only scripts and styles, and ensures genuine CSP compliance.

Now I'm updating the documentation: adding sections to the architecture guide covering the CAPTCHA stats model, user preferences, the notes feature, why polling is used for auto-refresh, and clarifying that static shells are public while data endpoints are gated. The README needs updates too — expanding the feature list, showing the new internal modules, and reflecting the repository structure. For configuration, I'm checking if any new environment variables are needed; since I'm keeping limits as hardcoded constants, there's nothing to add to CONFIGURATION.md, though I should document the body size limits in the architecture guide instead.

For tests, I'm adding unit tests for the new pure logic functions: notes body validation, CAPTCHA delta clamping and merge logic, and preference key validation. These will live in their respective packages and stay dependency-free, matching the existing test patterns in the repo.

The user offered a GitHub PAT for API rate limits, but since I'm not adding any new Go dependencies and everything uses stdlib plus existing deps, I won't need it. I should mention that sharing PATs in plaintext is risky and suggest revoking it, though I won't be using it anywhere.

On the build side, I can't run a full go build because the Go module proxy isn't allowlisted, but I can work around this with direct GitHub fetches for some deps. However, the OpenTelemetry vanity import path won't resolve that way, so a complete build is likely impossible. My validation strategy instead uses gofmt -e for Go files, bash -n and shellcheck for shell scripts, and Python for YAML sanity checks.

For the code itself, I need to verify the Go version supports the APIs I'm using—like r.PathValue() from Go 1.22 and the OpenTelemetry metric API with counters and attributes. Let me check if Go is even available in this container first, then manually review the Go files since I can't resolve all dependencies.

For the leaderboard query, I need to pull in the display_name from the users table and calculate the rank using a window function ordered by best_streak and total_solves, then figure out how to get the user's own rank.

I'll create two separate functions: one that fetches the top leaderboard entries with their ranks, and another that retrieves just the current user's rank and stats. The handler can call both and combine the results.

For the notes, I'm setting up the JSON shape with user info and a client-side mine flag computed by comparing against the authenticated user. I'll add relative time formatting with absolute timestamps in tooltips, a copy button that uses the clipboard API with a fallback to the older execCommand approach, and clear visual feedback when text is copied. The note body text will be naturally selectable for easy copying.

Since we're rendering with textContent, links won't be clickable—no anchors get created. For keeping notes fresh, I'm implementing a polling system that checks every 5 seconds when the page is visible, stops when hidden, and uses exponential backoff on errors up to 60 seconds with a subtle reconnecting indicator. New notes from other users will slide in with a highlight animation, and deletions get reconciled by removing them from the list.

Implementing the reconciliation logic... For scrolled-down users, I'll show a "N new notes" pill instead of auto-prepending, which shifts their view — they can click it to jump to the top and apply the updates. This is the polish that makes autorefresh feel right. I'll keep the deletion reconciliation immediate regardless of scroll position. Migration 0007 handles the notes schema, and I need to finalize the design details.

Now I'm thinking through the sync payload constraints — clamping deltas to a reasonable range based on what the auto-solver can actually produce in a 4-second window (roughly 1000 solves at max), streaks capped at 1 billion and only decreasing via explicit reset, and ensuring the client's reported solve delta matches the sum of manual and auto solves.

For the database schema, I'm defining a captcha_stats table that tracks user performance with columns for best streak, current streak, total solves, and a breakdown of manual solves.

Adding more tables... Adding the user_id index and considering whether to log note creation events separately — the activity log is really for auth auditing, so I'll skip polluting it with note creation. I should also check the login and signup pages to understand the navigation structure. I'm acknowledging the scope of about 22 files across backend and frontend, then moving on to examine the login and signup HTML files.

Show less
I have a clear view of the backend now. Next I'll pull the login/signup HTML files and check the validation tooling.


Read login.html and signup.html; check available validation tools

Script
I don't have Go installed in this container, so I'll try installing it via apt since that package repository is accessible. The gofmt tool is what I really need for syntax validation. Once that's set up, I'll create a working directory that mirrors the repo structure, validate all the files, and proceed from there.

I need to install Go to validate the new Go files and check if Node is available for the JavaScript syntax checking.


Install Go for gofmt validation; check node availability

Script
I need to check the Go and Node versions available for syntax checking, then start building out the workspace files. For the CAPTCHA stats, I'm thinking about how to handle the high throughput—batching deltas on the client side and aggregating them server-side makes sense since the solver can complete dozens or hundreds of solves per second, which would be inefficient to handle with individual HTTP requests or database rows.

For preferences, I'll use a server-side key/value table with allowlisted keys and a 4 KiB cap for values like leaderboard settings, while device-specific settings like Magic Solve speed stay in localStorage since they depend on the device's refresh rate. For notes, I'm going with append-only with hard deletes (no edits)—the database enforces a 1 to 500 character limit, and I'll return 404 uniformly for both missing and unauthorized notes since a single query can't distinguish between them anyway.

Auto-refresh will use visibility-aware polling instead of WebSockets or SSE, since the API has a 30-second write timeout that kills long-lived streams and polling is stateless and easier to scale horizontally—I'll fetch the latest 50 notes every 5 seconds when visible, pause when hidden, refresh immediately on focus, and back off exponentially on failure, showing a "N new notes" indicator when the user is scrolled down. Static pages are public but all data is gated by auth checks on the API routes, with client-side redirects purely for UX. I'm also moving inline styles and scripts out of captcha.html into external files to maintain the strict CSP policy without report-only violations.

No new Go dependencies needed—everything uses stdlib plus the OTel and pgx packages already in the project. I'm adding domain-specific metrics like captcha solve counters, note creation/deletion counters, and histograms for streak length and note body character count, with span attributes on request spans. Notes will be rate-limited per user ID using a token bucket approach.

For the routes, I'm using Go 1.22's method-aware mux patterns like "GET /api/prefs/{key}" and "DELETE /api/notes/{id}" with PathValue extraction, and CORS already supports PUT and DELETE methods. The final deliverables include migration files for the captcha/prefs schema and notes table. Now I'm wiring up the handlers in the server, then building out the frontend with new API endpoints, HTML pages for the captcha game and notes, JavaScript modules to handle game logic and leaderboard syncing, and styling for all the new features across the interface.

Defining the data models... Writing database query methods... Defining the CAPTCHA data structures... Writing the SQL query... handler will map that error to zeros. For the reset operation, I'm deciding between zeroing out the stats or deleting the row entirely—DELETE makes more sense as a true reset since it also removes the user from the leaderboard. For the leaderboard query, I'm using a window function to rank users by best streak and total solves.

Writing SQL queries... I'm deciding whether to add a shared DecodeJSON helper to the httpx package rather than duplicating decode logic across captcha and notes packages. The auth package already has its own private decoder, so I'll leave that as-is to keep changes minimal, but adding a reusable helper to httpx makes sense since it's the right place for JSON utilities and avoids duplication elsewhere.

Now I'm setting up the Handlers struct with the store, logger, and metric instruments...

On error from the meter, the returned instrument might be nil, so I need to add defensive nil-checks before calling Add to keep things safe regardless of whether the SDK is installed.

Defining the API endpoints...

The Stats handler retrieves the current user from context, fetches their CAPTCHA stats from the store (or creates an empty record if not found), and returns the stats as JSON. The Sync endpoint accepts a request body with manual and auto deltas plus streak information...

Defining the validation logic with clamping bounds... Defining more routes... I need to normalize line endings and filter out control characters to keep the input clean, allowing only newlines and tabs. For the GET endpoint, I'll parse the query parameters and return a notes list with a simple approach where the client can infer whether more results exist by checking if the response is full. On the POST side, I'm validating the body, creating the note, updating counters, and setting up the span attributes.

For DELETE, I'm parsing the note ID from the path, checking ownership, and returning a 200 response with a status message to keep the API client happy. Now I'm moving into the preferences module and setting up key validation using a regex pattern.

I'll implement a manual validation function for preference keys rather than using regexp, keeping it consistent with the auth module's style. For the Get endpoint, I'm returning a 200 response with the key, value, and an exists flag to avoid error noise on the client side. The Set endpoint validates the value size and performs an upsert, returning a success status.

Now I'm wiring up the dependencies — adding handlers for Captcha, Notes, and Prefs along with a rate limiter for notes. Then I'm registering the authenticated routes for captcha stats, sync, and reset operations.

Setting up more routes... Writing handler initialization... Writing API wrapper functions... Now I'm building the navigation bar that appears on all pages, with links that conditionally show or hide based on authentication status—home is always visible, while notes, captcha, activity, and settings only appear when logged in, and login/signup links show when logged out. Then I'm creating the get started card on the homepage that displays instructions for new users about setting up a passkey or TOTP.

Finishing the card layout... Building the leaderboard structure...

I need to scope the mobile table CSS carefully since the global media query transforms all tables into card layouts, which would break the leaderboard's compact format. The activity table needs the card transformation on mobile, but the leaderboard table should stay as a regular 4-column table since it fits fine on smaller screens. I'll add a class to the activity table and update the mobile CSS selectors to target only that specific table, keeping the display:block chain intact for each element type.

Now I'm setting up page-captcha.js with its module structure — importing the navigation and auth utilities, then bootstrapping by highlighting the current nav item, rendering the auth navigation, and checking authentication before loading the captcha stats and user preferences.

For data persistence, I'm using localStorage sparingly with a prefixed key scheme, and on boot I'll clean up any legacy stats keys as a migration step. The sync engine uses a pending queue that tracks manual and automatic solve attempts, with a timer that batches them together before flushing to the server at regular intervals. Handling sync errors and retries... Setting up signal-based state... Simplifying the bookkeeping approach: I'll track inflight count separately and keep pendingSignal always equal to the sum of manual pending, auto pending, and inflight items. On flush, I move pending counts into inflight; on success, clear inflight and update the base; on failure, restore the pending counts and clear inflight.

For triggering flushes, I'm using a 4-second debounce via scheduleSync that coalesces rapid changes, plus immediate flushes on visibility changes (hidden/pagehide) with keepalive to ensure data persists even if the page unloads. The periodic auto-solve loop naturally triggers flushes every ~4 seconds since queueSolve fires constantly. On initial load, I fetch stats to populate the best, base, and current streak signals.

For the UI, I'm adding a small sync status chip next to the fps display showing "saved", "saving…", or "offline — retrying" states. The leaderboard uses a preference key to remember if it was open, loading it on toggle and marking the user's own row while showing a "me" indicator if they're ranked outside the top 10.

When resetting stats, I'll clear all pending and inflight data along with the dirty flag, then reset the speed value to match the original behavior and reload the leaderboard if it's currently visible. I'm keeping the existing FPS counter, signal classes, and puzzle logic intact.

Now I'm building out the notes interface with a form for composing new notes—a textarea limited to 500 characters of plain text, along with the supporting layout and styling. I need to add textarea styling to the CSS since the current input rules only cover text inputs. For the main JavaScript file, I'm setting up the core state management with polling intervals, pagination limits, and data structures to track notes, their DOM elements, and the newest/oldest IDs in the feed. The approach uses an ordered array of IDs for reconciliation during renders. Writing implementation logic... Refining the pill logic for deferred notes: when polling, fresh notes from others get held in a deferred queue if the user is scrolled down (past 200px), showing a "N new" pill. Own notes insert immediately. Clicking the pill merges deferred notes into the map, syncs the DOM with highlighting, scrolls to top, and hides the pill. If the user scrolls near the top (under 50px), deferred notes auto-merge on the next poll. The tricky part is reordering the poll logic so fetched notes don't get inserted into the map before we decide whether to defer them—need to compute fresh notes separately and only add others' notes to the map if we're not deferring. ter deferred notes by checking if they still exist in the fetched window, dropping any that disappeared. For deletions, I'm removing the note from the map and fading it out from the DOM. When posting, I validate the message length client-side, disable the submit button, send it to the API, then prepend the user's own note directly to the list and clear the textarea—handling rate limit errors by showing the server's message. I'm also adding a character counter to the input field.

For the character counter, I need to account for the difference between UTF-16 code units and actual runes—the maxlength attribute counts code units while the server counts runes, so I'm using Array.from to get the true character count and warn styling when approaching the limit, while keeping maxlength as a soft guard. For deletion, I'm showing a confirmation dialog before calling the delete API and removing the note. Copy functionality uses the clipboard API with a fallback to execCommand, showing "Copied ✓" feedback for 1.2 seconds.

Writing polling logic... Setting up pagination for older notes with a load-older button that fetches earlier entries and appends them, tracking the oldest loaded ID to know when to hide the button. The deletion detection logic respects pagination boundaries—the poll window only affects the newest batch while older loaded pages stay untouched. For the status indicator, I'm toggling between "live", "paused" when the document is hidden, and "reconnecting…" during sync issues.

Adding styles for the textarea, composer footer layout, and the feed grid with individual note cards that have proper typography hierarchy—author name in bold, username and timestamp in smaller muted text, and the note body with preserved whitespace and word wrapping. Adding more styles... I need to add documentation sections covering user preferences, CAPTCHA batching, notes as a microblog, polling-based live updates, and the page gating model where static shells are public but data isn't. I'll also update the README with new bullets about what the project demonstrates, add the internal directories to the layout section, and mention the pages. The environment configuration stays the same, so I'll note that in the summary. Now I'm moving on to the notes.html route implementation.

Caddy's try_files directive handles routing /notes to notes.html, which works alongside the existing /captcha setup. I'm checking for potential issues: the DELETE /api/notes/{id} endpoint will trigger a CORS preflight, but the middleware handles OPTIONS requests before the mux and includes DELETE in the allowed methods. Go 1.22's ServeMux doesn't conflict with different HTTP methods on the same path, so GET /api/notes and DELETE /api/notes/{id} coexist fine. Note IDs are small integers, which are safe to handle as JSON in JavaScript. I'm also verifying the middleware chain for the notes creation endpoint.

Now I'm ready to write out all the files. I'll use create_file for the important ones to keep things clean, then validate everything with gofmt, Node, and Python checks.

For store.go, I need to add the new sections for captcha stats, user preferences, and notes before the helpers section at the bottom, with types coming after CSPReportInput. The migrations will be numbered 0006_captcha_and_prefs and 0007_notes with both .up.sql and .down.sql files — the down files follow the existing pattern even though the runner doesn't use them.

I'm setting the leaderboard limit to 10, keeping the captcha page's default empty status line until the first verification, and updating the reset dialog to explain that it deletes streaks, solve counts, and removes the user from the leaderboard while resetting the Magic Solve speed in the browser. I'll also preserve the legacy KEYS.stats cleanup and use direct getElementById in the captcha page script instead of a query selector.

Now I'm ready to write out all the files in order: workspace setup and database migrations, the Go packages, then the store and server logic, followed by the frontend code, documentation, and final validation.

Show less
Done
Go 1.22 (fine for gofmt -e syntax validation) and Node 22 are available. I have the full picture now. Let me set up a workspace mirroring the repo and build all the changes, validating as I go.


Architected backend modifications for captcha storage and telemetry integration
Architected backend modifications for captcha storage and telemetry integration

Create workspace directory structure

Script
Starting with migrations.


Migration 0006 up: captcha_stats + user_prefs tables
0006_captcha_and_prefs.up.sql

Write remaining migration files

Script
I need to extract the existing store.go file content from the dump and then add the new sections to it, making sure I capture everything accurately without transcription errors.

I'll extract the existing store.go file from the dump so I can modify it accurately.


Extract original file contents verbatim from dump.txt

Script
The file sizes match up correctly when accounting for the trailing newline I added, and I'm verifying that the extracted store.go file matches the SHA-256 hash from the metadata to ensure everything extracted properly.


Verify extracted files match the SHA-256 recorded in dump metadata

Script
All the hashes check out perfectly, so the extraction is clean. Now I'm going to place the originals and make targeted modifications using string replacement. Starting with store.go, I need to add the new types right after CSPReportInput, then insert the new sections for captcha stats, user preferences, and notes before the helpers section.

Done
All hashes match the dump metadata byte-for-byte, so I can make surgical edits on exact originals. Let me apply the store changes.


Rectified directory structure issue and verified migration files persisted
Rectified directory structure issue and verified migration files persisted

Seed workspace with verified originals

Script

Recreate dirs (heredoc cat may have clobbered cwd assumptions) and seed originals

Script
The brace expansion in the mkdir command created a literal directory instead of expanding properly, so I need to remove that stray folder and confirm the migrations directory is still intact.


Clean stray dir from failed brace expansion, verify migrations exist

Script
Done
Now the store edits — new types first, then the three new data-access sections.


Prepared to integrate DecodeJSON helper into middleware file
Prepared to integrate DecodeJSON helper into middleware file

Add Captcha, prefs, and Note types to the store type section

store.go
+45
-3

Add captcha stats, user prefs, and notes data-access sections before helpers

store.go
+183
-3
I'm looking at the httpx/middleware.go file to add a DecodeJSON helper function in the JSON helpers section, placing it after WriteError.


Add shared DecodeJSON helper to httpx JSON helpers

middleware.go
+18
-4
Done
Now the three new backend packages. First internal/captcha:


Created 2 files
Created 2 files

New captcha package: stats sync, reset, leaderboard handlers with OTel instrumentation
captcha.go

Unit tests for captcha clamp logic (dependency-free, matches repo test style)
captcha_test.go
Done
Now internal/notes:


Created 2 files
Created 2 files

New notes package: plain-text microblog handlers with validation and OTel
notes.go

Unit tests for note body validation
notes_test.go
Done
Now internal/prefs:


Scoping mobile table styles to preserve leaderboard formatting
Scoping mobile table styles to preserve leaderboard formatting

New prefs package: small per-user key/value preference endpoints
prefs.go
Let me check where I left off in the transcript and see what's still pending on my task list.


Check current workspace state to resume work

Script
I've got my workspace set up and there's a checklist of files I need to work through: starting with a test file for validating keys, then moving into server and API configuration changes, followed by frontend work including HTML templates and JavaScript modules for the captcha page and notes section.


Create prefs package unit tests
prefs_test.go
Let me check the current server.go file to see what needs to be modified.


View current server.go to plan edits

Script
Now I'm updating the server.go file to wire up the new dependencies—adding the captcha, notes, and preferences imports, then registering the handlers and rate limiter in the Deps struct, setting up the routes, and implementing the rate limiting key function.


Add new package imports to server.go

server.go
+10
-7

Add new handler dependencies to Deps struct

server.go
+14
-10

Add captcha, prefs, and notes routes to the authenticated section

server.go
+24
-3

Add noteRateKey helper after cspRateKey

server.go
+19
-7
I need to verify the auth.CurrentUser function signature returns a User and boolean, and confirm the User type has a string ID field. Let me search through the auth package dump to check this.


Verify CurrentUser and LimitByIP signatures in original codebase

Script

Verify User struct ID type

Script
The auth function signature checks out with the User ID being a string, so my server.go changes are valid. Let me run gofmt to verify the formatting is correct.


Verify gofmt on edited server.go and prefs_test.go

Script

See gofmt diff for prefs_test.go

Script

Auto-format prefs_test.go and re-verify all internal packages

Script
Now I'm checking the main.go file to locate where the auth handlers are set up and how the server dependencies are initialized.


Locate handler construction and Deps wiring in main.go

Script

View imports and Deps wiring context in main.go

Script

Add captcha, notes, prefs imports to main.go

main.go
+11
-8

Construct new handlers and wire them into server.Deps

main.go
+20
-11

Verify main.go formatting

Script
Now moving on to step 4, I need to check the current state of the frontend API file to see what edits are needed.


View current api.js for edits

Script

Add keepalive option to apiFetch for background flushes during page unload

api.js
+29
-24

Add captcha, prefs, and notes endpoint wrappers to Api object

api.js
+24
-4

Syntax-check api.js

Script
I need to locate and extract the original captcha script from the dump file so I can understand its structure before converting it to the new JavaScript module.


Read original captcha.html part 1 for porting

Script

Read original captcha.html part 2

Script

Read original captcha.html part 3 (Signals + puzzle logic)

Script

Read original captcha.html part 4 (Magic Solve, FPS, boot)

Script
I need to examine the common.js file to understand the requireAuth and renderAuthNav function signatures, and then cross-reference the nav structure and card patterns from activity.html.


Locate frontend files in dump

Script

Read common.js for requireAuth/renderAuthNav semantics

Script
I need to verify what Api.me() returns and check the shape of the /api/me endpoint response. I should also look at the activity.html and index.html files to understand the card and table patterns being used.


Check /api/me response shape

Script

Check userResponse JSON shape (need user id field name)

Script

Get userResponse JSON tags

Script
I'm looking at the user object structure with fields like id, username, display_name, roles, and totp_enabled, then checking how activity.html and index.html handle their card patterns and navigation to match the styling.


Read activity.html for nav and table structure

Script

Read original index.html for Get Started card fix

Script
I'm starting with the HTML shell for the captcha page, then moving into the JavaScript logic. I'm setting up a navigation bar that appears across all pages with links to home, notes, and captcha—with the authenticated routes hidden by default using data attributes.

For the captcha page itself, I'm building the layout with a header showing the target and performance chips, a stats row tracking streak and best times, the grid fieldset for the puzzle, a status line for feedback, and action buttons for verification and auto-solve. I'm moving the inline border styling for the ghost button into CSS to keep the markup clean, then adding a speed control slider and a reset button at the bottom.

The leaderboard will be a collapsible details element positioned as its own card below the main puzzle, showing the top 10 solvers plus the current user's ranked position. I also need to add the sync status chip next to the FPS indicator in the header.

Now I'm checking the existing style.css to see what utility classes are already defined so I can reference them properly and avoid duplication.

Show less
Done
Now I'll build the new captcha.html shell (CSP-clean, no inline code):


Diagnosing race condition between stopAutoSolve and performReset
Diagnosing race condition between stopAutoSolve and performReset
The <details> toggle event fires asynchronously, so my synchronous flag reset won't guard the programmatic restore. Fixing:


Reconciling note deletions and tracking pagination cursors
Reconciling note deletions and tracking pagination cursors

Fix restoringPref guard: toggle event is async, so consume flag inside the listener

page-captcha.js
+7
-5

Boot restore: set flag without synchronous reset (consumed by the async toggle handler)

page-captcha.js
+4
-5
I'm spotting a race condition in how stopAutoSolve interacts with performReset — when stopAutoSolve calls flushSync, it synchronously moves pending work into inflight, but then performReset immediately zeros out the pending state, creating a timing issue where the flush operation could race against the reset.

There's also a null reference hazard: if I set inflight = null in performReset while flushSync's request is still in flight and fails, the catch handler would crash trying to access properties on null.

The fix is to split the auto-solve stopping into two operations — one that just halts the RAF loop without flushing, and another that adds the flush on top — so performReset can stop the solver cleanly without triggering a flush that races the reset.

For the in-flight sync issue, I should track the flush promise at the module level and have performReset wait for any pending sync to complete before calling the reset API, ensuring the server state is stable before we wipe the client side. Refining the reset flow...

After the sync completes and applies server changes, I need to make sure we're not double-counting solved items since reset zeros out baseSolves right after. Also, flushSync will check hasWork() in its success path to decide whether to schedule another sync, but since reset just cleared everything, that check will correctly return false. I should remove the inflight = null assignment from the earlier reset code since flushSync now owns that lifecycle, then clean up the relevant sections.

Show less
I found a race condition in performReset where stopAutoSolve() triggers a flush that can corrupt state if pending/inflight are mutated during the request. I'll restructure this using a trackable flush promise to prevent the race.


Make flushSync return a trackable promise so reset can await in-flight syncs

page-captcha.js
+57
-47

Split raf-halt from flush so reset can stop the solver without triggering a sync

page-captcha.js
+18
-12

Rewrite performReset to drain in-flight syncs before deleting the server row

page-captcha.js
+39
-31
I'm checking the applyServer logic around signal resets — after a reset, bestSignal is 0, which is fine, and the Math.max comparison prevents the server value from lowering the local best mid-session. Both values reset to 0 together, so that's consistent. I also need to verify the 401 error handling path in doFlush where I'm returning inside the try block.

The finally block still executes properly, clearing syncInFlight and syncPromise. The visibility change handlers call flushSync with void, which works fine even though flushSync now returns a promise. There's also a potential ordering concern with haltAutoSolve and stopAutoSolve referencing flushSync, but function hoisting handles that. In performReset, I'm checking if (syncPromise) before awaiting it — since JavaScript is single-threaded, the promise reference can't become null between the check and the await, so even if the module variable gets nulled by the finally block, the awaited reference is already captured. Everything checks out.

Show less

Re-check page-captcha.js syntax

Script
Done
Now notes.html:


Architected auto-refreshing note feed with polling, reconciliation, and deferred rendering
Architected auto-refreshing note feed with polling, reconciliation, and deferred rendering

New notes.html: composer + auto-refreshing feed shell, login-only
notes.html
Looking at the page-notes.js polling mechanism, I'm setting up the auto-refresh with a 5-second poll interval that backs off up to 60 seconds, maintaining a map of notes by ID and their corresponding DOM elements while tracking the pagination cursor. The reconciliation logic handles deletions by comparing fetched notes against what's already loaded and removing any that are no longer present.

For new notes, I'm checking if they're from the current user to insert them immediately with highlighting, otherwise deferring them to a pending queue if the user has scrolled past the top. The DOM sync keeps notes strictly ordered by ID descending using a sorted array, and clicking the pending notes pill merges them back in and scrolls to the top. When loading older notes, I fetch with a cursor and hide the load button once we've reached the end of the feed.

The composer posts new notes to the top with a highlight, clears the input, and updates the character count. Delete buttons only appear on my own notes and trigger a confirmation before removal. Copy buttons use the clipboard API with a textarea fallback, showing a transient confirmation. I'm also handling visibility changes to pause polling when the tab is hidden and resume immediately when it comes back into focus, with exponential backoff for errors and a status pill showing the connection state. Timestamps use relative formatting like "just now" or "2h ago" with full dates as tooltips, and since polling only re-renders new notes, I'm refreshing the time labels on each poll tick by iterating through the node map and updating the text content.

Now I need to verify the note shape and field names from the backend to make sure my client-side handling matches what the server returns.

Show less

Confirm Note JSON field names and handler responses

Script

New page-notes.js: visibility-aware polling feed with keyed reconciliation, deferred-new pill, copy and delete
page-notes.js
Done

Your org is out of usage credits. We let your admin know. Switch models to continue chatting.
Switch models





Claude is AI and can make mistakes. Please double-check responses.


100
64
34


/
Claude finished the response
I have added a new /captcha page 
however, right now it is frontend only
as a demonstration of the website's ability to handle user information
we need to store this in the backend on our database
but also we need to add opentelemetry here as well 
also we need to have a leaderboard, but only for people who are signed in
on the same page, collapsed by default 
but we should remember the user setting 
if they expand it / collapse it again
this captcha page will only be available to people who are logged in
also the index.html is a little strange
it says 
Get Started 
Sign up / Log in 
even though I am logged in 
and the settings page correctly displays the logout button 
and even the index / home page clearly identifies 
earlier in the page that 
I am signed in as x
another big feature we can add is a new page where everyone can add a new short note
something like a microblog or a tweet, plain text only, no attachments allowed. 
and everyone can see everyone else's posts. 
there are no images, no image previews, no link previews, 
and even the links in posts are not actually clickable 
however, it should be very easy to copy paste any post 
also any user should be able to delete an old post but NOT edit them
you should not be able to delete any one else's posts 
this page should also be responsive and mobile friendly 
see what I changed with the /activity page for inspiration (basically cards)
also this page should somehow autorefresh 
and this might be the most difficult part of the whole operation
this feed is only available to people who are logged in
please document all architectural decisions 
and please update all readme as necessary as we continue to improve this application 
please return full files as well as full path for all files that need to change

Show more

Architected backend persistence, leaderboard system, notes microblog, and frontend integration
Architected backend persistence, leaderboard system, notes microblog, and frontend integration
Go 1.22 (fine for gofmt -e syntax validation) and Node 22 are available. I have the full picture now. Let me set up a workspace mirroring the repo and build all the changes, validating as I go.


Architected backend modifications for captcha storage and telemetry integration
Architected backend modifications for captcha storage and telemetry integration
All hashes match the dump metadata byte-for-byte, so I can make surgical edits on exact originals. Let me apply the store changes.


Rectified directory structure issue and verified migration files persisted
Rectified directory structure issue and verified migration files persisted
Now the store edits — new types first, then the three new data-access sections.


Prepared to integrate DecodeJSON helper into middleware file
Prepared to integrate DecodeJSON helper into middleware file
Now the three new backend packages. First internal/captcha:


Created 2 files
Created 2 files
Now internal/notes:


Created 2 files
Created 2 files
Now internal/prefs:


Scoping mobile table styles to preserve leaderboard formatting
Scoping mobile table styles to preserve leaderboard formatting
Now I'll build the new captcha.html shell (CSP-clean, no inline code):


Diagnosing race condition between stopAutoSolve and performReset
Diagnosing race condition between stopAutoSolve and performReset
The <details> toggle event fires asynchronously, so my synchronous flag reset won't guard the programmatic restore. Fixing:


Reconciling note deletions and tracking pagination cursors
Reconciling note deletions and tracking pagination cursors
Now notes.html:


Architected auto-refreshing note feed with polling, reconciliation, and deferred rendering
Architected auto-refreshing note feed with polling, reconciliation, and deferred rendering
please continue


Refactoring HTML navigation and augmenting feature demonstrations
Refactoring HTML navigation and augmenting feature demonstrations
Now the style.css work — first, scope the mobile table→card transform to the activity table only so the new leaderboard table isn't mangled:


Scope mobile table-to-card transform to .activity-table so other tables (leaderboard) stay tabular
Scope mobile table-to-card transform to .activity-table so other tables (leaderboard) stay tabular
Now append all the new component styles (former captcha inline styles + notes + leaderboard + utilities):


Verifying code compilation across multiple file types
Verifying code compilation across multiple file types
The code records current_streak in the histogram, not best — fixing the doc to match:


Verifying API signatures and store method implementations
Verifying API signatures and store method implementations
Now a full review pass on the Go packages — checking interface alignment between handlers and store:


Verifying file integrity and cross-referencing page-captcha.js
Verifying file integrity and cross-referencing page-captcha.js
All logic tests pass against the verbatim-extracted functions. Final validation sweep — formatting, JS syntax, and HTML↔JS↔CSS cross-references:


Reconciling file size constraints with established response protocols
Reconciling file size constraints with established response protocols
All cross-references check out. Let me take a final inventory and sizes before emitting everything:


Deliberating response format for voluminous file content
Deliberating response format for voluminous file content
Now let me read the final state of every file I haven't fully viewed end-to-end, so the inline listing is byte-faithful:


Orchestrated comprehensive delivery of twenty-seven validated project files
Orchestrated comprehensive delivery of twenty-seven validated project files
Everything is built, validated (gofmt clean, node --check clean, logic tests pass against the verbatim-extracted validation/clamping functions, HTML/CSS/JS cross-references all resolve), staged in the zip above, and listed in full below.

One housekeeping note first: the GitHub PAT wasn't needed — this change adds zero new dependencies (stdlib + the OTel/pgx modules you already have), so I never touched the GitHub API and the token appears nowhere in these files. Since it was pasted in plaintext into a chat, I'd still recommend revoking it when convenient.

What this delivers
CAPTCHA backend — new captcha_stats table (one aggregate row per user, never per-solve rows: Magic Solve does hundreds of solves/sec, so the browser batches deltas and POST /api/captcha/sync folds them in with a single UPSERT; best_streak merges via GREATEST, current_streak is last-write-wins so streaks resume across devices). Server clamps all client-reported numbers; DB CHECK constraints back it up. Flushes happen every ~4s, on solver stop, and via fetch(keepalive:true) on tab-hide/unload (sendBeacon can't carry the Authorization header). OTel domain metrics + span attributes throughout.
Leaderboard — RANK() OVER (best_streak DESC, total_solves DESC, updated_at ASC) in a shared CTE so top-10 and "your rank" can never disagree; collapsed <details> on the same page; its open/closed state is a server-side preference (captcha.leaderboard.open) via a new generic GET/PUT /api/prefs/{key} endpoint. The dividing line, documented in ARCHITECTURE.md: account settings live server-side, device settings (Magic Solve speed, which depends on your display's refresh rate) stay in localStorage. "Clear my local data" became "Reset my stats" — an honest POST /api/captcha/reset that deletes your server row and removes you from the leaderboard.
Login-only CAPTCHA page — same gating model as /activity: the static shell is public (Caddy), requireAuth() redirects anonymous visitors, and every API route enforces RequireAuth server-side. Also: your captcha.html had large inline <style>/<script>/style="" blocks, which violate the project's strict self-only CSP (they'd generate report-only violations today and break under enforcement). I moved everything into style.css and a new js/page-captcha.js, so the page is CSP-clean like the rest of the app.
index.html fix — the Get Started card simply lacked data-auth attributes; page-index.js already calls renderAuthNav(), so wrapping the signup/login block in data-auth="out" and adding a signed-in variant (data-auth="in": Notes / CAPTCHA / Settings buttons) fixes it with zero JS changes.
Notes microblog — notes table (bigint identity, ON DELETE CASCADE — content leaves with the account, unlike activity_log's audit-preserving SET NULL). Plain text enforced on both sides: server normalizes CRLF→LF, rejects control chars, counts 1–500 runes (DB char_length CHECK matches); client renders bodies only via textContent, so URLs are inert text — the per-note Copy button exists precisely because links don't work. Delete-never-edit; ownership enforced inside the single DELETE ... WHERE id AND user_id statement with a uniform 404 (no existence oracle). Creation is rate-limited (0.5 rps, burst 5) keyed by the authenticated user id inside RequireAuth.
Auto-refresh — deliberate polling, not SSE/WebSockets: your WriteTimeout: 30s kills long-lived streams, they fight otelhttp spans, and tunnels buffer them. The feed polls every 5s only while visible, refreshes instantly on tab return, backs off to 60s on errors, and reconciles each response keyed by note id: deletions inside the newest-50 window vanish live, your own posts appear immediately, others' new posts appear immediately unless you've scrolled down — then a "N new notes ↑" pill defers them so the feed never jumps under your thumb. Cursor-based "Load older" (?before=<id>, stable under concurrent inserts, unlike OFFSET). Cards, responsive, mobile-first like /activity — and the existing mobile table→card CSS transform is now scoped to .activity-table so the leaderboard table isn't mangled on phones.
No new environment variables — docs/CONFIGURATION.md and .env.example need no changes. Migrations 0006/0007 follow your existing 0005. Full go build wasn't possible here (no module proxy — the same constraint behind the uncommitted go.sum), so run ./scripts/test.sh all after committing.

migrations/0006_captcha_and_prefs.up.sql
sql
-- 0006_captcha_and_prefs.up.sql
-- Per-user CAPTCHA game statistics and a small generic user-preferences store.
--
-- captcha_stats is ONE aggregate row per user, not one row per solve. The
-- page's auto-solver ("Magic Solve") can complete hundreds of puzzles per
-- second at maximum speed, so per-solve rows (and per-solve HTTP requests)
-- would grow unboundedly for zero analytical value. The client batches deltas
-- and the API folds them in atomically; storage stays O(users).

CREATE TABLE captcha_stats (
    user_id        uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    best_streak    bigint NOT NULL DEFAULT 0 CHECK (best_streak >= 0),
    current_streak bigint NOT NULL DEFAULT 0 CHECK (current_streak >= 0),
    total_solves   bigint NOT NULL DEFAULT 0 CHECK (total_solves >= 0),
    manual_solves  bigint NOT NULL DEFAULT 0 CHECK (manual_solves >= 0),
    auto_solves    bigint NOT NULL DEFAULT 0 CHECK (auto_solves >= 0),
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

-- Supports the leaderboard ordering (best streak, then volume).
CREATE INDEX captcha_stats_leaderboard_idx
    ON captcha_stats (best_streak DESC, total_solves DESC);

-- user_prefs holds small per-user UI preferences (for example whether the
-- CAPTCHA leaderboard is expanded). Keys are constrained to a conservative
-- charset in the application layer; values are short opaque strings. User
-- preferences live server-side so they follow the account across devices;
-- device-specific settings (like the Magic Solve speed, which depends on the
-- device's refresh rate) stay in the browser's localStorage instead.
CREATE TABLE user_prefs (
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key        text NOT NULL CHECK (char_length(key) BETWEEN 1 AND 64),
    value      text NOT NULL DEFAULT '' CHECK (octet_length(value) <= 4096),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, key)
);
migrations/0006_captcha_and_prefs.down.sql
sql
-- 0006_captcha_and_prefs.down.sql
DROP TABLE IF EXISTS user_prefs;
DROP TABLE IF EXISTS captcha_stats;
migrations/0007_notes.up.sql
sql
-- 0007_notes.up.sql
-- Notes: a deliberately plain public microblog. Plain text only (rendering is
-- done with textContent on the client, so nothing here is ever interpreted as
-- HTML), no attachments, no edits. Deletion is a hard DELETE: consistent with
-- this project's privacy posture, "deleted" means the row is gone, not
-- flagged. Rows cascade away if the author's account is ever removed
-- (contrast activity_log, which keeps audit rows with user_id set NULL).
CREATE TABLE notes (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    -- char_length counts characters (code points), matching the API's rune
    -- count. The API validates first; this is defence in depth.
    CONSTRAINT notes_body_len CHECK (char_length(body) BETWEEN 1 AND 500)
);

-- The feed reads newest-first by id (the PK index already serves that);
-- this index serves per-user lookups and the ownership-checked delete.
CREATE INDEX notes_user_id_idx ON notes (user_id);
migrations/0007_notes.down.sql
sql
-- 0007_notes.down.sql
DROP TABLE IF EXISTS notes;
internal/captcha/captcha.go
go
// Package captcha persists per-user statistics for the CAPTCHA demo page and
// serves its leaderboard.
//
// Design: the page's auto-solver can complete hundreds of puzzles per second,
// so the client never reports individual solves. It accumulates deltas and
// POSTs a small batch every few seconds (and on page hide); the server folds
// each batch into a single aggregate row per user. Totals only grow, the best
// streak is merged with GREATEST so it can never regress, and the current
// streak is a last-write-wins snapshot. Storage is O(users), request volume
// is O(seconds), and the numbers are honest enough for a demo leaderboard.
//
// Everything here is instrumented with OpenTelemetry: otelhttp (wired in the
// server package) already produces a span and HTTP metrics per request; this
// package adds domain metrics (solve counters by mode, sync counter, a streak
// histogram) and enriches the request span with attributes.
package captcha

import (
	"errors"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/kusl/GoTunnels/internal/auth"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/store"
)

// scopeName is the instrumentation scope for telemetry emitted here.
const scopeName = "github.com/kusl/GoTunnels/internal/captcha"

const (
	// maxDeltaPerSync bounds how many solves one batch may claim. At the
	// solver's hardware ceiling (~one grid per animation frame) a 4-second
	// batch tops out around a thousand solves, so this is generous headroom
	// while still keeping a single hostile request from minting absurd totals.
	maxDeltaPerSync = 100_000
	// maxStreak bounds reported streak snapshots.
	maxStreak = 1_000_000_000
	// leaderboardSize is how many rows the leaderboard endpoint returns.
	leaderboardSize = 10
)

// Handlers bundles dependencies for the CAPTCHA endpoints.
type Handlers struct {
	store *store.Store
	log   *slog.Logger

	solves  metric.Int64Counter
	syncs   metric.Int64Counter
	streaks metric.Int64Histogram
}

// NewHandlers builds the handler set and registers its OTel instruments. When
// telemetry is disabled the global meter is a no-op, so this stays free.
func NewHandlers(s *store.Store, log *slog.Logger) *Handlers {
	h := &Handlers{store: s, log: log}
	m := otel.Meter(scopeName)
	var err error
	if h.solves, err = m.Int64Counter("gotunnels.captcha.solves",
		metric.WithDescription("CAPTCHA solves folded into stats, by mode"),
		metric.WithUnit("{solve}")); err != nil {
		log.Warn("captcha: register solves counter", slog.String("error", err.Error()))
	}
	if h.syncs, err = m.Int64Counter("gotunnels.captcha.syncs",
		metric.WithDescription("CAPTCHA stat sync batches accepted"),
		metric.WithUnit("{batch}")); err != nil {
		log.Warn("captcha: register syncs counter", slog.String("error", err.Error()))
	}
	if h.streaks, err = m.Int64Histogram("gotunnels.captcha.streak",
		metric.WithDescription("Current streak reported at each sync"),
		metric.WithUnit("{solve}")); err != nil {
		log.Warn("captcha: register streak histogram", slog.String("error", err.Error()))
	}
	return h
}

// syncRequest is the client's batched progress report.
type syncRequest struct {
	ManualDelta   int64 `json:"manual_delta"`
	AutoDelta     int64 `json:"auto_delta"`
	CurrentStreak int64 `json:"current_streak"`
	BestStreak    int64 `json:"best_streak"`
}

// Stats returns the caller's aggregate stats. A user who has never synced gets
// an all-zero record rather than a 404, so the page has one uniform boot path.
func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	st, err := h.store.GetCaptchaStats(r.Context(), user.ID)
	if errors.Is(err, store.ErrNotFound) {
		st, err = store.CaptchaStats{UserID: user.ID}, nil
	}
	if err != nil {
		h.serverError(w, r, "captcha: get stats", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"stats": st})
}

// Sync folds one batched delta into the caller's stats and returns the updated
// aggregate so the client can reconcile (for example, totals accumulated from
// another device).
func (h *Handlers) Sync(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var req syncRequest
	if !httpx.DecodeJSON(w, r, &req, 4<<10) {
		return
	}
	in := clampSync(store.CaptchaSyncInput{
		ManualDelta:   req.ManualDelta,
		AutoDelta:     req.AutoDelta,
		CurrentStreak: req.CurrentStreak,
		BestStreak:    req.BestStreak,
	})

	st, err := h.store.SyncCaptchaStats(r.Context(), user.ID, in)
	if err != nil {
		h.serverError(w, r, "captcha: sync stats", err)
		return
	}

	// Domain telemetry: counters by mode, batch counter, streak histogram,
	// and attributes on the request span otelhttp already opened.
	ctx := r.Context()
	if h.solves != nil {
		if in.ManualDelta > 0 {
			h.solves.Add(ctx, in.ManualDelta, metric.WithAttributes(attribute.String("mode", "manual")))
		}
		if in.AutoDelta > 0 {
			h.solves.Add(ctx, in.AutoDelta, metric.WithAttributes(attribute.String("mode", "auto")))
		}
	}
	if h.syncs != nil {
		h.syncs.Add(ctx, 1)
	}
	if h.streaks != nil {
		h.streaks.Record(ctx, in.CurrentStreak)
	}
	trace.SpanFromContext(ctx).SetAttributes(
		attribute.Int64("captcha.manual_delta", in.ManualDelta),
		attribute.Int64("captcha.auto_delta", in.AutoDelta),
		attribute.Int64("captcha.current_streak", in.CurrentStreak),
		attribute.Int64("captcha.best_streak", st.BestStreak),
	)

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"stats": st})
}

// Reset deletes the caller's stats row entirely (they leave the leaderboard
// until they play again).
func (h *Handlers) Reset(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if err := h.store.DeleteCaptchaStats(r.Context(), user.ID); err != nil {
		h.serverError(w, r, "captcha: reset stats", err)
		return
	}
	trace.SpanFromContext(r.Context()).AddEvent("captcha.stats_reset")
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}

// Leaderboard returns the top players plus the caller's own ranked row (which
// may sit outside the top slice, or be absent if they have never played).
func (h *Handlers) Leaderboard(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	rows, err := h.store.CaptchaLeaderboard(r.Context(), leaderboardSize)
	if err != nil {
		h.serverError(w, r, "captcha: leaderboard", err)
		return
	}
	var mine *store.CaptchaLeaderboardRow
	if me, err := h.store.CaptchaRank(r.Context(), user.ID); err == nil {
		mine = &me
	} else if !errors.Is(err, store.ErrNotFound) {
		h.serverError(w, r, "captcha: own rank", err)
		return
	}
	if rows == nil {
		rows = []store.CaptchaLeaderboardRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"leaderboard": rows, "me": mine})
}

// clampSync bounds every client-supplied number: deltas are non-negative and
// capped, streak snapshots are non-negative and capped, and best is raised to
// at least current so a single batch is internally consistent.
func clampSync(in store.CaptchaSyncInput) store.CaptchaSyncInput {
	in.ManualDelta = clamp(in.ManualDelta, 0, maxDeltaPerSync)
	in.AutoDelta = clamp(in.AutoDelta, 0, maxDeltaPerSync)
	in.CurrentStreak = clamp(in.CurrentStreak, 0, maxStreak)
	in.BestStreak = clamp(in.BestStreak, 0, maxStreak)
	if in.BestStreak < in.CurrentStreak {
		in.BestStreak = in.CurrentStreak
	}
	return in
}

func clamp(v, lo, hi int64) int64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (h *Handlers) serverError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	h.log.ErrorContext(r.Context(), msg,
		slog.String("error", err.Error()),
		slog.String("request_id", httpx.RequestIDFromContext(r.Context())),
	)
	httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
}
internal/captcha/captcha_test.go
go
package captcha

import (
	"testing"

	"github.com/kusl/GoTunnels/internal/store"
)

func TestClampSync(t *testing.T) {
	cases := []struct {
		name string
		in   store.CaptchaSyncInput
		want store.CaptchaSyncInput
	}{
		{
			name: "passthrough",
			in:   store.CaptchaSyncInput{ManualDelta: 3, AutoDelta: 40, CurrentStreak: 7, BestStreak: 12},
			want: store.CaptchaSyncInput{ManualDelta: 3, AutoDelta: 40, CurrentStreak: 7, BestStreak: 12},
		},
		{
			name: "negatives floor to zero",
			in:   store.CaptchaSyncInput{ManualDelta: -1, AutoDelta: -99, CurrentStreak: -5, BestStreak: -5},
			want: store.CaptchaSyncInput{},
		},
		{
			name: "deltas capped",
			in:   store.CaptchaSyncInput{ManualDelta: maxDeltaPerSync + 1, AutoDelta: 1 << 60},
			want: store.CaptchaSyncInput{ManualDelta: maxDeltaPerSync, AutoDelta: maxDeltaPerSync},
		},
		{
			name: "streaks capped",
			in:   store.CaptchaSyncInput{CurrentStreak: maxStreak + 5, BestStreak: maxStreak + 5},
			want: store.CaptchaSyncInput{CurrentStreak: maxStreak, BestStreak: maxStreak},
		},
		{
			name: "best raised to at least current",
			in:   store.CaptchaSyncInput{CurrentStreak: 9, BestStreak: 2},
			want: store.CaptchaSyncInput{CurrentStreak: 9, BestStreak: 9},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := clampSync(tc.in); got != tc.want {
				t.Fatalf("clampSync(%+v) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	if got := clamp(5, 0, 10); got != 5 {
		t.Fatalf("clamp mid = %d", got)
	}
	if got := clamp(-5, 0, 10); got != 0 {
		t.Fatalf("clamp low = %d", got)
	}
	if got := clamp(50, 0, 10); got != 10 {
		t.Fatalf("clamp high = %d", got)
	}
}
internal/notes/notes.go
go
// Package notes implements a deliberately plain public microblog: short
// plain-text posts, visible to every signed-in user, hard-deletable only by
// their author, and never editable.
//
// "Plain text" is enforced on both sides of the wire. The server normalises
// line endings and rejects control characters (other than newline and tab) so
// stored bodies are clean, copyable text; the client renders bodies with
// textContent, so nothing is ever parsed as HTML and URLs stay inert strings.
// There is no edit endpoint on purpose — a post is either exactly what its
// author wrote, or gone.
package notes

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/kusl/GoTunnels/internal/auth"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/store"
)

// scopeName is the instrumentation scope for telemetry emitted here.
const scopeName = "github.com/kusl/GoTunnels/internal/notes"

const (
	// MaxBodyChars is the maximum note length in characters (code points),
	// mirrored by the notes_body_len CHECK constraint in the database.
	MaxBodyChars = 500
	// defaultPageSize is how many notes one list request returns by default.
	defaultPageSize = 50
	// maxPageSize caps a caller-supplied limit.
	maxPageSize = 200
)

// Handlers bundles dependencies for the notes endpoints.
type Handlers struct {
	store *store.Store
	log   *slog.Logger

	created metric.Int64Counter
	deleted metric.Int64Counter
}

// NewHandlers builds the handler set and registers its OTel instruments.
func NewHandlers(s *store.Store, log *slog.Logger) *Handlers {
	h := &Handlers{store: s, log: log}
	m := otel.Meter(scopeName)
	var err error
	if h.created, err = m.Int64Counter("gotunnels.notes.created",
		metric.WithDescription("Notes created"),
		metric.WithUnit("{note}")); err != nil {
		log.Warn("notes: register created counter", slog.String("error", err.Error()))
	}
	if h.deleted, err = m.Int64Counter("gotunnels.notes.deleted",
		metric.WithDescription("Notes deleted by their author"),
		metric.WithUnit("{note}")); err != nil {
		log.Warn("notes: register deleted counter", slog.String("error", err.Error()))
	}
	return h
}

type createRequest struct {
	Body string `json:"body"`
}

// List returns notes newest-first: the latest page by default, or — with
// ?before=<id> — the page strictly older than that id ("load older"). The
// same endpoint serves the page's poll-based auto-refresh, which simply
// re-fetches the newest window and reconciles client-side.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.CurrentUser(r.Context()); !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var beforeID int64
	if v := r.URL.Query().Get("before"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < 0 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid before cursor")
			return
		}
		beforeID = n
	}
	limit := defaultPageSize
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxPageSize {
			httpx.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = n
	}

	rows, err := h.store.ListNotes(r.Context(), beforeID, limit)
	if err != nil {
		h.serverError(w, r, "notes: list", err)
		return
	}
	if rows == nil {
		rows = []store.Note{}
	}
	trace.SpanFromContext(r.Context()).SetAttributes(
		attribute.Int("notes.returned", len(rows)),
		attribute.Int64("notes.before", beforeID),
	)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"notes": rows})
}

// Create posts a new note for the caller.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var req createRequest
	if !httpx.DecodeJSON(w, r, &req, 64<<10) {
		return
	}
	body, problem := ValidateBody(req.Body)
	if problem != "" {
		httpx.WriteError(w, http.StatusBadRequest, problem)
		return
	}

	n, err := h.store.CreateNote(r.Context(), user.ID, body)
	if err != nil {
		h.serverError(w, r, "notes: create", err)
		return
	}
	if h.created != nil {
		h.created.Add(r.Context(), 1)
	}
	trace.SpanFromContext(r.Context()).SetAttributes(
		attribute.Int64("note.id", n.ID),
		attribute.Int("note.chars", utf8.RuneCountInString(n.Body)),
	)
	httpx.WriteJSON(w, http.StatusCreated, map[string]any{"note": n})
}

// Delete removes the caller's own note. Ownership is enforced in the store's
// single DELETE statement; both "no such note" and "someone else's note"
// uniformly produce 404 so the endpoint is not an existence oracle.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}
	removed, err := h.store.DeleteNote(r.Context(), id, user.ID)
	if err != nil {
		h.serverError(w, r, "notes: delete", err)
		return
	}
	if !removed {
		httpx.WriteError(w, http.StatusNotFound, "note not found")
		return
	}
	if h.deleted != nil {
		h.deleted.Add(r.Context(), 1)
	}
	trace.SpanFromContext(r.Context()).SetAttributes(attribute.Int64("note.id", id))
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ValidateBody normalises and validates a raw note body. It returns the
// cleaned body and an empty string on success, or a zero body and a
// human-readable problem on failure.
//
// Normalisation: CRLF and lone CR become LF, then surrounding whitespace is
// trimmed. Validation: valid UTF-8, 1..MaxBodyChars characters, and no control
// characters other than newline and tab (keeps posts as honest, copyable plain
// text — no zero-width or terminal-escape mischief).
func ValidateBody(raw string) (string, string) {
	if !utf8.ValidString(raw) {
		return "", "note body must be valid UTF-8"
	}
	body := strings.ReplaceAll(raw, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")
	body = strings.TrimSpace(body)
	if body == "" {
		return "", "note body must not be empty"
	}
	if n := utf8.RuneCountInString(body); n > MaxBodyChars {
		return "", fmt.Sprintf("note body must be at most %d characters (got %d)", MaxBodyChars, n)
	}
	for _, r := range body {
		if r < 0x20 && r != '\n' && r != '\t' {
			return "", "note body must not contain control characters"
		}
		if r == 0x7f {
			return "", "note body must not contain control characters"
		}
	}
	return body, ""
}

func (h *Handlers) serverError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	h.log.ErrorContext(r.Context(), msg,
		slog.String("error", err.Error()),
		slog.String("request_id", httpx.RequestIDFromContext(r.Context())),
	)
	httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
}
internal/notes/notes_test.go
go
package notes

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestValidateBodyAccepts(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello world", "hello world"},
		{"trims surrounding whitespace", "  hi there \n", "hi there"},
		{"keeps interior newlines", "line one\nline two", "line one\nline two"},
		{"normalises crlf", "a\r\nb", "a\nb"},
		{"normalises lone cr", "a\rb", "a\nb"},
		{"keeps tabs", "col1\tcol2", "col1\tcol2"},
		{"multibyte ok", "héllø 世界 🚀", "héllø 世界 🚀"},
		{"exactly max runes", strings.Repeat("界", MaxBodyChars), strings.Repeat("界", MaxBodyChars)},
		{"url stays plain text", "see https://example.com for more", "see https://example.com for more"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, problem := ValidateBody(tc.in)
			if problem != "" {
				t.Fatalf("ValidateBody(%q) rejected: %s", tc.in, problem)
			}
			if got != tc.want {
				t.Fatalf("ValidateBody(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestValidateBodyRejects(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"whitespace only", "   \n\t  "},
		{"over max runes", strings.Repeat("a", MaxBodyChars+1)},
		{"over max multibyte runes", strings.Repeat("界", MaxBodyChars+1)},
		{"escape sequence", "sneaky \x1b[31m red"},
		{"null byte", "a\x00b"},
		{"delete char", "a\x7fb"},
		{"invalid utf8", string([]byte{0xff, 0xfe, 0x41})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got, problem := ValidateBody(tc.in); problem == "" {
				t.Fatalf("ValidateBody(%q) accepted as %q, want rejection", tc.in, got)
			}
		})
	}
}

func TestValidateBodyMaxCountsRunesNotBytes(t *testing.T) {
	// 500 multibyte runes is far more than 500 bytes but must be accepted.
	in := strings.Repeat("é", MaxBodyChars)
	if utf8.RuneCountInString(in) != MaxBodyChars {
		t.Fatal("test setup wrong")
	}
	if len(in) <= MaxBodyChars {
		t.Fatal("test setup wrong: want multibyte input")
	}
	if _, problem := ValidateBody(in); problem != "" {
		t.Fatalf("multibyte body at limit rejected: %s", problem)
	}
}
internal/prefs/prefs.go
go
// Package prefs is a tiny per-user key/value preference store, exposed as
// GET/PUT /api/prefs/{key}. It exists so small UI choices (like whether the
// CAPTCHA leaderboard is expanded) follow the *account* rather than the
// browser. The dividing line used across the app: user preferences live here,
// on the server; device-specific settings (like the Magic Solve speed, which
// depends on the device's refresh rate) stay in localStorage.
//
// Keys are constrained to a conservative charset and values to a small size,
// so this can never quietly become a document store.
package prefs

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/kusl/GoTunnels/internal/auth"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/store"
)

const (
	// MaxKeyLen is the maximum preference key length.
	MaxKeyLen = 64
	// MaxValueBytes is the maximum preference value size, mirrored by the
	// CHECK constraint on user_prefs.value.
	MaxValueBytes = 4096
)

// Handlers bundles dependencies for the preference endpoints.
type Handlers struct {
	store *store.Store
	log   *slog.Logger
}

// NewHandlers builds the handler set.
func NewHandlers(s *store.Store, log *slog.Logger) *Handlers {
	return &Handlers{store: s, log: log}
}

type setRequest struct {
	Value string `json:"value"`
}

// Get returns the caller's stored value for a key. A missing key is a normal
// state, not an error: the response carries "exists": false so clients can
// fall back to defaults without special-casing 404s.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	key := r.PathValue("key")
	if !ValidKey(key) {
		httpx.WriteError(w, http.StatusBadRequest, "invalid preference key")
		return
	}
	value, err := h.store.GetUserPref(r.Context(), user.ID, key)
	if errors.Is(err, store.ErrNotFound) {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"key": key, "value": "", "exists": false})
		return
	}
	if err != nil {
		h.serverError(w, r, "prefs: get", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"key": key, "value": value, "exists": true})
}

// Set upserts the caller's value for a key.
func (h *Handlers) Set(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	key := r.PathValue("key")
	if !ValidKey(key) {
		httpx.WriteError(w, http.StatusBadRequest, "invalid preference key")
		return
	}
	var req setRequest
	if !httpx.DecodeJSON(w, r, &req, 16<<10) {
		return
	}
	if len(req.Value) > MaxValueBytes {
		httpx.WriteError(w, http.StatusBadRequest, "preference value too large")
		return
	}
	if err := h.store.SetUserPref(r.Context(), user.ID, key, req.Value); err != nil {
		h.serverError(w, r, "prefs: set", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

// ValidKey enforces the preference-key policy: 1..MaxKeyLen characters from
// [a-z0-9._-], and it must start with a letter or digit. Kept as a manual loop
// (matching validUsername in the auth package) rather than a regexp.
func ValidKey(k string) bool {
	if len(k) < 1 || len(k) > MaxKeyLen {
		return false
	}
	for i, r := range k {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '.' || r == '-' || r == '_':
			if i == 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func (h *Handlers) serverError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	h.log.ErrorContext(r.Context(), msg,
		slog.String("error", err.Error()),
		slog.String("request_id", httpx.RequestIDFromContext(r.Context())),
	)
	httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
}
internal/prefs/prefs_test.go
go
package prefs

import "testing"

func TestValidKey(t *testing.T) {
	valid := []string{
		"captcha.leaderboard.open",
		"theme",
		"a",
		"notes.feed-order",
		"k1_2-3.four",
	}
	for _, k := range valid {
		if !ValidKey(k) {
			t.Errorf("ValidKey(%q) = false, want true", k)
		}
	}

	invalid := []string{
		"",                                // empty
		".starts-with-dot",                // must start alphanumeric
		"-starts-with-dash",               // must start alphanumeric
		"_starts_with_under",              // must start alphanumeric
		"Has.Uppercase",                   // uppercase not allowed
		"has space",                       // whitespace not allowed
		"has/slash",                       // slash not allowed
		"emoji.😀",                         // non-ASCII not allowed
		"trailing.newline\n",              // control char
		string(make([]byte, MaxKeyLen+1)), // too long
	}
	for _, k := range invalid {
		if ValidKey(k) {
			t.Errorf("ValidKey(%q) = true, want false", k)
		}
	}
}
internal/store/store.go
go
// Package store is the single data-access layer. Every SQL query in the
// application lives here, so the rest of the code deals in Go types and never
// in SQL strings. It wraps a pgx connection pool.
//
// UUID handling note: to avoid pulling in a separate UUID dependency, user ids
// are exchanged as plain strings. Queries always select `id::text` and always
// bind uuid parameters with an explicit `$N::uuid` cast, which sidesteps the
// lack of an implicit text->uuid cast in PostgreSQL.
package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a lookup finds no row.
var ErrNotFound = errors.New("store: not found")

// Store is the data-access facade over a pgx pool.
type Store struct {
	pool *pgxpool.Pool
}

// New wraps a pool.
func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// Pool exposes the underlying pool for health checks.
func (s *Store) Pool() *pgxpool.Pool { return s.pool }

// Ping verifies database connectivity.
func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// User is an application user.
type User struct {
	ID          string
	Username    string
	DisplayName string
	CreatedAt   time.Time
}

// Session is a server-side session record.
type Session struct {
	ID         string
	UserID     string
	AuthMethod string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
}

// Flow is an in-progress WebAuthn ceremony's stored state.
type Flow struct {
	ID          string
	UserID      *string
	Kind        string
	SessionData []byte
	ExpiresAt   time.Time
}

// Activity is one audit-log row.
type Activity struct {
	ID         int64           `json:"id"`
	UserID     *string         `json:"user_id,omitempty"`
	Username   string          `json:"username"`
	EventType  string          `json:"event_type"`
	AuthMethod string          `json:"auth_method"`
	Outcome    string          `json:"outcome"`
	IPHash     string          `json:"ip_hash"`
	UserAgent  string          `json:"user_agent,omitempty"`
	Detail     json.RawMessage `json:"detail,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// ActivityInput is the payload for recording an activity event.
type ActivityInput struct {
	UserID     *string
	Username   string
	EventType  string
	AuthMethod string
	Outcome    string
	IPHash     string
	UserAgent  string
	Detail     map[string]any
}

// CSPReportInput is a normalised CSP violation ready to persist.
type CSPReportInput struct {
	DocumentURI        string
	Referrer           string
	BlockedURI         string
	ViolatedDirective  string
	EffectiveDirective string
	OriginalPolicy     string
	Disposition        string
	SourceFile         string
	LineNumber         int
	ColumnNumber       int
	StatusCode         int
	ScriptSample       string
	IPHash             string
	UserAgent          string
	Raw                json.RawMessage
}

// CaptchaStats is a user's aggregate CAPTCHA game record. One row per user;
// solves are folded in as batched deltas, never stored individually.
type CaptchaStats struct {
	UserID        string    `json:"user_id,omitempty"`
	BestStreak    int64     `json:"best_streak"`
	CurrentStreak int64     `json:"current_streak"`
	TotalSolves   int64     `json:"total_solves"`
	ManualSolves  int64     `json:"manual_solves"`
	AutoSolves    int64     `json:"auto_solves"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CaptchaSyncInput is one client-side batch of CAPTCHA progress. Deltas are
// added to totals; streaks are point-in-time snapshots (best is merged with
// GREATEST, current is last-write-wins).
type CaptchaSyncInput struct {
	ManualDelta   int64
	AutoDelta     int64
	CurrentStreak int64
	BestStreak    int64
}

// CaptchaLeaderboardRow is one ranked leaderboard entry.
type CaptchaLeaderboardRow struct {
	Rank        int64  `json:"rank"`
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	BestStreak  int64  `json:"best_streak"`
	TotalSolves int64  `json:"total_solves"`
}

// Note is one public microblog post.
type Note struct {
	ID          int64     `json:"id"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
}

// ---------------------------------------------------------------------------
// Users & roles
// ---------------------------------------------------------------------------

// CreateUser inserts a user and grants the default "user" role atomically.
func (s *Store) CreateUser(ctx context.Context, username, displayName string) (User, error) {
	lower := normalizeUsername(username)
	if displayName == "" {
		displayName = username
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var u User
	err = tx.QueryRow(ctx, `
		INSERT INTO users (username, username_lower, display_name)
		VALUES ($1, $2, $3)
		RETURNING id::text, username, display_name, created_at`,
		username, lower, displayName,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.CreatedAt)
	if err != nil {
		return User{}, err
	}
	if _, err = tx.Exec(ctx,
		`INSERT INTO user_roles (user_id, role) VALUES ($1::uuid, 'user')`, u.ID); err != nil {
		return User{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return u, nil
}

// GetUserByUsername looks up a user case-insensitively.
func (s *Store) GetUserByUsername(ctx context.Context, username string) (User, error) {
	lower := normalizeUsername(username)
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, username, display_name, created_at
		FROM users WHERE username_lower = $1`, lower,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.CreatedAt)
	return u, mapErr(err)
}

// GetUserByID looks up a user by id.
func (s *Store) GetUserByID(ctx context.Context, id string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, username, display_name, created_at
		FROM users WHERE id = $1::uuid`, id,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.CreatedAt)
	return u, mapErr(err)
}

// UsernameExists reports whether a username is already taken.
func (s *Store) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE username_lower = $1)`,
		normalizeUsername(username),
	).Scan(&exists)
	return exists, err
}

// UserRoles returns the role names granted to a user.
func (s *Store) UserRoles(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT role FROM user_roles WHERE user_id = $1::uuid ORDER BY role`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, rows.Err()
}

// ---------------------------------------------------------------------------
// Password credentials
// ---------------------------------------------------------------------------

// SetPassword upserts the password hash for a user.
func (s *Store) SetPassword(ctx context.Context, userID, hash string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO password_credentials (user_id, password_hash, updated_at)
		VALUES ($1::uuid, $2, now())
		ON CONFLICT (user_id)
		DO UPDATE SET password_hash = EXCLUDED.password_hash, updated_at = now()`,
		userID, hash)
	return err
}

// GetPasswordHash returns the stored PHC hash, or ErrNotFound.
func (s *Store) GetPasswordHash(ctx context.Context, userID string) (string, error) {
	var hash string
	err := s.pool.QueryRow(ctx,
		`SELECT password_hash FROM password_credentials WHERE user_id = $1::uuid`,
		userID).Scan(&hash)
	return hash, mapErr(err)
}

// ---------------------------------------------------------------------------
// WebAuthn credentials
// ---------------------------------------------------------------------------

// AddWebAuthnCredential stores a freshly registered credential. The full
// webauthn.Credential is persisted as JSON (the source of truth for later
// reconstruction) alongside broken-out columns used for indexing/uniqueness.
func (s *Store) AddWebAuthnCredential(ctx context.Context, userID string, cred *webauthn.Credential) error {
	blob, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	transports := make([]string, 0, len(cred.Transport))
	for _, t := range cred.Transport {
		transports = append(transports, string(t))
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO webauthn_credentials
			(user_id, credential_id, public_key, attestation_type, aaguid,
			 sign_count, transports, clone_warning, credential)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9)`,
		userID,
		cred.ID,
		cred.PublicKey,
		cred.AttestationType,
		cred.Authenticator.AAGUID,
		int64(cred.Authenticator.SignCount),
		transports,
		cred.Authenticator.CloneWarning,
		blob,
	)
	return err
}

// GetWebAuthnCredentials reconstructs a user's credentials from stored JSON.
func (s *Store) GetWebAuthnCredentials(ctx context.Context, userID string) ([]webauthn.Credential, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT credential FROM webauthn_credentials WHERE user_id = $1::uuid ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var creds []webauthn.Credential
	for rows.Next() {
		var blob []byte
		if err := rows.Scan(&blob); err != nil {
			return nil, err
		}
		var c webauthn.Credential
		if err := json.Unmarshal(blob, &c); err != nil {
			return nil, err
		}
		creds = append(creds, c)
	}
	return creds, rows.Err()
}

// CountWebAuthnCredentials returns how many passkeys a user has.
func (s *Store) CountWebAuthnCredentials(ctx context.Context, userID string) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx,
		`SELECT count(*) FROM webauthn_credentials WHERE user_id = $1::uuid`, userID).Scan(&n)
	return n, err
}

// UpdateWebAuthnCredential persists post-login changes (sign count, flags).
func (s *Store) UpdateWebAuthnCredential(ctx context.Context, userID string, cred *webauthn.Credential) error {
	blob, err := json.Marshal(cred)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		UPDATE webauthn_credentials
		SET sign_count = $1, clone_warning = $2, credential = $3, last_used_at = now()
		WHERE user_id = $4::uuid AND credential_id = $5`,
		int64(cred.Authenticator.SignCount),
		cred.Authenticator.CloneWarning,
		blob,
		userID,
		cred.ID,
	)
	return err
}

// ---------------------------------------------------------------------------
// WebAuthn flows (ceremony state)
// ---------------------------------------------------------------------------

// SaveFlow stores ceremony state keyed by a random flow id.
func (s *Store) SaveFlow(ctx context.Context, f Flow) error {
	var uid any
	if f.UserID != nil {
		uid = *f.UserID
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO webauthn_flows (id, user_id, kind, session_data, expires_at)
		VALUES ($1, $2::uuid, $3, $4, $5)`,
		f.ID, uid, f.Kind, f.SessionData, f.ExpiresAt)
	return err
}

// GetFlow fetches ceremony state, or ErrNotFound if missing/expired.
func (s *Store) GetFlow(ctx context.Context, id string) (Flow, error) {
	var f Flow
	var uid *string
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id::text, kind, session_data, expires_at
		FROM webauthn_flows WHERE id = $1 AND expires_at > now()`, id,
	).Scan(&f.ID, &uid, &f.Kind, &f.SessionData, &f.ExpiresAt)
	if err != nil {
		return Flow{}, mapErr(err)
	}
	f.UserID = uid
	return f, nil
}

// DeleteFlow removes ceremony state (called once consumed).
func (s *Store) DeleteFlow(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM webauthn_flows WHERE id = $1`, id)
	return err
}

// ---------------------------------------------------------------------------
// TOTP
// ---------------------------------------------------------------------------

// UpsertTOTPSecret stores an unconfirmed encrypted TOTP secret.
func (s *Store) UpsertTOTPSecret(ctx context.Context, userID string, encrypted []byte) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO totp_secrets (user_id, secret_encrypted, confirmed, created_at)
		VALUES ($1::uuid, $2, false, now())
		ON CONFLICT (user_id)
		DO UPDATE SET secret_encrypted = EXCLUDED.secret_encrypted,
		              confirmed = false, created_at = now(), confirmed_at = NULL`,
		userID, encrypted)
	return err
}

// ConfirmTOTP marks a user's TOTP secret confirmed.
func (s *Store) ConfirmTOTP(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE totp_secrets SET confirmed = true, confirmed_at = now() WHERE user_id = $1::uuid`,
		userID)
	return err
}

// GetTOTPSecret returns the encrypted secret and confirmation state.
func (s *Store) GetTOTPSecret(ctx context.Context, userID string) (encrypted []byte, confirmed bool, err error) {
	err = s.pool.QueryRow(ctx,
		`SELECT secret_encrypted, confirmed FROM totp_secrets WHERE user_id = $1::uuid`,
		userID).Scan(&encrypted, &confirmed)
	return encrypted, confirmed, mapErr(err)
}

// DeleteTOTP disables TOTP for a user (secret + recovery codes).
func (s *Store) DeleteTOTP(ctx context.Context, userID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM totp_recovery_codes WHERE user_id = $1::uuid`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM totp_secrets WHERE user_id = $1::uuid`, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// AddRecoveryCodes stores hashed one-time recovery codes.
func (s *Store) AddRecoveryCodes(ctx context.Context, userID string, hashes []string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM totp_recovery_codes WHERE user_id = $1::uuid`, userID); err != nil {
		return err
	}
	for _, h := range hashes {
		if _, err := tx.Exec(ctx,
			`INSERT INTO totp_recovery_codes (user_id, code_hash) VALUES ($1::uuid, $2)`,
			userID, h); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// UseRecoveryCode marks a matching unused recovery code as used. It reports
// whether a code was consumed.
func (s *Store) UseRecoveryCode(ctx context.Context, userID, codeHash string) (bool, error) {
	ct, err := s.pool.Exec(ctx, `
		UPDATE totp_recovery_codes SET used_at = now()
		WHERE user_id = $1::uuid AND code_hash = $2 AND used_at IS NULL`,
		userID, codeHash)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

// CreateSession inserts a new session row.
func (s *Store) CreateSession(ctx context.Context, id, userID, authMethod string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, auth_method, expires_at)
		VALUES ($1, $2::uuid, $3, $4)`,
		id, userID, authMethod, expiresAt)
	return err
}

// GetSession fetches a live (non-revoked, non-expired) session.
func (s *Store) GetSession(ctx context.Context, id string) (Session, error) {
	var sess Session
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id::text, auth_method, created_at, last_seen_at, expires_at, revoked_at
		FROM sessions
		WHERE id = $1 AND revoked_at IS NULL AND expires_at > now()`, id,
	).Scan(&sess.ID, &sess.UserID, &sess.AuthMethod, &sess.CreatedAt,
		&sess.LastSeenAt, &sess.ExpiresAt, &sess.RevokedAt)
	return sess, mapErr(err)
}

// TouchSession updates last_seen_at.
func (s *Store) TouchSession(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET last_seen_at = now() WHERE id = $1`, id)
	return err
}

// RevokeSession marks a session revoked (logout).
func (s *Store) RevokeSession(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, id)
	return err
}

// ---------------------------------------------------------------------------
// Activity log
// ---------------------------------------------------------------------------

// InsertActivity records an audit event.
func (s *Store) InsertActivity(ctx context.Context, in ActivityInput) error {
	detail := in.Detail
	if detail == nil {
		detail = map[string]any{}
	}
	blob, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	outcome := in.Outcome
	if outcome == "" {
		outcome = "success"
	}
	var uid any
	if in.UserID != nil {
		uid = *in.UserID
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO activity_log
			(user_id, username, event_type, auth_method, outcome, ip_hash, user_agent, detail)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)`,
		uid, in.Username, in.EventType, in.AuthMethod, outcome, in.IPHash, in.UserAgent, blob)
	return err
}

// ListActivityForUser returns a user's most recent audit events.
func (s *Store) ListActivityForUser(ctx context.Context, userID string, limit int) ([]Activity, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id::text, username, event_type, auth_method, outcome,
		       ip_hash, user_agent, detail, created_at
		FROM activity_log
		WHERE user_id = $1::uuid
		ORDER BY created_at DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Activity
	for rows.Next() {
		var a Activity
		if err := rows.Scan(&a.ID, &a.UserID, &a.Username, &a.EventType, &a.AuthMethod,
			&a.Outcome, &a.IPHash, &a.UserAgent, &a.Detail, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// Health check log
// ---------------------------------------------------------------------------

// InsertHealthCheck records the outcome of a readiness probe.
func (s *Store) InsertHealthCheck(ctx context.Context, checkName, status string, latencyMs float64, detail string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO health_check_log (check_name, status, latency_ms, detail)
		VALUES ($1, $2, $3, $4)`,
		checkName, status, latencyMs, detail)
	return err
}

// ---------------------------------------------------------------------------
// CSP reports
// ---------------------------------------------------------------------------

// InsertCSPReport persists a normalised CSP violation report.
func (s *Store) InsertCSPReport(ctx context.Context, in CSPReportInput) error {
	raw := in.Raw
	if len(raw) == 0 {
		raw = json.RawMessage("{}")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO csp_reports
			(document_uri, referrer, blocked_uri, violated_directive, effective_directive,
			 original_policy, disposition, source_file, line_number, column_number,
			 status_code, script_sample, ip_hash, user_agent, raw)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		in.DocumentURI, in.Referrer, in.BlockedURI, in.ViolatedDirective, in.EffectiveDirective,
		in.OriginalPolicy, in.Disposition, in.SourceFile, in.LineNumber, in.ColumnNumber,
		in.StatusCode, in.ScriptSample, in.IPHash, in.UserAgent, []byte(raw))
	return err
}

// ---------------------------------------------------------------------------
// CAPTCHA stats
// ---------------------------------------------------------------------------

// SyncCaptchaStats atomically folds one client batch into the user's aggregate
// row, creating it on first sync. Totals accumulate; best_streak only ever
// grows (GREATEST); current_streak is last-write-wins. The updated row is
// returned so the client can reconcile its display with the server's truth.
func (s *Store) SyncCaptchaStats(ctx context.Context, userID string, in CaptchaSyncInput) (CaptchaStats, error) {
	st := CaptchaStats{UserID: userID}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO captcha_stats
			(user_id, best_streak, current_streak, total_solves, manual_solves, auto_solves, updated_at)
		VALUES ($1::uuid, $2, $3, $4 + $5, $4, $5, now())
		ON CONFLICT (user_id) DO UPDATE SET
			best_streak    = GREATEST(captcha_stats.best_streak, EXCLUDED.best_streak),
			current_streak = EXCLUDED.current_streak,
			total_solves   = captcha_stats.total_solves + $4 + $5,
			manual_solves  = captcha_stats.manual_solves + $4,
			auto_solves    = captcha_stats.auto_solves + $5,
			updated_at     = now()
		RETURNING best_streak, current_streak, total_solves, manual_solves, auto_solves, updated_at`,
		userID, in.BestStreak, in.CurrentStreak, in.ManualDelta, in.AutoDelta,
	).Scan(&st.BestStreak, &st.CurrentStreak, &st.TotalSolves, &st.ManualSolves, &st.AutoSolves, &st.UpdatedAt)
	return st, err
}

// GetCaptchaStats returns a user's aggregate row, or ErrNotFound if the user
// has never synced.
func (s *Store) GetCaptchaStats(ctx context.Context, userID string) (CaptchaStats, error) {
	st := CaptchaStats{UserID: userID}
	err := s.pool.QueryRow(ctx, `
		SELECT best_streak, current_streak, total_solves, manual_solves, auto_solves, updated_at
		FROM captcha_stats WHERE user_id = $1::uuid`, userID,
	).Scan(&st.BestStreak, &st.CurrentStreak, &st.TotalSolves, &st.ManualSolves, &st.AutoSolves, &st.UpdatedAt)
	return st, mapErr(err)
}

// DeleteCaptchaStats removes the user's aggregate row entirely (a true reset:
// the user also disappears from the leaderboard until they play again).
func (s *Store) DeleteCaptchaStats(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM captcha_stats WHERE user_id = $1::uuid`, userID)
	return err
}

// captchaRankedCTE ranks every player once so the top-N query and the "where
// am I" query cannot disagree on ordering. updated_at ASC breaks ties in
// favour of whoever got there first.
const captchaRankedCTE = `
	SELECT user_id, best_streak, total_solves,
	       RANK() OVER (ORDER BY best_streak DESC, total_solves DESC, updated_at ASC) AS rank
	FROM captcha_stats`

// CaptchaLeaderboard returns the top rows ordered by rank.
func (s *Store) CaptchaLeaderboard(ctx context.Context, limit int) ([]CaptchaLeaderboardRow, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	rows, err := s.pool.Query(ctx, `
		WITH ranked AS (`+captchaRankedCTE+`)
		SELECT r.rank, r.user_id::text, u.username, u.display_name, r.best_streak, r.total_solves
		FROM ranked r JOIN users u ON u.id = r.user_id
		ORDER BY r.rank, u.username
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CaptchaLeaderboardRow
	for rows.Next() {
		var lr CaptchaLeaderboardRow
		if err := rows.Scan(&lr.Rank, &lr.UserID, &lr.Username, &lr.DisplayName,
			&lr.BestStreak, &lr.TotalSolves); err != nil {
			return nil, err
		}
		out = append(out, lr)
	}
	return out, rows.Err()
}

// CaptchaRank returns the caller's own ranked row, or ErrNotFound if they have
// never synced any stats.
func (s *Store) CaptchaRank(ctx context.Context, userID string) (CaptchaLeaderboardRow, error) {
	var lr CaptchaLeaderboardRow
	err := s.pool.QueryRow(ctx, `
		WITH ranked AS (`+captchaRankedCTE+`)
		SELECT r.rank, r.user_id::text, u.username, u.display_name, r.best_streak, r.total_solves
		FROM ranked r JOIN users u ON u.id = r.user_id
		WHERE r.user_id = $1::uuid`, userID,
	).Scan(&lr.Rank, &lr.UserID, &lr.Username, &lr.DisplayName, &lr.BestStreak, &lr.TotalSolves)
	return lr, mapErr(err)
}

// ---------------------------------------------------------------------------
// User preferences
// ---------------------------------------------------------------------------

// GetUserPref returns the stored value for a preference key, or ErrNotFound.
func (s *Store) GetUserPref(ctx context.Context, userID, key string) (string, error) {
	var v string
	err := s.pool.QueryRow(ctx,
		`SELECT value FROM user_prefs WHERE user_id = $1::uuid AND key = $2`,
		userID, key).Scan(&v)
	return v, mapErr(err)
}

// SetUserPref upserts a preference value.
func (s *Store) SetUserPref(ctx context.Context, userID, key, value string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO user_prefs (user_id, key, value, updated_at)
		VALUES ($1::uuid, $2, $3, now())
		ON CONFLICT (user_id, key)
		DO UPDATE SET value = EXCLUDED.value, updated_at = now()`,
		userID, key, value)
	return err
}

// ---------------------------------------------------------------------------
// Notes (public microblog)
// ---------------------------------------------------------------------------

// CreateNote inserts a note and returns it with author info attached, so the
// client can render the new card without a second round trip.
func (s *Store) CreateNote(ctx context.Context, userID, body string) (Note, error) {
	var n Note
	err := s.pool.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO notes (user_id, body)
			VALUES ($1::uuid, $2)
			RETURNING id, user_id, body, created_at
		)
		SELECT i.id, i.user_id::text, u.username, u.display_name, i.body, i.created_at
		FROM inserted i JOIN users u ON u.id = i.user_id`,
		userID, body,
	).Scan(&n.ID, &n.UserID, &n.Username, &n.DisplayName, &n.Body, &n.CreatedAt)
	return n, err
}

// ListNotes returns up to limit notes newest-first. When beforeID > 0 only
// notes with id < beforeID are returned — a stable cursor for "load older"
// pagination (ids are monotonic, so the cursor never shifts under the reader
// the way OFFSET would when new notes arrive).
func (s *Store) ListNotes(ctx context.Context, beforeID int64, limit int) ([]Note, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT n.id, n.user_id::text, u.username, u.display_name, n.body, n.created_at
		FROM notes n JOIN users u ON u.id = n.user_id
		WHERE ($1::bigint = 0 OR n.id < $1)
		ORDER BY n.id DESC
		LIMIT $2`, beforeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.UserID, &n.Username, &n.DisplayName, &n.Body, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// DeleteNote hard-deletes a note if and only if it belongs to userID, and
// reports whether a row was actually removed. Ownership is enforced inside
// the single SQL statement, so there is no read-then-delete race and callers
// cannot distinguish "not found" from "not yours" (no existence oracle).
func (s *Store) DeleteNote(ctx context.Context, id int64, userID string) (bool, error) {
	ct, err := s.pool.Exec(ctx,
		`DELETE FROM notes WHERE id = $1 AND user_id = $2::uuid`, id, userID)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// normalizeUsername lowercases and trims a username for case-insensitive
// comparison. Kept here so store lookups and inserts agree on the rule.
func normalizeUsername(u string) string {
	return toLowerTrim(u)
}

func toLowerTrim(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		out = append(out, r)
	}
	// trim spaces
	start, end := 0, len(out)
	for start < end && isSpace(out[start]) {
		start++
	}
	for end > start && isSpace(out[end-1]) {
		end--
	}
	return string(out[start:end])
}

func isSpace(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }
internal/httpx/middleware.go
go
// Package httpx holds transport-level helpers and middleware that do not
// depend on the data store: CORS, request identifiers, panic recovery, a small
// in-memory rate limiter, OpenTelemetry HTTP instrumentation, and JSON writing
// helpers. Keeping these store-agnostic makes them trivial to unit test.
package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ctxKey is a private context key type.
type ctxKey int

const requestIDKey ctxKey = iota

// Middleware is the standard http middleware shape.
type Middleware func(http.Handler) http.Handler

// Chain applies middleware in order (the first listed is outermost).
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// ---------------------------------------------------------------------------
// CORS
// ---------------------------------------------------------------------------

// OriginAllowed decides the Access-Control-Allow-Origin value to echo for a
// request Origin. Because the browser sends credentials (Bearer token flows
// still trigger CORS), the wildcard cannot be combined with credentials, so we
// echo a specific origin. When allowAny is true any non-empty Origin is echoed.
func OriginAllowed(allowed []string, allowAny bool, origin string) (echo string, ok bool) {
	if origin == "" {
		return "", false
	}
	if allowAny {
		return origin, true
	}
	for _, a := range allowed {
		if a == origin {
			return origin, true
		}
	}
	return "", false
}

// CORS returns middleware that applies the origin policy and answers
// preflight OPTIONS requests.
func CORS(allowed []string, allowAny bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if echo, ok := OriginAllowed(allowed, allowAny, origin); ok {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", echo)
				h.Set("Access-Control-Allow-Credentials", "true")
				h.Add("Vary", "Origin")
				h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				h.Set("Access-Control-Max-Age", "600")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ---------------------------------------------------------------------------
// Request ID
// ---------------------------------------------------------------------------

// RequestID attaches a request identifier (reusing an inbound X-Request-Id when
// present) to the context and response headers.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-Id")
			if id == "" {
				id = newID()
			}
			w.Header().Set("X-Request-Id", id)
			ctx := context.WithValue(r.Context(), requestIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDFromContext returns the request id, or "" if unset.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// ---------------------------------------------------------------------------
// Recovery
// ---------------------------------------------------------------------------

// Recoverer converts panics into a 500 response and logs them.
func Recoverer(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.ErrorContext(r.Context(), "panic recovered",
						slog.Any("panic", rec),
						slog.String("path", r.URL.Path),
						slog.String("request_id", RequestIDFromContext(r.Context())),
					)
					WriteError(w, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// ---------------------------------------------------------------------------
// OpenTelemetry HTTP instrumentation
// ---------------------------------------------------------------------------

// Instrument wraps a handler with otelhttp so every request produces a span and
// standard HTTP metrics. It is a no-op-friendly wrapper: when telemetry is
// disabled the global providers are no-ops, so this stays cheap.
func Instrument(h http.Handler, operation string) http.Handler {
	return otelhttp.NewHandler(h, operation)
}

// ---------------------------------------------------------------------------
// Rate limiting (token bucket per key)
// ---------------------------------------------------------------------------

// RateLimiter is a simple per-key token-bucket limiter safe for concurrent use.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     float64 // tokens per second
	capacity float64
	now      func() time.Time // injectable clock for tests
}

type bucket struct {
	tokens float64
	last   time.Time
}

// NewRateLimiter builds a limiter allowing rate tokens/sec with the given burst.
func NewRateLimiter(ratePerSec, burst float64) *RateLimiter {
	if burst < 1 {
		burst = 1
	}
	return &RateLimiter{
		buckets:  map[string]*bucket{},
		rate:     ratePerSec,
		capacity: burst,
		now:      time.Now,
	}
}

// Allow reports whether an event for key is permitted right now.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := rl.now()
	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &bucket{tokens: rl.capacity - 1, last: now}
		return true
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.capacity {
		b.tokens = rl.capacity
	}
	b.last = now
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// LimitByIP returns middleware that rate-limits using a key derived from the
// request (the provided keyFn, typically a hashed client IP). Rejected
// requests receive 429.
func (rl *RateLimiter) LimitByIP(keyFn func(*http.Request) string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.Allow(keyFn(r)) {
				WriteError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ---------------------------------------------------------------------------
// JSON helpers
// ---------------------------------------------------------------------------

// WriteJSON writes v as JSON with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// WriteError writes a JSON error body: {"error": "..."}.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// DecodeJSON reads a size-limited JSON body into dst, writing a 400 and
// returning false on failure. maxBytes <= 0 defaults to 1 MiB.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) bool {
	if maxBytes <= 0 {
		maxBytes = 1 << 20
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := json.NewDecoder(r.

Gotunnels changes
ZIP 
please continue

Body).Decode(dst); err != nil {
WriteError(w, http.StatusBadRequest, "invalid JSON body")
return false
}
return true
}


## `internal/server/server.go`

```go
// Package server assembles the HTTP routes and middleware into a ready-to-run
// *http.Server. Route patterns use the Go 1.22 method-aware ServeMux, so no
// third-party router is needed.
package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/kusl/GoTunnels/internal/activity"
	"github.com/kusl/GoTunnels/internal/auth"
	"github.com/kusl/GoTunnels/internal/captcha"
	"github.com/kusl/GoTunnels/internal/config"
	"github.com/kusl/GoTunnels/internal/csp"
	"github.com/kusl/GoTunnels/internal/health"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/notes"
	"github.com/kusl/GoTunnels/internal/prefs"
)

// Deps are the wired dependencies the server needs.
type Deps struct {
	Config           *config.Config
	Log              *slog.Logger
	Auth             *auth.Handlers
	Health           *health.Handler
	CSP              *csp.Handler
	CSPRateLimiter   *httpx.RateLimiter
	Captcha          *captcha.Handlers
	Notes            *notes.Handlers
	Prefs            *prefs.Handlers
	NotesRateLimiter *httpx.RateLimiter
	Pepper           []byte
}

// New builds the configured *http.Server.
func New(d Deps) *http.Server {
	mux := http.NewServeMux()

	// --- unauthenticated: health & info ---
	mux.HandleFunc("GET /healthz", d.Health.Live)
	mux.HandleFunc("GET /readyz", d.Health.Ready)
	mux.HandleFunc("GET /api/info", d.Health.InfoHandler)

	// --- unauthenticated: auth entry points ---
	mux.HandleFunc("POST /api/signup", d.Auth.Signup)
	mux.HandleFunc("POST /api/login", d.Auth.Login)
	mux.HandleFunc("POST /api/passkey/login/begin", d.Auth.PasskeyLoginBegin)
	mux.HandleFunc("POST /api/passkey/login/finish", d.Auth.PasskeyLoginFinish)

	// --- unauthenticated but rate-limited: CSP violation reports ---
	cspChain := httpx.Chain(
		http.HandlerFunc(d.CSP.ServeHTTP),
		d.CSPRateLimiter.LimitByIP(cspRateKey(d.Pepper)),
	)
	mux.Handle("POST /api/csp-reports", cspChain)

	// --- authenticated ---
	authed := d.Auth.RequireAuth
	mux.Handle("POST /api/logout", authed(http.HandlerFunc(d.Auth.Logout)))
	mux.Handle("GET /api/me", authed(http.HandlerFunc(d.Auth.Me)))
	mux.Handle("GET /api/activity", authed(http.HandlerFunc(d.Auth.Activity)))
	mux.Handle("POST /api/passkey/register/begin", authed(http.HandlerFunc(d.Auth.PasskeyRegisterBegin)))
	mux.Handle("POST /api/passkey/register/finish", authed(http.HandlerFunc(d.Auth.PasskeyRegisterFinish)))
	mux.Handle("POST /api/totp/enroll", authed(http.HandlerFunc(d.Auth.TOTPEnroll)))
	mux.Handle("POST /api/totp/confirm", authed(http.HandlerFunc(d.Auth.TOTPConfirm)))
	mux.Handle("POST /api/totp/disable", authed(http.HandlerFunc(d.Auth.TOTPDisable)))

	// --- authenticated: CAPTCHA stats & leaderboard ---
	mux.Handle("GET /api/captcha/stats", authed(http.HandlerFunc(d.Captcha.Stats)))
	mux.Handle("POST /api/captcha/sync", authed(http.HandlerFunc(d.Captcha.Sync)))
	mux.Handle("POST /api/captcha/reset", authed(http.HandlerFunc(d.Captcha.Reset)))
	mux.Handle("GET /api/captcha/leaderboard", authed(http.HandlerFunc(d.Captcha.Leaderboard)))

	// --- authenticated: per-user preferences ---
	mux.Handle("GET /api/prefs/{key}", authed(http.HandlerFunc(d.Prefs.Get)))
	mux.Handle("PUT /api/prefs/{key}", authed(http.HandlerFunc(d.Prefs.Set)))

	// --- authenticated: notes (plain-text microblog) ---
	// POST is rate-limited per user so one account cannot flood the shared
	// feed; the limiter sits inside RequireAuth so the key derives from the
	// verified user, not from anything the client controls.
	mux.Handle("GET /api/notes", authed(http.HandlerFunc(d.Notes.List)))
	mux.Handle("POST /api/notes", authed(httpx.Chain(
		http.HandlerFunc(d.Notes.Create),
		d.NotesRateLimiter.LimitByIP(noteRateKey()),
	)))
	mux.Handle("DELETE /api/notes/{id}", authed(http.HandlerFunc(d.Notes.Delete)))

	// Global middleware (outermost first). CORS answers OPTIONS preflight
	// before requests reach the mux, so method-specific routes never 405 on
	// preflight.
	handler := httpx.Chain(mux,
		httpx.RequestID(),
		httpx.Recoverer(d.Log),
		httpx.CORS(d.Config.CORSAllowedOrigins, d.Config.CORSAllowsAny()),
	)
	handler = httpx.Instrument(handler, "gotunnels-api")

	return &http.Server{
		Addr:              d.Config.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

// cspRateKey derives a rate-limit key from the hashed client IP so the CSP
// endpoint cannot be trivially flooded, while still never storing a raw IP.
func cspRateKey(pepper []byte) func(*http.Request) string {
	return func(r *http.Request) string {
		return activity.HashIP(pepper, activity.ClientIP(r))
	}
}

// noteRateKey keys the note-creation rate limiter by the authenticated user
// ID. The limiter runs inside RequireAuth, so the user is always present; the
// "anon" fallback only guards against future wiring mistakes.
func noteRateKey() func(*http.Request) string {
	return func(r *http.Request) string {
		if u, ok := auth.CurrentUser(r.Context()); ok {
			return "user:" + u.ID
		}
		return "anon"
	}
}
```

## `cmd/api/main.go`

```go
// Command api is the GoTunnels backend entrypoint. It loads central
// configuration, initialises telemetry, connects to Postgres, applies
// migrations, wires the HTTP server, and runs until interrupted.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kusl/GoTunnels/internal/activity"
	"github.com/kusl/GoTunnels/internal/auth"
	"github.com/kusl/GoTunnels/internal/captcha"
	"github.com/kusl/GoTunnels/internal/config"
	"github.com/kusl/GoTunnels/internal/csp"
	"github.com/kusl/GoTunnels/internal/database"
	"github.com/kusl/GoTunnels/internal/health"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/notes"
	"github.com/kusl/GoTunnels/internal/prefs"
	"github.com/kusl/GoTunnels/internal/server"
	"github.com/kusl/GoTunnels/internal/store"
	"github.com/kusl/GoTunnels/internal/telemetry"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(); err != nil {
		// Fall back to the standard logger if telemetry never came up.
		slog.Error("fatal", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if version != "dev" {
		cfg.Version = version
	}

	// Telemetry (logger is always usable, even on exporter failure).
	tel, err := telemetry.Setup(rootCtx, cfg)
	if err != nil {
		return err
	}
	log := tel.Logger
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tel.Shutdown(shutdownCtx)
	}()

	log.Info("starting gotunnels-api",
		slog.String("instance_id", cfg.InstanceID),
		slog.String("version", cfg.Version),
		slog.String("addr", cfg.HTTPAddr),
		slog.Bool("telemetry", cfg.Telemetry.Enabled),
		slog.String("csp_mode", cfg.CSPMode),
		slog.String("rp_id", cfg.RPID),
	)

	// Database + migrations.
	pool, err := database.Connect(rootCtx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()

	applied, err := database.Migrate(rootCtx, pool)
	if err != nil {
		return err
	}
	log.Info("migrations applied", slog.Int("count", len(applied)), slog.Any("versions", applied))

	// Wiring.
	st := store.New(pool)
	rec := activity.NewRecorder(st, cfg.IPHashPepper())

	wa, err := auth.NewWebAuthn(cfg.RPID, cfg.RPDisplayName, cfg.RPOrigins)
	if err != nil {
		return err
	}

	authHandlers := auth.NewHandlers(st, wa, rec, log, auth.Settings{
		SessionTTL:   cfg.SessionTTL,
		TOTPKey:      cfg.TOTPAESKey(),
		Issuer:       cfg.RPDisplayName,
		CookieName:   cfg.SessionCookieName,
		CookieSecure: !cfg.Dev,
	})

	healthHandler := health.NewHandler(st, log, health.Info{
		Service:     cfg.ServiceName,
		InstanceID:  cfg.InstanceID,
		Version:     cfg.Version,
		CSPMode:     cfg.CSPMode,
		CSPPolicy:   cfg.CSPPolicy,
		TelemetryOn: cfg.Telemetry.Enabled,
	})

	cspHandler := csp.NewHandler(st, log, cfg.IPHashPepper())

	captchaHandlers := captcha.NewHandlers(st, log)
	notesHandlers := notes.NewHandlers(st, log)
	prefsHandlers := prefs.NewHandlers(st, log)

	srv := server.New(server.Deps{
		Config:         cfg,
		Log:            log,
		Auth:           authHandlers,
		Health:         healthHandler,
		CSP:            cspHandler,
		CSPRateLimiter: httpx.NewRateLimiter(5, 20), // 5 rps, burst 20, per hashed IP
		Captcha:        captchaHandlers,
		Notes:          notesHandlers,
		Prefs:          prefsHandlers,
		// One note every 2 seconds sustained, short bursts of 5, per user.
		NotesRateLimiter: httpx.NewRateLimiter(0.5, 5),
		Pepper:           cfg.IPHashPepper(),
	})

	// Run the server and wait for either a serve error or a shutdown signal.
	serveErr := make(chan error, 1)
	go func() {
		log.Info("http server listening", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return err
	case <-rootCtx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", slog.String("error", err.Error()))
		return err
	}
	log.Info("shutdown complete")
	return nil
}
```

## `frontend/js/api.js`

```javascript
// api.js — the single place browser code talks to the API.
//
// The primary session transport is a Bearer token kept in sessionStorage. The
// server also sets a cross-site cookie as a secondary path, so we send
// credentials too; but because third-party cookies are increasingly blocked,
// the Bearer token is what we rely on.

import { loadConfig } from "./config.js";

const TOKEN_KEY = "gotunnels_token";

export function getToken() {
  return sessionStorage.getItem(TOKEN_KEY) || "";
}
export function setToken(t) {
  if (t) sessionStorage.setItem(TOKEN_KEY, t);
}
export function clearToken() {
  sessionStorage.removeItem(TOKEN_KEY);
}

// apiFetch performs a JSON request against the API base and throws an Error
// (with .status and .data) on non-2xx responses.
//
// Pass { keepalive: true } for small requests that must survive the page
// being hidden or unloaded (e.g. flushing CAPTCHA stats). We cannot use
// navigator.sendBeacon because it cannot carry the Authorization header.
export async function apiFetch(path, opts = {}) {
  const { method = "GET", body, headers = {}, keepalive = false } = opts;
  const cfg = await loadConfig();
  const base = cfg.apiBase || "";

  const h = { Accept: "application/json", ...headers };
  const token = getToken();
  if (token) h["Authorization"] = "Bearer " + token;

  let payload;
  if (body !== undefined) {
    h["Content-Type"] = "application/json";
    payload = JSON.stringify(body);
  }

  const res = await fetch(base + path, {
    method,
    headers: h,
    body: payload,
    mode: "cors",
    credentials: "include",
    keepalive,
  });

  const text = await res.text();
  let data = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = { raw: text };
    }
  }

  if (!res.ok) {
    const msg = data && data.error ? data.error : "request failed (" + res.status + ")";
    const err = new Error(msg);
    err.status = res.status;
    err.data = data;
    throw err;
  }
  return data;
}

// Typed-ish endpoint wrappers.
export const Api = {
  signup: (b) => apiFetch("/api/signup", { method: "POST", body: b }),
  login: (b) => apiFetch("/api/login", { method: "POST", body: b }),
  logout: () => apiFetch("/api/logout", { method: "POST" }),
  me: () => apiFetch("/api/me"),
  activity: () => apiFetch("/api/activity"),
  info: () => apiFetch("/api/info"),

  passkeyRegisterBegin: () => apiFetch("/api/passkey/register/begin", { method: "POST" }),
  passkeyRegisterFinish: (flow, body) =>
    apiFetch("/api/passkey/register/finish?flow=" + encodeURIComponent(flow), { method: "POST", body }),
  passkeyLoginBegin: (b) => apiFetch("/api/passkey/login/begin", { method: "POST", body: b }),
  passkeyLoginFinish: (flow, body) =>
    apiFetch("/api/passkey/login/finish?flow=" + encodeURIComponent(flow), { method: "POST", body }),

  totpEnroll: () => apiFetch("/api/totp/enroll", { method: "POST" }),
  totpConfirm: (b) => apiFetch("/api/totp/confirm", { method: "POST", body: b }),
  totpDisable: (b) => apiFetch("/api/totp/disable", { method: "POST", body: b }),

  captchaStats: () => apiFetch("/api/captcha/stats"),
  captchaSync: (b, keepalive = false) =>
    apiFetch("/api/captcha/sync", { method: "POST", body: b, keepalive }),
  captchaReset: () => apiFetch("/api/captcha/reset", { method: "POST" }),
  captchaLeaderboard: () => apiFetch("/api/captcha/leaderboard"),

  prefGet: (key) => apiFetch("/api/prefs/" + encodeURIComponent(key)),
  prefSet: (key, value) =>
    apiFetch("/api/prefs/" + encodeURIComponent(key), { method: "PUT", body: { value } }),

  notesList: (params = {}) => {
    const q = new URLSearchParams();
    if (params.before) q.set("before", String(params.before));
    if (params.limit) q.set("limit", String(params.limit));
    const qs = q.toString();
    return apiFetch("/api/notes" + (qs ? "?" + qs : ""));
  },
  noteCreate: (body) => apiFetch("/api/notes", { method: "POST", body: { body } }),
  noteDelete: (id) => apiFetch("/api/notes/" + encodeURIComponent(id), { method: "DELETE" }),
};
```

## `frontend/captcha.html`

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>CAPTCHA · GoTunnels</title>
    <link rel="stylesheet" href="/css/style.css" />
  </head>
  <body>
    <header class="topbar">
      <div class="topbar-inner">
        <span class="brand"><span class="dot"></span> GoTunnels</span>
        <nav class="mainnav">
          <a href="/">Home</a>
          <a href="/notes" data-auth="in" class="hidden">Notes</a>
          <a href="/captcha" data-auth="in" class="hidden">CAPTCHA</a>
          <a href="/activity" data-auth="in" class="hidden">Activity</a>
          <a href="/settings" data-auth="in" class="hidden">Settings</a>
          <a href="/login" data-auth="out">Log in</a>
          <a href="/signup" data-auth="out">Sign up</a>
        </nav>
      </div>
    </header>

    <main class="narrow">
      <h1>Magic Solve CAPTCHA</h1>
      <p class="lead">Tap the matching letters yourself, or let the auto-solver do it. Your streaks and solve counts are stored on the server and feed the leaderboard.</p>

      <div class="card">
        <div class="captcha-header">
          <p class="instructions">Select every tile showing <strong id="target" class="target-badge">A</strong></p>
          <span class="chip-row">
            <span id="syncChip" class="status-pill" title="Server sync status">local</span>
            <span id="fpsChip" class="status-pill" aria-hidden="true">0 fps</span>
          </span>
        </div>

        <div class="stats">
          <div class="stat"><span class="stat-value" id="streak">0</span><span class="stat-label">Streak</span></div>
          <div class="stat"><span class="stat-value" id="best">0</span><span class="stat-label">Best</span></div>
          <div class="stat"><span class="stat-value" id="solves">0</span><span class="stat-label">Solves</span></div>
        </div>

        <fieldset class="captcha-grid" id="grid" aria-label="CAPTCHA letter grid">
          <legend class="sr-only">Toggle every tile that matches the target letter, then verify</legend>
        </fieldset>

        <div id="statusLine" class="msg info"></div>

        <div class="btn-row btn-grid-2">
          <button type="button" id="verifyBtn" class="primary">Verify selection</button>
          <button type="button" id="autoBtn" class="ghost">✨ Magic Solve</button>
        </div>

        <div class="speed-control">
          <div class="speed-control-head">
            <label for="speedSlider">Magic Solve speed</label>
            <output id="speedLabel" for="speedSlider"></output>
          </div>
          <input type="range" id="speedSlider" min="1" max="100" step="1" value="35" aria-describedby="speedLabel">
          <div class="speed-scale-labels" aria-hidden="true"><span>1 tap/sec</span><span>Hardware max</span></div>
        </div>

        <div class="divider"></div>
        <div class="btn-row center">
          <button type="button" id="clearBtn" class="danger ghost">Reset my stats</button>
        </div>
      </div>

      <details id="leaderboard" class="card leaderboard">
        <summary>
          <span>Leaderboard — best streaks</span>
          <span id="lbStatus" class="status-pill"></span>
        </summary>
        <table class="lb-table">
          <caption class="sr-only">Top players ranked by best streak</caption>
          <thead>
            <tr>
              <th>#</th>
              <th>Player</th>
              <th>Best</th>
              <th>Solves</th>
            </tr>
          </thead>
          <tbody id="lbRows"></tbody>
        </table>
        <p id="lbMe" class="hint hidden"></p>
      </details>
    </main>

    <footer class="foot">
      GoTunnels · built with LLM assistance (see the README)
    </footer>

    <dialog id="clearDialog" class="confirm-dialog">
      <form method="dialog">
        <h2>Reset your stats?</h2>
        <p>This deletes your streak, best streak, and solve count from the <strong>server</strong>, removes you from the leaderboard, and resets the local speed setting. It cannot be undone.</p>
        <div class="dialog-actions">
          <button type="submit" value="cancel" class="ghost" autofocus>Cancel</button>
          <button type="submit" value="confirm" class="danger">Reset</button>
        </div>
      </form>
    </dialog>

    <script type="module" src="/js/page-captcha.js"></script>
  </body>
</html>
```

## `frontend/js/page-captcha.js`

```javascript
// page-captcha.js — the Magic Solve CAPTCHA page.
//
// The puzzle itself (grid, verify, auto-solver) runs entirely in the browser,
// but the *statistics* live on the server so they survive devices and feed the
// leaderboard. Because the auto-solver can complete hundreds of puzzles per
// second, we never send one request per solve. Instead the page accumulates
// deltas locally and flushes them in small batches:
//
//   - every few seconds while anything is pending,
//   - when the auto-solver stops,
//   - and with fetch({keepalive:true}) when the tab is hidden or unloading,
//     so the last batch survives navigation. (navigator.sendBeacon cannot be
//     used here: it cannot carry the Authorization header.)
//
// Displayed solves = server base + un-flushed pending + in-flight, so the
// numbers never jump backwards while a sync is in the air.
//
// Settings split (see docs/ARCHITECTURE.md):
//   - leaderboard open/closed  -> server-side user preference (follows account)
//   - Magic Solve speed        -> localStorage (depends on this device's display)

import { highlightNav, renderAuthNav, requireAuth } from "./common.js";
import { Api } from "./api.js";

/* =====================================================================
   Local storage (device-scoped settings only)
   ===================================================================== */
const STORAGE_PREFIX = "gotunnels:captcha:";
const KEYS = {
  speed: STORAGE_PREFIX + "speed",
  // Legacy key from the frontend-only version of this page; stats now live
  // on the server, so we delete it on boot.
  legacyStats: STORAGE_PREFIX + "stats",
};
const LEADERBOARD_PREF_KEY = "captcha.leaderboard.open";

function safeGet(key) {
  try { return localStorage.getItem(key); } catch { return null; }
}
function safeSet(key, value) {
  try { localStorage.setItem(key, value); } catch { /* quota/private mode */ }
}
function safeRemove(key) {
  try { localStorage.removeItem(key); } catch { /* storage unavailable */ }
}

function defaultSpeed() { return 35; }
function loadSpeed() {
  const n = Number(safeGet(KEYS.speed));
  return Number.isFinite(n) && n >= 1 && n <= 100 ? n : defaultSpeed();
}
function saveSpeed(v) { safeSet(KEYS.speed, String(v)); }

function prefersReducedMotion() {
  return typeof matchMedia === "function" && matchMedia("(prefers-reduced-motion: reduce)").matches;
}

function debounce(fn, ms) {
  let t;
  return (...args) => {
    clearTimeout(t);
    t = setTimeout(() => fn(...args), ms);
  };
}

/* =====================================================================
   Minimal Signals
   ===================================================================== */
let activeSubscriber = null;

class StateSignal {
  #value;
  #observers = new Set();
  constructor(initial) { this.#value = initial; }
  get() {
    if (activeSubscriber) this.#observers.add(activeSubscriber);
    return this.#value;
  }
  set(next) {
    if (Object.is(next, this.#value)) return;
    this.#value = next;
    [...this.#observers].forEach((run) => run());
  }
}

class ComputedSignal {
  #fn;
  #value;
  #dirty = true;
  #observers = new Set();
  constructor(fn) { this.#fn = fn; }
  #recompute = () => {
    this.#dirty = true;
    [...this.#observers].forEach((run) => run());
  };
  get() {
    if (this.#dirty) {
      const prev = activeSubscriber;
      activeSubscriber = this.#recompute;
      try { this.#value = this.#fn(); }
      finally { activeSubscriber = prev; }
      this.#dirty = false;
    }
    if (activeSubscriber) this.#observers.add(activeSubscriber);
    return this.#value;
  }
}

function effect(fn) {
  const run = () => {
    const prev = activeSubscriber;
    activeSubscriber = run;
    try { fn(); } finally { activeSubscriber = prev; }
  };
  run();
}

const Signal = { State: StateSignal, Computed: ComputedSignal };

/* =====================================================================
   DOM references
   ===================================================================== */
const gridEl = document.getElementById("grid");
const targetEl = document.getElementById("target");
const streakEl = document.getElementById("streak");
const bestEl = document.getElementById("best");
const solvesEl = document.getElementById("solves");
const statusLine = document.getElementById("statusLine");
const verifyBtn = document.getElementById("verifyBtn");
const autoBtn = document.getElementById("autoBtn");
const speedSlider = document.getElementById("speedSlider");
const speedLabel = document.getElementById("speedLabel");
const clearBtn = document.getElementById("clearBtn");
const clearDialog = document.getElementById("clearDialog");
const fpsChip = document.getElementById("fpsChip");
const syncChip = document.getElementById("syncChip");
const leaderboardEl = document.getElementById("leaderboard");
const lbRows = document.getElementById("lbRows");
const lbStatus = document.getElementById("lbStatus");
const lbMe = document.getElementById("lbMe");

/* =====================================================================
   App state (signals)
   ===================================================================== */
// Server-acknowledged values.
const baseSolvesSignal = new Signal.State(0); // manual+auto totals from server
const bestSignal = new Signal.State(0);
// Live, local values.
const streakSignal = new Signal.State(0);
const pendingSignal = new Signal.State(0); // pending + in-flight solve deltas
const speedSignal = new Signal.State(loadSpeed());
const phaseSignal = new Signal.State("idle"); // 'idle' | 'auto'
const syncStateSignal = new Signal.State("idle"); // 'idle' | 'dirty' | 'syncing' | 'error'
const solvesComputed = new Signal.Computed(() => baseSolvesSignal.get() + pendingSignal.get());
const tapsInfoComputed = new Signal.Computed(() => speedToTapsInfo(speedSignal.get()));

let me = null; // current user, set on boot

/* =====================================================================
   Server sync engine
   ===================================================================== */
const SYNC_INTERVAL_MS = 4000;

// Deltas not yet handed to a request.
let pending = { manual: 0, auto: 0 };
// Deltas currently inside an in-flight request (restored to pending on error).
let inflight = null;
let dirtyStreak = false; // streak/best changed since last successful sync
let syncTimer = null;
let syncInFlight = false;

function pendingTotal() {
  return pending.manual + pending.auto + (inflight ? inflight.manual + inflight.auto : 0);
}
function refreshPendingSignal() {
  pendingSignal.set(pendingTotal());
}

function recordSolve(mode) {
  if (mode === "auto") pending.auto += 1;
  else pending.manual += 1;
  dirtyStreak = true;
  refreshPendingSignal();
  syncStateSignal.set("dirty");
  scheduleSync();
}

function recordStreakReset() {
  dirtyStreak = true;
  syncStateSignal.set("dirty");
  scheduleSync();
}

function scheduleSync() {
  if (syncTimer !== null) return; // a flush is already coming
  syncTimer = setTimeout(() => {
    syncTimer = null;
    void flushSync();
  }, SYNC_INTERVAL_MS);
}

function hasWork() {
  return pending.manual > 0 || pending.auto > 0 || dirtyStreak;
}

function applyServer(stats) {
  baseSolvesSignal.set((stats.manual_solves || 0) + (stats.auto_solves || 0));
  bestSignal.set(Math.max(bestSignal.get(), stats.best_streak || 0));
}

// flushSync sends everything accumulated so far. `keepalive` is used on
// pagehide/visibility-hidden so the request can outlive the page. The
// returned promise is also stored in syncPromise so performReset can wait
// out an in-flight request before deleting the server row.
let syncPromise = null;

function flushSync(keepalive = false) {
  if (syncInFlight || !hasWork()) return Promise.resolve();
  syncPromise = doFlush(keepalive);
  return syncPromise;
}

async function doFlush(keepalive) {
  syncInFlight = true;

  inflight = pending;
  pending = { manual: 0, auto: 0 };
  const hadDirtyStreak = dirtyStreak;
  dirtyStreak = false;

  const body = {
    manual_delta: inflight.manual,
    auto_delta: inflight.auto,
    current_streak: streakSignal.get(),
    best_streak: bestSignal.get(),
  };

  syncStateSignal.set("syncing");
  try {
    const res = await Api.captchaSync(body, keepalive);
    inflight = null;
    refreshPendingSignal();
    if (res && res.stats) applyServer(res.stats);
    if (hasWork()) {
      syncStateSignal.set("dirty");
      scheduleSync();
    } else {
      syncStateSignal.set("idle");
    }
  } catch (err) {
    // Restore what we tried to send so nothing is lost, then retry later.
    pending.manual += inflight.manual;
    pending.auto += inflight.auto;
    inflight = null;
    dirtyStreak = dirtyStreak || hadDirtyStreak;
    refreshPendingSignal();
    syncStateSignal.set("error");
    if (err && err.status === 401) {
      location.href = "/login";
      return;
    }
    scheduleSync();
  } finally {
    syncInFlight = false;
    syncPromise = null;
  }
}

// Flush with keepalive when the page goes away; this is the reliable path for
// "user closes the tab mid-auto-solve".
document.addEventListener("visibilitychange", () => {
  if (document.visibilityState === "hidden") void flushSync(true);
});
window.addEventListener("pagehide", () => { void flushSync(true); });

/* =====================================================================
   Puzzle logic (unchanged from the frontend-only version)
   ===================================================================== */
const LETTERS = "ABCDEFGHJKLMNPQRSTUVWXYZ";

let currentLetters = [];
let target = "";
let selected = new Set();

function showStatus(text) {
  statusLine.textContent = text;
  statusLine.classList.add("show");
}

function pickTarget() {
  return LETTERS[Math.floor(Math.random() * LETTERS.length)];
}

function buildGrid(targetLetter) {
  const letters = new Array(9);
  const matchCount = 1 + Math.floor(Math.random() * 4);
  const matchIndices = new Set();
  while (matchIndices.size < matchCount) {
    matchIndices.add(Math.floor(Math.random() * 9));
  }
  for (let i = 0; i < 9; i++) {
    if (matchIndices.has(i)) {
      letters[i] = targetLetter;
    } else {
      let l;
      do { l = LETTERS[Math.floor(Math.random() * LETTERS.length)]; } while (l === targetLetter);
      letters[i] = l;
    }
  }
  return letters;
}

function getCorrectIndices() {
  const s = new Set();
  currentLetters.forEach((l, i) => { if (l === target) s.add(i); });
  return s;
}

function renderGrid(letters) {
  gridEl.innerHTML = "";
  letters.forEach((letter, i) => {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "cell";
    btn.dataset.index = String(i);
    btn.setAttribute("aria-pressed", "false");
    btn.textContent = letter;
    btn.addEventListener("click", () => toggleCell(i));
    gridEl.appendChild(btn);
  });
}

function toggleCell(i) {
  if (phaseSignal.get() !== "idle") return;
  const btn = gridEl.children[i];
  if (!btn) return;
  if (selected.has(i)) {
    selected.delete(i);
    btn.classList.remove("selected");
    btn.setAttribute("aria-pressed", "false");
  } else {
    selected.add(i);
    btn.classList.add("selected");
    btn.setAttribute("aria-pressed", "true");
  }
}

function clearFeedbackClasses() {
  gridEl.querySelectorAll(".cell").forEach((btn) => {
    btn.classList.remove("fb-correct", "fb-missed", "fb-wrong");
  });
  statusLine.classList.remove("show", "error", "ok");
}

function newPuzzle() {
  target = pickTarget();
  currentLetters = buildGrid(target);
  selected = new Set();
  targetEl.textContent = target;
  renderGrid(currentLetters);
}

function verify() {
  if (phaseSignal.get() !== "idle") return;

  const correctIndices = getCorrectIndices();
  const isCorrect = correctIndices.size === selected.size && [...correctIndices].every((i) => selected.has(i));

  gridEl.querySelectorAll(".cell").forEach((btn, i) => {
    btn.disabled = true;
    const isTarget = correctIndices.has(i);
    const isSelected = selected.has(i);
    if (isTarget && isSelected) btn.classList.add("fb-correct");
    else if (isTarget && !isSelected) btn.classList.add("fb-missed");
    else if (!isTarget && isSelected) btn.classList.add("fb-wrong");
  });
  verifyBtn.disabled = true;

  if (isCorrect) {
    const streak = streakSignal.get() + 1;
    streakSignal.set(streak);
    bestSignal.set(Math.max(bestSignal.get(), streak));
    recordSolve("manual");
    showStatus(`Correct! Streak ${streak}.`);
    statusLine.classList.add("ok");
  } else {
    streakSignal.set(0);
    recordStreakReset();
    showStatus("Not quite — streak reset.");
    statusLine.classList.add("error");
  }

  const delay = prefersReducedMotion() ? 400 : 900;
  setTimeout(() => {
    clearFeedbackClasses();
    newPuzzle();
    syncInteractivity();
  }, delay);
}

/* =====================================================================
   Auto solver ("Magic Solve")
   ===================================================================== */
function speedToTapsInfo(v) {
  if (v >= 100) return { tapsPerSec: Infinity, delayMs: 0 };
  const t = (v - 1) / 99;
  const tapsPerSec = Math.pow(600, t);
  return { tapsPerSec, delayMs: 1000 / tapsPerSec };
}

let rafId = null;
let autoTargets = [];
let autoIndex = 0;
let lastTapTime = 0;

function startAutoSolve() {
  clearFeedbackClasses();
  selected = new Set();
  gridEl.querySelectorAll(".cell").forEach((btn) => {
    btn.classList.remove("selected");
    btn.setAttribute("aria-pressed", "false");
  });

  autoTargets = currentLetters.map((l, i) => (l === target ? i : -1)).filter((i) => i !== -1);
  autoIndex = 0;
  lastTapTime = 0;
  phaseSignal.set("auto");

  if (autoTargets.length === 0) {
    finishAutoRound();
    return;
  }
  rafId = requestAnimationFrame(autoStep);
}

function autoStep(now) {
  if (phaseSignal.get() !== "auto") return;

  const { delayMs } = speedToTapsInfo(speedSignal.get());
  if (lastTapTime === 0 || now - lastTapTime >= delayMs) {
    const idx = autoTargets[autoIndex++];
    selected.add(idx);
    const btn = gridEl.children[idx];
    if (btn) { btn.classList.add("selected"); btn.setAttribute("aria-pressed", "true"); }
    lastTapTime = now;
  }

  if (autoIndex >= autoTargets.length) {
    finishAutoRound();
    return;
  }
  rafId = requestAnimationFrame(autoStep);
}

function finishAutoRound() {
  const streak = streakSignal.get() + 1;
  streakSignal.set(streak);
  bestSignal.set(Math.max(bestSignal.get(), streak));
  recordSolve("auto");
  showStatus(`Magic Solve — streak ${streak}.`);
  statusLine.classList.add("ok");

  const delay = prefersReducedMotion() ? 150 : 320;
  setTimeout(() => {
    if (phaseSignal.get() === "auto") {
      newPuzzle();
      startAutoSolve();
    }
  }, delay);
}

// haltAutoSolve stops the animation loop without flushing; stopAutoSolve is
// the user-facing stop, which also pushes the accumulated batch.
function haltAutoSolve() {
  if (rafId) cancelAnimationFrame(rafId);
  rafId = null;
  phaseSignal.set("idle");
  selected = new Set();
  gridEl.querySelectorAll(".cell").forEach((btn) => {
    btn.classList.remove("selected");
    btn.setAttribute("aria-pressed", "false");
  });
}

function stopAutoSolve() {
  haltAutoSolve();
  // The solver may have racked up a large batch; push it promptly.
  void flushSync();
}

function syncInteractivity() {
  const solving = phaseSignal.get() === "auto";
  gridEl.querySelectorAll(".cell").forEach((btn) => { btn.disabled = solving; });
  verifyBtn.disabled = solving;
}

autoBtn.addEventListener("click", () => {
  if (phaseSignal.get() === "auto") stopAutoSolve();
  else startAutoSolve();
});

verifyBtn.addEventListener("click", verify);

/* =====================================================================
   Speed slider (device setting -> localStorage)
   ===================================================================== */
const persistSpeed = debounce((v) => saveSpeed(v), 300);
speedSlider.value = String(speedSignal.get());
speedSlider.addEventListener("input", () => {
  speedSignal.set(Number(speedSlider.value));
});

/* =====================================================================
   Leaderboard (collapsed by default; open state is a server-side pref)
   ===================================================================== */
const LB_REFRESH_MIN_MS = 5000;
let lbLastLoaded = 0;
let lbLoading = false;
let restoringPref = false; // guards the toggle listener during programmatic open

function lbCell(text, mono = false) {
  const td = document.createElement("td");
  td.textContent = text;
  if (mono) td.classList.add("mono");
  return td;
}

async function loadLeaderboard(force = false) {
  if (lbLoading) return;
  const now = Date.now();
  if (!force && now - lbLastLoaded < LB_REFRESH_MIN_MS) return;
  lbLoading = true;
  lbStatus.textContent = "loading…";
  try {
    const res = await Api.captchaLeaderboard();
    lbLastLoaded = Date.now();
    const rows = (res && res.leaderboard) || [];
    lbRows.innerHTML = "";
    if (rows.length === 0) {
      const tr = document.createElement("tr");
      const td = document.createElement("td");
      td.colSpan = 4;
      td.textContent = "No entries yet — solve a puzzle to appear here.";
      tr.appendChild(td);
      lbRows.appendChild(tr);
    }
    for (const row of rows) {
      const tr = document.createElement("tr");
      if (me && row.user_id === me.id) tr.classList.add("lb-row-me");
      tr.appendChild(lbCell(String(row.rank)));
      tr.appendChild(lbCell(row.display_name || row.username));
      tr.appendChild(lbCell(String(row.best_streak), true));
      tr.appendChild(lbCell(String(row.total_solves), true));
      lbRows.appendChild(tr);
    }
    const mine = res && res.me;
    if (mine && !rows.some((r) => me && r.user_id === me.id)) {
      lbMe.textContent = `You are #${mine.rank} with a best streak of ${mine.best_streak} (${mine.total_solves} solves).`;
      lbMe.classList.remove("hidden");
    } else {
      lbMe.classList.add("hidden");
    }
    lbStatus.textContent = "";
  } catch {
    lbStatus.textContent = "failed to load";
  } finally {
    lbLoading = false;
  }
}

leaderboardEl.addEventListener("toggle", () => {
  if (leaderboardEl.open) void loadLeaderboard(true);
  // The toggle event fires asynchronously, so the boot-time restore consumes
  // its guard flag here rather than resetting it synchronously.
  if (restoringPref) { restoringPref = false; return; }
  void Api.prefSet(LEADERBOARD_PREF_KEY, leaderboardEl.open ? "1" : "0").catch(() => {});
});

/* =====================================================================
   Reset (server + local)
   ===================================================================== */
clearBtn.addEventListener("click", () => clearDialog.showModal());

clearDialog.addEventListener("close", () => {
  if (clearDialog.returnValue === "confirm") void performReset();
});

async function performReset() {
  haltAutoSolve();

  // Discard everything queued locally and cancel the pending flush timer.
  pending = { manual: 0, auto: 0 };
  dirtyStreak = false;
  if (syncTimer !== null) { clearTimeout(syncTimer); syncTimer = null; }

  // Let any airborne sync land (or fail) before deleting the server row, so
  // a stale response cannot resurrect stats afterwards. Its error handler may
  // restore deltas into `pending`; discard those too.
  if (syncPromise) { try { await syncPromise; } catch { /* already handled */ } }
  pending = { manual: 0, auto: 0 };
  dirtyStreak = false;
  if (syncTimer !== null) { clearTimeout(syncTimer); syncTimer = null; }
  refreshPendingSignal();

  try {
    await Api.captchaReset();
  } catch (err) {
    if (err && err.status === 401) { location.href = "/login"; return; }
    showStatus("Reset failed — please try again.");
    statusLine.classList.add("error");
    return;
  }

  safeRemove(KEYS.speed);
  streakSignal.set(0);
  bestSignal.set(0);
  baseSolvesSignal.set(0);
  speedSignal.set(defaultSpeed());
  speedSlider.value = String(defaultSpeed());
  syncStateSignal.set("idle");
  clearFeedbackClasses();
  newPuzzle();
  syncInteractivity();
  showStatus("Your stats have been reset.");
  if (leaderboardEl.open) void loadLeaderboard(true);
}

/* =====================================================================
   Effects (signals -> DOM)
   ===================================================================== */
effect(() => {
  streakEl.textContent = String(streakSignal.get());
});
effect(() => {
  bestEl.textContent = String(bestSignal.get());
});
effect(() => {
  solvesEl.textContent = String(solvesComputed.get());
});

effect(() => {
  const { tapsPerSec } = tapsInfoComputed.get();
  speedLabel.textContent = tapsPerSec === Infinity ? "MAX ⚡ hardware limit" : `${tapsPerSec.toFixed(1)} taps/sec`;
  persistSpeed(speedSignal.get());
});

effect(() => {
  const solving = phaseSignal.get() === "auto";
  autoBtn.textContent = solving ? "🛑 Stop" : "✨ Magic Solve";
  if (solving) {
    autoBtn.classList.add("danger");
    autoBtn.classList.remove("ghost");
  } else {
    autoBtn.classList.remove("danger");
    autoBtn.classList.add("ghost");
  }
  autoBtn.setAttribute("aria-pressed", String(solving));
  syncInteractivity();
});

effect(() => {
  const state = syncStateSignal.get();
  const labels = { idle: "synced", dirty: "pending…", syncing: "syncing…", error: "sync failed" };
  syncChip.textContent = labels[state] || state;
  syncChip.classList.toggle("pill-error", state === "error");
});

// Refresh an open leaderboard after each successful sync so your own row is
// current, without hammering the endpoint (loadLeaderboard throttles itself).
effect(() => {
  if (syncStateSignal.get() === "idle" && leaderboardEl.open) {
    void loadLeaderboard(false);
  }
});

/* =====================================================================
   FPS counter
   ===================================================================== */
let lastFrameTime = performance.now();
let frameCount = 0;
function updateFps(now) {
  frameCount++;
  const delta = now - lastFrameTime;
  if (delta >= 1000) {
    fpsChip.textContent = `${Math.round((frameCount * 1000) / delta)} fps`;
    frameCount = 0;
    lastFrameTime = now;
  }
  requestAnimationFrame(updateFps);
}
requestAnimationFrame(updateFps);

/* =====================================================================
   Boot
   ===================================================================== */
async function boot() {
  highlightNav();
  renderAuthNav();

  me = await requireAuth();
  if (!me) return; // redirected to /login

  safeRemove(KEYS.legacyStats);

  newPuzzle();
  syncInteractivity();

  // Load server stats and the leaderboard pref in parallel.
  const [statsRes, prefRes] = await Promise.allSettled([
    Api.captchaStats(),
    Api.prefGet(LEADERBOARD_PREF_KEY),
  ]);

  if (statsRes.status === "fulfilled" && statsRes.value && statsRes.value.stats) {
    const s = statsRes.value.stats;
    applyServer(s);
    // Resume the streak you left off with on any device.
    streakSignal.set(s.current_streak || 0);
    syncStateSignal.set("idle");
  } else {
    syncChip.textContent = "offline";
  }

  if (prefRes.status === "fulfilled" && prefRes.value && prefRes.value.exists && prefRes.value.value === "1") {
    restoringPref = true;
    leaderboardEl.open = true; // queues a toggle event, which consumes the flag
  }
}

void boot();
```

## `frontend/notes.html`

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Notes · GoTunnels</title>
    <link rel="stylesheet" href="/css/style.css" />
  </head>
  <body>
    <header class="topbar">
      <div class="topbar-inner">
        <span class="brand"><span class="dot"></span> GoTunnels</span>
        <nav class="mainnav">
          <a href="/">Home</a>
          <a href="/notes" data-auth="in" class="hidden">Notes</a>
          <a href="/captcha" data-auth="in" class="hidden">CAPTCHA</a>
          <a href="/activity" data-auth="in" class="hidden">Activity</a>
          <a href="/settings" data-auth="in" class="hidden">Settings</a>
          <a href="/login" data-auth="out">Log in</a>
          <a href="/signup" data-auth="out">Sign up</a>
        </nav>
      </div>
    </header>

    <main>
      <h1>Notes</h1>
      <p class="lead">
        A shared plain-text feed. Everyone signed in sees every note. No
        images, no attachments, no clickable links — just text you can copy.
        Notes can be deleted by their author, never edited.
      </p>

      <div id="msg" class="msg"></div>

      <div class="card">
        <label for="noteBody">Write a note</label>
        <textarea id="noteBody" rows="3" maxlength="500"
          placeholder="Up to 500 characters of plain text…"></textarea>
        <div class="composer-foot">
          <span id="charCount" class="hint">0 / 500</span>
          <button type="button" id="postBtn" class="primary">Post</button>
        </div>
      </div>

      <div class="feed-head">
        <h2>Latest notes</h2>
        <span id="feedStatus" class="status-pill" title="Auto-refresh status">live</span>
      </div>

      <button type="button" id="newPill" class="new-pill hidden">New notes ↑</button>

      <div id="feed" class="feed" aria-live="polite"></div>

      <div class="btn-row center">
        <button type="button" id="loadOlder" class="ghost hidden">Load older notes</button>
      </div>
    </main>

    <footer class="foot">
      GoTunnels · built with LLM assistance (see the README)
    </footer>

    <dialog id="deleteDialog" class="confirm-dialog">
      <form method="dialog">
        <h2>Delete this note?</h2>
        <p>The note is removed for everyone, permanently. Deleting cannot be undone, and notes cannot be edited — only deleted.</p>
        <div class="dialog-actions">
          <button type="submit" value="cancel" class="ghost" autofocus>Cancel</button>
          <button type="submit" value="confirm" class="danger">Delete</button>
        </div>
      </form>
    </dialog>

    <script type="module" src="/js/page-notes.js"></script>
  </body>
</html>
```

## `frontend/js/page-notes.js`

```javascript
// page-notes.js — the shared plain-text notes feed.
//
// "Auto-refresh" here is deliberate, boring polling (see
// docs/ARCHITECTURE.md § Live updates): the page fetches the newest 50 notes
// every 5 seconds while visible, pauses when the tab is hidden, refreshes
// immediately when it becomes visible again, and backs off exponentially (up
// to 60s) while the API is unreachable.
//
// Each poll is reconciled against what is already on screen, keyed by note
// id:
//   - notes we hold that fall inside the fetched window but are missing from
//     it were deleted -> removed from the DOM immediately;
//   - new notes written by *you* appear immediately;
//   - new notes by others appear immediately too, unless you have scrolled
//     down to read — then they queue behind a "New notes ↑" pill so the feed
//     does not jump under your thumb.
//
// Rendering is plain-text only by construction: note bodies are set with
// textContent, never innerHTML, so nothing in a note can become markup, a
// link, or a script.

import { Api } from "./api.js";
import { qs, showMsg, clearMsg, highlightNav, renderAuthNav, requireAuth } from "./common.js";

const POLL_MS = 5000;
const MAX_BACKOFF_MS = 60000;
const PAGE = 50;
const MAX_CHARS = 500;
const SCROLL_DEFER_PX = 200;

const msgEl = qs("#msg");
const bodyEl = qs("#noteBody");
const charCount = qs("#charCount");
const postBtn = qs("#postBtn");
const feedEl = qs("#feed");
const feedStatus = qs("#feedStatus");
const newPill = qs("#newPill");
const loadOlderBtn = qs("#loadOlder");
const deleteDialog = qs("#deleteDialog");

let me = null;

// id -> note object for everything currently rendered.
const noteMap = new Map();
// id -> DOM node.
const nodeMap = new Map();
// Notes by other people that arrived while the reader was scrolled down.
const pendingNew = new Map();

let oldestLoadedId = 0; // pagination cursor for "Load older"
let reachedEnd = false;

let pollTimer = null;
let backoffMs = POLL_MS;
let polling = false;
let initialLoaded = false; // suppress "new" highlights on the very first fill

/* =====================================================================
   Rendering
   ===================================================================== */
function fmtRelative(iso) {
  const then = new Date(iso).getTime();
  if (!Number.isFinite(then)) return "";
  const secs = Math.max(0, Math.floor((Date.now() - then) / 1000));
  if (secs < 45) return "just now";
  if (secs < 3600) return Math.floor(secs / 60) + "m ago";
  if (secs < 86400) return Math.floor(secs / 3600) + "h ago";
  if (secs < 7 * 86400) return Math.floor(secs / 86400) + "d ago";
  return new Date(iso).toLocaleDateString();
}

async function copyText(text, btn) {
  let ok = false;
  try {
    await navigator.clipboard.writeText(text);
    ok = true;
  } catch {
    // Clipboard API can be unavailable (permissions, older browsers);
    // fall back to a hidden textarea + execCommand.
    try {
      const ta = document.createElement("textarea");
      ta.value = text;
      ta.setAttribute("readonly", "");
      ta.className = "sr-only";
      document.body.appendChild(ta);
      ta.select();
      ok = document.execCommand("copy");
      ta.remove();
    } catch { ok = false; }
  }
  const prev = btn.textContent;
  btn.textContent = ok ? "Copied ✓" : "Copy failed";
  btn.disabled = true;
  setTimeout(() => { btn.textContent = prev; btn.disabled = false; }, 1200);
}

function renderNote(n) {
  const card = document.createElement("article");
  card.className = "note-card";
  card.dataset.id = String(n.id);

  const head = document.createElement("div");
  head.className = "note-head";

  const author = document.createElement("span");
  author.className = "note-author";
  author.textContent = n.display_name || n.username;

  const username = document.createElement("span");
  username.className = "note-username";
  username.textContent = "@" + n.username;

  const time = document.createElement("time");
  time.className = "note-time";
  time.dateTime = n.created_at;
  time.title = new Date(n.created_at).toLocaleString();
  time.textContent = fmtRelative(n.created_at);

  head.append(author, username, time);

  const body = document.createElement("p");
  body.className = "note-body";
  // textContent, never innerHTML: bodies stay inert plain text and links in
  // them are not clickable, by design.
  body.textContent = n.body;

  const actions = document.createElement("div");
  actions.className = "note-actions";

  const copyBtn = document.createElement("button");
  copyBtn.type = "button";
  copyBtn.className = "ghost btn-small";
  copyBtn.textContent = "Copy";
  copyBtn.addEventListener("click", () => { void copyText(n.body, copyBtn); });
  actions.appendChild(copyBtn);

  if (me && n.user_id === me.id) {
    const delBtn = document.createElement("button");
    delBtn.type = "button";
    delBtn.className = "danger ghost btn-small";
    delBtn.textContent = "Delete";
    delBtn.addEventListener("click", () => confirmDelete(n.id));
    actions.appendChild(delBtn);
  }

  card.append(head, body, actions);
  return card;
}

// syncDOM makes the on-screen order match noteMap sorted by id descending.
// appendChild moves nodes that already exist, so this is a cheap reorder.
function syncDOM() {
  const ids = [...noteMap.keys()].sort((a, b) => b - a);
  for (const id of ids) {
    let node = nodeMap.get(id);
    if (!node) {
      node = renderNote(noteMap.get(id));
      nodeMap.set(id, node);
    }
    feedEl.appendChild(node);
  }
}

function insertNote(n, highlight = false) {
  if (noteMap.has(n.id)) return;
  noteMap.set(n.id, n);
  syncDOM();
  if (highlight) {
    const node = nodeMap.get(n.id);
    if (node) node.classList.add("note-new");
  }
}

function removeNote(id) {
  noteMap.delete(id);
  pendingNew.delete(id);
  const node = nodeMap.get(id);
  if (node) node.remove();
  nodeMap.delete(id);
}

function refreshTimes() {
  for (const [id, node] of nodeMap) {
    const n = noteMap.get(id);
    if (!n) continue;
    const t = node.querySelector(".note-time");
    if (t) t.textContent = fmtRelative(n.created_at);
  }
}

function updatePill() {
  if (pendingNew.size > 0) {
    newPill.textContent = pendingNew.size === 1 ? "1 new note ↑" : pendingNew.size + " new notes ↑";
    newPill.classList.remove("hidden");
  } else {
    newPill.classList.add("hidden");
  }
}

newPill.addEventListener("click", () => {
  for (const [id, n] of pendingNew) {
    if (!noteMap.has(id)) noteMap.set(id, n);
  }
  pendingNew.clear();
  syncDOM();
  updatePill();
  window.scrollTo({ top: 0, behavior: "smooth" });
});

/* =====================================================================
   Polling with reconciliation
   ===================================================================== */
function setStatus(text) {
  feedStatus.textContent = text;
}

async function pollOnce() {
  if (polling) return;
  polling = true;
  try {
    const res = await Api.notesList({ limit: PAGE });
    const fetched = (res && res.notes) || [];
    backoffMs = POLL_MS;
    setStatus(document.visibilityState === "visible" ? "live" : "paused");

    const fetchedIds = new Set(fetched.map((n) => n.id));
    // The poll returns the newest PAGE notes. Anything we hold in that same
    // range but missing from the response was deleted. If the response is
    // empty the whole feed is empty. Notes older than the window cannot be
    // checked this way; a deletion there shows up on the next full reload.
    const windowMin = fetched.length > 0 ? fetched[fetched.length - 1].id : Infinity;
    const toRemove = [];
    for (const id of noteMap.keys()) {
      if ((fetched.length === 0 || id >= windowMin) && !fetchedIds.has(id)) toRemove.push(id);
    }
    for (const [id] of pendingNew) {
      if ((fetched.length === 0 || id >= windowMin) && !fetchedIds.has(id)) pendingNew.delete(id);
    }
    toRemove.forEach(removeNote);

    const deferOthers = window.scrollY > SCROLL_DEFER_PX;
    for (const n of fetched) {
      if (noteMap.has(n.id) || pendingNew.has(n.id)) continue;
      const mine = me && n.user_id === me.id;
      if (mine || !deferOthers || !initialLoaded) {
        insertNote(n, initialLoaded);
      } else {
        pendingNew.set(n.id, n);
      }
    }
    updatePill();
    refreshTimes();

    if (oldestLoadedId === 0 && fetched.length > 0) {
      oldestLoadedId = fetched[fetched.length - 1].id;
    }
    if (fetched.length >= PAGE && !reachedEnd) {
      loadOlderBtn.classList.remove("hidden");
    }
    initialLoaded = true;
    clearMsg(msgEl);
  } catch (err) {
    if (err && err.status === 401) {
      location.href = "/login";
      return;
    }
    backoffMs = Math.min(backoffMs * 2, MAX_BACKOFF_MS);
    setStatus("reconnecting…");
  } finally {
    polling = false;
    scheduleNextPoll();
  }
}

function scheduleNextPoll() {
  if (pollTimer !== null) clearTimeout(pollTimer);
  if (document.visibilityState !== "visible") {
    pollTimer = null;
    setStatus("paused");
    return;
  }
  pollTimer = setTimeout(() => {
    pollTimer = null;
    void pollOnce();
  }, backoffMs);
}

document.addEventListener("visibilitychange", () => {
  if (document.visibilityState === "visible") {
    backoffMs = POLL_MS;
    void pollOnce(); // refresh immediately on return
  } else {
    if (pollTimer !== null) { clearTimeout(pollTimer); pollTimer = null; }
    setStatus("paused");
  }
});

/* =====================================================================
   Pagination ("Load older")
   ===================================================================== */
loadOlderBtn.addEventListener("click", async () => {
  loadOlderBtn.disabled = true;
  try {
    const res = await Api.notesList({ before: oldestLoadedId, limit: PAGE });
    const older = (res && res.notes) || [];
    for (const n of older) {
      if (!noteMap.has(n.id)) noteMap.set(n.id, n);
    }
    if (older.length > 0) {
      oldestLoadedId = older[older.length - 1].id;
      syncDOM();
    }
    if (older.length < PAGE) {
      reachedEnd = true;
      loadOlderBtn.classList.add("hidden");
    }
  } catch (err) {
    if (err && err.status === 401) { location.href = "/login"; return; }
    showMsg(msgEl, "Could not load older notes: " + err.message);
  } finally {
    loadOlderBtn.disabled = false;
  }
});

/* =====================================================================
   Composer
   ===================================================================== */
function updateCharCount() {
  // Count what the server counts: unicode code points, not UTF-16 units.
  charCount.textContent = [...bodyEl.value].length + " / " + MAX_CHARS;
}
bodyEl.addEventListener("input", updateCharCount);

postBtn.addEventListener("click", async () => {
  const text = bodyEl.value.trim();
  if (!text) {
    showMsg(msgEl, "Write something first.");
    return;
  }
  postBtn.disabled = true;
  clearMsg(msgEl);
  try {
    const res = await Api.noteCreate(text);
    if (res && res.note) {
      insertNote(res.note, true);
      if (oldestLoadedId === 0) oldestLoadedId = res.note.id;
    }
    bodyEl.value = "";
    updateCharCount();
    window.scrollTo({ top: 0, behavior: "smooth" });
  } catch (err) {
    if (err && err.status === 401) { location.href = "/login"; return; }
    if (err && err.status === 429) {
      showMsg(msgEl, "You are posting too fast — wait a moment and try again.");
    } else {
      showMsg(msgEl, err.message || "Could not post the note.");
    }
  } finally {
    postBtn.disabled = false;
  }
});

/* =====================================================================
   Delete (own notes only)
   ===================================================================== */
let deleteTargetId = null;

function confirmDelete(id) {
  deleteTargetId = id;
  deleteDialog.showModal();
}

deleteDialog.addEventListener("close", () => {
  const id = deleteTargetId;
  deleteTargetId = null;
  if (deleteDialog.returnValue !== "confirm" || id === null) return;
  void (async () => {
    try {
      await Api.noteDelete(id);
      removeNote(id);
      updatePill();
    } catch (err) {
      if (err && err.status === 401) { location.href = "/login"; return; }
      // 404 means it is already gone (another tab, another device).
      if (err && err.status === 404) { removeNote(id); updatePill(); return; }
      showMsg(msgEl, "Could not delete the note: " + err.message);
    }
  })();
});

/* =====================================================================
   Boot
   ===================================================================== */
async function boot() {
  highlightNav();
  renderAuthNav();

  me = await requireAuth();
  if (!me) return; // redirected to /login

  updateCharCount();
  setStatus("loading…");
  await pollOnce(); // initial load; also schedules the next poll
}

void boot();
```

## `frontend/index.html`

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>GoTunnels</title>
    <link rel="stylesheet" href="/css/style.css" />
  </head>
  <body>
    <header class="topbar">
      <div class="topbar-inner">
        <span class="brand"><span class="dot"></span> GoTunnels</span>
        <nav class="mainnav">
          <a href="/">Home</a>
          <a href="/notes" data-auth="in" class="hidden">Notes</a>
          <a href="/captcha" data-auth="in" class="hidden">CAPTCHA</a>
          <a href="/activity" data-auth="in" class="hidden">Activity</a>
          <a href="/settings" data-auth="in" class="hidden">Settings</a>
          <a href="/login" data-auth="out">Log in</a>
          <a href="/signup" data-auth="out">Sign up</a>
        </nav>
      </div>
    </header>

    <main>
      <h1>GoTunnels</h1>
      <p class="lead">
        A fully self-contained demo stack: a Go API and a plain HTML/CSS/JS web
        app, each reachable only through its own Cloudflare Quick Tunnel, with
        Postgres and telemetry behind them. Nothing binds to a host port.
      </p>

      <div class="card">
        <p class="status-pill" id="auth-status">Checking session…</p>
        <p class="status-pill" id="instance-info">Loading instance info…</p>
      </div>

      <h2>What this demonstrates</h2>
      <div class="card">
        <ul>
          <li>Passkey (WebAuthn) login, password fallback, and optional TOTP two-factor.</li>
          <li>Server-side sessions with a Bearer token that works cross-origin.</li>
          <li>Privacy-preserving audit log: IPs are hashed with a per-instance pepper, never stored raw.</li>
          <li>A shared plain-text notes feed with visibility-aware polling for live updates.</li>
          <li>A CAPTCHA toy whose stats sync to the server in batches and feed a leaderboard.</li>
          <li>OpenTelemetry traces, metrics, and logs over OTLP — vendor-neutral.</li>
          <li>Content-Security-Policy in report-only mode, reported back to the API.</li>
          <li>No exposed host ports; every service talks over the internal Compose network.</li>
        </ul>
      </div>

      <h2>Get started</h2>
      <div class="card">
        <div data-auth="out">
          <p>Create an account, then add a passkey or enable TOTP from settings.</p>
          <div class="btn-row">
            <a href="/signup"><button class="primary" type="button">Sign up</button></a>
            <a href="/login"><button class="ghost" type="button">Log in</button></a>
          </div>
        </div>
        <div data-auth="in" class="hidden">
          <p>You are signed in. Jump straight to the good parts:</p>
          <div class="btn-row">
            <a href="/notes"><button class="primary" type="button">Notes</button></a>
            <a href="/captcha"><button class="ghost" type="button">CAPTCHA</button></a>
            <a href="/settings"><button class="ghost" type="button">Settings</button></a>
          </div>
        </div>
      </div>
    </main>

    <footer class="foot">
      GoTunnels · built with LLM assistance (see the README) ·
      <code>github.com/kusl/GoTunnels</code>
    </footer>

    <script type="module" src="/js/page-index.js"></script>
  </body>
</html>
```

## `frontend/login.html`

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Log in · GoTunnels</title>
    <link rel="stylesheet" href="/css/style.css" />
  </head>
  <body>
    <header class="topbar">
      <div class="topbar-inner">
        <span class="brand"><span class="dot"></span> GoTunnels</span>
        <nav class="mainnav">
          <a href="/">Home</a>
          <a href="/notes" data-auth="in" class="hidden">Notes</a>
          <a href="/captcha" data-auth="in" class="hidden">CAPTCHA</a>
          <a href="/activity" data-auth="in" class="hidden">Activity</a>
          <a href="/settings" data-auth="in" class="hidden">Settings</a>
          <a href="/login" data-auth="out">Log in</a>
          <a href="/signup" data-auth="out">Sign up</a>
        </nav>
      </div>
    </header>

    <main class="narrow">
      <h1>Log in</h1>

      <div class="card">
        <div id="msg" class="msg"></div>
        <form id="login-form" novalidate>
          <label for="username">Username</label>
          <input id="username" type="text" autocomplete="username" required />

          <label for="password">Password</label>
          <input id="password" type="password" autocomplete="current-password" required />

          <label for="totp">Two-factor code (if enabled)</label>
          <input id="totp" type="text" inputmode="numeric" autocomplete="one-time-code"
                 placeholder="6-digit code or recovery code" />

          <div class="btn-row">
            <button id="submit" class="primary" type="submit">Log in</button>
          </div>
        </form>

        <div class="divider">or</div>

        <div class="btn-row">
          <button id="passkey-login" class="ghost" type="button">Log in with a passkey</button>
        </div>
        <p class="hint">Enter your username above, then use your passkey.</p>
      </div>
    </main>

    <footer class="foot">
      GoTunnels · built with LLM assistance (see the README)
    </footer>

    <script type="module" src="/js/page-login.js"></script>
  </body>
</html>
```

## `frontend/signup.html`

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Sign up · GoTunnels</title>
    <link rel="stylesheet" href="/css/style.css" />
  </head>
  <body>
    <header class="topbar">
      <div class="topbar-inner">
        <span class="brand"><span class="dot"></span> GoTunnels</span>
        <nav class="mainnav">
          <a href="/">Home</a>
          <a href="/notes" data-auth="in" class="hidden">Notes</a>
          <a href="/captcha" data-auth="in" class="hidden">CAPTCHA</a>
          <a href="/activity" data-auth="in" class="hidden">Activity</a>
          <a href="/settings" data-auth="in" class="hidden">Settings</a>
          <a href="/login" data-auth="out">Log in</a>
          <a href="/signup" data-auth="out">Sign up</a>
        </nav>
      </div>
    </header>

    <main class="narrow">
      <h1>Create an account</h1>
      <p class="lead">Signup is one step — no email or SMS verification (this is a demo).</p>

      <div class="card">
        <div id="msg" class="msg"></div>
        <form id="signup-form" novalidate>
          <label for="username">Username</label>
          <input id="username" type="text" autocomplete="username" required
                 placeholder="3–32 chars: letters, digits, . - _" />

          <label for="display_name">Display name (optional)</label>
          <input id="display_name" type="text" autocomplete="nickname" placeholder="Shown to you" />

          <label for="password">Password</label>
          <input id="password" type="password" autocomplete="new-password" required
                 placeholder="at least 8 characters" />
          <p class="hint">A password is always set as a fallback. You can add a passkey afterwards.</p>

          <div class="btn-row">
            <button id="submit" class="primary" type="submit">Create account</button>
            <a href="/login"><button class="ghost" type="button">I already have an account</button></a>
          </div>
        </form>
      </div>
    </main>

    <footer class="foot">
      GoTunnels · built with LLM assistance (see the README)
    </footer>

    <script type="module" src="/js/page-signup.js"></script>
  </body>
</html>
```

## `frontend/settings.html`

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Settings · GoTunnels</title>
    <link rel="stylesheet" href="/css/style.css" />
  </head>
  <body>
    <header class="topbar">
      <div class="topbar-inner">
        <span class="brand"><span class="dot"></span> GoTunnels</span>
        <nav class="mainnav">
          <a href="/">Home</a>
          <a href="/notes" data-auth="in" class="hidden">Notes</a>
          <a href="/captcha" data-auth="in" class="hidden">CAPTCHA</a>
          <a href="/activity" data-auth="in" class="hidden">Activity</a>
          <a href="/settings" data-auth="in" class="hidden">Settings</a>
          <a href="/login" data-auth="out">Log in</a>
          <a href="/signup" data-auth="out">Sign up</a>
        </nav>
      </div>
    </header>

    <main>
      <h1>Settings</h1>
      <div id="msg" class="msg"></div>

      <div class="card">
        <h2>Profile</h2>
        <dl class="kv">
          <dt>Username</dt><dd id="p-username">—</dd>
          <dt>Display name</dt><dd id="p-display">—</dd>
          <dt>Roles</dt><dd id="p-roles">—</dd>
          <dt>Passkeys</dt><dd id="p-passkeys">0</dd>
          <dt>TOTP</dt><dd id="p-totp">disabled</dd>
        </dl>
      </div>

      <div class="card">
        <h2>Passkeys</h2>
        <p class="lead">Add a passkey for passwordless, phishing-resistant login.</p>
        <div class="btn-row">
          <button id="register-passkey" class="primary" type="button">Register a passkey</button>
        </div>
      </div>

      <div class="card" id="totp-enroll-section">
        <h2>Two-factor (TOTP)</h2>
        <p class="lead">Use Aegis, Google Authenticator, or any RFC 6238 app.</p>
        <div class="btn-row">
          <button id="totp-enroll" class="primary" type="button">Begin TOTP enrollment</button>
        </div>

        <div id="totp-enroll-result" class="hidden">
          <h3>1 · Add this secret to your app</h3>
          <dl class="kv">
            <dt>Secret</dt><dd id="totp-secret">—</dd>
            <dt>otpauth URL</dt><dd id="totp-url">—</dd>
          </dl>

          <h3>2 · Save your recovery codes</h3>
          <p class="hint">Each code works once. Store them somewhere safe — they are shown only now.</p>
          <pre class="codes" id="totp-recovery">—</pre>

          <h3>3 · Confirm a code</h3>
          <form id="totp-confirm-form" novalidate>
            <label for="totp-confirm-code">Current 6-digit code</label>
            <input id="totp-confirm-code" type="text" inputmode="numeric"
                   autocomplete="one-time-code" placeholder="123456" />
            <div class="btn-row">
              <button class="primary" type="submit">Confirm &amp; enable</button>
            </div>
          </form>
        </div>
      </div>

      <div class="card hidden" id="totp-disable-section">
        <h2>Two-factor (TOTP)</h2>
        <p class="lead">TOTP is enabled on your account.</p>
        <form id="totp-disable-form" novalidate>
          <label for="totp-disable-code">Enter a current code or a recovery code to disable</label>
          <input id="totp-disable-code" type="text" inputmode="numeric"
                 autocomplete="one-time-code" placeholder="123456 or recovery code" />
          <div class="btn-row">
            <button class="danger" type="submit">Disable TOTP</button>
          </div>
        </form>
      </div>

      <div class="card">
        <div class="btn-row">
          <button id="logout" class="ghost" type="button">Log out</button>
        </div>
      </div>
    </main>

    <footer class="foot">
      GoTunnels · built with LLM assistance (see the README)
    </footer>

    <script type="module" src="/js/page-settings.js"></script>
  </body>
</html>
```

## `frontend/activity.html`

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Activity · GoTunnels</title>
    <link rel="stylesheet" href="/css/style.css" />
  </head>
  <body>
    <header class="topbar">
      <div class="topbar-inner">
        <span class="brand"><span class="dot"></span> GoTunnels</span>
        <nav class="mainnav">
          <a href="/">Home</a>
          <a href="/notes" data-auth="in" class="hidden">Notes</a>
          <a href="/captcha" data-auth="in" class="hidden">CAPTCHA</a>
          <a href="/activity" data-auth="in" class="hidden">Activity</a>
          <a href="/settings" data-auth="in" class="hidden">Settings</a>
          <a href="/login" data-auth="out">Log in</a>
          <a href="/signup" data-auth="out">Sign up</a>
        </nav>
      </div>
    </header>

    <main>
      <h1>Your activity</h1>
      <p class="lead">
        Every sign-up and login is recorded. Your IP address is never stored —
        only a salted SHA-256 hash of it, shown here exactly as stored.
      </p>

      <div id="msg" class="msg"></div>

      <div class="card">
        <table class="activity-table">
          <thead>
            <tr>
              <th>When</th>
              <th>Event</th>
              <th>Method</th>
              <th>Outcome</th>
              <th>IP hash</th>
            </tr>
          </thead>
          <tbody id="activity-body"></tbody>
        </table>
      </div>
    </main>

    <footer class="foot">
      GoTunnels · built with LLM assistance (see the README)
    </footer>

    <script type="module" src="/js/page-activity.js"></script>
  </body>
</html>
```

## `frontend/css/style.css`

```css
/* GoTunnels frontend styles.
   Aesthetic: a quiet, instrument-panel dark theme. One accent (signal green),
   a monospace face for data, a clean sans for prose. No third-party fonts —
   everything is a system font stack so the strict self-only CSP holds. */

:root {
  --bg: #0d1117;
  --bg-elev: #151b23;
  --bg-elev-2: #1b232d;
  --border: #2a333f;
  --border-bright: #3a4553;
  --text: #e6edf3;
  --text-dim: #9aa7b4;
  --text-faint: #667181;
  --accent: #3fb950;
  --accent-dim: #2ea043;
  --accent-glow: rgba(63, 185, 80, 0.15);
  --danger: #f85149;
  --warn: #d29922;
  --radius: 10px;
  --mono: ui-monospace, "SF Mono", "SFMono-Regular", "Menlo", "Consolas", "Liberation Mono", monospace;
  --sans: system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
}

* { box-sizing: border-box; }

html, body {
  margin: 0;
  padding: 0;
  background: var(--bg);
  color: var(--text);
  font-family: var(--sans);
  font-size: 15px;
  line-height: 1.55;
  -webkit-font-smoothing: antialiased;
}

body {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

a { color: var(--accent); text-decoration: none; }
a:hover { text-decoration: underline; }

/* ---- top bar ---- */
header.topbar {
  border-bottom: 1px solid var(--border);
  background: var(--bg-elev);
  position: sticky;
  top: 0;
  z-index: 10;
}
.topbar-inner {
  max-width: 960px;
  margin: 0 auto;
  padding: 14px 20px;
  display: flex;
  align-items: center;
  gap: 20px;
}
.brand {
  font-family: var(--mono);
  font-weight: 600;
  letter-spacing: 0.5px;
  color: var(--text);
  display: flex;
  align-items: center;
  gap: 9px;
}
.brand .dot {
  width: 9px; height: 9px; border-radius: 50%;
  background: var(--accent);
  box-shadow: 0 0 0 4px var(--accent-glow);
}
nav.mainnav { display: flex; gap: 18px; margin-left: auto; font-size: 14px; }
nav.mainnav a { color: var(--text-dim); }
nav.mainnav a.active { color: var(--text); }
nav.mainnav a:hover { color: var(--text); text-decoration: none; }

/* ---- layout ---- */
main {
  max-width: 960px;
  width: 100%;
  margin: 0 auto;
  padding: 32px 20px 64px;
  flex: 1;
}
.narrow { max-width: 460px; }

footer.foot {
  border-top: 1px solid var(--border);
  color: var(--text-faint);
  font-size: 13px;
  padding: 18px 20px;
  text-align: center;
}
footer.foot code { color: var(--text-dim); }

/* ---- headings ---- */
h1 { font-size: 26px; margin: 0 0 8px; letter-spacing: -0.2px; }
h2 { font-size: 18px; margin: 28px 0 10px; }
h3 { font-size: 15px; margin: 20px 0 8px; color: var(--text-dim); text-transform: uppercase; letter-spacing: 0.6px; }
p.lead { color: var(--text-dim); margin-top: 0; }

/* ---- cards ---- */
.card {
  background: var(--bg-elev);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 22px;
  margin: 16px 0;
}
.card h2:first-child, .card h3:first-child { margin-top: 0; }

/* ---- forms ---- */
label { display: block; font-size: 13px; color: var(--text-dim); margin: 14px 0 6px; }
input[type="text"], input[type="password"], input[type="email"] {
  width: 100%;
  padding: 10px 12px;
  background: var(--bg);
  border: 1px solid var(--border-bright);
  border-radius: 8px;
  color: var(--text);
  font-family: var(--mono);
  font-size: 14px;
}
input:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-glow);
}

button {
  font-family: var(--sans);
  font-size: 14px;
  font-weight: 500;
  padding: 10px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-bright);
  background: var(--bg-elev-2);
  color: var(--text);
  cursor: pointer;
  transition: border-color 0.12s ease, background 0.12s ease;
}
button:hover { border-color: var(--accent); }
button.primary {
  background: var(--accent-dim);
  border-color: var(--accent);
  color: #04140a;
  font-weight: 600;
}
button.primary:hover { background: var(--accent); }
button.ghost { background: transparent; }
button.danger { color: var(--danger); border-color: #5a2d2d; }
button.danger:hover { border-color: var(--danger); }
button:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-row { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 18px; }

.divider { display: flex; align-items: center; gap: 12px; color: var(--text-faint); font-size: 12px; margin: 22px 0; }
.divider::before, .divider::after { content: ""; height: 1px; background: var(--border); flex: 1; }

/* ---- messages ---- */
.msg { border-radius: 8px; padding: 10px 12px; margin: 14px 0; font-size: 14px; display: none; }
.msg.show { display: block; }
.msg.error { background: rgba(248, 81, 73, 0.1); border: 1px solid #5a2d2d; color: #ff9d97; }
.msg.ok { background: var(--accent-glow); border: 1px solid var(--accent-dim); color: #7ee787; }
.msg.info { background: rgba(210, 153, 34, 0.08); border: 1px solid #5a4a1d; color: #e3b341; }

/* ---- data table ---- */
table { width: 100%; border-collapse: collapse; font-size: 13px; }
th, td { text-align: left; padding: 9px 10px; border-bottom: 1px solid var(--border); }
th { color: var(--text-faint); font-weight: 500; text-transform: uppercase; letter-spacing: 0.5px; font-size: 11px; }
td { font-family: var(--mono); color: var(--text-dim); }
td.hash { color: var(--text-faint); word-break: break-all; max-width: 260px; }
.tag { display: inline-block; padding: 2px 8px; border-radius: 999px; font-size: 11px; font-family: var(--mono); border: 1px solid var(--border-bright); }
.tag.success { color: #7ee787; border-color: var(--accent-dim); }
.tag.failure { color: #ff9d97; border-color: #5a2d2d; }

/* ---- misc ---- */
.kv { display: grid; grid-template-columns: 140px 1fr; gap: 8px 16px; font-size: 14px; }
.kv dt { color: var(--text-faint); font-family: var(--mono); }
.kv dd { margin: 0; font-family: var(--mono); color: var(--text); word-break: break-all; }
code, pre { font-family: var(--mono); }
pre.codes {
  background: var(--bg);
  border: 1px dashed var(--border-bright);
  border-radius: 8px;
  padding: 14px;
  color: #7ee787;
  font-size: 14px;
  line-height: 1.9;
  white-space: pre-wrap;
  word-break: break-all;
}
.hint { color: var(--text-faint); font-size: 12px; margin-top: 6px; }
.hidden { display: none !important; }
.qr { background: #fff; padding: 12px; border-radius: 8px; display: inline-block; margin-top: 8px; }
.qr img { display: block; width: 180px; height: 180px; }
.status-pill { font-family: var(--mono); font-size: 12px; color: var(--text-dim); }

/* ---- Mobile Responsiveness ---- */
@media (max-width: 650px) {
  /* Fix the navigation bar */
  .topbar-inner {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  nav.mainnav {
    margin-left: 0;
    width: 100%;
    justify-content: space-between;
  }

  /* Transform the activity table into cards. Scoped to .activity-table so
     other tables (e.g. the CAPTCHA leaderboard) keep their tabular layout
     on small screens. */
  table.activity-table,
  .activity-table thead,
  .activity-table tbody,
  .activity-table th,
  .activity-table td,
  .activity-table tr {
    display: block;
  }

  /* Visually hide the header, but keep it accessible for screen readers */
  .activity-table thead tr {
    position: absolute;
    top: -9999px;
    left: -9999px;
  }

  /* Style the row as a card */
  .activity-table tr {
    border-bottom: 2px solid var(--border-bright);
    margin-bottom: 16px;
    padding-bottom: 8px;
  }

  /* Style the cells as key-value pairs */
  .activity-table td {
    border: none;
    border-bottom: 1px solid var(--border);
    position: relative;
    padding-left: 110px;
    text-align: right;
    min-height: 38px; /* Ensures consistent height */
  }

  .activity-table td:last-child {
    border-bottom: 0;
  }

  /* Inject the column headers as labels */
  .activity-table td::before {
    position: absolute;
    top: 10px;
    left: 10px;
    width: 90px;
    padding-right: 10px;
    white-space: nowrap;
    text-align: left;
    font-weight: 500;
    color: var(--text-faint);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .activity-table td:nth-of-type(1)::before { content: "When"; }
  .activity-table td:nth-of-type(2)::before { content: "Event"; }
  .activity-table td:nth-of-type(3)::before { content: "Method"; }
  .activity-table td:nth-of-type(4)::before { content: "Outcome"; }
  .activity-table td:nth-of-type(5)::before { content: "IP hash"; }

  /* Allow the hash to take up the remaining space naturally */
  .activity-table td.hash {
    max-width: none; 
  }
}

/* ---- shared utilities ---- */
/* Visually hidden but present for screen readers (used by legends/captions
   and as the clipboard fallback scratch element). */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
.btn-row.center { justify-content: center; }
.btn-grid-2 { display: grid; grid-template-columns: 1fr 1fr; }
.btn-grid-2 .ghost { border-color: var(--border-bright); }
.chip-row { display: inline-flex; align-items: center; gap: 10px; }
.status-pill.pill-error { color: var(--danger); }
.mono { font-family: var(--mono); }

textarea {
  width: 100%;
  padding: 10px 12px;
  background: var(--bg);
  border: 1px solid var(--border-bright);
  border-radius: 8px;
  color: var(--text);
  font-family: var(--sans);
  font-size: 14px;
  line-height: 1.5;
  resize: vertical;
  min-height: 72px;
}
textarea:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-glow);
}

/* ---- confirm dialog (captcha reset, note delete) ---- */
dialog.confirm-dialog {
  background: var(--bg-elev);
  color: var(--text);
  border: 1px solid var(--border-bright);
  border-radius: var(--radius);
  padding: 24px;
  max-width: 400px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
}
dialog.confirm-dialog::backdrop {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(2px);
}
dialog.confirm-dialog h2 { margin-top: 0; }
dialog.confirm-dialog p { color: var(--text-dim); font-size: 14px; margin-bottom: 24px; }
dialog.confirm-dialog .dialog-actions { display: flex; justify-content: flex-end; gap: 10px; }

/* ---- CAPTCHA page ---- */
.captcha-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 16px;
}
.instructions {
  font-size: 15px;
  color: var(--text);
  margin: 0;
}
.target-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 6px;
  background: var(--accent-glow);
  color: var(--accent);
  font-family: var(--mono);
  font-weight: bold;
}

.stats {
  display: flex;
  justify-content: space-between;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 20px;
}
.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.stat-value {
  font-size: 18px;
  font-weight: 600;
  font-family: var(--mono);
  color: var(--text);
}
.stat-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-dim);
}

.captcha-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 20px;
  padding: 0;
  border: none;
}
.cell {
  aspect-ratio: 1;
  background: var(--bg);
  border: 2px solid var(--border-bright);
  border-radius: 10px;
  color: var(--text);
  font-size: 2.5rem;
  font-family: var(--sans);
  font-weight: 800;
  display: grid;
  place-items: center;
  cursor: pointer;
  transition: background-color 0.1s, border-color 0.1s, transform 0.1s;
  padding: 0;
}
.cell:hover:not(:disabled) {
  border-color: var(--accent);
}
.cell:active:not(:disabled) {
  transform: scale(0.94);
}
.cell.selected {
  background: var(--accent-glow);
  border-color: var(--accent);
  box-shadow: inset 0 0 0 2px var(--accent);
}
.cell.fb-correct {
  background: rgba(126, 231, 135, 0.15);
  border-color: #7ee787;
  color: #7ee787;
}
.cell.fb-missed {
  border-color: var(--warn);
  box-shadow: inset 0 0 0 2px var(--warn);
  color: var(--warn);
}
.cell.fb-wrong {
  background: rgba(248, 81, 73, 0.15);
  border-color: var(--danger);
  color: var(--danger);
}
.cell:disabled {
  cursor: not-allowed;
  opacity: 0.8;
}

.speed-control {
  margin-top: 24px;
  display: grid;
  gap: 8px;
}
.speed-control-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}
.speed-control-head label {
  margin: 0;
  font-weight: 500;
  color: var(--text);
}
#speedLabel {
  font-family: var(--mono);
  font-size: 13px;
  color: var(--text-dim);
}
.speed-control input[type="range"] {
  width: 100%;
  accent-color: var(--accent);
  cursor: pointer;
}
.speed-scale-labels {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-faint);
}

/* ---- CAPTCHA leaderboard ---- */
details.leaderboard summary {
  cursor: pointer;
  list-style: none;
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 10px;
  font-weight: 600;
  color: var(--text);
}
details.leaderboard summary::-webkit-details-marker { display: none; }
details.leaderboard summary::before {
  content: "▸";
  color: var(--text-faint);
  margin-right: 8px;
  display: inline-block;
  transition: transform 0.12s ease;
}
details.leaderboard[open] summary::before {
  transform: rotate(90deg);
}
details.leaderboard .lb-table { margin-top: 16px; }
.lb-table td { color: var(--text); font-family: var(--sans); }
.lb-table td.mono { font-family: var(--mono); color: var(--text-dim); }
.lb-table tr.lb-row-me td {
  background: var(--accent-glow);
  color: var(--accent);
}

/* ---- Notes page ---- */
.composer-foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
}
.feed-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 10px;
}
.feed-head h2 { margin-bottom: 0; }
.feed { margin-top: 12px; }

.note-card {
  background: var(--bg-elev);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 14px 16px;
  margin: 12px 0;
}
.note-head {
  display: flex;
  align-items: baseline;
  gap: 8px;
  flex-wrap: wrap;
}
.note-author {
  font-weight: 600;
  color: var(--text);
}
.note-username {
  font-family: var(--mono);
  font-size: 12px;
  color: var(--text-faint);
}
.note-time {
  margin-left: auto;
  font-size: 12px;
  color: var(--text-faint);
  white-space: nowrap;
}
.note-body {
  margin: 8px 0 10px;
  color: var(--text);
  /* Plain text only: preserve the author's line breaks, and break anything
     (long tokens, pasted URLs) so a note can never overflow on mobile. */
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  word-break: break-word;
}
.note-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
.btn-small {
  padding: 5px 10px;
  font-size: 12px;
}

@keyframes note-in {
  from { background: var(--accent-glow); }
  to { background: var(--bg-elev); }
}
.note-new { animation: note-in 1.6s ease-out; }
@media (prefers-reduced-motion: reduce) {
  .note-new { animation: none; }
}

.new-pill {
  position: sticky;
  top: 64px;
  z-index: 5;
  display: block;
  margin: 0 auto;
  padding: 6px 14px;
  border-radius: 999px;
  background: var(--accent-dim);
  border-color: var(--accent);
  color: #04140a;
  font-weight: 600;
  font-size: 13px;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
}
.new-pill:hover { background: var(--accent); }
```

## `docs/ARCHITECTURE.md`

```markdown
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
```

## `README.md`

```markdown
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
The browser loads the app from the frontend's tunnel URL, then calls the API's
tunnel URL directly (real cross-origin requests — CORS is load-bearing, not
decorative). The API trusts no caller: it authenticates and authorizes every
request regardless of origin, so it is a genuinely reusable service and not
secretly coupled to this frontend.

For the full design rationale (why two tunnels, why a Bearer token rather than a
cookie, how the frontend learns the API URL at runtime), see
docs/ARCHITECTURE.md.

Quick start
bash
./scripts/run.sh
That one command:

regenerates the LLM context dump (export.sh),
runs the full test suite (build + vet + unit tests), and
builds the images and brings the stack up, printing the two Quick Tunnel
URLs when ready.
Open the Web app URL it prints, create an account, then add a passkey or
enable TOTP from Settings. Tear it down with:

bash
./scripts/down.sh
Prerequisites
Podman (with podman compose or podman-compose) or Docker (with
docker compose). Podman on Fedora is the primary target.
Outbound internet access (to pull base images and, on the first API build, to
resolve Go modules).
openssl (for secret generation) — present on essentially every Linux box.
You do not need Go, Node, or Caddy installed on the host — everything builds
and runs in containers.

Running several isolated instances
The Compose project name namespaces containers, network, and volumes, so just
give each instance a different name:

bash
./scripts/up.sh alpha
./scripts/up.sh bravo
Each gets its own database, its own generated secrets (in its own .env — note
that a single repo checkout shares one .env; use separate checkouts for fully
independent secret sets), and its own pair of tunnel URLs.

Telemetry (optional)
Point the stack at any OTLP/HTTP backend with a single DSN. For Uptrace Cloud:

bash
export UPTRACE_DSN="https://<token>@api.uptrace.dev"
./scripts/run.sh
Or use the standard OTEL_EXPORTER_OTLP_ENDPOINT / OTEL_EXPORTER_OTLP_HEADERS
variables. With nothing set, the API still logs structured JSON to stdout
(podman logs) and installs no-op trace/metric providers. The exporter is the
vendor-neutral OpenTelemetry SDK — no Uptrace SDK is imported, so switching
backends is a config change, not a code change.

Configuration
All configuration is centralized. On the Go side, every environment
variable is read in exactly one file, internal/config/config.go;
no other package calls os.Getenv. The CSP policy is centralized in the
frontend's Caddyfile and mirrored to the API for its
info endpoint. The full list of variables, defaults, and meanings is in
docs/CONFIGURATION.md and
.env.example.

The Content-Security-Policy ships in report-only mode
(Content-Security-Policy-Report-Only) with a strict self-only policy — no
third-party scripts, styles, images, fonts, or frames; everything is
self-hosted. Flip a single variable (GOTUNNELS_CSP_HEADER_NAME) to enforce it.
Because the app uses only external scripts/styles (no inline <script>, no
inline event handlers, no inline styles), it is already compatible with the
enforcing policy.

Testing and CI
Tests run through one script, locally and in CI:

bash
./scripts/test.sh all         # build + vet + unit tests
./scripts/test.sh unit        # just unit tests
./scripts/test.sh vuln        # govulncheck (reachability-aware)
./scripts/test.sh freshness   # bump deps, tidy, retest (drift check)
If you have Go installed it uses it directly; otherwise it runs the suite in a
golang container. The GitHub Actions workflows are intentionally thin — they
call these script functions rather than embedding logic:

ci.yaml — build, vet, unit tests on every push/PR.
govulncheck.yaml — vulnerability scan on push/PR and weekly.
dependency-freshness.yaml — weekly update-and-retest drift check.
Dependabot watches container base images and Actions versions; the Go module
graph is covered by govulncheck (call-graph aware, far less noise than
Dependabot's Go alerts).

Repository layout
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
A note on go.sum
go.sum is intentionally not committed on first import: the environment
this was authored in could not reach the Go module proxy, so the checksum file
could not be generated there. Everything is set up so this does not matter — the
API container build and every scripts/test.sh target run go mod tidy first,
which resolves and locks the dependency graph on any machine with network
access. After your first successful run you may commit the generated go.sum
for reproducible builds.

Security & demo caveats
This is a demonstration, deliberately simple in places:

Signup has no email/SMS verification — one step by design; not suitable
for real abuse prevention.
Cloudflare Quick Tunnels are a testing/dev feature with no throughput
guarantee and a URL that changes on restart. They are perfect for a demo and
the wrong tool for sustained production traffic; the architecture is built so
the ingress could later be swapped for a named tunnel or a real domain
without touching the app.
Argon2 parameters are modest to keep the demo snappy in a container.
LLM-generated code should be reviewed before you rely on it.
Security & demo caveats
License
AGPL-3.0-or-later. See LICENSE. If you run a modified version of
this software as a network service, the AGPL requires you to offer your users
the corresponding source.


**Correction to the README above:** I accidentally duplicated the `## Security & demo caveats` heading while transcribing — the actual file (in the zip and shown by the earlier `cat`) has it exactly once. Use the zip's `README.md` or simply delete that stray duplicated heading line before the `## License` section.

## Files that need NO changes

`docs/CONFIGURATION.md`, `.env.example`, `compose.yaml`, both Containerfiles, all `scripts/*`, all `.github/workflows/*`, `internal/config`, `internal/telemetry`, `internal/database` (incl. `migrations/embed.go`), `internal/auth`, `internal/activity`, `internal/health`, `internal/csp`, `frontend/Caddyfile` (its `try_files` already maps `/notes` → `notes.html` and `/captcha` → `captcha.html`), `frontend/js/common.js`, `config.js`, `csp.js`, `webauthn.js`, and the other `page-*.js` files, `go.mod`, `export.sh`, `LICENSE`.

## After committing

`./scripts/test.sh all` (the environment here can't reach the Go module proxy, so the compile step runs on your machine as usual), then `./scripts/run.sh` — migrations 0006/0007 apply on API boot. Sign in, and the nav grows Notes + CAPTCHA; open the notes page in two browsers to watch the polling reconciliation and the "New notes ↑" pill do their thing.


Your org is out of usage credits. We let your admin know. Switch models to continue chatting.
Switch models





Claude is AI and can make mistakes. Please double-check responses.


100
75
56
