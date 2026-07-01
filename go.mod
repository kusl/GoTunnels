module github.com/kusl/GoTunnels

go 1.23.0

// NOTE: go.sum is intentionally not committed on first import. Run
// `go mod tidy` once (the local network can reach the Go module proxy) to
// resolve and lock the full dependency graph, then commit go.sum. The
// container build (Containerfile.api) also runs `go mod tidy` so the very
// first `./run.sh` works without a pre-existing go.sum.
require (
	github.com/go-webauthn/webauthn v0.11.2
	github.com/jackc/pgx/v5 v5.6.0
	github.com/pquerna/otp v1.4.0
	go.opentelemetry.io/contrib/bridges/otelslog v0.4.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.53.0
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.4.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.28.0
	go.opentelemetry.io/otel/log v0.4.0
	go.opentelemetry.io/otel/metric v1.28.0
	go.opentelemetry.io/otel/sdk v1.28.0
	go.opentelemetry.io/otel/sdk/log v0.4.0
	go.opentelemetry.io/otel/sdk/metric v1.28.0
	go.opentelemetry.io/otel/trace v1.28.0
	golang.org/x/crypto v0.25.0
)
