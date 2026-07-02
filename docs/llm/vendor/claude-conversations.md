25
12

What went wrong now? 

cd ~/src/golang/GoTunnels/; export UPTRACE_DSN="https://JhkQqxGHXMjQwCptK5Qpzr@api.uptrace.dev?grpc=4317"; time bash scripts/up.sh
[gotunnels] using runtime: podman / compose: podman compose
[gotunnels] using existing /home/kushal/src/golang/GoTunnels/.env
[gotunnels] project (instance): gotunnels-66bdbw
[gotunnels] building images…
>>>> Executing external compose provider "/usr/bin/podman-compose". Please see podman-compose(1) for how to disable this message. <<<<

STEP 1/5: FROM docker.io/library/caddy:2-alpine
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
