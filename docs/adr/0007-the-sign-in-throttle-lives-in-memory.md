# 7. The sign-in throttle lives in memory

- Status: accepted
- Date: 2026-10-17

## Context

A password can be guessed by asking often enough. Signing in has to slow down whoever keeps getting it wrong, without slowing down anyone else.

## Decision

- **Per email address, as normalised:** five attempts at once, then one more every twelve seconds. Each address has a token bucket from `golang.org/x/time/rate`, held in a map behind a mutex (`adapter/ratelimit`).
- **A refused attempt spends nothing**, so waiting is enough to try again. A successful sign-in empties the address's record.
- **The answer to a throttled attempt is 429** with code `rate_limited` and a `Retry-After` header, which the sign-in page turns into "try again in N seconds".
- **Buckets that are full again are swept every minute**, so memory grows with the addresses under attack, not with the people who sign in.
- **By address, not by IP:** guessing targets an account, and IP addresses are cheap to change and shared behind every office's NAT.

## Consequences

- Each API replica counts on its own: with N replicas, an attacker gets N times the rate. Compose runs one.
- A restart forgets every count.
- Someone can keep an address from signing in by failing on purpose, for as long as they keep at it, twelve seconds at a time.

## Alternatives considered

- **Counters in Postgres or Redis:** shared by every replica, for one more write per attempt, and one more service for Redis. `identity.Limiter` is the seam: an adapter over Postgres would replace this one without touching the use case.
- **By IP:** stops one machine, misses a botnet, and locks out an office.
