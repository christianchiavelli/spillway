# 18. Toolchains, dependencies and images pinned

- Status: accepted
- Date: 2026-10-17

## Context

A build that picks whatever is newest on the day it runs can fail, or change, without a commit to blame.

## Decision

- **Go 1.27.3**, on the `go` line of `go.mod`: an older Go downloads this one before it builds.
- **Node 24.22.0**, under `devEngines` in `package.json`, which pnpm downloads on the first install, and **pnpm 12.14.0**, under `packageManager` with its checksum.
- **One version of each JavaScript dependency**, exact, in the catalog of `pnpm-workspace.yaml`: a package names `catalog:`, so an upgrade is one line, reviewed once. pnpm waits a day before it installs a new release, and the few pinned while younger than that are excused there by name.
- **TypeScript 6.0**, which SvelteKit 3 asks for. typescript-eslint takes nothing newer, which keeps out the native TypeScript 7 for now.
- **Base images by digest**: Postgres 18.6, Go 1.27.3 and Node 24.22.0 on Alpine 3.24, and distroless static on Debian 13.
- **Actions by commit** (ADR 0017), and the CI tools by version: golangci-lint 2.15, govulncheck 1.9, oasdiff 1.33, sqlc 1.31.

## Consequences

- A build today and a build next month of the same commit get the same tools.
- Upgrades are commits of their own, and someone has to make them.

## Alternatives considered

- **Ranges and floating tags:** fresher by default, and a build that changes under a commit that did not.
