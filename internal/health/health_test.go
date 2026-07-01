package health

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLive(t *testing.T) {
	h := NewHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), Info{
		Service:    "gotunnels-api",
		InstanceID: "inst-1",
	})
	rec := httptest.NewRecorder()
	h.Live(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body: %v", err)
	}
	if body["status"] != "alive" {
		t.Fatalf("status field = %v", body["status"])
	}
	if body["instance_id"] != "inst-1" {
		t.Fatalf("instance_id = %v", body["instance_id"])
	}
}

func TestInfoHandler(t *testing.T) {
	h := NewHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), Info{
		Service:     "gotunnels-api",
		Version:     "1.2.3",
		CSPMode:     "report-only",
		TelemetryOn: true,
	})
	rec := httptest.NewRecorder()
	h.InfoHandler(rec, httptest.NewRequest(http.MethodGet, "/api/info", nil))

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body: %v", err)
	}
	if body["version"] != "1.2.3" {
		t.Fatalf("version = %v", body["version"])
	}
	if body["csp_mode"] != "report-only" {
		t.Fatalf("csp_mode = %v", body["csp_mode"])
	}
	if body["telemetry_on"] != true {
		t.Fatalf("telemetry_on = %v", body["telemetry_on"])
	}
}
