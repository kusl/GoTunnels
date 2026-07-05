00
00
00

This is great progress. 
However, there are a few issues. 
First, I got some CORS error when signing up 
Turns out that was because I still had the old `.env` file
the workaround there was to delete the env file 
```bash
kushal@virginia:~/src/golang/GoTunnels$ cd ~/src/golang/GoTunnels/; time git status; time git fetch; time git status; time git remote show origin; time git pull; time git status; export UPTRACE_DSN="https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"; cat .env; rm .env; time bash scripts/up.sh; podman ps -a;
```
I was able to sign up to the website after that. 
Next problem was when I was in the captcha page, 
I noticed sync failed. 
upon closer inspection, I observed that the sync endpoint always returned 500 error 
```json
{"error":"internal server error"}
```
also please check that we are sending all logs, traces, spans, metrics to uptrace 
here is some guidance
```bash
# Uncomment the appropriate protocol for your programming language.
# Only for OTLP/gRPC.
#export OTEL_EXPORTER_OTLP_ENDPOINT="https://otlp.uptrace.dev:4317"
# Only for OTLP/HTTP.
#export OTEL_EXPORTER_OTLP_ENDPOINT="https://otlp.uptrace.dev"

# Pass Uptrace DSN in gRPC/HTTP headers.
export OTEL_EXPORTER_OTLP_HEADERS="uptrace-dsn=https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"

# Enable gzip compression.
export OTEL_EXPORTER_OTLP_COMPRESSION=gzip

# Enable exponential histograms.
export OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION=BASE2_EXPONENTIAL_BUCKET_HISTOGRAM

# Prefer delta temporality.
export OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=DELTA
```
You can see the code in `dump.txt` and my deployment at `virginia.txt`
Also please review all the tests once again. 
Among others, please check if the github actions actually test our container files and podman compose process. 

Please return FULL files for all files that need to change. 
























































/
Claude finished the response
This is great progress. 
However, there are a few issues. 
First, I got some CORS error when signing up 
Turns out that was because I still had the old .env file
the workaround there was to delete the env file 
bash
kushal@virginia:~/src/golang/GoTunnels$ cd ~/src/golang/GoTunnels/; time git status; time git fetch; time git status; time git remote show origin; time git pull; time git status; export UPTRACE_DSN="https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"; cat .env; rm .env; time bash scripts/up.sh; podman ps -a;

I was able to sign up to the website after that. 
Next problem was when I was in the captcha page, 
I noticed sync failed. 
upon closer inspection, I observed that the sync endpoint always returned 500 error 
json
{"error":"internal server error"}

also please check that we are sending all logs, traces, spans, metrics to uptrace 
here is some guidance
bash
# Uncomment the appropriate protocol for your programming language.
# Only for OTLP/gRPC.
#export OTEL_EXPORTER_OTLP_ENDPOINT="https://otlp.uptrace.dev:4317"
# Only for OTLP/HTTP.
#export OTEL_EXPORTER_OTLP_ENDPOINT="https://otlp.uptrace.dev"
# Pass Uptrace DSN in gRPC/HTTP headers.
export OTEL_EXPORTER_OTLP_HEADERS="uptrace-dsn=https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"
# Enable gzip compression.
export OTEL_EXPORTER_OTLP_COMPRESSION=gzip
# Enable exponential histograms.
export OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION=BASE2_EXPONENTIAL_BUCKET_HISTOGRAM
# Prefer delta temporality.
export OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=DELTA

You can see the code in dump.txt and my deployment at virginia.txt
Also please review all the tests once again. 
Among others, please check if the github actions actually test our container files and podman compose process. 
Please return FULL files for all files that need to change.

Show less

Diagnosed SQL parameter type bug, traced CORS/env persistence issues, and architected telemetry configuration
Diagnosed SQL parameter type bug, traced CORS/env persistence issues, and architected telemetry configuration
I need to review the project structure and identify the main issues: there's a CORS problem during signup that required deleting the .env file as a workaround, the sync endpoint is consistently throwing 500 errors on the captcha page, and I need to verify that all logging, tracing, spans, and metrics are being captured properly.

The user shared credentials in plaintext—a GitHub PAT and an Uptrace DSN—which should be rotated immediately per security best practices, though they mentioned not to worry about the PAT. I'll avoid embedding the PAT in any files and can use it only for read-only API calls via bash. Let me start by examining the dump.txt and virginia.txt files to understand the current codebase and deployment setup.


Checking project files exist and their sizes

Script
I should start by examining the structure of dump.txt to understand what I'm working with.


Listing all files in the dump to understand project structure

Script

Inspecting dump.txt format around a file marker to write an accurate extractor

Script

Viewing more of the dump format to find content delimiters

Script
I'm working through the file format structure to build an extractor that identifies the --- CONTENT --- marker following each # FILE: header and captures everything until the next file block begins, being careful to strip any trailing blank lines.


Checking the boundary between two files in the dump

Script

Extracting all files from dump.txt into a working repo directory

Script
Now I'm examining the deployment log to identify the specific errors, particularly the sync 500 error and CORS issue.

All 91 files extracted. Now I'll check the deployment log to see what errors occurred.


Reading the deployment log from virginia.txt

Script

Searching deployment log for errors, CORS, sync failures

Script
I'm spotting container name collision errors in the deployment logs. There's a stale tunnel URL lingering in the environment variables from an old .env file that needs to be cleaned up. Let me dig deeper into the logs around that section to understand the full scope of the sync failure.


Viewing the section around the old .env file and its CORS config

Script

Viewing the rest of the deployment log after the successful bring-up

Script
Now I'm diving into the captcha sync handler to track down what's causing that 500 error. I need to examine the captcha implementation, the store logic, and the migration that set up the schema to find where things are breaking.

I'm examining the captcha sync handler and store layer to locate the 500 error.


Reading the captcha handler code to find the sync 500 bug

Script

Finding captcha-related store functions

Script

Reading the SyncCaptchaStats SQL and surrounding store code

Script
I need to check the migration 0006 file to understand the captcha_stats table structure — the issue could be a mismatch between the INSERT statement and the actual schema, like incorrect column names or missing defaults.


Reading migration 0006 (captcha_stats table), 0007, and the embed logic

Script
I'm spotting a bug in the captcha stats insertion query where the manual solves parameter is being added to the total instead of used directly, and the parameters being passed seem to be cut off mid-way through. I'm working through how PostgreSQL resolves parameter types across the entire statement. Each parameter gets assigned a single type that applies everywhere it's used, with direct column assignments providing strong type hints. The tricky part is the order of analysis — PostgreSQL processes the VALUES list left to right, so when it encounters $4 + $5 in the total_solves position, both parameters are still unknown types. This creates an ambiguity with the + operator since there's no clear way to resolve unknown + unknown to a specific operator, which typically triggers a "operator is not unique" error.

This parse-time failure would explain why the sync endpoint consistently returns 500 errors — the SQL statement itself is malformed and fails during parsing before any execution happens. The unit tests wouldn't catch this because they only test the clamping logic in-process without hitting the database, so the SQL error never surfaces during testing. operator is not unique error when Postgres tries to resolve the + operator with two unknown parameter types. The extended protocol that pgx uses forces parameter type resolution at parse time, so every call to SyncCaptchaStats fails immediately with a 500 error. The solution is to add explicit casts like ($4::bigint + $5::bigint) to disambiguate the operator. I should also add casts in the UPDATE clause for consistency. Let me reconsider whether $2 and $3 actually need casts in the VALUES clause — they're being assigned directly to bigint columns, so assignment context should handle the type coercion fine. The real issue is the expression $4 + $5 where operator resolution happens before assignment coercion can kick in. I want to verify this behavior empirically if I can get Postgres running locally, and I should also check whether pgx uses prepared statements by default.

pgx v5 defaults to QueryExecModeCacheStatement, which means it's using the extended protocol and prepared statements — that would definitely trigger the unknown + unknown error at the parse level. I should search through store.go for other similar expressions with the pattern $ followed by  + $ to see if there are more places with this issue. Let me also check what's actually being tested in the test files, particularly around the captcha logic, and then look at how the .env file is being managed in the setup scripts since that was the workaround for the CORS issue.

The real problem with the .env file is that it was persisting stale tunnel URLs from a previous run — Quick Tunnels generate a new URL each time, so the old GOTUNNELS_RP_ID, GOTUNNELS_RP_ORIGINS, and GOTUNNELS_CORS_ALLOWED_ORIGINS values were causing CORS and WebAuthn failures. The fix should be to regenerate these tunnel-dependent values after discovering the new tunnel URL while keeping stable secrets like POSTGRES_PASSWORD intact. Let me check how the scripts currently handle this.

Now looking at the telemetry setup — they're using OTLP HTTP exporters for traces, metrics, and logs. Uptrace expects the endpoint at https://otlp.uptrace.dev for HTTP (or https://otlp.uptrace.dev:4317 for gRPC), with the DSN passed as an uptrace-dsn header. Let me examine how telemetry.go configures this.

I also need to check what environment variables are passed to the API container in compose.yaml, and then configure the exporters with gzip compression, exponential histograms, and delta temporality — which in the Go SDK means setting the temporality selector, aggregation selector, and compression options on the metric and trace exporters. Both env vars are supported by the Go OTLP metric exporters, so I'll set these defaults in compose.yaml and telemetry.go for the cleanest approach. Now I need to review the test setup and GitHub Actions workflows to see what's currently being tested and add coverage for the container files and podman compose.

I'll need to create a workflow that builds both Containerfiles with podman, spins up the full stack with podman-compose, waits for health checks, hits the health endpoint to verify everything's working, then tears it down. Since ubuntu-24.04 runners have podman preinstalled, I just need to install podman-compose via pip and handle the cloudflared container requirements.

The tricky part is that cloudflared needs outbound internet access, which might be flaky or rate-limited in CI. A better approach would be to skip cloudflared entirely in the CI test — I can use podman compose up -d db api frontend to start only the essential services while letting compose handle their dependencies. I should check the existing scripts first to see if there's already a CI mode, or if I need to write a targeted test workflow from scratch.

For verifying the API is healthy, since there are no published host ports by design, I'll need to either exec into a container to run a health check, or spin up a curl container on the same network to test the endpoint. Let me check the compose file and the API's main.go to see what health check mechanisms are available.

Now I need to review the telemetry configuration to confirm we're sending all logs, traces, spans, and metrics to Uptrace, and check how the UPTRACE_DSN is wired up in the compose environment.

Show less
Done
The store SQL has my prime suspect. Let me now read the telemetry, scripts, compose, and workflows.


Unraveling the .env lifecycle bug's cascading consequences
Unraveling the .env lifecycle bug's cascading consequences

Reading telemetry setup to check Uptrace OTLP export configuration

Script

Finding telemetry config parsing including Uptrace DSN handling

Script

Reading how UPTRACE_DSN is resolved into endpoint and headers

Script

Reading endpointFromDSN and compose.yaml env wiring

Script
I need to examine the up.sh and lib.sh scripts to see how the .env file gets stale with old tunnel URLs and how the API restarts with the newly discovered ones.


Reading up.sh to understand env generation and stale .env issue

Script
I need to clean up the extracted up.sh file by removing the dump footer that got included, and I'm realizing the real issue: when ensure_env runs and finds a stale .env from a previous execution, it's not handling that scenario properly.

Looking at the logs, the first run failed partway through with container name conflicts — the script never reached the steps where it would refresh the RP and CORS values, so those stale settings persisted. The project name gotunnels-novzg appears to be generated with a random suffix, and when containers with that name already existed from the failed run, the next attempt couldn't proceed. I'm noticing the collision errors for db and api containers happen within the same run, which suggests podman-compose might be attempting to recreate all containers even when using --no-deps, rather than respecting the existing ones from earlier steps. ves the stale GOTUNNELS_INSTANCE_ID from the previous run. Looking at the logs, the first run shows using existing /home/kushal/src/golang/GoTunnels/.env with project (instance): gotunnels-novzg, but the dumped .env after down.sh still has GOTUNNELS_INSTANCE_ID=default. This suggests the set_env_var call isn't properly updating the instance ID in the .env file, or there's something odd happening with how the variable gets written.

Actually, I think I see the issue now: if GOTUNNELS_INSTANCE_ID was already set to "default" in the stale .env (possibly from .env.example or an older generation), then when ensure_env sources it, the variable stays "default" because the export line uses ${GOTUNNELS_INSTANCE_ID:-$PROJECT}, which keeps the existing value. And looking at resolve_project, it seems to generate a random suffix like novzg or ekh each run unless GOTUNNELS_PROJECT is explicitly set, so the project names are different every time.

The container name collision on the first run—gotunnels-novzg_db_1 is already in use—suggests a leftover container from a previous session with the same random suffix, which is unlikely with 5-character randomness, OR it's the same-run double-create behavior from podman-compose. The error happens right after the build succeeds and before the database even starts, which points to a stale container from an earlier session still hanging around.

I'm wondering if the .env file persists the GOTUNNELS_PROJECT suffix across runs—if resolve_project stores gotunnels-novzg in .env, then re-running with that stale .env would reuse the same suffix and collide with containers still existing from the previous run. The dumped .env doesn't show GOTUNNELS_PROJECT, but down.sh might be removing it. I need to check lib.sh to understand how the project name and suffix are managed, and I should also re-extract up.sh properly since the previous dump got contaminated with trailing summary output.

Show less
Done
The extracted up.sh has the dump footer appended. Let me fix that, then read lib.sh.


Unraveling podman-compose's service creation and dependency behavior
Unraveling podman-compose's service creation and dependency behavior

Removing the dump summary footer from the extracted up.sh

Script

Reading lib.sh to understand env generation, project resolution, and helpers

Script
So the bug stems from .env persistence—ensure_env preserves the existing file to keep secrets safe, but resolve_project generates a fresh random project name each time since the stale .env still has GOTUNNELS_INSTANCE_ID=default. This leaves old tunnel configuration values like GOTUNNELS_RP_ID, GOTUNNELS_RP_ORIGINS, and GOTUNNELS_CORS_ALLOWED_ORIGINS from the previous run lingering in the file.

The container collision happens because up.sh only overwrites those stale values after discovering the frontend tunnel in step 5—but if the run fails before that point (like with the container name conflicts), or if only the database is started with up -d db, those old values stay put. Looking at the virginia.txt log, the error about gotunnels-novzg_db_1 being in use is puzzling since novzg was just resolved this run, unless up.sh ran twice with the same .env or there's a leftover container from a previous execution.

Actually, I think I see it now—podman-compose's --no-deps flag might not prevent creation of all services in some versions. So when step 3 tries up -d --no-deps frontend cloudflared-frontend, it could be creating the db again even though it already exists from step 2, causing the "name already in use" error. Then step 6 does the same thing, which explains why both db_1 and api_1 show as already in use even though api was supposedly only created in step 6 itself.

Show less
Done
Let me study the first failed run in the log more carefully to understand the exact failure sequence.


Scrutinizing database query execution and health check endpoints
Scrutinizing database query execution and health check endpoints

Reading the first failed run sequence in detail

Script
I'm seeing the actual behavior of podman-compose now—when running up -d with specific services, it creates the entire dependency chain, and even with --no-deps it still attempts to create all listed services plus their dependencies, generating "already in use" errors for ones that exist but continuing anyway. At step 3, it even created gotunnels-novzg unexpectedly.

The real issue is that --no-deps isn't being honored—podman-compose created and started the API container at step 3 before the frontend tunnel URL was even known. This is the CORS bug's root cause: the API was initialized with whatever environment variables existed at that moment (stale .env values or localhost defaults), and when step 6 tried to bring up the API again with --no-deps, podman-compose found it already running and just started it without recreating it. Unlike docker compose, podman-compose doesn't detect config changes or recreate containers based on environment variable updates, so the API kept running with the stale CORS origins captured from step 3. When the .env was deleted and regenerated with GOTUNNELS_CORS_ALLOWED_ORIGINS=*, the API that was already created at step 3 had already locked in its CORS configuration. So the real issue is that podman-compose creates the api container at step 3 when the frontend declares its dependency, and then at step 6 when we try to recreate it with the discovered tunnel URL, the container's already running so the create fails and start does nothing. The solution is to force-recreate the api container after we've set the environment variables with the actual frontend URL — using --force-recreate --no-deps flags on the up command, assuming podman-compose 1.x supports that option.

But there's a safer approach that doesn't rely on specific flag support: explicitly remove the api container if it exists before bringing it back up, which ensures it gets created fresh with the new environment variables. The tricky part is handling the dependency chain — api depends on db with a health check condition, and cloudflared-api depends on api, so I need to make sure the startup order works correctly when recreating these containers.

Actually, looking back at the logs, the health check didn't cause a hang, and podman-compose tends to ignore depends_on conditions anyway, mostly just ordering creation. So maybe the better strategy is to stop fighting the tool's behavior and instead always recreate api after discovering the URL, since that's when the environment changes. Frontend doesn't need the same treatment since it gets its config injected via podman cp later. The real issue is that podman-compose ignores both --no-deps and depends_on conditions, so I can't rely on those flags to control what gets created.

Writing bash script... After removing the stale api container, podman-compose up -d will recreate it fresh with the current environment variables. I could use --force-recreate, but explicitly removing the container with rm -f is more robust across different versions of podman-compose and docker compose. With docker compose specifically, the api container won't even exist at step 6 since --no-deps skipped it earlier, so the rm becomes a no-op. I'm also wondering about the db health condition when creating the api container at step 6 with --no-deps — that might be an issue to consider.

Actually, --no-deps skips starting dependencies, but the db is already running, so docker compose should still respect the depends_on conditions. That should be fine. Beyond the recreation fix, I should make ensure_env more robust by resetting tunnel-derived values to safe defaults on every run, since these are per-run values, not secrets. If a run crashes mid-way, stale tunnel values in .env could cause the api to use outdated origins on the next run, even though recreation at step 6 would fix it anyway. Being explicit about resetting these values makes the script more self-documenting and safer.

I'm setting the tunnel-derived environment variables to sensible defaults and wondering whether the wildcard CORS setting is appropriate. Since the auth uses Bearer tokens in the Authorization header rather than cookies, and sendBeacon can't carry Authorization headers, allowing any origin with Access-Control-Allow-Origin: * should be safe.

The real issue is that ensure_env gets called by multiple scripts, but only up.sh should reset per-run values like the tunnel URL. I need to call the reset function right after ensure_env in up.sh and re-export those variables to override what was loaded from the environment file.

I'm creating a helper function that resets the tunnel-derived values to safe defaults — the RP ID back to localhost, CORS origins to the local dev server, and clearing the tunnel URL itself. This way stale values from a previous run won't silently break things with opaque CORS errors.

Now I'm thinking through the sync 500 bug hypothesis: the issue might be with $4 + $5 when those parameters have unknown types. I should test this empirically by setting up a real Postgres instance and running the exact query through the extended protocol to see if type coercion is the culprit. $5` would work. Let me check database.go for the QueryExecMode setting.

For the Uptrace telemetry config, I need to verify the correct endpoint — the DSN resolves to https://api.uptrace.dev with the uptrace-dsn header, but Uptrace's actual OTLP/HTTP endpoint should be https://otlp.uptrace.dev, not api.uptrace.dev. Let me check their current documentation to confirm the right endpoint format. Fixing the endpoint mapping logic...

For self-hosted instances, I'll keep the scheme and host as-is since they typically listen on the same port. I also need to add the exporter tuning settings from the guidance—gzip compression is already handled by default, but I should set the exponential histogram aggregation environment variable. The Go OTLP metrics exporter should pick up OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION to use base2 exponential bucket histograms.

Rather than relying on environment variables, I'll configure these settings programmatically using the metric HTTP options with custom selectors for both temporality preference (delta) and histogram aggregation (exponential), while still allowing environment overrides for flexibility.

Implementing the selector functions... Adding config fields...

I should also update compose.yaml to expose the OTEL environment variables so users can override compression, temporality, and histogram settings, and make sure lib.sh writes these defaults. I'll set gzip as the default compression in compose. But there's another issue: UPTRACE_DSN only gets wired during initial .env generation, so if a user exports it in their shell after the .env file exists, load_env will overwrite it with whatever's already in the file. Writing shell script logic...

Now I'm thinking through the telemetry instrumentation — traces are flowing through the HTTP server with enriched spans, and metrics come from the OpenTelemetry meter plus HTTP metrics, but runtime metrics aren't currently included. Adding runtime instrumentation would mean a new Go dependency, which conflicts with the constraint of not introducing new dependencies.

The core issue is verifying that logs, traces, spans, and metrics all reach Uptrace. There's a potential endpoint mismatch (api.uptrace.dev vs otlp.uptrace.dev) that could prevent data from arriving, the metrics interval looks reasonable at 30 seconds, and logs are fanned out through a multihandler to OTLP. The question is whether all packages are actually using the logger from main or if some are bypassing it.

For traces and spans, the HTTP server is wrapped with otelhttp, but database operations aren't instrumented since pgx tracing isn't wired up. The shutdown sequence should flush telemetry on SIGTERM, and the insecure flag derivation handles HTTP schemes correctly for self-hosted setups. There's also a potential issue with how the endpoint URL is being passed to the trace exporter—if the URL lacks a path component, the exporter should append the default signal path automatically, but I need to verify this behavior against the documentation. Parsing the endpoint URL to extract the host and path components, then using a helper function to apply the default traces path if no path was explicitly provided. I should also check if the environment variable for metric temporality preference is supported in v1.44, though since I'm setting the selectors programmatically from the config file, that's less critical. For the GitHub Actions workflow, I need to examine the existing tests and add a new job that builds both Containerfiles with podman, spins up the compose stack with the database, API, and frontend services, verifies the API is healthy through the internal network, confirms migrations ran successfully, and checks that Caddy is applying the CSP headers correctly before tearing everything down.

I'll create a dedicated script at scripts/ci-compose-test.sh to handle this so the workflow just calls it — that way developers can reproduce the exact same test locally with a single command. On Ubuntu 24.04, podman is already available but I need to install podman-compose via pip since the apt version is outdated and has known issues; the pip version gives me 1.5.x which is more reliable and closer to what users on Fedora would have.

For health verification, the api container is built from scratch so there's no curl or wget available to probe the health endpoints — I'll need to use a different approach than exec probes since the container only contains the binary itself.

Instead, I can run a temporary curl container on the same network to check the health endpoint, or reuse the frontend container which has wget available through its alpine base. The frontend approach is cleaner since it avoids pulling an extra image. I should also verify the frontend is serving properly by checking if it responds to requests on its configured port.

Now I'm deciding whether to keep the CI simple with a single solid job or test against multiple podman compose configurations. Before committing to that, I need to review the existing CI setup, test scripts, and all the relevant code files — the health checks, server logic, database setup, container configurations, and middleware — to understand what's already in place.

Show less
Done
Now let me read the remaining key files: workflows, Containerfiles, server, database, health, main, Caddyfile, and the CORS middleware.


Diagnosed database query protocol issue and mapped health endpoint verification tasks
Diagnosed database query protocol issue and mapped health endpoint verification tasks

Reading CI workflow, test script, and Containerfiles

Script
The distroless API container won't support exec probes since it lacks a shell, so I'll need to handle health checks through the frontend container with wget or set up a separate curl request on the same network. Let me trace through database.go to understand the pgx query execution, check health.go for the endpoint configuration, and review main.go, server.go, and the CORS middleware setup along with the frontend Caddyfile.


Reading database pool setup, health endpoints, and Caddyfile

Script
I'm pretty confident the issue is with pgx's default query execution mode using the extended protocol, which causes the server to fail on operator resolution when it encounters $4 + $5 with unknown parameter types. Let me verify this empirically by testing against a real PostgreSQL instance before I modify the SQL. I also need to check the exact health endpoint route paths in server.go.

I'll check the routes in server.go to understand how pgx handles the extended protocol, then test the SQL bug against a real PostgreSQL instance.


Reading server routes and checking health paths

Script
Done
Now let me empirically verify the sync 500 hypothesis with a real PostgreSQL before touching the SQL.


Verifying Uptrace endpoint configuration against current documentation
Verifying Uptrace endpoint configuration against current documentation

Installing PostgreSQL to empirically verify the sync SQL bug

Script

Installing PostgreSQL as current user (likely root)

Script

Updating apt indexes then installing PostgreSQL

Script

Starting PostgreSQL and verifying it runs

Script
PostgreSQL 16.14 is already running with the same version as the project, so I'm setting up the exact scenario by creating the users and captcha_stats tables from the migrations, then preparing a statement without explicit parameter types to match how pgx handles unknown OIDs.

I've set up PostgreSQL 16 to match production and I'm now testing the exact failing statement using pgx's extended protocol with undeclared parameter types.


Reproducing the exact production SQL with untyped parameters like pgx sends

Script
I've pinpointed the exact issue: when pgx v5 sends a Parse message with undeclared parameter types in QueryExecModeCacheStatement, PostgreSQL can't resolve the unknown + unknown operator in the VALUES clause at $4 + $5, causing the Parse to fail and returning a 500 error from SyncCaptchaStats.

The handler then blankets every error with {"error":"internal server error"}, so every sync consistently fails. The fix is to add explicit ::bigint casts to resolve the operator ambiguity.

I've confirmed the issue: PostgreSQL's Parse fails on the untyped + operator when pgx sends parameters without declared types via the extended protocol, causing every sync to return 500. Now I'll verify the fix.


Verifying the cast fix works for both insert and update paths

Script
The fix is working correctly — the insert and update paths both preserve the expected values and semantics. Now I'm scanning through the store queries to find similar issues where arithmetic or logical operators are used between untyped parameters.

The fix is working correctly with semantics preserved. Now I'll search the codebase for other instances of this same bug pattern.


Searching for other param-vs-param operator expressions in SQL

Script
I'll cast that expression for clarity even though the left-associativity handles it correctly. Now I'm looking at the CORS and environment variable issue — the logs show that podman-compose is ignoring the --no-deps flag when bringing up the frontend and cloudflared-frontend services.

The api container gets created as a dependency at that step with whatever environment was set at that moment (either stale values from .env or fresh defaults), and then when the api service is explicitly brought up later, podman-compose doesn't recreate it since there's no config-hash diffing — it just tries to create again, fails with "already in use" errors, and moves on. So the api never picks up the discovered tunnel URL for the RP_ID and CORS settings.

This explains the user's CORS failure with a stale .env pointing to a dead tunnel URL, and why deleting .env made signup work with the permissive CORS=* default. But now passkeys are actually broken on the deployed instance since the RP_ID is set to localhost while the page is served from trycloudflare.com — the WebAuthn validation will fail on origin and RP mismatch. The user hasn't noticed because they signed up with a password instead.

I need to fix this by removing any pre-existing api container before bringing it up in step 6, forcing a recreation with the corrected environment. I should also reset the tunnel-derived .env keys at the start of every up.sh run to prevent stale values from leaking in, which also handles crashes between steps. There's also an issue where exported UPTRACE_DSN gets clobbered by an older empty value in .env during the load — I need to persist caller-exported telemetry variables into .env before loading it.

For the Uptrace endpoint mapping, the current code extracts https://api.uptrace.dev from the DSN, but according to the guidance, OTLP/HTTP should target https://otlp.uptrace.dev instead. The code uses HTTP exporters, so sending them to api.uptrace.dev is wrong and likely why telemetry isn't working.

Let me examine the endpointFromDSN function to see how it parses the DSN and whether it's stripping the ?grpc=4317 query parameter correctly.

Show less
Done
Only the one spot. Now let me see the full endpointFromDSN to assess the Uptrace endpoint bug:


Reconciling sed extraction gaps in signup handler code
Reconciling sed extraction gaps in signup handler code

Reading endpointFromDSN implementation and its tests

