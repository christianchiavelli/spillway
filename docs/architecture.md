# Architecture

How Spillway is put together, and the path one request takes through it. Why each piece is the way it is lives in the [decision records](adr/README.md).

## The pieces

```text
browser ──HTML forms──▶ web (SvelteKit 3, Node 24) ──JSON over HTTP──▶ api (Go 1.27) ──pgx──▶ Postgres 18
          one origin,      renders every page,          bearer token,        rules, sessions,      ▲
          one cookie       keeps the session cookie     private network      storage              │ pgx
                                                                                                   │
the ONS's S3 bucket ◀──HTTPS, ETags── sync (`api sync`, the same binary) ─────────────────────────┘
   a Parquet file per                  sums each new or revised file,
   source and month                    and replaces its month whole
```

- The browser only ever talks to the web server, on one origin (ADR 0001).
- The web server only ever talks to the API, with the session's token as a bearer token. Nothing publishes the API's port.
- The API only ever talks to Postgres.
- The sync service is the only one that reaches the internet, to read the ONS's files every hour (ADR 0020).
- `contracts/openapi.yaml` describes the one interface between the two servers, and both sides are generated from it (ADR 0002).

## The repository

```text
spillway/
├── contracts/
│   └── openapi.yaml            the API: every request, answer, limit and error code
├── services/api/               Go module: the API
│   ├── cmd/api/                main: `serve`, `sync`, `migrate`, `healthcheck`; wires everything by hand
│   ├── internal/
│   │   ├── domain/             entities and rules, no I/O
│   │   │   ├── errors.go       sentinels, ValidationError, ConflictError, violation codes
│   │   │   ├── account/        email and password rules, sessions and their tokens
│   │   │   ├── alert/          what makes an alert, the limit per account, firings
│   │   │   └── curtailment/    sources, subsystems, reasons, periods, a month's sums
│   │   ├── app/                use cases, each with the interfaces it needs
│   │   │   ├── alerting/       create, list and delete alerts, and read their firings
│   │   │   ├── identity/       sign up, sign in, authenticate, sign out, purge
│   │   │   ├── ingest/         load the ONS's new and revised files, the latest month first
│   │   │   └── reports/        a period's overview, the plants ranked, the months, the latest days
│   │   ├── adapter/            the outside world, behind the use cases' interfaces
│   │   │   ├── httpapi/        the server ogen generated (oas/), handlers, problems, middleware
│   │   │   ├── postgres/       migrations/, queries/, sqlc's sqlcdb/, and the stores
│   │   │   ├── ons/            the ONS's bucket: its listing, and its Parquet files
│   │   │   ├── argon2id/       password hashing, a bounded number at a time
│   │   │   └── ratelimit/      the sign-in throttle, a token bucket per address
│   │   └── platform/           config, logging, httpserver, health, telemetry, schedule
│   ├── sqlc.yaml
│   ├── .golangci.yml           the linters, and the layer rules (depguard)
│   └── Dockerfile              a static binary on distroless
├── apps/web/                   SvelteKit 3: the pages
│   ├── src/
│   │   ├── hooks.ts            takes /pt-br off an address before routing (ADR 0026)
│   │   ├── hooks.server.ts     request id, the edition, security headers, the session on every request
│   │   ├── env.ts              API_URL, checked when the server starts
│   │   ├── app.css             the design system's tokens and type styles
│   │   ├── routes/             the overview (/), plants, history, alerts and a new one, signing in and out
│   │   └── lib/
│   │       ├── api/generated/  Hey API's types, client and Zod schemas: not edited by hand
│   │       ├── server/         server only: the API client, session and flash cookies, form readers
│   │       ├── forms/          what every form does: field errors, and use:enhance's additions
│   │       ├── features/       the domain's words for the screen: curtailment, alerts, account
│   │       └── components/     ui/ (fields, buttons, tables, toasts), charts/ (SVG) and layout/
│   ├── messages/               every text the app shows, en.json and pt-BR.json, compiled by Paraglide
│   ├── vite.config.ts          SvelteKit's options (CSP, CSRF, origin), Paraglide, and the Vitest projects
│   └── Dockerfile              adapter-node's output on Node
├── e2e/                        Playwright against the running stack: desktop, phone, no JavaScript
├── docs/                       this file, the decision records in adr/, the logo and the map
├── compose.yaml                db, migrate, api, sync and web
└── .github/workflows/ci.yml
```

