import type { ResolvedPathname } from '$app/types';
import { getLocale, type Locale, localizeHref } from '#lib/paraglide/runtime.js';

// SvelteKit sends /pt-br/ on to /pt-br, so an edition's home goes without the
// slash its prefix leaves: /pt-br and /pt-br?period=week, one request each.
const tidy = (href: string) => href.replace(/^(\/[^?#]+)\/(?=[?#]|$)/, '$1');

/**
 * A path inside the app, in the edition being read: /pt-br/plants for /plants
 * in Portuguese. It takes resolve()'s output, so the routes stay typed.
 */
export const localized = (href: ResolvedPathname): ResolvedPathname =>
  tidy(localizeHref(href)) as ResolvedPathname;

/** The page at this address in another edition, its query kept. */
export const inEdition = (
  url: { pathname: string; search: string },
  locale: Locale,
): ResolvedPathname =>
  tidy(localizeHref(url.pathname + url.search, { locale })) as ResolvedPathname;

/**
 * The tag Intl formats with: English as the United States writes it, and
 * Portuguese as Brazil does, 1.234,5 and 17 de outubro.
 */
export const intlLocale = (): 'en-US' | 'pt-BR' => (getLocale() === 'pt-BR' ? 'pt-BR' : 'en-US');

/** Each edition, named in its own language, for the switch between them. */
export const editions = [
  { locale: 'en', short: 'EN', name: 'English' },
  { locale: 'pt-BR', short: 'PT', name: 'Português' },
] as const satisfies readonly { locale: Locale; short: string; name: string }[];