Script
I'm noticing a discrepancy between what the DSN parser returns (https://api.uptrace.dev) and what Uptrace's documentation specifies as the OTLP/HTTP endpoint (https://otlp.uptrace.dev). I should verify against Uptrace's current documentation to confirm whether api.uptrace.dev also handles OTLP/HTTP requests or if otlp.uptrace.dev is the correct endpoint.

I found a discrepancy between the DSN endpoint and Uptrace's documentation—I'll check their current docs to verify the correct endpoint.


Uptrace OTLP endpoint otlp.uptrace.dev Go opentelemetry configure DSN
10 results


OpenTelemetry Go distro for Uptrace | Uptrace
uptrace.dev

Monitor OpenTelemetry Node.js with Uptrace | Uptrace
uptrace.dev

Ingesting telemetry using OpenTelemetry Distros for Uptrace | Uptrace
uptrace.dev

Direct OTLP Configuration for OpenTelemetry Python | Uptrace
uptrace.dev

OpenTelemetry NestJS Instrumentation | Uptrace
uptrace.dev

Getting Started with the OpenTelemetry Collector | Uptrace
uptrace.dev

Ingesting telemetry using OpenTelemetry Collector | Uptrace
uptrace.dev

OpenTelemetry Collector Configuration Tutorial | Uptrace
uptrace.dev

OpenTelemetry Java for Uptrace | Uptrace
uptrace.dev

otlp-metrics command - github.com/uptrace/uptrace-go/example/otlp-metrics - Go Packages
pkg.go.dev
I'm noticing that Uptrace's official documentation consistently specifies https://api.uptrace.dev as the OTLP/HTTP endpoint across multiple language implementations—Node.js, Python, and the collector configuration all reference this same endpoint.

Looking at the checklist against what's currently configured: the endpoint derivation is correct, headers with the DSN are set, and gzip compression is applied by default through the telemetry resolver. However, exponential histograms aren't being configured—the OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION=BASE2_EXPONENTIAL_BUCKET_HISTOGRAM environment variable isn't set anywhere in the compose setup, and the code doesn't specify an aggregation preference for metrics.

The same issue applies to delta temporality—OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=DELTA is missing. My plan is to add MetricsTemporality and MetricsHistogram fields to the TelemetryConfig struct, read them from environment variables with sensible defaults (using the spec-standard lowercase values and accepting any case), then wire them into the otlpmetrichttp exporter using the appropriate selector options from the pinned otel-go v1.44.0 version.

I'm trying to figure out if I can actually fetch and verify the exact option signatures from the Go module source on GitHub by configuring the Go proxy to bypass the sum database, since the standard proxy and sum server aren't in the allowed domains list. strumentKindCounter, InstrumentKindObservableCounter, InstrumentKindHistogram, InstrumentKindUpDownCounter, InstrumentKindObservableUpDownCounter, InstrumentKindObservableGauge, and InstrumentKindGauge. I also need to verify the aggregation options like sdkmetric.AggregationBase2ExponentialHistogram` with its configuration parameters and the default aggregation selector. Then I'll map out which instrument kinds should prefer delta temporality versus cumulative based on the OTLP spec.

I'm also planning to support three aggregation modes: delta, cumulative, and a lowmemory option that uses delta for counters and histograms but cumulative for observables. After that, I need to tackle the CORS and API recreation fixes, handle the ensure_env corrections, and get the CI working properly. I should run the full test suite with go vet and go test, validate shell scripts with bash and shellcheck, check YAML with Python, and potentially review the page-captcha.js error handling around that 500 error.

The frontend sync failure was actually a server-side issue, so once that's fixed the captcha page should report correctly — no JavaScript changes needed there. I also need to update the GitHub Actions workflow to add a new container testing job that detects the runtime and runs container builds with compose.

Writing container test script... can parse the token from signup responses using grep, then verify the frontend serves index.html with proper CSP headers. After that, clean up the containers and volumes. I'll keep the JSON parsing simple with a grep pattern, and add compose validation using Python's yaml module — though I'll keep that in a separate container test script rather than cluttering test.sh. Now I'm wondering if ubuntu-latest has podman available — checking the runner images, podman 4.9.3 is preinstalled, but podman-compose isn't, so I'll need to pip install it. The tricky part is that podman compose looks for an external provider like docker-compose or podman-compose in PATH, and GitHub runners have Docker Compose v2 as a plugin, but I need to figure out which one podman will actually find first.

To keep things simple and match the user's environment exactly, I should explicitly set COMPOSE=podman-compose in the CI script, since the detect_runtime function already respects pre-set CR and COMPOSE variables.

For the rootless setup in GitHub Actions, the runner user is non-root with rootless podman already configured, which mirrors the user's Fedora environment and naturally reproduces the health timer issue that wait_healthy handles. The container job needs to skip setup-go since the build happens inside the golang builder image instead.

Updating up.sh: I'm replacing step 6 to restart the API with correct reverse proxy and CORS settings plus its tunnel, accounting for podman-compose's behavior where it doesn't honor --no-deps, so the frontend and cloudflared-frontend containers were already created and started in step 3.

Writing shell script logic... I'm realizing that podman rm -f completely removes the container, so compose will recreate it even with the restart policy. The "already in use" warnings for the db at steps 3 and 6 are just noise from podman-compose itself—not worth suppressing. For the ensure_env function, I need to handle persisting any caller-exported telemetry overrides before loading the environment file so that the .env values don't override them. Adding more OTEL configuration variables to the template and caller-override loop, then creating a new helper to reset runtime environment variables that gets called in up.sh after ensure_env, plus updating compose.yaml to pass through the OTEL exporter settings to the api service. Writing selector functions... Writing aggregation selector... I'm realizing that GOPROXY=direct won't work for vanity import paths like go.opentelemetry.io since those domains aren't in the allowlist and require fetching the go-import metadata first. I need to find another way to verify the compilation without relying on those restricted domains.

I'll fetch the specific source files directly from GitHub raw content instead — checking the otel-go v1.44.0 exporter config files, aggregation types, and instrument kinds to verify the API signatures match what I'm expecting, particularly confirming WithEndpointURL and the histogram aggregation options. raw.githubusercontent.com supports the explicit /refs/tags/ syntax to disambiguate tag names containing slashes, so I can construct the URL as /open-telemetry/opentelemetry-go/refs/tags/exporters/otlp/otlpmetric/otlpmetrichttp/v1.44.0/ followed by the file path. The PAT should only be used for API calls that benefit from authentication, not for raw.githubusercontent.com which doesn't require it. I'll keep the token out of any committed files and use it only as a transient header in curl requests. I should also remind Kushal to rotate the PAT and check on the Uptrace DSN in virginia.txt since that file might end up in the repo.

For the GitHub Actions workflows, I need to add a container testing job to ci.yaml that installs podman-compose and runs an end-to-end test of the compose stack.

The script should set default values for the container runtime and compose command, but let the detection logic run normally so it works both locally and in CI — on the GitHub runner, podman compose version will work fine without forcing anything.

Writing shell script... I'm checking the exact log message that the API outputs — the wait_for_log function is looking for 'http server listening' but it might be timing out because either the message wording in main.go doesn't match, or the API already logged it before the grep started searching through the logs.

Now I'm setting up a way to test the frontend's health by running wget commands inside the caddy container to verify the API is reachable through the frontend network, including testing POST requests with JSON headers and body data using busybox wget's --header and --post-data flags.

I need to check the signup endpoint's expected field names and response format from the auth handlers, then chain together tests for captcha sync with a Bearer token, stats verification, notes CRUD operations, and preferences endpoints. For the frontend itself, I'll use wget with the -S flag to capture response headers and verify the CSP header is present, piping stderr to stdout so I can grep for it. All assertions will use grep -q for silent matching, and on failure I'll dump the full response for debugging.

For cleanup, I'm wrapping the teardown in a trap that calls the down.sh script with error handling. I also need to look at main.go to understand the "http server listening" log message that the wait_for_log function is checking for, verify the CORS middleware behavior to confirm it sends the right headers, and check what the default CORS configuration is in config.go.

Show less
I've confirmed the OTLP endpoint is correct, but the guidance is missing delta temporality and base2 exponential histogram settings. Now I'll check the API startup logs and auth request/response shapes to debug the CI smoke test timeout.


Checking startup log line, auth API shapes, CORS default, and down.sh

Script
The log line matches the grep pattern, so the timeout in production must stem from something else. The api container was started much earlier during tunnel polling, and its logs should contain the "http server listening" message in JSON format. When I grep for it with podman logs, it should match the "msg":"http server listening" field — unless there's something about how the output is being piped or formatted that's preventing the match. The API process might be crash-looping due to the stale environment variables, but that doesn't explain why requests worked later. The database volume is project-specific and freshly initialized, so the connection should work with the current password from the environment.

Looking at the successful g1gzbw run, it also showed the same "API listening log not observed" warning within 60 seconds, yet the app functioned fine — the user was able to sign up. So the API was definitely listening. The issue might be that cid_of is picking the wrong container when multiple containers with the same service label exist in the project, since the grep uses -a flag and takes the first match. At step 6, there was already a message about "api_1 already in use," suggesting multiple containers could be present.

I'm also considering whether the log line appears before podman logs actually starts capturing output, or if there's something about how podman-compose manages cgroups that prevents logs from being read properly — though step 10 successfully read cloudflared logs, so that seems less likely. ARTED (podman-compose creates deps without starting when they're not in the service list?) and then step 6's "up api" hit "already in use" on CREATE and... did it START the existing one? The output at step 6 (novzg): after errors, it printed gotunnels-novzg_db_1 and gotunnels-novzg_api_1 — podman-compose prints container names when it STARTS them (podman start outputs the name/id). So step 6 DID start db (no-op, running) and api. So api started at step 6... wait but then it started AFTER the env export — with the NEW env? NO — the container was CREATED at step 3 (its config/env frozen at creation); step 6 merely started it. Podman env is fixed at create time. So api runs with step-3 env. Consistent with my analysis.

But then: api started at step 6, wait_for_log at step 7 polls for 60s... api boots in ~1-3s (db ready), logs the line. Why not found? The printed names at step 6 are podman start outputs, so api definitely started at the beginning of step 6. Maybe the issue is the pattern matching inside the JSON log format — the substring 'http server listening' should be there, and grep -Eq should catch it. Let me check if wait_for_log itself has a bug.

Ah, I see it now — when running podman logs | grep -q, grep exits 0 on the first match and closes the pipe, which sends SIGPIPE to podman logs. Under set -o pipefail (inherited from up.sh), the pipeline returns the rightmost non-zero exit code. Grep exits 0 but podman logs exits 141 from SIGPIPE, so the pipeline status becomes 141, making the if condition false even though the match succeeded. That's the bug.

The race condition depends on log volume — small logs finish before grep exits (status 0), but verbose logs like API request logging trigger the SIGPIPE kill. This is exactly the pattern the team already fixed in poll_tunnel_url using || true and full draining, and it matches the documented lesson about set -euo pipefail with pipes silently killing scripts when commands like head -1 close early.

The real issue is that podman (written in Go) treats EPIPE as a fatal error and exits nonzero, so under pipefail the whole pipeline fails. The fix wraps the logs command with || true to neutralize its exit status, letting the pipeline status depend entirely on grep's result instead. The same wait_for_log issue appears in the CI script too, so I'll apply the same fix there. Let me quickly test the SIGPIPE behavior empirically with a simple bash command under pipefail to confirm my analysis. After that, I need to look at the auth request shapes — specifically what fields signupRequest needs — by checking the handlers for signup and login logic, plus the CORS and error handling in httpx.

Show less
The log pattern matches, but wait_for_log failing in both runs suggests a SIGPIPE issue where grep -q exits early and kills the pipe. I'll verify this empirically.


Reproducing the SIGPIPE+pipefail bug in wait_for_log and validating the fix

Script
So the "fixed shape" approach also failed because SIGPIPE terminates the chatty function itself — the subshell on the left side of the pipe receives SIGPIPE when echo tries to write, and while || true should catch that exit code, it's actually applying to the group's exit status rather than preventing the signal from killing the process. The group exits cleanly with 0, pipefail sees matching exit codes from both sides, and grep succeeds, but the underlying issue is that the signal is still terminating the function before the error handling can take effect.

The real solution is to avoid triggering SIGPIPE in the first place by fully draining the input — either by removing the -q flag so grep reads all input and prints matches to /dev/null, or by redirecting output explicitly so the pipeline doesn't close prematurely.

The naive fix doesn't work because SIGPIPE kills the subshell before the fallback can execute, so I'll use the project's established pattern of fully draining the stream instead.


Validating the full-drain grep fix under pipefail

Script
The drain fix resolved that SIGPIPE race condition — the verbose JSON logging on every request was triggering it consistently, which explains why the API listening log message kept appearing even though the service was working fine. Now I'm reviewing the auth handler's signup and login request/response structures, the CORS middleware setup, and the data shapes for preferences and notes to prepare the CI smoke test.

Done
Drain fix validated. Now the auth/notes/prefs API shapes for the CI smoke test:


Verifying endpoint URL path fallback behavior systematically
Verifying endpoint URL path fallback behavior systematically

Reading signup/login request-response shapes and CORS middleware

Script
I'm spotting the signup handler around line 163 where it records the user ID, username, and signup event, then returns a session response with a token. The extraction skipped some middle lines covering password validation and user creation, but that's already been reviewed. Now I need to cross-check the OpenTelemetry v1.44.0 metric API signatures to make sure they align with what's being used.

I've confirmed the signup response returns a token. Now I'll check the OTel v1.44.0 API signatures against the GitHub source since the module proxy isn't available.


Verifying otlpmetrichttp option functions exist in pinned v1.44.0

Script

Finding the module tag for otlpmetrichttp v1.44.0 via GitHub API

Script
I need to use the PAT the user provided instead of relying on the environment variable, so I'll pass it directly in the curl command as an inline variable without writing it to any deliverable files.


Locating the exact otlpmetrichttp v1.44.0 tag using the provided read-only PAT

Script

Fetching otlpmetrichttp options.go at the v1.44.0 module tag

Script

Inspecting error and listing the exporter directory at the tag

Script

Fetching config.go where the options live for otlpmetrichttp

Script

Verifying TemporalitySelector, AggregationSelector, exponential histogram struct, and instrument kinds in sdk/metric v1.44.0

Script
I'm verifying the API signatures against the pinned OpenTelemetry metrics version, checking that the temporality and aggregation selector options are properly defined with the correct function signatures.

Done
All signatures verified. Let me confirm the WithEndpointURL path behavior and metricdata temporality constants:


Verified SDK metric selectors exist, enabling telemetry implementation
Verified SDK metric selectors exist, enabling telemetry implementation

Checking WithEndpointURL path handling and temporality constants

Script
I'm checking whether WithEndpointURL properly handles empty paths by falling back to the default /v1/metrics signal path through the cleanPath function, and verifying that the telemetry implementation is actually working end-to-end as expected.


Verifying empty URL path falls back to the default /v1/metrics path

Script

Checking whether NewHTTPConfig normalizes an empty URL path to the default

Script
The path normalization is working as expected—when no explicit path is provided, it defaults to /v1/metrics, and the exporters correctly POST to the Uptrace endpoints with the DSN header. The only remaining gaps are what the user actually needs to implement.

I need to tackle four things: fix the SQL casting issue in the captcha stats sync, add a regression test via the CI smoke test that explicitly exercises the sync endpoint, extend the telemetry config to support delta temporality and exponential histograms with proper environment variable parsing and case normalization, and update the config tests accordingly.

Then I'm wiring up the temporality and aggregation selectors in the telemetry module with logging of the effective settings, fixing the shell script's SIGPIPE issue in the log-waiting function and ensuring environment variables persist before loading from .env so an empty DSN can't override what's already exported, adding a helper to reset tunnel-derived values between runs, updating the startup script to call that reset and recreate the API container so it picks up the discovered RP and CORS settings, and passing through the OTEL compression variable in the compose file.

Next I need to document the new telemetry variables in the example env file, create a new end-to-end smoke test script that spins up the full stack locally without tunnels and validates the signup flow including captcha sync and CSP headers, add a CI job to run that test with podman-compose, update the configuration docs to cover the new telemetry options, and then run the full validation suite across all changed files to check formatting and syntax. I need to verify the telemetry section in CONFIGURATION.md and the captcha frontend logic, check that httpx.DecodeJSON aligns with the auth implementation, and refine the up.sh script — particularly how wait_for_log resolves the container ID on each poll after the api recreation to ensure it catches the listening line reliably.

Now I'm thinking through whether GOTUNNELS_INSTANCE_ID should be reset — it shouldn't, since it's a stable user-facing identifier. I should also check if run.sh needs the telemetry variable passthrough for the bare go run dev path, and review the remaining GitHub workflows to see if they actually test the container and podman-compose setup like the user asked.

Show less
The endpoint derivation is correct—the real gaps are just the temporality and histogram settings. Now I'll check the remaining files and local toolchain.


Reading config tests and docs, installing validation tools

Script
I've got the tools set up and I'm ready to write out all the changes. There are five files to update: the store module needs a SQL cast fix, the config module needs new fields for metrics temporality and histogram settings along with corresponding tests, the telemetry module needs selector logic and a log line, and the shell script library needs a small adjustment.

Now I'm looking at the remaining files to create and modify: the main scripts need updates to call the reset function and recreate the API, there's a new end-to-end container test to write, the compose file needs to pass the three OpenTelemetry variables, the example environment file needs documentation for the new variables, the CI workflow needs a containers job, and the configuration docs need telemetry table rows. Before I start writing everything out, I should check the current .env.example and run.sh to stay consistent, then validate the Go code with formatting and vet checks. Since the config package only imports stdlib, I can compile and test it standalone to verify the shape of the temporality and aggregation selector logic.

For the remaining files that have external dependencies—store.go imports pgx, telemetry.go imports otel, captcha_test imports otel—I can't compile them in isolation, but I can use gofmt and careful diffs to ensure safety. The store.go fix modifies the SQL query for SyncCaptchaStats to properly calculate totals and individual counts...

Finishing the SQL query and adding a comment about the parse failure, then delivering the full store.go file via string replacement. I'll stick with pasting the full files directly in the chat since that's his established workflow — he copies from the response and commits them directly. Now I'm mapping out the specific changes needed, starting with config.go where I need to add metrics export configuration fields. Writing config initialization... Writing test cases... Adding metric configuration options... Finishing the condition check to verify the container is running and the log pattern matches, then looping with a one-second sleep until timeout. Then I'm updating the ensure_env function to add OpenTelemetry configuration variables to the template and handle caller overrides before loading the environment. I need to ensure that environment variables passed by the caller take precedence over any stale values left in the .env file, so I'm iterating through the key telemetry settings to preserve their current values before loading from the file. Writing shell script documentation... Writing shell script setup... Adding OpenTelemetry configuration...

Now I'm sketching out a bash script for the CI container test that sources utility functions, detects the runtime environment, and resolves the project name—though I notice the resolve_project function appears to lowercase the input.

Setting up a cleanup trap that captures container logs for the database, API, and frontend services if the test fails, then tears down the Docker Compose stack. The test itself starts by building the images, then bringing up the database service and waiting for it to be healthy before proceeding with the remaining steps.

Continuing with starting the API and frontend services without their dependencies, then waiting for the API to log its listening confirmation. I'm realizing that podman-compose's --no-deps flag won't prevent the database from being created again since the dependency graph flows the other way — frontend and API depend on the database, not the reverse — so cloudflared services won't be spun up either since they're not explicitly requested.

Now I'm setting up an HTTP driver that executes wget commands inside the frontend container to make requests to the API, building out helper functions that handle different HTTP methods and optional authentication tokens.

The issue is that busybox wget doesn't support PUT or DELETE methods—it only handles GET and POST through --post-data. Since the preferences endpoint uses PUT and the notes deletion uses DELETE, I need to decide between pulling in a curl container for full HTTP verb support or limiting test coverage to just GET and POST operations.

I'm going with option (a): spinning up a curl container on the compose network with proper assertions using -f for HTTP error codes and --write-out for response details. The curl image is small (~4MB) and a one-time pull. The key question is whether podman-compose's named network (<project>_default) supports DNS resolution of service names in rootless mode—it should, since podman uses internal DNS for that.

Writing the curl helper... Writing test assertions... Writing shell tests... Writing database checks... Deciding whether to validate the compose file with podman-compose config — it's supported but might introduce quirks, so I'll skip it since the build and up commands themselves serve as validation. Now I'm setting up the CI job that installs podman-compose via pipx, checks the runtime versions, and...

Writing CI steps... The user wants me to review tests and GitHub Actions, with a README update being optional. I should check if the README has a scripts section with a natural place to mention the ci-container-test.sh script—if so, I'll add it; otherwise I'll skip it to keep scope manageable. Then I need to validate the preferences and notes shapes, CaptchaStats JSON tags, OriginAllowed configuration, .env.example, and potentially the README scripts mention.

Show less
I'm checking the JSON schemas for the CI smoke test, the environment example file, and the README scripts section.


Reading JSON struct tags for captcha, prefs, and notes

Script
I'm picking up where I left off with 11 files to write. The repo is already extracted at /home/claude/gotunnels/repo/, and all root causes have been verified, so I can jump straight into the work without waiting for confirmation. Let me start by checking the current state of the working directory and reviewing the key files I need to modify.


Verify extracted repo state after compaction

Script
The repo structure looks good — scripts directory contains the shell utilities and .github/workflows has the CI configuration files. Now I'm pulling up store.go to see what needs to be modified.


View the SyncCaptchaStats function to apply the bigint cast fix

Script
I'm applying the bigint casts to store.go and adding a comment to explain the fix.


Fix the captcha sync 500: explicit bigint casts so Postgres can type $4 + $5 under the extended protocol

store.go
+32
-22

Locate telemetry config sections to modify

Script

Read full telemetry config section

Script
I'm updating config.go to add MetricsTemporality and MetricsHistogram fields to TelemetryConfig, then parsing them in resolveTelemetry with lowercase normalization alongside the Compression field.


Add metrics temporality + histogram aggregation fields to TelemetryConfig

config.go
+19
-9

Parse the two new OTEL metrics env vars (lowercase-normalized) and lowercase compression

config.go
+36
-23

Read existing config tests to extend them

Script
Adding new tests to config_test.go...


Add tests for the new metrics temporality/histogram defaults and overrides

config_test.go
+57
-14

Read telemetry.go in full before wiring temporality/aggregation selectors

Script
I need to update telemetry.go by adding the necessary imports for strings and OpenTelemetry metric data, then implement selector functions for temporality and aggregation behavior. These will be wired into the metric HTTP options and logged when telemetry is enabled. Writing aggregation selector... I need to check the go.mod file to confirm which version of the SDK is pinned and verify that DefaultAggregationSelector is available in it.


Confirm pinned otel versions from go.mod

Script

Verify DefaultAggregationSelector and DefaultTemporalitySelector exist in pinned sdk/metric v1.44.0

Script
I've confirmed that both DefaultTemporalitySelector and DefaultAggregationSelector are available in the current version of sdk/metric, so I can proceed with updating the telemetry.go file.

Done
Both default selectors confirmed in the pinned v1.44.0 source. Now wiring telemetry.go:


Refactoring up.sh script and resolving podman-compose dependency issues
Refactoring up.sh script and resolving podman-compose dependency issues

Add strings + metricdata imports to telemetry.go

telemetry.go
+24
-22

Log the effective metrics settings in the telemetry-enabled line

telemetry.go
+7
-4

Wire temporality + aggregation selectors into metric exporter options and add the selector functions

telemetry.go
+77
-13
Looking at the changes needed for scripts/lib.sh: fixing the wait_for_log drain issue, updating ensure_env to handle caller overrides for six telemetry variables before loading the environment plus three new OTEL variables in the template, and adding a new reset_runtime_env helper function. Let me examine the current lib.sh to understand what needs to be modified.


Read full lib.sh before editing

Script
I'm working through the lib.sh changes, starting with the wait_for_log fix. The current implementation uses grep -q which exits immediately on the first match, causing a signal issue. I need to switch to a full-drain pattern that pipes the logs through grep and redirects to /dev/null, then neutralizes the exit code with || true to prevent the script from failing.

Writing the corrected function...

The real issue was that grep with -q exits immediately on the first match, leaving the logs process hanging with SIGPIPE, which under pipefail makes the whole condition fail even when the pattern exists. By redirecting to /dev/null instead, grep reads the entire stream to completion, so the logs process exits cleanly and the condition works correctly. Now I'm moving on to the next issue with ensure_env caller-override.

The problem is when .env already has an empty UPTRACE_DSN= and the user exports a DSN in their shell, load_env clobbers it. The fix is to preserve any caller-exported telemetry variables before loading .env by saving those six specific variables into the file first.

I'll also add three new OTEL configuration variables to the generated template with sensible defaults: compression set to gzip, metrics temporality preference to delta, and histogram aggregation to base2 exponential bucket histogram.

Now I'm working through the ensure_env function logic, which checks if the environment file exists and generates it if needed, otherwise reuses the existing one. The key part is preserving any telemetry settings that were exported in the caller's shell before load_env runs, since an existing .env file with empty values would otherwise silently override those exports and disable telemetry without any warning.

Writing shell variable handling... Writing reset function...

Show less

Add 3 new OTEL vars to the .env template and the caller-override persistence loop in ensure_env

lib.sh
+56
-31

Add reset_runtime_env helper after load_env

lib.sh
+32
-9
I caught a syntax error in the bash script where I accidentally used // for a comment instead of #, which would break the whole thing — need to fix that.

Done
I introduced a // comment typo in that bash block — fixing:


Investigating sync endpoint failures and Uptrace integration
Investigating sync endpoint failures and Uptrace integration

Fix // comment typo that would be a bash syntax error

lib.sh
+2
-2

Apply the wait_for_log SIGPIPE fix with full-drain grep

lib.sh
+22
-12

Read up.sh in full before editing

Script
I need to modify up.sh to call reset_runtime_env after ensure_env, and then fix step 6 by removing the pre-created api container before bringing it up with --no-deps, since podman-compose ignores that flag when frontend depends on api and ends up starting the api container with frozen environment variables. The api service is the only one that needs removal since it's a dependency of frontend, which gets processed at step 3. cloudflared-api isn't a dependency of anything requested, so it won't be created then. Let me verify the compose.yaml structure and check for the env passthrough additions.


Read compose.yaml for dependency graph and env passthrough edits

Script
So the frontend service pulls in the api as a dependency when created at step 3, while cloudflared-api stays dormant since it's not explicitly requested. I need to update up.sh to reset the runtime environment after ensuring it's set up, and add a pre-removal step for any stale api containers before step 6. Then in compose.yaml, I'll add the three OTEL passthrough variables following the existing OTEL_E configuration.


Call reset_runtime_env at every up.sh start

up.sh
+7
-2

Step 6: remove any dependency-created api container so it is recreated with the tunnel-derived env

up.sh
+20
-3

Add the three OTEL metrics/compression passthrough vars to the api service env

compose.yaml
+8
-5
I need to examine the existing CI configuration files—the ci.yaml workflow, down.sh script, and Containerfiles—along with the .env.example and configuration documentation to understand the current setup before writing the container test script.


Read ci.yaml, down.sh, and check .env.example existence

Script

Read rest of .env.example and the telemetry table in docs/CONFIGURATION.md

Script

Document the two new OTEL metrics vars in .env.example

.env.example
+24
-11

Add the temporality + histogram rows to the telemetry table in CONFIGURATION.md

CONFIGURATION.md
+9
-7
I'm working through the CI container test script now—setting up the runtime detection and environment initialization, then configuring Docker Compose to build and start the database with health checks before bringing up the API and frontend services without their dependencies since the database is already running. So in CI, I don't need to recreate the api container since the environment from the bootstrap defaults is fine for the whole run. I'm skipping the cloudflared containers entirely since they'd require outbound network access to Cloudflare and would be flaky anyway. Instead, I'll wait for the api to log that the HTTP server is listening, then run smoke tests using a curl container on the compose-created network.

The network name is ${PROJECT}_default for both podman-compose and docker compose, and the api service gets a network alias matching its service name, so DNS resolution works. I'm creating a helper function that wraps the container runtime to run curl commands against that network, and I'll use sed to extract tokens from responses. Then I'll set up assertions to check the health endpoints and verify migrations ran correctly. I need to verify the password requirements from the auth handlers and confirm the exact field names for the signup response and captcha sync request structure.

For the captcha sync test, I'm planning two requests: the first with best_streak 4, current_streak 2, manual 3, auto 5 to reach total_solves 8, then a second sync that adds manual 4 and auto 8 to push the total to 20 while current_streak drops to 1 and best_streak stays at 4. I need to check the exact JSON field names in the captcha handler to construct the request bodies correctly.

Then I'll verify the leaderboard includes the username field, and test the preferences endpoint with a PUT request to set the theme to dark and confirm the response.

For the notes API, I'm checking POST to create a note with a body, GET to retrieve the list and verify it's there, and DELETE to remove it by ID—need to confirm what status code the delete handler returns. I'm also testing the OPTIONS preflight request expecting a 204 response, and finally checking that the frontend serves HTML with the content-security-policy header present.

I need to look at the captcha, notes, preferences, and auth handler files to understand the exact JSON structure for requests and responses.

Show less
Done
Now the CI test script. Let me verify exact request/response JSON shapes from the handlers first:


Catalogued API response formats and identified outstanding verification tasks
Catalogued API response formats and identified outstanding verification tasks

Verify JSON field names and status codes for captcha/notes/auth handlers

Script
I'm reviewing the API response shapes for the sync endpoint, which returns stats like manual and auto deltas along with streak information, and confirming the note operations—POST creates with a 201 status, DELETE removes with a 200 response, and GET retrieves the full list. The signup validation checks out with the password requirement and token in the response, but I need to verify what status code issueSession returns and whether the sync stats are wrapped in an object or returned directly.


Check session issue status code and captcha/prefs response shapes

Script
Done
please continue


Scrutinizing script behavior under set -e conditions
Scrutinizing script behavior under set -e conditions
I need to track down the issueSession return value since the grep only showed where it's called, not defined. I should also verify the CaptchaStats JSON tags in store.go and confirm the migrations count is at least 7, then move on to writing the CI container test script and updating ci.yaml before doing a final validation pass.


Find issueSession status code, CaptchaStats JSON tags, and migration count

Script
I'm confirming the API response formats across all endpoints: signup returns a token at the top level, the sync endpoint provides stats with compact JSON formatting from Go's encoder, the leaderboard includes both the list and the current user's data with a username field, and the preferences endpoint responds to PUT requests with a saved status confirmation.

For the remaining endpoints, GET requests return the preferences with the key-value pair and an exists flag, notes endpoints handle POST with a 201 response containing the note object, GET to retrieve all notes as an array, and DELETE to confirm deletion. The database has 7 migration files so the max version should be 7, and password validation requires at least 8 characters.

I'm verifying that Go's JSON encoder produces compact output without spaces after colons, so grepping for the exact format like "total_solves":8 will work correctly. The Authorization header follows the Bearer token format, and I need to extract note IDs from the POST response using a sed pattern to capture the UUID from the note object.

Refining the sed pattern to avoid matching the user_id field — I need to anchor it to the note object structure itself, like matching {"id":" at the start of the note object to ensure I'm grabbing the right ID field rather than accidentally capturing from user_id.

Show less

Verify Note struct field order, Bearer auth parsing, and WriteJSON encoding

Script
I need to remember that note IDs are integers, not UUIDs, so the JSON structure has a numeric id field. I can extract the ID using a sed pattern that captures the numeric value between the id field markers, and then use that numeric ID in the DELETE request to the notes endpoint. Now I'm looking for where the Bearer token parsing happens in the auth middleware—it's not showing up in session.go, so I need to search more broadly through the codebase.


Locate Bearer token parsing in auth middleware

Script
I'm setting up the CI container test script with the standard shell options and sourcing the necessary utilities, then configuring the project name from an argument with a default fallback, and defining the curl image and network naming conventions.

The cleanup trap will dump logs from all services on failure and always tear down the compose stack, then I'll build the images, start the database with a health check using pg_isready, bring up the API and frontend, and wait for the API to log that it's listening — failing the test if that doesn't happen within 90 seconds since this is CI and we need strict validation.

I'm creating helper functions to run curl commands inside a throwaway container on the compose network, plus assertion helpers to check response bodies and HTTP status codes, with careful attention to quoting under the IFS setting.

One thing to verify: the curl image needs to be pulled on first use, which works fine with podman pulling docker.io images. For DNS resolution, podman-compose sets up the network with DNS enabled and adds network aliases matching service names, so containers can reach each other by service name or container name — though docker compose uses dashes in container names while podman uses dashes too, so I should double-check which naming convention to rely on.

GitHub Actions runners have podman 4.9.x with netavark and aardvark-dns, so DNS on user-defined networks is supported. After bringing up the api and frontend services with up -d --no-deps, the api will wait for the database to be healthy, then run smoke tests. For the migrations check, I'm querying the database container to verify the migration status. Writing signup and token extraction...

Now I'm testing the stats endpoint with two different captcha solve scenarios to verify the totals accumulate correctly—first call returns 8 total solves with a best streak of 4, then the second call should show 20 total solves while the best streak stays at 4. This is the key regression test that was failing with 500 errors before the fix. Next I'm checking the leaderboard endpoint. Writing shell commands... Instead of using grep's quiet flag, I should redirect to /dev/null to ensure grep reads the entire input and avoids SIGPIPE issues. For the health checks, I'm adding curl calls to /healthz and /readyz endpoints with the -f flag to fail on HTTP errors, wrapped in a step logger. The -sS flag is already in the curl wrapper, so I just need to add -f where needed. The initial curl call also pulls the image, and the 20-second timeout only applies after the container starts. I'm also noting that podman-compose ignores the --no-deps flag and will still process the db service even when trying to start just api and frontend. With --no-deps in local dev, the db stays untouched and api+frontend start up; the depends_on health check gets skipped but that's fine since db is already healthy from the earlier wait. The cleanup trap is set up to catch failures and dump the last 100 lines of logs from each service (db, api, frontend) when the test exits non-zero. I need to be careful with the EXIT trap behavior — capturing the original exit status upfront and using exit "$status" at the end ensures the final code is explicit and unambiguous, especially under set -e where failing commands could interfere. I'll guard all trap commands with || true to prevent early exits. I also need to check that CR is set before calling cid_of since the trap might fire before ensure_env completes, and I should verify that the environment setup properly configures POSTGRES_PASSWORD and other secrets for the compose stack.

The GitHub Actions workflow installs podman-compose via pipx, which is already available on the runner, then validates the runtime versions before running the container smoke test script. Since podman and podman-compose come preinstalled on ubuntu-latest, the setup is straightforward.

I'm checking the migration table name next before writing the file.

Show less

Confirm migrations table name and health route paths

Script
Now I'm writing the CI container test script.

Done
All facts confirmed. Writing the CI container test script:


Validated script error handling and curl behavior mechanics
Validated script error handling and curl behavior mechanics

New CI script: builds both Containerfiles, brings up db+api+frontend via compose, and runs an end-to-end HTTP smoke test including the captcha-sync regression
ci-container-test.sh
I'm realizing there's a subtle issue with how the health check command behaves under set -e — when the curl fails, the && operator causes the entire compound statement to fail, which triggers the error trap and dumps logs as intended. That's actually the desired behavior for a hard fail, but I need to think through what happens when a command...

Actually, looking at the rest of the logic: the wait_healthy function passes arguments through correctly, the sed pattern for token extraction handles the single occurrence fine, and the note ID extraction regex is solid. The FAILED counter works because the assertion functions run in the current shell and their guards prevent nonzero returns, keeping set -e safe. The main concern is when ccurl fails during variable assignment — that triggers set -e and the trap, which is acceptable since transport failures are fatal anyway, and curl without -f still captures HTTP error bodies.

Now I need to update the CI configuration to add the containers job.

Show less
Done
Now the updated GitHub Actions workflow:


Navigated Go dependency resolution and verification constraints
Navigated Go dependency resolution and verification constraints

Add the containers job that builds the images and runs the compose smoke test with podman

ci.yaml
+38
-14
Running a full validation pass across gofmt, go test, bash syntax, shellcheck, YAML parsing, and secrets scanning.


gofmt validation on all changed Go files

Script

Run the config package tests for real in a stdlib-only scratch module

Script
The network is blocking access to proxy.golang.org, so I need to check what Go version is currently installed locally to work around this constraint.


Pin scratch module to the installed Go version and run the config tests

Script
The grep pattern with -P flag isn't extracting the version correctly from the go version output — it's matching "1.22" from "go1.22.2" but the -P option might not be supported on this system, so the extraction is failing silently.

Actually, I think I see the real issue: when I cd into the scratch/cfg directory and run go version there, it's triggering the toolchain resolution from the go.mod file which specifies go 1.26, causing it to try downloading that version and fail. That's why grep gets empty output and the sed substitution produces just "go " with nothing after it. The simpler fix is to just write the go.mod file directly with the go 1.22 version instead of trying to extract it dynamically.


Fix go.mod and run config tests with local toolchain

Script
The config tests all pass, so now I need to verify that telemetry.go compiles with the OpenTelemetry dependencies. The network restrictions block the standard Go proxy, but I can work around this by setting GOPROXY=direct, which will fetch modules directly from GitHub via git over HTTPS — a protocol that's allowed. Since the OpenTelemetry modules are hosted on github.com, this approach should let me download the dependencies I need.

However, there's a toolchain version mismatch: the go.mod specifies go 1.26, but I only have go 1.22.2 available locally. The OpenTelemetry v1.44.0 likely requires go 1.24 or higher, so even with the dependencies downloaded, the local compiler won't be able to build the code. I could rewrite the go directive temporarily, but that risks pulling in dependencies that require a newer Go version than what I have.

Since full compilation isn't feasible, I'll verify the code through syntax checking with gofmt and by manually validating that the OpenTelemetry API calls match the actual function signatures from the v1.44.0 source — I've already confirmed that WithTemporalitySelector and WithAggregationSelector exist with the correct types.

Show less
Done
Config tests all pass. The telemetry package needs otel deps I can't fetch here, so I verify the exact API usage against the pinned v1.44.0 source directly:


Verifying no new errors introduced by changes
Verifying no new errors introduced by changes

Final verification of every otel symbol used in telemetry.go against pinned v1.44.0

Script
Now I'm running validation checks across the shell scripts and configuration files—checking syntax with bash and shellcheck, parsing the YAML files, and scanning for any exposed secrets.

Done
Every otel symbol confirmed. Shell + YAML validation:


Verified test results and prepared multiple files for delivery
Verified test results and prepared multiple files for delivery

bash -n and shellcheck on all changed/new scripts

Script
The shellcheck findings are all pre-existing issues from before my changes—info-level warnings about shell directives and an unused loop counter pattern that's consistent throughout the codebase. No new errors were introduced, so I should verify the exit code with error-level severity filtering to confirm there are truly zero errors to address.


Confirm zero shellcheck errors and validate YAML files

Script

Secrets scan on deliverables, then empirically test the ensure_env DSN-persist + reset_runtime_env behavior

Script
The "FOUND SECRETS" alert is a false positive—it just matched the placeholder format in the .env.example file, not an actual secret. The sandbox test confirms the fix is working: the shell-exported UPTRACE_DSN persisted correctly into .env without being overwritten.

Now I need to verify the wait_for_log fix handles the SIGPIPE scenario properly under pipefail, and then test the sed extraction patterns against realistic JSON data. I should also validate that the assert helper functions in ci-container-test.sh work correctly by testing the critical extraction and assertion logic directly.

The .env.example file only contains placeholder values with no real secrets, and I've confirmed both environment fixes work in the sandbox. Now I'll verify the wait_for_log fix and JSON extraction patterns.


Empirically verify the wait_for_log fix, sed extractions, and assert helper semantics under strict mode

Script

Rerun harness under bash explicitly (default shell was sh)

Script
I've verified that the new drain pattern correctly catches the present pattern issue that the old -Eq approach missed under pipefail, and confirmed the token and note-id extractions work properly against the Go-encoder JSON output. The assert_contains function handles set -e correctly with FAILED tracking intact. Now I need to validate the SQL fix by running it against the local Postgres instance to ensure the final query in store.go has all the necessary casts and ON CONFLICT handling.

Running the exact SQL from store.go against real Postgres to verify the fix works end-to-end through pgx's untyped-parameter path.


Final SQL regression check: the fixed statement plans and executes correctly on both insert and update paths

Script

Run the SQL regression as the postgres user

Script
The store.go implementation is working correctly through the untyped-Parse path, with insert operations returning the expected 8/3/5 metrics and best staying at 4, while update operations show 20/7/13 with best at 4 and current dropping to 1—exactly matching what the CI test expects. Now I need to do a final visual review of all the modified files to ensure nothing's broken, then provide the complete files with their exact paths.

Done
The exact statement from store.go now plans and executes correctly through the untyped-Parse path — insert 8/3/5, update 20/7/13 with GREATEST keeping best=4. Let me pull the final file contents for delivery:


Validated all deliverables and prepared comprehensive file submission
Validated all deliverables and prepared comprehensive file submission

Line counts of all 11 deliverables

Script

Read final store.go for delivery

Script

Read final config.go for delivery

Script

Read final telemetry.go and lib.sh for delivery

Script
I've gathered all the files and now I'm ready to compile the complete response with all 11 full files, including the final version of up.sh and the remaining smaller files.


Read final up.sh, ci.yaml, compose.yaml, .env.example for delivery

Script
I need to finalize the remaining test files and documentation — I have the config_test.go changes ready to insert after the TestResolveTelemetry_UptraceDSN test, and I should verify the full content of docs/CONFIGURATION.md and ci-container-test.sh before pasting them out.


Read final CONFIGURATION.md for delivery

Script
I've validated all the core files—store.go passes gofmt and SQL verification against PostgreSQL 16, config.go and its tests compile cleanly with all 13 tests passing, telemetry.go has every OpenTelemetry symbol checked against the pinned version, and lib.sh passes bash syntax and shellcheck validation with the helper functions tested in a sandbox environment.

Now I'm finishing the validation sweep: up.sh and ci-container-test.sh both pass syntax and linting checks with extraction patterns and assertions verified, the YAML files parse correctly, documentation is updated, and I've confirmed no credentials are exposed in any deliverable.

I'm ready to compile the full response with all 11 files, their paths, the root-cause findings with empirical verification, and deployment notes including the passkey issue on the current instance and the credential rotation steps needed for the PAT and Uptrace DSN.

Show less
Done
All 11 files are finalized and validated. Here's what was actually wrong, each root cause verified empirically, followed by every full file.

Root causes
1. Captcha sync always 500. SyncCaptchaStats had VALUES (..., $4 + $5, ...). pgx v5's default statement-cache mode sends the extended-protocol Parse with no declared parameter types, so Postgres must infer them — and $4 + $5 fails at plan time with operator is not unique: unknown + unknown. The statement could never even execute, hence a 500 on every single call. I reproduced the exact error on Postgres 16 via PREPARE without a type list, and verified the fix (::bigint casts on every integer placeholder, in both the VALUES and the ON CONFLICT clauses): insert path returns 8/3/5, update path returns 20/7/13 with GREATEST keeping best=4 and current dropping to 1.

2. CORS failure on signup (and silently broken passkeys). Your virginia.txt log shows podman-compose ignores --no-deps: step 3 (up -d --no-deps frontend cloudflared-frontend) also created and started the api container (frontend depends_on api) with env frozen at that moment. Step 6's up then hit "name already in use" and never recreated it — so the api never received the tunnel-derived RP_ID/RP_ORIGINS/CORS_ALLOWED_ORIGINS. Deleting .env "fixed" CORS only because a fresh .env defaults to *. Side effect: passkeys on your current deployment are broken (RP ID is stuck at localhost while the origin is trycloudflare.com) until you redeploy. Fix is two-part: reset_runtime_env returns the three tunnel-derived keys to bootstrap defaults at every run start (so deleting .env is never needed and your secrets/DSN survive), and step 6 now removes any dependency-created api container before starting it so compose recreates it with the correct env (a no-op under docker compose).

3. "did not observe API listening log yet" on every run. wait_for_log used grep -Eq: grep exits at first match, podman logs takes SIGPIPE (exit 141), and under pipefail the pipeline reports failure even though the pattern matched. Every success was inverted into a miss. Verified both directions in a harness: the old form misses a present pattern; the new full-drain form (grep -E ... >/dev/null) matches correctly. Same bug family you already fixed in cid_of and poll_tunnel_url.

4. Uptrace telemetry gaps. Good news first: your endpoint derivation was already correct — https://api.uptrace.dev is a valid OTLP/HTTP base and otel-go v1.44.0's WithEndpointURL appends /v1/{traces,metrics,logs} itself; gzip was already on. What was missing per Uptrace's guidance: delta temporality and base2 exponential histograms. Both are now wired via WithTemporalitySelector/WithAggregationSelector (every symbol verified against the pinned v1.44.0 source). Plus one nasty footgun I found and fixed: an existing .env with an empty UPTRACE_DSN= line silently clobbered a shell-exported DSN (load_env re-exports the empty value), turning telemetry off with zero errors — ensure_env now persists caller-exported telemetry vars into .env before loading it. Verified in a sandbox: exported DSN survives, stale tunnel values reset, generated secrets preserved.

5. CI. New scripts/ci-container-test.sh builds both Containerfiles, brings up db+api+frontend via podman-compose (no tunnels), and runs an HTTP smoke test from a curl container on the compose network — including the captcha-sync regression (which only reproduces against real Postgres), prefs, notes, leaderboard, CORS preflight, and the frontend CSP header. Wired into ci.yaml as a second job.

Validation: gofmt -e clean on all 4 Go files; all 13 config tests pass (ran for real in a stdlib-only scratch module); bash -n + shellcheck error-free on all scripts; YAML parses; no PAT or DSN in any file.

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
//
// Every integer placeholder carries an explicit ::bigint cast. pgx v5 uses the
// extended query protocol and, in its default statement-cache mode, sends the
// Parse step without declared parameter types. Postgres then has to infer them
// and the expression `$4 + $5` fails at plan time with
// `ERROR: operator is not unique: unknown + unknown` — the statement can never
// even execute, so the sync endpoint 500s on every request. Casting each
// parameter pins the types and makes the statement plannable. (Placeholders
// compared against a typed column, like `$1::uuid = user_id`, would be
// inferable anyway, but we cast uniformly for clarity.)
func (s *Store) SyncCaptchaStats(ctx context.Context, userID string, in CaptchaSyncInput) (CaptchaStats, error) {
	st := CaptchaStats{UserID: userID}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO captcha_stats
			(user_id, best_streak, current_streak, total_solves, manual_solves, auto_solves, updated_at)
		VALUES ($1::uuid, $2::bigint, $3::bigint, $4::bigint + $5::bigint, $4::bigint, $5::bigint, now())
		ON CONFLICT (user_id) DO UPDATE SET
			best_streak    = GREATEST(captcha_stats.best_streak, EXCLUDED.best_streak),
			current_streak = EXCLUDED.current_streak,
			total_solves   = captcha_stats.total_solves + $4::bigint + $5::bigint,
			manual_solves  = captcha_stats.manual_solves + $4::bigint,
			auto_solves    = captcha_stats.auto_solves + $5::bigint,
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
internal/config/config.go
go
// Package config is the single source of truth for every runtime setting.
//
// Everything the application can be tuned with is read here, from environment
// variables, and nowhere else. No other package reads os.Getenv. This keeps
// the "all configuration is central in one location" promise from the design
// notes: to see everything GoTunnels can be configured with, read this file.
package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all resolved runtime settings for the API service.
type Config struct {
	// Identity -----------------------------------------------------------
	InstanceID  string // unique per running instance; also an OTel resource attribute
	ServiceName string // OTel service.name
	Version     string // build version (set via -ldflags at build time)

	// HTTP ---------------------------------------------------------------
	HTTPAddr        string        // listen address, e.g. ":8080"
	ShutdownTimeout time.Duration // graceful shutdown budget

	// Database -----------------------------------------------------------
	DatabaseURL      string
	DBMaxConns       int32
	DBMinConns       int32
	DBConnectTimeout time.Duration

	// Sessions -----------------------------------------------------------
	SessionCookieName string
	SessionTTL        time.Duration

	// Secrets (generated per-instance by scripts/up.sh, never committed) -
	ipHashPepper []byte   // used as sha256(pepper || ip)
	totpAESKey   [32]byte // AES-256-GCM key for encrypting TOTP secrets at rest

	// CORS ---------------------------------------------------------------
	// Exact allowed origins. Because the browser talks to the API cross-origin
	// (frontend tunnel URL != API tunnel URL) and sends credentials, we cannot
	// use the "*" wildcard together with credentials; we echo an allowed origin.
	CORSAllowedOrigins []string

	// WebAuthn / passkeys -----------------------------------------------
	RPID          string   // relying-party ID: the frontend's registrable domain
	RPDisplayName string   // human-readable name shown by the authenticator
	RPOrigins     []string // full origins the browser will present, e.g. https://x.trycloudflare.com

	// Content Security Policy -------------------------------------------
	// The header itself is emitted by Caddy on the frontend. These values let
	// the API report what mode it believes is active and are surfaced on the
	// health/info endpoint so the whole system agrees on one central policy.
	CSPMode   string // "report-only" or "enforce"
	CSPPolicy string // the policy string (informational mirror of Caddy's)

	// Telemetry ----------------------------------------------------------
	Telemetry TelemetryConfig

	// Dev conveniences ---------------------------------------------------
	Dev bool // when true, missing secrets are generated ephemerally with a warning
}

// TelemetryConfig captures OTLP/HTTP exporter settings. When Enabled is false
// the service logs to stdout only and installs no-op trace/metric providers.
type TelemetryConfig struct {
	Enabled     bool
	EndpointURL string            // full base URL, e.g. https://api.uptrace.dev
	Headers     map[string]string // e.g. {"uptrace-dsn": "..."}
	Insecure    bool              // allow http:// (in-cluster collectors)
	Compression string            // "gzip" or ""

	// MetricsTemporality mirrors OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE
	// ("delta", "cumulative", or "lowmemory"; normalized to lowercase). Uptrace
	// prefers delta temporality, so that is the default.
	MetricsTemporality string
	// MetricsHistogram mirrors OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION
	// ("base2_exponential_bucket_histogram" or "explicit_bucket_histogram";
	// normalized to lowercase). Exponential histograms compress better and give
	// Uptrace more accurate percentiles, so they are the default.
	MetricsHistogram string
}

// Load reads and validates configuration from the environment.
func Load() (*Config, error) {
	c := &Config{
		InstanceID:         getenv("GOTUNNELS_INSTANCE_ID", defaultInstanceID()),
		ServiceName:        getenv("OTEL_SERVICE_NAME", getenv("GOTUNNELS_SERVICE_NAME", "gotunnels-api")),
		Version:            getenv("GOTUNNELS_VERSION", "dev"),
		HTTPAddr:           getenv("GOTUNNELS_HTTP_ADDR", ":8080"),
		ShutdownTimeout:    getdur("GOTUNNELS_SHUTDOWN_TIMEOUT", 15*time.Second),
		DatabaseURL:        getenv("DATABASE_URL", ""),
		DBMaxConns:         int32(getint("GOTUNNELS_DB_MAX_CONNS", 20)),
		DBMinConns:         int32(getint("GOTUNNELS_DB_MIN_CONNS", 2)),
		DBConnectTimeout:   getdur("GOTUNNELS_DB_CONNECT_TIMEOUT", 30*time.Second),
		SessionCookieName:  getenv("GOTUNNELS_SESSION_COOKIE_NAME", "gotunnels_session"),
		SessionTTL:         getdur("GOTUNNELS_SESSION_TTL", 24*time.Hour),
		CORSAllowedOrigins: splitList(getenv("GOTUNNELS_CORS_ALLOWED_ORIGINS", "*")),
		RPID:               getenv("GOTUNNELS_RP_ID", "localhost"),
		RPDisplayName:      getenv("GOTUNNELS_RP_DISPLAY_NAME", "GoTunnels"),
		RPOrigins:          splitList(getenv("GOTUNNELS_RP_ORIGINS", "http://localhost:8080")),
		CSPMode:            getenv("GOTUNNELS_CSP_MODE", "report-only"),
		CSPPolicy:          getenv("GOTUNNELS_CSP_POLICY", DefaultCSPPolicy),
		Dev:                getbool("GOTUNNELS_DEV", false),
	}

	if err := c.resolveSecrets(); err != nil {
		return nil, err
	}
	c.resolveTelemetry()

	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// DefaultCSPPolicy locks everything to same-origin: no third-party scripts,
// styles, images, fonts, media, or frames. Everything is self-hosted. It is
// emitted in report-only mode by default (see the Caddyfile), so violations
// are reported but nothing is blocked yet.
const DefaultCSPPolicy = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self'; " +
	"img-src 'self'; " +
	"font-src 'self'; " +
	"connect-src 'self' https:; " +
	"media-src 'self'; " +
	"object-src 'none'; " +
	"frame-src 'none'; " +
	"frame-ancestors 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'"

func (c *Config) resolveSecrets() error {
	// IP hashing pepper.
	if pepper := os.Getenv("GOTUNNELS_IP_HASH_PEPPER"); pepper != "" {
		c.ipHashPepper = []byte(pepper)
	} else if c.Dev {
		c.ipHashPepper = mustRandom(32)
	} else {
		return fmt.Errorf("config: GOTUNNELS_IP_HASH_PEPPER is required (set GOTUNNELS_DEV=1 to auto-generate for local dev)")
	}

	// TOTP encryption key. Accept any string; derive a fixed 32-byte AES key
	// via SHA-256 so operators can supply `openssl rand -base64 32` output
	// without worrying about exact byte length.
	if raw := os.Getenv("GOTUNNELS_TOTP_ENCRYPTION_KEY"); raw != "" {
		c.totpAESKey = sha256.Sum256([]byte(raw))
	} else if c.Dev {
		c.totpAESKey = sha256.Sum256(mustRandom(32))
	} else {
		return fmt.Errorf("config: GOTUNNELS_TOTP_ENCRYPTION_KEY is required (set GOTUNNELS_DEV=1 to auto-generate for local dev)")
	}
	return nil
}

func (c *Config) resolveTelemetry() {
	// Preference order:
	//   1. UPTRACE_DSN (convenience) -> derive endpoint + uptrace-dsn header
	//   2. OTEL_EXPORTER_OTLP_ENDPOINT (+ OTEL_EXPORTER_OTLP_HEADERS)
	//   3. disabled (stdout logging, no-op traces/metrics)
	//
	// The two metrics knobs follow the OpenTelemetry spec's environment
	// variables. Their spec values are UPPERCASE (e.g. DELTA,
	// BASE2_EXPONENTIAL_BUCKET_HISTOGRAM) but we normalize to lowercase so
	// comparisons elsewhere are simple and either casing works.
	tc := TelemetryConfig{
		Headers:     map[string]string{},
		Compression: strings.ToLower(getenv("OTEL_EXPORTER_OTLP_COMPRESSION", "gzip")),
		MetricsTemporality: strings.ToLower(getenv(
			"OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "delta")),
		MetricsHistogram: strings.ToLower(getenv(
			"OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION",
			"base2_exponential_bucket_histogram")),
	}

	if dsn := os.Getenv("UPTRACE_DSN"); dsn != "" {
		endpoint, insecure := endpointFromDSN(dsn)
		tc.Enabled = true
		tc.EndpointURL = endpoint
		tc.Insecure = insecure
		tc.Headers["uptrace-dsn"] = dsn
	} else if ep := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); ep != "" {
		tc.Enabled = true
		tc.EndpointURL = ep
		tc.Insecure = strings.HasPrefix(ep, "http://")
		for k, v := range parseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")) {
			tc.Headers[k] = v
		}
	}
	c.Telemetry = tc
}

// Validate checks invariants that must hold before the server starts.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: DATABASE_URL is required")
	}
	if c.HTTPAddr == "" {
		return fmt.Errorf("config: GOTUNNELS_HTTP_ADDR is required")
	}
	if c.SessionTTL <= 0 {
		return fmt.Errorf("config: GOTUNNELS_SESSION_TTL must be positive")
	}
	switch c.CSPMode {
	case "report-only", "enforce":
	default:
		return fmt.Errorf("config: GOTUNNELS_CSP_MODE must be 'report-only' or 'enforce', got %q", c.CSPMode)
	}
	if len(c.RPOrigins) == 0 {
		return fmt.Errorf("config: GOTUNNELS_RP_ORIGINS must contain at least one origin")
	}
	return nil
}

// IPHashPepper returns a copy-safe reference to the pepper bytes.
func (c *Config) IPHashPepper() []byte { return c.ipHashPepper }

// TOTPAESKey returns the 32-byte AES key for TOTP secret encryption.
func (c *Config) TOTPAESKey() [32]byte { return c.totpAESKey }

// CORSAllowsAny reports whether the wildcard "*" origin is configured.
func (c *Config) CORSAllowsAny() bool {
	for _, o := range c.CORSAllowedOrigins {
		if o == "*" {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// small env helpers
// ---------------------------------------------------------------------------

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getint(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return def
}

func getbool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
			return b
		}
	}
	return def
}

func getdur(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(strings.TrimSpace(v)); err == nil {
			return d
		}
	}
	return def
}

// splitList parses a comma-or-space separated list, trimming empties.
func splitList(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' })
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

// parseHeaders parses OTEL_EXPORTER_OTLP_HEADERS ("k1=v1,k2=v2").
func parseHeaders(s string) map[string]string {
	out := map[string]string{}
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out
}

// endpointFromDSN extracts the OTLP base endpoint from an Uptrace DSN of the
// form "https://TOKEN@host[:port][/path]". Returns the base URL and whether it
// is insecure (http).
func endpointFromDSN(dsn string) (endpoint string, insecure bool) {
	// Very small, dependency-free parse: scheme://[userinfo@]host
	scheme := "https"
	rest := dsn
	if i := strings.Index(rest, "://"); i >= 0 {
		scheme = rest[:i]
		rest = rest[i+3:]
	}
	if at := strings.LastIndex(rest, "@"); at >= 0 {
		rest = rest[at+1:]
	}
	// strip any path/query
	if q := strings.IndexAny(rest, "/?"); q >= 0 {
		rest = rest[:q]
	}
	return scheme + "://" + rest, scheme == "http"
}

// defaultInstanceID returns a stable-ish identifier used when the operator did
// not supply one (mostly for bare `go run` during development).
func defaultInstanceID() string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return "local-" + hex.EncodeToString(mustRandom(4))
}

func mustRandom(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("config: crypto/rand failed: " + err.Error())
	}
	return b
}

// DecodeKeyMaterial is a helper exposed for tests and tooling: it accepts hex,
// base64 (std or url), or raw bytes and returns them. Unused by Load directly
// but kept for completeness where callers want exact-length keys.
func DecodeKeyMaterial(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if b, err := hex.DecodeString(s); err == nil && len(s)%2 == 0 {
		return b, nil
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return []byte(s), nil
}
internal/config/config_test.go
go
package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	t.Setenv("GOTUNNELS_DEV", "1") // auto-generate secrets
	// DATABASE_URL deliberately unset.
	if _, err := Load(); err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoad_RequiresSecretsWhenNotDev(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x/y")
	t.Setenv("GOTUNNELS_DEV", "0")
	// no pepper / totp key
	if _, err := Load(); err == nil {
		t.Fatal("expected error when secrets are missing and not in dev mode")
	}
}

func TestLoad_DevGeneratesSecrets(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_DEV", "1")
	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.IPHashPepper()) == 0 {
		t.Error("expected a generated IP hash pepper in dev mode")
	}
	var zero [32]byte
	if c.TOTPAESKey() == zero {
		t.Error("expected a non-zero TOTP AES key in dev mode")
	}
	if c.CSPMode != "report-only" {
		t.Errorf("expected default CSP mode report-only, got %q", c.CSPMode)
	}
	if !strings.Contains(c.CSPPolicy, "default-src 'self'") {
		t.Errorf("default CSP policy should lock to self, got %q", c.CSPPolicy)
	}
}

func TestLoad_TOTPKeyDerivedDeterministically(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_IP_HASH_PEPPER", "pepper")
	t.Setenv("GOTUNNELS_TOTP_ENCRYPTION_KEY", "the-same-secret")

	c1, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	c2, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c1.TOTPAESKey() != c2.TOTPAESKey() {
		t.Error("TOTP key derivation must be deterministic for a given secret")
	}
}

func TestValidate_CSPMode(t *testing.T) {
	c := &Config{
		DatabaseURL: "x",
		HTTPAddr:    ":8080",
		SessionTTL:  time.Hour,
		RPOrigins:   []string{"http://localhost"},
		CSPMode:     "nonsense",
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected invalid CSP mode to fail validation")
	}
	c.CSPMode = "enforce"
	if err := c.Validate(); err != nil {
		t.Fatalf("enforce should be valid: %v", err)
	}
}

func TestSplitList(t *testing.T) {
	cases := map[string][]string{
		"a,b,c":            {"a", "b", "c"},
		"a b c":            {"a", "b", "c"},
		" a , b ,, c ":     {"a", "b", "c"},
		"https://x https:": {"https://x", "https:"},
		"":                 {},
	}
	for in, want := range cases {
		got := splitList(in)
		if len(got) != len(want) {
			t.Errorf("splitList(%q) len = %d, want %d (%v)", in, len(got), len(want), got)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("splitList(%q)[%d] = %q, want %q", in, i, got[i], want[i])
			}
		}
	}
}

func TestParseHeaders(t *testing.T) {
	h := parseHeaders("uptrace-dsn=https://tok@api.uptrace.dev,x-extra = val ")
	if h["uptrace-dsn"] != "https://tok@api.uptrace.dev" {
		t.Errorf("unexpected dsn header: %q", h["uptrace-dsn"])
	}
	if h["x-extra"] != "val" {
		t.Errorf("unexpected x-extra header: %q", h["x-extra"])
	}
}

func TestEndpointFromDSN(t *testing.T) {
	cases := []struct {
		dsn          string
		wantEndpoint string
		wantInsecure bool
	}{
		{"https://TOKEN@api.uptrace.dev?grpc=4317", "https://api.uptrace.dev", false},
		{"https://TOKEN@api.uptrace.dev:443/v1", "https://api.uptrace.dev:443", false},
		{"http://token@localhost:14318", "http://localhost:14318", true},
	}
	for _, tc := range cases {
		ep, insecure := endpointFromDSN(tc.dsn)
		if ep != tc.wantEndpoint {
			t.Errorf("endpointFromDSN(%q) endpoint = %q, want %q", tc.dsn, ep, tc.wantEndpoint)
		}
		if insecure != tc.wantInsecure {
			t.Errorf("endpointFromDSN(%q) insecure = %v, want %v", tc.dsn, insecure, tc.wantInsecure)
		}
	}
}

func TestResolveTelemetry_DisabledByDefault(t *testing.T) {
	// Ensure a clean env for the OTLP vars.
	t.Setenv("UPTRACE_DSN", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	c := &Config{}
	c.resolveTelemetry()
	if c.Telemetry.Enabled {
		t.Error("telemetry should be disabled when no endpoint/DSN is configured")
	}
}

func TestResolveTelemetry_UptraceDSN(t *testing.T) {
	t.Setenv("UPTRACE_DSN", "https://secret@api.uptrace.dev?grpc=4317")
	c := &Config{}
	c.resolveTelemetry()
	if !c.Telemetry.Enabled {
		t.Fatal("telemetry should be enabled with an Uptrace DSN")
	}
	if c.Telemetry.EndpointURL != "https://api.uptrace.dev" {
		t.Errorf("endpoint = %q", c.Telemetry.EndpointURL)
	}
	if c.Telemetry.Headers["uptrace-dsn"] == "" {
		t.Error("expected uptrace-dsn header to be set")
	}
}

func TestResolveTelemetry_MetricsDefaults(t *testing.T) {
	// Uptrace-friendly defaults must hold even when the OTEL_* metrics vars
	// are entirely absent: delta temporality + base2 exponential histograms
	// + gzip compression.
	t.Setenv("UPTRACE_DSN", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_COMPRESSION", "")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION", "")
	c := &Config{}
	c.resolveTelemetry()
	if got := c.Telemetry.Compression; got != "gzip" {
		t.Errorf("default compression = %q, want gzip", got)
	}
	if got := c.Telemetry.MetricsTemporality; got != "delta" {
		t.Errorf("default temporality = %q, want delta", got)
	}
	if got := c.Telemetry.MetricsHistogram; got != "base2_exponential_bucket_histogram" {
		t.Errorf("default histogram aggregation = %q, want base2_exponential_bucket_histogram", got)
	}
}

func TestResolveTelemetry_MetricsOverrides(t *testing.T) {
	// The OTel spec documents UPPERCASE values (DELTA, CUMULATIVE,
	// BASE2_EXPONENTIAL_BUCKET_HISTOGRAM...); we must accept them and
	// normalize to lowercase.
	t.Setenv("UPTRACE_DSN", "https://secret@api.uptrace.dev")
	t.Setenv("OTEL_EXPORTER_OTLP_COMPRESSION", "NONE")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "CUMULATIVE")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION", "EXPLICIT_BUCKET_HISTOGRAM")
	c := &Config{}
	c.resolveTelemetry()
	if got := c.Telemetry.Compression; got != "none" {
		t.Errorf("compression = %q, want none", got)
	}
	if got := c.Telemetry.MetricsTemporality; got != "cumulative" {
		t.Errorf("temporality = %q, want cumulative", got)
	}
	if got := c.Telemetry.MetricsHistogram; got != "explicit_bucket_histogram" {
		t.Errorf("histogram aggregation = %q, want explicit_bucket_histogram", got)
	}
}

func TestGetHelpers(t *testing.T) {
	t.Setenv("X_INT", "42")
	if getint("X_INT", 0) != 42 {
		t.Error("getint failed")
	}
	if getint("X_MISSING", 7) != 7 {
		t.Error("getint default failed")
	}
	t.Setenv("X_BOOL", "true")
	if !getbool("X_BOOL", false) {
		t.Error("getbool failed")
	}
	t.Setenv("X_DUR", "250ms")
	if getdur("X_DUR", 0) != 250*time.Millisecond {
		t.Error("getdur failed")
	}
}
internal/telemetry/telemetry.go
go
// Package telemetry wires the vendor-neutral OpenTelemetry Go SDK for all three
// signals (traces, metrics, logs) and exports them over OTLP/HTTP.
//
// We deliberately never import any Uptrace SDK. Uptrace (or any OTLP backend)
// is configured purely via a DSN / endpoint, so swapping backends is a config
// change, not a code change.
//
// Logs are always written to stdout as JSON (so `podman logs` is useful) and,
// when an OTLP endpoint is configured, additionally shipped to the collector
// via the OpenTelemetry log bridge. When no endpoint is configured, traces and
// metrics use no-op providers and only stdout logging remains.
package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/kusl/GoTunnels/internal/config"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellog "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// scopeName is the instrumentation scope for logs/traces/metrics we emit
// directly (as opposed to via otelhttp instrumentation).
const scopeName = "github.com/kusl/GoTunnels"

// Providers bundles what the application needs after setup.
type Providers struct {
	Logger   *slog.Logger
	shutdown []func(context.Context) error
}

// Shutdown flushes and stops all providers. Safe to call once during teardown.
func (p *Providers) Shutdown(ctx context.Context) error {
	var errs []error
	for i := len(p.shutdown) - 1; i >= 0; i-- {
		if err := p.shutdown[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Setup initializes telemetry from configuration. The returned Providers always
// has a usable Logger, even on partial failure.
func Setup(ctx context.Context, cfg *config.Config) (*Providers, error) {
	p := &Providers{}

	// stdout logging is always on.
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	if !cfg.Telemetry.Enabled {
		p.Logger = slog.New(stdoutHandler).With(
			slog.String("service.name", cfg.ServiceName),
			slog.String("service.instance.id", cfg.InstanceID),
		)
		p.Logger.Info("telemetry disabled: logging to stdout only, no-op traces/metrics")
		return p, nil
	}

	res, err := buildResource(ctx, cfg)
	if err != nil {
		// A resource failure is not fatal; fall back to stdout logging.
		l := slog.New(stdoutHandler)
		l.Error("telemetry: failed to build resource; continuing with stdout logging", slog.Any("err", err))
		p.Logger = l
		return p, nil
	}

	// ---- Traces -------------------------------------------------------
	traceExp, err := otlptracehttp.New(ctx, traceHTTPOpts(cfg)...)
	if err != nil {
		return fallback(stdoutHandler, cfg, err), nil
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))
	p.shutdown = append(p.shutdown, tp.Shutdown)

	// ---- Metrics ------------------------------------------------------
	metricExp, err := otlpmetrichttp.New(ctx, metricHTTPOpts(cfg)...)
	if err != nil {
		return fallback(stdoutHandler, cfg, err), nil
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp,
			sdkmetric.WithInterval(30*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	p.shutdown = append(p.shutdown, mp.Shutdown)

	// ---- Logs ---------------------------------------------------------
	logExp, err := otlploghttp.New(ctx, logHTTPOpts(cfg)...)
	if err != nil {
		return fallback(stdoutHandler, cfg, err), nil
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)
	otellog.SetLoggerProvider(lp)
	p.shutdown = append(p.shutdown, lp.Shutdown)

	// Fan out slog to BOTH stdout and the OTel log bridge.
	otelHandler := otelslog.NewHandler(scopeName, otelslog.WithLoggerProvider(lp))
	p.Logger = slog.New(NewMultiHandler(stdoutHandler, otelHandler)).With(
		slog.String("service.name", cfg.ServiceName),
		slog.String("service.instance.id", cfg.InstanceID),
	)
	p.Logger.Info("telemetry enabled",
		slog.String("otlp.endpoint", cfg.Telemetry.EndpointURL),
		slog.Bool("otlp.insecure", cfg.Telemetry.Insecure),
		slog.String("otlp.compression", cfg.Telemetry.Compression),
		slog.String("otlp.metrics.temporality", cfg.Telemetry.MetricsTemporality),
		slog.String("otlp.metrics.histogram", cfg.Telemetry.MetricsHistogram),
	)
	return p, nil
}

func fallback(h slog.Handler, cfg *config.Config, cause error) *Providers {
	l := slog.New(h).With(slog.String("service.name", cfg.ServiceName))
	l.Error("telemetry: exporter setup failed; continuing with stdout logging", slog.Any("err", cause))
	return &Providers{Logger: l}
}

func buildResource(ctx context.Context, cfg *config.Config) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.Version),
			semconv.ServiceInstanceID(cfg.InstanceID),
		),
	)
}

func traceHTTPOpts(cfg *config.Config) []otlptracehttp.Option {
	opts := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(cfg.Telemetry.EndpointURL)}
	if cfg.Telemetry.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(cfg.Telemetry.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Telemetry.Headers))
	}
	if cfg.Telemetry.Compression == "gzip" {
		opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.GzipCompression))
	}
	return opts
}

func metricHTTPOpts(cfg *config.Config) []otlpmetrichttp.Option {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpointURL(cfg.Telemetry.EndpointURL),
		otlpmetrichttp.WithTemporalitySelector(temporalitySelector(cfg.Telemetry.MetricsTemporality)),
		otlpmetrichttp.WithAggregationSelector(aggregationSelector(cfg.Telemetry.MetricsHistogram)),
	}
	if cfg.Telemetry.Insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	if len(cfg.Telemetry.Headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(cfg.Telemetry.Headers))
	}
	if cfg.Telemetry.Compression == "gzip" {
		opts = append(opts, otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression))
	}
	return opts
}

// temporalitySelector maps OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE
// (already lowercased by config) onto an SDK TemporalitySelector, following
// the OTLP exporter spec:
//
//	delta:      Counter, ObservableCounter, Histogram -> delta; others cumulative
//	lowmemory:  Counter, Histogram -> delta; others cumulative
//	cumulative: everything cumulative (the SDK default)
//
// Uptrace prefers delta, which is our config default. Unknown values fall back
// to the SDK default (cumulative) rather than failing startup.
func temporalitySelector(pref string) sdkmetric.TemporalitySelector {
	switch strings.ToLower(pref) {
	case "delta":
		return func(k sdkmetric.InstrumentKind) metricdata.Temporality {
			switch k {
			case sdkmetric.InstrumentKindCounter,
				sdkmetric.InstrumentKindObservableCounter,
				sdkmetric.InstrumentKindHistogram:
				return metricdata.DeltaTemporality
			default:
				return metricdata.CumulativeTemporality
			}
		}
	case "lowmemory":
		return func(k sdkmetric.InstrumentKind) metricdata.Temporality {
			switch k {
			case sdkmetric.InstrumentKindCounter,
				sdkmetric.InstrumentKindHistogram:
				return metricdata.DeltaTemporality
			default:
				return metricdata.CumulativeTemporality
			}
		}
	default: // "cumulative" or anything unrecognized
		return sdkmetric.DefaultTemporalitySelector
	}
}

// aggregationSelector maps OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION
// (already lowercased by config) onto an SDK AggregationSelector. With
// base2_exponential_bucket_histogram (our default, and what Uptrace
// recommends), Histogram instruments use auto-scaling exponential buckets
// (MaxSize 160 / MaxScale 20, the SDK's own defaults for this aggregation)
// instead of fixed explicit boundaries; every other instrument kind keeps the
// SDK default aggregation.
func aggregationSelector(hist string) sdkmetric.AggregationSelector {
	if strings.ToLower(hist) != "base2_exponential_bucket_histogram" {
		return sdkmetric.DefaultAggregationSelector
	}
	return func(k sdkmetric.InstrumentKind) sdkmetric.Aggregation {
		if k == sdkmetric.InstrumentKindHistogram {
			return sdkmetric.AggregationBase2ExponentialHistogram{
				MaxSize:  160,
				MaxScale: 20,
			}
		}
		return sdkmetric.DefaultAggregationSelector(k)
	}
}

func logHTTPOpts(cfg *config.Config) []otlploghttp.Option {
	opts := []otlploghttp.Option{otlploghttp.WithEndpointURL(cfg.Telemetry.EndpointURL)}
	if cfg.Telemetry.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}
	if len(cfg.Telemetry.Headers) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(cfg.Telemetry.Headers))
	}
	if cfg.Telemetry.Compression == "gzip" {
		opts = append(opts, otlploghttp.WithCompression(otlploghttp.GzipCompression))
	}
	return opts
}
scripts/lib.sh
bash
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
detect_runtime() {
  if [ -n "${CR:-}" ] && [ -n "${COMPOSE:-}" ]; then return 0; fi
  if command -v podman >/dev/null 2>&1; then
    CR=podman
    if podman compose version >/dev/null 2>&1; then
      COMPOSE="podman compose"
    elif command -v podman-compose >/dev/null 2>&1; then
      COMPOSE="podman-compose"
    else
      die "podman found but neither 'podman compose' nor 'podman-compose' is available"
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
GOTUNNELS_CSP_POLICY="default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self' https:; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
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
  if [ -z "$p" ]; then p="gotunnels-$(gen_secret 4 | tr 'A-Z' 'a-z' | tr -cd 'a-z0-9' | cut -c1-6)"; fi
  # compose project names must be lowercase alnum/dash/underscore.
  p="$(printf '%s' "$p" | tr 'A-Z' 'a-z' | tr -cd 'a-z0-9_-')"
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
  local i cid status st hs
  log "waiting for '$svc' to become ready (up to ${timeout}s)"
  for i in $(seq 1 "$timeout"); do
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
  local project="$1" svc="$2" timeout="${3:-60}" i cid url
  for i in $(seq 1 "$timeout"); do
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
  local project="$1" svc="$2" pat="$3" timeout="${4:-60}" i cid
  for i in $(seq 1 "$timeout"); do
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
scripts/up.sh
bash
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
#    podman-compose does NOT honor `--no-deps`: step 3's
#    `up -d --no-deps frontend cloudflared-frontend` also created and started
#    the api container (frontend depends_on api), freezing its environment
#    with the pre-discovery values. A plain `up -d` here then hits "name
#    already in use", keeps that stale container, and the API never sees the
#    tunnel-derived GOTUNNELS_RP_ID / RP_ORIGINS / CORS_ALLOWED_ORIGINS just
#    exported above — which is exactly the CORS-on-signup failure (and, worse,
#    silently broken passkeys: RP ID stuck at its bootstrap value). So: if an
#    api container already exists for this project, remove it and let compose
#    recreate it with the current environment. Under docker compose (which
#    honors --no-deps) no such container exists yet and this is a no-op.
_stale_api_cid="$(cid_of "$PROJECT" api)"
if [ -n "$_stale_api_cid" ]; then
  warn "api container was pre-created by a dependency with pre-discovery env; recreating it"
  "$CR" rm -f "$_stale_api_cid" >/dev/null 2>&1 || true
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
scripts/ci-container-test.sh (new file)
bash
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

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pw
please continue

d)"

shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"

detect_runtime
ensure_env
reset_runtime_env

PROJECT="
(
r
e
s
o
l
v
e
p
r
o
j
e
c
t
"
(resolve 
p
​
 roject"{1:-gotunnels-ci}")"
export GOTUNNELS_INSTANCE_ID="{GOTUNNELS_INSTANCE_ID:- PROJECT}"
NET="${PROJECT}_default"

Pinned, fully qualified (podman does not assume docker.io). busybox wget in
the caddy image can't do PUT/DELETE, hence a real curl.
CURL_IMAGE="docker.io/curlimages/curl:8.10.1"

log "container test instance: $PROJECT"

---------------------------------------------------------------------------
teardown + failure diagnostics
---------------------------------------------------------------------------
cleanup() {
status=$?
if [ "status" -ne 0 ] && [ -n " {CR:-}" ]; then
err "container test FAILED (exit $status) — dumping service logs"
for _svc in db api frontend; do
_cid="
(
c
i
d
o
f
"
(cid 
o
​
 f"PROJECT" "$_svc")"
if [ -n "$_cid" ]; then
err "----- logs: $_svc -----"
"
C
R
"
l
o
g
s
−
−
t
a
i
l
120
"
CR"logs−−tail120"_cid" 2>&1 | sed 's/^/ /' >&2 || true
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

---------------------------------------------------------------------------
assertion helpers
---------------------------------------------------------------------------
FAILED=0

ccurl — run curl inside a one-shot container on the compose network. -sS so
transport errors surface; --max-time caps a wedged request, not the test.
ccurl() {
"
C
R
"
r
u
n
−
−
r
m
−
−
n
e
t
w
o
r
k
"
CR"run−−rm−−network"NET" "
C
U
R
L
I
M
A
G
E
"
−
s
S
−
−
m
a
x
−
t
i
m
e
30
"
CURL 
I
​
 MAGE"−sS−−max−time30"@"
}

assert_contains BODY NEEDLE LABEL — grep must DRAIN stdin (>/dev/null, not
-q): with -q it exits at first match, printf takes SIGPIPE, and under
pipefail a PASS would read as FAIL. Same trap as wait_for_log in lib.sh.
assert_contains() {
if printf '%s' "$1" | grep -F -- "$2" >/dev/null; then
ok "PASS: $3"
else
err "FAIL: $3"
err " expected to find: $2"
printf '%s\n' "$1" | head -20 | sed 's/^/ got: /' >&2
FAILED=1
fi
}

assert_icontains() { # case-insensitive variant (headers, HTML)
if printf '%s' "$1" | grep -Fi -- "$2" >/dev/null; then
ok "PASS: $3"
else
err "FAIL: $3"
err " expected to find (case-insensitive): $2"
printf '%s\n' "$1" | head -20 | sed 's/^/ got: /' >&2
FAILED=1
fi
}

assert_not_contains() {
if printf '%s' "$1" | grep -F -- "$2" >/dev/null; then
err "FAIL: $3"
err " expected NOT to find: $2"
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

---------------------------------------------------------------------------
1) build both images
---------------------------------------------------------------------------
log "building images (Containerfile.api + Containerfile.frontend)…"
dc -p "$PROJECT" build

---------------------------------------------------------------------------
2) database, then api + frontend (no tunnels in CI)
---------------------------------------------------------------------------
log "starting database…"
dc -p "$PROJECT" up -d db
wait_healthy "$PROJECT" db 90
pg_isready -U "
P
O
S
T
G
R
E
S
U
S
E
R
:
−
g
o
t
u
n
n
e
l
s
"
−
d
"
POSTGRES 
U
​
 SER:−gotunnels"−d"{POSTGRES_DB:-gotunnels}" -q

log "starting api + frontend…"

podman-compose ignores --no-deps and will re-process db; it prints a
"name already in use" for the existing container but exits 0, so this is
safe under both runtimes (docker compose honors --no-deps outright).
dc -p "$PROJECT" up -d --no-deps api frontend

Unlike up.sh (best-effort warn), CI must hard-fail if the API never comes up.
wait_for_log "$PROJECT" api 'http server listening' 90
|| die "api never logged 'http server listening' — startup failed"
ok "api is listening"

---------------------------------------------------------------------------
3) migrations actually ran
---------------------------------------------------------------------------
db_cid="
(
c
i
d
o
f
"
(cid 
o
​
 f"PROJECT" db)"
[ -n "$db_cid" ] || die "no db container id"
mig="
(
"
("CR" exec "
d
b
c
i
d
"
p
s
q
l
−
U
"
db 
c
​
 id"psql−U"{POSTGRES_USER:-gotunnels}" -d "${POSTGRES_DB:-gotunnels}"
-tAc 'SELECT COALESCE(max(version), 0) FROM schema_migrations' | tr -d '[:space:]')"
log "schema_migrations max(version) = $mig"
if [ "${mig:-0}" -ge 7 ] 2>/dev/null; then
ok "PASS: migrations applied (>= 7)"
else
err "FAIL: expected migration version >= 7, got '$mig'"
FAILED=1
fi

---------------------------------------------------------------------------
4) health endpoints
---------------------------------------------------------------------------
ccurl -f http://api:8080/healthz >/dev/null && ok "PASS: /healthz"
ccurl -f http://api:8080/readyz >/dev/null && ok "PASS: /readyz (db reachable)"

---------------------------------------------------------------------------
5) signup -> bearer token
---------------------------------------------------------------------------
signup_resp="$(ccurl -X POST http://api:8080/api/signup
-H 'Content-Type: application/json'
-d '{"username":"ciuser","password":"ci-password-123","display_name":"CI User"}')"
TOKEN="
(
p
r
i
n
t
f
′
(printf 
′
 signup_resp" | sed -n 's/."token":"([^"])".*/\1/p')"
if [ -n "$TOKEN" ]; then
ok "PASS: signup issued a session token"
else
err "FAIL: signup did not return a token"
printf '%s\n' "$signup_resp" | head -5 | sed 's/^/ got: /' >&2
die "cannot continue without a session token"
fi
AUTH="Authorization: Bearer $TOKEN"

---------------------------------------------------------------------------
6) captcha sync — REGRESSION for the 'unknown + unknown' 500.
Before the ::bigint casts in store.SyncCaptchaStats, every call to this
endpoint failed with {"error":"internal server error"}; both the insert
and the update paths below would have 500ed.
---------------------------------------------------------------------------
sync1="$(ccurl -X POST http://api:8080/api/captcha/sync
-H "$AUTH" -H 'Content-Type: application/json'
-d '{"manual_delta":3,"auto_delta":5,"current_streak":2,"best_streak":4}')"
assert_contains "$sync1" '"total_solves":8' "captcha sync insert path (3 manual + 5 auto = 8)"
assert_contains "$sync1" '"best_streak":4' "captcha sync insert path records best streak"
assert_not_contains "$sync1" 'internal server error' "captcha sync does not 500 (insert)"

sync2="$(ccurl -X POST http://api:8080/api/captcha/sync
-H "$AUTH" -H 'Content-Type: application/json'
-d '{"manual_delta":4,"auto_delta":8,"current_streak":1,"best_streak":2}')"
assert_contains "$sync2" '"total_solves":20' "captcha sync update path accumulates (8 + 12 = 20)"
assert_contains "$sync2" '"best_streak":4' "captcha sync keeps GREATEST best streak (4 > 2)"
assert_contains "$sync2" '"current_streak":1' "captcha sync current streak is last-write-wins"
assert_not_contains "$sync2" 'internal server error' "captcha sync does not 500 (update)"

stats="
(
c
c
u
r
l
−
H
"
(ccurl−H"AUTH" http://api:8080/api/captcha/stats)"
assert_contains "$stats" '"total_solves":20' "captcha stats reads back the synced totals"

lb="
(
c
c
u
r
l
−
H
"
(ccurl−H"AUTH" http://api:8080/api/captcha/leaderboard)"
assert_contains "$lb" '"username":"ciuser"' "leaderboard ranks the CI user"

---------------------------------------------------------------------------
7) prefs round trip
---------------------------------------------------------------------------
put_pref="$(ccurl -X PUT http://api:8080/api/prefs/theme
-H "$AUTH" -H 'Content-Type: application/json' -d '{"value":"dark"}')"
assert_contains "$put_pref" '"status":"saved"' "prefs PUT saves"

get_pref="
(
c
c
u
r
l
−
H
"
(ccurl−H"AUTH" http://api:8080/api/prefs/theme)"
assert_contains "$get_pref" '"value":"dark"' "prefs GET returns the stored value"
assert_contains "$get_pref" '"exists":true' "prefs GET reports exists=true"

---------------------------------------------------------------------------
8) notes create / list / delete
---------------------------------------------------------------------------
note_resp="$(ccurl -X POST http://api:8080/api/notes
-H "$AUTH" -H 'Content-Type: application/json'
-d '{"body":"hello from the CI smoke test"}')"
NOTE_ID="
(
p
r
i
n
t
f
′
(printf 
′
 note_resp" | sed -n 's/."note":{"id":([0-9][0-9]).*/\1/p')"
if [ -n "$NOTE_ID" ]; then
ok "PASS: note created (id=$NOTE_ID)"
else
err "FAIL: note creation did not return a numeric id"
printf '%s\n' "$note_resp" | head -5 | sed 's/^/ got: /' >&2
FAILED=1
fi

notes_list="
(
c
c
u
r
l
−
H
"
(ccurl−H"AUTH" http://api:8080/api/notes)"
assert_contains "$notes_list" 'hello from the CI smoke test' "notes list contains the new note"

if [ -n "$NOTE_ID" ]; then
del_resp="$(ccurl -X DELETE "http://api:8080/api/notes/$NOTE_ID" -H "$AUTH")"
assert_contains "$del_resp" '"status":"deleted"' "note delete (hard delete semantics)"
notes_after="
(
c
c
u
r
l
−
H
"
(ccurl−H"AUTH" http://api:8080/api/notes)"
assert_not_contains "$notes_after" 'hello from the CI smoke test' "deleted note is gone from the list"
fi

---------------------------------------------------------------------------
9) CORS preflight (the middleware must answer OPTIONS itself with 204)
---------------------------------------------------------------------------
pre_code="$(ccurl -o /dev/null -w '%{http_code}' -X OPTIONS http://api:8080/api/signup
-H 'Origin: https://example.com' -H 'Access-Control-Request-Method: POST')"
assert_eq "$pre_code" "204" "CORS preflight answered with 204"

---------------------------------------------------------------------------
10) frontend serves the app with the CSP header
---------------------------------------------------------------------------
front="$(ccurl -D - http://frontend:8080/)"
assert_icontains "$front" 'content-security-policy' "frontend sends a CSP header"
assert_icontains "$front" '<html' "frontend serves the HTML app"

---------------------------------------------------------------------------
verdict
---------------------------------------------------------------------------
if [ "$FAILED" -ne 0 ]; then
die "container smoke test had failures (see FAIL lines above)"
fi
ok "container smoke test passed."


**Wait — stop.** I spotted a typo I introduced while pasting: line 22 above says `pw d)` instead of `pwd)`. The file on disk is correct (it passed `bash -n` and shellcheck); that corruption happened only in this chat rendering. Use this corrected line 22:

```bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
```

Everything else in the block above is byte-identical to the validated file.

## `.github/workflows/ci.yaml`

```yaml
# Build and test. Deliberately thin: all logic lives in scripts/test.sh so the
# exact same commands run locally and here.
name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7

      - uses: actions/setup-go@v6
        with:
          go-version: "stable"
          # go.sum is intentionally absent on first import; do not fail on it.
          cache: false

      - name: Build, vet, unit tests
        run: bash scripts/test.sh all

  # Everything scripts/test.sh cannot see: that both Containerfiles build,
  # that compose wiring + env substitution is correct, that migrations run
  # against a real Postgres, and that the HTTP surface works end to end —
  # including the captcha-sync regression that only reproduces against real
  # Postgres via the pgx extended protocol. Ubuntu runners ship podman;
  # podman-compose comes from pipx. No tunnels are started and no host ports
  # are published; assertions run from a curl container on the compose network.
  containers:
    runs-on: ubuntu-latest
    timeout-minutes: 25
    steps:
      - uses: actions/checkout@v7

      - name: Install podman-compose
        run: pipx install podman-compose

      - name: Runtime versions
        run: |
          podman version
          podman-compose version

      - name: Build images + container smoke test
        run: bash scripts/ci-container-test.sh gotunnels-ci
