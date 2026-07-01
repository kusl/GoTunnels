module github.com/kusl/GoTunnels

go 1.26.0

// NOTE: go.sum is intentionally not committed on first import. Run
// `go mod tidy` once (the local network can reach the Go module proxy) to
// resolve and lock the full dependency graph, then commit go.sum. The
// container build (Containerfile.api) also runs `go mod tidy` so the very
// first `./run.sh` works without a pre-existing go.sum.
require (
	github.com/go-webauthn/webauthn v0.11.2
	github.com/jackc/pgx/v5 v5.10.0
	github.com/pquerna/otp v1.4.0
	go.opentelemetry.io/contrib/bridges/otelslog v0.19.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0
	go.opentelemetry.io/otel v1.44.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.20.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.44.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.44.0
	go.opentelemetry.io/otel/log v0.20.0
	go.opentelemetry.io/otel/metric v1.44.0
	go.opentelemetry.io/otel/sdk v1.44.0
	go.opentelemetry.io/otel/sdk/log v0.20.0
	go.opentelemetry.io/otel/sdk/metric v1.44.0
	go.opentelemetry.io/otel/trace v1.44.0
	golang.org/x/crypto v0.53.0
)

// Security floor for GO-2025-3553 (excessive memory allocation during JWT
// header parsing). jwt/v5 is not imported directly by our code; it is pulled
// in transitively by go-webauthn, which currently requires the vulnerable
// v5.2.1. This explicit require raises the minimum to the patched v5.2.2 —
// `go mod tidy` will keep it as the selected version and will not downgrade
// below an explicit require. Once go-webauthn requires v5.2.2+ this line can
// be removed.
require github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
