# 29. The workflow is audited too

- Status: accepted, extending [ADR 0017](0017-ci-least-privilege-pinned-actions.md)
- Date: 2026-10-20

## Context

ADR 0017 keeps the workflow's token read-only and pins every action to a commit. Nothing checked that the workflow kept to it: a new job could check out the code with the token saved in `.git`, an expression could put an untrusted value into a script, or a pin could name a commit that is not in the action's repository at all, which runs a stranger's code under the action's name.

## Decision

- **A `workflows` job, first in the file**, lints and audits `.github/workflows` on every push.
- **actionlint 1.7.12** checks the workflow itself: its syntax, its expressions and the job graph, and every `run` script through the runner's shellcheck.
- **zizmor 1.31.1** audits it for security: template injection, excessive permissions, credentials persisted by a checkout, and, with the job's token, actions pinned to impostor commits or to versions with a known advisory.
- **No checkout persists the token.** zizmor found all five keeping it in `.git`, where a later step, or an artifact uploaded from the workspace, could read it; each now sets `persist-credentials: false`.
- Both run at a pinned version, from the Go module proxy and PyPI, as govulncheck and oasdiff already do.

## Consequences

- A workflow change that weakens ADR 0017 fails its own push.
- The tools' versions are pinned like everything else, and updated by hand.
- zizmor's online audits ask GitHub's API with the job's read-only token.

## Alternatives considered

- **GitHub's code scanning with zizmor's SARIF:** the findings as alerts in the Security tab, for a token that can write security events, which no job here has.
- **A pre-commit hook:** checks a laptop that has the tools, not every push.
