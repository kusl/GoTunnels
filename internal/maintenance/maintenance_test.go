package maintenance

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"
)

type fakePruner struct {
	flows, health, csp int
	healthBefore       time.Time
	cspBefore          time.Time
	failFlows          bool
}

func (f *fakePruner) DeleteExpiredFlows(context.Context) (int64, error) {
	f.flows++
	if f.failFlows {
		return 0, errors.New("boom")
	}
	return 1, nil
}

func (f *fakePruner) PruneHealthChecks(_ context.Context, before time.Time) (int64, error) {
	f.health++
	f.healthBefore = before
	return 2, nil
}

func (f *fakePruner) PruneCSPReports(_ context.Context, before time.Time) (int64, error) {
	f.csp++
	f.cspBefore = before
	return 3, nil
}

func newJanitor(p Pruner, cfg Config, now time.Time) *Janitor {
	j := New(p, slog.New(slog.DiscardHandler), cfg)
	j.now = func() time.Time { return now }
	return j
}

func TestRunOnceUsesRetentionCutoffs(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	p := &fakePruner{}
	j := newJanitor(p, Config{Interval: time.Hour, HealthLogRetention: 24 * time.Hour, CSPReportRetention: 48 * time.Hour}, now)
	if err := j.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.flows != 1 || p.health != 1 || p.csp != 1 {
		t.Fatalf("calls = %+v", p)
	}
	if !p.healthBefore.Equal(now.Add(-24*time.Hour)) || !p.cspBefore.Equal(now.Add(-48*time.Hour)) {
		t.Fatalf("cutoffs = %v / %v", p.healthBefore, p.cspBefore)
	}
}

func TestRunOnceZeroRetentionKeepsRows(t *testing.T) {
	p := &fakePruner{}
	j := newJanitor(p, Config{Interval: time.Hour}, time.Now())
	if err := j.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.health != 0 || p.csp != 0 || p.flows != 1 {
		t.Fatalf("zero retention must skip pruning, calls = %+v", p)
	}
}

func TestRunOnceContinuesAfterFailure(t *testing.T) {
	p := &fakePruner{failFlows: true}
	j := newJanitor(p, Config{Interval: time.Hour, HealthLogRetention: time.Hour, CSPReportRetention: time.Hour}, time.Now())
	if err := j.RunOnce(context.Background()); err == nil {
		t.Fatal("expected the flow failure to be reported")
	}
	if p.health != 1 || p.csp != 1 {
		t.Fatalf("later steps must still run, calls = %+v", p)
	}
}

func TestRunStopsOnCancel(t *testing.T) {
	p := &fakePruner{}
	j := newJanitor(p, Config{Interval: time.Hour}, time.Now())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { j.Run(ctx); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}
