import { expect, expectAccessible, test } from './fixtures.ts';

// The pages anyone can read, from the ONS's files the sync service loaded.

test('the overview reads a day, a week and a month, and steps through them', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByRole('heading', { name: 'Overview', level: 1 })).toBeVisible();
  await expect(page.getByText(/^ONS data to \w+ \d+, \d\d:\d\d$/)).toBeVisible();
  await expect(
    page.getByRole('img', { name: /^Energy curtailed in each half hour of / }),
  ).toBeVisible();
  // Nothing is published after the latest day, so its arrow leads nowhere.
  await expect(page.getByRole('link', { name: 'Next day' })).toHaveAttribute(
    'aria-disabled',
    'true',
  );
  await expectAccessible(page);

  const period = page.getByRole('navigation', { name: 'Period' });
  await period.getByRole('link', { name: 'Week' }).click();
  await expect(page).toHaveURL(/\/\?period=week$/);
  await expect(period.getByRole('link', { name: 'Week' })).toHaveAttribute('aria-current', 'page');
  await expect(page.getByRole('heading', { name: 'Day by Day' })).toBeVisible();

  const title = page.getByRole('heading', { level: 2 }).first();
  const thisWeek = await title.textContent();
  await page.getByRole('link', { name: 'Previous week' }).click();
  await expect(page).toHaveURL(/\/\?period=week&date=\d{4}-\d{2}-\d{2}$/);
  await expect(title).not.toHaveText(thisWeek ?? '');
  await expect(page.getByText('Curtailed in the week')).toBeVisible();
  await page.getByRole('link', { name: 'Next week' }).click();
  await expect(title).toHaveText(thisWeek ?? '');

  await period.getByRole('link', { name: 'Month' }).click();
  // The day stays when the kind changes: this month is the one that holds it.
  await expect(page).toHaveURL(/\/\?period=month&date=\d{4}-\d{2}-\d{2}$/);
  await expectAccessible(page);

  // The chart's figures, for whoever cannot see it.
  await page.getByText('Show as table').click();
  await expect(page.getByRole('table', { name: /^Energy curtailed each day, / })).toBeVisible();
});

test('a day the data does not reach is not found', async ({ request }) => {
  expect((await request.get('/?date=2001-01-01')).status()).toBe(404);
  expect((await request.get('/plants?period=year')).status()).toBe(404);
});

test('the plants are ranked, filtered and paged', async ({ page }) => {
  await page.goto('/plants?period=month');
  await expect(page.getByRole('heading', { name: /^\d[\d,]* Plants? Limited$/ })).toBeVisible();
  await expectAccessible(page);

  const ranking = page.getByRole('table', { name: /^Plants limited, / });
  await page
    .getByRole('navigation', { name: 'Source' })
    .getByRole('link', { name: 'Solar' })
    .click();
  await expect(page).toHaveURL(/\/plants\?period=month&source=solar$/);
  await expect(ranking.getByRole('row').nth(1)).toContainText('Solar');
  await expect(ranking.getByRole('cell', { name: 'Wind', exact: true })).toHaveCount(0);

  await page.getByRole('combobox', { name: 'Subsystem' }).selectOption({ label: 'Northeast' });
  await expect(page).toHaveURL(/\/plants\?period=month&source=solar&subsystem=NE$/);

  await page.getByRole('link', { name: 'Next', exact: true }).click();
  await expect(page).toHaveURL(/&page=2$/);
  await expect(page.getByText(/^21 to \d+ of \d+$/)).toBeVisible();
  await expect(ranking.getByRole('row')).not.toHaveCount(0);
});

test('the history goes back to the first file', async ({ page }) => {
  await page.goto('/history');
  await expect(
    page.getByRole('img', { name: /^Energy curtailed each month since / }),
  ).toBeVisible();
  await expect(page.getByRole('rowheader', { name: '2025', exact: true })).toBeVisible();
  await expectAccessible(page);
});
