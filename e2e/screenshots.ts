import { mkdir } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

import { type Browser, chromium, expect, type Page } from '@playwright/test';

/*
 * The README's screens, taken from the stack `docker compose up --wait --build`
 * started: `pnpm run screenshots`, or `pnpm run screenshots <name>…` for some.
 * They are taken as Ana, an account of the script's own at example.com, made
 * the first time and given the same five alerts every time.
 */

const base = process.env['E2E_BASE_URL'] ?? 'http://localhost:8090';
const out = fileURLToPath(new URL('../docs/screenshots/', import.meta.url));

const LAPTOP = { width: 1440, height: 900 };
const PHONE = { width: 390, height: 844 };
const ANA = { email: 'ana@example.com', password: 'correct horse battery staple' };

/** Where and what the grid curtails most, at thresholds of the right size. */
const ALERTS = [
  { name: 'Northeast wind', source: 'Wind', subsystem: 'Northeast', threshold: '12000' },
  { name: 'Northeast solar', source: 'Solar', subsystem: 'Northeast', threshold: '4500' },
  {
    name: 'Southeast solar',
    source: 'Solar',
    subsystem: 'Southeast and Centre-West',
    threshold: '1200',
  },
  { name: 'South wind', source: 'Wind', subsystem: 'South', threshold: '800' },
  { name: 'North wind', source: 'Wind', subsystem: 'North', threshold: '300' },
] as const;

async function open(browser: Browser, viewport = LAPTOP) {
  const context = await browser.newContext({
    baseURL: base,
    viewport,
    locale: 'en-US',
    deviceScaleFactor: 1,
  });
  return { context, page: await context.newPage() };
}

/** Signs Ana in, making her account the first time. */
async function signIn(page: Page) {
  await page.goto('/sign-in');
  await page.getByLabel('Email address').fill(ANA.email);
  await page.getByLabel('Password').fill(ANA.password);
  await page.getByRole('button', { name: 'Sign In' }).click();
  const list = page.getByRole('heading', { name: 'Alerts', level: 1 });
  await expect(list.or(page.getByRole('alert'))).toBeVisible();
  if (await list.isVisible()) return;

  await page.goto('/sign-up');
  await page.getByLabel('Email address').fill(ANA.email);
  await page.getByLabel('Password').fill(ANA.password);
  await page.getByRole('button', { name: 'Create Account' }).click();
  await expect(list).toBeVisible();
}

/** Leaves Ana with the five alerts, or with none. */
async function alerts(page: Page, wanted: 'five' | 'none') {
  const deletes = page.getByRole('button', { name: /^Delete “/ });
  // A button reads "Deleting…" until the list comes back without its alert.
  const deleting = page.getByRole('button', { name: /^Deleting/ });
  for (let left = await deletes.count(); left > 0; left--) {
    await deletes.first().click();
    await expect(deletes).toHaveCount(left - 1);
    await expect(deleting).toHaveCount(0);
  }
  if (wanted === 'none') return;

  // The list is newest first, so the last one made heads it.
  for (const alert of ALERTS.toReversed()) {
    await page.getByLabel('Name').fill(alert.name);
    await page.getByRole('radio', { name: alert.source }).check();
    await page.getByLabel('Subsystem').selectOption({ label: alert.subsystem });
    await page.getByLabel('Daily threshold').fill(alert.threshold);
    await page.getByRole('button', { name: 'Create Alert' }).click();
    await expect(page.getByRole('row', { name: new RegExp(`^${alert.name} `) })).toBeVisible();
  }
}

/** Until the page holds still: fonts in, animations done, the pointer out of the way. */
async function settle(page: Page) {
  await page.mouse.move(page.viewportSize()?.width ?? 0, page.viewportSize()?.height ?? 0);
  await page.evaluate(async () => {
    await document.fonts.ready;
    await Promise.allSettled(document.getAnimations().map((animation) => animation.finished));
  });
}

const screens: Record<string, (browser: Browser) => Promise<Page>> = {
  async alerts(browser) {
    const { page } = await open(browser);
    await signIn(page);
    await alerts(page, 'five');
    await page.reload(); // without the toast
    return page;
  },

  async 'alert-refused'(browser) {
    const { page } = await open(browser);
    await signIn(page);
    await alerts(page, 'five');
    await page.reload();
    await page.getByRole('button', { name: 'Create Alert' }).click();
    await expect(page.getByRole('alert')).toBeFocused();
    return page;
  },

  async 'name-taken'(browser) {
    const { page } = await open(browser);
    await signIn(page);
    await alerts(page, 'five');
    await page.reload();
    await page.getByLabel('Name').fill('NORTHEAST WIND');
    await page.getByLabel('Subsystem').selectOption({ label: 'Northeast' });
    await page.getByLabel('Daily threshold').fill('9000');
    await page.getByRole('button', { name: 'Create Alert' }).click();
    await expect(page.getByLabel('Name')).toHaveAccessibleDescription(
      'You already have an alert with this name',
    );
    return page;
  },

  async 'no-alerts'(browser) {
    const { page } = await open(browser);
    await signIn(page);
    await alerts(page, 'none');
    await page.reload();
    return page;
  },

  async 'alerts-phone'(browser) {
    const { page } = await open(browser, PHONE);
    await signIn(page);
    await alerts(page, 'five');
    await page.reload();
    await page.getByRole('heading', { name: 'Your Alerts' }).scrollIntoViewIfNeeded();
    return page;
  },

  async 'sign-in-refused'(browser) {
    const { page } = await open(browser);
    await page.goto('/sign-in');
    await page.getByLabel('Email address').fill(ANA.email);
    await page.getByLabel('Password').fill('not the password at all');
    await page.getByRole('button', { name: 'Sign In' }).click();
    await expect(page.getByRole('alert')).toContainText('The email or the password is wrong');
    return page;
  },
};

const names = process.argv.slice(2);
const unknown = names.filter((name) => !(name in screens));
if (unknown.length > 0) {
  throw new Error(
    `No screen named ${unknown.join(', ')}. There are: ${Object.keys(screens).join(', ')}.`,
  );
}

await mkdir(out, { recursive: true });
const browser = await chromium.launch();
try {
  for (const [name, take] of Object.entries(screens)) {
    if (names.length > 0 && !names.includes(name)) continue;
    const page = await take(browser);
    await settle(page);
    await page.screenshot({ path: `${out}${name}.png` });
    await page.context().close();
    console.log(`docs/screenshots/${name}.png`);
  }
} finally {
  await browser.close();
}
