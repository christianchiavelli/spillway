import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { page } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';

import Toaster from './Toaster.svelte';

const eightSeconds = 8000;

const toasts = [
  { id: 'flash-1', title: 'Alert created', body: '“Wind in the South” is on your list.' },
];

describe('Toaster', () => {
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ['setTimeout', 'clearTimeout'] });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('closes a toast once it has been up for its lifetime', async () => {
    const ondismiss = vi.fn();
    await render(Toaster, { toasts, ondismiss });
    await expect
      .element(page.getByRole('status').getByText('Alert created', { exact: true }))
      .toBeVisible();

    vi.advanceTimersByTime(eightSeconds - 1);
    expect(ondismiss).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1);
    expect(ondismiss).toHaveBeenCalledExactlyOnceWith('flash-1');
  });

  it('keeps a toast while the pointer is on it, and counts again once it leaves', async () => {
    const ondismiss = vi.fn();
    await render(Toaster, { toasts, ondismiss });
    const toast = page.getByRole('status').getByText('Alert created', { exact: true });

    await toast.hover();
    vi.advanceTimersByTime(eightSeconds * 2);
    expect(ondismiss).not.toHaveBeenCalled();

    await toast.unhover();
    vi.advanceTimersByTime(eightSeconds);
    expect(ondismiss).toHaveBeenCalledExactlyOnceWith('flash-1');
  });

  it('keeps a toast while focus is in it', async () => {
    const ondismiss = vi.fn();
    await render(Toaster, { toasts, ondismiss });
    const close = page.getByRole('button', { name: 'Dismiss “Alert created”' });

    close.element().focus();
    vi.advanceTimersByTime(eightSeconds * 2);
    expect(ondismiss).not.toHaveBeenCalled();

    (close.element() as HTMLElement).blur();
    vi.advanceTimersByTime(eightSeconds);
    expect(ondismiss).toHaveBeenCalledExactlyOnceWith('flash-1');
  });

  it('keeps a toast while the tab is hidden', async () => {
    const ondismiss = vi.fn();
    await render(Toaster, { toasts, ondismiss });
    const hidden = vi.spyOn(document, 'hidden', 'get').mockReturnValue(true);

    document.dispatchEvent(new Event('visibilitychange'));
    vi.advanceTimersByTime(eightSeconds * 2);
    expect(ondismiss).not.toHaveBeenCalled();

    hidden.mockReturnValue(false);
    document.dispatchEvent(new Event('visibilitychange'));
    vi.advanceTimersByTime(eightSeconds);
    expect(ondismiss).toHaveBeenCalledExactlyOnceWith('flash-1');
    hidden.mockRestore();
  });

  it('closes a toast from its button', async () => {
    const ondismiss = vi.fn();
    await render(Toaster, { toasts, ondismiss });

    await page.getByRole('button', { name: 'Dismiss “Alert created”' }).click();
    expect(ondismiss).toHaveBeenCalledExactlyOnceWith('flash-1');
  });
});
