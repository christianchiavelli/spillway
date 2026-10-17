# 4. Errors are values inside, RFC 9457 problems at the edge

- Status: accepted
- Date: 2026-10-17

## Context

An error crosses three layers on its way to a person: the domain names it, the API answers with it, and the web shows it next to the field it is about. Each layer needs to tell causes apart without reading messages meant for people.

## Decision

- **The domain names failures** as sentinels (`domain.ErrNotFound`, `ErrConflict`, `ErrUnauthenticated`) and as types that carry what the caller needs: `ValidationError` with every violation, `ConflictError` with the field, `identity.ThrottledError` with the wait.
- **Use cases wrap with `%w`** and say what they were doing (`add alert: %w`). Nothing panics for control flow, and a panic that happens anyway is recovered, logged with its stack and answered as a 500.
- **One function, `httpapi.NewError`, maps an error to an answer**, with `errors.Is` and `errors.As`, never by its text. An error it does not recognise is logged and answered as a bare 500 with code `internal`: its cause never reaches the client.
- **Every answer that is not a success is an RFC 9457 problem** (`application/problem+json`) with two extensions: `code`, a stable name from a closed enum in the contract, and `errors`, a list of `{pointer, code}` whose RFC 6901 pointers name the fields of the request.
- **Status codes:**

| Status | Code                              | When                                                |
| ------ | --------------------------------- | --------------------------------------------------- |
| 400    | `malformed_request`               | A body or a path the operation cannot read          |
| 400    | `invalid_credentials`             | A wrong email or password                           |
| 401    | `unauthenticated`                 | No session, or one that is over or unknown          |
| 403    | `forbidden_origin`                | A request from a browser (ADR 0006)                 |
| 404    | `not_found`                       | Nothing there, or nothing that is yours             |
| 405    | `method_not_allowed`              | With an `Allow` header                              |
| 409    | `conflict`, `alert_limit_reached` | A name already taken, the twenty-first alert        |
| 413    | `request_too_large`               | A body over 64 KiB                                  |
| 415    | `unsupported_media_type`          | A body that is not JSON                             |
| 422    | `validation_failed`               | A body that breaks a rule, with a pointer per field |
| 429    | `rate_limited`                    | With `Retry-After` (ADR 0007)                       |
| 500    | `internal`                        | Anything else, logged                               |

- **Wrong credentials answer 400**, as an OAuth token endpoint answers `invalid_grant`, so that 401 means one thing: sign in again. Every 401 carries `WWW-Authenticate: Bearer`, as RFC 9110 requires, with `error="invalid_token"` when a session came and was refused (RFC 6750).
- **Another account's alert is not found**, not forbidden: the answer says nothing about what exists.
- A request its client abandoned is logged at debug level: it is not the server's failure.

## Consequences

- The web switches on `code`, never on a message, and puts each field's error next to its input.
- A message can be reworded without breaking a client.
- Adding a code is a contract change, checked by CI like any other (ADR 0002).

## Alternatives considered

- **An error envelope of our own:** one more format to document. RFC 9457 is the one HTTP tooling already knows.
- **Status codes alone:** a 409 cannot say which field, or why.
- **401 for wrong credentials:** common, and a client that answers every 401 by signing in again would loop on the sign-in itself.
