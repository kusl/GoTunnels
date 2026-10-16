package telemetry

import (
	"context"
	"testing"

	"github.com/kusl/GoTunnels/internal/config"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestTemporalitySelector(t *testing.T) {
	cases := []struct {
		pref string
		kind sdkmetric.InstrumentKind
		want metricdata.Temporality
	}{
		{"delta", sdkmetric.InstrumentKindCounter, metricdata.DeltaTemporality},
		{"delta", sdkmetric.InstrumentKindObservableCounter, metricdata.DeltaTemporality},
		{"delta", sdkmetric.InstrumentKindUpDownCounter, metricdata.CumulativeTemporality},
		{"lowmemory", sdkmetric.InstrumentKindObservableCounter, metricdata.CumulativeTemporality},
		{"lowmemory", sdkmetric.InstrumentKindHistogram, metricdata.DeltaTemporality},
		{"cumulative", sdkmetric.InstrumentKindCounter, metricdata.CumulativeTemporality},
		{"bogus", sdkmetric.InstrumentKindCounter, metricdata.CumulativeTemporality},
	}
	for _, c := range cases {
		if got := temporalitySelector(c.pref)(c.kind); got != c.want {
			t.Errorf("%s/%v = %v, want %v", c.pref, c.kind, got, c.want)
		}
	}
}

func TestAggregationSelector(t *testing.T) {
	agg := aggregationSelector("base2_exponential_bucket_histogram")(sdkmetric.InstrumentKindHistogram)
	if _, ok := agg.(sdkmetric.AggregationBase2ExponentialHistogram); !ok {
		t.Fatalf("histogram aggregation = %T, want exponential", agg)
	}
	agg = aggregationSelector("explicit_bucket_histogram")(sdkmetric.InstrumentKindHistogram)
	if _, ok := agg.(sdkmetric.AggregationBase2ExponentialHistogram); ok {
		t.Fatal("explicit preference must not select exponential buckets")
	}
}

func TestSetupDisabledLogsToStdoutOnly(t *testing.T) {
	cfg := &config.Config{ServiceName: "svc", InstanceID: "i-1"}
	p, err := Setup(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if p.Logger == nil {
		t.Fatal("logger must always be usable")
	}
	if err := p.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown with no providers: %v", err)
	}
}
