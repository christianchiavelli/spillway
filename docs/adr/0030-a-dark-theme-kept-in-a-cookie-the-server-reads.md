# 30. A dark theme, kept in a cookie the server reads

- Status: accepted, giving the look of [ADR 0009](0009-a-dense-dashboard-look-accessibility-first.md) a dark theme
- Date: 2026-10-28

## Context

Spillway was light only: dark chrome around white panels (ADR 0009). Many of the people it is for set their devices to dark, at work and at night. A theme touches every colour, so it has to be cheap to keep right: a component painting a colour of its own would stay light in the dark, and nothing would say so. It also has to keep the app's other rules: every page rendered on the server, a content security policy that lets in no inline script, and every form working without JavaScript (ADR 0010).

## Decision

- **Every colour role holds both its colours, side by side, in `light-dark()`**, as `--color-surface-canvas: light-dark(cool-grey-25, cool-grey-500)` does. A theme is then only the `color-scheme` `<html>` declares: `light dark` follows the device, `light` or `dark` is a choice. The primitives keep one value each. A role that is the same in both themes, as what sits on the navy chrome is, is named in `src/tokens.spec.ts`, which fails on a role with one colour that is not on its list, and on a component that paints with a primitive or with a colour written out.
- **The dark palette is the light one's navy, deepened**, as Larder's dark theme is: the chrome deepest, `#0a1020`, then the canvas, `#11192d`, then the cards, `#18223a`, each lighter as it comes nearer. Wind and solar go a step brighter, `#3f8ff0` and `#e08a1f`, to clear 3:1 on a dark card, and the map's steps turn round, so the most curtailed region is the lightest, as it is the darkest on white. Every text clears 4.5:1 on its ground, links `#5cc2ff`, errors `#f47070` and the grey of what is disabled, `#8890a3`, included, and axe checks each page in both themes.
- **The choice is a cookie the server reads**, `spillway-theme`, holding `light` or `dark`; no cookie is System. A server hook writes `data-theme` on `<html>` before the page leaves, and the layout's `color-scheme` and `theme-color` follow, so a page is drawn in its theme from the first byte: no flash of the other one, and no inline script in the head to prevent it, which the content security policy would have to let in by its hash. Every page is rendered per request and none is cached; a cache put in front would have to vary by the cookie.
- **Light, Dark or System, from a menu in the header**, beside EN | PT, as in Faultline. It is a native popover anchored to its button: it opens, closes on Escape or a click outside, and comes after its button in the tab order, all without a script. Each choice is a submit button of one form, which posts to `/theme`; its action sets the cookie, HTTP-only, for 400 days, and redirects back. Without JavaScript the page comes back in the theme picked. With it, `use:enhance` sends the same post and the redirect is not followed: the menu hands the theme to the root layout, which puts it on `<html>` at once, with every transition held for that frame, and puts back the one before if the post is refused or lost. The layout's is the one copy in the browser, so data loaded again, which brings the cookie another tab may have changed, moves `<html>`, the menu and the bar together.
- **Each choice shows a scrap of the app in its colours**, and System says which theme the device gives now. That line is CSS, a media query, so it is right without a script and when the device changes; a screen reader hears that System follows the device.
- **The browser's bar on a phone takes the header's navy** (`theme-color`), one per scheme for System.
- **A browser without popovers, Safari before 17 or Chrome before 114, shows neither the menu nor its button**, and the page follows the device.
- **The focus ring in the header and the rail is the dark theme's in both**: they are navy in both themes, and the light theme's blue had 1.9:1 on them, short of the 3:1 a focus indicator needs.

## Consequences

- A new colour is a role with two values, or a role with one and a reason in the test's list.
- The build resolves the roles on `<html>`. Tailwind 4.3.3 and Vite 8.3.4 run Lightning CSS for the browsers Tailwind supports, from Safari 16.4 and Chrome 111, older than `light-dark()`, and it turns each one into a pair of variables that the `color-scheme` rules in `app.css` switch, at the root. So only the whole page changes theme: a part of it cannot be drawn in the other theme by setting its `color-scheme`, which is why the menu's scraps name their primitives.
- Every request carries the cookie, and a cache in front would need `Vary: Cookie`.
- Picking a theme is a request that only stores it, and one refused or lost, offline or from another origin, takes the theme back.
- The end-to-end suite runs axe twice for every state it audits.

## Alternatives considered

- **localStorage and an inline script in the head, as Larder does:** no request, but the server renders the default and the script has to run before the first paint; here the content security policy would have to let it in, and a reader without JavaScript could not choose.
- **The cookie written by a script, as Faultline does:** at once and offline, but two writers of one cookie to keep agreeing, and a cookie every script can read.
- **The device's scheme alone, with no choice:** a reader who wants light on a dark device cannot have it.
- **A toggle between two themes, the moon and sun of Larder's rail:** one click, and no way back to following the device once pressed.
- **`dark:` variants in each component:** the dark colour of a role written at every use, not once, and a missed one found only by looking. `dark:` stays, redefined to follow the theme picked rather than the device alone, so a utility under it cannot disagree with the tokens.
- **A block redefining every role under `[data-theme=dark]`, and the same block again in `@media (prefers-color-scheme: dark)` for System:** no transform needed in any browser, for every dark value written twice.
- **A neutral grey for the dark pages, with the navy chrome kept:** shown beside this palette in the screenshots, a navy frame around a page of another colour; the navy was chosen.
