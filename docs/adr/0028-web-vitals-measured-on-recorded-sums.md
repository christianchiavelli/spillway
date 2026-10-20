# 28. Web Vitals, measured on recorded sums

- Status: accepted
- Date: 2026-10-20

## Context

CI proved that every page works, reads well to a screen reader and keeps its links, and said nothing of how fast it is. A page could put on a second of load, a shift or a slow interaction, and every check would pass. Larder and Faultline measure their Core Web Vitals in CI; Spillway, whose pages draw charts of up to five years of data on the server, did not.

Measured on the live stack, the numbers would move with the data: the sync loads whatever the ONS last published, the latest day changes every morning, and a run downloads 230 MB of Parquet first.

## Decision

- **A `vitals` job measures LCP, CLS and INP** with web-vitals 6.2.3 and its attribution build, in Chrome on the phone and connection of Lighthouse's mobile run: a Moto G Power's screen, a CPU four times slower than the runner's, slow 4G.
- **Four pages, five visits each, as a reader makes them:** the overview, the plants, the history, and a new alert behind the sign-in. Each visit arrives cold, waits for the page to hydrate, reads it to the end a screen at a time, then uses it: a chart switched to its table, the next period or page, a threshold typed a key at a time while the preview redraws. LCP is read before the first scroll, CLS and INP as the visit ends.
- **Each metric's median is held against a baseline in the repository** (`e2e/vitals/baseline.json`), and the job fails when one rises past its tolerance: a tenth of LCP or 100 ms, 0.01 of CLS, three tenths of INP or 16 ms, above the noise of GitHub's runners. The table lands in the run's summary, with what each number is about: the element painted, shifted or touched, and where its time went.
- **A baseline comes from CI, never from a laptop**: `pnpm run vitals:accept` takes the median of one or more runs' measurements, writes the baseline and the README's table, and the change is committed like any other. Off CI a measurement is compared, not judged.
- **The stack runs on recorded sums.** `e2e/vitals/sums.sql.gz`, 5 MB, holds the curtailment tables the sync had built from every file on 20 October 2026; a compose file laid over the usual one leaves the sync off and restores them before the API starts. Every run measures the same five years, the same latest day and the same plants, and nothing on the runner reaches the ONS. `pnpm run vitals:record` takes a new recording from a synced stack.

## Consequences

- A change that slows a page fails the push that makes it, and says which element and which part of its time.
- The recording is 5 MB in the repository, and 5 MB more each time it is taken again. It restores into the schema as the migrations leave it, so a migration that changes these tables needs a new recording, and a new baseline after it.
- The pages are measured on data that grows old. Their shape does not change with the date, and the e2e job still runs on the ONS's latest files.
- The job builds the stack a second time, beside the e2e job, and takes about ten minutes.

## Alternatives considered

- **Measuring in the e2e job, on its synced data:** numbers that move with each day the ONS publishes, and a measurement sharing the machine with the sync.
- **A recording of the ONS's Parquet files, served by a stub of the bucket:** the sync's own path, but 230 MB for the history, or a history page with two months on it.
- **Lighthouse CI:** a score per page from one load, with no interaction, so no INP, and less to say about what moved.
- **Fixed budgets in place of a baseline:** a limit set once, far above what the pages do, catches nothing until the page is already slow.