```

## `compose.yaml`

```yaml
# compose.yaml — the whole GoTunnels stack.
#
# Design rules embodied here:
#   * No service publishes a host port. The only traffic leaving the machine is
#     the two outbound Cloudflare Quick Tunnel connections.
#   * Everything is addressed by service name on the internal Compose network.
#   * Every instance is isolated by the Compose project name (passed via -p),
#     which prefixes container, network, and volume names — so N instances run
#     side by side without collision.
#   * Values that can be generated are generated by scripts/up.sh into .env;
#     nothing secret is hardcoded here.
#
# Runtime-discovered values (the frontend's tunnel URL, used for the WebAuthn
# RP ID and CORS allow-list) are exported into the shell by scripts/up.sh
# before the api service is (re)started, and read here via ${VAR} substitution.

services:
  db:
    image: docker.io/library/postgres:16-alpine
    environment:
      POSTGRES_USER: "${POSTGRES_USER:-gotunnels}"
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD must be set (run scripts/up.sh)}"
      POSTGRES_DB: "${POSTGRES_DB:-gotunnels}"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-gotunnels} -d ${POSTGRES_DB:-gotunnels}"]
      interval: 3s
      timeout: 3s
      retries: 20
    restart: unless-stopped

  api:
    build:
      context: .
      dockerfile: Containerfile.api
      args:
        VERSION: "${GOTUNNELS_VERSION:-dev}"
    depends_on:
      db:
        condition: service_healthy
    environment:
      GOTUNNELS_INSTANCE_ID: "${GOTUNNELS_INSTANCE_ID:-default}"
      GOTUNNELS_VERSION: "${GOTUNNELS_VERSION:-dev}"
      GOTUNNELS_HTTP_ADDR: ":8080"
      DATABASE_URL: "postgres://${POSTGRES_USER:-gotunnels}:${POSTGRES_PASSWORD}@db:5432/${POSTGRES_DB:-gotunnels}?sslmode=disable"
      GOTUNNELS_IP_HASH_PEPPER: "${GOTUNNELS_IP_HASH_PEPPER:?set by scripts/up.sh}"
      GOTUNNELS_TOTP_ENCRYPTION_KEY: "${GOTUNNELS_TOTP_ENCRYPTION_KEY:?set by scripts/up.sh}"
      # Discovered at runtime by scripts/up.sh (frontend tunnel URL). Defaults
      # keep `compose config` valid before discovery.
      GOTUNNELS_RP_ID: "${GOTUNNELS_RP_ID:-localhost}"
      GOTUNNELS_RP_DISPLAY_NAME: "${GOTUNNELS_RP_DISPLAY_NAME:-GoTunnels}"
      GOTUNNELS_RP_ORIGINS: "${GOTUNNELS_RP_ORIGINS:-http://localhost:8080}"
      GOTUNNELS_CORS_ALLOWED_ORIGINS: "${GOTUNNELS_CORS_ALLOWED_ORIGINS:-*}"
      GOTUNNELS_CSP_MODE: "${GOTUNNELS_CSP_MODE:-report-only}"
      # Telemetry: a single Uptrace (cloud or self-hosted) DSN, or OTLP env vars.
      UPTRACE_DSN: "${UPTRACE_DSN:-}"
      OTEL_EXPORTER_OTLP_ENDPOINT: "${OTEL_EXPORTER_OTLP_ENDPOINT:-}"
      OTEL_EXPORTER_OTLP_HEADERS: "${OTEL_EXPORTER_OTLP_HEADERS:-}"
      OTEL_EXPORTER_OTLP_COMPRESSION: "${OTEL_EXPORTER_OTLP_COMPRESSION:-gzip}"
      OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE: "${OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE:-delta}"
      OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION: "${OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION:-base2_exponential_bucket_histogram}"
      OTEL_SERVICE_NAME: "${OTEL_SERVICE_NAME:-gotunnels-api}"
      GOTUNNELS_DEV: "${GOTUNNELS_DEV:-false}"
    restart: unless-stopped

  frontend:
    build:
      context: .
      dockerfile: Containerfile.frontend
    depends_on:
      - api
    environment:
      GOTUNNELS_CSP_HEADER_NAME: "${GOTUNNELS_CSP_HEADER_NAME:-Content-Security-Policy-Report-Only}"
      GOTUNNELS_CSP_POLICY: "${GOTUNNELS_CSP_POLICY:-default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self' https:; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'}"
    restart: unless-stopped

  cloudflared-frontend:
    image: docker.io/cloudflare/cloudflared:latest
    depends_on:
      - frontend
    command: ["tunnel", "--no-autoupdate", "--url", "http://frontend:8080"]
    restart: unless-stopped

  cloudflared-api:
    image: docker.io/cloudflare/cloudflared:latest
    depends_on:
      - api
    command: ["tunnel", "--no-autoupdate", "--url", "http://api:8080"]
    restart: unless-stopped