## Layers, and which way they point

In the API, dependencies point inwards, and depguard fails the lint when they do not (ADR 0003):

| Package | May import | Holds |
| --- | --- | --- |
| `domain/...` | the standard library, `domain` | What an alert, an account and a session are, and their rules |
| `app/...` | `domain/...` | The use cases, and the interfaces they need from the outside |
| `adapter/...` | `app/...`, `domain/...`, `platform/...` | HTTP, Postgres, hashing, throttling |
| `platform/...` | the standard library and libraries | Configuration, logs, servers, health, telemetry |
| `cmd/api` | everything | The wiring, and nothing else |

In the web app, the same idea in SvelteKit's terms (ADR 0014):

| Folder | Does | Never |
| --- | --- | --- |
| `routes/` | Reads in load functions, writes in form actions, picks components | Draws more than the page's layout |
| `lib/server/` | Talks to the API and to cookies | Runs in a browser: SvelteKit refuses to bundle it |
| `lib/features/` | Knows the domain's words: labels, messages, formats, limits, in both editions | Fetches |
| `lib/components/` | Draws, from props, and calls back | Knows what an alert is |

## One request, end to end: creating an alert

Someone fills in the form on `/alerts` and presses Create Alert.

| Step | Where | What happens |
| --- | --- | --- |
| 1 | `routes/alerts/+page.svelte` | A `<form method="POST" action="?/create">`, which posts without JavaScript. With it, `enhanceForm` blocks a second submit and shows the pending state. |
| 2 | `hooks.server.ts` | The request gets an id, and the edition its address names: `/pt-br/alerts` is the same route in Portuguese, and every message and link the request renders follows it. The session cookie becomes `locals.session` after the API confirms it (`GET /v1/sessions/current`), and its deadline moves forward. |
| 3 | `routes/alerts/+page.server.ts` | The `create` action reads the form, and parses it with `zNewAlert`, generated from the contract. A broken rule answers 422 here, with the values as typed and an error per field, without calling the API. |
| 4 | `lib/server/api.ts` | A client for this request alone: the bearer token, `X-Request-Id`, a five-second timeout. The answer is parsed with `zAlert` before anyone reads it. |
| 5 | `httpapi/router.go` | Middleware, in order: request id, access log, panic recovery, API headers, origin check, body limit, deadline. |
| 6 | `httpapi/oas` | ogen routes `POST /v1/alerts`, calls `HandleSession` with the token, and decodes the body against the contract: a broken rule answers 422 with a JSON pointer per field. |
| 7 | `httpapi/handler.go` | `CreateAlert` turns the request into an `alert.Draft` and calls the use case with the account from the context. |
| 8 | `app/alerting` | `Create` gives the alert a UUIDv7 and asks the domain whether the draft is an alert. |
| 9 | `domain/alert` | `alert.New` trims the name, counts its characters, refuses control characters, checks the source, subsystem and threshold, and collects every violation. |
| 10 | `adapter/postgres` | `Add` runs one transaction: lock the account's row, count its alerts, insert under twenty. A name taken in any case trips a unique index and becomes `alert.ErrNameTaken` (ADR 0008). |
| 11 | `httpapi/problem.go` | Success answers 201 with the alert. A failure becomes a problem: 409 `conflict` pointing at `/name`, 409 `alert_limit_reached`, 422 `validation_failed` (ADR 0004). |
| 12 | `routes/alerts/+page.server.ts` | Success sets a flash cookie and redirects with 303, so a reload cannot post again; the list loads afresh and a toast says the alert was created. A problem becomes field errors through its pointers, and the page comes back with them. |
| 13 | `lib/forms/enhance.ts` | With JavaScript, after a refusal, focus moves to the error summary, which links to each field. After success, the form is emptied. |

