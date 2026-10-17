# 9. A dense dashboard look, with accessibility first

- Status: accepted
- Date: 2026-10-17

## Context

Spillway is a tool for people who watch the grid: it should look like the analytics products they use at work, with dense tables, compact controls, dark chrome and white panels. That look often trades accessibility for density.

## Decision

- **Tokens before components.** `app.css` holds the design system in Tailwind 4's `@theme`: colours named for their role (`text-`, `surface-`, `stroke-`, `icon-`), a spacing scale named 100 to 800, radii, elevation, and type styles (`ts-*`) that set family, size, line height and weight in one class. Tailwind's own scales are reset, so no value outside the system can be written.
- **Open Sans for the interface, Lora for page titles**, self-hosted from Fontsource with metric-matched fallbacks: no third-party request and no shift when they load. Lucide icons, at a light stroke.
- **Where the look and accessibility disagree, accessibility wins:**
  - Every field's label always shows, notched into its border. None doubles as a placeholder that leaves as you type.
  - Choices are native controls: a `<select>` in the field's frame, and radio buttons under the segmented control, laid over each option. They post without JavaScript, and the platform gives them keyboard and screen reader support everywhere.
  - One focus ring for every control.
  - A refused form opens with an error summary that takes focus and links to each field, as GOV.UK's design system does.
  - Toasts stay until they are closed: one that leaves on a timer can go before it is read.
  - Error text and control borders are a step darker than the palette's defaults, for 4.5:1 and 3:1.
  - Each table row is headed by its alert's name, so a screen reader reads it with every cell, and each delete button says what it deletes.
  - Motion stops for anyone who asks the system for less of it.
- **The checks are automatic:** axe audits every state of every page in the end-to-end suite, against WCAG 2.2 AA, and a test fails on any console error, a content security policy report included.

## Consequences

- Components take more markup than a library's: a native radio under a styled label, a summary above every form.
- Every new screen starts from the tokens and the components in `src/lib/components/ui`, not from Tailwind's defaults.

## Alternatives considered

- **A component library (Skeleton, shadcn-svelte, Bits UI):** a faster start, a look that is everyone's, and custom selects that need JavaScript.
