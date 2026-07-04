// Package captcha persists per-user statistics for the CAPTCHA demo page and
// serves its leaderboard.
//
// Design: the page's auto-solver can complete hundreds of puzzles per second,
// so the client never reports individual solves. It accumulates deltas and
// POSTs a small batch every few seconds (and on page hide); the server folds
// each batch into a single aggregate row per user. Totals only grow, the best
// streak is merged with GREATEST so it can never regress, and the current
// streak is a last-write-wins snapshot. Storage is O(users), request volume
// is O(seconds), and the numbers are honest enough for a demo leaderboard.
//
// Everything here is instrumented with OpenTelemetry: otelhttp (wired in the
// server package) already produces a span and HTTP metrics per request; this
// package adds domain metrics (solve counters by mode, sync counter, a streak
// histogram) and enriches the request span with attributes.
package captcha

import (
	"errors"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/kusl/GoTunnels/internal/auth"
	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/store"
)

// scopeName is the instrumentation scope for telemetry emitted here.
const scopeName = "github.com/kusl/GoTunnels/internal/captcha"

const (
	// maxDeltaPerSync bounds how many solves one batch may claim. At the
	// solver's hardware ceiling (~one grid per animation frame) a 4-second
	// batch tops out around a thousand solves, so this is generous headroom
	// while still keeping a single hostile request from minting absurd totals.
	maxDeltaPerSync = 100_000
	// maxStreak bounds reported streak snapshots.
	maxStreak = 1_000_000_000
	// leaderboardSize is how many rows the leaderboard endpoint returns.
	leaderboardSize = 10
)

// Handlers bundles dependencies for the CAPTCHA endpoints.
type Handlers struct {
	store *store.Store
	log   *slog.Logger

	solves  metric.Int64Counter
	syncs   metric.Int64Counter
	streaks metric.Int64Histogram
}

// NewHandlers builds the handler set and registers its OTel instruments. When
// telemetry is disabled the global meter is a no-op, so this stays free.
func NewHandlers(s *store.Store, log *slog.Logger) *Handlers {
	h := &Handlers{store: s, log: log}
	m := otel.Meter(scopeName)
	var err error
	if h.solves, err = m.Int64Counter("gotunnels.captcha.solves",
		metric.WithDescription("CAPTCHA solves folded into stats, by mode"),
		metric.WithUnit("{solve}")); err != nil {
		log.Warn("captcha: register solves counter", slog.String("error", err.Error()))
	}
	if h.syncs, err = m.Int64Counter("gotunnels.captcha.syncs",
		metric.WithDescription("CAPTCHA stat sync batches accepted"),
		metric.WithUnit("{batch}")); err != nil {
		log.Warn("captcha: register syncs counter", slog.String("error", err.Error()))
	}
	if h.streaks, err = m.Int64Histogram("gotunnels.captcha.streak",
		metric.WithDescription("Current streak reported at each sync"),
		metric.WithUnit("{solve}")); err != nil {
		log.Warn("captcha: register streak histogram", slog.String("error", err.Error()))
	}
	return h
}

// syncRequest is the client's batched progress report.
type syncRequest struct {
	ManualDelta   int64 `json:"manual_delta"`
	AutoDelta     int64 `json:"auto_delta"`
	CurrentStreak int64 `json:"current_streak"`
	BestStreak    int64 `json:"best_streak"`
}

// Stats returns the caller's aggregate stats. A user who has never synced gets
// an all-zero record rather than a 404, so the page has one uniform boot path.
func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	st, err := h.store.GetCaptchaStats(r.Context(), user.ID)
	if errors.Is(err, store.ErrNotFound) {
		st, err = store.CaptchaStats{UserID: user.ID}, nil
	}
	if err != nil {
		h.serverError(w, r, "captcha: get stats", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"stats": st})
}

