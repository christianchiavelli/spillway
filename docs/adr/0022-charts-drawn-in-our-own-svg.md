# 22. Charts drawn in our own SVG

- Status: accepted; its table under each chart replaced by [ADR 0025](0025-a-chart-or-its-table-by-a-switch.md)
- Date: 2026-10-19

## Context

The overview, the history and the new alert form need stacked columns, horizontal bars and lines, rendered on the server like the rest of each page. The content security policy, set in `vite.config.ts`, allows one style attribute, by its hash, and no `'unsafe-inline'`, and the end-to-end suite fails on any report it makes (ADR 0009).

## Decision

- **Three Svelte components draw the charts**, `Columns`, `BarRows` and `Lines`, as SVG with classes from the design system's tokens and never a style attribute. d3-scale 4.0.2 gives the scales and their nice ticks, and d3-shape 3.2.0 the line paths; both run on the server and in the browser.
- **One spec for every mark**: bars at most 24px thick, a 4px rounded end and a square base, 2px between stacked segments, 2px lines and hairline grids. Wind is the blue of the sky and solar the amber of the sun, both 3:1 on white.
- **Laid out on the server, then measured.** The server draws for the usual width and the browser scales that to fit, so the page arrives whole without JavaScript; once it runs, each chart measures its frame and lays itself out again. A chart's height is set by classes, so the measuring never moves the page.
- **Labels that would overlap are left out**, left to right, and a chart with shorter labels uses them first.
- **The tooltip is drawn in the same SVG**, a group moved with `transform`, so it needs no style attribute either. It is for the pointer only: every chart has its figures in a table under it, in a `details` element that opens without a script.

## Consequences

- No chart library's runtime in the browser, and nothing it writes has to be allowed by the policy.
- Every chart the pages show is ours to keep: a new kind, such as an area or a map, is a new component.
- Before the script runs, on a phone, a chart is the laptop's drawing scaled down, small but whole.

## Alternatives considered

- **ECharts 6.1**, server-side SVG included: its SVG carries `style` attributes and a `<style>` element.
- **LayerChart 2.8:** Svelte, and built on the same d3 modules; its `Chart` and tooltip write inline styles, with no option to stop them.
- **Hashing each style the libraries write:** the hashes change with the data.
