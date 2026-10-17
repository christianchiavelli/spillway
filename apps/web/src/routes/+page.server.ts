import { redirect } from '@sveltejs/kit';

import type { PageServerLoad } from './$types';

// The overview of the grid will live here. Until it does, the front door opens
// onto the one page that exists, instead of a page about pages to come.
export const load: PageServerLoad = ({ locals }) => {
  redirect(307, locals.session === null ? '/sign-in' : '/alerts');
};
