import { mkdir } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

import { type Browser, chromium, expect, type Page } from '@playwright/test';

/*
 * The README's screens, taken from the stack `docker compose up --wait --build`
 * started, once the sync has loaded the ONS's files: `pnpm run screenshots`,
 * or `pnpm run screenshots <name>…` for some. The data pages show DAY, its
 * week or its month, and the history and the alerts the latest published; the
 * alert pages are taken as Ana, an account of the script's own at example.com,
 * made the first time and given the same five alerts. A screen in the dark
 * theme asks for it as a device would, through System.
 */

const base = process.env['E2E_BASE_URL'] ?? 'http://localhost:8090';
const out = fileURLToPath(new URL('../docs/screenshots/', import.meta.url));

const LAPTOP = { width: 1440, height: 900 };
const PHONE = { width: 390, height: 844 };
const ANA = { email: 'ana@example.com', password: 'correct horse battery staple' };

/**
 * A Sunday when half of what the plants could have generated was curtailed: a
 * full chart, and the same one, with the figures the README's descriptions
 * give, each time the screens are taken, its month's once the month is all
 * published. Only a revision of the ONS's file moves them.
 */
const DAY = '2026-10-18';

/** Where and what the grid curtails most, at thresholds that some days pass and most do not. */
const ALERTS = [
  { name: 'Northeast wind', source: 'Wind', subsystem: 'Northeast', threshold: '150000' },
  { name: 'Northeast solar', source: 'Solar', subsystem: 'Northeast', threshold: '50000' },
  {
    name: 'Southeast solar',
    source: 'Solar',
    subsystem: 'Southeast and Centre-West',
    threshold: '40000',
  },
  { name: 'South wind', source: 'Wind', subsystem: 'South', threshold: '12000' },
  { name: 'North wind', source: 'Wind', subsystem: 'North', threshold: '1500' },
] as const;

