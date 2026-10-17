package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/trace"

	"github.com/christianchiavelli/spillway/services/api/internal/platform/logging"
)

func TestLinesCarryTheContextsIDs(t *testing.T) {
	var out bytes.Buffer
	logger, err := logging.New(&out, "json", slog.LevelInfo)
	if err != nil {
		t.Fatal(err)
	}
	ctx := logging.WithAttrs(t.Context(), slog.String("request_id", "req-1"))
	ctx = trace.ContextWithSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1},
		SpanID:     trace.SpanID{2},
		TraceFlags: trace.FlagsSampled,
	}))

	logger.With("component", "test").InfoContext(ctx, "hello")

	var line map[string]any
	if err := json.Unmarshal(out.Bytes(), &line); err != nil {
		t.Fatalf("not one JSON line: %s", out.String())
	}
	for key, want := range map[string]string{
		"msg":        "hello",
		"component":  "test",
		"request_id": "req-1",
		"trace_id":   "01000000000000000000000000000000",
		"span_id":    "0200000000000000",
	} {
		if line[key] != want {
			t.Errorf("%s = %v, want %q", key, line[key], want)
		}
	}
}

func TestWithAttrsDoesNotShareTheParentsSlice(t *testing.T) {
	var out bytes.Buffer
	logger, _ := logging.New(&out, "json", slog.LevelInfo)
	parent := logging.WithAttrs(t.Context(), slog.String("a", "1"))
	_ = logging.WithAttrs(parent, slog.String("b", "2"))
	_ = logging.WithAttrs(parent, slog.String("c", "3"))

	logger.InfoContext(parent, "parent")
	var line map[string]any
	_ = json.Unmarshal(out.Bytes(), &line)
	if _, ok := line["b"]; ok {
		t.Error("a child's attribute leaked into its parent")
	}
}

func TestUnknownFormat(t *testing.T) {
	if _, err := logging.New(&bytes.Buffer{}, "xml", slog.LevelInfo); err == nil {
		t.Error("New() accepted an unknown format")
	}
}
