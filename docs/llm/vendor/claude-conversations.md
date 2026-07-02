25
12

What went wrong now? 

kushal@fedora:~/src/golang/GoTunnels$ cd ~/src/golang/GoTunnels/; export UPTRACE_DSN="https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"; time bash export.sh > docs/llm/vendor/output.txt

real	0m2.367s
user	0m1.190s
sys	0m1.473s
kushal@fedora:~/src/golang/GoTunnels$ cd ~/src/golang/GoTunnels/; export UPTRACE_DSN="https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"; time bash scripts/up.sh
[gotunnels] using runtime: podman / compose: podman compose
[gotunnels] using existing /home/kushal/src/golang/GoTunnels/.env
[gotunnels] project (instance): gotunnels-66bdbw
[gotunnels] building images…
>>>> Executing external compose provider "/usr/bin/podman-compose". Please see podman-compose(1) for how to disable this message. <<<<

STEP 1/5: FROM docker.io/library/caddy:2-alpine
[1/2] STEP 1/6: FROM golang:1.26-bookworm AS build
STEP 2/5: COPY frontend/ /srv/
--> Using cache 8de81c4f823c69b139983dfef93489bd1d9e76c3c0bd71f2456c1fe0489bc6f7
--> 8de81c4f823c
STEP 3/5: COPY frontend/Caddyfile /etc/caddy/Caddyfile
--> Using cache 1dcdf59f71ba7b2ce5081770fedc6ddd4da70570a6766ebbeeeedc6e70b7f509
--> 1dcdf59f71ba
STEP 4/5: RUN rm -f /srv/Caddyfile
--> Using cache e170163757ffe0bd8d211ece66b0b9ff4f2423db1809d11d43375364333bc892
--> e170163757ff
STEP 5/5: EXPOSE 8080
--> Using cache 1d92b326891da32bb747a4dd759c2d8fc6e4bc622f724fb076c00bc30f0cd0dd
COMMIT gotunnels-66bdbw_frontend
--> 1d92b326891d
Successfully tagged localhost/gotunnels-66bdbw_frontend:latest
Successfully tagged localhost/gotunnels-d5vnw_frontend:latest
Successfully tagged localhost/gotunnels-ym1zdq_frontend:latest
Successfully tagged localhost/gotunnels-dt6vjg_frontend:latest
Successfully tagged localhost/gotunnels-qdiyew_frontend:latest
Successfully tagged localhost/gotunnels-fifbya_frontend:latest
Successfully tagged localhost/gotunnels-eclukq_frontend:latest
1d92b326891da32bb747a4dd759c2d8fc6e4bc622f724fb076c00bc30f0cd0dd
[1/2] STEP 2/6: WORKDIR /src
--> Using cache babdaca5bcc2fececbbc19b62b0bfc9eaedfc787aa587199af487dc7ea867421
--> babdaca5bcc2
[1/2] STEP 3/6: ENV CGO_ENABLED=0     GOFLAGS=-mod=mod     GOTOOLCHAIN=local
--> Using cache 2048cbee3e7520d1cefe0c1d68b41d63b22a875582bb19fed499dd483de336d7
--> 2048cbee3e75
[1/2] STEP 4/6: COPY . .
--> f9869653d044
[1/2] STEP 5/6: ARG VERSION=dev
--> 0f5ae77146fc
[1/2] STEP 6/6: RUN go mod tidy  && go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/api ./cmd/api
go: downloading github.com/go-webauthn/webauthn v0.11.2
go: downloading github.com/jackc/pgx/v5 v5.10.0
go: downloading go.opentelemetry.io/contrib/bridges/otelslog v0.19.0
go: downloading go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0
go: downloading go.opentelemetry.io/otel v1.44.0
go: downloading github.com/pquerna/otp v1.4.0
go: downloading go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.20.0
go: downloading go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.44.0
go: downloading golang.org/x/crypto v0.53.0
go: downloading go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.44.0
go: downloading go.opentelemetry.io/otel/log v0.20.0
go: downloading go.opentelemetry.io/otel/sdk/log v0.20.0
go: downloading go.opentelemetry.io/otel/sdk/metric v1.44.0
go: downloading go.opentelemetry.io/otel/sdk v1.44.0
go: downloading go.opentelemetry.io/otel/metric v1.44.0
go: downloading go.opentelemetry.io/otel/trace v1.44.0
go: downloading github.com/jackc/puddle/v2 v2.2.2
go: downloading github.com/stretchr/testify v1.11.1
go: downloading github.com/felixge/httpsnoop v1.0.4
go: downloading github.com/google/uuid v1.6.0
go: downloading golang.org/x/sys v0.46.0
go: downloading github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc
go: downloading github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
go: downloading github.com/jackc/pgpassfile v1.0.0
go: downloading golang.org/x/text v0.38.0
go: downloading github.com/go-logr/logr v1.4.3
go: downloading github.com/go-logr/stdr v1.2.2
go: downloading github.com/google/go-cmp v0.7.0
go: downloading go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.44.0
go: downloading go.opentelemetry.io/proto/otlp v1.10.0
go: downloading google.golang.org/protobuf v1.36.11
go: downloading github.com/go-webauthn/x v0.1.14
go: downloading github.com/mitchellh/mapstructure v1.5.0
go: downloading github.com/golang-jwt/jwt/v5 v5.2.2
go: downloading github.com/google/go-tpm v0.9.1
go: downloading go.opentelemetry.io/otel/metric/x v0.66.0
go: downloading go.uber.org/goleak v1.3.0
go: downloading go.opentelemetry.io/auto/sdk v1.2.1
go: downloading google.golang.org/grpc v1.81.1
go: downloading github.com/cenkalti/backoff/v5 v5.0.3
go: downloading github.com/davecgh/go-spew v1.1.1
go: downloading golang.org/x/sync v0.17.0
go: downloading github.com/pmezard/go-difflib v1.0.0
go: downloading github.com/fxamacker/cbor/v2 v2.7.0
go: downloading github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0
go: downloading github.com/cespare/xxhash/v2 v2.3.0
go: downloading go.opentelemetry.io/otel/sdk/log/logtest v0.20.0
go: downloading gopkg.in/yaml.v3 v3.0.1
go: downloading github.com/x448/float16 v0.8.4
go: downloading google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa
go: downloading google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa
go: downloading golang.org/x/net v0.55.0
go: downloading golang.org/x/sync v0.21.0
go: downloading github.com/golang/protobuf v1.5.4
go: downloading gonum.org/v1/gonum v0.17.0
--> a52de2f9f104
[2/2] STEP 1/6: FROM gcr.io/distroless/static-debian12:nonroot
[2/2] STEP 2/6: WORKDIR /
--> Using cache ea9af167ae5e058f672df8dadcc271a9bf0f3c9cde9671a600e8616c9b0b230b
--> ea9af167ae5e
[2/2] STEP 3/6: COPY --from=build /out/api /api
--> Using cache 8a9f1e128ef6872a254a54e5f04402dcce3647879841cf2729bfe5d4160c831f
--> 8a9f1e128ef6
[2/2] STEP 4/6: USER nonroot:nonroot
--> Using cache 633c109c36ef05d19fa7ed1c073822cb909a4a84055690d74249d0d360baed5b
--> 633c109c36ef
[2/2] STEP 5/6: EXPOSE 8080
--> Using cache 0f87d03655e699e3a9ae0b4ee80d6799a5a1f336775af9542c3ff6ee15a4679c
--> 0f87d03655e6
[2/2] STEP 6/6: ENTRYPOINT ["/api"]
--> Using cache 5c2fd56b4fade2cca3d1ea47b115832735edd6f028a65cbee2ffa844721fc752
[2/2] COMMIT gotunnels-66bdbw_api
--> 5c2fd56b4fad
Successfully tagged localhost/gotunnels-66bdbw_api:latest
Successfully tagged localhost/gotunnels-d5vnw_api:latest
Successfully tagged localhost/gotunnels-ym1zdq_api:latest
Successfully tagged localhost/gotunnels-dt6vjg_api:latest
Successfully tagged localhost/gotunnels-qdiyew_api:latest
Successfully tagged localhost/gotunnels-fifbya_api:latest
Successfully tagged localhost/gotunnels-eclukq_api:latest
5c2fd56b4fade2cca3d1ea47b115832735edd6f028a65cbee2ffa844721fc752
[gotunnels] starting database…
>>>> Executing external compose provider "/usr/bin/podman-compose". Please see podman-compose(1) for how to disable this message. <<<<

ba519ad3a526400fda3ef6e4adbdcaf9c852e9936d7e9e76cf3531ded825509a
a9b2e3537d0e8b0ec702d8cae9f9be0ba4933e83c1838785ef3d6ebdd89ff72f
gotunnels-66bdbw_db_1
[gotunnels] waiting for 'db' to become ready (up to 60s)
[gotunnels] 'db' did not become ready in 60s
[gotunnels]   no container id resolved for 'db' — was it created? (check: dc -p gotunnels-66bdbw ps)

real	1m26.567s
user	1m18.108s
sys	0m22.206s
kushal@fedora:~/src/golang/GoTunnels$ 
