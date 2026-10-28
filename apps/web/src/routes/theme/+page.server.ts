import { error, redirect } from '@sveltejs/kit';

import { resolve } from '$app/paths';
import { localized } from '#lib/i18n.ts';
import { m } from '#lib/paraglide/messages.js';
import { text } from '#lib/server/forms.ts';
import { safeRedirectTarget } from '#lib/server/session.ts';
import { isTheme, THEME_COOKIE, THEME_COOKIE_MAX_AGE } from '#lib/theme.ts';

import type { Actions, PageServerLoad } from './$types';

// Nothing to show here: the address is where the theme menu posts.
export const load: PageServerLoad = () => redirect(307, localized(resolve('/')));

export const actions = {
  /**
   * Keeps the theme picked, then goes back to the page it was picked on. Only
   * the server writes the cookie, HTTP-only like the session's, and System is
   * no cookie at all. With JavaScript the page has taken the theme already,
   * and the menu does not follow the redirect.
   */
  default: async ({ request, cookies, url }) => {
    const form = await request.formData();
    const theme = text(form, 'theme');
    if (!isTheme(theme)) error(400, m.error_refused());

    const secure = url.protocol === 'https:';
    if (theme === 'system') {
      cookies.delete(THEME_COOKIE, { path: '/', secure });
    } else {
      cookies.set(THEME_COOKIE, theme, {
        path: '/',
        httpOnly: true,
        sameSite: 'lax',
        secure,
        maxAge: THEME_COOKIE_MAX_AGE,
      });
    }
    redirect(303, safeRedirectTarget(text(form, 'back') || null, localized(resolve('/'))));
  },
} satisfies Actions;
