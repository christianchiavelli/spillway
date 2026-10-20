import { tick } from 'svelte';

import type { SubmitFunction } from '$app/forms';

type Options = {
  /** Called with true when a submit starts and false when it ends. */
  onpending?: (pending: boolean) => void;
  /** Empties the form after a redirect: what it held is now saved. */
  resetOnRedirect?: boolean;
};

/**
 * What every form here does when JavaScript is on, on top of what it does
 * without: one submit at a time, a pending state to show, and, when the
 * server refuses the form, focus on the summary of what to fix, so a keyboard
 * or screen reader user lands where the problem is described.
 */
export function enhanceForm({ onpending, resetOnRedirect = false }: Options = {}): SubmitFunction {
  let busy = false;
  return ({ cancel, formElement }) => {
    if (busy) {
      cancel();
      return;
    }
    busy = true;
    onpending?.(true);

    return async ({ result, update }) => {
      try {
        if (result.type === 'redirect' && resetOnRedirect) formElement.reset();
        await update({ reset: false });
        if (result.type === 'failure') {
          // Without JavaScript the page never sends a password back, so a
          // refused form comes back without one. The same here, shown or not.
          for (const input of formElement.querySelectorAll<HTMLInputElement>(
            'input[type="password"], input[data-password]',
          )) {
            input.value = '';
          }
          await tick();
          const summary =
            formElement.ownerDocument.querySelector<HTMLElement>('[data-error-summary]');
          const field = formElement.querySelector<HTMLElement>('[aria-invalid="true"]');
          (summary ?? field)?.focus();
        }
      } finally {
        busy = false;
        onpending?.(false);
      }
    };
  };
}
