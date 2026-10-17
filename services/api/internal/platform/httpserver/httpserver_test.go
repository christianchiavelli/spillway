package httpserver_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/christianchiavelli/spillway/services/api/internal/platform/httpserver"
)

// A request in flight when the context ends still gets its whole answer.
func TestShutdownLetsRequestsInFlightFinish(t *testing.T) {
	entered := make(chan struct{})
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		time.Sleep(200 * time.Millisecond)
		_, _ = io.WriteString(w, "finished")
	})
	srv := httpserver.New("", h, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- httpserver.Serve(ctx, srv, ln, 5*time.Second) }()

	type result struct {
		body string
		err  error
	}
	got := make(chan result, 1)
	go func() {
		res, err := http.Get("http://" + ln.Addr().String())
		if err != nil {
			got <- result{err: err}
			return
		}
		defer res.Body.Close()
		b, err := io.ReadAll(res.Body)
		got <- result{body: string(b), err: err}
	}()

	<-entered
	cancel()

	if r := <-got; r.err != nil || r.body != "finished" {
		t.Errorf("request in flight = %q, %v; want its whole answer", r.body, r.err)
	}
	if err := <-done; err != nil {
		t.Errorf("Serve() = %v, want a clean stop", err)
	}
	if _, err := http.Get("http://" + ln.Addr().String()); err == nil {
		t.Error("the server still answers after shutting down")
	}
}

func TestShutdownGivesUpAfterTheTimeout(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
	})
	srv := httpserver.New("", h, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- httpserver.Serve(ctx, srv, ln, 50*time.Millisecond) }()
	go func() { _, _ = http.Get("http://" + ln.Addr().String()) }()

	<-entered
	cancel()
	if err := <-done; err == nil || !strings.Contains(err.Error(), "shut down") {
		t.Errorf("Serve() = %v, want a shutdown timeout", err)
	}
}

func TestRunReportsAPortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	srv := httpserver.New(ln.Addr().String(), http.NotFoundHandler(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := httpserver.Run(t.Context(), srv, time.Second); err == nil || !strings.Contains(err.Error(), "listen on") {
		t.Errorf("Run() on a used port = %v", err)
	}
}
