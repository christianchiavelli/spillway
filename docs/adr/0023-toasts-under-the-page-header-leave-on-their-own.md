# 23. Toasts under the page header, leaving on their own

- Status: accepted, replacing the rule for toasts in ADR 0009
- Date: 2026-10-19

## Context

ADR 0009 kept a toast on the screen until it was closed, at the bottom right, so that none would go before it was read. Most dashboards show their notices at the top right, where people look for them. Moved there and kept until closed, a toast covers the page header's actions: after signing up, the welcome toast sat on the New Alert button, and the end-to-end suite could not press it. A toast that never leaves also has to be closed by hand after every save.

## Decision

- **At the top right, 1.5rem under the page header's band**, the white band with the page's title and actions. The band is a CSS anchor (`anchor-name: --page-header`) and the toast region's `top` is read from where it ends, so a taller band, as on a phone, puts the toasts lower. They are never higher than 1.5rem under the app header, which is where a page without a band shows them, and where a browser without anchor positioning shows them on every page.
- **The band is a named anchor, not the region's default one** (`position-anchor`): the toasts stay where they appeared while the page scrolls, where a default anchor would carry them off the screen with the band. A toast that arrives on a page already scrolled is placed from where the band is at that moment.
- **A toast leaves after 8 seconds**, counted from the start again each time the pointer or focus leaves it, or the tab is shown again. It stays while the pointer is on it, while focus is in it and while the tab is hidden. A Svelte attachment on each toast holds its timer; the close button works as before.
- **It leaves without an animation.** Svelte's transitions run through the Web Animations API, which the reduced-motion rule in `app.css` does not reach.

## Consequences

- A slow reader can lose a toast before finishing it, unless they point at it or tab into it: the trade-off ADR 0009 refused, taken for the page it no longer covers. A screen reader announces each toast as it arrives in the live region; the timer only decides how long it stays on the screen.
- What a toast says must also show on the page it lands on: a new alert is in the list, a deleted one is gone from it.
- Without JavaScript, a toast stays until the next page, as before.

## Alternatives considered

- **At the bottom right, where it was:** clear of the page header, and away from where notices are looked for.
- **At the top right, kept until closed:** it covered the New Alert button until someone closed it.
- **A fixed offset from the top:** the bands differ in height, from a title alone to a title, a description and a button that wraps on a phone, so one offset leaves a gap on some pages and covers the band on others. Measuring the band in a script would mean writing a style, or the same rule through the CSSOM, for what CSS now does alone.