volumes:
  pgdata:
```

## `.env.example`

```bash
# .env.example — shape of the environment file.
#
# There are NO real secrets to copy here. scripts/up.sh generates a .env with
# fresh random secrets on first run if one does not already exist. This file
# exists only to document what the stack reads.
#
# Copy to .env and fill in ONLY if you are not using scripts/up.sh:
#   cp .env.example .env

# --- Instance identity -------------------------------------------------------
# Unique per running instance; prefixes containers/network/volumes and appears
# as an OpenTelemetry resource attribute.
GOTUNNELS_INSTANCE_ID=default
GOTUNNELS_VERSION=dev

# --- Database (password is generated by up.sh) -------------------------------
POSTGRES_USER=gotunnels
POSTGRES_DB=gotunnels
POSTGRES_PASSWORD=generate-me

# --- Server secrets (generated by up.sh; never commit real values) -----------
# Pepper mixed into sha256(pepper||ip) so stored IP hashes are not rainbow-tableable.
GOTUNNELS_IP_HASH_PEPPER=generate-me
# Any string; the app derives a 32-byte AES key from it to encrypt TOTP secrets.
GOTUNNELS_TOTP_ENCRYPTION_KEY=generate-me

# --- WebAuthn / CORS (discovered at runtime from the frontend tunnel URL) ----
# up.sh overwrites these once the frontend Quick Tunnel URL is known.
GOTUNNELS_RP_ID=localhost
GOTUNNELS_RP_DISPLAY_NAME=GoTunnels
GOTUNNELS_RP_ORIGINS=http://localhost:8080
GOTUNNELS_CORS_ALLOWED_ORIGINS=*

# --- Content Security Policy (central configuration) -------------------------
# Header name selects report-only vs enforcing:
#   Content-Security-Policy-Report-Only  (default; observe only)
#   Content-Security-Policy              (enforce)
GOTUNNELS_CSP_HEADER_NAME=Content-Security-Policy-Report-Only
GOTUNNELS_CSP_MODE=report-only
GOTUNNELS_CSP_POLICY="default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self' https:; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"

# --- Telemetry (optional) ----------------------------------------------------
# Point at Uptrace Cloud (or any OTLP/HTTP backend) with a single DSN:
#   UPTRACE_DSN=https://<token>@api.uptrace.dev
# Or use the standard OTLP env vars instead:
#   OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp.example.com
#   OTEL_EXPORTER_OTLP_HEADERS=key1=value1,key2=value2
# Leave blank to disable exporting (the app still logs JSON to stdout).
#
# NOTE: exporting UPTRACE_DSN in your shell before running scripts/up.sh also
# works — ensure_env persists it into this file before loading it, so a blank
# UPTRACE_DSN= line here can no longer clobber your export.
UPTRACE_DSN=
OTEL_EXPORTER_OTLP_ENDPOINT=
OTEL_EXPORTER_OTLP_HEADERS=
# gzip-compress OTLP payloads (Uptrace recommendation; set empty to disable).
OTEL_EXPORTER_OTLP_COMPRESSION=gzip
# Metrics temporality: delta | cumulative | lowmemory (spec casing DELTA etc.
# is also accepted). Uptrace prefers delta.
OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta
# Histogram aggregation for Histogram instruments:
#   base2_exponential_bucket_histogram (Uptrace recommendation; default)
#   explicit_bucket_histogram
OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION=base2_exponential_bucket_histogram
OTEL_SERVICE_NAME=gotunnels-api

# --- Dev convenience ---------------------------------------------------------
# When true, missing secrets are auto-generated ephemerally and cookies are not
# marked Secure. Never set true in anything resembling production.
GOTUNNELS_DEV=false
```

## `docs/CONFIGURATION.md`

```markdown
# Configuration

All runtime configuration is centralized. On the Go side every variable is read
in exactly one place — [`internal/config/config.go`](../internal/config/config.go)
— and nowhere else. This file is the reference for what each variable does.

Most values are generated for you by `scripts/up.sh` on first run and written to
a gitignored `.env`. You only ever set variables by hand if you are bypassing
the scripts. See also [`.env.example`](../.env.example).

## Identity

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOTUNNELS_INSTANCE_ID` | hostname / `default` | Unique per instance; also an OTel resource attribute and used to derive the Compose project name. |
| `GOTUNNELS_VERSION` | `dev` | Build/version label (also settable at build time via `-ldflags -X main.version`). |

## HTTP

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOTUNNELS_HTTP_ADDR` | `:8080` | Listen address inside the container. |
| `GOTUNNELS_SHUTDOWN_TIMEOUT` | `15s` | Graceful shutdown budget. |

## Database

| Variable | Default | Meaning |
| --- | --- | --- |
| `DATABASE_URL` | (required) | pgx connection string. Compose builds it from the `POSTGRES_*` values. |
| `POSTGRES_USER` | `gotunnels` | DB user (Compose/Postgres). |
| `POSTGRES_DB` | `gotunnels` | DB name. |
| `POSTGRES_PASSWORD` | generated | DB password; generated by `up.sh`. |
| `GOTUNNELS_DB_MAX_CONNS` | `20` | Pool max connections. |
| `GOTUNNELS_DB_MIN_CONNS` | `2` | Pool min connections. |
| `GOTUNNELS_DB_CONNECT_TIMEOUT` | `30s` | How long startup retries connecting before giving up. |

## Sessions

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOTUNNELS_SESSION_COOKIE_NAME` | `gotunnels_session` | Name of the secondary session cookie. |
| `GOTUNNELS_SESSION_TTL` | `24h` | Session lifetime. |

## Secrets (generated per instance; never commit real values)

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOTUNNELS_IP_HASH_PEPPER` | required (unless `GOTUNNELS_DEV`) | Mixed into `sha256(pepper||ip)` so stored IP hashes are not rainbow-tableable. |
| `GOTUNNELS_TOTP_ENCRYPTION_KEY` | required (unless `GOTUNNELS_DEV`) | Any string; a 32-byte AES-256-GCM key is derived from it (SHA-256) to encrypt TOTP secrets at rest. |

When `GOTUNNELS_DEV=true`, missing secrets are generated ephemerally at startup
with a warning (fine for local hacking, never for anything real).

## WebAuthn / CORS (discovered at runtime)

`scripts/up.sh` overwrites these once the frontend Quick Tunnel URL is known.

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOTUNNELS_RP_ID` | `localhost` | WebAuthn Relying Party ID = the frontend's registrable domain. |
| `GOTUNNELS_RP_DISPLAY_NAME` | `GoTunnels` | Human-readable RP name (also the TOTP issuer). |
| `GOTUNNELS_RP_ORIGINS` | `http://localhost:8080` | Full origin(s) the browser presents (comma/space separated). |
| `GOTUNNELS_CORS_ALLOWED_ORIGINS` | `*` | Exact allowed origin(s), or `*`. Because credentials are used, `*` is echoed as the request origin rather than literally, and never treated as authorization. |

## Content Security Policy (central)

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOTUNNELS_CSP_HEADER_NAME` | `Content-Security-Policy-Report-Only` | Set to `Content-Security-Policy` to **enforce**. |
| `GOTUNNELS_CSP_MODE` | `report-only` | Informational mirror surfaced on `/api/info` (`report-only` or `enforce`). |
| `GOTUNNELS_CSP_POLICY` | strict self-only (see below) | The policy string. Emitted by Caddy; mirrored to the API for `/api/info`. |

The default policy is fully self-hosted — no third-party scripts, styles,
images, fonts, media, or frames:

```
default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self';
font-src 'self'; connect-src 'self' https:; media-src 'self'; object-src 'none';
frame-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'
Because the app has no inline scripts, inline event handlers, or inline styles,
it already satisfies the enforcing form of this policy.

Telemetry (optional)
Resolution order: UPTRACE_DSN first, then the standard OTLP variables, then
disabled (stdout logging only).

Variable	Default	Meaning
UPTRACE_DSN	(empty)	e.g. https://<token>@api.uptrace.dev. Derives the OTLP endpoint and sets the uptrace-dsn header.
OTEL_EXPORTER_OTLP_ENDPOINT	(empty)	Base OTLP/HTTP endpoint, if not using a DSN.
OTEL_EXPORTER_OTLP_HEADERS	(empty)	k1=v1,k2=v2 headers for the exporter.
OTEL_EXPORTER_OTLP_COMPRESSION	gzip	Exporter compression.
OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE	delta	delta, cumulative, or lowmemory (spec casing like DELTA also accepted). Uptrace prefers delta.
OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION	base2_exponential_bucket_histogram	Or explicit_bucket_histogram. Exponential buckets give Uptrace accurate percentiles at lower cost.
OTEL_SERVICE_NAME / GOTUNNELS_SERVICE_NAME	gotunnels-api	OTel service.name.
Dev
Variable	Default	Meaning
GOTUNNELS_DEV	false	Auto-generate missing secrets; mark cookies non-Secure. Local use only.
Script-only variables
These are read by the shell scripts (not the Go app) to control instance
naming and the test toolchain:

Variable	Meaning
GOTUNNELS_PROJECT	Explicit Compose project name (else derived from the instance id or generated).
GO_IMAGE	Go image used by scripts/test.sh when no host Go toolchain is present (default golang:1.26-bookworm).
CR / COMPOSE	Override the detected container runtime / compose command.

---

## Deploying and housekeeping

