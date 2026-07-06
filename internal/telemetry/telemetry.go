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
	"go.opentelemetry.io/otel/attribute"
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

// buildResource assembles the OTel resource attached to every span, metric,
// and log record. host.name deserves a note: the SDK does not detect it by
// default (resource.Default() covers service/SDK attributes only), which is
// why backends show host_name as an empty string out of the box. When
// cfg.HostName is set — scripts/up.sh exports the podman host's `hostname`,
// e.g. "virginia" — it is pinned explicitly, which is the value people
// actually want (the machine the stack runs on). Otherwise resource.WithHost()
// fills in the container's hostname: a random short id, less meaningful, but
// still better than empty.
func buildResource(ctx context.Context, cfg *config.Config) (*resource.Resource, error) {
	opts := []resource.Option{
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	}
	attrs := []attribute.KeyValue{
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.Version),
		semconv.ServiceInstanceID(cfg.InstanceID),
	}
	if cfg.HostName != "" {
		attrs = append(attrs, semconv.HostName(cfg.HostName))
	} else {
		opts = append(opts, resource.WithHost())
	}
	opts = append(opts, resource.WithAttributes(attrs...))
	return resource.New(ctx, opts...)
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
