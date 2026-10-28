import { afterEach, describe, expect, it, vi } from 'vitest';
import { page } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';

import { inPortuguese } from '#lib/testing/portuguese.ts';

import ThemeMenu from './ThemeMenu.svelte';

/** The answer SvelteKit gives an enhanced form whose action redirects. */
const redirected = () =>
  new Response(JSON.stringify({ type: 'redirect', status: 303, location: '/' }), {
    headers: { 'content-type': 'application/json' },
  });

describe('ThemeMenu', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('names its button after the theme in use', async () => {
    await render(ThemeMenu, { theme: 'dark', onchange: () => undefined });

    await expect.element(page.getByRole('button', { name: 'Theme: Dark' })).toBeVisible();
  });

  it('marks the theme in use, and says System follows the device', async () => {
    await render(ThemeMenu, { theme: 'system', onchange: () => undefined });
    await page.getByRole('button', { name: 'Theme: System' }).click();

    const menu = page.getByRole('group', { name: 'Theme' });
    await expect.element(menu).toBeVisible();
    await expect
      .element(menu.getByRole('button', { name: 'Light', exact: true }))
      .toHaveAttribute('aria-pressed', 'false');
    await expect
      .element(menu.getByRole('button', { name: 'Dark', exact: true }))
      .toHaveAttribute('aria-pressed', 'false');
    const system = menu.getByRole('button', { name: 'System', exact: true });
    await expect.element(system).toHaveAttribute('aria-pressed', 'true');
    await expect.element(system).toHaveAccessibleDescription('Follows your device');
    await expect.element(menu.getByText('Light on this device')).toBeVisible();
  });

  it('hands the theme picked on at once, closes, and only posts it to be kept', async () => {
    const fetch = vi.spyOn(window, 'fetch').mockResolvedValue(redirected());
    const onchange = vi.fn();
    await render(ThemeMenu, { theme: 'system', onchange });
    await page.getByRole('button', { name: 'Theme: System' }).click();
    await page.getByRole('button', { name: 'Dark', exact: true }).click();

    expect(onchange).toHaveBeenCalledExactlyOnceWith('dark');
    await expect.element(page.getByRole('group', { name: 'Theme' })).not.toBeInTheDocument();

    await vi.waitFor(() => {
      expect(fetch).toHaveBeenCalledOnce();
    });
    const [input, init] = fetch.mock.calls[0] ?? [];
    const url = input instanceof Request ? input.url : input instanceof URL ? input.href : input;
    expect(url).toMatch(/\/theme$/);
    expect(init?.method).toBe('POST');
    const body = init?.body as FormData;
    expect(body.get('theme')).toBe('dark');
    expect(body.get('back')).toMatch(/^\//);
    // The redirect back is not followed: it would load every page's data again.
    expect(fetch).toHaveBeenCalledOnce();
  });

  it('hands the theme it had back when the server does not keep the one picked', async () => {
    // What the action answers an enhanced post it refuses.
    vi.spyOn(window, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ type: 'error', error: { message: 'Refused' } }), {
        status: 400,
        headers: { 'content-type': 'application/json' },
      }),
    );
    const onchange = vi.fn();
    await render(ThemeMenu, { theme: 'light', onchange });
    await page.getByRole('button', { name: 'Theme: Light' }).click();
    await page.getByRole('button', { name: 'Dark', exact: true }).click();

    await vi.waitFor(() => {
      expect(onchange.mock.calls).toEqual([['dark'], ['light']]);
    });
  });

  describe('in Portuguese', () => {
    inPortuguese();

    it('reads in Portuguese', async () => {
      await render(ThemeMenu, { theme: 'light', onchange: () => undefined });
      await page.getByRole('button', { name: 'Tema: Claro' }).click();

      const menu = page.getByRole('group', { name: 'Tema' });
      await expect.element(menu.getByRole('button', { name: 'Escuro', exact: true })).toBeVisible();
      await expect.element(menu.getByText('Claro neste aparelho')).toBeVisible();
    });
  });
});
