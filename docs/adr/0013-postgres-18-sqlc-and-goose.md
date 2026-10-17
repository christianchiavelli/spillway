# 13. Postgres 18, sqlc and goose

- Status: accepted
- Date: 2026-10-17

## Context

The API stores accounts, sessions and alerts, with rules the database has to hold (ADR 0008). The code that reads and writes them should show the SQL that runs.

## Decision

- **Postgres 18.** Accounts and alerts are keyed by UUIDv7, made by the API with Go 1.27's `uuid.NewV7`, so keys sort by creation and new rows land at the end of each index. A session is keyed by its token's hash (ADR 0006). Times are `timestamptz`, and leave the adapter in UTC.
- **sqlc 1.31.** Queries are SQL files in `queries/`, checked against the migrations when the code is generated, and turned into typed Go over pgx v5. No ORM and no query builder.
- **goose 3.29.** Migrations are plain SQL, embedded in the binary. `api migrate` applies them under a Postgres advisory lock, so two runs at once take turns, and compose runs it as a job of its own before the API starts.
- **One pgx pool per process**, ten connections by default (`DATABASE_MAX_CONNS`), with every query traced (ADR 0011).
- **Tests run against a real Postgres 18**, in Testcontainers, or the one `TEST_DATABASE_URL` names. Never a mock of the database.

## Consequences

- The SQL in review is the SQL that runs, and a query that no longer fits the schema fails at generation, not in production.
- The API's server could run as a role that cannot change the schema, since it never migrates.
- sqlc runs at a pinned version through `go run`, not as a tool in `go.mod`, so its parsers stay out of the API's module graph.

## Alternatives considered

- **GORM or ent:** reflection or a DSL between the reader and the query that runs.
- **golang-migrate:** much the same. goose's provider takes the embedded files and the lock as values, with no global state.
- **Migrations as the server starts:** one service fewer, and every replica needs the right to change tables.