To redeploy: `git pull && bash scripts/up.sh` — no need to delete `.env` anymore, and doing so is now actively worse (you'd lose the persisted DSN and secrets). The first run after this fix will restore passkey functionality, since the api container will finally boot with the real tunnel-derived RP ID. Existing passkeys registered while RP ID was `localhost` won't work against the tunnel origin and will need re-registering — that's inherent to WebAuthn, not a bug here.

Two credentials need rotating: the GitHub PAT you pasted this session, and the Uptrace DSN — the DSN also sits in plaintext inside `virginia.txt`, so treat both as exposed regardless of whether the files stay private. Neither appears in any file above (I scanned all eleven).

One heads-up on the CI job: `pipx install podman-compose` pulls the latest podman-compose, which may behave slightly differently from your Fedora package around the "name already in use" tolerance in step 2 of the smoke test. If the `containers` job ever fails there, the log dump in the cleanup trap will show exactly which service, and pinning `pipx install podman-compose==1.2.0` is the one-line remedy.


Want to be notified when Claude responds?






Claude is AI and can make mistakes. Please double-check responses.




83
17
31









































































83
17
31

Can you please make sure the reply is complete and give me all the files? 


/
Claude finished the response
This is great progress. 
However, there are a few issues. 
First, I got some CORS error when signing up 
Turns out that was because I still had the old .env file
the workaround there was to delete the env file 
bash
kushal@virginia:~/src/golang/GoTunnels$ cd ~/src/golang/GoTunnels/; time git status; time git fetch; time git status; time git remote show origin; time git pull; time git status; export UPTRACE_DSN="https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"; cat .env; rm .env; time bash scripts/up.sh; podman ps -a;

I was able to sign up to the website after that. 
Next problem was when I was in the captcha page, 
I noticed sync failed. 
upon closer inspection, I observed that the sync endpoint always returned 500 error 
json
{"error":"internal server error"}

also please check that we are sending all logs, traces, spans, metrics to uptrace 
here is some guidance
bash
# Uncomment the appropriate protocol for your programming language.
# Only for OTLP/gRPC.
#export OTEL_EXPORTER_OTLP_ENDPOINT="https://otlp.uptrace.dev:4317"
# Only for OTLP/HTTP.
#export OTEL_EXPORTER_OTLP_ENDPOINT="https://otlp.uptrace.dev"
# Pass Uptrace DSN in gRPC/HTTP headers.
export OTEL_EXPORTER_OTLP_HEADERS="uptrace-dsn=https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"
# Enable gzip compression.
export OTEL_EXPORTER_OTLP_COMPRESSION=gzip
# Enable exponential histograms.
export OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION=BASE2_EXPONENTIAL_BUCKET_HISTOGRAM
# Prefer delta temporality.
export OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=DELTA

You can see the code in dump.txt and my deployment at virginia.txt
Also please review all the tests once again. 
Among others, please check if the github actions actually test our container files and podman compose process. 
Please return FULL files for all files that need to change.

Show less

Diagnosed SQL parameter type bug, traced CORS/env persistence issues, and architected telemetry configuration
Diagnosed SQL parameter type bug, traced CORS/env persistence issues, and architected telemetry configuration
I need to review the project structure and identify the main issues: there's a CORS problem during signup that required deleting the .env file as a workaround, the sync endpoint is consistently throwing 500 errors on the captcha page, and I need to verify that all logging, tracing, spans, and metrics are being captured properly.

The user shared credentials in plaintext—a GitHub PAT and an Uptrace DSN—which should be rotated immediately per security best practices, though they mentioned not to worry about the PAT. I'll avoid embedding the PAT in any files and can use it only for read-only API calls via bash. Let me start by examining the dump.txt and virginia.txt files to understand the current codebase and deployment setup.


Checking project files exist and their sizes

Script
I should start by examining the structure of dump.txt to understand what I'm working with.


Listing all files in the dump to understand project structure

Script

Inspecting dump.txt format around a file marker to write an accurate extractor

Script

Viewing more of the dump format to find content delimiters

Script
I'm working through the file format structure to build an extractor that identifies the --- CONTENT --- marker following each # FILE: header and captures everything until the next file block begins, being careful to strip any trailing blank lines.


Checking the boundary between two files in the dump

Script

Extracting all files from dump.txt into a working repo directory

Script
Now I'm examining the deployment log to identify the specific errors, particularly the sync 500 error and CORS issue.

All 91 files extracted. Now I'll check the deployment log to see what errors occurred.


Reading the deployment log from virginia.txt

Script

Searching deployment log for errors, CORS, sync failures

Script
I'm spotting container name collision errors in the deployment logs. There's a stale tunnel URL lingering in the environment variables from an old .env file that needs to be cleaned up. Let me dig deeper into the logs around that section to understand the full scope of the sync failure.


Viewing the section around the old .env file and its CORS config

Script

Viewing the rest of the deployment log after the successful bring-up

Script
Now I'm diving into the captcha sync handler to track down what's causing that 500 error. I need to examine the captcha implementation, the store logic, and the migration that set up the schema to find where things are breaking.

I'm examining the captcha sync handler and store layer to locate the 500 error.


Reading the captcha handler code to find the sync 500 bug

Script

Finding captcha-related store functions

Script

Reading the SyncCaptchaStats SQL and surrounding store code

Script
I need to check the migration 0006 file to understand the captcha_stats table structure — the issue could be a mismatch between the INSERT statement and the actual schema, like incorrect column names or missing defaults.


Reading migration 0006 (captcha_stats table), 0007, and the embed logic

Script
I'm spotting a bug in the captcha stats insertion query where the manual solves parameter is being added to the total instead of used directly, and the parameters being passed seem to be cut off mid-way through. I'm working through how PostgreSQL resolves parameter types across the entire statement. Each parameter gets assigned a single type that applies everywhere it's used, with direct column assignments providing strong type hints. The tricky part is the order of analysis — PostgreSQL processes the VALUES list left to right, so when it encounters $4 + $5 in the total_solves position, both parameters are still unknown types. This creates an ambiguity with the + operator since there's no clear way to resolve unknown + unknown to a specific operator, which typically triggers a "operator is not unique" error.

This parse-time failure would explain why the sync endpoint consistently returns 500 errors — the SQL statement itself is malformed and fails during parsing before any execution happens. The unit tests wouldn't catch this because they only test the clamping logic in-process without hitting the database, so the SQL error never surfaces during testing. operator is not unique error when Postgres tries to resolve the + operator with two unknown parameter types. The extended protocol that pgx uses forces parameter type resolution at parse time, so every call to SyncCaptchaStats fails immediately with a 500 error. The solution is to add explicit casts like ($4::bigint + $5::bigint) to disambiguate the operator. I should also add casts in the UPDATE clause for consistency. Let me reconsider whether $2 and $3 actually need casts in the VALUES clause — they're being assigned directly to bigint columns, so assignment context should handle the type coercion fine. The real issue is the expression $4 + $5 where operator resolution happens before assignment coercion can kick in. I want to verify this behavior empirically if I can get Postgres running locally, and I should also check whether pgx uses prepared statements by default.

pgx v5 defaults to QueryExecModeCacheStatement, which means it's using the extended protocol and prepared statements — that would definitely trigger the unknown + unknown error at the parse level. I should search through store.go for other similar expressions with the pattern $ followed by  + $ to see if there are more places with this issue. Let me also check what's actually being tested in the test files, particularly around the captcha logic, and then look at how the .env file is being managed in the setup scripts since that was the workaround for the CORS issue.

The real problem with the .env file is that it was persisting stale tunnel URLs from a previous run — Quick Tunnels generate a new URL each time, so the old GOTUNNELS_RP_ID, GOTUNNELS_RP_ORIGINS, and GOTUNNELS_CORS_ALLOWED_ORIGINS values were causing CORS and WebAuthn failures. The fix should be to regenerate these tunnel-dependent values after discovering the new tunnel URL while keeping stable secrets like POSTGRES_PASSWORD intact. Let me check how the scripts currently handle this.

Now looking at the telemetry setup — they're using OTLP HTTP exporters for traces, metrics, and logs. Uptrace expects the endpoint at https://otlp.uptrace.dev for HTTP (or https://otlp.uptrace.dev:4317 for gRPC), with the DSN passed as an uptrace-dsn header. Let me examine how telemetry.go configures this.

I also need to check what environment variables are passed to the API container in compose.yaml, and then configure the exporters with gzip compression, exponential histograms, and delta temporality — which in the Go SDK means setting the temporality selector, aggregation selector, and compression options on the metric and trace exporters. Both env vars are supported by the Go OTLP metric exporters, so I'll set these defaults in compose.yaml and telemetry.go for the cleanest approach. Now I need to review the test setup and GitHub Actions workflows to see what's currently being tested and add coverage for the container files and podman compose.

I'll need to create a workflow that builds both Containerfiles with podman, spins up the full stack with podman-compose, waits for health checks, hits the health endpoint to verify everything's working, then tears it down. Since ubuntu-24.04 runners have podman preinstalled, I just need to install podman-compose via pip and handle the cloudflared container requirements.

The tricky part is that cloudflared needs outbound internet access, which might be flaky or rate-limited in CI. A better approach would be to skip cloudflared entirely in the CI test — I can use podman compose up -d db api frontend to start only the essential services while letting compose handle their dependencies. I should check the existing scripts first to see if there's already a CI mode, or if I need to write a targeted test workflow from scratch.

For verifying the API is healthy, since there are no published host ports by design, I'll need to either exec into a container to run a health check, or spin up a curl container on the same network to test the endpoint. Let me check the compose file and the API's main.go to see what health check mechanisms are available.

Now I need to review the telemetry configuration to confirm we're sending all logs, traces, spans, and metrics to Uptrace, and check how the UPTRACE_DSN is wired up in the compose environment.

Show less
Done
The store SQL has my prime suspect. Let me now read the telemetry, scripts, compose, and workflows.


Unraveling the .env lifecycle bug's cascading consequences
Unraveling the .env lifecycle bug's cascading consequences

Reading telemetry setup to check Uptrace OTLP export configuration

Script

Finding telemetry config parsing including Uptrace DSN handling

Script

Reading how UPTRACE_DSN is resolved into endpoint and headers

Script

Reading endpointFromDSN and compose.yaml env wiring

Script
I need to examine the up.sh and lib.sh scripts to see how the .env file gets stale with old tunnel URLs and how the API restarts with the newly discovered ones.


Reading up.sh to understand env generation and stale .env issue

Script
I need to clean up the extracted up.sh file by removing the dump footer that got included, and I'm realizing the real issue: when ensure_env runs and finds a stale .env from a previous execution, it's not handling that scenario properly.

Looking at the logs, the first run failed partway through with container name conflicts — the script never reached the steps where it would refresh the RP and CORS values, so those stale settings persisted. The project name gotunnels-novzg appears to be generated with a random suffix, and when containers with that name already existed from the failed run, the next attempt couldn't proceed. I'm noticing the collision errors for db and api containers happen within the same run, which suggests podman-compose might be attempting to recreate all containers even when using --no-deps, rather than respecting the existing ones from earlier steps. ves the stale GOTUNNELS_INSTANCE_ID from the previous run. Looking at the logs, the first run shows using existing /home/kushal/src/golang/GoTunnels/.env with project (instance): gotunnels-novzg, but the dumped .env after down.sh still has GOTUNNELS_INSTANCE_ID=default. This suggests the set_env_var call isn't properly updating the instance ID in the .env file, or there's something odd happening with how the variable gets written.

Actually, I think I see the issue now: if GOTUNNELS_INSTANCE_ID was already set to "default" in the stale .env (possibly from .env.example or an older generation), then when ensure_env sources it, the variable stays "default" because the export line uses ${GOTUNNELS_INSTANCE_ID:-$PROJECT}, which keeps the existing value. And looking at resolve_project, it seems to generate a random suffix like novzg or ekh each run unless GOTUNNELS_PROJECT is explicitly set, so the project names are different every time.

The container name collision on the first run—gotunnels-novzg_db_1 is already in use—suggests a leftover container from a previous session with the same random suffix, which is unlikely with 5-character randomness, OR it's the same-run double-create behavior from podman-compose. The error happens right after the build succeeds and before the database even starts, which points to a stale container from an earlier session still hanging around.

I'm wondering if the .env file persists the GOTUNNELS_PROJECT suffix across runs—if resolve_project stores gotunnels-novzg in .env, then re-running with that stale .env would reuse the same suffix and collide with containers still existing from the previous run. The dumped .env doesn't show GOTUNNELS_PROJECT, but down.sh might be removing it. I need to check lib.sh to understand how the project name and suffix are managed, and I should also re-extract up.sh properly since the previous dump got contaminated with trailing summary output.

Show less
Done
The extracted up.sh has the dump footer appended. Let me fix that, then read lib.sh.


Unraveling podman-compose's service creation and dependency behavior
Unraveling podman-compose's service creation and dependency behavior

Removing the dump summary footer from the extracted up.sh

Script

Reading lib.sh to understand env generation, project resolution, and helpers

Script
So the bug stems from .env persistence—ensure_env preserves the existing file to keep secrets safe, but resolve_project generates a fresh random project name each time since the stale .env still has GOTUNNELS_INSTANCE_ID=default. This leaves old tunnel configuration values like GOTUNNELS_RP_ID, GOTUNNELS_RP_ORIGINS, and GOTUNNELS_CORS_ALLOWED_ORIGINS from the previous run lingering in the file.

The container collision happens because up.sh only overwrites those stale values after discovering the frontend tunnel in step 5—but if the run fails before that point (like with the container name conflicts), or if only the database is started with up -d db, those old values stay put. Looking at the virginia.txt log, the error about gotunnels-novzg_db_1 being in use is puzzling since novzg was just resolved this run, unless up.sh ran twice with the same .env or there's a leftover container from a previous execution.

Actually, I think I see it now—podman-compose's --no-deps flag might not prevent creation of all services in some versions. So when step 3 tries up -d --no-deps frontend cloudflared-frontend, it could be creating the db again even though it already exists from step 2, causing the "name already in use" error. Then step 6 does the same thing, which explains why both db_1 and api_1 show as already in use even though api was supposedly only created in step 6 itself.

Show less
Done
Let me study the first failed run in the log more carefully to understand the exact failure sequence.


Scrutinizing database query execution and health check endpoints
Scrutinizing database query execution and health check endpoints

Reading the first failed run sequence in detail

Script
I'm seeing the actual behavior of podman-compose now—when running up -d with specific services, it creates the entire dependency chain, and even with --no-deps it still attempts to create all listed services plus their dependencies, generating "already in use" errors for ones that exist but continuing anyway. At step 3, it even created gotunnels-novzg unexpectedly.

The real issue is that --no-deps isn't being honored—podman-compose created and started the API container at step 3 before the frontend tunnel URL was even known. This is the CORS bug's root cause: the API was initialized with whatever environment variables existed at that moment (stale .env values or localhost defaults), and when step 6 tried to bring up the API again with --no-deps, podman-compose found it already running and just started it without recreating it. Unlike docker compose, podman-compose doesn't detect config changes or recreate containers based on environment variable updates, so the API kept running with the stale CORS origins captured from step 3. When the .env was deleted and regenerated with GOTUNNELS_CORS_ALLOWED_ORIGINS=*, the API that was already created at step 3 had already locked in its CORS configuration. So the real issue is that podman-compose creates the api container at step 3 when the frontend declares its dependency, and then at step 6 when we try to recreate it with the discovered tunnel URL, the container's already running so the create fails and start does nothing. The solution is to force-recreate the api container after we've set the environment variables with the actual frontend URL — using --force-recreate --no-deps flags on the up command, assuming podman-compose 1.x supports that option.

But there's a safer approach that doesn't rely on specific flag support: explicitly remove the api container if it exists before bringing it back up, which ensures it gets created fresh with the new environment variables. The tricky part is handling the dependency chain — api depends on db with a health check condition, and cloudflared-api depends on api, so I need to make sure the startup order works correctly when recreating these containers.

Actually, looking back at the logs, the health check didn't cause a hang, and podman-compose tends to ignore depends_on conditions anyway, mostly just ordering creation. So maybe the better strategy is to stop fighting the tool's behavior and instead always recreate api after discovering the URL, since that's when the environment changes. Frontend doesn't need the same treatment since it gets its config injected via podman cp later. The real issue is that podman-compose ignores both --no-deps and depends_on conditions, so I can't rely on those flags to control what gets created.

Writing bash script... After removing the stale api container, podman-compose up -d will recreate it fresh with the current environment variables. I could use --force-recreate, but explicitly removing the container with rm -f is more robust across different versions of podman-compose and docker compose. With docker compose specifically, the api container won't even exist at step 6 since --no-deps skipped it earlier, so the rm becomes a no-op. I'm also wondering about the db health condition when creating the api container at step 6 with --no-deps — that might be an issue to consider.

Actually, --no-deps skips starting dependencies, but the db is already running, so docker compose should still respect the depends_on conditions. That should be fine. Beyond the recreation fix, I should make ensure_env more robust by resetting tunnel-derived values to safe defaults on every run, since these are per-run values, not secrets. If a run crashes mid-way, stale tunnel values in .env could cause the api to use outdated origins on the next run, even though recreation at step 6 would fix it anyway. Being explicit about resetting these values makes the script more self-documenting and safer.

I'm setting the tunnel-derived environment variables to sensible defaults and wondering whether the wildcard CORS setting is appropriate. Since the auth uses Bearer tokens in the Authorization header rather than cookies, and sendBeacon can't carry Authorization headers, allowing any origin with Access-Control-Allow-Origin: * should be safe.

The real issue is that ensure_env gets called by multiple scripts, but only up.sh should reset per-run values like the tunnel URL. I need to call the reset function right after ensure_env in up.sh and re-export those variables to override what was loaded from the environment file.

I'm creating a helper function that resets the tunnel-derived values to safe defaults — the RP ID back to localhost, CORS origins to the local dev server, and clearing the tunnel URL itself. This way stale values from a previous run won't silently break things with opaque CORS errors.

Now I'm thinking through the sync 500 bug hypothesis: the issue might be with $4 + $5 when those parameters have unknown types. I should test this empirically by setting up a real Postgres instance and running the exact query through the extended protocol to see if type coercion is the culprit. $5` would work. Let me check database.go for the QueryExecMode setting.

For the Uptrace telemetry config, I need to verify the correct endpoint — the DSN resolves to https://api.uptrace.dev with the uptrace-dsn header, but Uptrace's actual OTLP/HTTP endpoint should be https://otlp.uptrace.dev, not api.uptrace.dev. Let me check their current documentation to confirm the right endpoint format. Fixing the endpoint mapping logic...

For self-hosted instances, I'll keep the scheme and host as-is since they typically listen on the same port. I also need to add the exporter tuning settings from the guidance—gzip compression is already handled by default, but I should set the exponential histogram aggregation environment variable. The Go OTLP metrics exporter should pick up OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION to use base2 exponential bucket histograms.

Rather than relying on environment variables, I'll configure these settings programmatically using the metric HTTP options with custom selectors for both temporality preference (delta) and histogram aggregation (exponential), while still allowing environment overrides for flexibility.

Implementing the selector functions... Adding config fields...

I should also update compose.yaml to expose the OTEL environment variables so users can override compression, temporality, and histogram settings, and make sure lib.sh writes these defaults. I'll set gzip as the default compression in compose. But there's another issue: UPTRACE_DSN only gets wired during initial .env generation, so if a user exports it in their shell after the .env file exists, load_env will overwrite it with whatever's already in the file. Writing shell script logic...

Now I'm thinking through the telemetry instrumentation — traces are flowing through the HTTP server with enriched spans, and metrics come from the OpenTelemetry meter plus HTTP metrics, but runtime metrics aren't currently included. Adding runtime instrumentation would mean a new Go dependency, which conflicts with the constraint of not introducing new dependencies.

The core issue is verifying that logs, traces, spans, and metrics all reach Uptrace. There's a potential endpoint mismatch (api.uptrace.dev vs otlp.uptrace.dev) that could prevent data from arriving, the metrics interval looks reasonable at 30 seconds, and logs are fanned out through a multihandler to OTLP. The question is whether all packages are actually using the logger from main or if some are bypassing it.

For traces and spans, the HTTP server is wrapped with otelhttp, but database operations aren't instrumented since pgx tracing isn't wired up. The shutdown sequence should flush telemetry on SIGTERM, and the insecure flag derivation handles HTTP schemes correctly for self-hosted setups. There's also a potential issue with how the endpoint URL is being passed to the trace exporter—if the URL lacks a path component, the exporter should append the default signal path automatically, but I need to verify this behavior against the documentation. Parsing the endpoint URL to extract the host and path components, then using a helper function to apply the default traces path if no path was explicitly provided. I should also check if the environment variable for metric temporality preference is supported in v1.44, though since I'm setting the selectors programmatically from the config file, that's less critical. For the GitHub Actions workflow, I need to examine the existing tests and add a new job that builds both Containerfiles with podman, spins up the compose stack with the database, API, and frontend services, verifies the API is healthy through the internal network, confirms migrations ran successfully, and checks that Caddy is applying the CSP headers correctly before tearing everything down.

I'll create a dedicated script at scripts/ci-compose-test.sh to handle this so the workflow just calls it — that way developers can reproduce the exact same test locally with a single command. On Ubuntu 24.04, podman is already available but I need to install podman-compose via pip since the apt version is outdated and has known issues; the pip version gives me 1.5.x which is more reliable and closer to what users on Fedora would have.

For health verification, the api container is built from scratch so there's no curl or wget available to probe the health endpoints — I'll need to use a different approach than exec probes since the container only contains the binary itself.

Instead, I can run a temporary curl container on the same network to check the health endpoint, or reuse the frontend container which has wget available through its alpine base. The frontend approach is cleaner since it avoids pulling an extra image. I should also verify the frontend is serving properly by checking if it responds to requests on its configured port.

Now I'm deciding whether to keep the CI simple with a single solid job or test against multiple podman compose configurations. Before committing to that, I need to review the existing CI setup, test scripts, and all the relevant code files — the health checks, server logic, database setup, container configurations, and middleware — to understand what's already in place.

Show less
Done
Now let me read the remaining key files: workflows, Containerfiles, server, database, health, main, Caddyfile, and the CORS middleware.


Diagnosed database query protocol issue and mapped health endpoint verification tasks
Diagnosed database query protocol issue and mapped health endpoint verification tasks

Reading CI workflow, test script, and Containerfiles

Script
The distroless API container won't support exec probes since it lacks a shell, so I'll need to handle health checks through the frontend container with wget or set up a separate curl request on the same network. Let me trace through database.go to understand the pgx query execution, check health.go for the endpoint configuration, and review main.go, server.go, and the CORS middleware setup along with the frontend Caddyfile.


Reading database pool setup, health endpoints, and Caddyfile

Script
I'm pretty confident the issue is with pgx's default query execution mode using the extended protocol, which causes the server to fail on operator resolution when it encounters $4 + $5 with unknown parameter types. Let me verify this empirically by testing against a real PostgreSQL instance before I modify the SQL. I also need to check the exact health endpoint route paths in server.go.

I'll check the routes in server.go to understand how pgx handles the extended protocol, then test the SQL bug against a real PostgreSQL instance.


Reading server routes and checking health paths

Script
Done
Now let me empirically verify the sync 500 hypothesis with a real PostgreSQL before touching the SQL.


Verifying Uptrace endpoint configuration against current documentation
Verifying Uptrace endpoint configuration against current documentation

Installing PostgreSQL to empirically verify the sync SQL bug

Script

Installing PostgreSQL as current user (likely root)

Script

Updating apt indexes then installing PostgreSQL

Script

Starting PostgreSQL and verifying it runs

Script
PostgreSQL 16.14 is already running with the same version as the project, so I'm setting up the exact scenario by creating the users and captcha_stats tables from the migrations, then preparing a statement without explicit parameter types to match how pgx handles unknown OIDs.

I've set up PostgreSQL 16 to match production and I'm now testing the exact failing statement using pgx's extended protocol with undeclared parameter types.


Reproducing the exact production SQL with untyped parameters like pgx sends

Script
I've pinpointed the exact issue: when pgx v5 sends a Parse message with undeclared parameter types in QueryExecModeCacheStatement, PostgreSQL can't resolve the unknown + unknown operator in the VALUES clause at $4 + $5, causing the Parse to fail and returning a 500 error from SyncCaptchaStats.

The handler then blankets every error with {"error":"internal server error"}, so every sync consistently fails. The fix is to add explicit ::bigint casts to resolve the operator ambiguity.

I've confirmed the issue: PostgreSQL's Parse fails on the untyped + operator when pgx sends parameters without declared types via the extended protocol, causing every sync to return 500. Now I'll verify the fix.


Verifying the cast fix works for both insert and update paths

Script
The fix is working correctly — the insert and update paths both preserve the expected values and semantics. Now I'm scanning through the store queries to find similar issues where arithmetic or logical operators are used between untyped parameters.

The fix is working correctly with semantics preserved. Now I'll search the codebase for other instances of this same bug pattern.


Searching for other param-vs-param operator expressions in SQL

Script
I'll cast that expression for clarity even though the left-associativity handles it correctly. Now I'm looking at the CORS and environment variable issue — the logs show that podman-compose is ignoring the --no-deps flag when bringing up the frontend and cloudflared-frontend services.

The api container gets created as a dependency at that step with whatever environment was set at that moment (either stale values from .env or fresh defaults), and then when the api service is explicitly brought up later, podman-compose doesn't recreate it since there's no config-hash diffing — it just tries to create again, fails with "already in use" errors, and moves on. So the api never picks up the discovered tunnel URL for the RP_ID and CORS settings.

This explains the user's CORS failure with a stale .env pointing to a dead tunnel URL, and why deleting .env made signup work with the permissive CORS=* default. But now passkeys are actually broken on the deployed instance since the RP_ID is set to localhost while the page is served from trycloudflare.com — the WebAuthn validation will fail on origin and RP mismatch. The user hasn't noticed because they signed up with a password instead.

I need to fix this by removing any pre-existing api container before bringing it up in step 6, forcing a recreation with the corrected environment. I should also reset the tunnel-derived .env keys at the start of every up.sh run to prevent stale values from leaking in, which also handles crashes between steps. There's also an issue where exported UPTRACE_DSN gets clobbered by an older empty value in .env during the load — I need to persist caller-exported telemetry variables into .env before loading it.

For the Uptrace endpoint mapping, the current code extracts https://api.uptrace.dev from the DSN, but according to the guidance, OTLP/HTTP should target https://otlp.uptrace.dev instead. The code uses HTTP exporters, so sending them to api.uptrace.dev is wrong and likely why telemetry isn't working.

Let me examine the endpointFromDSN function to see how it parses the DSN and whether it's stripping the ?grpc=4317 query parameter correctly.

Show less
Done
Only the one spot. Now let me see the full endpointFromDSN to assess the Uptrace endpoint bug:


Reconciling sed extraction gaps in signup handler code
Reconciling sed extraction gaps in signup handler code

Reading endpointFromDSN implementation and its tests

Script
I'm noticing a discrepancy between what the DSN parser returns (https://api.uptrace.dev) and what Uptrace's documentation specifies as the OTLP/HTTP endpoint (https://otlp.uptrace.dev). I should verify against Uptrace's current documentation to confirm whether api.uptrace.dev also handles OTLP/HTTP requests or if otlp.uptrace.dev is the correct endpoint.

I found a discrepancy between the DSN endpoint and Uptrace's documentation—I'll check their current docs to verify the correct endpoint.


Uptrace OTLP endpoint otlp.uptrace.dev Go opentelemetry configure DSN
10 results


OpenTelemetry Go distro for Uptrace | Uptrace
uptrace.dev

Monitor OpenTelemetry Node.js with Uptrace | Uptrace
uptrace.dev

Ingesting telemetry using OpenTelemetry Distros for Uptrace | Uptrace
uptrace.dev

Direct OTLP Configuration for OpenTelemetry Python | Uptrace
uptrace.dev

OpenTelemetry NestJS Instrumentation | Uptrace
uptrace.dev

Getting Started with the OpenTelemetry Collector | Uptrace
uptrace.dev

Ingesting telemetry using OpenTelemetry Collector | Uptrace
uptrace.dev

OpenTelemetry Collector Configuration Tutorial | Uptrace
uptrace.dev

OpenTelemetry Java for Uptrace | Uptrace
uptrace.dev

otlp-metrics command - github.com/uptrace/uptrace-go/example/otlp-metrics - Go Packages
pkg.go.dev
I'm noticing that Uptrace's official documentation consistently specifies https://api.uptrace.dev as the OTLP/HTTP endpoint across multiple language implementations—Node.js, Python, and the collector configuration all reference this same endpoint.

Looking at the checklist against what's currently configured: the endpoint derivation is correct, headers with the DSN are set, and gzip compression is applied by default through the telemetry resolver. However, exponential histograms aren't being configured—the OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION=BASE2_EXPONENTIAL_BUCKET_HISTOGRAM environment variable isn't set anywhere in the compose setup, and the code doesn't specify an aggregation preference for metrics.

The same issue applies to delta temporality—OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=DELTA is missing. My plan is to add MetricsTemporality and MetricsHistogram fields to the TelemetryConfig struct, read them from environment variables with sensible defaults (using the spec-standard lowercase values and accepting any case), then wire them into the otlpmetrichttp exporter using the appropriate selector options from the pinned otel-go v1.44.0 version.

I'm trying to figure out if I can actually fetch and verify the exact option signatures from the Go module source on GitHub by configuring the Go proxy to bypass the sum database, since the standard proxy and sum server aren't in the allowed domains list. strumentKindCounter, InstrumentKindObservableCounter, InstrumentKindHistogram, InstrumentKindUpDownCounter, InstrumentKindObservableUpDownCounter, InstrumentKindObservableGauge, and InstrumentKindGauge. I also need to verify the aggregation options like sdkmetric.AggregationBase2ExponentialHistogram` with its configuration parameters and the default aggregation selector. Then I'll map out which instrument kinds should prefer delta temporality versus cumulative based on the OTLP spec.

I'm also planning to support three aggregation modes: delta, cumulative, and a lowmemory option that uses delta for counters and histograms but cumulative for observables. After that, I need to tackle the CORS and API recreation fixes, handle the ensure_env corrections, and get the CI working properly. I should run the full test suite with go vet and go test, validate shell scripts with bash and shellcheck, check YAML with Python, and potentially review the page-captcha.js error handling around that 500 error.

The frontend sync failure was actually a server-side issue, so once that's fixed the captcha page should report correctly — no JavaScript changes needed there. I also need to update the GitHub Actions workflow to add a new container testing job that detects the runtime and runs container builds with compose.

Writing container test script... can parse the token from signup responses using grep, then verify the frontend serves index.html with proper CSP headers. After that, clean up the containers and volumes. I'll keep the JSON parsing simple with a grep pattern, and add compose validation using Python's yaml module — though I'll keep that in a separate container test script rather than cluttering test.sh. Now I'm wondering if ubuntu-latest has podman available — checking the runner images, podman 4.9.3 is preinstalled, but podman-compose isn't, so I'll need to pip install it. The tricky part is that podman compose looks for an external provider like docker-compose or podman-compose in PATH, and GitHub runners have Docker Compose v2 as a plugin, but I need to figure out which one podman will actually find first.

To keep things simple and match the user's environment exactly, I should explicitly set COMPOSE=podman-compose in the CI script, since the detect_runtime function already respects pre-set CR and COMPOSE variables.

For the rootless setup in GitHub Actions, the runner user is non-root with rootless podman already configured, which mirrors the user's Fedora environment and naturally reproduces the health timer issue that wait_healthy handles. The container job needs to skip setup-go since the build happens inside the golang builder image instead.

Updating up.sh: I'm replacing step 6 to restart the API with correct reverse proxy and CORS settings plus its tunnel, accounting for podman-compose's behavior where it doesn't honor --no-deps, so the frontend and cloudflared-frontend containers were already created and started in step 3.

Writing shell script logic... I'm realizing that podman rm -f completely removes the container, so compose will recreate it even with the restart policy. The "already in use" warnings for the db at steps 3 and 6 are just noise from podman-compose itself—not worth suppressing. For the ensure_env function, I need to handle persisting any caller-exported telemetry overrides before loading the environment file so that the .env values don't override them. Adding more OTEL configuration variables to the template and caller-override loop, then creating a new helper to reset runtime environment variables that gets called in up.sh after ensure_env, plus updating compose.yaml to pass through the OTEL exporter settings to the api service. Writing selector functions... Writing aggregation selector... I'm realizing that GOPROXY=direct won't work for vanity import paths like go.opentelemetry.io since those domains aren't in the allowlist and require fetching the go-import metadata first. I need to find another way to verify the compilation without relying on those restricted domains.

I'll fetch the specific source files directly from GitHub raw content instead — checking the otel-go v1.44.0 exporter config files, aggregation types, and instrument kinds to verify the API signatures match what I'm expecting, particularly confirming WithEndpointURL and the histogram aggregation options. raw.githubusercontent.com supports the explicit /refs/tags/ syntax to disambiguate tag names containing slashes, so I can construct the URL as /open-telemetry/opentelemetry-go/refs/tags/exporters/otlp/otlpmetric/otlpmetrichttp/v1.44.0/ followed by the file path. The PAT should only be used for API calls that benefit from authentication, not for raw.githubusercontent.com which doesn't require it. I'll keep the token out of any committed files and use it only as a transient header in curl requests. I should also remind Kushal to rotate the PAT and check on the Uptrace DSN in virginia.txt since that file might end up in the repo.

For the GitHub Actions workflows, I need to add a container testing job to ci.yaml that installs podman-compose and runs an end-to-end test of the compose stack.

The script should set default values for the container runtime and compose command, but let the detection logic run normally so it works both locally and in CI — on the GitHub runner, podman compose version will work fine without forcing anything.

Writing shell script... I'm checking the exact log message that the API outputs — the wait_for_log function is looking for 'http server listening' but it might be timing out because either the message wording in main.go doesn't match, or the API already logged it before the grep started searching through the logs.

Now I'm setting up a way to test the frontend's health by running wget commands inside the caddy container to verify the API is reachable through the frontend network, including testing POST requests with JSON headers and body data using busybox wget's --header and --post-data flags.

I need to check the signup endpoint's expected field names and response format from the auth handlers, then chain together tests for captcha sync with a Bearer token, stats verification, notes CRUD operations, and preferences endpoints. For the frontend itself, I'll use wget with the -S flag to capture response headers and verify the CSP header is present, piping stderr to stdout so I can grep for it. All assertions will use grep -q for silent matching, and on failure I'll dump the full response for debugging.

For cleanup, I'm wrapping the teardown in a trap that calls the down.sh script with error handling. I also need to look at main.go to understand the "http server listening" log message that the wait_for_log function is checking for, verify the CORS middleware behavior to confirm it sends the right headers, and check what the default CORS configuration is in config.go.

Show less
I've confirmed the OTLP endpoint is correct, but the guidance is missing delta temporality and base2 exponential histogram settings. Now I'll check the API startup logs and auth request/response shapes to debug the CI smoke test timeout.


Checking startup log line, auth API shapes, CORS default, and down.sh

Script
The log line matches the grep pattern, so the timeout in production must stem from something else. The api container was started much earlier during tunnel polling, and its logs should contain the "http server listening" message in JSON format. When I grep for it with podman logs, it should match the "msg":"http server listening" field — unless there's something about how the output is being piped or formatted that's preventing the match. The API process might be crash-looping due to the stale environment variables, but that doesn't explain why requests worked later. The database volume is project-specific and freshly initialized, so the connection should work with the current password from the environment.

Looking at the successful g1gzbw run, it also showed the same "API listening log not observed" warning within 60 seconds, yet the app functioned fine — the user was able to sign up. So the API was definitely listening. The issue might be that cid_of is picking the wrong container when multiple containers with the same service label exist in the project, since the grep uses -a flag and takes the first match. At step 6, there was already a message about "api_1 already in use," suggesting multiple containers could be present.

I'm also considering whether the log line appears before podman logs actually starts capturing output, or if there's something about how podman-compose manages cgroups that prevents logs from being read properly — though step 10 successfully read cloudflared logs, so that seems less likely. ARTED (podman-compose creates deps without starting when they're not in the service list?) and then step 6's "up api" hit "already in use" on CREATE and... did it START the existing one? The output at step 6 (novzg): after errors, it printed gotunnels-novzg_db_1 and gotunnels-novzg_api_1 — podman-compose prints container names when it STARTS them (podman start outputs the name/id). So step 6 DID start db (no-op, running) and api. So api started at step 6... wait but then it started AFTER the env export — with the NEW env? NO — the container was CREATED at step 3 (its config/env frozen at creation); step 6 merely started it. Podman env is fixed at create time. So api runs with step-3 env. Consistent with my analysis.

But then: api started at step 6, wait_for_log at step 7 polls for 60s... api boots in ~1-3s (db ready), logs the line. Why not found? The printed names at step 6 are podman start outputs, so api definitely started at the beginning of step 6. Maybe the issue is the pattern matching inside the JSON log format — the substring 'http server listening' should be there, and grep -Eq should catch it. Let me check if wait_for_log itself has a bug.

Ah, I see it now — when running podman logs | grep -q, grep exits 0 on the first match and closes the pipe, which sends SIGPIPE to podman logs. Under set -o pipefail (inherited from up.sh), the pipeline returns the rightmost non-zero exit code. Grep exits 0 but podman logs exits 141 from SIGPIPE, so the pipeline status becomes 141, making the if condition false even though the match succeeded. That's the bug.

The race condition depends on log volume — small logs finish before grep exits (status 0), but verbose logs like API request logging trigger the SIGPIPE kill. This is exactly the pattern the team already fixed in poll_tunnel_url using || true and full draining, and it matches the documented lesson about set -euo pipefail with pipes silently killing scripts when commands like head -1 close early.

The real issue is that podman (written in Go) treats EPIPE as a fatal error and exits nonzero, so under pipefail the whole pipeline fails. The fix wraps the logs command with || true to neutralize its exit status, letting the pipeline status depend entirely on grep's result instead. The same wait_for_log issue appears in the CI script too, so I'll apply the same fix there. Let me quickly test the SIGPIPE behavior empirically with a simple bash command under pipefail to confirm my analysis. After that, I need to look at the auth request shapes — specifically what fields signupRequest needs — by checking the handlers for signup and login logic, plus the CORS and error handling in httpx.

Show less
The log pattern matches, but wait_for_log failing in both runs suggests a SIGPIPE issue where grep -q exits early and kills the pipe. I'll verify this empirically.


Reproducing the SIGPIPE+pipefail bug in wait_for_log and validating the fix

Script
So the "fixed shape" approach also failed because SIGPIPE terminates the chatty function itself — the subshell on the left side of the pipe receives SIGPIPE when echo tries to write, and while || true should catch that exit code, it's actually applying to the group's exit status rather than preventing the signal from killing the process. The group exits cleanly with 0, pipefail sees matching exit codes from both sides, and grep succeeds, but the underlying issue is that the signal is still terminating the function before the error handling can take effect.

The real solution is to avoid triggering SIGPIPE in the first place by fully draining the input — either by removing the -q flag so grep reads all input and prints matches to /dev/null, or by redirecting output explicitly so the pipeline doesn't close prematurely.

The naive fix doesn't work because SIGPIPE kills the subshell before the fallback can execute, so I'll use the project's established pattern of fully draining the stream instead.


Validating the full-drain grep fix under pipefail

Script
The drain fix resolved that SIGPIPE race condition — the verbose JSON logging on every request was triggering it consistently, which explains why the API listening log message kept appearing even though the service was working fine. Now I'm reviewing the auth handler's signup and login request/response structures, the CORS middleware setup, and the data shapes for preferences and notes to prepare the CI smoke test.

Done
Drain fix validated. Now the auth/notes/prefs API shapes for the CI smoke test:


Verifying endpoint URL path fallback behavior systematically
Verifying endpoint URL path fallback behavior systematically

Reading signup/login request-response shapes and CORS middleware

Script
I'm spotting the signup handler around line 163 where it records the user ID, username, and signup event, then returns a session response with a token. The extraction skipped some middle lines covering password validation and user creation, but that's already been reviewed. Now I need to cross-check the OpenTelemetry v1.44.0 metric API signatures to make sure they align with what's being used.

I've confirmed the signup response returns a token. Now I'll check the OTel v1.44.0 API signatures against the GitHub source since the module proxy isn't available.


Verifying otlpmetrichttp option functions exist in pinned v1.44.0

Script

Finding the module tag for otlpmetrichttp v1.44.0 via GitHub API

Script
I need to use the PAT the user provided instead of relying on the environment variable, so I'll pass it directly in the curl command as an inline variable without writing it to any deliverable files.


Locating the exact otlpmetrichttp v1.44.0 tag using the provided read-only PAT

Script

Fetching otlpmetrichttp options.go at the v1.44.0 module tag

Script

Inspecting error and listing the exporter directory at the tag

Script

Fetching config.go where the options live for otlpmetrichttp

Script

Verifying TemporalitySelector, AggregationSelector, exponential histogram struct, and instrument kinds in sdk/metric v1.44.0

Script
I'm verifying the API signatures against the pinned OpenTelemetry metrics version, checking that the temporality and aggregation selector options are properly defined with the correct function signatures.

Done
All signatures verified. Let me confirm the WithEndpointURL path behavior and metricdata temporality constants:


Verified SDK metric selectors exist, enabling telemetry implementation
Verified SDK metric selectors exist, enabling telemetry implementation

Checking WithEndpointURL path handling and temporality constants

Script
I'm checking whether WithEndpointURL properly handles empty paths by falling back to the default /v1/metrics signal path through the cleanPath function, and verifying that the telemetry implementation is actually working end-to-end as expected.


Verifying empty URL path falls back to the default /v1/metrics path

Script

Checking whether NewHTTPConfig normalizes an empty URL path to the default

Script
The path normalization is working as expected—when no explicit path is provided, it defaults to /v1/metrics, and the exporters correctly POST to the Uptrace endpoints with the DSN header. The only remaining gaps are what the user actually needs to implement.

I need to tackle four things: fix the SQL casting issue in the captcha stats sync, add a regression test via the CI smoke test that explicitly exercises the sync endpoint, extend the telemetry config to support delta temporality and exponential histograms with proper environment variable parsing and case normalization, and update the config tests accordingly.

Then I'm wiring up the temporality and aggregation selectors in the telemetry module with logging of the effective settings, fixing the shell script's SIGPIPE issue in the log-waiting function and ensuring environment variables persist before loading from .env so an empty DSN can't override what's already exported, adding a helper to reset tunnel-derived values between runs, updating the startup script to call that reset and recreate the API container so it picks up the discovered RP and CORS settings, and passing through the OTEL compression variable in the compose file.

Next I need to document the new telemetry variables in the example env file, create a new end-to-end smoke test script that spins up the full stack locally without tunnels and validates the signup flow including captcha sync and CSP headers, add a CI job to run that test with podman-compose, update the configuration docs to cover the new telemetry options, and then run the full validation suite across all changed files to check formatting and syntax. I need to verify the telemetry section in CONFIGURATION.md and the captcha frontend logic, check that httpx.DecodeJSON aligns with the auth implementation, and refine the up.sh script — particularly how wait_for_log resolves the container ID on each poll after the api recreation to ensure it catches the listening line reliably.

Now I'm thinking through whether GOTUNNELS_INSTANCE_ID should be reset — it shouldn't, since it's a stable user-facing identifier. I should also check if run.sh needs the telemetry variable passthrough for the bare go run dev path, and review the remaining GitHub workflows to see if they actually test the container and podman-compose setup like the user asked.

Show less
The endpoint derivation is correct—the real gaps are just the temporality and histogram settings. Now I'll check the remaining files and local toolchain.


Reading config tests and docs, installing validation tools

Script
I've got the tools set up and I'm ready to write out all the changes. There are five files to update: the store module needs a SQL cast fix, the config module needs new fields for metrics temporality and histogram settings along with corresponding tests, the telemetry module needs selector logic and a log line, and the shell script library needs a small adjustment.

Now I'm looking at the remaining files to create and modify: the main scripts need updates to call the reset function and recreate the API, there's a new end-to-end container test to write, the compose file needs to pass the three OpenTelemetry variables, the example environment file needs documentation for the new variables, the CI workflow needs a containers job, and the configuration docs need telemetry table rows. Before I start writing everything out, I should check the current .env.example and run.sh to stay consistent, then validate the Go code with formatting and vet checks. Since the config package only imports stdlib, I can compile and test it standalone to verify the shape of the temporality and aggregation selector logic.

For the remaining files that have external dependencies—store.go imports pgx, telemetry.go imports otel, captcha_test imports otel—I can't compile them in isolation, but I can use gofmt and careful diffs to ensure safety. The store.go fix modifies the SQL query for SyncCaptchaStats to properly calculate totals and individual counts...

Finishing the SQL query and adding a comment about the parse failure, then delivering the full store.go file via string replacement. I'll stick with pasting the full files directly in the chat since that's his established workflow — he copies from the response and commits them directly. Now I'm mapping out the specific changes needed, starting with config.go where I need to add metrics export configuration fields. Writing config initialization... Writing test cases... Adding metric configuration options... Finishing the condition check to verify the container is running and the log pattern matches, then looping with a one-second sleep until timeout. Then I'm updating the ensure_env function to add OpenTelemetry configuration variables to the template and handle caller overrides before loading the environment. I need to ensure that environment variables passed by the caller take precedence over any stale values left in the .env file, so I'm iterating through the key telemetry settings to preserve their current values before loading from the file. Writing shell script documentation... Writing shell script setup... Adding OpenTelemetry configuration...

Now I'm sketching out a bash script for the CI container test that sources utility functions, detects the runtime environment, and resolves the project name—though I notice the resolve_project function appears to lowercase the input.

Setting up a cleanup trap that captures container logs for the database, API, and frontend services if the test fails, then tears down the Docker Compose stack. The test itself starts by building the images, then bringing up the database service and waiting for it to be healthy before proceeding with the remaining steps.

Continuing with starting the API and frontend services without their dependencies, then waiting for the API to log its listening confirmation. I'm realizing that podman-compose's --no-deps flag won't prevent the database from being created again since the dependency graph flows the other way — frontend and API depend on the database, not the reverse — so cloudflared services won't be spun up either since they're not explicitly requested.

Now I'm setting up an HTTP driver that executes wget commands inside the frontend container to make requests to the API, building out helper functions that handle different HTTP methods and optional authentication tokens.

The issue is that busybox wget doesn't support PUT or DELETE methods—it only handles GET and POST through --post-data. Since the preferences endpoint uses PUT and the notes deletion uses DELETE, I need to decide between pulling in a curl container for full HTTP verb support or limiting test coverage to just GET and POST operations.

I'm going with option (a): spinning up a curl container on the compose network with proper assertions using -f for HTTP error codes and --write-out for response details. The curl image is small (~4MB) and a one-time pull. The key question is whether podman-compose's named network (<project>_default) supports DNS resolution of service names in rootless mode—it should, since podman uses internal DNS for that.

Writing the curl helper... Writing test assertions... Writing shell tests... Writing database checks... Deciding whether to validate the compose file with podman-compose config — it's supported but might introduce quirks, so I'll skip it since the build and up commands themselves serve as validation. Now I'm setting up the CI job that installs podman-compose via pipx, checks the runtime versions, and...

Writing CI steps... The user wants me to review tests and GitHub Actions, with a README update being optional. I should check if the README has a scripts section with a natural place to mention the ci-container-test.sh script—if so, I'll add it; otherwise I'll skip it to keep scope manageable. Then I need to validate the preferences and notes shapes, CaptchaStats JSON tags, OriginAllowed configuration, .env.example, and potentially the README scripts mention.

Show less
I'm checking the JSON schemas for the CI smoke test, the environment example file, and the README scripts section.


Reading JSON struct tags for captcha, prefs, and notes

Script
I'm picking up where I left off with 11 files to write. The repo is already extracted at /home/claude/gotunnels/repo/, and all root causes have been verified, so I can jump straight into the work without waiting for confirmation. Let me start by checking the current state of the working directory and reviewing the key files I need to modify.


Verify extracted repo state after compaction

Script
The repo structure looks good — scripts directory contains the shell utilities and .github/workflows has the CI configuration files. Now I'm pulling up store.go to see what needs to be modified.


View the SyncCaptchaStats function to apply the bigint cast fix

Script
I'm applying the bigint casts to store.go and adding a comment to explain the fix.


Fix the captcha sync 500: explicit bigint casts so Postgres can type $4 + $5 under the extended protocol

store.go
+32
-22

Locate telemetry config sections to modify

Script

Read full telemetry config section

Script
I'm updating config.go to add MetricsTemporality and MetricsHistogram fields to TelemetryConfig, then parsing them in resolveTelemetry with lowercase normalization alongside the Compression field.


Add metrics temporality + histogram aggregation fields to TelemetryConfig

config.go
+19
-9

Parse the two new OTEL metrics env vars (lowercase-normalized) and lowercase compression

config.go
+36
-23

Read existing config tests to extend them

Script
Adding new tests to config_test.go...


Add tests for the new metrics temporality/histogram defaults and overrides

config_test.go
+57
-14

Read telemetry.go in full before wiring temporality/aggregation selectors

Script
I need to update telemetry.go by adding the necessary imports for strings and OpenTelemetry metric data, then implement selector functions for temporality and aggregation behavior. These will be wired into the metric HTTP options and logged when telemetry is enabled. Writing aggregation selector... I need to check the go.mod file to confirm which version of the SDK is pinned and verify that DefaultAggregationSelector is available in it.


Confirm pinned otel versions from go.mod

Script

Verify DefaultAggregationSelector and DefaultTemporalitySelector exist in pinned sdk/metric v1.44.0

Script
I've confirmed that both DefaultTemporalitySelector and DefaultAggregationSelector are available in the current version of sdk/metric, so I can proceed with updating the telemetry.go file.

Done
Both default selectors confirmed in the pinned v1.44.0 source. Now wiring telemetry.go:


Refactoring up.sh script and resolving podman-compose dependency issues
Refactoring up.sh script and resolving podman-compose dependency issues

Add strings + metricdata imports to telemetry.go

telemetry.go
+24
-22

Log the effective metrics settings in the telemetry-enabled line

telemetry.go
+7
-4

Wire temporality + aggregation selectors into metric exporter options and add the selector functions

telemetry.go
+77
-13
Looking at the changes needed for scripts/lib.sh: fixing the wait_for_log drain issue, updating ensure_env to handle caller overrides for six telemetry variables before loading the environment plus three new OTEL variables in the template, and adding a new reset_runtime_env helper function. Let me examine the current lib.sh to understand what needs to be modified.


Read full lib.sh before editing

Script
I'm working through the lib.sh changes, starting with the wait_for_log fix. The current implementation uses grep -q which exits immediately on the first match, causing a signal issue. I need to switch to a full-drain pattern that pipes the logs through grep and redirects to /dev/null, then neutralizes the exit code with || true to prevent the script from failing.

Writing the corrected function...

The real issue was that grep with -q exits immediately on the first match, leaving the logs process hanging with SIGPIPE, which under pipefail makes the whole condition fail even when the pattern exists. By redirecting to /dev/null instead, grep reads the entire stream to completion, so the logs process exits cleanly and the condition works correctly. Now I'm moving on to the next issue with ensure_env caller-override.

The problem is when .env already has an empty UPTRACE_DSN= and the user exports a DSN in their shell, load_env clobbers it. The fix is to preserve any caller-exported telemetry variables before loading .env by saving those six specific variables into the file first.

I'll also add three new OTEL configuration variables to the generated template with sensible defaults: compression set to gzip, metrics temporality preference to delta, and histogram aggregation to base2 exponential bucket histogram.

Now I'm working through the ensure_env function logic, which checks if the environment file exists and generates it if needed, otherwise reuses the existing one. The key part is preserving any telemetry settings that were exported in the caller's shell before load_env runs, since an existing .env file with empty values would otherwise silently override those exports and disable telemetry without any warning.

Writing shell variable handling... Writing reset function...

Show less

Add 3 new OTEL vars to the .env template and the caller-override persistence loop in ensure_env

lib.sh
+56
-31

Add reset_runtime_env helper after load_env

lib.sh
+32
-9
I caught a syntax error in the bash script where I accidentally used // for a comment instead of #, which would break the whole thing — need to fix that.

Done
I introduced a // comment typo in that bash block — fixing:


Investigating sync endpoint failures and Uptrace integration
Investigating sync endpoint failures and Uptrace integration

Fix // comment typo that would be a bash syntax error

lib.sh
+2
-2

Apply the wait_for_log SIGPIPE fix with full-drain grep

lib.sh
+22
-12

Read up.sh in full before editing

Script
I need to modify up.sh to call reset_runtime_env after ensure_env, and then fix step 6 by removing the pre-created api container before bringing it up with --no-deps, since podman-compose ignores that flag when frontend depends on api and ends up starting the api container with frozen environment variables. The api service is the only one that needs removal since it's a dependency of frontend, which gets processed at step 3. cloudflared-api isn't a dependency of anything requested, so it won't be created then. Let me verify the compose.yaml structure and check for the env passthrough additions.


Read compose.yaml for dependency graph and env passthrough edits

Script
So the frontend service pulls in the api as a dependency when created at step 3, while cloudflared-api stays dormant since it's not explicitly requested. I need to update up.sh to reset the runtime environment after ensuring it's set up, and add a pre-removal step for any stale api containers before step 6. Then in compose.yaml, I'll add the three OTEL passthrough variables following the existing OTEL_E configuration.


Call reset_runtime_env at every up.sh start

up.sh
+7
-2

Step 6: remove any dependency-created api container so it is recreated with the tunnel-derived env

up.sh
+20
-3

Add the three OTEL metrics/compression passthrough vars to the api service env

compose.yaml
+8
-5
I need to examine the existing CI configuration files—the ci.yaml workflow, down.sh script, and Containerfiles—along with the .env.example and configuration documentation to understand the current setup before writing the container test script.


Read ci.yaml, down.sh, and check .env.example existence

Script

Read rest of .env.example and the telemetry table in docs/CONFIGURATION.md

Script

Document the two new OTEL metrics vars in .env.example

.env.example
+24
-11

Add the temporality + histogram rows to the telemetry table in CONFIGURATION.md

CONFIGURATION.md
+9
-7
I'm working through the CI container test script now—setting up the runtime detection and environment initialization, then configuring Docker Compose to build and start the database with health checks before bringing up the API and frontend services without their dependencies since the database is already running. So in CI, I don't need to recreate the api container since the environment from the bootstrap defaults is fine for the whole run. I'm skipping the cloudflared containers entirely since they'd require outbound network access to Cloudflare and would be flaky anyway. Instead, I'll wait for the api to log that the HTTP server is listening, then run smoke tests using a curl container on the compose-created network.

The network name is ${PROJECT}_default for both podman-compose and docker compose, and the api service gets a network alias matching its service name, so DNS resolution works. I'm creating a helper function that wraps the container runtime to run curl commands against that network, and I'll use sed to extract tokens from responses. Then I'll set up assertions to check the health endpoints and verify migrations ran correctly. I need to verify the password requirements from the auth handlers and confirm the exact field names for the signup response and captcha sync request structure.

For the captcha sync test, I'm planning two requests: the first with best_streak 4, current_streak 2, manual 3, auto 5 to reach total_solves 8, then a second sync that adds manual 4 and auto 8 to push the total to 20 while current_streak drops to 1 and best_streak stays at 4. I need to check the exact JSON field names in the captcha handler to construct the request bodies correctly.

Then I'll verify the leaderboard includes the username field, and test the preferences endpoint with a PUT request to set the theme to dark and confirm the response.

For the notes API, I'm checking POST to create a note with a body, GET to retrieve the list and verify it's there, and DELETE to remove it by ID—need to confirm what status code the delete handler returns. I'm also testing the OPTIONS preflight request expecting a 204 response, and finally checking that the frontend serves HTML with the content-security-policy header present.

I need to look at the captcha, notes, preferences, and auth handler files to understand the exact JSON structure for requests and responses.

Show less
Done
Now the CI test script. Let me verify exact request/response JSON shapes from the handlers first:


Catalogued API response formats and identified outstanding verification tasks
Catalogued API response formats and identified outstanding verification tasks

Verify JSON field names and status codes for captcha/notes/auth handlers

Script
I'm reviewing the API response shapes for the sync endpoint, which returns stats like manual and auto deltas along with streak information, and confirming the note operations—POST creates with a 201 status, DELETE removes with a 200 response, and GET retrieves the full list. The signup validation checks out with the password requirement and token in the response, but I need to verify what status code issueSession returns and whether the sync stats are wrapped in an object or returned directly.


Check session issue status code and captcha/prefs response shapes

Script
Done
please continue


Scrutinizing script behavior under set -e conditions
Scrutinizing script behavior under set -e conditions
I need to track down the issueSession return value since the grep only showed where it's called, not defined. I should also verify the CaptchaStats JSON tags in store.go and confirm the migrations count is at least 7, then move on to writing the CI container test script and updating ci.yaml before doing a final validation pass.


Find issueSession status code, CaptchaStats JSON tags, and migration count

Script
I'm confirming the API response formats across all endpoints: signup returns a token at the top level, the sync endpoint provides stats with compact JSON formatting from Go's encoder, the leaderboard includes both the list and the current user's data with a username field, and the preferences endpoint responds to PUT requests with a saved status confirmation.

For the remaining endpoints, GET requests return the preferences with the key-value pair and an exists flag, notes endpoints handle POST with a 201 response containing the note object, GET to retrieve all notes as an array, and DELETE to confirm deletion. The database has 7 migration files so the max version should be 7, and password validation requires at least 8 characters.

I'm verifying that Go's JSON encoder produces compact output without spaces after colons, so grepping for the exact format like "total_solves":8 will work correctly. The Authorization header follows the Bearer token format, and I need to extract note IDs from the POST response using a sed pattern to capture the UUID from the note object.

Refining the sed pattern to avoid matching the user_id field — I need to anchor it to the note object structure itself, like matching {"id":" at the start of the note object to ensure I'm grabbing the right ID field rather than accidentally capturing from user_id.

Show less

Verify Note struct field order, Bearer auth parsing, and WriteJSON encoding

Script
I need to remember that note IDs are integers, not UUIDs, so the JSON structure has a numeric id field. I can extract the ID using a sed pattern that captures the numeric value between the id field markers, and then use that numeric ID in the DELETE request to the notes endpoint. Now I'm looking for where the Bearer token parsing happens in the auth middleware—it's not showing up in session.go, so I need to search more broadly through the codebase.


Locate Bearer token parsing in auth middleware

Script
I'm setting up the CI container test script with the standard shell options and sourcing the necessary utilities, then configuring the project name from an argument with a default fallback, and defining the curl image and network naming conventions.

The cleanup trap will dump logs from all services on failure and always tear down the compose stack, then I'll build the images, start the database with a health check using pg_isready, bring up the API and frontend, and wait for the API to log that it's listening — failing the test if that doesn't happen within 90 seconds since this is CI and we need strict validation.

I'm creating helper functions to run curl commands inside a throwaway container on the compose network, plus assertion helpers to check response bodies and HTTP status codes, with careful attention to quoting under the IFS setting.

One thing to verify: the curl image needs to be pulled on first use, which works fine with podman pulling docker.io images. For DNS resolution, podman-compose sets up the network with DNS enabled and adds network aliases matching service names, so containers can reach each other by service name or container name — though docker compose uses dashes in container names while podman uses dashes too, so I should double-check which naming convention to rely on.

GitHub Actions runners have podman 4.9.x with netavark and aardvark-dns, so DNS on user-defined networks is supported. After bringing up the api and frontend services with up -d --no-deps, the api will wait for the database to be healthy, then run smoke tests. For the migrations check, I'm querying the database container to verify the migration status. Writing signup and token extraction...

Now I'm testing the stats endpoint with two different captcha solve scenarios to verify the totals accumulate correctly—first call returns 8 total solves with a best streak of 4, then the second call should show 20 total solves while the best streak stays at 4. This is the key regression test that was failing with 500 errors before the fix. Next I'm checking the leaderboard endpoint. Writing shell commands... Instead of using grep's quiet flag, I should redirect to /dev/null to ensure grep reads the entire input and avoids SIGPIPE issues. For the health checks, I'm adding curl calls to /healthz and /readyz endpoints with the -f flag to fail on HTTP errors, wrapped in a step logger. The -sS flag is already in the curl wrapper, so I just need to add -f where needed. The initial curl call also pulls the image, and the 20-second timeout only applies after the container starts. I'm also noting that podman-compose ignores the --no-deps flag and will still process the db service even when trying to start just api and frontend. With --no-deps in local dev, the db stays untouched and api+frontend start up; the depends_on health check gets skipped but that's fine since db is already healthy from the earlier wait. The cleanup trap is set up to catch failures and dump the last 100 lines of logs from each service (db, api, frontend) when the test exits non-zero. I need to be careful with the EXIT trap behavior — capturing the original exit status upfront and using exit "$status" at the end ensures the final code is explicit and unambiguous, especially under set -e where failing commands could interfere. I'll guard all trap commands with || true to prevent early exits. I also need to check that CR is set before calling cid_of since the trap might fire before ensure_env completes, and I should verify that the environment setup properly configures POSTGRES_PASSWORD and other secrets for the compose stack.

The GitHub Actions workflow installs podman-compose via pipx, which is already available on the runner, then validates the runtime versions before running the container smoke test script. Since podman and podman-compose come preinstalled on ubuntu-latest, the setup is straightforward.

I'm checking the migration table name next before writing the file.

Show less

Confirm migrations table name and health route paths

Script
Now I'm writing the CI container test script.

Done
All facts confirmed. Writing the CI container test script:


Validated script error handling and curl behavior mechanics
Validated script error handling and curl behavior mechanics

New CI script: builds both Containerfiles, brings up db+api+frontend via compose, and runs an end-to-end HTTP smoke test including the captcha-sync regression
ci-container-test.sh
I'm realizing there's a subtle issue with how the health check command behaves under set -e — when the curl fails, the && operator causes the entire compound statement to fail, which triggers the error trap and dumps logs as intended. That's actually the desired behavior for a hard fail, but I need to think through what happens when a command...

Actually, looking at the rest of the logic: the wait_healthy function passes arguments through correctly, the sed pattern for token extraction handles the single occurrence fine, and the note ID extraction regex is solid. The FAILED counter works because the assertion functions run in the current shell and their guards prevent nonzero returns, keeping set -e safe. The main concern is when ccurl fails during variable assignment — that triggers set -e and the trap, which is acceptable since transport failures are fatal anyway, and curl without -f still captures HTTP error bodies.

Now I need to update the CI configuration to add the containers job.

Show less
Done
Now the updated GitHub Actions workflow:


Navigated Go dependency resolution and verification constraints
Navigated Go dependency resolution and verification constraints

Add the containers job that builds the images and runs the compose smoke test with podman

ci.yaml
+38
-14
Running a full validation pass across gofmt, go test, bash syntax, shellcheck, YAML parsing, and secrets scanning.


gofmt validation on all changed Go files

Script

Run the config package tests for real in a stdlib-only scratch module

Script
The network is blocking access to proxy.golang.org, so I need to check what Go version is currently installed locally to work around this constraint.


Pin scratch module to the installed Go version and run the config tests

Script
The grep pattern with -P flag isn't extracting the version correctly from the go version output — it's matching "1.22" from "go1.22.2" but the -P option might not be supported on this system, so the extraction is failing silently.

Actually, I think I see the real issue: when I cd into the scratch/cfg directory and run go version there, it's triggering the toolchain resolution from the go.mod file which specifies go 1.26, causing it to try downloading that version and fail. That's why grep gets empty output and the sed substitution produces just "go " with nothing after it. The simpler fix is to just write the go.mod file directly with the go 1.22 version instead of trying to extract it dynamically.


Fix go.mod and run config tests with local toolchain

Script
The config tests all pass, so now I need to verify that telemetry.go compiles with the OpenTelemetry dependencies. The network restrictions block the standard Go proxy, but I can work around this by setting GOPROXY=direct, which will fetch modules directly from GitHub via git over HTTPS — a protocol that's allowed. Since the OpenTelemetry modules are hosted on github.com, this approach should let me download the dependencies I need.

However, there's a toolchain version mismatch: the go.mod specifies go 1.26, but I only have go 1.22.2 available locally. The OpenTelemetry v1.44.0 likely requires go 1.24 or higher, so even with the dependencies downloaded, the local compiler won't be able to build the code. I could rewrite the go directive temporarily, but that risks pulling in dependencies that require a newer Go version than what I have.

Since full compilation isn't feasible, I'll verify the code through syntax checking with gofmt and by manually validating that the OpenTelemetry API calls match the actual function signatures from the v1.44.0 source — I've already confirmed that WithTemporalitySelector and WithAggregationSelector exist with the correct types.

Show less
Done
Config tests all pass. The telemetry package needs otel deps I can't fetch here, so I verify the exact API usage against the pinned v1.44.0 source directly:


Verifying no new errors introduced by changes
Verifying no new errors introduced by changes

Final verification of every otel symbol used in telemetry.go against pinned v1.44.0

Script
Now I'm running validation checks across the shell scripts and configuration files—checking syntax with bash and shellcheck, parsing the YAML files, and scanning for any exposed secrets.

Done
Every otel symbol confirmed. Shell + YAML validation:


Verified test results and prepared multiple files for delivery
Verified test results and prepared multiple files for delivery

bash -n and shellcheck on all changed/new scripts

Script
The shellcheck findings are all pre-existing issues from before my changes—info-level warnings about shell directives and an unused loop counter pattern that's consistent throughout the codebase. No new errors were introduced, so I should verify the exit code with error-level severity filtering to confirm there are truly zero errors to address.


Confirm zero shellcheck errors and validate YAML files

Script

Secrets scan on deliverables, then empirically test the ensure_env DSN-persist + reset_runtime_env behavior

Script
The "FOUND SECRETS" alert is a false positive—it just matched the placeholder format in the .env.example file, not an actual secret. The sandbox test confirms the fix is working: the shell-exported UPTRACE_DSN persisted correctly into .env without being overwritten.

Now I need to verify the wait_for_log fix handles the SIGPIPE scenario properly under pipefail, and then test the sed extraction patterns against realistic JSON data. I should also validate that the assert helper functions in ci-container-test.sh work correctly by testing the critical extraction and assertion logic directly.

The .env.example file only contains placeholder values with no real secrets, and I've confirmed both environment fixes work in the sandbox. Now I'll verify the wait_for_log fix and JSON extraction patterns.


Empirically verify the wait_for_log fix, sed extractions, and assert helper semantics under strict mode

Script

Rerun harness under bash explicitly (default shell was sh)

Script
I've verified that the new drain pattern correctly catches the present pattern issue that the old -Eq approach missed under pipefail, and confirmed the token and note-id extractions work properly against the Go-encoder JSON output. The assert_contains function handles set -e correctly with FAILED tracking intact. Now I need to validate the SQL fix by running it against the local Postgres instance to ensure the final query in store.go has all the necessary casts and ON CONFLICT handling.

Running the exact SQL from store.go against real Postgres to verify the fix works end-to-end through pgx's untyped-parameter path.


Final SQL regression check: the fixed statement plans and executes correctly on both insert and update paths

Script

Run the SQL regression as the postgres user

Script
The store.go implementation is working correctly through the untyped-Parse path, with insert operations returning the expected 8/3/5 metrics and best staying at 4, while update operations show 20/7/13 with best at 4 and current dropping to 1—exactly matching what the CI test expects. Now I need to do a final visual review of all the modified files to ensure nothing's broken, then provide the complete files with their exact paths.

Done
The exact statement from store.go now plans and executes correctly through the untyped-Parse path — insert 8/3/5, update 20/7/13 with GREATEST keeping best=4. Let me pull the final file contents for delivery:


Validated all deliverables and prepared comprehensive file submission
Validated all deliverables and prepared comprehensive file submission

Line counts of all 11 deliverables

Script

Read final store.go for delivery

Script

Read final config.go for delivery

Script

Read final telemetry.go and lib.sh for delivery

Script
I've gathered all the files and now I'm ready to compile the complete response with all 11 full files, including the final version of up.sh and the remaining smaller files.


Read final up.sh, ci.yaml, compose.yaml, .env.example for delivery

Script
I need to finalize the remaining test files and documentation — I have the config_test.go changes ready to insert after the TestResolveTelemetry_UptraceDSN test, and I should verify the full content of docs/CONFIGURATION.md and ci-container-test.sh before pasting them out.


Read final CONFIGURATION.md for delivery

Script
I've validated all the core files—store.go passes gofmt and SQL verification against PostgreSQL 16, config.go and its tests compile cleanly with all 13 tests passing, telemetry.go has every OpenTelemetry symbol checked against the pinned version, and lib.sh passes bash syntax and shellcheck validation with the helper functions tested in a sandbox environment.

Now I'm finishing the validation sweep: up.sh and ci-container-test.sh both pass syntax and linting checks with extraction patterns and assertions verified, the YAML files parse correctly, documentation is updated, and I've confirmed no credentials are exposed in any deliverable.

I'm ready to compile the full response with all 11 files, their paths, the root-cause findings with empirical verification, and deployment notes including the passkey issue on the current instance and the credential rotation steps needed for the PAT and Uptrace DSN.

Show less
Done
All 11 files are finalized and validated. Here's what was actually wrong, each root cause verified empirically, followed by every full file.

Root causes
1. Captcha sync always 500. SyncCaptchaStats had VALUES (..., $4 + $5, ...). pgx v5's default statement-cache mode sends the extended-protocol Parse with no declared parameter types, so Postgres must infer them — and $4 + $5 fails at plan time with operator is not unique: unknown + unknown. The statement could never even execute, hence a 500 on every single call. I reproduced the exact error on Postgres 16 via PREPARE without a type list, and verified the fix (::bigint casts on every integer placeholder, in both the VALUES and the ON CONFLICT clauses): insert path returns 8/3/5, update path returns 20/7/13 with GREATEST keeping best=4 and current dropping to 1.

2. CORS failure on signup (and silently broken passkeys). Your virginia.txt log shows podman-compose ignores --no-deps: step 3 (up -d --no-deps frontend cloudflared-frontend) also created and started the api container (frontend depends_on api) with env frozen at that moment. Step 6's up then hit "name already in use" and never recreated it — so the api never received the tunnel-derived RP_ID/RP_ORIGINS/CORS_ALLOWED_ORIGINS. Deleting .env "fixed" CORS only because a fresh .env defaults to *. Side effect: passkeys on your current deployment are broken (RP ID is stuck at localhost while the origin is trycloudflare.com) until you redeploy. Fix is two-part: reset_runtime_env returns the three tunnel-derived keys to bootstrap defaults at every run start (so deleting .env is never needed and your secrets/DSN survive), and step 6 now removes any dependency-created api container before starting it so compose recreates it with the correct env (a no-op under docker compose).

3. "did not observe API listening log yet" on every run. wait_for_log used grep -Eq: grep exits at first match, podman logs takes SIGPIPE (exit 141), and under pipefail the pipeline reports failure even though the pattern matched. Every success was inverted into a miss. Verified both directions in a harness: the old form misses a present pattern; the new full-drain form (grep -E ... >/dev/null) matches correctly. Same bug family you already fixed in cid_of and poll_tunnel_url.

4. Uptrace telemetry gaps. Good news first: your endpoint derivation was already correct — https://api.uptrace.dev is a valid OTLP/HTTP base and otel-go v1.44.0's WithEndpointURL appends /v1/{traces,metrics,logs} itself; gzip was already on. What was missing per Uptrace's guidance: delta temporality and base2 exponential histograms. Both are now wired via WithTemporalitySelector/WithAggregationSelector (every symbol verified against the pinned v1.44.0 source). Plus one nasty footgun I found and fixed: an existing .env with an empty UPTRACE_DSN= line silently clobbered a shell-exported DSN (load_env re-exports the empty value), turning telemetry off with zero errors — ensure_env now persists caller-exported telemetry vars into .env before loading it. Verified in a sandbox: exported DSN survives, stale tunnel values reset, generated secrets preserved.

5. CI. New scripts/ci-container-test.sh builds both Containerfiles, brings up db+api+frontend via podman-compose (no tunnels), and runs an HTTP smoke test from a curl container on the compose network — including the captcha-sync regression (which only reproduces against real Postgres), prefs, notes, leaderboard, CORS preflight, and the frontend CSP header. Wired into ci.yaml as a second job.

Validation: gofmt -e clean on all 4 Go files; all 13 config tests pass (ran for real in a stdlib-only scratch module); bash -n + shellcheck error-free on all scripts; YAML parses; no PAT or DSN in any file.

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
//
// Every integer placeholder carries an explicit ::bigint cast. pgx v5 uses the
// extended query protocol and, in its default statement-cache mode, sends the
// Parse step without declared parameter types. Postgres then has to infer them
// and the expression `$4 + $5` fails at plan time with
// `ERROR: operator is not unique: unknown + unknown` — the statement can never
// even execute, so the sync endpoint 500s on every request. Casting each
// parameter pins the types and makes the statement plannable. (Placeholders
// compared against a typed column, like `$1::uuid = user_id`, would be
// inferable anyway, but we cast uniformly for clarity.)
func (s *Store) SyncCaptchaStats(ctx context.Context, userID string, in CaptchaSyncInput) (CaptchaStats, error) {
	st := CaptchaStats{UserID: userID}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO captcha_stats
			(user_id, best_streak, current_streak, total_solves, manual_solves, auto_solves, updated_at)
		VALUES ($1::uuid, $2::bigint, $3::bigint, $4::bigint + $5::bigint, $4::bigint, $5::bigint, now())
		ON CONFLICT (user_id) DO UPDATE SET
			best_streak    = GREATEST(captcha_stats.best_streak, EXCLUDED.best_streak),
			current_streak = EXCLUDED.current_streak,
			total_solves   = captcha_stats.total_solves + $4::bigint + $5::bigint,
			manual_solves  = captcha_stats.manual_solves + $4::bigint,
			auto_solves    = captcha_stats.auto_solves + $5::bigint,
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
internal/config/config.go
go
// Package config is the single source of truth for every runtime setting.
//
// Everything the application can be tuned with is read here, from environment
// variables, and nowhere else. No other package reads os.Getenv. This keeps
// the "all configuration is central in one location" promise from the design
// notes: to see everything GoTunnels can be configured with, read this file.
package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all resolved runtime settings for the API service.
type Config struct {
	// Identity -----------------------------------------------------------
	InstanceID  string // unique per running instance; also an OTel resource attribute
	ServiceName string // OTel service.name
	Version     string // build version (set via -ldflags at build time)

	// HTTP ---------------------------------------------------------------
	HTTPAddr        string        // listen address, e.g. ":8080"
	ShutdownTimeout time.Duration // graceful shutdown budget

	// Database -----------------------------------------------------------
	DatabaseURL      string
	DBMaxConns       int32
	DBMinConns       int32
	DBConnectTimeout time.Duration

	// Sessions -----------------------------------------------------------
	SessionCookieName string
	SessionTTL        time.Duration

	// Secrets (generated per-instance by scripts/up.sh, never committed) -
	ipHashPepper []byte   // used as sha256(pepper || ip)
	totpAESKey   [32]byte // AES-256-GCM key for encrypting TOTP secrets at rest

	// CORS ---------------------------------------------------------------
	// Exact allowed origins. Because the browser talks to the API cross-origin
	// (frontend tunnel URL != API tunnel URL) and sends credentials, we cannot
	// use the "*" wildcard together with credentials; we echo an allowed origin.
	CORSAllowedOrigins []string

	// WebAuthn / passkeys -----------------------------------------------
	RPID          string   // relying-party ID: the frontend's registrable domain
	RPDisplayName string   // human-readable name shown by the authenticator
	RPOrigins     []string // full origins the browser will present, e.g. https://x.trycloudflare.com

	// Content Security Policy -------------------------------------------
	// The header itself is emitted by Caddy on the frontend. These values let
	// the API report what mode it believes is active and are surfaced on the
	// health/info endpoint so the whole system agrees on one central policy.
	CSPMode   string // "report-only" or "enforce"
	CSPPolicy string // the policy string (informational mirror of Caddy's)

	// Telemetry ----------------------------------------------------------
	Telemetry TelemetryConfig

	// Dev conveniences ---------------------------------------------------
	Dev bool // when true, missing secrets are generated ephemerally with a warning
}

// TelemetryConfig captures OTLP/HTTP exporter settings. When Enabled is false
// the service logs to stdout only and installs no-op trace/metric providers.
type TelemetryConfig struct {
	Enabled     bool
	EndpointURL string            // full base URL, e.g. https://api.uptrace.dev
	Headers     map[string]string // e.g. {"uptrace-dsn": "..."}
	Insecure    bool              // allow http:// (in-cluster collectors)
	Compression string            // "gzip" or ""

	// MetricsTemporality mirrors OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE
	// ("delta", "cumulative", or "lowmemory"; normalized to lowercase). Uptrace
	// prefers delta temporality, so that is the default.
	MetricsTemporality string
	// MetricsHistogram mirrors OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION
	// ("base2_exponential_bucket_histogram" or "explicit_bucket_histogram";
	// normalized to lowercase). Exponential histograms compress better and give
	// Uptrace more accurate percentiles, so they are the default.
	MetricsHistogram string
}

// Load reads and validates configuration from the environment.
func Load() (*Config, error) {
	c := &Config{
		InstanceID:         getenv("GOTUNNELS_INSTANCE_ID", defaultInstanceID()),
		ServiceName:        getenv("OTEL_SERVICE_NAME", getenv("GOTUNNELS_SERVICE_NAME", "gotunnels-api")),
		Version:            getenv("GOTUNNELS_VERSION", "dev"),
		HTTPAddr:           getenv("GOTUNNELS_HTTP_ADDR", ":8080"),
		ShutdownTimeout:    getdur("GOTUNNELS_SHUTDOWN_TIMEOUT", 15*time.Second),
		DatabaseURL:        getenv("DATABASE_URL", ""),
		DBMaxConns:         int32(getint("GOTUNNELS_DB_MAX_CONNS", 20)),
		DBMinConns:         int32(getint("GOTUNNELS_DB_MIN_CONNS", 2)),
		DBConnectTimeout:   getdur("GOTUNNELS_DB_CONNECT_TIMEOUT", 30*time.Second),
		SessionCookieName:  getenv("GOTUNNELS_SESSION_COOKIE_NAME", "gotunnels_session"),
		SessionTTL:         getdur("GOTUNNELS_SESSION_TTL", 24*time.Hour),
		CORSAllowedOrigins: splitList(getenv("GOTUNNELS_CORS_ALLOWED_ORIGINS", "*")),
		RPID:               getenv("GOTUNNELS_RP_ID", "localhost"),
		RPDisplayName:      getenv("GOTUNNELS_RP_DISPLAY_NAME", "GoTunnels"),
		RPOrigins:          splitList(getenv("GOTUNNELS_RP_ORIGINS", "http://localhost:8080")),
		CSPMode:            getenv("GOTUNNELS_CSP_MODE", "report-only"),
		CSPPolicy:          getenv("GOTUNNELS_CSP_POLICY", DefaultCSPPolicy),
		Dev:                getbool("GOTUNNELS_DEV", false),
	}

	if err := c.resolveSecrets(); err != nil {
		return nil, err
	}
	c.resolveTelemetry()

	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// DefaultCSPPolicy locks everything to same-origin: no third-party scripts,
// styles, images, fonts, media, or frames. Everything is self-hosted. It is
// emitted in report-only mode by default (see the Caddyfile), so violations
// are reported but nothing is blocked yet.
const DefaultCSPPolicy = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self'; " +
	"img-src 'self'; " +
	"font-src 'self'; " +
	"connect-src 'self' https:; " +
	"media-src 'self'; " +
	"object-src 'none'; " +
	"frame-src 'none'; " +
	"frame-ancestors 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'"

func (c *Config) resolveSecrets() error {
	// IP hashing pepper.
	if pepper := os.Getenv("GOTUNNELS_IP_HASH_PEPPER"); pepper != "" {
		c.ipHashPepper = []byte(pepper)
	} else if c.Dev {
		c.ipHashPepper = mustRandom(32)
	} else {
		return fmt.Errorf("config: GOTUNNELS_IP_HASH_PEPPER is required (set GOTUNNELS_DEV=1 to auto-generate for local dev)")
	}

	// TOTP encryption key. Accept any string; derive a fixed 32-byte AES key
	// via SHA-256 so operators can supply `openssl rand -base64 32` output
	// without worrying about exact byte length.
	if raw := os.Getenv("GOTUNNELS_TOTP_ENCRYPTION_KEY"); raw != "" {
		c.totpAESKey = sha256.Sum256([]byte(raw))
	} else if c.Dev {
		c.totpAESKey = sha256.Sum256(mustRandom(32))
	} else {
		return fmt.Errorf("config: GOTUNNELS_TOTP_ENCRYPTION_KEY is required (set GOTUNNELS_DEV=1 to auto-generate for local dev)")
	}
	return nil
}

func (c *Config) resolveTelemetry() {
	// Preference order:
	//   1. UPTRACE_DSN (convenience) -> derive endpoint + uptrace-dsn header
	//   2. OTEL_EXPORTER_OTLP_ENDPOINT (+ OTEL_EXPORTER_OTLP_HEADERS)
	//   3. disabled (stdout logging, no-op traces/metrics)
	//
	// The two metrics knobs follow the OpenTelemetry spec's environment
	// variables. Their spec values are UPPERCASE (e.g. DELTA,
	// BASE2_EXPONENTIAL_BUCKET_HISTOGRAM) but we normalize to lowercase so
	// comparisons elsewhere are simple and either casing works.
	tc := TelemetryConfig{
		Headers:     map[string]string{},
		Compression: strings.ToLower(getenv("OTEL_EXPORTER_OTLP_COMPRESSION", "gzip")),
		MetricsTemporality: strings.ToLower(getenv(
			"OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "delta")),
		MetricsHistogram: strings.ToLower(getenv(
			"OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION",
			"base2_exponential_bucket_histogram")),
	}

	if dsn := os.Getenv("UPTRACE_DSN"); dsn != "" {
		endpoint, insecure := endpointFromDSN(dsn)
		tc.Enabled = true
		tc.EndpointURL = endpoint
		tc.Insecure = insecure
		tc.Headers["uptrace-dsn"] = dsn
	} else if ep := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); ep != "" {
		tc.Enabled = true
		tc.EndpointURL = ep
		tc.Insecure = strings.HasPrefix(ep, "http://")
		for k, v := range parseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")) {
			tc.Headers[k] = v
		}
	}
	c.Telemetry = tc
}

// Validate checks invariants that must hold before the server starts.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: DATABASE_URL is required")
	}
	if c.HTTPAddr == "" {
		return fmt.Errorf("config: GOTUNNELS_HTTP_ADDR is required")
	}
	if c.SessionTTL <= 0 {
		return fmt.Errorf("config: GOTUNNELS_SESSION_TTL must be positive")
	}
	switch c.CSPMode {
	case "report-only", "enforce":
	default:
		return fmt.Errorf("config: GOTUNNELS_CSP_MODE must be 'report-only' or 'enforce', got %q", c.CSPMode)
	}
	if len(c.RPOrigins) == 0 {
		return fmt.Errorf("config: GOTUNNELS_RP_ORIGINS must contain at least one origin")
	}
	return nil
}

// IPHashPepper returns a copy-safe reference to the pepper bytes.
func (c *Config) IPHashPepper() []byte { return c.ipHashPepper }

// TOTPAESKey returns the 32-byte AES key for TOTP secret encryption.
func (c *Config) TOTPAESKey() [32]byte { return c.totpAESKey }

// CORSAllowsAny reports whether the wildcard "*" origin is configured.
func (c *Config) CORSAllowsAny() bool {
	for _, o := range c.CORSAllowedOrigins {
		if o == "*" {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// small env helpers
// ---------------------------------------------------------------------------

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getint(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return def
}

func getbool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
			return b
		}
	}
	return def
}

func getdur(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(strings.TrimSpace(v)); err == nil {
			return d
		}
	}
	return def
}

// splitList parses a comma-or-space separated list, trimming empties.
func splitList(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' })
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

// parseHeaders parses OTEL_EXPORTER_OTLP_HEADERS ("k1=v1,k2=v2").
func parseHeaders(s string) map[string]string {
	out := map[string]string{}
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out
}

// endpointFromDSN extracts the OTLP base endpoint from an Uptrace DSN of the
// form "https://TOKEN@host[:port][/path]". Returns the base URL and whether it
// is insecure (http).
func endpointFromDSN(dsn string) (endpoint string, insecure bool) {
	// Very small, dependency-free parse: scheme://[userinfo@]host
	scheme := "https"
	rest := dsn
	if i := strings.Index(rest, "://"); i >= 0 {
		scheme = rest[:i]
		rest = rest[i+3:]
	}
	if at := strings.LastIndex(rest, "@"); at >= 0 {
		rest = rest[at+1:]
	}
	// strip any path/query
	if q := strings.IndexAny(rest, "/?"); q >= 0 {
		rest = rest[:q]
	}
	return scheme + "://" + rest, scheme == "http"
}

// defaultInstanceID returns a stable-ish identifier used when the operator did
// not supply one (mostly for bare `go run` during development).
func defaultInstanceID() string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return "local-" + hex.EncodeToString(mustRandom(4))
}

func mustRandom(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("config: crypto/rand failed: " + err.Error())
	}
	return b
}

// DecodeKeyMaterial is a helper exposed for tests and tooling: it accepts hex,
// base64 (std or url), or raw bytes and returns them. Unused by Load directly
// but kept for completeness where callers want exact-length keys.
func DecodeKeyMaterial(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if b, err := hex.DecodeString(s); err == nil && len(s)%2 == 0 {
		return b, nil
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return []byte(s), nil
}
internal/config/config_test.go
go
package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	t.Setenv("GOTUNNELS_DEV", "1") // auto-generate secrets
	// DATABASE_URL deliberately unset.
	if _, err := Load(); err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoad_RequiresSecretsWhenNotDev(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x/y")
	t.Setenv("GOTUNNELS_DEV", "0")
	// no pepper / totp key
	if _, err := Load(); err == nil {
		t.Fatal("expected error when secrets are missing and not in dev mode")
	}
}

func TestLoad_DevGeneratesSecrets(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_DEV", "1")
	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.IPHashPepper()) == 0 {
		t.Error("expected a generated IP hash pepper in dev mode")
	}
	var zero [32]byte
	if c.TOTPAESKey() == zero {
		t.Error("expected a non-zero TOTP AES key in dev mode")
	}
	if c.CSPMode != "report-only" {
		t.Errorf("expected default CSP mode report-only, got %q", c.CSPMode)
	}
	if !strings.Contains(c.CSPPolicy, "default-src 'self'") {
		t.Errorf("default CSP policy should lock to self, got %q", c.CSPPolicy)
	}
}

func TestLoad_TOTPKeyDerivedDeterministically(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_IP_HASH_PEPPER", "pepper")
	t.Setenv("GOTUNNELS_TOTP_ENCRYPTION_KEY", "the-same-secret")

	c1, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	c2, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c1.TOTPAESKey() != c2.TOTPAESKey() {
		t.Error("TOTP key derivation must be deterministic for a given secret")
	}
}

func TestValidate_CSPMode(t *testing.T) {
	c := &Config{
		DatabaseURL: "x",
		HTTPAddr:    ":8080",
		SessionTTL:  time.Hour,
		RPOrigins:   []string{"http://localhost"},
		CSPMode:     "nonsense",
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected invalid CSP mode to fail validation")
	}
	c.CSPMode = "enforce"
	if err := c.Validate(); err != nil {
		t.Fatalf("enforce should be valid: %v", err)
	}
}

func TestSplitList(t *testing.T) {
	cases := map[string][]string{
		"a,b,c":            {"a", "b", "c"},
		"a b c":            {"a", "b", "c"},
		" a , b ,, c ":     {"a", "b", "c"},
		"https://x https:": {"https://x", "https:"},
		"":                 {},
	}
	for in, want := range cases {
		got := splitList(in)
		if len(got) != len(want) {
			t.Errorf("splitList(%q) len = %d, want %d (%v)", in, len(got), len(want), got)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("splitList(%q)[%d] = %q, want %q", in, i, got[i], want[i])
			}
		}
	}
}

func TestParseHeaders(t *testing.T) {
	h := parseHeaders("uptrace-dsn=https://tok@api.uptrace.dev,x-extra = val ")
	if h["uptrace-dsn"] != "https://tok@api.uptrace.dev" {
		t.Errorf("unexpected dsn header: %q", h["uptrace-dsn"])
	}
	if h["x-extra"] != "val" {
		t.Errorf("unexpected x-extra header: %q", h["x-extra"])
	}
}

func TestEndpointFromDSN(t *testing.T) {
	cases := []struct {
		dsn          string
		wantEndpoint string
		wantInsecure bool
	}{
		{"https://TOKEN@api.uptrace.dev?grpc=4317", "https://api.uptrace.dev", false},
		{"https://TOKEN@api.uptrace.dev:443/v1", "https://api.uptrace.dev:443", false},
		{"http://token@localhost:14318", "http://localhost:14318", true},
	}
	for _, tc := range cases {
		ep, insecure := endpointFromDSN(tc.dsn)
		if ep != tc.wantEndpoint {
			t.Errorf("endpointFromDSN(%q) endpoint = %q, want %q", tc.dsn, ep, tc.wantEndpoint)
		}
		if insecure != tc.wantInsecure {
			t.Errorf("endpointFromDSN(%q) insecure = %v, want %v", tc.dsn, insecure, tc.wantInsecure)
		}
	}
}

func TestResolveTelemetry_DisabledByDefault(t *testing.T) {
	// Ensure a clean env for the OTLP vars.
	t.Setenv("UPTRACE_DSN", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	c := &Config{}
	c.resolveTelemetry()
	if c.Telemetry.Enabled {
		t.Error("telemetry should be disabled when no endpoint/DSN is configured")
	}
}

func TestResolveTelemetry_UptraceDSN(t *testing.T) {
	t.Setenv("UPTRACE_DSN", "https://secret@api.uptrace.dev?grpc=4317")
	c := &Config{}
	c.resolveTelemetry()
	if !c.Telemetry.Enabled {
		t.Fatal("telemetry should be enabled with an Uptrace DSN")
	}
	if c.Telemetry.EndpointURL != "https://api.uptrace.dev" {
		t.Errorf("endpoint = %q", c.Telemetry.EndpointURL)
	}
	if c.Telemetry.Headers["uptrace-dsn"] == "" {
		t.Error("expected uptrace-dsn header to be set")
	}
}

func TestResolveTelemetry_MetricsDefaults(t *testing.T) {
	// Uptrace-friendly defaults must hold even when the OTEL_* metrics vars
	// are entirely absent: delta temporality + base2 exponential histograms
	// + gzip compression.
	t.Setenv("UPTRACE_DSN", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_COMPRESSION", "")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION", "")
	c := &Config{}
	c.resolveTelemetry()
	if got := c.Telemetry.Compression; got != "gzip" {
		t.Errorf("default compression = %q, want gzip", got)
	}
	if got := c.Telemetry.MetricsTemporality; got != "delta" {
		t.Errorf("default temporality = %q, want delta", got)
	}
	if got := c.Telemetry.MetricsHistogram; got != "base2_exponential_bucket_histogram" {
		t.Errorf("default histogram aggregation = %q, want base2_exponential_bucket_histogram", got)
	}
}

func TestResolveTelemetry_MetricsOverrides(t *testing.T) {
	// The OTel spec documents UPPERCASE values (DELTA, CUMULATIVE,
	// BASE2_EXPONENTIAL_BUCKET_HISTOGRAM...); we must accept them and
	// normalize to lowercase.
	t.Setenv("UPTRACE_DSN", "https://secret@api.uptrace.dev")
	t.Setenv("OTEL_EXPORTER_OTLP_COMPRESSION", "NONE")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "CUMULATIVE")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION", "EXPLICIT_BUCKET_HISTOGRAM")
	c := &Config{}
	c.resolveTelemetry()
	if got := c.Telemetry.Compression; got != "none" {
		t.Errorf("compression = %q, want none", got)
	}
	if got := c.Telemetry.MetricsTemporality; got != "cumulative" {
		t.Errorf("temporality = %q, want cumulative", got)
	}
	if got := c.Telemetry.MetricsHistogram; got != "explicit_bucket_histogram" {
		t.Errorf("histogram aggregation = %q, want explicit_bucket_histogram", got)
	}
}

func TestGetHelpers(t *testing.T) {
	t.Setenv("X_INT", "42")
	if getint("X_INT", 0) != 42 {
		t.Error("getint failed")
	}
	if getint("X_MISSING", 7) != 7 {
		t.Error("getint default failed")
	}
	t.Setenv("X_BOOL", "true")
	if !getbool("X_BOOL", false) {
		t.Error("getbool failed")
	}
	t.Setenv("X_DUR", "250ms")
	if getdur("X_DUR", 0) != 250*time.Millisecond {
		t.Error("getdur failed")
	}
}
internal/telemetry/telemetry.go
go
// Package telemetry wires the vendor-neutral OpenTelemetry Go SDK for all three
// signals (traces, metrics, logs) and exports them over OTLP/HTTP.
//
// We deliberately never import any Uptrace SDK. Uptrace (or any OTLP backend)
// is configured purely via a DSN / endpoint, so swapping backends is a config
// change, not a code change.
//
// Logs are always written to stdout as JSON (so `podman logs` is useful) and,
// when an OTLP endpoint is configured, additionally shipped to the collector
// via the OpenTelemetry log bridge. When no endpoint is configured, traces and
// metrics use no-op providers and only stdout logging remains.
package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/kusl/GoTunnels/internal/config"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellog "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// scopeName is the instrumentation scope for logs/traces/metrics we emit
// directly (as opposed to via otelhttp instrumentation).
const scopeName = "github.com/kusl/GoTunnels"

// Providers bundles what the application needs after setup.
type Providers struct {
	Logger   *slog.Logger
	shutdown []func(context.Context) error
}

// Shutdown flushes and stops all providers. Safe to call once during teardown.
func (p *Providers) Shutdown(ctx context.Context) error {
	var errs []error
	for i := len(p.shutdown) - 1; i >= 0; i-- {
		if err := p.shutdown[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Setup initializes telemetry from configuration. The returned Providers always
// has a usable Logger, even on partial failure.
func Setup(ctx context.Context, cfg *config.Config) (*Providers, error) {
	p := &Providers{}

	// stdout logging is always on.
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	if !cfg.Telemetry.Enabled {
		p.Logger = slog.New(stdoutHandler).With(
			slog.String("service.name", cfg.ServiceName),
			slog.String("service.instance.id", cfg.InstanceID),
		)
		p.Logger.Info("telemetry disabled: logging to stdout only, no-op traces/metrics")
		return p, nil
	}

	res, err := buildResource(ctx, cfg)
	if err != nil {
		// A resource failure is not fatal; fall back to stdout logging.
		l := slog.New(stdoutHandler)
		l.Error("telemetry: failed to build resource; continuing with stdout logging", slog.Any("err", err))
		p.Logger = l
		return p, nil
	}

	// ---- Traces -------------------------------------------------------
	traceExp, err := otlptracehttp.New(ctx, traceHTTPOpts(cfg)...)
	if err != nil {
		return fallback(stdoutHandler, cfg, err), nil
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))
	p.shutdown = append(p.shutdown, tp.Shutdown)

	// ---- Metrics ------------------------------------------------------
	metricExp, err := otlpmetrichttp.New(ctx, metricHTTPOpts(cfg)...)
	if err != nil {
		return fallback(stdoutHandler, cfg, err), nil
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp,
			sdkmetric.WithInterval(30*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	p.shutdown = append(p.shutdown, mp.Shutdown)

	// ---- Logs ---------------------------------------------------------
	logExp, err := otlploghttp.New(ctx, logHTTPOpts(cfg)...)
	if err != nil {
		return fallback(stdoutHandler, cfg, err), nil
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)
	otellog.SetLoggerProvider(lp)
	p.shutdown = append(p.shutdown, lp.Shutdown)

	// Fan out slog to BOTH stdout and the OTel log bridge.
	otelHandler := otelslog.NewHandler(scopeName, otelslog.WithLoggerProvider(lp))
	p.Logger = slog.New(NewMultiHandler(stdoutHandler, otelHandler)).With(
		slog.String("service.name", cfg.ServiceName),
		slog.String("service.instance.id", cfg.InstanceID),
	)
	p.Logger.Info("telemetry enabled",
		slog.String("otlp.endpoint", cfg.Telemetry.EndpointURL),
		slog.Bool("otlp.insecure", cfg.Telemetry.Insecure),
		slog.String("otlp.compression", cfg.Telemetry.Compression),
		slog.String("otlp.metrics.temporality", cfg.Telemetry.MetricsTemporality),
		slog.String("otlp.metrics.histogram", cfg.Telemetry.MetricsHistogram),
	)
	return p, nil
}

func fallback(h slog.Handler, cfg *config.Config, cause error) *Providers {
	l := slog.New(h).With(slog.String("service.name", cfg.ServiceName))
	l.Error("telemetry: exporter setup failed; continuing with stdout logging", slog.Any("err", cause))
	return &Providers{Logger: l}
}

func buildResource(ctx context.Context, cfg *config.Config) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.Version),
			semconv.ServiceInstanceID(cfg.InstanceID),
		),
	)
}

func traceHTTPOpts(cfg *config.Config) []otlptracehttp.Option {
	opts := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(cfg.Telemetry.EndpointURL)}
	if cfg.Telemetry.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(cfg.Telemetry.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Telemetry.Headers))
	}
	if cfg.Telemetry.Compression == "gzip" {
		opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.GzipCompression))
	}
	return opts
}

func metricHTTPOpts(cfg *config.Config) []otlpmetrichttp.Option {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpointURL(cfg.Telemetry.EndpointURL),
		otlpmetrichttp.WithTemporalitySelector(temporalitySelector(cfg.Telemetry.MetricsTemporality)),
		otlpmetrichttp.WithAggregationSelector(aggregationSelector(cfg.Telemetry.MetricsHistogram)),
	}
	if cfg.Telemetry.Insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	if len(cfg.Telemetry.Headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(cfg.Telemetry.Headers))
	}
	if cfg.Telemetry.Compression == "gzip" {
		opts = append(opts, otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression))
	}
	return opts
}

// temporalitySelector maps OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE
// (already lowercased by config) onto an SDK TemporalitySelector, following
// the OTLP exporter spec:
//
//	delta:      Counter, ObservableCounter, Histogram -> delta; others cumulative
//	lowmemory:  Counter, Histogram -> delta; others cumulative
//	cumulative: everything cumulative (the SDK default)
//
// Uptrace prefers delta, which is our config default. Unknown values fall back
// to the SDK default (cumulative) rather than failing startup.
func temporalitySelector(pref string) sdkmetric.TemporalitySelector {
	switch strings.ToLower(pref) {
	case "delta":
		return func(k sdkmetric.InstrumentKind) metricdata.Temporality {
			switch k {
			case sdkmetric.InstrumentKindCounter,
				sdkmetric.InstrumentKindObservableCounter,
				sdkmetric.InstrumentKindHistogram:
				return metricdata.DeltaTemporality
			default:
				return metricdata.CumulativeTemporality
			}
		}
	case "lowmemory":
		return func(k sdkmetric.InstrumentKind) metricdata.Temporality {
			switch k {
			case sdkmetric.InstrumentKindCounter,
				sdkmetric.InstrumentKindHistogram:
				return metricdata.DeltaTemporality
			default:
				return metricdata.CumulativeTemporality
			}
		}
	default: // "cumulative" or anything unrecognized
		return sdkmetric.DefaultTemporalitySelector
	}
}

// aggregationSelector maps OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION
// (already lowercased by config) onto an SDK AggregationSelector. With
// base2_exponential_bucket_histogram (our default, and what Uptrace
// recommends), Histogram instruments use auto-scaling exponential buckets
// (MaxSize 160 / MaxScale 20, the SDK's own defaults for this aggregation)
// instead of fixed explicit boundaries; every other instrument kind keeps the
// SDK default aggregation.
func aggregationSelector(hist string) sdkmetric.AggregationSelector {
	if strings.ToLower(hist) != "base2_exponential_bucket_histogram" {
		return sdkmetric.DefaultAggregationSelector
	}
	return func(k sdkmetric.InstrumentKind) sdkmetric.Aggregation {
		if k == sdkmetric.InstrumentKindHistogram {
			return sdkmetric.AggregationBase2ExponentialHistogram{
				MaxSize:  160,
				MaxScale: 20,
			}
		}
		return sdkmetric.DefaultAggregationSelector(k)
	}
}

func logHTTPOpts(cfg *config.Config) []otlploghttp.Option {
	opts := []otlploghttp.Option{otlploghttp.WithEndpointURL(cfg.Telemetry.EndpointURL)}
	if cfg.Telemetry.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}
	if len(cfg.Telemetry.Headers) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(cfg.Telemetry.Headers))
	}
	if cfg.Telemetry.Compression == "gzip" {
		opts = append(opts, otlploghttp.WithCompression(otlploghttp.GzipCompression))
	}
	return opts
}
scripts/lib.sh
bash
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
detect_runtime() {
  if [ -n "${CR:-}" ] && [ -n "${COMPOSE:-}" ]; then return 0; fi
  if command -v podman >/dev/null 2>&1; then
    CR=podman
    if podman compose version >/dev/null 2>&1; then
      COMPOSE="podman compose"
    elif command -v podman-compose >/dev/null 2>&1; then
      COMPOSE="podman-compose"
    else
      die "podman found but neither 'podman compose' nor 'podman-compose' is available"
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
GOTUNNELS_CSP_POLICY="default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self' https:; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
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
  if [ -z "$p" ]; then p="gotunnels-$(gen_secret 4 | tr 'A-Z' 'a-z' | tr -cd 'a-z0-9' | cut -c1-6)"; fi
  # compose project names must be lowercase alnum/dash/underscore.
  p="$(printf '%s' "$p" | tr 'A-Z' 'a-z' | tr -cd 'a-z0-9_-')"
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
  local i cid status st hs
  log "waiting for '$svc' to become ready (up to ${timeout}s)"
  for i in $(seq 1 "$timeout"); do
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
  local project="$1" svc="$2" timeout="${3:-60}" i cid url
  for i in $(seq 1 "$timeout"); do
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
  local project="$1" svc="$2" pat="$3" timeout="${4:-60}" i cid
  for i in $(seq 1 "$timeout"); do
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
scripts/up.sh
bash
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
#    podman-compose does NOT honor `--no-deps`: step 3's
#    `up -d --no-deps frontend cloudflared-frontend` also created and started
#    the api container (frontend depends_on api), freezing its environment
#    with the pre-discovery values. A plain `up -d` here then hits "name
#    already in use", keeps that stale container, and the API never sees the
#    tunnel-derived GOTUNNELS_RP_ID / RP_ORIGINS / CORS_ALLOWED_ORIGINS just
#    exported above — which is exactly the CORS-on-signup failure (and, worse,
#    silently broken passkeys: RP ID stuck at its bootstrap value). So: if an
#    api container already exists for this project, remove it and let compose
#    recreate it with the current environment. Under docker compose (which
#    honors --no-deps) no such container exists yet and this is a no-op.
_stale_api_cid="$(cid_of "$PROJECT" api)"
if [ -n "$_stale_api_cid" ]; then
  warn "api container was pre-created by a dependency with pre-discovery env; recreating it"
  "$CR" rm -f "$_stale_api_cid" >/dev/null 2>&1 || true
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
scripts/ci-container-test.sh (new file)
bash
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

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pw
please continue

d)"

shellcheck source=lib.sh
. "$SCRIPT_DIR/lib.sh"

detect_runtime
ensure_env
reset_runtime_env

PROJECT="
(
r
e
s
o
l
v
e
p
r
o
j
e
c
t
"
(resolve 
p
​
 roject"{1:-gotunnels-ci}")"
export GOTUNNELS_INSTANCE_ID="{GOTUNNELS_INSTANCE_ID:- PROJECT}"
NET="${PROJECT}_default"

Pinned, fully qualified (podman does not assume docker.io). busybox wget in
the caddy image can't do PUT/DELETE, hence a real curl.
CURL_IMAGE="docker.io/curlimages/curl:8.10.1"

log "container test instance: $PROJECT"

---------------------------------------------------------------------------
teardown + failure diagnostics
---------------------------------------------------------------------------
cleanup() {
status=$?
if [ "status" -ne 0 ] && [ -n " {CR:-}" ]; then
err "container test FAILED (exit $status) — dumping service logs"
for _svc in db api frontend; do
_cid="
(
c
i
d
o
f
"
(cid 
o
​
 f"PROJECT" "$_svc")"
if [ -n "$_cid" ]; then
err "----- logs: $_svc -----"
"
C
R
"
l
o
g
s
−
−
t
a
i
l
120
"
CR"logs−−tail120"_cid" 2>&1 | sed 's/^/ /' >&2 || true
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

---------------------------------------------------------------------------
assertion helpers
---------------------------------------------------------------------------
FAILED=0

ccurl — run curl inside a one-shot container on the compose network. -sS so
transport errors surface; --max-time caps a wedged request, not the test.
ccurl() {
"
C
R
"
r
u
n
−
−
r
m
−
−
n
e
t
w
o
r
k
"
CR"run−−rm−−network"NET" "
C
U
R
L
I
M
A
G
E
"
−
s
S
−
−
m
a
x
−
t
i
m
e
30
"
CURL 
I
​
 MAGE"−sS−−max−time30"@"
}

assert_contains BODY NEEDLE LABEL — grep must DRAIN stdin (>/dev/null, not
-q): with -q it exits at first match, printf takes SIGPIPE, and under
pipefail a PASS would read as FAIL. Same trap as wait_for_log in lib.sh.
assert_contains() {
if printf '%s' "$1" | grep -F -- "$2" >/dev/null; then
ok "PASS: $3"
else
err "FAIL: $3"
err " expected to find: $2"
printf '%s\n' "$1" | head -20 | sed 's/^/ got: /' >&2
FAILED=1
fi
}

assert_icontains() { # case-insensitive variant (headers, HTML)
if printf '%s' "$1" | grep -Fi -- "$2" >/dev/null; then
ok "PASS: $3"
else
err "FAIL: $3"
err " expected to find (case-insensitive): $2"
printf '%s\n' "$1" | head -20 | sed 's/^/ got: /' >&2
FAILED=1
fi
}

assert_not_contains() {
if printf '%s' "$1" | grep -F -- "$2" >/dev/null; then
err "FAIL: $3"
err " expected NOT to find: $2"
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

---------------------------------------------------------------------------
1) build both images
---------------------------------------------------------------------------
log "building images (Containerfile.api + Containerfile.frontend)…"
dc -p "$PROJECT" build

---------------------------------------------------------------------------
2) database, then api + frontend (no tunnels in CI)
---------------------------------------------------------------------------
log "starting database…"
dc -p "$PROJECT" up -d db
wait_healthy "$PROJECT" db 90
pg_isready -U "
P
O
S
T
G
R
E
S
U
S
E
R
:
−
g
o
t
u
n
n
e
l
s
"
−
d
"
POSTGRES 
U
​
 SER:−gotunnels"−d"{POSTGRES_DB:-gotunnels}" -q

log "starting api + frontend…"

podman-compose ignores --no-deps and will re-process db; it prints a
"name already in use" for the existing container but exits 0, so this is
safe under both runtimes (docker compose honors --no-deps outright).
dc -p "$PROJECT" up -d --no-deps api frontend

Unlike up.sh (best-effort warn), CI must hard-fail if the API never comes up.
wait_for_log "$PROJECT" api 'http server listening' 90
|| die "api never logged 'http server listening' — startup failed"
ok "api is listening"

---------------------------------------------------------------------------
3) migrations actually ran
---------------------------------------------------------------------------
db_cid="
(
c
i
d
o
f
"
(cid 
o
​
 f"PROJECT" db)"
[ -n "$db_cid" ] || die "no db container id"
mig="
(
"
("CR" exec "
d
b
c
i
d
"
p
s
q
l
−
U
"
db 
c
​
 id"psql−U"{POSTGRES_USER:-gotunnels}" -d "${POSTGRES_DB:-gotunnels}"
-tAc 'SELECT COALESCE(max(version), 0) FROM schema_migrations' | tr -d '[:space:]')"
log "schema_migrations max(version) = $mig"
if [ "${mig:-0}" -ge 7 ] 2>/dev/null; then
ok "PASS: migrations applied (>= 7)"
else
err "FAIL: expected migration version >= 7, got '$mig'"
FAILED=1
fi

---------------------------------------------------------------------------
4) health endpoints
---------------------------------------------------------------------------
ccurl -f http://api:8080/healthz >/dev/null && ok "PASS: /healthz"
ccurl -f http://api:8080/readyz >/dev/null && ok "PASS: /readyz (db reachable)"

---------------------------------------------------------------------------
5) signup -> bearer token
---------------------------------------------------------------------------
signup_resp="$(ccurl -X POST http://api:8080/api/signup
-H 'Content-Type: application/json'
-d '{"username":"ciuser","password":"ci-password-123","display_name":"CI User"}')"
TOKEN="
(
p
r
i
n
t
f
′
(printf 
′
 signup_resp" | sed -n 's/."token":"([^"])".*/\1/p')"
if [ -n "$TOKEN" ]; then
ok "PASS: signup issued a session token"
else
err "FAIL: signup did not return a token"
printf '%s\n' "$signup_resp" | head -5 | sed 's/^/ got: /' >&2
die "cannot continue without a session token"
fi
AUTH="Authorization: Bearer $TOKEN"

---------------------------------------------------------------------------
6) captcha sync — REGRESSION for the 'unknown + unknown' 500.
Before the ::bigint casts in store.SyncCaptchaStats, every call to this
endpoint failed with {"error":"internal server error"}; both the insert
and the update paths below would have 500ed.
---------------------------------------------------------------------------
sync1="$(ccurl -X POST http://api:8080/api/captcha/sync
-H "$AUTH" -H 'Content-Type: application/json'
-d '{"manual_delta":3,"auto_delta":5,"current_streak":2,"best_streak":4}')"
assert_contains "$sync1" '"total_solves":8' "captcha sync insert path (3 manual + 5 auto = 8)"
assert_contains "$sync1" '"best_streak":4' "captcha sync insert path records best streak"
assert_not_contains "$sync1" 'internal server error' "captcha sync does not 500 (insert)"

sync2="$(ccurl -X POST http://api:8080/api/captcha/sync
-H "$AUTH" -H 'Content-Type: application/json'
-d '{"manual_delta":4,"auto_delta":8,"current_streak":1,"best_streak":2}')"
assert_contains "$sync2" '"total_solves":20' "captcha sync update path accumulates (8 + 12 = 20)"
assert_contains "$sync2" '"best_streak":4' "captcha sync keeps GREATEST best streak (4 > 2)"
assert_contains "$sync2" '"current_streak":1' "captcha sync current streak is last-write-wins"
assert_not_contains "$sync2" 'internal server error' "captcha sync does not 500 (update)"

stats="
(
c
c
u
r
l
−
H
"
(ccurl−H"AUTH" http://api:8080/api/captcha/stats)"
assert_contains "$stats" '"total_solves":20' "captcha stats reads back the synced totals"

lb="
(
c
c
u
r
l
−
H
"
(ccurl−H"AUTH" http://api:8080/api/captcha/leaderboard)"
assert_contains "$lb" '"username":"ciuser"' "leaderboard ranks the CI user"

---------------------------------------------------------------------------
7) prefs round trip
---------------------------------------------------------------------------
put_pref="$(ccurl -X PUT http://api:8080/api/prefs/theme
-H "$AUTH" -H 'Content-Type: application/json' -d '{"value":"dark"}')"
assert_contains "$put_pref" '"status":"saved"' "prefs PUT saves"

get_pref="
(
c
c
u
r
l
−
H
"
(ccurl−H"AUTH" http://api:8080/api/prefs/theme)"
assert_contains "$get_pref" '"value":"dark"' "prefs GET returns the stored value"
assert_contains "$get_pref" '"exists":true' "prefs GET reports exists=true"

---------------------------------------------------------------------------
8) notes create / list / delete
---------------------------------------------------------------------------
note_resp="$(ccurl -X POST http://api:8080/api/notes
-H "$AUTH" -H 'Content-Type: application/json'
-d '{"body":"hello from the CI smoke test"}')"
NOTE_ID="
(
p
r
i
n
t
f
′
(printf 
′
 note_resp" | sed -n 's/."note":{"id":([0-9][0-9]).*/\1/p')"
if [ -n "$NOTE_ID" ]; then
ok "PASS: note created (id=$NOTE_ID)"
else
err "FAIL: note creation did not return a numeric id"
printf '%s\n' "$note_resp" | head -5 | sed 's/^/ got: /' >&2
FAILED=1
fi

notes_list="
(
c
c
u
r
l
−
H
"
(ccurl−H"AUTH" http://api:8080/api/notes)"
assert_contains "$notes_list" 'hello from the CI smoke test' "notes list contains the new note"

if [ -n "$NOTE_ID" ]; then
del_resp="$(ccurl -X DELETE "http://api:8080/api/notes/$NOTE_ID" -H "$AUTH")"
assert_contains "$del_resp" '"status":"deleted"' "note delete (hard delete semantics)"
notes_after="
(
c
c
u
r
l
−
H
"
(ccurl−H"AUTH" http://api:8080/api/notes)"
assert_not_contains "$notes_after" 'hello from the CI smoke test' "deleted note is gone from the list"
fi

---------------------------------------------------------------------------
9) CORS preflight (the middleware must answer OPTIONS itself with 204)
---------------------------------------------------------------------------
pre_code="$(ccurl -o /dev/null -w '%{http_code}' -X OPTIONS http://api:8080/api/signup
-H 'Origin: https://example.com' -H 'Access-Control-Request-Method: POST')"
assert_eq "$pre_code" "204" "CORS preflight answered with 204"

---------------------------------------------------------------------------
10) frontend serves the app with the CSP header
---------------------------------------------------------------------------
front="$(ccurl -D - http://frontend:8080/)"
assert_icontains "$front" 'content-security-policy' "frontend sends a CSP header"
assert_icontains "$front" '<html' "frontend serves the HTML app"

---------------------------------------------------------------------------
verdict
---------------------------------------------------------------------------
if [ "$FAILED" -ne 0 ]; then
die "container smoke test had failures (see FAIL lines above)"
fi
ok "container smoke test passed."


**Wait — stop.** I spotted a typo I introduced while pasting: line 22 above says `pw d)` instead of `pwd)`. The file on disk is correct (it passed `bash -n` and shellcheck); that corruption happened only in this chat rendering. Use this corrected line 22:

```bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
```

Everything else in the block above is byte-identical to the validated file.

## `.github/workflows/ci.yaml`

```yaml
# Build and test. Deliberately thin: all logic lives in scripts/test.sh so the
# exact same commands run locally and here.
name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7

      - uses: actions/setup-go@v6
        with:
          go-version: "stable"
          # go.sum is intentionally absent on first import; do not fail on it.
          cache: false

      - name: Build, vet, unit tests
        run: bash scripts/test.sh all

  # Everything scripts/test.sh cannot see: that both Containerfiles build,
  # that compose wiring + env substitution is correct, that migrations run
  # against a real Postgres, and that the HTTP surface works end to end —
  # including the captcha-sync regression that only reproduces against real
  # Postgres via the pgx extended protocol. Ubuntu runners ship podman;
  # podman-compose comes from pipx. No tunnels are started and no host ports
  # are published; assertions run from a curl container on the compose network.
  containers:
    runs-on: ubuntu-latest
    timeout-minutes: 25
    steps:
      - uses: actions/checkout@v7

      - name: Install podman-compose
        run: pipx install podman-compose

      - name: Runtime versions
        run: |
          podman version
          podman-compose version

      - name: Build images + container smoke test
        run: bash scripts/ci-container-test.sh gotunnels-ci
