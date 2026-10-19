# 21. Firings are read from the data, not recorded

- Status: accepted
- Date: 2026-10-19

## Context

An alert fires on each day its source, in its subsystem, loses more energy than its threshold. The alerts page shows when each one last fired, how many of the last 30 days it fired on, and every firing in those days; the new alert form previews the same count before the alert exists.

The ONS revises days after it publishes them (ADR 0020), and Spillway sends nothing: no email, no push.

## Decision

- **A firing is a query, not a row.** The alerts join the subsystem days where the energy lost is more than the threshold; as much does not fire. The rule lives in three queries in `queries/alerts.sql`, and an integration test holds them to it, the equal case included.
- **The days are the 30 up to the latest published**, the window the preview counts too, so the form and the page agree on what an alert would catch.
- **A firing follows the data**: when the ONS revises a day, its firings change with it, and the API's contract says so.

## Consequences

- Nothing to schedule and nothing to keep in step: a new alert shows its firings at once, and a revision never leaves a firing the data no longer supports.
- An alert fires on days before it was made. Without a delivery, a firing says what the data shows, not that anyone was told.
- Sending a message for each firing would need what this leaves out: a record of each one, made once, with the figure it was made from, so a revision cannot send it twice.

## Alternatives considered

- **A table of firings, written by the sync after each run:** the shape a delivery would need, and with nothing to deliver, a second copy of the data that a revision makes wrong.
- **Firings counted from the day an alert is made:** honest about time, and an empty page for every new alert until the next day is published.
