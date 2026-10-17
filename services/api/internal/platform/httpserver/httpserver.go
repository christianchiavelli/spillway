// Package httpserver runs an http.Server for as long as a context lives and
// then shuts it down without cutting the requests in flight.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// New returns a server with every timeout set. The zero values of net/http
// wait forever, which lets a slow client hold a connection open for good.
func New(addr string, h http.Handler, logger *slog.Logger) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    64 << 10,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
	}
}

// Run listens on srv.Addr and serves until ctx is done. Then it stops taking
// connections and waits up to timeout for the requests in flight, which keep
// their own contexts: cancelling ctx does not cancel them.
func Run(ctx context.Context, srv *http.Server, timeout time.Duration) error {
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", srv.Addr, err)
	}
	return Serve(ctx, srv, ln, timeout)
}

// Serve is Run on a listener the caller opened, as tests do on port 0.
func Serve(ctx context.Context, srv *http.Server, ln net.Listener, timeout time.Duration) error {
	served := make(chan error, 1)
	go func() { served <- srv.Serve(ln) }()

	select {
	case err := <-served:
		return fmt.Errorf("serve on %s: %w", ln.Addr(), err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close() // whatever did not finish in time is cut now
		return fmt.Errorf("shut down %s: %w", ln.Addr(), err)
	}
	if err := <-served; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve on %s: %w", ln.Addr(), err)
	}
	return nil
}
