# 25. A chart or its table, by a switch

- Status: accepted, replacing the table under each chart in ADR 0022
- Date: 2026-10-19

## Context

ADR 0022 put each chart's figures in a `details` element under the chart, opened by a "Show as table" link. It read as a footnote: a small line under every chart, for figures that matter as much as the drawing to a screen reader and to anyone who wants the numbers. Opened, the table landed under the chart and pushed the page down, and some are long: a day's 48 half hours made the card four times its height.

## Decision

- **A switch in the card's header, Chart or Table**, the segmented control in a new small size, with its legend for a screen reader only. The table takes the chart's place; the card opens on the chart.
- **Radio buttons and CSS, no script.** `ChartCard` wraps the card in an element with `display: contents` that holds the switch and both views, and the view not chosen is hidden from it with `:has()`. The arrow keys move between the two, a screen reader hears a group named Show as, and the view not shown is not displayed, so it leaves the accessibility tree too.
- **The table is at most 20rem tall**, about a chart's height, and scrolls under its header, so the card keeps its size either way. `Table`'s header cells are sticky for it, and its borders separate, since a collapsed border scrolls away from a sticky cell.
- **The legend goes over the chart**, in the chart's view, and leaves the header to the switch. A card's title takes the room its aside leaves, so the description wraps before the aside drops under it.
- **Every chart has a table**: the share of potential on the history page gets one too.

## Consequences

- One more control in each chart's header; on a phone it sits under the title.
- The choice is a page's: a reload opens every card on its chart again.
- A screen reader finds the figures by choosing Table, where it opened the details before.
- Every table's frame is a named region in the tab order, since one that scrolls must take focus for the keyboard to scroll it: one more stop before each table.
- `:has()` asks for a browser from 2023 on; an older one shows the chart only, and the switch does nothing.

## Alternatives considered

- **Tabs, with a script:** the table out of reach without one, and ARIA's tab pattern to keep for two choices.
- **The view in the address,** as `?view=table`: a parameter for each chart, on pages with two and three of them.
- **The chart and the table, always:** every page twice as long.
- **The `details` element, kept:** see the context.
