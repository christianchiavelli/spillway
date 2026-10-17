# 14. Load functions and form actions, not remote functions

- Status: accepted
- Date: 2026-10-17

## Context

SvelteKit 3 has two ways to reach the server from a page: load functions and form actions, stable since SvelteKit 1, and remote functions (`query`, `form`, `command`), still behind an experimental flag.

## Decision

- **Pages read in `+page.server.ts` and `+layout.server.ts` load functions, and write through form actions.** Every read needs the session cookie, and the API is not reachable from a browser, so no load is universal: there is no `+page.ts`.
- **Each route's data and action results are typed by its generated `$types`**, so a page reads what its server code returned without a cast.
- **Components take props and raise events.** They never fetch: `src/lib/components` draws, `src/lib/features` knows the domain's words, and the routes orchestrate.

## Consequences

- A route's reads and writes are in one file, and every write works without JavaScript (ADR 0010).
- When remote functions are stable, a `form` could replace an action without changing what the page does.

## Alternatives considered

- **Remote functions now:** shorter code, and an experimental API on the path of every form.
