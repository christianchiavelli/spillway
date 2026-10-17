# 5. Validation twice: the contract at the edge, the rules in the domain

- Status: accepted
- Date: 2026-10-17

## Context

A rule can be checked by the browser, the web server, the API's decoder, the domain or the database. Checked in one place only, it is skipped by every caller that does not pass there.

## Decision

- **At the API's edge, the contract.** ogen's decoder checks every request before any handler runs: types, required fields, lengths, ranges, enums, and unknown fields, which `additionalProperties: false` refuses. A failed check answers 422 with a pointer per field, in the domain's codes: `required`, `too_short`, `too_long`, `out_of_range`, `invalid`. A body is read up to 64 KiB, and a longer one is refused with 413.
- **In the domain, the rules.** Constructors check again, and check what a schema cannot say: a name is trimmed and its length counted in characters, not bytes; control characters are refused; an email is normalised before it is compared. Every violation is collected, not only the first.
- **In the web, the same schemas.** Forms parse with the Zod schemas generated from the contract, so a form the API would refuse is refused by the web server first, without a request, and in the same words.
- **In the database, the last line.** Check constraints and a unique index hold the same limits.
- **Tests hold the layers together.** A Go test checks that each limit in the contract equals the domain's constant, and that the decoder's errors map to the domain's codes.

## Consequences

- A request that skips the web, from `curl` or another client, meets the same rules.
- The rules are written in three places in code and one in SQL, with tests that keep them equal.

## Alternatives considered

- **The edge alone:** the domain would trust its callers, and the next caller (a job, a command) would skip the rules.
- **The domain alone:** the HTTP layer would decode anything, and every error pointer would be written by hand.
