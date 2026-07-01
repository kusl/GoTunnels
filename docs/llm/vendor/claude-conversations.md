78
08

What went wrong now? 

kushal@fedora:~/src/golang/GoTunnels$ cd ~/src/golang/GoTunnels/; export UPTRACE_DSN="https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"; time bash export.sh > docs/llm/vendor/output.txt

real	0m2.295s
user	0m1.141s
sys	0m1.455s
kushal@fedora:~/src/golang/GoTunnels$ cd ~/src/golang/GoTunnels/; export UPTRACE_DSN="https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"; time bash export.sh > docs/llm/vendor/output.txt

real	0m2.313s
user	0m1.153s
sys	0m1.446s
kushal@fedora:~/src/golang/GoTunnels$ cd ~/src/golang/GoTunnels/; export UPTRACE_DSN="https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"; time bash scripts/up.sh
[gotunnels] using runtime: podman / compose: podman compose
[gotunnels] using existing /home/kushal/src/golang/GoTunnels/.env
[gotunnels] project (instance): gotunnels-eclukq
[gotunnels] building images…
>>>> Executing external compose provider "/usr/bin/podman-compose". Please see podman-compose(1) for how to disable this message. <<<<

STEP 1/5: FROM docker.io/library/caddy:2-alpine
[1/2] STEP 1/6: FROM golang:1.26-bookworm AS build
Trying to pull docker.io/library/caddy:2-alpine...
[1/2] STEP 2/6: WORKDIR /src
--> Using cache babdaca5bcc2fececbbc19b62b0bfc9eaedfc787aa587199af487dc7ea867421
--> babdaca5bcc2
[1/2] STEP 3/6: ENV CGO_ENABLED=0     GOFLAGS=-mod=mod     GOTOOLCHAIN=local
--> Using cache 2048cbee3e7520d1cefe0c1d68b41d63b22a875582bb19fed499dd483de336d7
--> 2048cbee3e75
[1/2] STEP 4/6: COPY . .
Getting image source signatures
Copying blob 4f4fb700ef54 skipped: already exists  
Copying blob ee31d5a470f0 done   | 
Copying blob f8432a27d075 [--------------------------------------] 0.0b / 2.6MiB
Copying blob f8432a27d075 [=====>--------------------------------] 447.4KiB / 2.6MiB | 7.2 MiB/s
Copying blob 4f4fb700ef54 skipped: already exists  
Copying blob ee31d5a470f0 done   | 
Copying blob f8432a27d075 [--------------------------------------] 0.0b / 2.6MiB
Copying blob f8432a27d075 [=====================================>] 2.6MiB / 2.6MiB | 7.8 MiB/s
Copying blob e6f31ffc071e [--------------------------------------] 0.0b / 3.7MiB
Copying blob e6f31ffc071e [=================>--------------------] 1.7MiB / 3.7MiB | 28.1 MiB/s
Copying blob a0449c657909 [--------------------------------------] 0.0b / 16.5MiB
Copying blob a0449c657909 [==>-----------------------------------] 1.1MiB / 16.5MiB | 8.9 MiB/s
go: downloading github.com/go-webauthn/webauthn v0.11.2
go: downloading go.opentelemetry.io/contrib/bridges/otelslog v0.19.0
go: downloading go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0
go: downloading github.com/jackc/pgx/v5 v5.10.0
go: downloading github.com/pquerna/otp v1.4.0
Copying blob 4f4fb700ef54 skipped: already exists  
Copying blob ee31d5a470f0 done   | 
Copying blob f8432a27d075 done   | 
Copying blob e6f31ffc071e done   | 
Copying blob a0449c657909 done   | 
Copying config af555904a0 done   | 
Writing manifest to image destination
STEP 2/5: COPY frontend/ /srv/
go: downloading go.opentelemetry.io/otel/sdk v1.44.0
--> 8de81c4f823c
STEP 3/5: COPY frontend/Caddyfile /etc/caddy/Caddyfile
go: downloading go.opentelemetry.io/otel/metric v1.44.0
go: downloading go.opentelemetry.io/otel/trace v1.44.0
--> 1dcdf59f71ba
STEP 4/5: RUN rm -f /srv/Caddyfile
go: downloading github.com/jackc/puddle/v2 v2.2.2
go: downloading github.com/stretchr/testify v1.11.1
go: downloading github.com/felixge/httpsnoop v1.0.4
go: downloading golang.org/x/sys v0.46.0
go: downloading github.com/jackc/pgpassfile v1.0.0
go: downloading github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
go: downloading golang.org/x/text v0.38.0
go: downloading google.golang.org/protobuf v1.36.11
go: downloading go.opentelemetry.io/proto/otlp v1.10.0
go: downloading github.com/google/go-cmp v0.7.0
--> e170163757ff
STEP 5/5: EXPOSE 8080
go: downloading github.com/go-logr/logr v1.4.3
COMMIT gotunnels-eclukq_frontend
go: downloading github.com/go-logr/stdr v1.2.2
go: downloading go.opentelemetry.io/otel/metric/x v0.66.0
--> 1d92b326891d
Successfully tagged localhost/gotunnels-eclukq_frontend:latest
1d92b326891da32bb747a4dd759c2d8fc6e4bc622f724fb076c00bc30f0cd0dd
go: downloading github.com/google/uuid v1.6.0
go: downloading go.uber.org/goleak v1.3.0
go: downloading go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.44.0
go: downloading github.com/cenkalti/backoff/v5 v5.0.3
go: downloading go.opentelemetry.io/otel/sdk/log/logtest v0.20.0
go: downloading go.opentelemetry.io/auto/sdk v1.2.1
go: downloading github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc
go: downloading google.golang.org/grpc v1.81.1
go: downloading golang.org/x/sync v0.17.0
go: downloading github.com/davecgh/go-spew v1.1.1
go: downloading github.com/pmezard/go-difflib v1.0.0
go: downloading github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0
go: downloading github.com/cespare/xxhash/v2 v2.3.0
go: downloading github.com/go-webauthn/x v0.1.14
go: downloading github.com/golang-jwt/jwt/v5 v5.2.2
go: downloading github.com/mitchellh/mapstructure v1.5.0
go: downloading github.com/google/go-tpm v0.9.1
go: downloading gopkg.in/yaml.v3 v3.0.1
go: downloading github.com/fxamacker/cbor/v2 v2.7.0
go: downloading google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa
go: downloading google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa
go: downloading golang.org/x/net v0.55.0
go: downloading github.com/x448/float16 v0.8.4
go: downloading golang.org/x/sync v0.21.0
go: downloading github.com/golang/protobuf v1.5.4
go: downloading gonum.org/v1/gonum v0.17.0
--> b89a512de4c9
[2/2] STEP 1/6: FROM gcr.io/distroless/static-debian12:nonroot
[2/2] STEP 2/6: WORKDIR /
--> Using cache ea9af167ae5e058f672df8dadcc271a9bf0f3c9cde9671a600e8616c9b0b230b
--> ea9af167ae5e
[2/2] STEP 3/6: COPY --from=build /out/api /api
--> 8a9f1e128ef6
[2/2] STEP 4/6: USER nonroot:nonroot
--> 633c109c36ef
[2/2] STEP 5/6: EXPOSE 8080
--> 0f87d03655e6
[2/2] STEP 6/6: ENTRYPOINT ["/api"]
[2/2] COMMIT gotunnels-eclukq_api
--> 5c2fd56b4fad
Successfully tagged localhost/gotunnels-eclukq_api:latest
5c2fd56b4fade2cca3d1ea47b115832735edd6f028a65cbee2ffa844721fc752
[gotunnels] starting database…
>>>> Executing external compose provider "/usr/bin/podman-compose". Please see podman-compose(1) for how to disable this message. <<<<

