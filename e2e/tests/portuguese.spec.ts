import { expect, expectAccessible, newEmail, password, test } from './fixtures.ts';

// The Portuguese edition is the same app under /pt-br: the address alone
// chooses it, so every link, redirect and form has to keep the prefix.

test('the same page is one link away in the other edition, its query kept', async ({ page }) => {
  await page.goto('/plants?source=solar');
  await page
    .getByRole('navigation', { name: 'Language' })
    .getByRole('link', { name: 'Português' })
    .click();
  await expect(page).toHaveURL(/^http:\/\/[^/]+\/pt-br\/plants\?source=solar$/);
  await expect(page.locator('html')).toHaveAttribute('lang', 'pt-BR');
  await expect(page.getByRole('heading', { level: 1, name: 'Usinas' })).toBeVisible();
  const languages = page.getByRole('navigation', { name: 'Idioma' });
  await expect(languages.getByRole('link', { name: 'Português' })).toHaveAttribute(
    'aria-current',
    'true',
  );
  await expectAccessible(page);

  await languages.getByRole('link', { name: 'English' }).click();
  await expect(page).toHaveURL(/^http:\/\/[^/]+\/plants\?source=solar$/);
  await expect(page.locator('html')).toHaveAttribute('lang', 'en');
});

test('each page names its other edition for search engines', async ({ page }) => {
  await page.goto('/pt-br/history');
  const alternates = await page
    .locator('link[rel=alternate]')
    .evaluateAll((links) =>
      links.map(
        (l) =>
          `${l.getAttribute('hreflang') ?? ''} ${new URL(l.getAttribute('href') ?? '').pathname}`,
      ),
    );
  expect(alternates).toEqual(['en /history', 'pt-BR /pt-br/history', 'x-default /history']);
});

test('the data pages read in Portuguese, numbers and dates included', async ({ page }) => {
  await page.goto('/pt-br');
  await expect(page.getByRole('heading', { level: 1, name: 'Visão geral' })).toBeVisible();
  await expect(page.getByText(/^Dados do ONS até \d{1,2} de [a-zç]+, \d{2}:\d{2}$/)).toBeVisible();
  await expectAccessible(page);

  await page.goto('/pt-br/history');
  await expect(page.getByRole('heading', { level: 1, name: 'Histórico' })).toBeVisible();
  await expectAccessible(page);
});

test('an account and an alert, made and refused in Portuguese', async ({ page }) => {
  await page.goto('/pt-br/sign-up');
  await page.getByLabel('E-mail').fill(newEmail());
  await page.getByLabel('Senha').fill(password);
  await page.getByRole('button', { name: 'Criar conta' }).click();
  await expect(page).toHaveURL(/\/pt-br\/alerts$/);
  await expect(page.getByRole('status')).toContainText('Boas-vindas ao Spillway');
  await expectAccessible(page);

  await page.getByRole('link', { name: 'Novo alerta' }).first().click();
  await expect(page).toHaveURL(/\/pt-br\/alerts\/new$/);
  await page.getByLabel('Nome').fill('Eólica no Norte');
  await page.getByLabel('Subsistema').selectOption({ label: 'Norte' });
  // A dot groups thousands in Portuguese and is a decimal point in English,
  // so a threshold is digits only, in both editions.
  await page.getByLabel('Limite diário').fill('1.500');
  await page.getByRole('button', { name: 'Criar alerta' }).click();
  await expect(page.getByRole('alert')).toContainText('Informe um número inteiro de MWh, como 500');
  await expectAccessible(page);

  await page.getByLabel('Limite diário').fill('1500');
  await page.getByRole('button', { name: 'Criar alerta' }).click();
  await expect(page).toHaveURL(/\/pt-br\/alerts$/);
  await expect(
    page
      .getByRole('table', { name: 'Seus alertas, os mais novos primeiro' })
      .getByRole('row', { name: /Eólica no Norte/ }),
  ).toContainText('1.500 MWh');

  await page.getByRole('button', { name: 'Sair' }).click();
  await expect(page).toHaveURL(/\/pt-br\/sign-in$/);
  await page.goto('/pt-br/alerts');
  await expect(page).toHaveURL(/\/pt-br\/sign-in\?redirectTo=%2Fpt-br%2Falerts$/);
  await expect(page.getByRole('heading', { level: 1, name: 'Entrar' })).toBeVisible();
});
