# 15. The app's origin is fixed at build time

- Status: accepted
- Date: 2026-10-17

## Context

SvelteKit refuses a form whose `Origin` header differs from the app's own origin (ADR 0006). adapter-node 6 no longer reads that origin from an `ORIGIN` variable at run time. Unless a proxy says otherwise, through `PROTOCOL_HEADER` and `HOST_HEADER`, it takes every request for HTTPS, so on plain `http://localhost` every form was refused with 403.

## Decision

- **The build sets `paths.origin` from `ORIGIN`**, `http://localhost:8090` by default, the one address compose publishes. The dev server keeps its own.
- **The image takes it as a build argument**: `docker build --build-arg ORIGIN=https://… -f apps/web/Dockerfile .`
- **Cookies follow the origin.** Over HTTP they carry neither `Secure` nor the `__Host-` prefix, which no browser would accept there; over HTTPS they carry both.

## Consequences

- An image serves one address. Another address is another build.
- `http://127.0.0.1:8090` is another origin: every form posted from it is refused, by design.

## Alternatives considered

- **A reverse proxy, such as Caddy, sending `X-Forwarded-Proto` and `X-Forwarded-Host`:** the origin at run time, for one more service. It is the right shape once TLS ends at a proxy.
