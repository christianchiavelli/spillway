package httpapi

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/christianchiavelli/spillway/services/api/internal/adapter/httpapi/oas"
)

// Options tune the edge.
type Options struct {
	Logger *slog.Logger
	// AllowedOrigins may call from a browser. Empty, the default, allows none.
	AllowedOrigins []string
	MaxBodyBytes   int64
	RequestTimeout time.Duration
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
}

// New returns the API's HTTP handler: the generated router, with the
// middleware every request passes through on its way in.
func New(alerts Alerts, ident Identity, opts Options) (http.Handler, error) {
	h := &handler{alerts: alerts, identity: ident, logger: opts.Logger}
	srv, err := oas.NewServer(h, h,
		oas.WithErrorHandler(h.handleRequestError),
		oas.WithNotFound(h.notFound),
		oas.WithMethodNotAllowed(h.methodNotAllowed),
		oas.WithTracerProvider(opts.TracerProvider),
		oas.WithMeterProvider(opts.MeterProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("build router: %w", err)
	}
	return chain(srv,
		requestID,
		accessLog(opts.Logger),
		recoverPanics(opts.Logger),
		apiHeaders,
		crossOrigin(opts.AllowedOrigins),
		limitBody(opts.MaxBodyBytes),
		deadline(opts.RequestTimeout),
	), nil
}
