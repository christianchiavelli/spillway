import { fail, redirect } from '@sveltejs/kit';

import { zCredentials } from '#lib/api/generated/zod.gen.ts';
import { accountFields } from '#lib/features/account/account.ts';
import { fromIssues, fromProblem } from '#lib/forms/errors.ts';
import { api } from '#lib/server/api.ts';
import { text } from '#lib/server/forms.ts';
import { keepSession, safeRedirectTarget } from '#lib/server/session.ts';

import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ locals, url }) => {
  if (locals.session !== null)
    redirect(303, safeRedirectTarget(url.searchParams.get('redirectTo')));
};

export const actions = {
  default: async (event) => {
    const form = await event.request.formData();
    // The password is checked and forgotten: it is never sent back to the page.
    const values = { email: text(form, 'email').trim() };
    const input = { ...values, password: text(form, 'password') };

    const parsed = zCredentials.safeParse(input);
    if (!parsed.success) {
      return fail(422, { values, errors: fromIssues(parsed.error.issues, accountFields, input) });
    }

    const result = await api(event).createSession(parsed.data);
    if (!result.ok) {
      const { code } = result.problem;
      if (code === 'invalid_credentials')
        return fail(400, { values, refused: 'credentials' as const });
      if (code === 'rate_limited') {
        return fail(429, { values, refused: 'throttled' as const, wait: result.retryAfterSeconds });
      }
      return fail(result.status, { values, errors: fromProblem(result.problem, accountFields) });
    }

    keepSession(event.cookies, event.url, result.data.token, result.data.expiresAt);
    redirect(303, safeRedirectTarget(event.url.searchParams.get('redirectTo')));
  },
} satisfies Actions;
