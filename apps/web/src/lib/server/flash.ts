import type { Cookies } from '@sveltejs/kit';
import * as z from 'zod';

/**
 * A message for the page after a redirect. The action that creates an alert
 * redirects to the list, so a reload never posts the form twice, and the list
 * says what happened. It lives in a cookie for a minute and is read once. Its
 * content is checked like any other input, since a cookie is one.
 */
const zFlash = z.discriminatedUnion('kind', [
  z.object({ id: z.uuid(), kind: z.literal('alert-created'), name: z.string().max(80) }),
  z.object({ id: z.uuid(), kind: z.literal('alert-deleted'), name: z.string().max(80) }),
  z.object({ id: z.uuid(), kind: z.literal('account-created') }),
]);

export type Flash = z.infer<typeof zFlash>;

/** A flash before it is sent: the id, which tells two alike apart, is added here. */
type NewFlash = Flash extends infer F ? (F extends Flash ? Omit<F, 'id'> : never) : never;

const NAME = 'spillway-flash';

export function setFlash(cookies: Cookies, url: URL, flash: NewFlash) {
  cookies.set(NAME, JSON.stringify({ ...flash, id: crypto.randomUUID() }), {
    path: '/',
    httpOnly: true,
    sameSite: 'lax',
    secure: url.protocol === 'https:',
    maxAge: 60,
  });
}

export function takeFlash(cookies: Cookies, url: URL): Flash | null {
  const raw = cookies.get(NAME);
  if (raw === undefined) return null;
  cookies.delete(NAME, { path: '/', secure: url.protocol === 'https:' });
  try {
    const parsed = zFlash.safeParse(JSON.parse(raw));
    return parsed.success ? parsed.data : null;
  } catch {
    return null;
  }
}
