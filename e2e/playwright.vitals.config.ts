import { defineConfig } from '@playwright/test';

/**
 * The Core Web Vitals of the stack as it ships, on the phone and connection
 * of Lighthouse's mobile run, against the sums of the ONS's files as they were
 * synced once (`vitals/sums.sql.gz`). Each page's median is held against the
 * last accepted measurement (`vitals/baseline.json`).
 *
 *   pnpm run vitals:stack   the stack, with the recorded sums in place of the sync
 *   pnpm run vitals         measures every page on it
 *
 * One visit at a time, and nothing recorded on the side: a measurement
 * shares the machine with nothing, not even a trace of itself.
 */
export default defineConfig({
  testDir: './vitals',
  workers: 1,
  fullyParallel: false,
  forbidOnly: !!process.env['CI'],
  // A measurement retried until it passes measures nothing.
  retries: 0,
  timeout: 10 * 60_000,
  // Four times slower, on slow 4G, a page takes longer to answer than the suite's default allows.
  expect: { timeout: 20_000 },
  reporter: [[process.env['CI'] ? 'github' : 'list'], ['./vitals/reporter.ts']],
  use: {
    baseURL: process.env['E2E_BASE_URL'] ?? 'http://localhost:8090',
    // Chrome itself, headless, rather than the lighter shell the suite runs in: it paints the way Chrome does.
    channel: 'chromium',
    trace: 'off',
    screenshot: 'off',
    video: 'off',
    // A tap that finds nothing to tap fails within a minute, not at the end of the visit.
    actionTimeout: 60_000,
  },
});
