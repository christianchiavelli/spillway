# 12. Graceful shutdown, with a drain

- Status: accepted
- Date: 2026-10-17

## Context

A deploy stops the old process while requests are in flight. Stopped abruptly, they fail; stopped before the load balancer notices, new ones fail too.

## Decision

- **A signal starts the shutdown** (`signal.NotifyContext` on SIGINT and SIGTERM). A second one kills the process at once.
- **First, `/readyz` turns 503.** The API keeps serving for `SHUTDOWN_DRAIN` (none by default; a few seconds behind a load balancer that polls readiness), then `http.Server.Shutdown` waits up to `SHUTDOWN_TIMEOUT` (20 s) for the requests in flight.
- **Everything stops in order.** The servers and the background jobs (the session purge and the throttle's sweep) run in one errgroup; the database pool closes after the servers, so no request loses its connection; telemetry flushes last.
- **Every request has a deadline**, `REQUEST_TIMEOUT` (10 s), which its context carries into every query.
- **The servers set their timeouts**: 5 s to read the headers, against slow clients, 15 s to read a request, 30 s to write an answer, 2 minutes idle, 64 KiB of headers.
- **The web server** stops on SIGTERM as adapter-node does, under tini in compose, since Node is not built to be PID 1.

## Consequences

- A rolling deploy loses no request: the load balancer stops sending before the server stops answering.
- Stopping takes up to the drain plus the timeout.

## Alternatives considered

- **`Shutdown` at the signal:** fine alone on a laptop, and behind a load balancer it refuses the requests already on their way.
