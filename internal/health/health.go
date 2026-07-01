// Package health provides the two standard health endpoints plus a small info
// endpoint. Liveness answers "is the process up" without touching
// dependencies; readiness answers "can I serve traffic" by probing Postgres
// and recording each probe to health_check_log for a queryable history.
package health

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/kusl/GoTunnels/internal/httpx"
	"github.com/kusl/GoTunnels/internal/store"
)

// Info is static metadata surfaced on the info endpoint so the whole system
// agrees on one central configuration.
type Info struct {
	Service     string
	InstanceID  string
	Version     string
	CSPMode     string
	CSPPolicy   string
	TelemetryOn bool
}

// Handler serves the health and info endpoints.
type Handler struct {
	store *store.Store
	log   *slog.Logger
	info  Info
}

// NewHandler builds a health handler.
func NewHandler(s *store.Store, log *slog.Logger, info Info) *Handler {
	return &Handler{store: s, log: log, info: info}
}

// Live reports process liveness. It never checks dependencies.
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":      "alive",
		"service":     h.info.Service,
		"instance_id": h.info.InstanceID,
	})
}

// Ready reports readiness by probing the database and logging the result.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	err := h.store.Ping(r.Context())
	latencyMs := float64(time.Since(start).Microseconds()) / 1000.0

	status := "ok"
	code := http.StatusOK
	detail := ""
	if err != nil {
		status = "fail"
		code = http.StatusServiceUnavailable
		detail = err.Error()
	}

	// Best-effort history write; never let logging failure change the verdict.
	if logErr := h.store.InsertHealthCheck(r.Context(), "database", status, latencyMs, detail); logErr != nil {
		h.log.WarnContext(r.Context(), "record health check", slog.String("error", logErr.Error()))
	}

	httpx.WriteJSON(w, code, map[string]any{
		"status": status,
		"checks": []map[string]any{
			{"name": "database", "status": status, "latency_ms": latencyMs},
		},
	})
}

// InfoHandler returns static configuration/metadata.
func (h *Handler) InfoHandler(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"service":      h.info.Service,
		"instance_id":  h.info.InstanceID,
		"version":      h.info.Version,
		"csp_mode":     h.info.CSPMode,
		"csp_policy":   h.info.CSPPolicy,
		"telemetry_on": h.info.TelemetryOn,
	})
}
