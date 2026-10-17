package health_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/christianchiavelli/spillway/services/api/internal/platform/health"
)

func TestReadiness(t *testing.T) {
	var down error
	p := health.New(func(context.Context) error { return down })

	status := func() int {
		rec := httptest.NewRecorder()
		p.Ready(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		return rec.Code
	}

	if got := status(); got != http.StatusServiceUnavailable {
		t.Errorf("before SetReady(true): %d, want 503", got)
	}
	p.SetReady(true)
	if got := status(); got != http.StatusOK {
		t.Errorf("ready: %d, want 200", got)
	}
	down = errors.New("database down")
	if got := status(); got != http.StatusServiceUnavailable {
		t.Errorf("dependency down: %d, want 503", got)
	}
	down = nil
	p.SetReady(false)
	if got := status(); got != http.StatusServiceUnavailable {
		t.Errorf("draining: %d, want 503", got)
	}

	rec := httptest.NewRecorder()
	p.Live(rec, httptest.NewRequest(http.MethodGet, "/livez", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("live while draining: %d, want 200", rec.Code)
	}
}
