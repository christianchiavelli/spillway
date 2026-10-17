# 16. Local only

- Status: accepted
- Date: 2026-10-17

## Context

Spillway is a portfolio piece. A deployment costs money every month and goes stale; a reviewer who wants to try it can run it.

## Decision

- **No cloud deployment and no monthly cost.** `docker compose up --wait --build` runs everything, with no account and no key.
- **Production concerns are built in anyway**: images from distroless and Alpine bases pinned by digest, health checks, graceful shutdown with a drain (ADR 0012), metrics and traces (ADR 0011), and secure cookies whenever the origin is HTTPS (ADR 0015).

## Consequences

- Anyone can run the whole system, and it does not rot behind an expired bill.
- The README has to show the app with screenshots, since many readers will not clone it.
- Nothing proves the stack on a cloud provider.

## Alternatives considered

- **A hosted demo:** easier to look at, and a standing cost for something that changes rarely.
