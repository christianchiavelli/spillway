package telemetry_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/platform/telemetry"
)

// Setup once refused to start the server: the SDK's default resource and
// ours named two versions of the semantic conventions. Starting it here
// catches that kind of mistake before a deploy does.
func TestSetupServesMetricsAndShutsDown(t *testing.T) {
	tel, err := telemetry.Setup(t.Context(), telemetry.Options{Service: "spillway-api", Version: "test"})
	if err != nil {
		t.Fatalf("Setup() = %v", err)
	}

	counter, err := tel.MeterProvider.Meter("test").Int64Counter("probe_total")
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(t.Context(), 3)

	rec := httptest.NewRecorder()
	tel.Metrics.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body, _ := io.ReadAll(rec.Body)
	if rec.Code != http.StatusOK || !strings.Contains(string(body), "probe_total") || !strings.Contains(string(body), "go_goroutines") {
		t.Errorf("/metrics = %d:\n%s", rec.Code, body)
	}

	if err := tel.Shutdown(t.Context()); err != nil {
		t.Errorf("Shutdown() = %v", err)
	}
}

func TestSetupRecordsSpansWhenTracesAreExported(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:1")
	tel, err := telemetry.Setup(t.Context(), telemetry.Options{Service: "spillway-api", Version: "test", ExportTraces: true})
	if err != nil {
		t.Fatalf("Setup() = %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		_ = tel.Shutdown(ctx) // the endpoint is unreachable on purpose
	})
	_, span := tel.TracerProvider.Tracer("test").Start(t.Context(), "probe")
	if !span.SpanContext().IsValid() {
		t.Error("spans are not recorded with traces on")
	}
	span.End()
}
