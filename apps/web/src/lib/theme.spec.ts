import { describe, expect, it } from 'vitest';

import { applyTheme, colorScheme, themeColors, themeOf } from './theme.ts';

describe('themeOf', () => {
  it.each(['light', 'dark', 'system'] as const)('reads %s', (theme) => {
    expect(themeOf(theme)).toBe(theme);
  });

  it.each([undefined, '', 'Dark', 'sepia', 'dark '])('follows the device for %j', (value) => {
    expect(themeOf(value)).toBe('system');
  });
});

describe('the page a theme renders', () => {
  it('declares one scheme for a theme picked, and both for System', () => {
    expect(colorScheme('light')).toBe('light');
    expect(colorScheme('dark')).toBe('dark');
    expect(colorScheme('system')).toBe('light dark');
  });

  it("colours the browser's bar as the header, for each scheme System may get", () => {
    expect(themeColors('light')).toEqual([{ content: '#11192d' }]);
    expect(themeColors('dark')).toEqual([{ content: '#0a1020' }]);
    expect(themeColors('system')).toEqual([
      { content: '#11192d', media: '(prefers-color-scheme: light)' },
      { content: '#0a1020', media: '(prefers-color-scheme: dark)' },
    ]);
  });
});

describe('applyTheme', () => {
  it('holds every transition back while the page changes theme, and lets them back after', () => {
    const steps: string[] = [];
    const dataset: Record<string, string> = {};
    const root = {
      classList: {
        add: (name: string) => steps.push(`add ${name}`),
        remove: (name: string) => steps.push(`remove ${name}`),
      },
      dataset: new Proxy(dataset, {
        set(target, key, value: string) {
          steps.push(`theme ${value}`);
          target[String(key)] = value;
          return true;
        },
      }),
      getBoundingClientRect() {
        steps.push('lay out');
      },
    };

    applyTheme('dark', root as unknown as HTMLElement);

    expect(steps).toEqual([
      'add theme-switching',
      'theme dark',
      'lay out',
      'remove theme-switching',
    ]);
    expect(dataset['theme']).toBe('dark');
  });

  it('leaves a page already in the theme alone', () => {
    const steps: string[] = [];
    const root = {
      classList: {
        add: (name: string) => steps.push(`add ${name}`),
        remove: (name: string) => steps.push(`remove ${name}`),
      },
      dataset: { theme: 'dark' },
      getBoundingClientRect() {
        steps.push('lay out');
      },
    };

    applyTheme('dark', root as unknown as HTMLElement);

    expect(steps).toEqual([]);
  });
});
