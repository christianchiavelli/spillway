import { expect, newEmail, password, test } from './fixtures.ts';

// Progressive enhancement, checked: with JavaScript off, every form posts, the
// server answers with the page, and the same rules and words come back.
test('without JavaScript, an account and an alert are made the old way', async ({ page }) => {
  await page.goto('/sign-up');
  await page.getByLabel('Email address').fill(newEmail());
  await page.getByLabel('Password').fill('too short');
  await page.getByRole('button', { name: 'Create Account' }).click();
  // The refused page opens on the field to put right, as the script would.
  await expect(page.getByLabel('Password')).toBeFocused();
  await expect(page.getByLabel('Password')).toHaveAccessibleDescription(
    'Use at least 15 characters',
  );
  await expect(page.getByRole('button', { name: 'Show password' })).toHaveCount(0);

  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Create Account' }).click();
  await expect(page).toHaveURL(/\/alerts$/);
  await expect(page.getByRole('status')).toContainText('Welcome to Spillway');
  // The server alone marks where you are.
  await expect(
    page.getByRole('navigation', { name: 'Main' }).getByRole('link', { name: 'Alerts' }),
  ).toHaveAttribute('aria-current', 'page');

  await page.getByRole('link', { name: 'New Alert' }).first().click();
  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(page.getByRole('alert')).toContainText('Enter a name for the alert');
  await expect(page.getByLabel('Name')).toHaveAttribute('aria-invalid', 'true');

  // The preview shows what the form held when the server sent it back.
  await page.getByLabel('Name').fill('Wind in the North');
  await page.getByLabel('Subsystem').selectOption({ label: 'North' });
  await page.getByLabel('Daily threshold').fill('1,500');
  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(
    page.getByRole('img', { name: /Wind in the North, the last 30 days/ }),
  ).toBeVisible();

  await page.getByLabel('Daily threshold').fill('750');
  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(page).toHaveURL(/\/alerts$/);
  await expect(
    page.getByRole('table', { name: 'Your alerts, newest first' }).getByRole('row', {
      name: /Wind in the North/,
    }),
  ).toContainText('750 MWh');

  await page.getByRole('button', { name: 'Delete “Wind in the North”' }).click();
  await expect(page.getByRole('heading', { name: 'No alerts yet' })).toBeVisible();

  await page.getByRole('button', { name: 'Sign Out' }).click();
  await expect(page).toHaveURL(/\/sign-in$/);
});

test('without JavaScript, the data pages page and filter through their addresses', async ({
  page,
}) => {
  await page.goto('/');
  await page
    .getByRole('navigation', { name: 'Period' })
    .getByRole('link', { name: 'Week' })
    .click();
  await expect(page).toHaveURL(/\?period=week$/);
  await expect(page.getByRole('heading', { name: 'Day by Day' })).toBeVisible();
  await page.getByRole('link', { name: 'Previous week' }).click();
  await expect(page).toHaveURL(/\?period=week&date=\d{4}-\d{2}-\d{2}$/);
  await expect(page.getByRole('link', { name: 'Next week' })).toBeVisible();

  // The subsystem is a form that asks for a new address, with a button for it.
  await page.goto('/plants');
  await page.getByRole('combobox', { name: 'Subsystem' }).selectOption({ label: 'South' });
  await page.getByRole('button', { name: 'Show' }).click();
  await expect(page).toHaveURL(/\/plants\?subsystem=S$/);
  await expect(page.getByRole('combobox', { name: 'Subsystem' })).toHaveValue('S');

  // A chart swaps for its table without a script: the switch is radios and CSS.
  await page.goto('/history');
  const months = page.getByRole('region', { name: 'Curtailed Each Month' });
  await months.getByRole('radio', { name: 'Table' }).check();
  await expect(
    page.getByRole('table', { name: 'Energy curtailed each month, in GWh' }),
  ).toBeVisible();
  await expect(months.getByRole('img')).toBeHidden();
});