async function open(browser: Browser, viewport = LAPTOP, colorScheme: 'light' | 'dark' = 'light') {
  const context = await browser.newContext({
    baseURL: base,
    viewport,
    locale: 'en-US',
    colorScheme,
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

/** Fills the new alert form. */
async function fill(
  page: Page,
  alert: { name: string; source: string; subsystem: string; threshold: string },
) {
  await page.getByLabel('Name').fill(alert.name);
  await page.getByRole('radio', { name: alert.source }).check();
  await page.getByLabel('Subsystem').selectOption({ label: alert.subsystem });
  await page.getByLabel('Daily threshold').fill(alert.threshold);
}

/** Leaves Ana with the five alerts, or with none, on the alerts page. */
async function alerts(page: Page, wanted: 'five' | 'none') {
  await page.goto('/alerts');
  const yours = page.getByRole('table', { name: 'Your alerts, newest first' });
  const deletes = yours.getByRole('button', { name: /^Delete “/ });
  // A button reads "Deleting…" until the list comes back without its alert.
  const deleting = yours.getByRole('button', { name: /^Deleting/ });
  for (let left = await deletes.count(); left > 0; left--) {
    await deletes.first().click();
    await expect(deletes).toHaveCount(left === 1 ? 0 : left - 1);
    await expect(deleting).toHaveCount(0);
  }
  if (wanted === 'none') return;

  // The list is newest first, so the last one made heads it.
  for (const alert of ALERTS.toReversed()) {
    await page.goto('/alerts/new');
    await fill(page, alert);
    await page.getByRole('button', { name: 'Create Alert' }).click();
    await expect(yours.getByRole('row', { name: new RegExp(`^${alert.name} `) })).toBeVisible();
  }
  await page.reload(); // without the toast
}

/** The month by subsystem, a row pointed at so the map picks its region out. */
async function where(page: Page) {
  await page.goto(`/?period=month&date=${DAY}`);
  const card = page.getByRole('region', { name: 'By Subsystem' });
  // The card's top just under the sticky header.
  await card.evaluate((el) => {
    window.scrollBy(0, el.getBoundingClientRect().top - 80);
  });
  await settle(page);
  await card.getByRole('link', { name: 'Southeast and Centre-West' }).hover();
}

/** Until the page holds still: fonts in, charts measured, animations done, the pointer away. */
async function settle(page: Page, pointer = true) {
  if (pointer) {
    await page.mouse.move(page.viewportSize()?.width ?? 0, page.viewportSize()?.height ?? 0);
  }
  await page.evaluate(async () => {
    await document.fonts.ready;
    await new Promise((resolve) => requestAnimationFrame(resolve));
    await Promise.allSettled(document.getAnimations().map((animation) => animation.finished));
  });
}

const screens: Record<string, (browser: Browser) => Promise<Page>> = {
  async overview(browser) {
    const { page } = await open(browser);
    await page.goto(`/?date=${DAY}`);
    return page;
  },

  async 'overview-month'(browser) {
    const { page } = await open(browser);
    await page.goto(`/?period=month&date=${DAY}`);
    return page;
  },

  async 'overview-tooltip'(browser) {
    const { page } = await open(browser);
    await page.goto(`/?date=${DAY}`);
    const chart = page.getByRole('img', { name: /^Energy curtailed in each half hour/ });
    await chart.scrollIntoViewIfNeeded();
    await settle(page);
    const box = await chart.boundingBox();
    if (box === null) throw new Error('The chart is not on the page.');
    // Midday, when solar is held back most.
    await page.mouse.move(box.x + box.width * 0.55, box.y + box.height * 0.6);
    return page;
  },

  async 'overview-where'(browser) {
    const { page } = await open(browser);
    await where(page);
    return page;
  },

  async 'overview-pt'(browser) {
    const { page } = await open(browser);
    await page.goto(`/pt-br?date=${DAY}`);
    return page;
  },

  async 'overview-dark'(browser) {
    const { page } = await open(browser, LAPTOP, 'dark');
    await page.goto(`/?date=${DAY}`);
    return page;
  },

  async 'overview-where-dark'(browser) {
    const { page } = await open(browser, LAPTOP, 'dark');
    await where(page);
    return page;
  },

  async 'theme-menu'(browser) {
    const { page } = await open(browser);
    await page.goto(`/?date=${DAY}`);
    await page.getByRole('button', { name: 'Theme: System' }).click();
    await expect(page.getByRole('group', { name: 'Theme' })).toBeVisible();
    return page;
  },

  async 'overview-phone'(browser) {
    const { page } = await open(browser, PHONE);
    await page.goto(`/?date=${DAY}`);
    return page;
  },

  async plants(browser) {
    const { page } = await open(browser);
    await page.goto(`/plants?period=week&date=${DAY}`);
    return page;
  },

  async history(browser) {
    const { page } = await open(browser);
    await page.goto('/history');
    return page;
  },

  async alerts(browser) {
    const { page } = await open(browser);
    await signIn(page);
    await alerts(page, 'five');
    return page;
  },

  async 'alert-preview'(browser) {
    const { page } = await open(browser);
    await signIn(page);
    await page.goto('/alerts/new');
    await fill(page, { ...ALERTS[0], name: 'Northeast wind, the worst days' });
    return page;
  },

  async 'alert-refused'(browser) {
    const { page } = await open(browser);
    await signIn(page);
    await page.goto('/alerts/new');
    await page.getByRole('button', { name: 'Create Alert' }).click();
    await expect(page.getByRole('alert')).toBeFocused();
    return page;
  },

  async 'name-taken'(browser) {
    const { page } = await open(browser);
    await signIn(page);
    await alerts(page, 'five');
    await page.goto('/alerts/new');
    await fill(page, { ...ALERTS[0], name: 'NORTHEAST WIND', threshold: '120000' });
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
    return page;
  },

  async 'sign-in-refused'(browser) {
    const { page } = await open(browser);
    await page.goto('/sign-in');
    await page.getByLabel('Email address').fill(ANA.email);
    await page.getByLabel('Password').fill('not the password at all');
    await page.getByRole('button', { name: 'Sign In' }).click();
    await expect(page.getByRole('alert')).toContainText('Wrong email address or password');
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
    // The tooltip and the map shots keep the pointer where they put it.
    await settle(
      page,
      !['overview-tooltip', 'overview-where', 'overview-where-dark'].includes(name),
    );
    await page.screenshot({ path: `${out}${name}.png` });
    await page.context().close();
    console.log(`docs/screenshots/${name}.png`);
  }
} finally {
  await browser.close();
}
