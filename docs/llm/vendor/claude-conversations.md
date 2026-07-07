00
44
76

Something isn't quite right here. Mozilla Firefox reports that the CSP is not set up right. 
Please review this and any other issue 
and please give me full file and full file path for all files that need to change 
also please make sure we have good test coverage as well
Content-Security-Policy: This site (https://insertion-revision-valuable-separated.trycloudflare.com) has a Report-Only policy without a report-uri directive nor a report-to directive. CSP will not block and cannot report violations of this policy.
also the containers action is failing on github 
Run bash scripts/ci-container-test.sh gotunnels-ci
[gotunnels] using runtime: podman / compose: podman compose
[gotunnels] generating fresh .env with per-instance secrets
[gotunnels] wrote /home/runner/work/GoTunnels/GoTunnels/.env (gitignored)
[gotunnels] reset tunnel-derived env (RP_ID / RP_ORIGINS / CORS) to bootstrap defaults
[gotunnels] container test instance: gotunnels-ci
[gotunnels] building images (Containerfile.api + Containerfile.frontend)…
>>>> Executing external compose provider "/usr/libexec/docker/cli-plugins/docker-compose". Please refer to the documentation for details. <<<<

time="2026-07-06T07:58:42Z" level=warning msg="Docker Compose is configured to build using Bake, but buildkit isn't enabled"
 Service frontend  Building
 Service api  Building
time="2026-07-06T07:58:42Z" level=error msg="Can't add file /home/runner/work/GoTunnels/GoTunnels/.containerignore to tar: io: read/write on closed pipe"
time="2026-07-06T07:58:42Z" level=error msg="Can't close tar writer: io: read/write on closed pipe"
time="2026-07-06T07:58:42Z" level=error msg="Can't add file /home/runner/work/GoTunnels/GoTunnels/.containerignore to tar: io: read/write on closed pipe"
time="2026-07-06T07:58:42Z" level=error msg="Can't close tar writer: io: read/write on closed pipe"
Cannot connect to the Docker daemon at unix:///run/user/1001/podman/podman.sock. Is the docker daemon running?
Error: executing /usr/libexec/docker/cli-plugins/docker-compose -f /home/runner/work/GoTunnels/GoTunnels/compose.yaml -p gotunnels-ci build: exit status 1
[gotunnels] container test FAILED (exit 1) — dumping service logs
[gotunnels] ----- logs: db ----- (no container)
[gotunnels] ----- logs: api ----- (no container)
[gotunnels] ----- logs: frontend ----- (no container)
[gotunnels] tearing down CI instance gotunnels-ci
Error: Process completed with exit code 1.