```

## `compose.yaml`

```yaml
# compose.yaml — the whole GoTunnels stack.
#
# Design rules embodied here:
#   * No service publishes a host port. The only traffic leaving the machine is
#     the two outbound Cloudflare Quick Tunnel connections.
#   * Everything is addressed by service name on the internal Compose network.
#   * Every instance is isolated by the Compose project name (passed via -p),
#     which prefixes container, network, and volume names — so N instances run
#     side by side without collision.
#   * Values that can be generated are generated by scripts/up.sh into .env;
#     nothing secret is hardcoded here.
#
# Runtime-discovered values (the frontend's tunnel URL, used for the WebAuthn
# RP ID and CORS allow-list) are exported into the shell by scripts/up.sh
# before the api service is (re)started, and read here via ${VAR} substitution.

services:
  db:
    image: docker.io/library/postgres:16-alpine
    environment:
      POSTGRES_USER: "${POSTGRES_USER:-gotunnels}"
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD must be set (run scripts/up.sh)}"
      POSTGRES_DB: "${POSTGRES_DB:-gotunnels}"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-gotunnels} -d ${POSTGRES_DB:-gotunnels}"]
      interval: 3s
      timeout: 3s
      retries: 20
    restart: unless-stopped

  api:
    build:
      context: .
      dockerfile: Containerfile.api
      args:
        VERSION: "${GOTUNNELS_VERSION:-dev}"
    depends_on:
      db:
        condition: service_healthy
    environment:
      GOTUNNELS_INSTANCE_ID: "${GOTUNNELS_INSTANCE_ID:-default}"
      GOTUNNELS_VERSION: "${GOTUNNELS_VERSION:-dev}"
      GOTUNNELS_HTTP_ADDR: ":8080"
      DATABASE_URL: "postgres://${POSTGRES_USER:-gotunnels}:${POSTGRES_PASSWORD}@db:5432/${POSTGRES_DB:-gotunnels}?sslmode=disable"
      GOTUNNELS_IP_HASH_PEPPER: "${GOTUNNELS_IP_HASH_PEPPER:?set by scripts/up.sh}"
      GOTUNNELS_TOTP_ENCRYPTION_KEY: "${GOTUNNELS_TOTP_ENCRYPTION_KEY:?set by scripts/up.sh}"
      # Discovered at runtime by scripts/up.sh (frontend tunnel URL). Defaults
      # keep `compose config` valid before discovery.
      GOTUNNELS_RP_ID: "${GOTUNNELS_RP_ID:-localhost}"
      GOTUNNELS_RP_DISPLAY_NAME: "${GOTUNNELS_RP_DISPLAY_NAME:-GoTunnels}"
      GOTUNNELS_RP_ORIGINS: "${GOTUNNELS_RP_ORIGINS:-http://localhost:8080}"
      GOTUNNELS_CORS_ALLOWED_ORIGINS: "${GOTUNNELS_CORS_ALLOWED_ORIGINS:-*}"
      GOTUNNELS_CSP_MODE: "${GOTUNNELS_CSP_MODE:-report-only}"
      # Telemetry: a single Uptrace (cloud or self-hosted) DSN, or OTLP env vars.
      UPTRACE_DSN: "${UPTRACE_DSN:-}"
      OTEL_EXPORTER_OTLP_ENDPOINT: "${OTEL_EXPORTER_OTLP_ENDPOINT:-}"
      OTEL_EXPORTER_OTLP_HEADERS: "${OTEL_EXPORTER_OTLP_HEADERS:-}"
      OTEL_EXPORTER_OTLP_COMPRESSION: "${OTEL_EXPORTER_OTLP_COMPRESSION:-gzip}"
      OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE: "${OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE:-delta}"
      OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION: "${OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION:-base2_exponential_bucket_histogram}"
      OTEL_SERVICE_NAME: "${OTEL_SERVICE_NAME:-gotunnels-api}"
      GOTUNNELS_DEV: "${GOTUNNELS_DEV:-false}"
    restart: unless-stopped

  frontend:
    build:
      context: .
      dockerfile: Containerfile.frontend
    depends_on:
      - api
    environment:
      GOTUNNELS_CSP_HEADER_NAME: "${GOTUNNELS_CSP_HEADER_NAME:-Content-Security-Policy-Report-Only}"
      GOTUNNELS_CSP_POLICY: "${GOTUNNELS_CSP_POLICY:-default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self' https:; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'}"
    restart: unless-stopped

  cloudflared-frontend:
    image: docker.io/cloudflare/cloudflared:latest
    depends_on:
      - frontend
    command: ["tunnel", "--no-autoupdate", "--url", "http://frontend:8080"]
    restart: unless-stopped

  cloudflared-api:
    image: docker.io/cloudflare/cloudflared:latest
    depends_on:
      - api
    command: ["tunnel", "--no-autoupdate", "--url", "http://api:8080"]
    restart: unless-stopped

