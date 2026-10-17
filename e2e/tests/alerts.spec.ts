import { expect, expectAccessible, signUp, test } from './fixtures.ts';

test('an alert is created, refused when it repeats a name, and deleted', async ({ page }) => {
  await signUp(page);

  await expect(page.getByRole('status')).toContainText('Welcome to Spillway');
  await expect(page.getByRole('heading', { name: 'No alerts yet' })).toBeVisible();
  await expectAccessible(page);

  // Nothing filled in: the summary takes focus and links to each field.
  await page.getByRole('button', { name: 'Create Alert' }).click();
  const summary = page.getByRole('alert');
  await expect(summary).toBeFocused();
  await expect(summary.getByRole('link')).toHaveText([
    'Enter a name for the alert',
    'Choose a subsystem',
    'Enter a daily threshold in MWh',
  ]);
  await summary.getByRole('link', { name: 'Choose a subsystem' }).click();
  await expect(page.getByLabel('Subsystem')).toBeFocused();
  await expectAccessible(page);

  // A threshold written with a separator is refused, and kept as typed.
  await page.getByLabel('Name').fill('Wind in the Northeast');
  await page.getByLabel('Subsystem').selectOption({ label: 'Northeast' });
  await page.getByLabel('Daily threshold').fill('1,500');
  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(page.getByLabel('Daily threshold')).toHaveValue('1,500');
  await expect(page.getByLabel('Daily threshold')).toHaveAccessibleDescription(
    'Enter a whole number of MWh, such as 500',
  );

  await page.getByLabel('Daily threshold').fill('1500');
  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(page.getByRole('status')).toContainText('Alert created');
  const row = page.getByRole('row', { name: /Wind in the Northeast/ });
  await expect(row).toContainText('1,500 MWh');
  await expect(row).toContainText('Wind');
  await expect(page.getByLabel('Name')).toHaveValue('');
  await expectAccessible(page);

  // Names are unique per account, whatever their case: the Go API says so.
  await page.getByLabel('Name').fill('WIND in the northeast');
  await page.getByRole('radio', { name: 'Solar' }).check();
  await page.getByLabel('Subsystem').selectOption({ label: 'South' });
  await page.getByLabel('Daily threshold').fill('20');
  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(page.getByLabel('Name')).toHaveAccessibleDescription(
    'You already have an alert with this name',
  );
  await expect(page.getByRole('radio', { name: 'Solar' })).toBeChecked();

  await row.getByRole('button', { name: 'Delete “Wind in the Northeast”' }).click();
  await expect(page.getByRole('status')).toContainText('Alert deleted');
  await expect(page.getByRole('heading', { name: 'No alerts yet' })).toBeVisible();
});

test('a reload after creating an alert does not post the form again', async ({ page }) => {
  await signUp(page);
  await page.getByLabel('Name').fill('Solar in the South');
  await page.getByRole('radio', { name: 'Solar' }).check();
  await page.getByLabel('Subsystem').selectOption({ label: 'South' });
  await page.getByLabel('Daily threshold').fill('300');
  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(page.getByRole('row', { name: /Solar in the South/ })).toBeVisible();

  await page.reload();
  await expect(page.getByRole('row')).toHaveCount(2); // the header and the one alert
  await expect(
    page.getByRole('navigation', { name: 'Main' }).getByRole('link', { name: 'Alerts' }),
  ).toHaveAttribute('aria-current', 'page');
});
