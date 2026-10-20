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
  await expect(page.getByRole('alert')).toContainText('Wrong email address or password');
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
  await expect(page.getByRole('alert')).toContainText('Wrong email address or password');
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

test('a field is checked when it is left, never while it is first typed', async ({ page }) => {
  await page.goto('/sign-up');
  const email = page.getByLabel('Email address');
  const secret = page.getByLabel('Password');
  // The reveal button comes with the page's script, which does the checking.
  await expect(page.getByRole('button', { name: 'Show password' })).toBeVisible();
  await expect(secret).toHaveAccessibleDescription(
    'At least 15 characters. A short phrase works well.',
  );

  await email.fill('ana@');
  await expect(email).not.toHaveAttribute('aria-invalid');
  await secret.focus();
  await expect(email).toHaveAttribute('aria-invalid', 'true');
  await expect(email).toHaveAccessibleDescription('Enter an email address like name@example.com');

  // Once found wrong, a field is checked as it is typed, and the error goes
  // as soon as it is put right.
  await email.fill('ana@example.com');
  await expect(email).not.toHaveAttribute('aria-invalid');
  await expectAccessible(page);
});

test('a form the browser can tell is wrong is not sent', async ({ page }) => {
  await page.goto('/sign-in');
  await expect(page.getByRole('button', { name: 'Show password' })).toBeVisible();
  const posts: string[] = [];
  page.on('request', (request) => {
    if (request.method() === 'POST') posts.push(request.url());
  });

  await page.getByRole('button', { name: 'Sign In' }).click();
  await expect(page.getByLabel('Email address')).toBeFocused();
  await expect(page.getByLabel('Email address')).toHaveAccessibleDescription(
    'Enter your email address',
  );
  await expect(page.getByLabel('Password')).toHaveAccessibleDescription('Enter a password');
  expect(posts).toEqual([]);
});

test('a password can be shown, and is cleared all the same when refused', async ({ page }) => {
  await page.goto('/sign-in');
  const reveal = page.getByRole('button', { name: 'Show password' });
  const secret = page.getByLabel('Password');
  await secret.fill('not the password at all');
  await expect(reveal).toHaveAttribute('aria-pressed', 'false');
  await reveal.click();
  await expect(reveal).toHaveAttribute('aria-pressed', 'true');
  await expect(secret).toHaveAttribute('type', 'text');

  await page.getByLabel('Email address').fill(newEmail());
  await page.getByRole('button', { name: 'Sign In' }).click();
  await expect(page.getByRole('alert')).toContainText('Wrong email address or password');
  await expect(secret).toHaveValue('');
});
