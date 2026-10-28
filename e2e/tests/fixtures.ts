import { randomUUID } from 'node:crypto';

import AxeBuilder from '@axe-core/playwright';
import { test as base, expect, type Page } from '@playwright/test';

export const password = 'correct horse battery staple';

/** An address no other test uses, so tests run in parallel against one database. */
export function newEmail() {
  return `e2e-${randomUUID()}@example.com`;
}

type Fixtures = {
  /** Fails the test on any console error, a broken content security policy included. */
  page: Page;
};

export const test = base.extend<Fixtures>({
  page: async ({ page }, use) => {
    const errors: { text: string; url: string }[] = [];
    const refusedForms = new Set<string>();
    page.on('response', (response) => {
      const status = response.status();
      if (response.request().method() === 'POST' && status >= 400 && status < 500) {
        refusedForms.add(response.url());
      }
    });
    page.on('console', (message) => {
      if (message.type() === 'error')
        errors.push({ text: message.text(), url: message.location().url });
    });
    page.on('pageerror', (error) => errors.push({ text: error.message, url: '' }));
    await use(page);
    // A form the server refuses answers 4xx, and Chromium logs that as a
    // failed load. The refusal is the point of those tests, which check what
    // the page says; every other error still fails the test.
    const unexpected = errors.filter(
      ({ text, url }) => !(text.startsWith('Failed to load resource') && refusedForms.has(url)),
    );
    expect(
      unexpected.map(({ text }) => text),
      'console errors',
    ).toEqual([]);
  },
});

export { expect };

/** Creates an account through the sign-up form and lands on the alerts. */
export async function signUp(page: Page, email = newEmail()) {
  await page.goto('/sign-up');
  await page.getByLabel('Email address').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Create Account' }).click();
  await expect(page).toHaveURL(/\/alerts$/);
  return email;
}

/**
 * Runs axe with every WCAG 2.2 AA rule on the page as it settles, in the
 * light theme and in the dark one: a page left to follow the device takes
 * either, as the test asks the browser for it. axe reads colours as they are
 * drawn, and a toast halfway through fading in, or a control halfway into
 * the other theme, has less contrast than it ends with, so finite animations
 * finish first. The device's scheme is put back after.
 */
export async function expectAccessible(page: Page) {
  const dark = await page.evaluate(() => matchMedia('(prefers-color-scheme: dark)').matches);
  for (const colorScheme of ['light', 'dark'] as const) {
    await page.emulateMedia({ colorScheme });
    await page.evaluate(async () => {
      await new Promise((resolve) => requestAnimationFrame(() => requestAnimationFrame(resolve)));
      await Promise.allSettled(
        document
          .getAnimations()
          .filter((animation) => animation.effect?.getTiming().iterations !== Infinity)
          .map((animation) => animation.finished),
      );
    });
    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'])
      .analyze();
    expect(
      results.violations.map(
        (v) => `${v.id}: ${v.nodes.map((n) => n.target.join(' ')).join(', ')}`,
      ),
      `axe, with the device asking for ${colorScheme}`,
    ).toEqual([]);
  }
  await page.emulateMedia({ colorScheme: dark ? 'dark' : 'light' });
}
