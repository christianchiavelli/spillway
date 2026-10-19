# 19. Material Symbols, and a wordmark drawn from the data

- Status: accepted, replacing the icon choice in ADR 0009
- Date: 2026-10-17

## Context

ADR 0009 chose Lucide at a light stroke. Drawn into the app, it looked like every other dashboard: Lucide is shadcn's default, and its wind and sun are weather, while what the app counts is what wind turbines and solar panels were kept from making. The logo was a few strokes in Lucide's style, a dam with water over it, and read as an icon next to a name.

## Decision

- **Material Symbols, Rounded, at weight 400**, Google's set (Apache 2.0), version 1.2.95 of `@iconify-json/material-symbols`. Of the four sets drawn into the app to compare, it is the only one with a wind turbine and a solar panel, which now mark the two sources.
- **Compiled at build time by unplugin-icons 24.0.0**: each icon imported from `~icons/material-symbols/*` becomes a Svelte component of inline SVG. Nothing loads at run time, no style attribute is written, so the content security policy is untouched, and every icon is `aria-hidden`, since each one sits next to its label.
- **The logo is a wordmark, and a chart.** "Spillway" set in Lora, the page titles' typeface, turned into outlines. Each l is a bar: its top, cut off by a hairline in its source's colour, is the share of that source's potential the ONS curtailed in 2025, 19.1% of wind's and 25.7% of solar's. The mark, the icon and the favicon draw the same two l's on a 32-unit grid, as stems with Lora's slanted flag.
- **One script draws it all**: `docs/brand/build.py` shapes the word with HarfBuzz, cuts it with Skia's path operations, and writes the SVGs to `docs/brand`, the header's wordmark and the favicon. The shares live in that script, so the logo can follow a year's data.

## Consequences

- The icons look heavier than Lucide's at the same size: Material draws filled forms where Lucide draws strokes.
- A Vite plugin and a 9 MB icon set join the web app's dev dependencies. The icon set is read at build time only; the browser gets the few icons the app uses.
- Changing the logo takes Python (through uv), not a design tool, and the outlines are only as good as Lora's, which is under the SIL Open Font License and may be turned into artwork.
- Under 16 px the cut l's are bars with coloured tops; the hairline shows from about 32 px.

## Alternatives considered

- **Phosphor:** the free set closest to the Font Awesome Light this look started from, with a windmill, not a turbine.
- **Hugeicons,** which has a turbine: newer, and less used than the others.
- **Fluent, from Microsoft:** drawn apart for each size and sharp at 16 px, with nothing for either source.
- **Material Symbols as raw SVG files** (`@material-symbols/svg-400`), read with `?raw` into `{@html}`: no plugin, and markup the linter rightly flags in every component that draws one.
- **A logo of a curve with its peak lifted off, or of a dam's ogee crest:** the first read as a cone, the second as an engineering drawing for an app about data.
