/**
 * The reader's theme: Light or Dark, picked in the header, or System, which
 * follows the device and is what no choice at all means. It is kept in a
 * cookie the server reads, so a page is rendered in it from its first byte:
 * no flash of the other theme, and no inline script to prevent one, which the
 * content security policy would have to let in (ADR 0030).
 */
export const themes = ['light', 'dark', 'system'] as const;
export type Theme = (typeof themes)[number];

export const THEME_COOKIE = 'spillway-theme';

/** As long as a browser keeps a cookie: Chromium caps them at 400 days. */
export const THEME_COOKIE_MAX_AGE = 400 * 24 * 60 * 60;

export const isTheme = (value: unknown): value is Theme =>
  typeof value === 'string' && (themes as readonly string[]).includes(value);

/** A cookie's value as a theme: System when there is none, or one no version of the app wrote. */
export const themeOf = (value: string | undefined): Theme => (isTheme(value) ? value : 'system');

/**
 * The header's colour in each theme, `--color-surface-primary-cool-dark` in
 * app.css, for the bar a phone's browser draws above the page: it reads as
 * part of the header. System gives one for each scheme the device may ask for.
 */
const BAR = { light: '#11192d', dark: '#0a1020' } as const;

export function themeColors(theme: Theme): { content: string; media?: string }[] {
  if (theme !== 'system') return [{ content: BAR[theme] }];
  return [
    { content: BAR.light, media: '(prefers-color-scheme: light)' },
    { content: BAR.dark, media: '(prefers-color-scheme: dark)' },
  ];
}

/** What `color-scheme` the page declares: both when it follows the device. */
export const colorScheme = (theme: Theme): string => (theme === 'system' ? 'light dark' : theme);

/**
 * Puts a theme on the page at once. Controls ease their colours on hover, so
 * without this each would fade into the new theme on its own while the rest
 * of the page changes in a frame: every transition is held while the styles
 * are worked out again, which measuring the page forces, and let back after.
 * A page already in the theme is left alone.
 */
export function applyTheme(theme: Theme, root: HTMLElement = document.documentElement): void {
  if (root.dataset['theme'] === theme) return;
  root.classList.add('theme-switching');
  root.dataset['theme'] = theme;
  root.getBoundingClientRect();
  root.classList.remove('theme-switching');
}
