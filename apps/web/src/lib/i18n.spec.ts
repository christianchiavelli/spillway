import { describe, expect, it } from 'vitest';

import type { ResolvedPathname } from '$app/types';
import { inPortuguese } from '#lib/testing/portuguese.ts';

import { inEdition, intlLocale, localized } from './i18n.ts';

const path = (href: string) => href as ResolvedPathname;

describe('localized', () => {
  it('leaves an English path as it is', () => {
    expect(localized(path('/plants?period=week'))).toBe('/plants?period=week');
    expect(intlLocale()).toBe('en-US');
  });

  describe('in Portuguese', () => {
    inPortuguese();

    it('puts the path under /pt-br, its query kept', () => {
      expect(localized(path('/plants?period=week'))).toBe('/pt-br/plants?period=week');
      expect(localized(path('/sign-in?redirectTo=%2Fpt-br%2Falerts'))).toBe(
        '/pt-br/sign-in?redirectTo=%2Fpt-br%2Falerts',
      );
      expect(intlLocale()).toBe('pt-BR');
    });

    it('writes the home without the slash SvelteKit would redirect away', () => {
      expect(localized(path('/'))).toBe('/pt-br');
      expect(localized(path('/?period=month'))).toBe('/pt-br?period=month');
    });
  });
});

describe('inEdition', () => {
  it('finds the same page in the other edition, whichever is read', () => {
    const plants = { pathname: '/pt-br/plants', search: '?source=solar' };
    expect(inEdition(plants, 'en')).toBe('/plants?source=solar');
    expect(inEdition(plants, 'pt-BR')).toBe('/pt-br/plants?source=solar');
    expect(inEdition({ pathname: '/', search: '' }, 'pt-BR')).toBe('/pt-br');
    expect(inEdition({ pathname: '/pt-br', search: '' }, 'en')).toBe('/');
  });
});
