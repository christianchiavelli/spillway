import { afterEach, beforeEach } from 'vitest';

import { getLocale, overwriteGetLocale } from '#lib/paraglide/runtime.js';

const asAddressed = getLocale;

/**
 * Reads every message and number in the describe block that calls it in
 * Portuguese, as a request under /pt-br does, and goes back to the address's
 * edition after each test.
 */
export function inPortuguese() {
  beforeEach(() => {
    overwriteGetLocale(() => 'pt-BR');
  });
  afterEach(() => {
    overwriteGetLocale(asAddressed);
  });
}
