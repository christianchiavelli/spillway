import { randomUUID } from 'node:crypto';

import { type BrowserContextOptions, expect, test } from '@playwright/test';

import { measure, settle, tap, typeInto, type Visit } from './measure.ts';
import {
  ATTACHMENT,
  explain,
  format,
  limitOf,
  type Measurement,
  METRICS,
  readBaseline,
  summarise,
} from './report.ts';

/**
 * Each page visited five times, the way a reader uses it on a phone: loaded
 * cold, read to the end, then put to work. The median of each metric is held
 * against the baseline, and a page that has none yet is measured, not judged.
 * The stack runs on the sums of the ONS's files as they were synced once
 * (`sums.sql.gz`), so a page is timed on the same data every run.
 */
const VISITS = 5;

const PAGES: readonly (Visit & { readonly signedIn?: boolean })[] = [
  {
    key: 'overview',
    name: 'The overview',
    path: '/',
    journey: async (page) => {
      // The day's half hours as a table, then the week.
      await tap(
        page.getByRole('region', { name: 'Hour by Hour' }).getByRole('radio', { name: 'Table' }),
      );
      await tap(
        page.getByRole('navigation', { name: 'Period' }).getByRole('link', { name: 'Week' }),
      );
      await expect(page).toHaveURL(/\/\?period=week$/);
      await expect(page.getByRole('heading', { name: 'Day by Day' })).toBeVisible();
    },
  },
  {
    key: 'plants',
    name: 'The plants',
    path: '/plants',
    journey: async (page) => {
      // The ranking narrowed to wind, then its second page.
      await tap(
        page.getByRole('navigation', { name: 'Source' }).getByRole('link', { name: 'Wind' }),
      );
      await expect(page).toHaveURL(/source=wind/);
      await tap(page.getByRole('link', { name: 'Next', exact: true }));
      await expect(page).toHaveURL(/page=2/);
    },
  },
  {
    key: 'history',
    name: 'The history',
    path: '/history',
    journey: async (page) => {
      // Every month since October 2021, as a table.
      await tap(
        page
          .getByRole('region', { name: 'Curtailed Each Month' })
          .getByRole('radio', { name: 'Table' }),
      );
      await expect(page.getByRole('rowheader', { name: '2025', exact: true })).toBeVisible();
    },
  },
  {
    key: 'new-alert',
    name: 'A new alert',
    path: '/alerts/new',
    signedIn: true,
    journey: async (page) => {
      // Solar in the Northeast, then a threshold typed a key at a time: the
      // preview of the last 30 days redraws at each.
      await tap(page.getByRole('radio', { name: 'Solar' }));
      await page.getByLabel('Subsystem').selectOption({ label: 'Northeast' });
      await settle(page);
      await typeInto(page.getByLabel('Daily threshold'), '50000');
      await expect(page.getByText(/^It would have fired on /)).toBeVisible();
    },
  },
];

const baseline = readBaseline();
// The baseline is the CI runners' own: another machine's numbers are set against it, not judged by it.
const judged = !!process.env['CI'];

// One account for the pages behind the sign-in, made before any is measured.
let session: BrowserContextOptions['storageState'];
test.beforeAll(async ({ browser, baseURL }) => {
  const context = await browser.newContext({ baseURL });
  const page = await context.newPage();
  await page.goto('/sign-up');
  await page.getByLabel('Email address').fill(`vitals-${randomUUID()}@example.com`);
  await page.getByLabel('Password').fill('correct horse battery staple');
  await page.getByRole('button', { name: 'Create Account' }).click();
  await expect(page).toHaveURL(/\/alerts$/);
  session = await context.storageState();
  await context.close();
});

for (const { journey, signedIn, ...page } of PAGES) {
  test(`${page.name}, ${page.path}`, async ({ browser, baseURL }, testInfo) => {
    const visits: Measurement[] = [];
    for (let visit = 1; visit <= VISITS; visit++) {
      visits.push(
        await test.step(`visit ${String(visit)} of ${String(VISITS)}`, () =>
          measure(browser, baseURL ?? '', { ...page, journey }, signedIn ? session : undefined)),
      );
    }

    const result = summarise(page, visits);
    await testInfo.attach(ATTACHMENT, {
      body: JSON.stringify(result),
      contentType: 'application/json',
    });

    const base = baseline.pages[page.key];
    if (!base || !judged) {
      testInfo.annotations.push({
        type: 'notice',
        description: base
          ? 'Measured off CI: compared, not judged'
          : 'No baseline yet: measured, not judged',
      });
      return;
    }
    for (const metric of METRICS) {
      const value = result.median[metric];
      const typical = visits.find((visit) => visit[metric].value === value)?.[metric];
      const about = typical ? `, ${explain(metric, typical)}` : '';
      expect
        .soft(
          value,
          `${metric} ${format(metric, value)} against a baseline of ${format(metric, base[metric])}${about}`,
        )
        .toBeLessThanOrEqual(limitOf(metric, base[metric]));
    }
  });
}
