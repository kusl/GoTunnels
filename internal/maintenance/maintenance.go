// Package maintenance runs periodic housekeeping so no table grows without
// bound: expired WebAuthn ceremonies are deleted, and probe history and CSP
// reports are pruned past their retention.
package maintenance

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

// Pruner is the slice of the store the janitor needs. Depending on this
// interface rather than *store.Store keeps the janitor unit-testable.
type Pruner interface {
	DeleteExpiredFlows(ctx context.Context) (int64, error)
	PruneHealthChecks(ctx context.Context, before time.Time) (int64, error)
	PruneCSPReports(ctx context.Context, before time.Time) (int64, error)
}

// Config controls the janitor. A zero retention keeps that table forever.
type Config struct {
	Interval           time.Duration
	HealthLogRetention time.Duration
	CSPReportRetention time.Duration
}

// Janitor performs housekeeping passes.
type Janitor struct {
	store Pruner
	log   *slog.Logger
	cfg   Config
	now   func() time.Time
}

// New builds a Janitor.
func New(store Pruner, log *slog.Logger, cfg Config) *Janitor {
	return &Janitor{store: store, log: log, cfg: cfg, now: time.Now}
}

// passTimeout bounds one housekeeping pass.
const passTimeout = time.Minute

// RunOnce performs a single housekeeping pass. Every step runs even when an
// earlier one fails; all failures are returned together.
func (j *Janitor) RunOnce(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, passTimeout)
	defer cancel()
	now := j.now()
	var errs []error

	if n, err := j.store.DeleteExpiredFlows(ctx); err != nil {
		errs = append(errs, err)
	} else if n > 0 {
		j.log.InfoContext(ctx, "maintenance: deleted expired webauthn flows", slog.Int64("rows", n))
	}
	if j.cfg.HealthLogRetention > 0 {
		if n, err := j.store.PruneHealthChecks(ctx, now.Add(-j.cfg.HealthLogRetention)); err != nil {
			errs = append(errs, err)
		} else if n > 0 {
			j.log.InfoContext(ctx, "maintenance: pruned health check log", slog.Int64("rows", n))
		}
	}
	if j.cfg.CSPReportRetention > 0 {
		if n, err := j.store.PruneCSPReports(ctx, now.Add(-j.cfg.CSPReportRetention)); err != nil {
			errs = append(errs, err)
		} else if n > 0 {
			j.log.InfoContext(ctx, "maintenance: pruned csp reports", slog.Int64("rows", n))
		}
	}
	return errors.Join(errs...)
}

// Run performs a pass immediately and then every Interval until ctx is done.
func (j *Janitor) Run(ctx context.Context) {
	ticker := time.NewTicker(j.cfg.Interval)
	defer ticker.Stop()
	for {
		if err := j.RunOnce(ctx); err != nil && ctx.Err() == nil {
			j.log.WarnContext(ctx, "maintenance pass failed", slog.String("error", err.Error()))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
