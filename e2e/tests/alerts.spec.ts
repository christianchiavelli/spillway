import { expect, expectAccessible, signUp, test } from './fixtures.ts';

test('an alert is created, refused when it repeats a name, and deleted', async ({ page }) => {
  await signUp(page);

  await expect(page.getByRole('status')).toContainText('Welcome to Spillway');
  await expect(page.getByRole('heading', { name: 'No alerts yet' })).toBeVisible();
  await expectAccessible(page);

  await page.getByRole('link', { name: 'New Alert' }).first().click();
  await expect(page).toHaveURL(/\/alerts\/new$/);
  await expect(page.getByRole('heading', { name: 'No subsystem chosen' })).toBeVisible();

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

  // The preview follows the fields: the last 30 days, against the threshold.
  await page.getByLabel('Name').fill('Wind in the Northeast');
  await page.getByLabel('Subsystem').selectOption({ label: 'Northeast' });
  await expect(
    page.getByRole('img', { name: /Wind in the Northeast, the last 30 days$/ }),
  ).toBeVisible();
  await expect(
    page.getByText('Enter a daily threshold to see how often it would have fired.'),
  ).toBeVisible();
  // A day in the Northeast loses far more than 1 MWh of wind, and never a million.
  await page.getByLabel('Daily threshold').fill('1');
  await expect(page.getByText(/It would have fired on \d+ of the last 30 days/)).toBeVisible();
  await page.getByLabel('Daily threshold').fill('1000000');
  await expect(page.getByText('It would not have fired in the last 30 days.')).toBeVisible();
  await expectAccessible(page);

  // A threshold written with a separator is refused, and kept as typed.
  await page.getByLabel('Daily threshold').fill('1,500');
  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(page.getByLabel('Daily threshold')).toHaveValue('1,500');
  await expect(page.getByLabel('Daily threshold')).toHaveAccessibleDescription(
    'Enter a whole number of MWh, such as 500',
  );

  // So low a threshold fires on nearly every day, which fills the firings.
  await page.getByLabel('Daily threshold').fill('1');
  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(page).toHaveURL(/\/alerts$/);
  await expect(page.getByRole('status')).toContainText('Alert created');
  const alerts = page.getByRole('table', { name: 'Your alerts, newest first' });
  const row = alerts.getByRole('row', { name: /Wind in the Northeast/ });
  await expect(row).toContainText('1 MWh');
  await expect(row).toContainText(/[12]\d days|30 days/);
  const firings = page.getByRole('table', { name: /The days an alert went over its threshold/ });
  await expect(firings.getByRole('row')).toHaveCount(11);
  await expect(page.getByText(/^1 to 10 of (2\d|30)$/)).toBeVisible();
  await expectAccessible(page);

  await page.getByRole('link', { name: 'Older' }).click();
  await expect(page).toHaveURL(/\/alerts\?page=2$/);
  await expect(page.getByText(/^11 to 20 of (2\d|30)$/)).toBeVisible();

  // Names are unique per account, whatever their case: the Go API says so.
  await page.goto('/alerts/new');
  await page.getByLabel('Name').fill('WIND in the northeast');
  await page.getByRole('radio', { name: 'Solar' }).check();
  await page.getByLabel('Subsystem').selectOption({ label: 'South' });
  await page.getByLabel('Daily threshold').fill('20');
  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(page.getByLabel('Name')).toHaveAccessibleDescription(
    'You already have an alert with this name',
  );
  await expect(page.getByRole('radio', { name: 'Solar' })).toBeChecked();

  await page.getByRole('link', { name: 'Cancel' }).click();
  await page.getByRole('button', { name: 'Delete “Wind in the Northeast”' }).click();
  await expect(page.getByRole('status')).toContainText('Alert deleted');
  await expect(page.getByRole('heading', { name: 'No alerts yet' })).toBeVisible();
});

test('a reload after creating an alert does not post the form again', async ({ page }) => {
  await signUp(page);
  await page.goto('/alerts/new');
  await page.getByLabel('Name').fill('Solar in the South');
  await page.getByRole('radio', { name: 'Solar' }).check();
  await page.getByLabel('Subsystem').selectOption({ label: 'South' });
  await page.getByLabel('Daily threshold').fill('300');
  await page.getByRole('button', { name: 'Create Alert' }).click();
  await expect(page.getByRole('row', { name: /Solar in the South/ }).first()).toBeVisible();

  await page.reload();
  await expect(
    page.getByRole('table', { name: 'Your alerts, newest first' }).getByRole('row'),
  ).toHaveCount(2);
  await expect(
    page.getByRole('navigation', { name: 'Main' }).getByRole('link', { name: 'Alerts' }),
  ).toHaveAttribute('aria-current', 'page');
});
