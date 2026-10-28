import type { Page } from '@playwright/test';

import { expect, expectAccessible, newEmail, password, test } from './fixtures.ts';

// The theme menu in the header: Light, Dark, or System, which follows the device.

const CANVAS = { light: 'rgb(247, 249, 252)', dark: 'rgb(17, 25, 45)' };

/** The colour the page is drawn on, under the cards. */
const canvas = (page: Page) =>
  page
    .getByRole('main')
    .evaluate((main) => getComputedStyle(main.parentElement ?? main).backgroundColor);

const hex = (rgb: string) =>
  `#${(rgb.match(/\d+/g) ?? []).map((n) => Number(n).toString(16).padStart(2, '0')).join('')}`;

/** The theme-color the browser uses now: System gives one per scheme. */
const bar = (page: Page) =>
  page
    .locator('meta[name="theme-color"]')
    .evaluateAll((metas) =>
      metas
        .filter(
          (meta) =>
            !meta.getAttribute('media') || matchMedia(meta.getAttribute('media') ?? '').matches,
        )
        .map((meta) => meta.getAttribute('content')),
    );

test('a theme picked in the header takes at once, and every page after it is drawn in it', async ({
  page,
}) => {
  await page.goto('/plants?period=week');
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'system');
  expect(await canvas(page)).toBe(CANVAS.light);
  // Gone if the page loads again.
  await page.locator('body').evaluate((body) => (body.dataset['kept'] = 'yes'));

  await page.getByRole('button', { name: 'Theme: System' }).click();
  const menu = page.getByRole('group', { name: 'Theme' });
  await expect(menu.getByRole('button', { name: 'System', exact: true })).toHaveAttribute(
    'aria-pressed',
    'true',
  );
  await expectAccessible(page);
  await menu.getByRole('button', { name: 'Dark', exact: true }).click();

  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
  expect(await canvas(page)).toBe(CANVAS.dark);
  await expect(menu).toBeHidden();
  await expect(page.getByRole('button', { name: 'Theme: Dark' })).toBeFocused();
  await expect(page).toHaveURL(/\/plants\?period=week$/);
  await expect(page.locator('body')).toHaveAttribute('data-kept', 'yes');

  // The server has it, and writes it into the page before any script runs.
  await expect
    .poll(async () => (await page.request.get('/history')).text())
    .toContain('data-theme="dark"');
  await page
    .getByRole('navigation', { name: 'Main' })
    .getByRole('link', { name: 'History' })
    .click();
  await expect(page).toHaveURL(/\/history$/);
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
  await expect(page.getByRole('button', { name: 'Theme: Dark' })).toBeVisible();
  await page.reload();
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
  expect(await canvas(page)).toBe(CANVAS.dark);
  await expectAccessible(page);

  await page.getByRole('button', { name: 'Theme: Dark' }).click();
  await page.getByRole('button', { name: 'System', exact: true }).click();
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'system');
  expect(await canvas(page)).toBe(CANVAS.light);
  await expect
    .poll(async () => (await page.request.get('/history')).text())
    .toContain('data-theme="system"');
});

test('a theme picked in another tab is taken up when the page loads its data again', async ({
  page,
  context,
}) => {
  await page.goto('/sign-up');
  const other = await context.newPage();
  await other.goto('/');
  await other.getByRole('button', { name: 'Theme: System' }).click();
  await other.getByRole('button', { name: 'Dark', exact: true }).click();
  await expect
    .poll(async () => (await other.request.get('/history')).text())
    .toContain('data-theme="dark"');
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'system');

  // Signing up loads the layout's data again, the cookie with it.
  await page.getByLabel('Email address').fill(newEmail());
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Create Account' }).click();
  await expect(page).toHaveURL(/\/alerts$/);
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
  await expect(page.getByRole('button', { name: 'Theme: Dark' })).toBeVisible();
});

test('System follows the device, and says which theme that gives', async ({ page }) => {
  await page.emulateMedia({ colorScheme: 'dark' });
  await page.goto('/');
  expect(await canvas(page)).toBe(CANVAS.dark);
  await page.getByRole('button', { name: 'Theme: System' }).click();
  const system = page.getByRole('button', { name: 'System', exact: true });
  await expect(system.getByText('Dark on this device')).toBeVisible();
  await expect(system.getByText('Light on this device')).toBeHidden();
  await expect(system).toHaveAccessibleDescription('Follows your device');

  await page.emulateMedia({ colorScheme: 'light' });
  await expect(system.getByText('Light on this device')).toBeVisible();
  await expect(system.getByText('Dark on this device')).toBeHidden();
  await expect.poll(() => canvas(page)).toBe(CANVAS.light);
});

test("the browser's bar takes the header's colour in each theme", async ({ page }) => {
  const header = () =>
    page.getByRole('banner').evaluate((banner) => getComputedStyle(banner).backgroundColor);

  await page.goto('/');
  for (const colorScheme of ['light', 'dark'] as const) {
    await page.emulateMedia({ colorScheme });
    await expect.poll(async () => bar(page)).toEqual([hex(await header())]);
  }

  // Light, picked on a device that asks for dark.
  await page.getByRole('button', { name: 'Theme: System' }).click();
  await page.getByRole('button', { name: 'Light', exact: true }).click();
  expect(await canvas(page)).toBe(CANVAS.light);
  await expect.poll(async () => bar(page)).toEqual([hex(await header())]);
  await expect(page.locator('meta[name="color-scheme"]')).toHaveAttribute('content', 'light');
});

test('the menu is opened, used and closed from the keyboard', async ({ page, isMobile }) => {
  test.skip(isMobile, 'A phone has no Tab key.');
  await page.goto('/');
  const button = page.getByRole('button', { name: 'Theme: System' });
  await button.focus();
  await page.keyboard.press('Enter');
  const menu = page.getByRole('group', { name: 'Theme' });
  await expect(menu).toBeVisible();

  // The menu follows its button in the tab order, wherever it sits on the page.
  await page.keyboard.press('Tab');
  await expect(menu.getByRole('button', { name: 'Light', exact: true })).toBeFocused();
  await page.keyboard.press('Escape');
  await expect(menu).toBeHidden();
  await expect(button).toBeFocused();

  await page.keyboard.press('Enter');
  await page.keyboard.press('Tab');
  await page.keyboard.press('Tab');
  await page.keyboard.press('Enter');
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'dark');
  await expect(page.getByRole('button', { name: 'Theme: Dark' })).toBeFocused();
});
