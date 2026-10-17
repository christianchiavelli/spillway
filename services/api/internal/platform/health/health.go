// Package health answers the two questions an orchestrator asks: is the
// process alive, and should it get traffic.
package health

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"
)

// Probe holds the readiness of the process.
type Probe struct {
	ready atomic.Bool
	check func(context.Context) error
}

// New returns a Probe that is not ready yet, and that checks a dependency,
// such as the database, every time readiness is asked.
func New(check func(context.Context) error) *Probe {
	return &Probe{check: check}
}

// SetReady says whether the process should get traffic. It turns false first
// thing on shutdown, so traffic drains away before the servers stop.
func (p *Probe) SetReady(ready bool) {
	p.ready.Store(ready)
}

// Live answers 200 while the process can answer at all.
func (p *Probe) Live(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("live\n"))
}

// Ready answers 200 when the process is ready and its dependency answers
// within two seconds, and 503 otherwise.
func (p *Probe) Ready(w http.ResponseWriter, r *http.Request) {
	if !p.ready.Load() {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := p.check(ctx); err != nil {
		http.Error(w, "dependency down", http.StatusServiceUnavailable)
		return
	}
	_, _ = w.Write([]byte("ready\n"))
}
