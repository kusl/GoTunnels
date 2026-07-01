package telemetry

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestMultiHandler_FansOut(t *testing.T) {
	var a, b bytes.Buffer
	h := NewMultiHandler(
		slog.NewTextHandler(&a, nil),
		slog.NewJSONHandler(&b, nil),
	)
	logger := slog.New(h)
	logger.Info("hello", slog.String("k", "v"))

	if !strings.Contains(a.String(), "hello") {
		t.Errorf("text handler missing message: %q", a.String())
	}
	if !strings.Contains(b.String(), "hello") || !strings.Contains(b.String(), `"k":"v"`) {
		t.Errorf("json handler missing message/attr: %q", b.String())
	}
}

func TestMultiHandler_WithAttrsAndGroup(t *testing.T) {
	var buf bytes.Buffer
	h := NewMultiHandler(slog.NewJSONHandler(&buf, nil))
	logger := slog.New(h).With(slog.String("svc", "gotunnels")).WithGroup("req")
	logger.Info("done", slog.Int("status", 200))

	out := buf.String()
	if !strings.Contains(out, `"svc":"gotunnels"`) {
		t.Errorf("expected top-level attr, got %q", out)
	}
	if !strings.Contains(out, `"req":{`) || !strings.Contains(out, `"status":200`) {
		t.Errorf("expected grouped attr, got %q", out)
	}
}

func TestMultiHandler_Enabled(t *testing.T) {
	var buf bytes.Buffer
	// Only enabled at Error and above.
	errOnly := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})
	h := NewMultiHandler(errOnly)
	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("should not be enabled at Info when only Error handler is present")
	}
	if !h.Enabled(context.Background(), slog.LevelError) {
		t.Error("should be enabled at Error")
	}
}