758c55f24d6be2987b99185b7024fb515a5b9153cf7bb79348f17f3bb99b8645
Trying to pull docker.io/library/postgres:16-alpine...
Getting image source signatures
Copying blob f63c7a8df82b done   | 
Copying blob 55afa1ecc21d done   | 
Copying blob 0d4fedf9cad8 done   | 
Copying blob f7f6aac6fe13 done   | 
Copying blob f2511ae13411 done   | 
Copying blob 6c2eaa02a04a done   | 
Copying blob ecbe26720671 done   | 
Copying blob ddab922e8d89 done   | 
Copying blob 95e4c51fed83 done   | 
Copying blob 5de95df2a1fb done   | 
Copying blob d84bc3f3ded6 done   | 
Copying config e684c11a6c done   | 
Writing manifest to image destination
b35bfb2d3448161930cba7b61fc5c3b3c6045ae836b256d3410866fbd92e4b25
gotunnels-eclukq_db_1
[gotunnels] waiting for 'db' to become healthy (up to 120s)

real	0m25.370s
user	1m14.453s
sys	0m19.753s
kushal@fedora:~/src/golang/GoTunnels$ podman ps -a
CONTAINER ID  IMAGE                                             COMMAND               CREATED         STATUS                    PORTS                                                                                                        NAMES
91f36f84efef  docker.io/library/postgres:17                     postgres              3 months ago    Exited (0) 3 months ago   0.0.0.0:5432->5432/tcp                                                                                       otelshop-postgres
7116732afa27  mcr.microsoft.com/dotnet/aspire-dashboard:latest                        3 months ago    Exited (0) 3 months ago   0.0.0.0:4317-4318->18889-18890/tcp, 0.0.0.0:18888->18888/tcp                                                 otelshop-dashboard
d860bc098ccc  docker.io/library/rabbitmq:4-management           rabbitmq-server       3 weeks ago     Exited (0) 3 weeks ago    0.0.0.0:5672->5672/tcp, 0.0.0.0:15672->15672/tcp, 4369/tcp, 5671/tcp, 15671/tcp, 15691-15692/tcp, 25672/tcp  fedora-rabbit
a386f726e845  docker.io/library/rabbitmq:4-management           rabbitmq-server       3 weeks ago     Exited (0) 3 weeks ago    0.0.0.0:5672->5672/tcp, 0.0.0.0:15672->15672/tcp, 4369/tcp, 5671/tcp, 15671/tcp, 15691-15692/tcp, 25672/tcp  distributed-tasking_rabbitmq_1
b58ad2e975e0  docker.io/library/postgres:18                     postgres              3 weeks ago     Exited (0) 3 weeks ago    0.0.0.0:5432->5432/tcp                                                                                       distributed-tasking_postgres_1
dac3fb972737  docker.io/jaegertracing/jaeger:latest                                   3 weeks ago     Exited (0) 3 weeks ago    0.0.0.0:16686->16686/tcp, 4317-4318/tcp, 5778-5779/tcp, 9411/tcp, 13132-13133/tcp, 14250/tcp, 14268/tcp      distributed-tasking_jaeger_1
4ba7d986078c  localhost/distributed-tasking:local               ConsumerHost.dll      3 weeks ago     Exited (137) 3 weeks ago                                                                                                               distributed-tasking_consumer_1
dd261c151ea8  localhost/distributed-tasking:local               ProducerHost.dll      3 weeks ago     Exited (137) 3 weeks ago                                                                                                               distributed-tasking_producer_1
fbdb6998b953  localhost/iroh-ping:latest                                              2 weeks ago     Exited (0) 2 weeks ago                                                                                                                 iroh-ping
b031dd3f7f43  localhost/virginia:latest                                               2 weeks ago     Exited (137) 2 weeks ago  0.0.0.0:8080->8080/tcp                                                                                       virginia
33fcb8f958f3  localhost/virginia-iroh:latest                    listen-tcp --host...  2 weeks ago     Exited (137) 2 weeks ago                                                                                                               virginia-iroh
b322b0fba860  mcr.microsoft.com/dotnet/aspire-dashboard:9.0                           12 days ago     Exited (0) 8 days ago     0.0.0.0:18888->18888/tcp                                                                                     weather_aspire-dashboard_1
10118fcb0bbd  localhost/weather_otelcol:latest                  --config=/etc/ote...  12 days ago     Exited (0) 8 days ago     4317-4318/tcp, 55678-55679/tcp                                                                               weather_otelcol_1
881f639c5604  docker.io/cloudflare/cloudflared:latest           tunnel --no-autou...  12 days ago     Exited (0) 8 days ago                                                                                                                  weather_tunnel-aspire_1
4e78f13abb10  localhost/weather_api:latest                                            12 days ago     Exited (0) 8 days ago     0.0.0.0:8080->8080/tcp                                                                                       weather_api_1
407c161987e3  localhost/weather_web:latest                                            12 days ago     Exited (0) 8 days ago     0.0.0.0:8081->8080/tcp                                                                                       weather_web_1
34ece45d51fa  docker.io/cloudflare/cloudflared:latest           tunnel --no-autou...  12 days ago     Exited (0) 8 days ago                                                                                                                  weather_tunnel-api_1
32860347096f  docker.io/cloudflare/cloudflared:latest           tunnel --no-autou...  12 days ago     Exited (0) 8 days ago                                                                                                                  weather_tunnel-web_1
b35bfb2d3448  docker.io/library/postgres:16-alpine              postgres              26 seconds ago  Up 26 seconds (healthy)   5432/tcp                                                                                                     gotunnels-eclukq_db_1
kushal@fedora:~/src/golang/GoTunnels$ 

