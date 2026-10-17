import { expect, newEmail, password, test } from './fixtures.ts';

// Progressive enhancement, checked: with JavaScript off, every form posts, the
// server answers with the page, and the same rules and words come back.
test('without JavaScript, an account and an alert are made the old way', async ({ page }) => {
  await page.goto('/sign-up');
  await page.getByLabel('Email address').fill(newEmail());
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Create Account' }).click();
  await expect(page).toHaveURL(/\/alerts$/);
  await expect(page.getByRole('status')).toContainText('Welcome to Spillway');
  // The server alone marks where you are.
  await expect(
    page.getByRole('navigation', { name: 'Main' }).getByRole('link', { name: 'Alerts' }),
  ).toHaveAttribute('aria-current', 'page');

  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(page.getByRole('alert')).toContainText('Enter a name for the alert');
  await expect(page.getByLabel('Name')).toHaveAttribute('aria-invalid', 'true');

  await page.getByLabel('Name').fill('Wind in the North');
  await page.getByLabel('Subsystem').selectOption({ label: 'North' });
  await page.getByLabel('Daily threshold').fill('750');
  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(page).toHaveURL(/\/alerts$/);
  await expect(page.getByRole('row', { name: /Wind in the North/ })).toContainText('750 MWh');

  await page.getByRole('button', { name: 'Delete “Wind in the North”' }).click();
  await expect(page.getByRole('heading', { name: 'No alerts yet' })).toBeVisible();

  await page.getByRole('button', { name: 'Sign Out' }).click();
  await expect(page).toHaveURL(/\/sign-in$/);
});
