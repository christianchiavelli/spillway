import { redirect } from '@sveltejs/kit';

import { api } from '#lib/server/api.ts';
import { endSession } from '#lib/server/session.ts';

import type { RequestHandler } from './$types';

// A plain form post, from the header on every page. SvelteKit refuses it from
// any other origin, as it does every form post, so no site can sign anyone out.
export const POST: RequestHandler = async (event) => {
  if (event.locals.session !== null) {
    await api(event).endSession();
  }
  endSession(event.cookies, event.url);
  redirect(303, '/sign-in');
};
