# 20. The ONS's data is synced by a process of its own

- Status: accepted
- Date: 2026-10-19

## Context

The ONS publishes curtailment as one file per source and month, Parquet among other formats, in a public S3 bucket: wind since October 2021, solar since April 2024, about 280 MB in all, with ten million half-hour records for wind alone. It rewrites the current month's files as each day is added, and rewrites older months when it revises them: on 18 October 2026 the September files changed, and in early October every solar month did.

The pages read a period's totals, its half hours or days, its subsystems and reasons, and its plants ranked. None of them needs a single record.

## Decision

- **A `sync` command of the API's binary, run as a service of its own** in compose, separate from `serve`. It is the only process that reaches the internet; the API still talks to Postgres alone. It answers `/livez`, `/readyz` and `/metrics` on the admin port, so compose waits on it like the others.
- **The bucket is listed, not the portal.** S3's ListObjectsV2 gives every file with its ETag in one request per dataset. A file is downloaded when its ETag differs from the one stored, so a run reads only what changed, and an older month the ONS revises is picked up like a new one. `SYNC_INTERVAL` sets the cadence, an hour by default.
- **The latest month first**, so a fresh stack shows the latest figures within seconds and fills its history behind them, in under a minute.
- **A file is summed as it is read and replaced whole.** parquet-go 0.32.0 reads the columns Spillway uses, the INT96 timestamps included, which the ONS writes as Brasília's wall clock. The domain sums each record into plant days, subsystem days, reason days and half hours, and one transaction deletes the month and copies the sums in, so a reader sees the month before or after, never half of each. Reads that take several queries run in one repeatable-read snapshot.
- **Strict about what cannot be, lenient about what is noise.** An unknown subsystem or reason, a time outside the file's month or off the half hour, a negative curtailment: the file is refused and logged, and tried again next run. The count of limited minutes, which runs past thirty in a few dozen of millions of records, is clamped to the half hour.
- **Days are calendar days in Brasília**, at a fixed offset of -03:00: Brazil has kept no summer time since 2019, before the first file. A day is shown once both sources have published it.

## Consequences

- The first start downloads every file, about 280 MB, and the data pages say the data is on its way until the latest month is in. The end-to-end suite waits for the whole history before it starts.
- Postgres holds sums, not records: about 340,000 plant days, 130,000 half hours and smaller tables, rebuilt from the files at any time by deleting `ons_files`.
- A question the sums cannot answer, such as one plant's half hours, needs a new table and a full sync.
- Two sync processes at once would do the same work twice, harmlessly: each month is replaced whole.

## Alternatives considered

- **The CKAN API of the ONS's portal:** it lists the same URLs without ETags, so every file would be downloaded every time.
- **DuckDB, in Go or as a sidecar:** it reads the Parquet files in one query, and needs cgo in a binary that is static on distroless today.
- **Apache Arrow's Go Parquet reader:** heavier, for no column type parquet-go cannot read.
- **The records themselves in Postgres:** every page would sum millions of rows on each request, for detail no page shows.
- **The sync inside `serve`:** one service fewer, and an API that needs a way out to the internet and runs the job once per replica.
