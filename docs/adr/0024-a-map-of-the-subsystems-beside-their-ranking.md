# 24. A map of the subsystems, beside their ranking

- Status: accepted
- Date: 2026-10-19

## Context

The overview showed where energy was curtailed as four bars, one a subsystem, and why as three or four more, with a glossary of the reasons under them that repeated each name. Where is a question about places, and the ONS's four subsystems follow its grid, not the states anyone knows: a map has to say which is which. Drawing one needs outlines, a rule for which state goes in which subsystem, a way to shade them, and a way to read it for whoever cannot see it, all under the content security policy that keeps the charts in our own SVG (ADR 0022).

## Decision

- **The outlines are Natural Earth's states**, version 5.1.2 at 1:50m, in the public domain. `docs/map/build.py` projects them in the Albers equal-area conic IBGE uses for Brazil, so no region looks larger than it is, joins them into the subsystems, drops islands under 300 km², and simplifies the four together with GEOS's coverage simplifier, so neighbours keep one border. It writes `subsystem-shapes.ts`: an outline and a label point a subsystem, 524 points in 6.5 KB, with nothing to run in the browser. shapely 2.2.0 and pyproj 3.8.0 run it under uv, as the logo's script runs.
- **A state goes to the subsystem that holds most of its installed capacity** in the ONS's `capacidade-geracao` dataset, as read on 19 October 2026. So Tocantins, whose dams feed the Southeast, is drawn there, and Paraná in the South, though Itaipu is counted in the Southeast. Acre and the Federal District have no plant of their own: Acre is on Rondônia's lines and the District sits inside Goiás, both in the Southeast.
- **Shaded in four steps of the chrome's navy**, cool grey 100, 200 and 300 and blue 500, each a quarter of the period's largest total, wind and solar together. Navy, not wind's brighter blue, so no shade reads as a source: wind and solar keep their colours, and only in the bars. Labels are dark in a white halo, which reads on every step and where a label runs off the coast.
- **The map sits beside a ranking of the same subsystems**, a bar each on one scale, wind and solar stacked. Pointing at a region marks its row and pointing at a row, or focusing its link, fades the rest of the map. Each region and each row leads to the plants of that subsystem in the same period.
- **No tooltip on the map:** the figures it would show are in the row beside it, and a phone has no pointer to hover with.
- **For a screen reader and the keyboard, the ranking is the map.** The map is `aria-hidden` and its links are out of the tab order; the list says the same in words and links to the same places, and the card's table view gives each subsystem's wind and solar.
- **`BarRows` becomes that ranking for both cards**: an ordered list in HTML, a bar in SVG under each row, values in a column on the right. Each reason's meaning sits under its name, in place of the glossary that named each one twice. The subsystems take the wider card, three fifths, since the map needs the room; inside it the map goes beside the list once the card is 36rem wide, a container query, so it follows the card and not the window.

## Consequences

- The Northeast will be the darkest in most periods: it holds most of the plants. The map shows where energy was lost, not where curtailment bites hardest; that would be each subsystem's share of its potential, which the API does not give yet.
- The steps are a quarter of the period's largest, so shades compare within a period, not from one to the next; the figures beside the map do that.
- A state the ONS splits between two subsystems, such as Pará or Paraná, is drawn whole in one of them.
- Changing the map is a script, not a design tool, and the rule for each state lives in it.

## Alternatives considered

- **A circle at each region, sized by its total, or a pie of wind and solar:** truer to a total than shading an area, but four more marks with labels on a map that sits beside the same figures, and pies compare worse than the bars next to them.
- **A turquoise scale,** the colour of a spillway's water: one the app used nowhere else.
- **IBGE's state boundaries:** finer than a map some 270 px wide needs, and under IBGE's terms rather than the public domain.
- **d3-geo in the browser:** a projection worked out on every render for a shape that never changes.
- **mapshaper,** which does the same in one command: a Node tool for a step the logo's script already takes in Python.