// Sync folds one batched delta into the caller's stats and returns the updated
// aggregate so the client can reconcile (for example, totals accumulated from
// another device).
func (h *Handlers) Sync(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var req syncRequest
	if !httpx.DecodeJSON(w, r, &req, 4<<10) {
		return
	}
	in := clampSync(store.CaptchaSyncInput{
		ManualDelta:   req.ManualDelta,
		AutoDelta:     req.AutoDelta,
		CurrentStreak: req.CurrentStreak,
		BestStreak:    req.BestStreak,
	})

	st, err := h.store.SyncCaptchaStats(r.Context(), user.ID, in)
	if err != nil {
		h.serverError(w, r, "captcha: sync stats", err)
		return
	}

	// Domain telemetry: counters by mode, batch counter, streak histogram,
	// and attributes on the request span otelhttp already opened.
	ctx := r.Context()
	if h.solves != nil {
		if in.ManualDelta > 0 {
			h.solves.Add(ctx, in.ManualDelta, metric.WithAttributes(attribute.String("mode", "manual")))
		}
		if in.AutoDelta > 0 {
			h.solves.Add(ctx, in.AutoDelta, metric.WithAttributes(attribute.String("mode", "auto")))
		}
	}
	if h.syncs != nil {
		h.syncs.Add(ctx, 1)
	}
	if h.streaks != nil {
		h.streaks.Record(ctx, in.CurrentStreak)
	}
	trace.SpanFromContext(ctx).SetAttributes(
		attribute.Int64("captcha.manual_delta", in.ManualDelta),
		attribute.Int64("captcha.auto_delta", in.AutoDelta),
		attribute.Int64("captcha.current_streak", in.CurrentStreak),
		attribute.Int64("captcha.best_streak", st.BestStreak),
	)

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"stats": st})
}

// Reset deletes the caller's stats row entirely (they leave the leaderboard
// until they play again).
func (h *Handlers) Reset(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if err := h.store.DeleteCaptchaStats(r.Context(), user.ID); err != nil {
		h.serverError(w, r, "captcha: reset stats", err)
		return
	}
	trace.SpanFromContext(r.Context()).AddEvent("captcha.stats_reset")
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}

// Leaderboard returns the top players plus the caller's own ranked row (which
// may sit outside the top slice, or be absent if they have never played).
func (h *Handlers) Leaderboard(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	rows, err := h.store.CaptchaLeaderboard(r.Context(), leaderboardSize)
	if err != nil {
		h.serverError(w, r, "captcha: leaderboard", err)
		return
	}
	var mine *store.CaptchaLeaderboardRow
	if me, err := h.store.CaptchaRank(r.Context(), user.ID); err == nil {
		mine = &me
	} else if !errors.Is(err, store.ErrNotFound) {
		h.serverError(w, r, "captcha: own rank", err)
		return
	}
	if rows == nil {
		rows = []store.CaptchaLeaderboardRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"leaderboard": rows, "me": mine})
}

// clampSync bounds every client-supplied number: deltas are non-negative and
// capped, streak snapshots are non-negative and capped, and best is raised to
// at least current so a single batch is internally consistent.
func clampSync(in store.CaptchaSyncInput) store.CaptchaSyncInput {
	in.ManualDelta = clamp(in.ManualDelta, 0, maxDeltaPerSync)
	in.AutoDelta = clamp(in.AutoDelta, 0, maxDeltaPerSync)
	in.CurrentStreak = clamp(in.CurrentStreak, 0, maxStreak)
	in.BestStreak = clamp(in.BestStreak, 0, maxStreak)
	if in.BestStreak < in.CurrentStreak {
		in.BestStreak = in.CurrentStreak
	}
	return in
}

func clamp(v, lo, hi int64) int64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (h *Handlers) serverError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	h.log.ErrorContext(r.Context(), msg,
		slog.String("error", err.Error()),
		slog.String("request_id", httpx.RequestIDFromContext(r.Context())),
	)
	httpx.WriteError(w, http.StatusInternalServerError, "internal server error")
}
