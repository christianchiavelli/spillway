import { type Cookies, type RequestEvent, redirect } from '@sveltejs/kit';

/**
 * The session token lives in one cookie the browser keeps but never reads:
 * HTTP-only, so no script can take it, and SameSite=Lax, so forms on other
 * sites do not carry it. Over HTTPS it gets the __Host- prefix, which pins it
 * to this host and to Secure; plain http://localhost cannot carry the prefix.
 */
function cookieName(url: URL) {
  return url.protocol === 'https:' ? '__Host-spillway-session' : 'spillway-session';
}

export function readSessionToken(cookies: Cookies, url: URL): string | undefined {
  return cookies.get(cookieName(url)) || undefined;
}

/** Stores the token until the session's deadline, which moves as it is used. */
export function keepSession(cookies: Cookies, url: URL, token: string, expiresAt: string) {
  cookies.set(cookieName(url), token, {
    path: '/',
    httpOnly: true,
    sameSite: 'lax',
    secure: url.protocol === 'https:',
    expires: new Date(expiresAt),
  });
}

export function endSession(cookies: Cookies, url: URL) {
  cookies.delete(cookieName(url), { path: '/', secure: url.protocol === 'https:' });
}

/**
 * Returns the signed-in session, or sends the visitor to sign in and then
 * back to the page they asked for.
 */
export function requireSession(event: RequestEvent) {
  const { session } = event.locals;
  if (session === null) {
    const back = event.url.pathname + event.url.search;
    redirect(303, `/sign-in?${new URLSearchParams({ redirectTo: back }).toString()}`);
  }
  return session;
}

/**
 * For an API call that answers 401 after the hook found a session: it ended
 * in between, signed out elsewhere or expired. Sign in again, and come back.
 */
export function sessionLost(event: RequestEvent): never {
  endSession(event.cookies, event.url);
  redirect(303, `/sign-in?${new URLSearchParams({ redirectTo: event.url.pathname }).toString()}`);
}

/**
 * Where to go after signing in: the page asked for, when it is a path on this
 * site, and the alerts otherwise. Anything else, like `//evil.example` or an
 * absolute URL, would turn the sign-in form into a way to send people away.
 */
export function safeRedirectTarget(raw: string | null, fallback = '/alerts'): string {
  if (raw === null || !raw.startsWith('/') || raw.startsWith('//') || raw.startsWith('/\\')) {
    return fallback;
  }
  try {
    const url = new URL(raw, 'http://spillway.invalid');
    return url.origin === 'http://spillway.invalid'
      ? url.pathname + url.search + url.hash
      : fallback;
  } catch {
    return fallback;
  }
}
