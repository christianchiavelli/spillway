# 6. Sessions in an HttpOnly cookie, CSRF checked by origin, no CORS

- Status: accepted
- Date: 2026-10-17

## Context

A browser keeps a session for a person, the web server turns it into an account on every request, and the API decides whether it is still good. The token must stay out of reach of scripts and of other sites.

## Decision

- **An opaque token, hashed at rest.** Signing in makes 128 random bits (`crypto/rand.Text`) and returns them once. The API keeps only their SHA-256, so a copy of the database holds no session that works. A fast hash is enough: the token is random, not chosen by a person.
- **Seven days, sliding.** A request with less than half the time left moves the deadline, and the cookie follows. Expired sessions are purged every 15 minutes.
- **One cookie the browser keeps and never reads**: `HttpOnly`, `SameSite=Lax`, `Path=/`, expiring with the session. Over HTTPS it is also `Secure`, with the `__Host-` prefix, which pins it to the host. The web's hook asks the API about it on every request, so a session ended elsewhere ends here on the next one.
- **Passwords.** argon2id at OWASP's baseline (19 MiB, two passes, one lane), stored in PHC format so the costs can rise without a migration. At most one hash per CPU runs at once, and waiting for a turn respects the request's deadline. 15 to 128 characters, with no rules about which, as NIST SP 800-63B asks. An unknown address takes as long to refuse as a wrong password, by checking a decoy hash.
- **CSRF.** SvelteKit refuses a form posted from any origin but the app's own, and no other origin is trusted (ADR 0015 says how the app knows its origin). `SameSite=Lax` would keep the cookie off a cross-site post anyway.
- **No CORS.** The API answers no browser: a request with an `Origin` header not on `CORS_ALLOWED_ORIGINS`, empty by default, is refused with 403 `forbidden_origin` before routing.
- **Signing out is a POST**, which deletes the session in the API and the cookie in the browser. A link cannot sign anyone out.
- **After signing in, only paths on this site.** `?redirectTo=//evil.example` sends you to your alerts.

## Consequences

- A session can be ended at once, from anywhere, by deleting its row.
- Every page view costs one indexed lookup in the API. That is the price of ending sessions at once.

## Alternatives considered

- **JWTs:** no lookup, and no way to end a session before it expires. The web would still keep it in a cookie.
- **Sessions held by the web server:** the API would have to take the web server's word for who is asking.
- **A CSRF token in every form:** what SvelteKit's origin check makes unnecessary for a same-origin app, and one more value to render and compare.
