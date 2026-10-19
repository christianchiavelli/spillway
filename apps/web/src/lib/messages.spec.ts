import { readFileSync } from 'node:fs';

import { describe, expect, it } from 'vitest';

const read = (locale: string) =>
  JSON.parse(
    readFileSync(new URL(`../../messages/${locale}.json`, import.meta.url), 'utf8'),
  ) as Record<string, unknown>;

// Paraglide falls back to English for a message Portuguese lacks, so a missing
// translation would ship without a word; this is where it fails instead.
describe('the messages', () => {
  it('are the same in both editions, and none is left blank', () => {
    const en = read('en');
    const pt = read('pt-BR');
    expect(Object.keys(pt).sort()).toEqual(Object.keys(en).sort());
    for (const [key, value] of Object.entries(pt)) {
      expect(value, key).not.toEqual('');
    }
  });
});
