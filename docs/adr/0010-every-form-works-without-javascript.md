# 10. Every form works without JavaScript

- Status: accepted
- Date: 2026-10-17

## Context

A form that needs a script to post fails for anyone whose script has not loaded yet, failed to load, or is turned off, and it hides the failure.

## Decision

- **Every form is a `<form method="POST">` with a SvelteKit action.** The same action answers with JavaScript and without it.
- **`use:enhance` adds to the form, never replaces it** (`src/lib/forms/enhance.ts`): one submission at a time, a pending state, focus on the error summary after a refusal, and the password cleared, as the page without JavaScript would come back.
- **A refused form answers 4xx with the page**: the values typed, never a password, and an error per field.
- **An accepted form redirects with 303**, Post/Redirect/Get, so a reload never posts it again. What happened travels in a flash cookie, `HttpOnly`, for a minute, read once and parsed like any other input, and shows as a toast.
- **An end-to-end project runs with JavaScript off**: it signs up, has an alert refused and then accepted, deletes it and signs out.

## Consequences

- Each page answers before its script loads, in any browser.
- A confirmation costs a cookie and a redirect.
- With JavaScript on, a form that fails still shows 4xx in the browser's console, which the end-to-end suite expects.

## Alternatives considered

- **JSON endpoints called from the browser:** a second API to secure, and nothing works without the script.
