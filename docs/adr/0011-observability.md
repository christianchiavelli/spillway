# 11. Observability: slog, OpenTelemetry and Prometheus

- Status: accepted
- Date: 2026-10-17

## Context

When a request fails, someone has to find out which one, what it did and how long it took, from a server they may not be able to attach a debugger to.

## Decision

- **Logs are `log/slog`**: JSON in containers, text on a laptop. Every line written during a request carries its `request_id`, and its `trace_id` and `span_id` when tracing is on. One access line per request gives the method, path, status, bytes and `duration_ms`, at error level for a 5xx.
- **One request id end to end.** The web makes one per request and sends it as `X-Request-Id`; the API keeps it, logs it and returns it; an error page shows it as a reference to quote.
- **Metrics are OpenTelemetry's, read by Prometheus** at `/metrics` on the admin port: ogen's request counts and durations per operation, query metrics from otelpgx, and the pool's connections.
- **Traces go out over OTLP** when `OTEL_EXPORTER_OTLP_ENDPOINT` is set, with a span per operation and per query. Unset, spans are not even recorded.
- **Probes live on the admin port, 9090**, never on the public one: `/livez` says the process answers, and `/readyz` that Postgres answers a ping and the server is not shutting down.

## Consequences

- A failure on a page can be followed from the error page's reference to the API's log lines.
- The exporter is a choice made in `platform/telemetry`, not in the code that records.

## Alternatives considered

- **The Prometheus client library directly:** one vendor's API in every package that records.
- **zap or zerolog:** faster, and slog is in the standard library and fast enough here.
