100
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