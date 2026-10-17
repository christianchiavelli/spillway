import { defineConfig, devices } from '@playwright/test';

/**
 * The whole stack, as a visitor meets it: the web app on E2E_BASE_URL, the Go
 * API and Postgres behind it. `docker compose up` starts all three; nothing
 * here is mocked.
 */
export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env['CI'],
  retries: process.env['CI'] ? 1 : 0,
  reporter: process.env['CI'] ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: process.env['E2E_BASE_URL'] ?? 'http://localhost:8090',
    trace: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
      testIgnore: /no-js\.spec\.ts/,
    },
    {
      // The same forms with JavaScript off: they post and redirect like
      // forms always have, and the server words every error.
      name: 'chromium, no JavaScript',
      use: { ...devices['Desktop Chrome'], javaScriptEnabled: false },
      testMatch: /no-js\.spec\.ts/,
    },
    {
      name: 'mobile',
      use: { ...devices['Pixel 7'] },
      testMatch: /alerts\.spec\.ts/,
    },
  ],
});
