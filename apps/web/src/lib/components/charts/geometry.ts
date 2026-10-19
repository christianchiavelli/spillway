// The marks every chart shares, drawn to one spec: bars at most 24px thick, a
// 4px rounded end and a square base, a 2px gap between stacked segments, 2px
// lines and hairline grids. Everything is SVG geometry and classes, never a
// style attribute, which the content security policy would refuse.

export const GAP = 2;
export const RADIUS = 4;
export const MAX_BAR = 24;
/** The width of a 12px digit, near enough, to make room for an axis label. */
export const CHAR = 7;

/** A series of a chart: its name, and the classes that paint it. */
export type Series = { label: string; fill: string; stroke?: string; swatch?: string };

/** What a tooltip shows: a title, a row a series, and a total. */
export type Tip = {
  title: string;
  rows: { label: string; value: string; fill: string }[];
  total?: string | undefined;
};

/** Half a pixel off the grid, so a 1px line falls on one row of pixels. */
export const crisp = (v: number) => Math.round(v) + 0.5;

export const textWidth = (text: string) => text.length * CHAR;

/** A column with its top corners rounded and its base square. */
export function topRounded(x: number, y: number, w: number, h: number, r = RADIUS): string {
  const rr = Math.min(r, w / 2, h);
  return `M${x},${y + h}V${y + rr}A${rr},${rr} 0 0 1 ${x + rr},${y}H${x + w - rr}A${rr},${rr} 0 0 1 ${x + w},${y + rr}V${y + h}Z`;
}

/** A bar with its right corners rounded and its base, on the left, square. */
export function rightRounded(x: number, y: number, w: number, h: number, r = RADIUS): string {
  const rr = Math.min(r, h / 2, w);
  return `M${x},${y}H${x + w - rr}A${rr},${rr} 0 0 1 ${x + w},${y + rr}V${y + h - rr}A${rr},${rr} 0 0 1 ${x + w - rr},${y + h}H${x}Z`;
}

/**
 * Which labels to draw along an axis: left to right, each one that clears the
 * one before by gap and stays inside the chart. Returns whether all fit.
 */
export function fitLabels(
  labels: readonly { x: number; text: string }[],
  width: number,
  gap = 10,
): { shown: boolean[]; all: boolean } {
  let edge = -Infinity;
  let all = true;
  const shown = labels.map(({ x, text }) => {
    const half = textWidth(text) / 2;
    const fits = x - half >= edge + gap && x - half >= 0 && x + half <= width;
    if (fits) edge = x + half;
    else all = false;
    return fits;
  });
  return { shown, all };
}

/** Stacks values bottom up: each segment's start and end, a gap between. */
export function stack(
  values: readonly number[],
  scale: (v: number) => number,
): { from: number; to: number; index: number }[] {
  const out: { from: number; to: number; index: number }[] = [];
  let sum = 0;
  values.forEach((v, index) => {
    if (v <= 0) return;
    const from = scale(sum);
    sum += v;
    const to = scale(sum);
    out.push({ from, to, index });
  });
  return out;
}
