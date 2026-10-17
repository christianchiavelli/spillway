# 1. A SvelteKit server in front of a Go API

- Status: accepted
- Date: 2026-10-17

## Context

Spillway needs pages and an API that owns its rules and its data. The browser could call that API itself, or a server of the pages' own could stand between them.

## Decision

- **The browser talks to the SvelteKit server alone**, on one origin. Its server code (hooks, load functions and form actions) calls the Go API over HTTP, with the session's token as a bearer token.
- **The Go API owns the domain**: accounts, sessions and alerts, their rules and their storage. It knows nothing of pages, cookies or forms.
- **The SvelteKit server is a backend for the front end.** It keeps the session in a cookie, renders every page, and turns the API's problems into form errors and error pages.
- The API listens on a private network. Nothing publishes its port, and it refuses any request that comes from a browser (ADR 0006).

## Consequences

- The session token never reaches JavaScript, and the API needs no CORS for the app, since no browser ever calls it.
- Every page renders on the server and every form posts to it, so the app works before its script loads, or without it (ADR 0010).
- A request that needs data pays one more hop, on the private network.
- Two runtimes to build, test and ship. Compose and CI treat them as one stack.

## Alternatives considered

- **A single-page app that calls the API from the browser:** one server fewer, and a token within reach of any script, CORS on every route, and no page until the script has run.
- **SvelteKit alone, with its own database access:** one runtime, and nothing to show of Go, which is half of what the project is for.