volumes:
  pgdata:
```

## `.env.example`

```bash
# .env.example — shape of the environment file.
#
# There are NO real secrets to copy here. scripts/up.sh generates a .env with
# fresh random secrets on first run if one does not already exist. This file
# exists only to document what the stack reads.
#
# Copy to .env and fill in ONLY if you are not using scripts/up.sh:
#   cp .env.example .env

# --- Instance identity -------------------------------------------------------
# Unique per running instance; prefixes containers/network/volumes and appears
# as an OpenTelemetry resource attribute.
GOTUNNELS_INSTANCE_ID=default
GOTUNNELS_VERSION=dev

# --- Database (password is generated by up.sh) -------------------------------
POSTGRES_USER=gotunnels
POSTGRES_DB=gotunnels
POSTGRES_PASSWORD=generate-me

# --- Server secrets (generated by up.sh; never commit real values) -----------
# Pepper mixed into sha256(pepper||ip) so stored IP hashes are not rainbow-tableable.
GOTUNNELS_IP_HASH_PEPPER=generate-me
# Any string; the app derives a 32-byte AES key from it to encrypt TOTP secrets.
GOTUNNELS_TOTP_ENCRYPTION_KEY=generate-me

# --- WebAuthn / CORS (discovered at runtime from the frontend tunnel URL) ----
# up.sh overwrites these once the frontend Quick Tunnel URL is known.
GOTUNNELS_RP_ID=localhost
GOTUNNELS_RP_DISPLAY_NAME=GoTunnels
GOTUNNELS_RP_ORIGINS=http://localhost:8080
GOTUNNELS_CORS_ALLOWED_ORIGINS=*

# --- Content Security Policy (central configuration) -------------------------
# Header name selects report-only vs enforcing:
#   Content-Security-Policy-Report-Only  (default; observe only)
#   Content-Security-Policy              (enforce)
GOTUNNELS_CSP_HEADER_NAME=Content-Security-Policy-Report-Only
GOTUNNELS_CSP_MODE=report-only
GOTUNNELS_CSP_POLICY="default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self' https:; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"

# --- Telemetry (optional) ----------------------------------------------------
# Point at Uptrace Cloud (or any OTLP/HTTP backend) with a single DSN:
#   UPTRACE_DSN=https://<token>@api.uptrace.dev
# Or use the standard OTLP env vars instead:
#   OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp.example.com
#   OTEL_EXPORTER_OTLP_HEADERS=key1=value1,key2=value2
# Leave blank to disable exporting (the app still logs JSON to stdout).
#
# NOTE: exporting UPTRACE_DSN in your shell before running scripts/up.sh also
# works — ensure_env persists it into this file before loading it, so a blank
# UPTRACE_DSN= line here can no longer clobber your export.
UPTRACE_DSN=
OTEL_EXPORTER_OTLP_ENDPOINT=
OTEL_EXPORTER_OTLP_HEADERS=
# gzip-compress OTLP payloads (Uptrace recommendation; set empty to disable).
OTEL_EXPORTER_OTLP_COMPRESSION=gzip
# Metrics temporality: delta | cumulative | lowmemory (spec casing DELTA etc.
# is also accepted). Uptrace prefers delta.
OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta
# Histogram aggregation for Histogram instruments:
#   base2_exponential_bucket_histogram (Uptrace recommendation; default)
#   explicit_bucket_histogram
OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION=base2_exponential_bucket_histogram
OTEL_SERVICE_NAME=gotunnels-api

# --- Dev convenience ---------------------------------------------------------
# When true, missing secrets are auto-generated ephemerally and cookies are not
# marked Secure. Never set true in anything resembling production.
GOTUNNELS_DEV=false
```

## `docs/CONFIGURATION.md`

```markdown
# Configuration

All runtime configuration is centralized. On the Go side every variable is read
in exactly one place — [`internal/config/config.go`](../internal/config/config.go)
— and nowhere else. This file is the reference for what each variable does.

Most values are generated for you by `scripts/up.sh` on first run and written to
a gitignored `.env`. You only ever set variables by hand if you are bypassing
the scripts. See also [`.env.example`](../.env.example).

## Identity

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOTUNNELS_INSTANCE_ID` | hostname / `default` | Unique per instance; also an OTel resource attribute and used to derive the Compose project name. |
| `GOTUNNELS_VERSION` | `dev` | Build/version label (also settable at build time via `-ldflags -X main.version`). |

## HTTP

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOTUNNELS_HTTP_ADDR` | `:8080` | Listen address inside the container. |
| `GOTUNNELS_SHUTDOWN_TIMEOUT` | `15s` | Graceful shutdown budget. |

## Database

| Variable | Default | Meaning |
| --- | --- | --- |
| `DATABASE_URL` | (required) | pgx connection string. Compose builds it from the `POSTGRES_*` values. |
| `POSTGRES_USER` | `gotunnels` | DB user (Compose/Postgres). |
| `POSTGRES_DB` | `gotunnels` | DB name. |
| `POSTGRES_PASSWORD` | generated | DB password; generated by `up.sh`. |
| `GOTUNNELS_DB_MAX_CONNS` | `20` | Pool max connections. |
| `GOTUNNELS_DB_MIN_CONNS` | `2` | Pool min connections. |
| `GOTUNNELS_DB_CONNECT_TIMEOUT` | `30s` | How long startup retries connecting before giving up. |

## Sessions

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOTUNNELS_SESSION_COOKIE_NAME` | `gotunnels_session` | Name of the secondary session cookie. |
| `GOTUNNELS_SESSION_TTL` | `24h` | Session lifetime. |

## Secrets (generated per instance; never commit real values)

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOTUNNELS_IP_HASH_PEPPER` | required (unless `GOTUNNELS_DEV`) | Mixed into `sha256(pepper||ip)` so stored IP hashes are not rainbow-tableable. |
| `GOTUNNELS_TOTP_ENCRYPTION_KEY` | required (unless `GOTUNNELS_DEV`) | Any string; a 32-byte AES-256-GCM key is derived from it (SHA-256) to encrypt TOTP secrets at rest. |

When `GOTUNNELS_DEV=true`, missing secrets are generated ephemerally at startup
with a warning (fine for local hacking, never for anything real).

## WebAuthn / CORS (discovered at runtime)

`scripts/up.sh` overwrites these once the frontend Quick Tunnel URL is known.

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOTUNNELS_RP_ID` | `localhost` | WebAuthn Relying Party ID = the frontend's registrable domain. |
| `GOTUNNELS_RP_DISPLAY_NAME` | `GoTunnels` | Human-readable RP name (also the TOTP issuer). |
| `GOTUNNELS_RP_ORIGINS` | `http://localhost:8080` | Full origin(s) the browser presents (comma/space separated). |
| `GOTUNNELS_CORS_ALLOWED_ORIGINS` | `*` | Exact allowed origin(s), or `*`. Because credentials are used, `*` is echoed as the request origin rather than literally, and never treated as authorization. |

## Content Security Policy (central)

| Variable | Default | Meaning |
| --- | --- | --- |
| `GOTUNNELS_CSP_HEADER_NAME` | `Content-Security-Policy-Report-Only` | Set to `Content-Security-Policy` to **enforce**. |
| `GOTUNNELS_CSP_MODE` | `report-only` | Informational mirror surfaced on `/api/info` (`report-only` or `enforce`). |
| `GOTUNNELS_CSP_POLICY` | strict self-only (see below) | The policy string. Emitted by Caddy; mirrored to the API for `/api/info`. |

The default policy is fully self-hosted — no third-party scripts, styles,
images, fonts, media, or frames:

