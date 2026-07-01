98
10

What went wrong now? 

kushal@fedora:~/src/golang/GoTunnels$ cd ~/src/golang/GoTunnels/; export UPTRACE_DSN="https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"; time bash scripts/up.sh
[gotunnels] using runtime: podman / compose: podman compose
[gotunnels] using existing /home/kushal/src/golang/GoTunnels/.env
[gotunnels] project (instance): gotunnels-dt6vjg
[gotunnels] building images…
>>>> Executing external compose provider "/usr/bin/podman-compose". Please see podman-compose(1) for how to disable this message. <<<<

STEP 1/5: FROM docker.io/library/caddy:2-alpine
[1/2] STEP 1/6: FROM golang:1.26-bookworm AS build
[snip]
[2/2] STEP 6/6: ENTRYPOINT ["/api"]
--> Using cache 5c2fd56b4fade2cca3d1ea47b115832735edd6f028a65cbee2ffa844721fc752
[2/2] COMMIT gotunnels-dt6vjg_api
--> 5c2fd56b4fad
Successfully tagged localhost/gotunnels-dt6vjg_api:latest
Successfully tagged localhost/gotunnels-qdiyew_api:latest
Successfully tagged localhost/gotunnels-fifbya_api:latest
Successfully tagged localhost/gotunnels-eclukq_api:latest
5c2fd56b4fade2cca3d1ea47b115832735edd6f028a65cbee2ffa844721fc752
[gotunnels] starting database…
>>>> Executing external compose provider "/usr/bin/podman-compose". Please see podman-compose(1) for how to disable this message. <<<<

dac78f1502db8ab818a31d9f4687af833f8a631460619447aa2a891677acb14a
5a9f8a0ebb5d5bad7fd029e69584c6ddc4259291bef0d1593a3adaf55f028002
gotunnels-dt6vjg_db_1
[gotunnels] waiting for 'db' to become healthy (up to 120s)
[gotunnels] 'db' did not become healthy in 120s

real	2m30.831s
user	1m17.622s
sys	0m21.898s
kushal@fedora:~/src/golang/GoTunnels$ 
