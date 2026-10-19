import { expect, signUp, test } from './fixtures.ts';

// No link may lead to a 404: a page that does not exist yet has no link to it.
test('every link on every page leads somewhere', async ({ page, request }) => {
  test.slow(); // every period's links, as a visitor and signed in
  const seen = new Set<string>();
  const check = async () => {
    const hrefs = await page
      .locator('a[href]')
      .evaluateAll((links) =>
        links.map((a) => (a as HTMLAnchorElement).href).filter((href) => !href.includes('#')),
      );
    for (const href of hrefs) {
      if (seen.has(href)) continue;
      seen.add(href);
      const url = new URL(href);
      if (url.origin !== new URL(page.url()).origin) continue;
      const response = await request.get(href, { maxRedirects: 5 });
      expect(response.status(), href).toBeLessThan(400);
    }
  };

  for (const path of [
    '/',
    '/?period=week',
    '/?period=month',
    '/plants',
    '/history',
    '/sign-in',
    '/sign-up',
  ]) {
    await page.goto(path);
    await check();
  }
  await signUp(page);
  await check();
  await page.goto('/alerts/new');
  await check();

  const missing = await request.get('/no-such-page');
  expect(missing.status()).toBe(404);
});
