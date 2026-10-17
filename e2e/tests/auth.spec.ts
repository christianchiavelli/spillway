import { expect, expectAccessible, newEmail, password, signUp, test } from './fixtures.ts';

test('a visitor is sent to sign in, and back to the page asked for', async ({ page, context }) => {
  const email = await signUp(page);
  await page.getByRole('button', { name: 'Sign Out' }).click();
  await expect(page).toHaveURL(/\/sign-in$/);

  await page.goto('/alerts');
  await expect(page).toHaveURL(/\/sign-in\?redirectTo=%2Falerts$/);
  await expectAccessible(page);

  await page.getByLabel('Email address').fill(email);
  await page.getByLabel('Password').fill('not the password at all');
  await page.getByRole('button', { name: 'Sign In' }).click();
  await expect(page.getByRole('alert')).toContainText('The email or the password is wrong');
  await expect(page.getByLabel('Email address')).toHaveValue(email);
  await expect(page.getByLabel('Password')).toHaveValue('');

  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Sign In' }).click();
  await expect(page).toHaveURL(/\/alerts$/);

  const cookies = await context.cookies();
  const session = cookies.find((c) => c.name.endsWith('spillway-session'));
  expect(session?.httpOnly).toBe(true);
  expect(session?.sameSite).toBe('Lax');
});

test('an unknown address gets the same answer as a wrong password', async ({ page }) => {
  await page.goto('/sign-in');
  await page.getByLabel('Email address').fill(newEmail());
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Sign In' }).click();
  await expect(page.getByRole('alert')).toContainText('The email or the password is wrong');
});

test('signing in never sends anyone to another site', async ({ page }) => {
  const email = await signUp(page);
  await page.getByRole('button', { name: 'Sign Out' }).click();

  await page.goto('/sign-in?redirectTo=//evil.example/alerts');
  await page.getByLabel('Email address').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Sign In' }).click();
  await expect(page).toHaveURL(/^http:\/\/[^/]+\/alerts$/);
});

test('a second account with the same address is refused', async ({ page }) => {
  const email = await signUp(page);
  await page.getByRole('button', { name: 'Sign Out' }).click();

  await page.goto('/sign-up');
  await page.getByLabel('Email address').fill(email.toUpperCase());
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: 'Create Account' }).click();
  await expect(page.getByLabel('Email address')).toHaveAccessibleDescription(
    'An account already uses this email address',
  );
  await expectAccessible(page);
});

test('a short password is refused with the rule it broke', async ({ page }) => {
  await page.goto('/sign-up');
  await page.getByLabel('Email address').fill(newEmail());
  await page.getByLabel('Password').fill('too short');
  await page.getByRole('button', { name: 'Create Account' }).click();
  await expect(page.getByLabel('Password')).toHaveAccessibleDescription(
    'Use at least 15 characters',
  );
});
