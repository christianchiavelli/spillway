import { describe, expect, it } from 'vitest';
import { page } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';

import TextField from './TextField.svelte';

describe('TextField', () => {
  it('is named by its label, with the hint as its description', async () => {
    await render(TextField, {
      label: 'Name',
      name: 'name',
      hint: 'Unique among your alerts',
      required: true,
    });

    const input = page.getByRole('textbox', { name: 'Name' });
    await expect.element(input).toBeInTheDocument();
    await expect.element(input).toHaveAccessibleDescription('Unique among your alerts');
    await expect.element(input).toBeRequired();
    await expect.element(input).not.toHaveAttribute('aria-invalid');
  });

  it('replaces the hint with the error, and says the field is invalid', async () => {
    await render(TextField, {
      label: 'Daily threshold',
      name: 'thresholdMwh',
      value: '1,500',
      hint: 'Curtailed in one day',
      error: 'Enter a whole number of MWh, such as 500',
    });

    const input = page.getByRole('textbox', { name: 'Daily threshold' });
    await expect.element(input).toHaveAttribute('aria-invalid', 'true');
    await expect
      .element(input)
      .toHaveAccessibleDescription('Enter a whole number of MWh, such as 500');
    await expect.element(input).toHaveValue('1,500');
    await expect.element(page.getByText('Curtailed in one day')).not.toBeInTheDocument();
  });

  it('keeps the required star out of the accessible name', async () => {
    await render(TextField, {
      label: 'Email address',
      name: 'email',
      type: 'email',
      required: true,
    });

    const input = page.getByRole('textbox', { name: 'Email address', exact: true });
    await expect.element(input).toBeInTheDocument();
    // With neither a hint nor an error, it describes nothing, not an empty line.
    await expect.element(input).not.toHaveAttribute('aria-describedby');
  });
});
