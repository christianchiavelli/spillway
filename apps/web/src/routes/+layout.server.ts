import { takeFlash } from '#lib/server/flash.ts';

import type { LayoutServerLoad } from './$types';

// The account's address and the theme picked, for the header, and the message
// the last action left. The protocol is read untracked: it never changes, and
// tracking the URL would rerun this load on every navigation.
export const load: LayoutServerLoad = ({ locals, cookies, url, untrack }) => ({
  email: locals.session?.account.email ?? null,
  theme: locals.theme,
  flash: takeFlash(
    cookies,
    untrack(() => new URL(url.href)),
  ),
});
