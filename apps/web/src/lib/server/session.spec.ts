import { describe, expect, it } from 'vitest';

import { safeRedirectTarget } from './session.ts';

describe('safeRedirectTarget', () => {
  it('keeps a path on this site, with its query and hash', () => {
    expect(safeRedirectTarget('/alerts?sort=name#list')).toBe('/alerts?sort=name#list');
  });

  it('refuses everything that would leave the site', () => {
    for (const raw of [
      null,
      '',
      'alerts',
      '//evil.example/alerts',
      '/\\evil.example',
      'https://evil.example',
      'javascript:alert(1)',
    ]) {
      expect(safeRedirectTarget(raw), String(raw)).toBe('/alerts');
    }
  });
});