```
default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self';
font-src 'self'; connect-src 'self' https:; media-src 'self'; object-src 'none';
frame-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'
Because the app has no inline scripts, inline event handlers, or inline styles,
it already satisfies the enforcing form of this policy.

Telemetry (optional)
Resolution order: UPTRACE_DSN first, then the standard OTLP variables, then
disabled (stdout logging only).

Variable	Default	Meaning
UPTRACE_DSN	(empty)	e.g. https://<token>@api.uptrace.dev. Derives the OTLP endpoint and sets the uptrace-dsn header.
OTEL_EXPORTER_OTLP_ENDPOINT	(empty)	Base OTLP/HTTP endpoint, if not using a DSN.
OTEL_EXPORTER_OTLP_HEADERS	(empty)	k1=v1,k2=v2 headers for the exporter.
OTEL_EXPORTER_OTLP_COMPRESSION	gzip	Exporter compression.
OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE	delta	delta, cumulative, or lowmemory (spec casing like DELTA also accepted). Uptrace prefers delta.
OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION	base2_exponential_bucket_histogram	Or explicit_bucket_histogram. Exponential buckets give Uptrace accurate percentiles at lower cost.
OTEL_SERVICE_NAME / GOTUNNELS_SERVICE_NAME	gotunnels-api	OTel service.name.
Dev
Variable	Default	Meaning
GOTUNNELS_DEV	false	Auto-generate missing secrets; mark cookies non-Secure. Local use only.
Script-only variables
These are read by the shell scripts (not the Go app) to control instance
naming and the test toolchain:

Variable	Meaning
GOTUNNELS_PROJECT	Explicit Compose project name (else derived from the instance id or generated).
GO_IMAGE	Go image used by scripts/test.sh when no host Go toolchain is present (default golang:1.26-bookworm).
CR / COMPOSE	Override the detected container runtime / compose command.

---

## Deploying and housekeeping

To redeploy: `git pull && bash scripts/up.sh` — no need to delete `.env` anymore, and doing so is now actively worse (you'd lose the persisted DSN and secrets). The first run after this fix will restore passkey functionality, since the api container will finally boot with the real tunnel-derived RP ID. Existing passkeys registered while RP ID was `localhost` won't work against the tunnel origin and will need re-registering — that's inherent to WebAuthn, not a bug here.

Two credentials need rotating: the GitHub PAT you pasted this session, and the Uptrace DSN — the DSN also sits in plaintext inside `virginia.txt`, so treat both as exposed regardless of whether the files stay private. Neither appears in any file above (I scanned all eleven).

One heads-up on the CI job: `pipx install podman-compose` pulls the latest podman-compose, which may behave slightly differently from your Fedora package around the "name already in use" tolerance in step 2 of the smoke test. If the `containers` job ever fails there, the log dump in the cleanup trap will show exactly which service, and pinning `pipx install podman-compose==1.2.0` is the one-line remedy.
Can you please make sure the reply is complete and give me all the files?


Orchestrated complete file delivery with strategic ordering to prevent corruption
Orchestrated complete file delivery with strategic ordering to prevent corruption
I need to go back and verify that all 11 files were delivered completely and cleanly, since the CI container test script got cut off mid-stream in my previous response. Let me re-deliver everything to make sure nothing's truncated. I'll do a quick validation pass on all the files to confirm they're still intact, then re-deliver the complete set in one message so there's no ambiguity. I'll finalize the file order with the remaining entries and do a quick validation pass to ensure everything's ready before delivering the full list.


Re-validate all 11 files on disk before final complete delivery

Script
I've verified all 11 files are intact and ready, so now I'm pulling a fresh copy of the ci-container-test.sh script to include it accurately in the complete response.


Read the validated ci-container-test.sh fresh for clean delivery

Script
I've verified all the files are intact, and I need to make sure I deliver the complete set of 11 files. The first six files went through cleanly, but the shell script got split awkwardly across the message boundary, and the remaining files came in the second batch. The user wants me to ensure the full reply is complete.

I'll reorganize the delivery with the previously-mangled shell script first to guarantee it comes through cleanly, then follow with the rest in order. Given the volume, I'll keep my intro minimal and focus on getting all the content through without hitting the output limit again.

Done
Yes — verified just now: all 11 files on disk are intact and re-validated (gofmt clean, bash -n + shellcheck zero errors, YAML parses, no secrets). The only casualty last time was the chat rendering of scripts/ci-container-test.sh, which got split mid-line across the message boundary. Here is the complete set of all 11 files, with that one first.

1. scripts/ci-container-test.sh (new file)
bash
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
2. scripts/lib.sh
bash
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
detect_runtime() {
  if [ -n "${CR:-}" ] && [ -n "${COMPOSE:-}" ]; then return 0; fi
  if command -v podman >/dev/null 2>&1; then
    CR=podman
    if podman compose version >/dev/null 2>&1; then
      COMPOSE="podman compose"
    elif command -v podman-compose >/dev/null 2>&1; then
      COMPOSE="podman-compose"
    else
      die "podman found but neither 'podman compose' nor 'podman-compose' is available"
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
GOTUNNELS_CSP_POLICY="default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self' https:; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
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
  if [ -z "$p" ]; then p="gotunnels-$(gen_secret 4 | tr 'A-Z' 'a-z' | tr -cd 'a-z0-9' | cut -c1-6)"; fi
  # compose project names must be lowercase alnum/dash/underscore.
  p="$(printf '%s' "$p" | tr 'A-Z' 'a-z' | tr -cd 'a-z0-9_-')"
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
  local i cid status st hs
  log "waiting for '$svc' to become ready (up to ${timeout}s)"
  for i in $(seq 1 "$timeout"); do
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
  local project="$1" svc="$2" timeout="${3:-60}" i cid url
  for i in $(seq 1 "$timeout"); do
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
  local project="$1" svc="$2" pat="$3" timeout="${4:-60}" i cid
  for i in $(seq 1 "$timeout"); do
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
3. scripts/up.sh
bash
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
#    podman-compose does NOT honor `--no-deps`: step 3's
#    `up -d --no-deps frontend cloudflared-frontend` also created and started
#    the api container (frontend depends_on api), freezing its environment
#    with the pre-discovery values. A plain `up -d` here then hits "name
#    already in use", keeps that stale container, and the API never sees the
#    tunnel-derived GOTUNNELS_RP_ID / RP_ORIGINS / CORS_ALLOWED_ORIGINS just
#    exported above — which is exactly the CORS-on-signup failure (and, worse,
#    silently broken passkeys: RP ID stuck at its bootstrap value). So: if an
#    api container already exists for this project, remove it and let compose
#    recreate it with the current environment. Under docker compose (which
#    honors --no-deps) no such container exists yet and this is a no-op.
_stale_api_cid="$(cid_of "$PROJECT" api)"
if [ -n "$_stale_api_cid" ]; then
  warn "api container was pre-created by a dependency with pre-discovery env; recreating it"
  "$CR" rm -f "$_stale_api_cid" >/dev/null 2>&1 || true
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
4. internal/store/store.go
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
//
// Every integer placeholder carries an explicit ::bigint cast. pgx v5 uses the
// extended query protocol and, in its default statement-cache mode, sends the
// Parse step without declared parameter types. Postgres then has to infer them
// and the expression `$4 + $5` fails at plan time with
// `ERROR: operator is not unique: unknown + unknown` — the statement can never
// even execute, so the sync endpoint 500s on every request. Casting each
// parameter pins the types and makes the statement plannable. (Placeholders
// compared against a typed column, like `$1::uuid = user_id`, would be
// inferable anyway, but we cast uniformly for clarity.)
func (s *Store) SyncCaptchaStats(ctx context.Context, userID string, in CaptchaSyncInput) (CaptchaStats, error) {
	st := CaptchaStats{UserID: userID}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO captcha_stats
			(user_id, best_streak, current_streak, total_solves, manual_solves, auto_solves, updated_at)
		VALUES ($1::uuid, $2::bigint, $3::bigint, $4::bigint + $5::bigint, $4::bigint, $5::bigint, now())
		ON CONFLICT (user_id) DO UPDATE SET
			best_streak    = GREATEST(captcha_stats.best_streak, EXCLUDED.best_streak),
			current_streak = EXCLUDED.current_streak,
			total_solves   = captcha_stats.total_solves + $4::bigint + $5::bigint,
			manual_solves  = captcha_stats.manual_solves + $4::bigint,
			auto_solves    = captcha_stats.auto_solves + $5::bigint,
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
5. internal/config/config.go
go
// Package config is the single source of truth for every runtime setting.
//
// Everything the application can be tuned with is read here, from environment
// variables, and nowhere else. No other package reads os.Getenv. This keeps
// the "all configuration is central in one location" promise from the design
// notes: to see everything GoTunnels can be configured with, read this file.
package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all resolved runtime settings for the API service.
type Config struct {
	// Identity -----------------------------------------------------------
	InstanceID  string // unique per running instance; also an OTel resource attribute
	ServiceName string // OTel service.name
	Version     string // build version (set via -ldflags at build time)

	// HTTP ---------------------------------------------------------------
	HTTPAddr        string        // listen address, e.g. ":8080"
	ShutdownTimeout time.Duration // graceful shutdown budget

	// Database -----------------------------------------------------------
	DatabaseURL      string
	DBMaxConns       int32
	DBMinConns       int32
	DBConnectTimeout time.Duration

	// Sessions -----------------------------------------------------------
	SessionCookieName string
	SessionTTL        time.Duration

	// Secrets (generated per-instance by scripts/up.sh, never committed) -
	ipHashPepper []byte   // used as sha256(pepper || ip)
	totpAESKey   [32]byte // AES-256-GCM key for encrypting TOTP secrets at rest

	// CORS ---------------------------------------------------------------
	// Exact allowed origins. Because the browser talks to the API cross-origin
	// (frontend tunnel URL != API tunnel URL) and sends credentials, we cannot
	// use the "*" wildcard together with credentials; we echo an allowed origin.
	CORSAllowedOrigins []string

	// WebAuthn / passkeys -----------------------------------------------
	RPID          string   // relying-party ID: the frontend's registrable domain
	RPDisplayName string   // human-readable name shown by the authenticator
	RPOrigins     []string // full origins the browser will present, e.g. https://x.trycloudflare.com

	// Content Security Policy -------------------------------------------
	// The header itself is emitted by Caddy on the frontend. These values let
	// the API report what mode it believes is active and are surfaced on the
	// health/info endpoint so the whole system agrees on one central policy.
	CSPMode   string // "report-only" or "enforce"
	CSPPolicy string // the policy string (informational mirror of Caddy's)

	// Telemetry ----------------------------------------------------------
	Telemetry TelemetryConfig

	// Dev conveniences ---------------------------------------------------
	Dev bool // when true, missing secrets are generated ephemerally with a warning
}

// TelemetryConfig captures OTLP/HTTP exporter settings. When Enabled is false
// the service logs to stdout only and installs no-op trace/metric providers.
type TelemetryConfig struct {
	Enabled     bool
	EndpointURL string            // full base URL, e.g. https://api.uptrace.dev
	Headers     map[string]string // e.g. {"uptrace-dsn": "..."}
	Insecure    bool              // allow http:// (in-cluster collectors)
	Compression string            // "gzip" or ""

	// MetricsTemporality mirrors OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE
	// ("delta", "cumulative", or "lowmemory"; normalized to lowercase). Uptrace
	// prefers delta temporality, so that is the default.
	MetricsTemporality string
	// MetricsHistogram mirrors OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION
	// ("base2_exponential_bucket_histogram" or "explicit_bucket_histogram";
	// normalized to lowercase). Exponential histograms compress better and give
	// Uptrace more accurate percentiles, so they are the default.
	MetricsHistogram string
}

// Load reads and validates configuration from the environment.
func Load() (*Config, error) {
	c := &Config{
		InstanceID:         getenv("GOTUNNELS_INSTANCE_ID", defaultInstanceID()),
		ServiceName:        getenv("OTEL_SERVICE_NAME", getenv("GOTUNNELS_SERVICE_NAME", "gotunnels-api")),
		Version:            getenv("GOTUNNELS_VERSION", "dev"),
		HTTPAddr:           getenv("GOTUNNELS_HTTP_ADDR", ":8080"),
		ShutdownTimeout:    getdur("GOTUNNELS_SHUTDOWN_TIMEOUT", 15*time.Second),
		DatabaseURL:        getenv("DATABASE_URL", ""),
		DBMaxConns:         int32(getint("GOTUNNELS_DB_MAX_CONNS", 20)),
		DBMinConns:         int32(getint("GOTUNNELS_DB_MIN_CONNS", 2)),
		DBConnectTimeout:   getdur("GOTUNNELS_DB_CONNECT_TIMEOUT", 30*time.Second),
		SessionCookieName:  getenv("GOTUNNELS_SESSION_COOKIE_NAME", "gotunnels_session"),
		SessionTTL:         getdur("GOTUNNELS_SESSION_TTL", 24*time.Hour),
		CORSAllowedOrigins: splitList(getenv("GOTUNNELS_CORS_ALLOWED_ORIGINS", "*")),
		RPID:               getenv("GOTUNNELS_RP_ID", "localhost"),
		RPDisplayName:      getenv("GOTUNNELS_RP_DISPLAY_NAME", "GoTunnels"),
		RPOrigins:          splitList(getenv("GOTUNNELS_RP_ORIGINS", "http://localhost:8080")),
		CSPMode:            getenv("GOTUNNELS_CSP_MODE", "report-only"),
		CSPPolicy:          getenv("GOTUNNELS_CSP_POLICY", DefaultCSPPolicy),
		Dev:                getbool("GOTUNNELS_DEV", false),
	}

	if err := c.resolveSecrets(); err != nil {
		return nil, err
	}
	c.resolveTelemetry()

	if err := c.Validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// DefaultCSPPolicy locks everything to same-origin: no third-party scripts,
// styles, images, fonts, media, or frames. Everything is self-hosted. It is
// emitted in report-only mode by default (see the Caddyfile), so violations
// are reported but nothing is blocked yet.
const DefaultCSPPolicy = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self'; " +
	"img-src 'self'; " +
	"font-src 'self'; " +
	"connect-src 'self' https:; " +
	"media-src 'self'; " +
	"object-src 'none'; " +
	"frame-src 'none'; " +
	"frame-ancestors 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'"

func (c *Config) resolveSecrets() error {
	// IP hashing pepper.
	if pepper := os.Getenv("GOTUNNELS_IP_HASH_PEPPER"); pepper != "" {
		c.ipHashPepper = []byte(pepper)
	} else if c.Dev {
		c.ipHashPepper = mustRandom(32)
	} else {
		return fmt.Errorf("config: GOTUNNELS_IP_HASH_PEPPER is required (set GOTUNNELS_DEV=1 to auto-generate for local dev)")
	}

	// TOTP encryption key. Accept any string; derive a fixed 32-byte AES key
	// via SHA-256 so operators can supply `openssl rand -base64 32` output
	// without worrying about exact byte length.
	if raw := os.Getenv("GOTUNNELS_TOTP_ENCRYPTION_KEY"); raw != "" {
		c.totpAESKey = sha256.Sum256([]byte(raw))
	} else if c.Dev {
		c.totpAESKey = sha256.Sum256(mustRandom(32))
	} else {
		return fmt.Errorf("config: GOTUNNELS_TOTP_ENCRYPTION_KEY is required (set GOTUNNELS_DEV=1 to auto-generate for local dev)")
	}
	return nil
}

func (c *Config) resolveTelemetry() {
	// Preference order:
	//   1. UPTRACE_DSN (convenience) -> derive endpoint + uptrace-dsn header
	//   2. OTEL_EXPORTER_OTLP_ENDPOINT (+ OTEL_EXPORTER_OTLP_HEADERS)
	//   3. disabled (stdout logging, no-op traces/metrics)
	//
	// The two metrics knobs follow the OpenTelemetry spec's environment
	// variables. Their spec values are UPPERCASE (e.g. DELTA,
	// BASE2_EXPONENTIAL_BUCKET_HISTOGRAM) but we normalize to lowercase so
	// comparisons elsewhere are simple and either casing works.
	tc := TelemetryConfig{
		Headers:     map[string]string{},
		Compression: strings.ToLower(getenv("OTEL_EXPORTER_OTLP_COMPRESSION", "gzip")),
		MetricsTemporality: strings.ToLower(getenv(
			"OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "delta")),
		MetricsHistogram: strings.ToLower(getenv(
			"OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION",
			"base2_exponential_bucket_histogram")),
	}

	if dsn := os.Getenv("UPTRACE_DSN"); dsn != "" {
		endpoint, insecure := endpointFromDSN(dsn)
		tc.Enabled = true
		tc.EndpointURL = endpoint
		tc.Insecure = insecure
		tc.Headers["uptrace-dsn"] = dsn
	} else if ep := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); ep != "" {
		tc.Enabled = true
		tc.EndpointURL = ep
		tc.Insecure = strings.HasPrefix(ep, "http://")
		for k, v := range parseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")) {
			tc.Headers[k] = v
		}
	}
	c.Telemetry = tc
}

// Validate checks invariants that must hold before the server starts.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: DATABASE_URL is required")
	}
	if c.HTTPAddr == "" {
		return fmt.Errorf("config: GOTUNNELS_HTTP_ADDR is required")
	}
	if c.SessionTTL <= 0 {
		return fmt.Errorf("config: GOTUNNELS_SESSION_TTL must be positive")
	}
	switch c.CSPMode {
	case "report-only", "enforce":
	default:
		return fmt.Errorf("config: GOTUNNELS_CSP_MODE must be 'report-only' or 'enforce', got %q", c.CSPMode)
	}
	if len(c.RPOrigins) == 0 {
		return fmt.Errorf("config: GOTUNNELS_RP_ORIGINS must contain at least one origin")
	}
	return nil
}

// IPHashPepper returns a copy-safe reference to the pepper bytes.
func (c *Config) IPHashPepper() []byte { return c.ipHashPepper }

// TOTPAESKey returns the 32-byte AES key for TOTP secret encryption.
func (c *Config) TOTPAESKey() [32]byte { return c.totpAESKey }

// CORSAllowsAny reports whether the wildcard "*" origin is configured.
func (c *Config) CORSAllowsAny() bool {
	for _, o := range c.CORSAllowedOrigins {
		if o == "*" {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// small env helpers
// ---------------------------------------------------------------------------

func getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getint(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return def
}

func getbool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
			return b
		}
	}
	return def
}

func getdur(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(strings.TrimSpace(v)); err == nil {
			return d
		}
	}
	return def
}

// splitList parses a comma-or-space separated list, trimming empties.
func splitList(s string) []string {
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' })
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

// parseHeaders parses OTEL_EXPORTER_OTLP_HEADERS ("k1=v1,k2=v2").
func parseHeaders(s string) map[string]string {
	out := map[string]string{}
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out
}

// endpointFromDSN extracts the OTLP base endpoint from an Uptrace DSN of the
// form "https://TOKEN@host[:port][/path]". Returns the base URL and whether it
// is insecure (http).
func endpointFromDSN(dsn string) (endpoint string, insecure bool) {
	// Very small, dependency-free parse: scheme://[userinfo@]host
	scheme := "https"
	rest := dsn
	if i := strings.Index(rest, "://"); i >= 0 {
		scheme = rest[:i]
		rest = rest[i+3:]
	}
	if at := strings.LastIndex(rest, "@"); at >= 0 {
		rest = rest[at+1:]
	}
	// strip any path/query
	if q := strings.IndexAny(rest, "/?"); q >= 0 {
		rest = rest[:q]
	}
	return scheme + "://" + rest, scheme == "http"
}

// defaultInstanceID returns a stable-ish identifier used when the operator did
// not supply one (mostly for bare `go run` during development).
func defaultInstanceID() string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return "local-" + hex.EncodeToString(mustRandom(4))
}

func mustRandom(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("config: crypto/rand failed: " + err.Error())
	}
	return b
}

// DecodeKeyMaterial is a helper exposed for tests and tooling: it accepts hex,
// base64 (std or url), or raw bytes and returns them. Unused by Load directly
// but kept for completeness where callers want exact-length keys.
func DecodeKeyMaterial(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if b, err := hex.DecodeString(s); err == nil && len(s)%2 == 0 {
		return b, nil
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return []byte(s), nil
}
6. internal/config/config_test.go
go
package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoad_RequiresDatabaseURL(t *testing.T) {
	t.Setenv("GOTUNNELS_DEV", "1") // auto-generate secrets
	// DATABASE_URL deliberately unset.
	if _, err := Load(); err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoad_RequiresSecretsWhenNotDev(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x/y")
	t.Setenv("GOTUNNELS_DEV", "0")
	// no pepper / totp key
	if _, err := Load(); err == nil {
		t.Fatal("expected error when secrets are missing and not in dev mode")
	}
}

func TestLoad_DevGeneratesSecrets(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_DEV", "1")
	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.IPHashPepper()) == 0 {
		t.Error("expected a generated IP hash pepper in dev mode")
	}
	var zero [32]byte
	if c.TOTPAESKey() == zero {
		t.Error("expected a non-zero TOTP AES key in dev mode")
	}
	if c.CSPMode != "report-only" {
		t.Errorf("expected default CSP mode report-only, got %q", c.CSPMode)
	}
	if !strings.Contains(c.CSPPolicy, "default-src 'self'") {
		t.Errorf("default CSP policy should lock to self, got %q", c.CSPPolicy)
	}
}

func TestLoad_TOTPKeyDerivedDeterministically(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/gotunnels")
	t.Setenv("GOTUNNELS_IP_HASH_PEPPER", "pepper")
	t.Setenv("GOTUNNELS_TOTP_ENCRYPTION_KEY", "the-same-secret")

	c1, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	c2, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c1.TOTPAESKey() != c2.TOTPAESKey() {
		t.Error("TOTP key derivation must be deterministic for a given secret")
	}
}

func TestValidate_CSPMode(t *testing.T) {
	c := &Config{
		DatabaseURL: "x",
		HTTPAddr:    ":8080",
		SessionTTL:  time.Hour,
		RPOrigins:   []string{"http://localhost"},
		CSPMode:     "nonsense",
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected invalid CSP mode to fail validation")
	}
	c.CSPMode = "enforce"
	if err := c.Validate(); err != nil {
		t.Fatalf("enforce should be valid: %v", err)
	}
}

func TestSplitList(t *testing.T) {
	cases := map[string][]string{
		"a,b,c":            {"a", "b", "c"},
		"a b c":            {"a", "b", "c"},
		" a , b ,, c ":     {"a", "b", "c"},
		"https://x https:": {"https://x", "https:"},
		"":                 {},
	}
	for in, want := range cases {
		got := splitList(in)
		if len(got) != len(want) {
			t.Errorf("splitList(%q) len = %d, want %d (%v)", in, len(got), len(want), got)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("splitList(%q)[%d] = %q, want %q", in, i, got[i], want[i])
			}
		}
	}
}

func TestParseHeaders(t *testing.T) {
	h := parseHeaders("uptrace-dsn=https://tok@api.uptrace.dev,x-extra = val ")
	if h["uptrace-dsn"] != "https://tok@api.uptrace.dev" {
		t.Errorf("unexpected dsn header: %q", h["uptrace-dsn"])
	}
	if h["x-extra"] != "val" {
		t.Errorf("unexpected x-extra header: %q", h["x-extra"])
	}
}

func TestEndpointFromDSN(t *testing.T) {
	cases := []struct {
		dsn          string
		wantEndpoint string
		wantInsecure bool
	}{
		{"https://TOKEN@api.uptrace.dev?grpc=4317", "https://api.uptrace.dev", false},
		{"https://TOKEN@api.uptrace.dev:443/v1", "https://api.uptrace.dev:443", false},
		{"http://token@localhost:14318", "http://localhost:14318", true},
	}
	for _, tc := range cases {
		ep, insecure := endpointFromDSN(tc.dsn)
		if ep != tc.wantEndpoint {
			t.Errorf("endpointFromDSN(%q) endpoint = %q, want %q", tc.dsn, ep, tc.wantEndpoint)
		}
		if insecure != tc.wantInsecure {
			t.Errorf("endpointFromDSN(%q) insecure = %v, want %v", tc.dsn, insecure, tc.wantInsecure)
		}
	}
}

func TestResolveTelemetry_DisabledByDefault(t *testing.T) {
	// Ensure a clean env for the OTLP vars.
	t.Setenv("UPTRACE_DSN", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	c := &Config{}
	c.resolveTelemetry()
	if c.Telemetry.Enabled {
		t.Error("telemetry should be disabled when no endpoint/DSN is configured")
	}
}

func TestResolveTelemetry_UptraceDSN(t *testing.T) {
	t.Setenv("UPTRACE_DSN", "https://secret@api.uptrace.dev?grpc=4317")
	c := &Config{}
	c.resolveTelemetry()
	if !c.Telemetry.Enabled {
		t.Fatal("telemetry should be enabled with an Uptrace DSN")
	}
	if c.Telemetry.EndpointURL != "https://api.uptrace.dev" {
		t.Errorf("endpoint = %q", c.Telemetry.EndpointURL)
	}
	if c.Telemetry.Headers["uptrace-dsn"] == "" {
		t.Error("expected uptrace-dsn header to be set")
	}
}

func TestResolveTelemetry_MetricsDefaults(t *testing.T) {
	// Uptrace-friendly defaults must hold even when the OTEL_* metrics vars
	// are entirely absent: delta temporality + base2 exponential histograms
	// + gzip compression.
	t.Setenv("UPTRACE_DSN", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_COMPRESSION", "")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION", "")
	c := &Config{}
	c.resolveTelemetry()
	if got := c.Telemetry.Compression; got != "gzip" {
		t.Errorf("default compression = %q, want gzip", got)
	}
	if got := c.Telemetry.MetricsTemporality; got != "delta" {
		t.Errorf("default temporality = %q, want delta", got)
	}
	if got := c.Telemetry.MetricsHistogram; got != "base2_exponential_bucket_histogram" {
		t.Errorf("default histogram aggregation = %q, want base2_exponential_bucket_histogram", got)
	}
}

func TestResolveTelemetry_MetricsOverrides(t *testing.T) {
	// The OTel spec documents UPPERCASE values (DELTA, CUMULATIVE,
	// BASE2_EXPONENTIAL_BUCKET_HISTOGRAM...); we must accept them and
	// normalize to lowercase.
	t.Setenv("UPTRACE_DSN", "https://secret@api.uptrace.dev")
	t.Setenv("OTEL_EXPORTER_OTLP_COMPRESSION", "NONE")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "CUMULATIVE")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION", "EXPLICIT_BUCKET_HISTOGRAM")
	c := &Config{}
	c.resolveTelemetry()
	if got := c.Telemetry.Compression; got != "none" {
		t.Errorf("compression = %q, want none", got)
	}
	if got := c.Telemetry.MetricsTemporality; got != "cumulative" {
		t.Errorf("temporality = %q, want cumulative", got)
	}
	if got := c.Telemetry.MetricsHistogram; got != "explicit_bucket_histogram" {
		t.Errorf("histogram aggregation = %q, want explicit_bucket_histogram", got)
	}
}

func TestGetHelpers(t *testing.T) {
	t.Setenv("X_INT", "42")
	if getint("X_INT", 0) != 42 {
		t.Error("getint failed")
	}
	if getint("X_MISSING", 7) != 7 {
		t.Error("getint default failed")
	}
	t.Setenv("X_BOOL", "true")
	if !getbool("X_BOOL", false) {
		t.Error("getbool failed")
	}
	t.Setenv("X_DUR", "250ms")
	if getdur("X_DUR", 0) != 250*time.Millisecond {
		t.Error("getdur failed")
	}
}
7. internal/telemetry/telemetry.go
go
// Package telemetry wires the vendor-neutral OpenTelemetry Go SDK for all three
// signals (traces, metrics, logs) and exports them over OTLP/HTTP.
//
// We deliberately never import any Uptrace SDK. Uptrace (or any OTLP backend)
// is configured purely via a DSN / endpoint, so swapping backends is a config
// change, not a code change.
//
// Logs are always written to stdout as JSON (so `podman logs` is useful) and,
// when an OTLP endpoint is configured, additionally shipped to the collector
// via the OpenTelemetry log bridge. When no endpoint is configured, traces and
// metrics use no-op providers and only stdout logging remains.
package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/kusl/GoTunnels/internal/config"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellog "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// scopeName is the instrumentation scope for logs/traces/metrics we emit
// directly (as opposed to via otelhttp instrumentation).
const scopeName = "github.com/kusl/GoTunnels"

// Providers bundles what the application needs after setup.
type Providers struct {
	Logger   *slog.Logger
	shutdown []func(context.Context) error
}

// Shutdown flushes and stops all providers. Safe to call once during teardown.
func (p *Providers) Shutdown(ctx context.Context) error {
	var errs []error
	for i := len(p.shutdown) - 1; i >= 0; i-- {
		if err := p.shutdown[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Setup initializes telemetry from configuration. The returned Providers always
// has a usable Logger, even on partial failure.
func Setup(ctx context.Context, cfg *config.Config) (*Providers, error) {
	p := &Providers{}

	// stdout logging is always on.
	stdoutHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	if !cfg.Telemetry.Enabled {
		p.Logger = slog.New(stdoutHandler).With(
			slog.String("service.name", cfg.ServiceName),
			slog.String("service.instance.id", cfg.InstanceID),
		)
		p.Logger.Info("telemetry disabled: logging to stdout only, no-op traces/metrics")
		return p, nil
	}

	res, err := buildResource(ctx, cfg)
	if err != nil {
		// A resource failure is not fatal; fall back to stdout logging.
		l := slog.New(stdoutHandler)
		l.Error("telemetry: failed to build resource; continuing with stdout logging", slog.Any("err", err))
		p.Logger = l
		return p, nil
	}

	// ---- Traces -------------------------------------------------------
	traceExp, err := otlptracehttp.New(ctx, traceHTTPOpts(cfg)...)
	if err != nil {
		return fallback(stdoutHandler, cfg, err), nil
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))
	p.shutdown = append(p.shutdown, tp.Shutdown)

	// ---- Metrics ------------------------------------------------------
	metricExp, err := otlpmetrichttp.New(ctx, metricHTTPOpts(cfg)...)
	if err != nil {
		return fallback(stdoutHandler, cfg, err), nil
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp,
			sdkmetric.WithInterval(30*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	p.shutdown = append(p.shutdown, mp.Shutdown)

	// ---- Logs ---------------------------------------------------------
	logExp, err := otlploghttp.New(ctx, logHTTPOpts(cfg)...)
	if err != nil {
		return fallback(stdoutHandler, cfg, err), nil
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)
	otellog.SetLoggerProvider(lp)
	p.shutdown = append(p.shutdown, lp.Shutdown)

	// Fan out slog to BOTH stdout and the OTel log bridge.
	otelHandler := otelslog.NewHandler(scopeName, otelslog.WithLoggerProvider(lp))
	p.Logger = slog.New(NewMultiHandler(stdoutHandler, otelHandler)).With(
		slog.String("service.name", cfg.ServiceName),
		slog.String("service.instance.id", cfg.InstanceID),
	)
	p.Logger.Info("telemetry enabled",
		slog.String("otlp.endpoint", cfg.Telemetry.EndpointURL),
		slog.Bool("otlp.insecure", cfg.Telemetry.Insecure),
		slog.String("otlp.compression", cfg.Telemetry.Compression),
		slog.String("otlp.metrics.temporality", cfg.Telemetry.MetricsTemporality),
		slog.String("otlp.metrics.histogram", cfg.Telemetry.MetricsHistogram),
	)
	return p, nil
}

func fallback(h slog.Handler, cfg *config.Config, cause error) *Providers {
	l := slog.New(h).With(slog.String("service.name", cfg.ServiceName))
	l.Error("telemetry: exporter setup failed; continuing with stdout logging", slog.Any("err", cause))
	return &Providers{Logger: l}
}

func buildResource(ctx context.Context, cfg *config.Config) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.Version),
			semconv.ServiceInstanceID(cfg.InstanceID),
		),
	)
}

func traceHTTPOpts(cfg *config.Config) []otlptracehttp.Option {
	opts := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(cfg.Telemetry.EndpointURL)}
	if cfg.Telemetry.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(cfg.Telemetry.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.Telemetry.Headers))
	}
	if cfg.Telemetry.Compression == "gzip" {
		opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.GzipCompression))
	}
	return opts
}

func metricHTTPOpts(cfg *config.Config) []otlpmetrichttp.Option {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpointURL(cfg.Telemetry.EndpointURL),
		otlpmetrichttp.WithTemporalitySelector(temporalitySelector(cfg.Telemetry.MetricsTemporality)),
		otlpmetrichttp.WithAggregationSelector(aggregationSelector(cfg.Telemetry.MetricsHistogram)),
	}
	if cfg.Telemetry.Insecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	if len(cfg.Telemetry.Headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(cfg.Telemetry.Headers))
	}
	if cfg.Telemetry.Compression == "gzip" {
		opts = append(opts, otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression))
	}
	return opts
}

// temporalitySelector maps OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE
// (already lowercased by config) onto an SDK TemporalitySelector, following
// the OTLP exporter spec:
//
//	delta:      Counter, ObservableCounter, Histogram -> delta; others cumulative
//	lowmemory:  Counter, Histogram -> delta; others cumulative
//	cumulative: everything cumulative (the SDK default)
//
// Uptrace prefers delta, which is our config default. Unknown values fall back
// to the SDK default (cumulative) rather than failing startup.
func temporalitySelector(pref string) sdkmetric.TemporalitySelector {
	switch strings.ToLower(pref) {
	case "delta":
		return func(k sdkmetric.InstrumentKind) metricdata.Temporality {
			switch k {
			case sdkmetric.InstrumentKindCounter,
				sdkmetric.InstrumentKindObservableCounter,
				sdkmetric.InstrumentKindHistogram:
				return metricdata.DeltaTemporality
			default:
				return metricdata.CumulativeTemporality
			}
		}
	case "lowmemory":
		return func(k sdkmetric.InstrumentKind) metricdata.Temporality {
			switch k {
			case sdkmetric.InstrumentKindCounter,
				sdkmetric.InstrumentKindHistogram:
				return metricdata.DeltaTemporality
			default:
				return metricdata.CumulativeTemporality
			}
		}
	default: // "cumulative" or anything unrecognized
		return sdkmetric.DefaultTemporalitySelector
	}
}

// aggregationSelector maps OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION
// (already lowercased by config) onto an SDK AggregationSelector. With
// base2_exponential_bucket_histogram (our default, and what Uptrace
// recommends), Histogram instruments use auto-scaling exponential buckets
// (MaxSize 160 / MaxScale 20, the SDK's own defaults for this aggregation)
// instead of fixed explicit boundaries; every other instrument kind keeps the
// SDK default aggregation.
func aggregationSelector(hist string) sdkmetric.AggregationSelector {
	if strings.ToLower(hist) != "base2_exponential_bucket_histogram" {
		return sdkmetric.DefaultAggregationSelector
	}
	return func(k sdkmetric.InstrumentKind) sdkmetric.Aggregation {
		if k == sdkmetric.InstrumentKindHistogram {
			return sdkmetric.AggregationBase2ExponentialHistogram{
				MaxSize:  160,
				MaxScale: 20,
			}
		}
		return sdkmetric.DefaultAggregationSelector(k)
	}
}

func logHTTPOpts(cfg *config.Config) []otlploghttp.Option {
	opts := []otlploghttp.Option{otlploghttp.WithEndpointURL(cfg.Telemetry.EndpointURL)}
	if cfg.Telemetry.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}
	if len(cfg.Telemetry.Headers) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(cfg.Telemetry.Headers))
	}
	if cfg.Telemetry.Compression == "gzip" {
		opts = append(opts, otlploghttp.WithCompression(otlploghttp.GzipCompression))
	}
	return opts
}
8. compose.yaml
yaml
# compose.yaml — the whole GoTunnels stack.
#
# Design rules embodied here:
#   * No service publishes a host port. The only traffic leaving the machine is
#     the two outbound Cloudflare Quick Tunnel connections.
#   * Everything is addressed by service name on the internal Compose network.
#   * Every instance is isolated by the Compose project name (passed via -p),
#     which prefixes container, network, and volume names — so N instances run
#     side by side without collision.
#   * Values that can be generated are generated by scripts/up.sh into .env;
#     nothing secret is hardcoded here.
#
# Runtime-discovered values (the frontend's tunnel URL, used for the WebAuthn
# RP ID and CORS allow-list) are exported into the shell by scripts/up.sh
# before the api service is (re)started, and read here via ${VAR} substitution.

services:
  db:
    image: docker.io/library/postgres:16-alpine
    environment:
      POSTGRES_USER: "${POSTGRES_USER:-gotunnels}"
      POSTGRES_PASSWORD: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD must be set (run scripts/up.sh)}"
      POSTGRES_DB: "${POSTGRES_DB:-gotunnels}"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-gotunnels} -d ${POSTGRES_DB:-gotunnels}"]
      interval: 3s
      timeout: 3s
      retries: 20
    restart: unless-stopped

  api:
    build:
      context: .
      dockerfile: Containerfile.api
      args:
        VERSION: "${GOTUNNELS_VERSION:-dev}"
    depends_on:
      db:
        condition: service_healthy
    environment:
      GOTUNNELS_INSTANCE_ID: "${GOTUNNELS_INSTANCE_ID:-default}"
      GOTUNNELS_VERSION: "${GOTUNNELS_VERSION:-dev}"
      GOTUNNELS_HTTP_ADDR: ":8080"
      DATABASE_URL: "postgres://${POSTGRES_USER:-gotunnels}:${POSTGRES_PASSWORD}@db:5432/${POSTGRES_DB:-gotunnels}?sslmode=disable"
      GOTUNNELS_IP_HASH_PEPPER: "${GOTUNNELS_IP_HASH_PEPPER:?set by scripts/up.sh}"
      GOTUNNELS_TOTP_ENCRYPTION_KEY: "${GOTUNNELS_TOTP_ENCRYPTION_KEY:?set by scripts/up.sh}"
      # Discovered at runtime by scripts/up.sh (frontend tunnel URL). Defaults
      # keep `compose config` valid before discovery.
      GOTUNNELS_RP_ID: "${GOTUNNELS_RP_ID:-localhost}"
      GOTUNNELS_RP_DISPLAY_NAME: "${GOTUNNELS_RP_DISPLAY_NAME:-GoTunnels}"
      GOTUNNELS_RP_ORIGINS: "${GOTUNNELS_RP_ORIGINS:-http://localhost:8080}"
      GOTUNNELS_CORS_ALLOWED_ORIGINS: "${GOTUNNELS_CORS_ALLOWED_ORIGINS:-*}"
      GOTUNNELS_CSP_MODE: "${GOTUNNELS_CSP_MODE:-report-only}"
      # Telemetry: a single Uptrace (cloud or self-hosted) DSN, or OTLP env vars.
      UPTRACE_DSN: "${UPTRACE_DSN:-}"
      OTEL_EXPORTER_OTLP_ENDPOINT: "${OTEL_EXPORTER_OTLP_ENDPOINT:-}"
      OTEL_EXPORTER_OTLP_HEADERS: "${OTEL_EXPORTER_OTLP_HEADERS:-}"
      OTEL_EXPORTER_OTLP_COMPRESSION: "${OTEL_EXPORTER_OTLP_COMPRESSION:-gzip}"
      OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE: "${OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE:-delta}"
      OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION: "${OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION:-base2_exponential_bucket_histogram}"
      OTEL_SERVICE_NAME: "${OTEL_SERVICE_NAME:-gotunnels-api}"
      GOTUNNELS_DEV: "${GOTUNNELS_DEV:-false}"
    restart: unless-stopped

  frontend:
    build:
      context: .
      dockerfile: Containerfile.frontend
    depends_on:
      - api
    environment:
      GOTUNNELS_CSP_HEADER_NAME: "${GOTUNNELS_CSP_HEADER_NAME:-Content-Security-Policy-Report-Only}"
      GOTUNNELS_CSP_POLICY: "${GOTUNNELS_CSP_POLICY:-default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self' https:; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'}"
    restart: unless-stopped

  cloudflared-frontend:
    image: docker.io/cloudflare/cloudflared:latest
    depends_on:
      - frontend
    command: ["tunnel", "--no-autoupdate", "--url", "http://frontend:8080"]
    restart: unless-stopped

  cloudflared-api:
    image: docker.io/cloudflare/cloudflared:latest
    depends_on:
      - api
    command: ["tunnel", "--no-autoupdate", "--url", "http://api:8080"]
    restart: unless-stopped

volumes:
  pgdata:
9. .github/workflows/ci.yaml
yaml
# Build and test. Deliberately thin: all logic lives in scripts/test.sh so the
# exact same commands run locally and here.
name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

permissions:
  contents: read

concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7

      - uses: actions/setup-go@v6
        with:
          go-version: "stable"
          # go.sum is intentionally absent on first import; do not fail on it.
          cache: false

      - name: Build, vet, unit tests
        run: bash scripts/test.sh all

  # Everything scripts/test.sh cannot see: that both Containerfiles build,
  # that compose wiring + env substitution is correct, that migrations run
  # against a real Postgres, and that the HTTP surface works end to end —
  # including the captcha-sync regression that only reproduces against real
  # Postgres via the pgx extended protocol. Ubuntu runners ship podman;
  # podman-compose comes from pipx. No tunnels are started and no host ports
  # are published; assertions run from a curl container on the compose network.
  containers:
    runs-on: ubuntu-latest
    timeout-minutes: 25
    steps:
      - uses: actions/checkout@v7

      - name: Install podman-compose
        run: pipx install podman-compose

      - name: Runtime versions
        run: |
          podman version
          podman-compose version

      - name: Build images + container smoke test
        run: bash scripts/ci-container-test.sh gotunnels-ci
10. .env.example
bash
# .env.example — shape of the environment file.
#
# There are NO real secrets to copy here. scripts/up.sh generates a .env with
# fresh random secrets on first run if one does not already exist. This file
# exists only to document what the stack reads.
#
# Copy to .env and fill in ONLY if you are not using scripts/up.sh:
#   cp .env.example .env

# --- Instance identity -------------------------------------------------------
# Unique per running instance; prefixes containers/network/volumes and appears
# as an OpenTelemetry resource attribute.
GOTUNNELS_INSTANCE_ID=default
GOTUNNELS_VERSION=dev

# --- Database (password is generated by up.sh) -------------------------------
POSTGRES_USER=gotunnels
POSTGRES_DB=gotunnels
POSTGRES_PASSWORD=generate-me

# --- Server secrets (generated by up.sh; never commit real values) -----------
# Pepper mixed into sha256(pepper||ip) so stored IP hashes are not rainbow-tableable.
GOTUNNELS_IP_HASH_PEPPER=generate-me
# Any string; the app derives a 32-byte AES key from it to encrypt TOTP secrets.
GOTUNNELS_TOTP_ENCRYPTION_KEY=generate-me

# --- WebAuthn / CORS (discovered at runtime from the frontend tunnel URL) ----
# up.sh overwrites these once the frontend Quick Tunnel URL is known.
GOTUNNELS_RP_ID=localhost
GOTUNNELS_RP_DISPLAY_NAME=GoTunnels
GOTUNNELS_RP_ORIGINS=http://localhost:8080
GOTUNNELS_CORS_ALLOWED_ORIGINS=*

# --- Content Security Policy (central configuration) -------------------------
# Header name selects report-only vs enforcing:
#   Content-Security-Policy-Report-Only  (default; observe only)
#   Content-Security-Policy              (enforce)
GOTUNNELS_CSP_HEADER_NAME=Content-Security-Policy-Report-Only
GOTUNNELS_CSP_MODE=report-only
GOTUNNELS_CSP_POLICY="default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self' https:; object-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"

# --- Telemetry (optional) ----------------------------------------------------
# Point at Uptrace Cloud (or any OTLP/HTTP backend) with a single DSN:
#   UPTRACE_DSN=https://<token>@api.uptrace.dev
# Or use the standard OTLP env vars instead:
#   OTEL_EXPORTER_OTLP_ENDPOINT=https://otlp.example.com
#   OTEL_EXPORTER_OTLP_HEADERS=key1=value1,key2=value2
# Leave blank to disable exporting (the app still logs JSON to stdout).
#
# NOTE: exporting UPTRACE_DSN in your shell before running scripts/up.sh also
# works — ensure_env persists it into this file before loading it, so a blank
# UPTRACE_DSN= line here can no longer clobber your export.
UPTRACE_DSN=
OTEL_EXPORTER_OTLP_ENDPOINT=
OTEL_EXPORTER_OTLP_HEADERS=
# gzip-compress OTLP payloads (Uptrace recommendation; set empty to disable).
OTEL_EXPORTER_OTLP_COMPRESSION=gzip
# Metrics temporality: delta | cumulative | lowmemory (spec casing DELTA etc.
# is also accepted). Uptrace prefers delta.
OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta
# Histogram aggregation for Histogram instruments:
#   base2_exponential_bucket_histogram (Uptrace recommendation; default)
#   explicit_bucket_histogram
OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION=base2_exponential_bucket_histogram
OTEL_SERVICE_NAME=gotunnels-api

# --- Dev convenience ---------------------------------------------------------
# When true, missing secrets are auto-generated ephemerally and cookies are not
# marked Secure. Never set true in anything resembling production.
GOTUNNELS_DEV=false
11. docs/CONFIGURATION.md
markdown
# Configuration

All runtime configuration is centralized. On the Go side every variable is read
in exactly one place — [`internal/config/config.go`](../internal/config/config.go)
— and nowhere else. This file is the reference for what each variable does.

Most values are generated for you by `scripts/up.sh` on first run and written to
a gitignored `.env`. You only ever set variables by hand if you are bypassing
the scripts. See also [`.env.example`](../.env.example).

## Identity

|
 Variable 
|
 Default 
|
 Meaning 
|
|
---
|
---
|
---
|
|
`GOTUNNELS_INSTANCE_ID`
|
 hostname / 
`default`
|
 Unique per instance; also an OTel resource attribute and used to derive the Compose project name. 
|
|
`GOTUNNELS_VERSION`
|
`dev`
|
 Build/version label (also settable at build time via 
`-ldflags -X main.version`
). 
|

## HTTP

|
 Variable 
|
 Default 
|
 Meaning 
|
|
---
|
---
|
---
|
|
`GOTUNNELS_HTTP_ADDR`
|
`:8080`
|
 Listen address inside the container. 
|
|
`GOTUNNELS_SHUTDOWN_TIMEOUT`
|
`15s`
|
 Graceful shutdown budget. 
|

## Database

|
 Variable 
|
 Default 
|
 Meaning 
|
|
---
|
---
|
---
|
|
`DATABASE_URL`
|
 (required) 
|
 pgx connection string. Compose builds it from the 
`POSTGRES_*`
 values. 
|
|
`POSTGRES_USER`
|
`gotunnels`
|
 DB user (Compose/Postgres). 
|
|
`POSTGRES_DB`
|
`gotunnels`
|
 DB name. 
|
|
`POSTGRES_PASSWORD`
|
 generated 
|
 DB password; generated by 
`up.sh`
. 
|
|
`GOTUNNELS_DB_MAX_CONNS`
|
`20`
|
 Pool max connections. 
|
|
`GOTUNNELS_DB_MIN_CONNS`
|
`2`
|
 Pool min connections. 
|
|
`GOTUNNELS_DB_CONNECT_TIMEOUT`
|
`30s`
|
 How long startup retries connecting before giving up. 
|

## Sessions

|
 Variable 
|
 Default 
|
 Meaning 
|
|
---
|
---
|
---
|
|
`GOTUNNELS_SESSION_COOKIE_NAME`
|
`gotunnels_session`
|
 Name of the secondary session cookie. 
|
|
`GOTUNNELS_SESSION_TTL`
|
`24h`
|
 Session lifetime. 
|

## Secrets (generated per instance; never commit real values)

|
 Variable 
|
 Default 
|
 Meaning 
|
|
---
|
---
|
---
|
|
`GOTUNNELS_IP_HASH_PEPPER`
|
 required (unless 
`GOTUNNELS_DEV`
) 
|
 Mixed into 
`sha256(pepper||ip)`
 so stored IP hashes are not rainbow-tableable. 
|
|
`GOTUNNELS_TOTP_ENCRYPTION_KEY`
|
 required (unless 
`GOTUNNELS_DEV`
) 
|
 Any string; a 32-byte AES-256-GCM key is derived from it (SHA-256) to encrypt TOTP secrets at rest. 
|

When `GOTUNNELS_DEV=true`, missing secrets are generated ephemerally at startup
with a warning (fine for local hacking, never for anything real).

## WebAuthn / CORS (discovered at runtime)

`scripts/up.sh` overwrites these once the frontend Quick Tunnel URL is known.

|
 Variable 
|
 Default 
|
 Meaning 
|
|
---
|
---
|
---
|
|
`GOTUNNELS_RP_ID`
|
`localhost`
|
 WebAuthn Relying Party ID = the frontend's registrable domain. 
|
|
`GOTUNNELS_RP_DISPLAY_NAME`
|
`GoTunnels`
|
 Human-readable RP name (also the TOTP issuer). 
|
|
`GOTUNNELS_RP_ORIGINS`
|
`http://localhost:8080`
|
 Full origin(s) the browser presents (comma/space separated). 
|
|
`GOTUNNELS_CORS_ALLOWED_ORIGINS`
|
`*`
|
 Exact allowed origin(s), or 
`*`
. Because credentials are used, 
`*`
 is echoed as the request origin rather than literally, and never treated as authorization. 
|

## Content Security Policy (central)

|
 Variable 
|
 Default 
|
 Meaning 
|
|
---
|
---
|
---
|
|
`GOTUNNELS_CSP_HEADER_NAME`
|
`Content-Security-Policy-Report-Only`
|
 Set to 
`Content-Security-Policy`
 to 
**
enforce
**
. 
|
|
`GOTUNNELS_CSP_MODE`
|
`report-only`
|
 Informational mirror surfaced on 
`/api/info`
 (
`report-only`
 or 
`enforce`
). 
|
|
`GOTUNNELS_CSP_POLICY`
|
 strict self-only (see below) 
|
 The policy string. Emitted by Caddy; mirrored to the API for 
`/api/info`
. 
|

The default policy is fully self-hosted — no third-party scripts, styles,
images, fonts, media, or frames:

```
default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self';
font-src 'self'; connect-src 'self' https:; media-src 'self'; object-src 'none';
frame-src 'none'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'
```

Because the app has no inline scripts, inline event handlers, or inline styles,
it already satisfies the enforcing form of this policy.

## Telemetry (optional)

Resolution order: `UPTRACE_DSN` first, then the standard OTLP variables, then
disabled (stdout logging only).

|
 Variable 
|
 Default 
|
 Meaning 
|
|
---
|
---
|
---
|
|
`UPTRACE_DSN`
|
 (empty) 
|
 e.g. 
`https://<token>@api.uptrace.dev`
. Derives the OTLP endpoint and sets the 
`uptrace-dsn`
 header. 
|
|
`OTEL_EXPORTER_OTLP_ENDPOINT`
|
 (empty) 
|
 Base OTLP/HTTP endpoint, if not using a DSN. 
|
|
`OTEL_EXPORTER_OTLP_HEADERS`
|
 (empty) 
|
`k1=v1,k2=v2`
 headers for the exporter. 
|
|
`OTEL_EXPORTER_OTLP_COMPRESSION`
|
`gzip`
|
 Exporter compression. 
|
|
`OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE`
|
`delta`
|
`delta`
, 
`cumulative`
, or 
`lowmemory`
 (spec casing like 
`DELTA`
 also accepted). Uptrace prefers delta. 
|
|
`OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION`
|
`base2_exponential_bucket_histogram`
|
 Or 
`explicit_bucket_histogram`
. Exponential buckets give Uptrace accurate percentiles at lower cost. 
|
|
`OTEL_SERVICE_NAME`
 / 
`GOTUNNELS_SERVICE_NAME`
|
`gotunnels-api`
|
 OTel 
`service.name`
. 
|

## Dev

|
 Variable 
|
 Default 
|
 Meaning 
|
|
---
|
---
|
---
|
|
`GOTUNNELS_DEV`
|
`false`
|
 Auto-generate missing secrets; mark cookies non-Secure. Local use only. 
|

## Script-only variables

These are read by the shell scripts (not the Go app) to control instance
naming and the test toolchain:

|
 Variable 
|
 Meaning 
|
|
---
|
---
|
|
`GOTUNNELS_PROJECT`
|
 Explicit Compose project name (else derived from the instance id or generated). 
|
|
`GO_IMAGE`
|
 Go image used by 
`scripts/test.sh`
 when no host Go toolchain is present (default 
`golang:1.26-bookworm`
). 
|
|
`CR`
 / 
`COMPOSE`
|
 Override the detected container runtime / compose command. 
|
That's all 11 files, complete: 3 shell scripts, 4 Go files, 2 YAML files, .env.example, and the docs page. scripts/ci-container-test.sh is new and needs chmod +x if your git setup doesn't preserve the executable bit from a plain paste. Redeploy is just git pull && bash scripts/up.sh (no .env deletion needed anymore), and remember to rotate the GitHub PAT and the Uptrace DSN — the DSN also sits in plaintext in virginia.txt.


Your org is out of usage credits. We let your admin know. Switch models to continue chatting.
Switch models





Claude is AI and can make mistakes. Please double-check responses.




100
21
39
