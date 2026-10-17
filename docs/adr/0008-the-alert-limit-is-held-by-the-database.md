# 8. The alert limit is held by the database

- Status: accepted
- Date: 2026-10-17

## Context

An account holds at most twenty alerts. Two requests at once, each counting nineteen, would both insert, unless something makes the second count wait for the first insert.

## Decision

- **`Alerts.Add` is one transaction:** lock the account's row with `SELECT … FOR NO KEY UPDATE`, count its alerts, and insert only under the limit.
- **`FOR NO KEY UPDATE`, not `FOR UPDATE`:** creators for one account queue on it, while it leaves alone the key-share locks that foreign key checks take on the same row.
- **READ COMMITTED is enough**, since the count runs after the lock is held and sees every alert committed before it.
- **Names are unique per account whatever their case**, by a unique index on `(account_id, lower(name))`. The adapter turns its violation into `alert.ErrNameTaken` by the constraint's name, not by the message.
- **An integration test** fires forty creates at once for one account, against a real Postgres, and expects exactly twenty stored and twenty refused with `alert.ErrLimitReached`, which the API answers as 409 `alert_limit_reached`.

## Consequences

- The limit holds with any number of API replicas.
- Creates for one account run one at a time, which is no cost at the speed people create alerts.

## Alternatives considered

- **Count in the use case, then insert:** the race this record is about.
- **SERIALIZABLE:** correct, and every caller has to retry on serialisation failures.
- **A trigger:** the rule would hide in the database, away from the code that reads like the rule.
- **A counter column with a check constraint:** two writes per create, and a count that can drift from the rows.
