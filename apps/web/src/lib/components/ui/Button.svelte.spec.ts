import { createRawSnippet } from 'svelte';
import { describe, expect, it } from 'vitest';
import { page } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';

import Button from './Button.svelte';

const label = (text: string) => createRawSnippet(() => ({ render: () => `<span>${text}</span>` }));

describe('Button', () => {
  it('refuses another press while pending, without leaving the tab order', async () => {
    await render(Button, { type: 'submit', pending: true, children: label('Signing In…') });

    const button = page.getByRole('button', { name: 'Signing In…' });
    await expect.element(button).toHaveAttribute('aria-disabled', 'true');
    await expect.element(button).not.toHaveAttribute('disabled');
  });

  it('is an ordinary button otherwise', async () => {
    await render(Button, { children: label('Sign In') });

    await expect
      .element(page.getByRole('button', { name: 'Sign In' }))
      .not.toHaveAttribute('aria-disabled');
  });
});
