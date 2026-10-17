# 17. CI with least privilege and pinned actions

- Status: accepted
- Date: 2026-10-17

## Context

A workflow runs third-party code with a token that could write to the repository. A tag on an action can be moved to point at different code.

## Decision

- **The workflow's token can only read the repository.** No job publishes anything.
- **Every action is pinned to a full commit SHA**, with its version in a comment.
- **Five jobs, each answering one question:**
  - **api**: is the generated Go what the contract and the queries generate today, does it pass golangci-lint and the tests under the race detector, and does it call code with a known vulnerability (govulncheck)?
  - **api-integration**: does the Postgres adapter hold against a real Postgres 18, the alert limit under forty creates at once included?
  - **web**: format, lint, types, unit tests, component tests in Chromium, and the build.
  - **contract**: does the web's generated client match the contract, and does the contract break a client of the previous commit's (oasdiff)?
  - **e2e**: does the stack that compose builds work in a browser, on a laptop's screen, a phone's, and with JavaScript off?
- **The contract is compared with the previous commit** on a push, since changes go straight to `main`, and with the base on a pull request.

## Consequences

- A compromised or moved tag cannot change what runs.
- Pinned SHAs go stale; updating them is a deliberate change. Dependabot is not used, since it works through pull requests and this repository does not.
- A breaking contract change fails the push that makes it, after the fact: there is no pull request to stop it at.

## Alternatives considered

- **Actions by major tag:** less upkeep, and trust in whoever can move the tag.
