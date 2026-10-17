import { fail, redirect } from '@sveltejs/kit';

import { zNewAccount } from '#lib/api/generated/zod.gen.ts';
import { accountFields } from '#lib/features/account/account.ts';
import { fromIssues, fromProblem } from '#lib/forms/errors.ts';
import { api } from '#lib/server/api.ts';
import { setFlash } from '#lib/server/flash.ts';
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
    const values = { email: text(form, 'email').trim() };
    const input = { ...values, password: text(form, 'password') };

    const parsed = zNewAccount.safeParse(input);
    if (!parsed.success) {
      return fail(422, { values, errors: fromIssues(parsed.error.issues, accountFields, input) });
    }

    const result = await api(event).createAccount(parsed.data);
    if (!result.ok) {
      return fail(result.status, { values, errors: fromProblem(result.problem, accountFields) });
    }

    keepSession(event.cookies, event.url, result.data.token, result.data.expiresAt);
    setFlash(event.cookies, event.url, { kind: 'account-created' });
    redirect(303, safeRedirectTarget(event.url.searchParams.get('redirectTo')));
  },
} satisfies Actions;
