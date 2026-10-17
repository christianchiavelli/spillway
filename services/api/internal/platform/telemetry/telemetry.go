// Package telemetry sets up OpenTelemetry: metrics always, read by
// Prometheus at /metrics on the admin port, and traces when an OTLP endpoint
// is configured. ogen instruments every operation with both, and pgx every
// query, so a trace that starts in the web app ends at the SQL.
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Options say what the service is and where traces go.
type Options struct {
	Service string
	Version string
	// ExportTraces sends spans to the endpoint in the standard
	// OTEL_EXPORTER_OTLP_* variables. Off, spans are not even recorded.
	ExportTraces bool
}

// Telemetry holds the providers the rest of the service is given.
type Telemetry struct {
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	// Metrics serves every metric in the Prometheus text format.
	Metrics  http.Handler
	shutdown []func(context.Context) error
}

// Setup builds the providers. Call Shutdown before exiting, so spans still
// in memory are sent.
func Setup(ctx context.Context, opts Options) (*Telemetry, error) {
	// Schemaless: resource.Default carries the SDK's semantic conventions
	// version, and merging it with another version's schema URL is an error.
	res, err := resource.Merge(resource.Default(), resource.NewSchemaless(
		semconv.ServiceName(opts.Service),
		semconv.ServiceVersion(opts.Version),
	))
	if err != nil {
		return nil, fmt.Errorf("describe the service: %w", err)
	}
	t := &Telemetry{}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reader, err := otelprom.New(otelprom.WithRegisterer(registry))
	if err != nil {
		return nil, fmt.Errorf("create the prometheus exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader), sdkmetric.WithResource(res))
	t.MeterProvider = mp
	t.Metrics = promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	t.shutdown = append(t.shutdown, mp.Shutdown)

	t.TracerProvider = noop.NewTracerProvider()
	if opts.ExportTraces {
		exporter, err := otlptracehttp.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("create the trace exporter: %w", err)
		}
		tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(res))
		t.TracerProvider = tp
		t.shutdown = append(t.shutdown, tp.Shutdown)
	}

	// W3C trace context, so a trace the web app starts continues here.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(t.TracerProvider)
	otel.SetMeterProvider(t.MeterProvider)
	return t, nil
}

// QueryTracer returns a pgx tracer that records each query as a span.
func (t *Telemetry) QueryTracer() pgx.QueryTracer {
	return otelpgx.NewTracer(otelpgx.WithTracerProvider(t.TracerProvider), otelpgx.WithDisableSQLStatementInAttributes())
}

// RecordPool publishes the pool's connection counts as metrics: the first
// thing to read when requests queue for the database.
func (t *Telemetry) RecordPool(pool *pgxpool.Pool) error {
	if err := otelpgx.RecordStats(pool, otelpgx.WithStatsMeterProvider(t.MeterProvider)); err != nil {
		return fmt.Errorf("record pool stats: %w", err)
	}
	return nil
}

// Shutdown flushes and stops the providers.
func (t *Telemetry) Shutdown(ctx context.Context) error {
	var errs []error
	for _, f := range t.shutdown {
		errs = append(errs, f(ctx))
	}
	return errors.Join(errs...)
}
