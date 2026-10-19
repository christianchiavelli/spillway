# Spillway

Alerts on how much wind and solar energy Brazil's grid operator keeps from being generated. `README.md` says what it does and why, `docs/architecture.md` how it fits together; this file is how to work on it, for people and for coding agents alike.

## Layout

| Path | What it is |
| --- | --- |
| `contracts/openapi.yaml` | The API's contract, OpenAPI 3.1: both servers are generated from it (ADR 0002) |
| `services/api` | The Go API: `domain`, `app`, `adapter` and `platform` under `internal`, wired by hand in `cmd/api` (ADR 0003) |
| `apps/web` | The SvelteKit app: every page rendered on its server, the only thing a browser talks to (ADR 0001), in English and under `/pt-br` in Portuguese (ADR 0026) |
| `apps/web/messages` | Every text the app shows, `en.json` and `pt-BR.json`, compiled into `src/lib/paraglide`, which is not committed |
| `e2e` | Playwright against the running stack: desktop, a phone, and JavaScript off |
| `docs` | `architecture.md`, the decision records in `adr/`, the logo in `brand/` and the subsystems' map in `map/`, each with the script that draws it |

## Commands

- `docker compose up --wait --build`: the whole stack, healthchecked, at `http://localhost:8090`. Use `localhost`, not `127.0.0.1`: the app's origin is fixed at build time and a form posted from another one is refused (ADR 0015). The first start downloads the ONS's files, about 280 MB, in a minute or so; until the latest month is in, the data pages say so.
- `pnpm install`: Node 24.22.0 and pnpm 12.14.0 are pinned (`devEngines`, `packageManager`) and downloaded on the first install.
- `pnpm run ci`: format check, lint, types, unit and component tests in every package. Must pass before any change is done.
- `pnpm exec prettier --write <paths>`: from the repository root, which holds the ignore file.
- `pnpm --filter @spillway/web dev`: the web app on port 5173, against the API at `API_URL` (from `.env` in `apps/web`, e.g. `API_URL=http://localhost:8080`).
- `pnpm --filter @spillway/web run i18n`: compiles the messages after a change to them; `install`, `typecheck` and the build run it too.
- `pnpm --filter @spillway/web api:generate`: after any change to the contract. Commit what it writes under `src/lib/api/generated`; `pnpm run contract:check` fails while it is stale.
- `pnpm run e2e`: the browser checks, against the stack compose started, once the sync has loaded every month. The first time, `pnpm --filter @spillway/e2e exec playwright install chromium` fetches the browser.
- `pnpm run screenshots`: takes the README's screens again from the running stack, into `docs/screenshots`. Look at each before committing it.

In `services/api`:

- `go generate ./...`: after a change to the contract (ogen) or to a query or migration (sqlc). Commit what it writes; CI fails while it is stale.
- `go test -race ./...`: the unit tests.
- `go test -race -tags integration ./internal/adapter/postgres/...`: against a real Postgres 18, in Testcontainers (Docker must be running), or the empty database `TEST_DATABASE_URL` names.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./...`: the linters, the layer rules included.
- `go run ./cmd/api serve`, `go run ./cmd/api sync` and `go run ./cmd/api migrate`: with `DATABASE_URL` set. `LOG_FORMAT=text` reads better on a laptop, and `SYNC_INTERVAL=0` makes `sync` run once and exit.

## The API's configuration

| Variable | Default | What it sets |
| --- | --- | --- |
| `DATABASE_URL` | required | Postgres |
| `DATABASE_MAX_CONNS` | `10` | The pool's size |
| `API_ADDR` | `:8080` | The API |
| `ADMIN_ADDR` | `:9090` | `/livez`, `/readyz` and `/metrics` |
| `LOG_FORMAT`, `LOG_LEVEL` | `json`, `info` | slog's handler and level |
| `CORS_ALLOWED_ORIGINS` | empty | Browser origins allowed to call the API: none (ADR 0006) |
| `REQUEST_TIMEOUT` | `10s` | Each request's deadline |
| `SHUTDOWN_DRAIN`, `SHUTDOWN_TIMEOUT` | `0s`, `20s` | ADR 0012 |
| `PASSWORD_HASH_CONCURRENCY` | the CPUs | Hashes at once |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | unset | Traces over OTLP when set (ADR 0011) |
| `ONS_URL` | the ONS's bucket | Where `sync` lists and reads the files (ADR 0020) |
| `SYNC_INTERVAL` | `1h` | How often `sync` looks for new and revised files; `0` runs it once |

The configuration is read once, at start, and every bad value is reported together.

## Conventions

- Go 1.27 and the standard library first: `log/slog`, `net/http`, `uuid`, `crypto/rand.Text`, `errors.Is` and `errors.As`.
- In the API, dependencies point inwards and depguard enforces it. A use case declares the interfaces it needs; only `cmd/api` knows concrete types.
- Every error answer is an RFC 9457 problem with a `code` from the contract's `ProblemCode`. A new failure is a new code in the contract first (ADR 0004).
- A rule is checked at the edge by the contract and in the domain by its constructor, and a test keeps their limits equal (ADR 0005).
- In the web app, reads are load functions in `+page.server.ts` and writes are form actions; components never fetch (ADR 0014). Every form works without JavaScript (ADR 0010).
- Only `src/lib/server/api.ts` imports the generated client, which ESLint enforces, so every call has the request's token, id and timeout.
- Svelte 5 runes, strict TypeScript, and no `$effect` to keep state in sync.
- Tests sit next to what they test: `_test.go` in Go, `*.spec.ts` in TypeScript, `*.svelte.spec.ts` for components in a browser. Anything that needs Postgres runs against a real one, never a mock.
- Comments explain why, at the line that needs it. No comments that restate the code.
- Everything written in the repository is in English: code, docs, commit messages. The one exception is what the app shows, which is a message in both `messages/en.json` and `messages/pt-BR.json`, never a string in a component; links and redirects go through `localized()` (ADR 0026).
- Commits go straight to `main`, a title only: conventional, lowercase, imperative and short, like `fix(api): answer a wrong password with 400`.
- A decision with a trade-off gets a record in `docs/adr`, version pins included.

## Design changes

A new screen, or a visible change to one, starts as screenshots the owner approves, at a laptop's width and a phone's. Only then is it built, from the tokens in `app.css` and the components in `src/lib/components/ui`, and axe checks it in the end-to-end suite.
