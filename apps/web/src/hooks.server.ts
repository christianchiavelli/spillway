import { type Handle, type HandleServerError, sequence } from '@sveltejs/kit/hooks';

import { api } from '#lib/server/api.ts';
import { log } from '#lib/server/log.ts';
import { endSession, keepSession, readSessionToken } from '#lib/server/session.ts';

/** Gives the request an id, which the API's logs and the error page share. */
const identify: Handle = async ({ event, resolve }) => {
  event.locals.requestId = crypto.randomUUID();
  event.locals.session = null;
  const response = await resolve(event);
  response.headers.set('X-Request-Id', event.locals.requestId);
  return response;
};

/**
 * The headers the content security policy in vite.config.ts does not cover.
 * HSTS only over HTTPS, where it means something.
 */
const secureHeaders: Handle = async ({ event, resolve }) => {
  const response = await resolve(event);
  const h = response.headers;
  h.set('X-Content-Type-Options', 'nosniff');
  h.set('Referrer-Policy', 'strict-origin-when-cross-origin');
  h.set('Permissions-Policy', 'camera=(), microphone=(), geolocation=(), payment=(), usb=()');
  h.set('Cross-Origin-Opener-Policy', 'same-origin');
  if (event.url.protocol === 'https:') {
    h.set('Strict-Transport-Security', 'max-age=63072000; includeSubDomains');
  }
  return response;
};

/**
 * Turns the session cookie into a signed-in account, asking the API every
 * time: a session ended elsewhere (signed out on another device, expired)
 * ends here on the next request. The API may move the deadline forward, and
 * the cookie follows it.
 */
const authenticate: Handle = async ({ event, resolve }) => {
  const token = readSessionToken(event.cookies, event.url);
  if (token !== undefined) {
    const result = await api(event, token).currentSession();
    if (result.ok) {
      const { account, expiresAt } = result.data;
      event.locals.session = { token, account, expiresAt };
      keepSession(event.cookies, event.url, token, expiresAt);
    } else {
      endSession(event.cookies, event.url);
    }
  }
  return resolve(event);
};

export const handle = sequence(identify, secureHeaders, authenticate);

/**
 * Errors the app expects (a 404, a redirect to sign in) keep SvelteKit's
 * answer. Failures get the request id as a reference on the error page, and
 * the ones nobody caught are logged here, with the cause the page never shows.
 */
export const handleError: HandleServerError = (caught) => {
  const { event } = caught;
  if (caught.kind === 'framework' || caught.kind === 'validation') return;
  if (caught.kind === 'app') {
    return caught.error.status >= 500 ? { reference: event.locals.requestId } : undefined;
  }
  log.error('unhandled error', {
    request_id: event.locals.requestId,
    path: event.url.pathname,
    error: caught.error,
  });
  return {
    status: 500,
    message: 'Something went wrong on our side.',
    reference: event.locals.requestId,
  };
};
