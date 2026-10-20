import { describe, expect, it } from 'vitest';
import { page } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';

import PasswordField from './PasswordField.svelte';

describe('PasswordField', () => {
  it('shows what was typed at a press, and says whether it does', async () => {
    await render(PasswordField, {
      id: 'password',
      label: 'Password',
      name: 'password',
      showLabel: 'Show password',
      required: true,
    });

    const field = page.getByRole('textbox', { name: 'Password' });
    const reveal = page.getByRole('button', { name: 'Show password' });
    await expect.element(field).toHaveAttribute('type', 'password');
    await expect.element(reveal).toHaveAttribute('aria-pressed', 'false');

    await reveal.click();
    await expect.element(field).toHaveAttribute('type', 'text');
    await expect.element(reveal).toHaveAttribute('aria-pressed', 'true');
    // Shown, it is still never sent to a spelling service.
    await expect.element(field).toHaveAttribute('spellcheck', 'false');

    await reveal.click();
    await expect.element(field).toHaveAttribute('type', 'password');
  });
});
