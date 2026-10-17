# 3. Hexagonal, at Go's size

- Status: accepted
- Date: 2026-10-17

## Context

The API should keep its rules apart from HTTP and SQL, in the manner of Clean or Hexagonal Architecture. In Go, the textbook shapes of both (a package per layer, an interface for everything, mappers between identical structs) cost more than they return.

## Decision

Four kinds of package under `services/api/internal`, and dependencies that point inwards:

- **`domain/...`**: entities, values and rules. `alert.New` decides whether a draft is an alert, and `account.CheckPassword` whether a password will do. No I/O and no context: the standard library and the domain's own errors.
- **`app/...`**: the use cases, `alerting` and `identity`. They run the domain against the outside world through interfaces each declares where it uses them, as small as it needs: `alerting.Store` has three methods.
- **`adapter/...`**: what talks to the outside. `httpapi` (the generated server, its handlers and middleware), `postgres` (sqlc), `argon2id` and `ratelimit`. They satisfy the use cases' interfaces without naming them.
- **`platform/...`**: process plumbing with no business meaning: configuration, logging, HTTP servers, health, telemetry and schedules.

`cmd/api/main.go` is the only place that knows every concrete type, and it wires them by hand: no container, no globals. depguard fails the lint when the domain imports a use case, an adapter or the platform, or a use case imports an adapter or the platform.

## Consequences

- Use cases are tested with in-memory fakes of their own interfaces, in milliseconds. Adapters are tested against a real Postgres (ADR 0013).
- Replacing Postgres or the HTTP layer touches one adapter and `main.go`.
- The generated `oas.Alert` and the domain's `alert.Alert` are separate types, mapped in the handler, so the wire format can change without the domain noticing.

## Alternatives considered

- **A package per layer, with interfaces declared next to their implementations:** the shape of Java's layers, with interfaces nobody substitutes, and import cycles that end with everything in one package.
- **One flat package:** fine at this size, and no obvious place for the next feature to go.