The other operations take the same path: the contract, the generated decoder, a handler, a use case, the domain, an adapter.

## The data, from the ONS to a page

| Step | Where | What happens |
| --- | --- | --- |
| 1 | `cmd/api`, `sync` | Each hour, and once at start, the ingest use case runs. |
| 2 | `adapter/ons` | Lists both datasets in the ONS's S3 bucket, one request each, and keeps the monthly Parquet files with their ETags. |
| 3 | `app/ingest` | Skips every file whose ETag is the one stored, and takes the rest the latest month first. |
| 4 | `adapter/ons`, `domain/curtailment` | Downloads a file and reads it with parquet-go into a month builder, which checks each record and sums it into plant days, subsystem days, reason days and half hours. |
| 5 | `adapter/postgres` | One transaction: the plants upserted, the month's rows deleted, the sums copied in, the file's ETag recorded. A file that fails is logged and left for the next run. |
| 6 | `routes/+page.server.ts` | `/?period=week&date=2026-10-14` asks `GET /v1/curtailment` and `GET /v1/plants`, public operations that need no session. |
| 7 | `app/reports`, `domain/curtailment` | Works out the period from how far the data goes: the week from Monday that holds the day, up to the last day published, and as many days of the week before to compare with. |
| 8 | `adapter/postgres` | Reads the totals, the series, the subsystems, the reasons and the plant counts in one repeatable-read snapshot. |
| 9 | `features/curtailment`, `components/charts` | Words the period, and draws its columns, bars and the subsystems' map in SVG on the server, each with a switch to its figures as a table (ADR 0022, ADR 0024, ADR 0025). |

An alert's firings are a query over the same subsystem days, not rows of their own (ADR 0021).

## Conventions

**Go**

- Errors are values: sentinels and typed errors in the domain, wrapped with `%w` on the way up, told apart with `errors.Is` and `errors.As`, mapped to a status in one place (`httpapi.NewError`). No panics for control flow.
- Every function that does I/O takes a `context.Context` first, and passes it down. Background work gets its own context from `main`, never `context.Background()` in the middle of a call.
- Interfaces are declared by the package that uses them, with only the methods it calls. Constructors return concrete types.
- Goroutines only where something runs on its own: the two servers, the session purge and the throttle's sweep, in one errgroup; in `sync`, the admin server and the hourly run, in another. The password hasher bounds its own concurrency with a channel used as a semaphore.
- `log/slog` only, with the request's attributes taken from the context.
- Tests sit next to the code as `_test.go`. Integration tests carry the `integration` build tag and need a Postgres.

**SvelteKit**

- Svelte 5 runes only: `$props` for inputs, `$derived` for anything computed, `$state` for what a component owns. There is no `$effect` in the app: nothing is kept in sync by hand.
- A load function reads only what it needs, so it reruns only when that changes: the layout reads the URL's protocol through `untrack`, since tracking the URL would rerun it on every navigation.
- Strict TypeScript, `exactOptionalPropertyTypes` aside, which the generated client does not satisfy.
- Every form works without JavaScript (ADR 0010).
- Every component is checked by axe in the end-to-end suite, and every field has a visible label, a described error and a focus ring (ADR 0009).

**Both**

- The contract changes first, then `go generate ./...` and `pnpm --filter @spillway/web api:generate`, then the code. CI fails when generated code is stale.
- Times travel as RFC 3339 in UTC and are shown in Brasília's time zone, the grid operator's.
- Comments say why, at the line that needs it.
